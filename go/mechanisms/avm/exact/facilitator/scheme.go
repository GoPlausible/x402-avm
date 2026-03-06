package facilitator

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/types"

	x402 "github.com/GoPlausible/x402-avm/go"
	"github.com/GoPlausible/x402-avm/go/mechanisms/avm"
	x402types "github.com/GoPlausible/x402-avm/go/types"
)

// ExactAvmScheme implements the SchemeNetworkFacilitator interface for AVM (Algorand) exact payments (V2)
type ExactAvmScheme struct {
	signer avm.FacilitatorAvmSigner
}

// NewExactAvmScheme creates a new ExactAvmScheme
func NewExactAvmScheme(signer avm.FacilitatorAvmSigner) *ExactAvmScheme {
	return &ExactAvmScheme{
		signer: signer,
	}
}

// Scheme returns the scheme identifier
func (f *ExactAvmScheme) Scheme() string {
	return avm.SchemeExact
}

// CaipFamily returns the CAIP family pattern this facilitator supports
func (f *ExactAvmScheme) CaipFamily() string {
	return "algorand:*"
}

// GetExtra returns mechanism-specific extra data for the supported kinds endpoint.
// For AVM, this includes a randomly selected fee payer address.
func (f *ExactAvmScheme) GetExtra(network x402.Network) map[string]interface{} {
	addresses := f.signer.GetAddresses(context.Background(), string(network))
	if len(addresses) == 0 {
		return nil
	}

	randomIndex := rand.Intn(len(addresses))

	return map[string]interface{}{
		"feePayer": addresses[randomIndex],
	}
}

// GetSigners returns signer addresses used by this facilitator.
func (f *ExactAvmScheme) GetSigners(network x402.Network) []string {
	return f.signer.GetAddresses(context.Background(), string(network))
}

// Verify verifies a V2 payment payload against requirements
func (f *ExactAvmScheme) Verify(
	ctx context.Context,
	payload x402types.PaymentPayload,
	requirements x402types.PaymentRequirements,
	_ *x402.FacilitatorContext,
) (*x402.VerifyResponse, error) {
	// Step 1: Validate scheme
	if payload.Accepted.Scheme != avm.SchemeExact || requirements.Scheme != avm.SchemeExact {
		return nil, x402.NewVerifyError(ErrUnsupportedScheme, "", fmt.Sprintf("invalid scheme: %s", payload.Accepted.Scheme))
	}

	// V2: Validate payload network matches requirements
	if string(payload.Accepted.Network) != string(requirements.Network) {
		return nil, x402.NewVerifyError(ErrNetworkMismatch, "", fmt.Sprintf("network mismatch: %s != %s", payload.Accepted.Network, requirements.Network))
	}

	// Validate fee payer
	if requirements.Extra == nil || requirements.Extra["feePayer"] == nil {
		return nil, x402.NewVerifyError(ErrMissingFeePayer, "", "missing feePayer")
	}

	feePayerStr, ok := requirements.Extra["feePayer"].(string)
	if !ok {
		return nil, x402.NewVerifyError(ErrMissingFeePayer, "", fmt.Sprintf("invalid feePayer: %v", requirements.Extra["feePayer"]))
	}

	// Verify fee payer is managed by this facilitator
	signerAddresses := f.signer.GetAddresses(ctx, string(requirements.Network))
	feePayerManaged := false
	for _, addr := range signerAddresses {
		if addr == feePayerStr {
			feePayerManaged = true
			break
		}
	}
	if !feePayerManaged {
		return nil, x402.NewVerifyError(ErrFeePayerNotManaged, "", fmt.Sprintf("feePayer not managed: %s", feePayerStr))
	}

	// Step 2: Parse AVM payload
	avmPayload, err := avm.PayloadFromMap(payload.Payload)
	if err != nil {
		return nil, x402.NewVerifyError(ErrInvalidPayloadFormat, "", err.Error())
	}

	// Validate group size
	if len(avmPayload.PaymentGroup) > avm.MaxAtomicGroupSize {
		return nil, x402.NewVerifyError(ErrGroupSizeExceeded, "", fmt.Sprintf("group size %d exceeds maximum %d", len(avmPayload.PaymentGroup), avm.MaxAtomicGroupSize))
	}

	// Step 3: Decode all transactions in the group
	decodedTxns := make([]avm.DecodedTransactionInfo, len(avmPayload.PaymentGroup))
	for i, encoded := range avmPayload.PaymentGroup {
		txn, hasSig, rawBytes, err := avm.DecodeTransactionAuto(encoded)
		if err != nil {
			return nil, x402.NewVerifyError(ErrInvalidTransaction, "", fmt.Sprintf("failed to decode transaction %d: %v", i, err))
		}

		info := avm.DecodedTransactionInfo{
			Txn: *txn,
			Raw: rawBytes,
		}

		if hasSig {
			stxn, err := avm.DecodeSignedTransaction(encoded)
			if err != nil {
				return nil, x402.NewVerifyError(ErrInvalidTransaction, "", fmt.Sprintf("failed to decode signed transaction %d: %v", i, err))
			}
			info.Signed = stxn
		}

		decodedTxns[i] = info
	}

	// Step 4: Validate group IDs match (if multiple transactions)
	if len(decodedTxns) > 1 {
		firstGroup := decodedTxns[0].Txn.Group
		for i, info := range decodedTxns {
			if info.Txn.Group != firstGroup {
				return nil, x402.NewVerifyError(ErrInvalidGroupID, "", fmt.Sprintf("transaction %d has different group ID", i))
			}
		}
	}

	// Step 5: Get expected genesis hash for network
	expectedGenesisHash, err := avm.GetGenesisHashForNetwork(string(requirements.Network))
	if err != nil {
		return nil, x402.NewVerifyError(ErrGenesisHashMismatch, "", err.Error())
	}

	// Step 6: Security checks on ALL transactions
	for i, info := range decodedTxns {
		txn := info.Txn

		// Validate genesis hash
		genesisHash := base64.StdEncoding.EncodeToString(txn.GenesisHash[:])
		if genesisHash != expectedGenesisHash {
			return nil, x402.NewVerifyError(ErrGenesisHashMismatch, "", fmt.Sprintf("transaction %d genesis hash mismatch: expected %s, got %s", i, expectedGenesisHash, genesisHash))
		}

		// No keyreg transactions allowed
		if txn.Type == types.KeyRegistrationTx {
			return nil, x402.NewVerifyError(ErrSecurityKeyregNotAllowed, "", fmt.Sprintf("transaction %d is keyreg", i))
		}

		// No rekey allowed
		if !txn.RekeyTo.IsZero() {
			return nil, x402.NewVerifyError(ErrSecurityRekeyNotAllowed, "", fmt.Sprintf("transaction %d has rekeyTo", i))
		}

		// No close-remainder-to for payments
		if txn.Type == types.PaymentTx && !txn.CloseRemainderTo.IsZero() {
			return nil, x402.NewVerifyError(ErrSecurityCloseToNotAllowed, "", fmt.Sprintf("transaction %d has closeRemainderTo", i))
		}

		// No close-to for asset transfers
		if txn.Type == types.AssetTransferTx && !txn.AssetCloseTo.IsZero() {
			return nil, x402.NewVerifyError(ErrSecurityCloseToNotAllowed, "", fmt.Sprintf("transaction %d has assetCloseTo", i))
		}

		// Unsigned transactions must be from facilitator addresses
		if info.Signed == nil {
			sender := txn.Sender.String()
			isFacilitatorAddr := false
			for _, addr := range signerAddresses {
				if addr == sender {
					isFacilitatorAddr = true
					break
				}
			}
			if !isFacilitatorAddr {
				return nil, x402.NewVerifyError(ErrUnsignedNonFacilitatorTxn, "", fmt.Sprintf("unsigned transaction %d from non-facilitator address: %s", i, sender))
			}
		}
	}

	// Step 7: Verify payment transaction
	paymentTxnInfo := decodedTxns[avmPayload.PaymentIndex]
	paymentTxn := paymentTxnInfo.Txn
	payer := paymentTxn.Sender.String()

	// Payment must be an ASA transfer
	if paymentTxn.Type != types.AssetTransferTx {
		return nil, x402.NewVerifyError(ErrPaymentNotAssetTransfer, payer, fmt.Sprintf("expected axfer, got %s", paymentTxn.Type))
	}

	// Verify asset matches
	requiredAssetID, err := strconv.ParseUint(requirements.Asset, 10, 64)
	if err != nil {
		return nil, x402.NewVerifyError(ErrAssetMismatch, payer, fmt.Sprintf("invalid asset ID: %s", requirements.Asset))
	}
	if uint64(paymentTxn.XferAsset) != requiredAssetID {
		return nil, x402.NewVerifyError(ErrAssetMismatch, payer, fmt.Sprintf("expected asset %d, got %d", requiredAssetID, paymentTxn.XferAsset))
	}

	// Verify receiver matches payTo
	if paymentTxn.AssetReceiver.String() != requirements.PayTo {
		return nil, x402.NewVerifyError(ErrReceiverMismatch, payer, fmt.Sprintf("expected receiver %s, got %s", requirements.PayTo, paymentTxn.AssetReceiver.String()))
	}

	// Verify amount
	requiredAmount, err := strconv.ParseUint(requirements.Amount, 10, 64)
	if err != nil {
		return nil, x402.NewVerifyError(ErrAmountMismatch, payer, fmt.Sprintf("invalid amount: %s", requirements.Amount))
	}
	if paymentTxn.AssetAmount < requiredAmount {
		return nil, x402.NewVerifyError(ErrAmountMismatch, payer, fmt.Sprintf("expected amount >= %d, got %d", requiredAmount, paymentTxn.AssetAmount))
	}

	// SECURITY: Verify facilitator addresses are not the payment sender
	for _, addr := range signerAddresses {
		if payer == addr {
			return nil, x402.NewVerifyError(ErrFacilitatorTransferringFunds, payer, "facilitator address cannot be payment sender")
		}
	}

	// Payment transaction must be signed
	if paymentTxnInfo.Signed == nil {
		return nil, x402.NewVerifyError(ErrPaymentNotSigned, payer, "payment transaction is not signed")
	}

	// Verify Ed25519 signature on the payment transaction
	if !avm.HasSignature(paymentTxnInfo.Signed) {
		return nil, x402.NewVerifyError(ErrPaymentNotSigned, payer, "payment transaction has no signature")
	}

	// Verify signature: Algorand signs "TX" prefix + raw transaction bytes
	if err := verifyEd25519Signature(paymentTxnInfo.Signed); err != nil {
		return nil, x402.NewVerifyError(ErrInvalidSignature, payer, err.Error())
	}

	// Step 8: Verify fee payer transaction (if present and separate from payment)
	for i, info := range decodedTxns {
		if i == avmPayload.PaymentIndex {
			continue
		}

		txn := info.Txn
		sender := txn.Sender.String()

		// Check if this is a facilitator fee payer transaction
		isFacilitatorSender := false
		for _, addr := range signerAddresses {
			if sender == addr {
				isFacilitatorSender = true
				break
			}
		}

		if isFacilitatorSender {
			// Fee payer transaction: must be payment type
			if txn.Type != types.PaymentTx {
				return nil, x402.NewVerifyError(ErrFeePayerInvalidType, payer, fmt.Sprintf("fee payer txn %d: expected pay, got %s", i, txn.Type))
			}

			// Amount must be 0
			if txn.Amount != 0 {
				return nil, x402.NewVerifyError(ErrFeePayerNonZeroAmount, payer, fmt.Sprintf("fee payer txn %d has non-zero amount: %d", i, txn.Amount))
			}

			// Must be self-payment (receiver == sender)
			if txn.Receiver.String() != sender {
				return nil, x402.NewVerifyError(ErrFeePayerNotSelfPayment, payer, fmt.Sprintf("fee payer txn %d receiver mismatch", i))
			}

			// No closeRemainderTo
			if !txn.CloseRemainderTo.IsZero() {
				return nil, x402.NewVerifyError(ErrFeePayerHasCloseTo, payer, fmt.Sprintf("fee payer txn %d has closeRemainderTo", i))
			}

			// No rekeyTo
			if !txn.RekeyTo.IsZero() {
				return nil, x402.NewVerifyError(ErrFeePayerHasRekeyTo, payer, fmt.Sprintf("fee payer txn %d has rekeyTo", i))
			}

			// Fee must be reasonable
			if uint64(txn.Fee) > uint64(avm.MaxReasonableFee) {
				return nil, x402.NewVerifyError(ErrFeeTooHigh, payer, fmt.Sprintf("fee payer txn %d fee %d exceeds maximum %d", i, txn.Fee, avm.MaxReasonableFee))
			}
		}
	}

	// Step 9: Simulate transaction group
	// Build simulation group: sign facilitator txns temporarily, wrap unsigned for simulation
	simTxns := make([][]byte, len(avmPayload.PaymentGroup))
	for i, encoded := range avmPayload.PaymentGroup {
		rawBytes, err := avm.DecodeRawTransaction(encoded)
		if err != nil {
			return nil, x402.NewVerifyError(ErrSimulationFailed, payer, fmt.Sprintf("failed to decode txn %d for simulation: %v", i, err))
		}
		simTxns[i] = rawBytes
	}

	if err := f.signer.SimulateTransactions(ctx, simTxns, string(requirements.Network)); err != nil {
		return nil, x402.NewVerifyError(ErrSimulationFailed, payer, err.Error())
	}

	return &x402.VerifyResponse{
		IsValid: true,
		Payer:   payer,
	}, nil
}

// Settle settles a payment by signing facilitator transactions and submitting the group
func (f *ExactAvmScheme) Settle(
	ctx context.Context,
	payload x402types.PaymentPayload,
	requirements x402types.PaymentRequirements,
	fctx *x402.FacilitatorContext,
) (*x402.SettleResponse, error) {
	network := x402.Network(requirements.Network)

	// First verify the payment
	verifyResp, err := f.Verify(ctx, payload, requirements, fctx)
	if err != nil {
		ve := &x402.VerifyError{}
		if errors.As(err, &ve) {
			return nil, x402.NewSettleError(ve.InvalidReason, ve.Payer, network, "", ve.InvalidMessage)
		}
		return nil, x402.NewSettleError(ErrVerificationFailed, "", network, "", err.Error())
	}

	// Parse payload
	avmPayload, err := avm.PayloadFromMap(payload.Payload)
	if err != nil {
		return nil, x402.NewSettleError(ErrInvalidPayloadFormat, verifyResp.Payer, network, "", err.Error())
	}

	// Sign facilitator transactions and build the final group
	signedGroup := make([][]byte, len(avmPayload.PaymentGroup))
	for i, encoded := range avmPayload.PaymentGroup {
		rawBytes, err := avm.DecodeRawTransaction(encoded)
		if err != nil {
			return nil, x402.NewSettleError(ErrTransactionSigningFailed, verifyResp.Payer, network, "", fmt.Sprintf("failed to decode txn %d: %v", i, err))
		}

		txn, hasSig, _, decErr := avm.DecodeTransactionAuto(encoded)
		if decErr != nil {
			return nil, x402.NewSettleError(ErrTransactionSigningFailed, verifyResp.Payer, network, "", fmt.Sprintf("failed to decode txn %d: %v", i, decErr))
		}

		sender := txn.Sender.String()

		if !hasSig {
			// Unsigned transaction from facilitator - sign it
			signedBytes, err := f.signer.SignTransaction(ctx, rawBytes, sender, string(requirements.Network))
			if err != nil {
				return nil, x402.NewSettleError(ErrTransactionSigningFailed, verifyResp.Payer, network, "", fmt.Sprintf("failed to sign txn %d: %v", i, err))
			}
			signedGroup[i] = signedBytes
		} else {
			// Already signed (client's transaction) - use as-is
			signedGroup[i] = rawBytes
		}
	}

	// Submit the transaction group
	txID, err := f.signer.SendTransactions(ctx, signedGroup, string(requirements.Network))
	if err != nil {
		return nil, x402.NewSettleError(ErrTransactionFailed, verifyResp.Payer, network, "", err.Error())
	}

	// Wait for confirmation
	if err := f.signer.WaitForConfirmation(ctx, txID, string(requirements.Network), 4); err != nil {
		return nil, x402.NewSettleError(ErrTransactionConfirmationFailed, verifyResp.Payer, network, txID, err.Error())
	}

	return &x402.SettleResponse{
		Success:     true,
		Transaction: txID,
		Network:     network,
		Payer:       verifyResp.Payer,
	}, nil
}

// verifyEd25519Signature verifies the Ed25519 signature on a signed Algorand transaction.
// Algorand signs: "TX" prefix (0x54, 0x58) + msgpack-encoded transaction bytes.
func verifyEd25519Signature(stxn *types.SignedTxn) error {
	if stxn == nil {
		return fmt.Errorf("nil signed transaction")
	}

	// Only verify standard Ed25519 signatures (not multisig or logicsig)
	sig := stxn.Sig
	isZero := true
	for _, b := range sig {
		if b != 0 {
			isZero = false
			break
		}
	}
	if isZero {
		return fmt.Errorf("no Ed25519 signature present")
	}

	// Get the public key from sender address
	senderAddr := stxn.Txn.Sender
	pubKey := ed25519.PublicKey(senderAddr[:])

	// Build the signed message: "TX" + msgpack(transaction)
	txnBytes := msgpack.Encode(stxn.Txn)
	signedMsg := make([]byte, 2+len(txnBytes))
	signedMsg[0] = 0x54 // 'T'
	signedMsg[1] = 0x58 // 'X'
	copy(signedMsg[2:], txnBytes)

	// Verify
	if !ed25519.Verify(pubKey, signedMsg, sig[:]) {
		return fmt.Errorf("Ed25519 signature verification failed")
	}

	return nil
}

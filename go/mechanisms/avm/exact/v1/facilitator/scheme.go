package facilitator

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/types"

	x402 "github.com/coinbase/x402/go"
	"github.com/coinbase/x402/go/mechanisms/avm"
	x402types "github.com/coinbase/x402/go/types"
)

// ExactAvmSchemeV1 implements the SchemeNetworkFacilitatorV1 interface for AVM (Algorand) exact payments (V1)
type ExactAvmSchemeV1 struct {
	signer avm.FacilitatorAvmSigner
}

// NewExactAvmSchemeV1 creates a new ExactAvmSchemeV1
func NewExactAvmSchemeV1(signer avm.FacilitatorAvmSigner) *ExactAvmSchemeV1 {
	return &ExactAvmSchemeV1{
		signer: signer,
	}
}

// Scheme returns the scheme identifier
func (f *ExactAvmSchemeV1) Scheme() string {
	return avm.SchemeExact
}

// CaipFamily returns the CAIP family pattern this facilitator supports
func (f *ExactAvmSchemeV1) CaipFamily() string {
	return "algorand:*"
}

// GetExtra returns mechanism-specific extra data for the supported kinds endpoint
func (f *ExactAvmSchemeV1) GetExtra(network x402.Network) map[string]interface{} {
	addresses := f.signer.GetAddresses(context.Background(), string(network))
	if len(addresses) == 0 {
		return nil
	}

	randomIndex := rand.Intn(len(addresses))

	return map[string]interface{}{
		"feePayer": addresses[randomIndex],
	}
}

// GetSigners returns signer addresses used by this facilitator
func (f *ExactAvmSchemeV1) GetSigners(network x402.Network) []string {
	return f.signer.GetAddresses(context.Background(), string(network))
}

// Verify verifies a V1 payment payload against requirements
func (f *ExactAvmSchemeV1) Verify(
	ctx context.Context,
	payload x402types.PaymentPayloadV1,
	requirements x402types.PaymentRequirementsV1,
	_ *x402.FacilitatorContext,
) (*x402.VerifyResponse, error) {
	// Step 1: Validate scheme
	if payload.Scheme != avm.SchemeExact || requirements.Scheme != avm.SchemeExact {
		return nil, x402.NewVerifyError(ErrUnsupportedScheme, "", fmt.Sprintf("invalid scheme: %s", payload.Scheme))
	}

	if payload.Network != requirements.Network {
		return nil, x402.NewVerifyError(ErrNetworkMismatch, "", fmt.Sprintf("network mismatch: %s != %s", payload.Network, requirements.Network))
	}

	// Parse extra field for feePayer (V1 uses *json.RawMessage)
	var reqExtraMap map[string]interface{}
	if requirements.Extra != nil {
		if err := json.Unmarshal(*requirements.Extra, &reqExtraMap); err != nil {
			return nil, x402.NewVerifyError(ErrInvalidExtraField, "", err.Error())
		}
	}

	if reqExtraMap == nil || reqExtraMap["feePayer"] == nil {
		return nil, x402.NewVerifyError(ErrMissingFeePayer, "", "missing feePayer")
	}

	feePayerStr, ok := reqExtraMap["feePayer"].(string)
	if !ok {
		return nil, x402.NewVerifyError(ErrMissingFeePayer, "", fmt.Sprintf("invalid feePayer: %v", reqExtraMap["feePayer"]))
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

	if len(avmPayload.PaymentGroup) > avm.MaxAtomicGroupSize {
		return nil, x402.NewVerifyError(ErrGroupSizeExceeded, "", fmt.Sprintf("group size %d exceeds maximum %d", len(avmPayload.PaymentGroup), avm.MaxAtomicGroupSize))
	}

	// Step 3: Decode all transactions
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

	// Step 4: Validate group IDs
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

		genesisHash := base64.StdEncoding.EncodeToString(txn.GenesisHash[:])
		if genesisHash != expectedGenesisHash {
			return nil, x402.NewVerifyError(ErrGenesisHashMismatch, "", fmt.Sprintf("transaction %d genesis hash mismatch", i))
		}

		if txn.Type == types.KeyRegistrationTx {
			return nil, x402.NewVerifyError(ErrSecurityKeyregNotAllowed, "", fmt.Sprintf("transaction %d is keyreg", i))
		}

		if !txn.RekeyTo.IsZero() {
			return nil, x402.NewVerifyError(ErrSecurityRekeyNotAllowed, "", fmt.Sprintf("transaction %d has rekeyTo", i))
		}

		if txn.Type == types.PaymentTx && !txn.CloseRemainderTo.IsZero() {
			return nil, x402.NewVerifyError(ErrSecurityCloseToNotAllowed, "", fmt.Sprintf("transaction %d has closeRemainderTo", i))
		}

		if txn.Type == types.AssetTransferTx && !txn.AssetCloseTo.IsZero() {
			return nil, x402.NewVerifyError(ErrSecurityCloseToNotAllowed, "", fmt.Sprintf("transaction %d has assetCloseTo", i))
		}

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

	if paymentTxn.Type != types.AssetTransferTx {
		return nil, x402.NewVerifyError(ErrPaymentNotAssetTransfer, payer, fmt.Sprintf("expected axfer, got %s", paymentTxn.Type))
	}

	requiredAssetID, err := strconv.ParseUint(requirements.Asset, 10, 64)
	if err != nil {
		return nil, x402.NewVerifyError(ErrAssetMismatch, payer, fmt.Sprintf("invalid asset ID: %s", requirements.Asset))
	}
	if uint64(paymentTxn.XferAsset) != requiredAssetID {
		return nil, x402.NewVerifyError(ErrAssetMismatch, payer, fmt.Sprintf("expected asset %d, got %d", requiredAssetID, paymentTxn.XferAsset))
	}

	if paymentTxn.AssetReceiver.String() != requirements.PayTo {
		return nil, x402.NewVerifyError(ErrReceiverMismatch, payer, fmt.Sprintf("expected receiver %s, got %s", requirements.PayTo, paymentTxn.AssetReceiver.String()))
	}

	// V1: Use MaxAmountRequired
	requiredAmount, err := strconv.ParseUint(requirements.MaxAmountRequired, 10, 64)
	if err != nil {
		return nil, x402.NewVerifyError(ErrAmountMismatch, payer, fmt.Sprintf("invalid amount: %s", requirements.MaxAmountRequired))
	}
	if paymentTxn.AssetAmount < requiredAmount {
		return nil, x402.NewVerifyError(ErrAmountMismatch, payer, fmt.Sprintf("expected amount >= %d, got %d", requiredAmount, paymentTxn.AssetAmount))
	}

	// SECURITY: Facilitator addresses cannot be payment sender
	for _, addr := range signerAddresses {
		if payer == addr {
			return nil, x402.NewVerifyError(ErrFacilitatorTransferringFunds, payer, "facilitator address cannot be payment sender")
		}
	}

	// Verify signature
	if paymentTxnInfo.Signed == nil || !avm.HasSignature(paymentTxnInfo.Signed) {
		return nil, x402.NewVerifyError(ErrPaymentNotSigned, payer, "payment transaction is not signed")
	}

	if err := verifyEd25519Signature(paymentTxnInfo.Signed); err != nil {
		return nil, x402.NewVerifyError(ErrInvalidSignature, payer, err.Error())
	}

	// Step 8: Verify fee payer transactions
	for i, info := range decodedTxns {
		if i == avmPayload.PaymentIndex {
			continue
		}

		txn := info.Txn
		sender := txn.Sender.String()

		isFacilitatorSender := false
		for _, addr := range signerAddresses {
			if sender == addr {
				isFacilitatorSender = true
				break
			}
		}

		if isFacilitatorSender {
			if txn.Type != types.PaymentTx {
				return nil, x402.NewVerifyError(ErrFeePayerInvalidType, payer, fmt.Sprintf("fee payer txn %d: expected pay, got %s", i, txn.Type))
			}
			if txn.Amount != 0 {
				return nil, x402.NewVerifyError(ErrFeePayerNonZeroAmount, payer, fmt.Sprintf("fee payer txn %d has non-zero amount", i))
			}
			if txn.Receiver.String() != sender {
				return nil, x402.NewVerifyError(ErrFeePayerNotSelfPayment, payer, fmt.Sprintf("fee payer txn %d receiver mismatch", i))
			}
			if !txn.CloseRemainderTo.IsZero() {
				return nil, x402.NewVerifyError(ErrFeePayerHasCloseTo, payer, fmt.Sprintf("fee payer txn %d has closeRemainderTo", i))
			}
			if !txn.RekeyTo.IsZero() {
				return nil, x402.NewVerifyError(ErrFeePayerHasRekeyTo, payer, fmt.Sprintf("fee payer txn %d has rekeyTo", i))
			}
			if uint64(txn.Fee) > uint64(avm.MaxReasonableFee) {
				return nil, x402.NewVerifyError(ErrFeeTooHigh, payer, fmt.Sprintf("fee payer txn %d fee exceeds maximum", i))
			}
		}
	}

	// Step 9: Simulate
	simTxns := make([][]byte, len(avmPayload.PaymentGroup))
	for i, encoded := range avmPayload.PaymentGroup {
		rawBytes, err := avm.DecodeRawTransaction(encoded)
		if err != nil {
			return nil, x402.NewVerifyError(ErrSimulationFailed, payer, fmt.Sprintf("failed to decode txn %d: %v", i, err))
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

// Settle settles a payment by signing facilitator transactions and submitting the group (V1)
func (f *ExactAvmSchemeV1) Settle(
	ctx context.Context,
	payload x402types.PaymentPayloadV1,
	requirements x402types.PaymentRequirementsV1,
	fctx *x402.FacilitatorContext,
) (*x402.SettleResponse, error) {
	network := x402.Network(payload.Network)

	verifyResp, err := f.Verify(ctx, payload, requirements, fctx)
	if err != nil {
		ve := &x402.VerifyError{}
		if errors.As(err, &ve) {
			return nil, x402.NewSettleError(ve.InvalidReason, ve.Payer, network, "", ve.InvalidMessage)
		}
		return nil, x402.NewSettleError(ErrVerificationFailed, "", network, "", err.Error())
	}

	avmPayload, err := avm.PayloadFromMap(payload.Payload)
	if err != nil {
		return nil, x402.NewSettleError(ErrInvalidPayloadFormat, verifyResp.Payer, network, "", err.Error())
	}

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
			signedBytes, err := f.signer.SignTransaction(ctx, rawBytes, sender, string(requirements.Network))
			if err != nil {
				return nil, x402.NewSettleError(ErrTransactionSigningFailed, verifyResp.Payer, network, "", fmt.Sprintf("failed to sign txn %d: %v", i, err))
			}
			signedGroup[i] = signedBytes
		} else {
			signedGroup[i] = rawBytes
		}
	}

	txID, err := f.signer.SendTransactions(ctx, signedGroup, string(requirements.Network))
	if err != nil {
		return nil, x402.NewSettleError(ErrTransactionFailed, verifyResp.Payer, network, "", err.Error())
	}

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

// verifyEd25519Signature verifies the Ed25519 signature on a signed Algorand transaction
func verifyEd25519Signature(stxn *types.SignedTxn) error {
	if stxn == nil {
		return fmt.Errorf("nil signed transaction")
	}

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

	senderAddr := stxn.Txn.Sender
	pubKey := ed25519.PublicKey(senderAddr[:])

	txnBytes := msgpack.Encode(stxn.Txn)
	signedMsg := make([]byte, 2+len(txnBytes))
	signedMsg[0] = 0x54 // 'T'
	signedMsg[1] = 0x58 // 'X'
	copy(signedMsg[2:], txnBytes)

	if !ed25519.Verify(pubKey, signedMsg, sig[:]) {
		return fmt.Errorf("Ed25519 signature verification failed")
	}

	return nil
}

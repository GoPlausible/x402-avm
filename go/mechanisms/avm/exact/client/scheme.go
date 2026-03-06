package client

import (
	"context"
	"fmt"
	"strconv"

	algodClient "github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/algorand/go-algorand-sdk/v2/types"

	"github.com/GoPlausible/x402-avm/go/mechanisms/avm"
	x402types "github.com/GoPlausible/x402-avm/go/types"
)

// ExactAvmScheme implements the SchemeNetworkClient interface for AVM (Algorand) exact payments (V2)
type ExactAvmScheme struct {
	signer avm.ClientAvmSigner
	config *avm.ClientConfig
}

// NewExactAvmScheme creates a new ExactAvmScheme
// Config is optional - if not provided, uses network defaults
func NewExactAvmScheme(signer avm.ClientAvmSigner, config ...*avm.ClientConfig) *ExactAvmScheme {
	var cfg *avm.ClientConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	return &ExactAvmScheme{
		signer: signer,
		config: cfg,
	}
}

// Scheme returns the scheme identifier
func (c *ExactAvmScheme) Scheme() string {
	return avm.SchemeExact
}

// CreatePaymentPayload creates a V2 payment payload for the Exact scheme
func (c *ExactAvmScheme) CreatePaymentPayload(
	ctx context.Context,
	requirements x402types.PaymentRequirements,
) (x402types.PaymentPayload, error) {
	// Validate network
	networkStr := string(requirements.Network)
	if !avm.IsValidNetwork(networkStr) {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrUnsupportedNetwork+": %s", requirements.Network)
	}

	// Get network configuration
	config, err := avm.GetNetworkConfig(networkStr)
	if err != nil {
		return x402types.PaymentPayload{}, err
	}

	// Get Algod URL (custom or default)
	algodURL := config.AlgodURL
	algodToken := config.AlgodToken
	if c.config != nil && c.config.AlgodURL != "" {
		algodURL = c.config.AlgodURL
		algodToken = c.config.AlgodToken
	}

	// Create Algod client
	algod, err := algodClient.MakeClient(algodURL, algodToken)
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrFailedToGetTransactionParams+": %w", err)
	}

	// Get suggested transaction parameters
	txParams, err := algod.SuggestedParams().Do(ctx)
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrFailedToGetTransactionParams+": %w", err)
	}

	// Parse asset ID
	assetID, err := strconv.ParseUint(requirements.Asset, 10, 64)
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrInvalidAssetID+": %w", err)
	}

	// Validate payTo address
	if !avm.ValidateAlgorandAddress(requirements.PayTo) {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrInvalidPayToAddress+": %s", requirements.PayTo)
	}

	// Parse amount
	amount, err := strconv.ParseUint(requirements.Amount, 10, 64)
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrInvalidAmount+": %w", err)
	}

	// Get fee payer from requirements.extra
	feePayerAddr, ok := requirements.Extra["feePayer"].(string)
	if !ok || feePayerAddr == "" {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrFeePayerRequired)
	}

	if !avm.ValidateAlgorandAddress(feePayerAddr) {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrInvalidFeePayerAddress+": %s", feePayerAddr)
	}

	clientAddress := c.signer.Address()

	// Build transaction group: fee payer (index 0) + ASA transfer (index 1)
	groupSize := 2
	pooledFee := uint64(txParams.MinFee) * uint64(groupSize)

	// Fee payer transaction: self-payment, 0 amount, pooled fee covering entire group
	feePayerParams := types.SuggestedParams{
		Fee:             types.MicroAlgos(pooledFee),
		FirstRoundValid: txParams.FirstRoundValid,
		LastRoundValid:  txParams.LastRoundValid,
		GenesisID:       txParams.GenesisID,
		GenesisHash:     txParams.GenesisHash,
		FlatFee:         true,
		MinFee:          txParams.MinFee,
	}

	feePayerTxn, err := transaction.MakePaymentTxn(
		feePayerAddr, feePayerAddr, 0, nil, "", feePayerParams,
	)
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrFailedToBuildFeePayerTxn+": %w", err)
	}

	// ASA transfer transaction: client -> payTo, fee=0 (covered by fee payer)
	paymentParams := types.SuggestedParams{
		Fee:             0,
		FirstRoundValid: txParams.FirstRoundValid,
		LastRoundValid:  txParams.LastRoundValid,
		GenesisID:       txParams.GenesisID,
		GenesisHash:     txParams.GenesisHash,
		FlatFee:         true,
		MinFee:          txParams.MinFee,
	}

	assetTransferTxn, err := transaction.MakeAssetTransferTxn(
		clientAddress, requirements.PayTo, amount, nil, paymentParams, "", assetID,
	)
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrFailedToBuildAssetTransferTxn+": %w", err)
	}

	// Assign group ID to both transactions
	txns := []types.Transaction{feePayerTxn, assetTransferTxn}
	groupedTxns, err := transaction.AssignGroupID(txns, "")
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrFailedToAssignGroupID+": %w", err)
	}

	// Encode transactions to raw bytes for signing
	encodedTxns := make([][]byte, len(groupedTxns))
	for i, txn := range groupedTxns {
		encodedTxns[i] = msgpack.Encode(txn)
	}

	// Sign only the client's transaction (ASA transfer at index 1)
	paymentIndex := 1
	signedResults, err := c.signer.SignTransactions(ctx, encodedTxns, []int{paymentIndex})
	if err != nil {
		return x402types.PaymentPayload{}, fmt.Errorf(ErrFailedToSignTransaction+": %w", err)
	}

	// Build payment group: base64-encoded transactions
	// Index 0: unsigned fee payer txn (facilitator will sign)
	// Index 1: signed ASA transfer (client signed)
	paymentGroup := make([]string, len(groupedTxns))
	for i := range groupedTxns {
		if signedResults[i] != nil {
			// Signed transaction bytes (already msgpack-encoded SignedTxn)
			paymentGroup[i] = avm.EncodeSignedTransactionBytes(signedResults[i])
		} else {
			// Unsigned transaction (msgpack-encoded Transaction)
			paymentGroup[i] = avm.EncodeTransaction(groupedTxns[i])
		}
	}

	avmPayload := &avm.ExactAvmPayload{
		PaymentGroup: paymentGroup,
		PaymentIndex: paymentIndex,
	}

	return x402types.PaymentPayload{
		X402Version: 2,
		Payload:     avmPayload.ToMap(),
	}, nil
}

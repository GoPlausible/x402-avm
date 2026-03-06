package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	algodClient "github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/algorand/go-algorand-sdk/v2/types"

	"github.com/GoPlausible/x402-avm/go/mechanisms/avm"
	x402types "github.com/GoPlausible/x402-avm/go/types"
)

// ExactAvmSchemeV1 implements the SchemeNetworkClientV1 interface for AVM (Algorand) exact payments (V1)
type ExactAvmSchemeV1 struct {
	signer avm.ClientAvmSigner
	config *avm.ClientConfig
}

// NewExactAvmSchemeV1 creates a new ExactAvmSchemeV1
func NewExactAvmSchemeV1(signer avm.ClientAvmSigner, config ...*avm.ClientConfig) *ExactAvmSchemeV1 {
	var cfg *avm.ClientConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	return &ExactAvmSchemeV1{
		signer: signer,
		config: cfg,
	}
}

// Scheme returns the scheme identifier
func (c *ExactAvmSchemeV1) Scheme() string {
	return avm.SchemeExact
}

// CreatePaymentPayload creates a V1 payment payload for the Exact scheme
func (c *ExactAvmSchemeV1) CreatePaymentPayload(
	ctx context.Context,
	requirements x402types.PaymentRequirementsV1,
) (x402types.PaymentPayloadV1, error) {
	// Validate network (V1 uses simple names, normalize internally)
	networkStr := requirements.Network
	if !avm.IsValidNetwork(networkStr) {
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrUnsupportedNetwork+": %s", requirements.Network)
	}

	config, err := avm.GetNetworkConfig(networkStr)
	if err != nil {
		return x402types.PaymentPayloadV1{}, err
	}

	algodURL := config.AlgodURL
	algodToken := config.AlgodToken
	if c.config != nil && c.config.AlgodURL != "" {
		algodURL = c.config.AlgodURL
		algodToken = c.config.AlgodToken
	}

	algod, err := algodClient.MakeClient(algodURL, algodToken)
	if err != nil {
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrFailedToGetTransactionParams+": %w", err)
	}

	txParams, err := algod.SuggestedParams().Do(ctx)
	if err != nil {
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrFailedToGetTransactionParams+": %w", err)
	}

	assetID, err := strconv.ParseUint(requirements.Asset, 10, 64)
	if err != nil {
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrInvalidAssetID+": %w", err)
	}

	if !avm.ValidateAlgorandAddress(requirements.PayTo) {
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrInvalidPayToAddress+": %s", requirements.PayTo)
	}

	// V1: Use MaxAmountRequired field
	amount, err := strconv.ParseUint(requirements.MaxAmountRequired, 10, 64)
	if err != nil {
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrInvalidAmount+": %w", err)
	}

	// Parse fee payer from V1 extra (json.RawMessage)
	var extraMap map[string]interface{}
	if requirements.Extra != nil {
		if err := json.Unmarshal(*requirements.Extra, &extraMap); err != nil {
			return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrInvalidExtraField+": %w", err)
		}
	}

	feePayerAddr, ok := extraMap["feePayer"].(string)
	if !ok || feePayerAddr == "" {
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrFeePayerRequired)
	}

	if !avm.ValidateAlgorandAddress(feePayerAddr) {
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrInvalidFeePayerAddress+": %s", feePayerAddr)
	}

	clientAddress := c.signer.Address()

	// Build transaction group
	groupSize := 2
	pooledFee := uint64(txParams.MinFee) * uint64(groupSize)

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
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrFailedToBuildFeePayerTxn+": %w", err)
	}

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
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrFailedToBuildAssetTransferTxn+": %w", err)
	}

	txns := []types.Transaction{feePayerTxn, assetTransferTxn}
	groupedTxns, err := transaction.AssignGroupID(txns, "")
	if err != nil {
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrFailedToAssignGroupID+": %w", err)
	}

	encodedTxns := make([][]byte, len(groupedTxns))
	for i, txn := range groupedTxns {
		encodedTxns[i] = msgpack.Encode(txn)
	}

	paymentIndex := 1
	signedResults, err := c.signer.SignTransactions(ctx, encodedTxns, []int{paymentIndex})
	if err != nil {
		return x402types.PaymentPayloadV1{}, fmt.Errorf(ErrFailedToSignTransaction+": %w", err)
	}

	paymentGroup := make([]string, len(groupedTxns))
	for i := range groupedTxns {
		if signedResults[i] != nil {
			paymentGroup[i] = avm.EncodeSignedTransactionBytes(signedResults[i])
		} else {
			paymentGroup[i] = avm.EncodeTransaction(groupedTxns[i])
		}
	}

	avmPayload := &avm.ExactAvmPayload{
		PaymentGroup: paymentGroup,
		PaymentIndex: paymentIndex,
	}

	return x402types.PaymentPayloadV1{
		X402Version: 1,
		Scheme:      requirements.Scheme,
		Network:     requirements.Network,
		Payload:     avmPayload.ToMap(),
	}, nil
}

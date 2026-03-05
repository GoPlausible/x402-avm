package avm

import (
	"context"
	"encoding/json"
	"fmt"

	algodClient "github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/types"
)

// ExactAvmPayload represents an AVM (Algorand) payment payload
type ExactAvmPayload struct {
	PaymentGroup []string `json:"paymentGroup"` // Base64-encoded msgpack transactions
	PaymentIndex int      `json:"paymentIndex"` // Index of payment txn in group
}

// ExactAvmPayloadV1 - alias for v1 compatibility
type ExactAvmPayloadV1 = ExactAvmPayload

// ExactAvmPayloadV2 - alias for v2 (currently identical)
type ExactAvmPayloadV2 = ExactAvmPayload

// ClientAvmSigner defines client-side operations for Algorand
type ClientAvmSigner interface {
	// Address returns the signer's Algorand address (58-char base32)
	Address() string

	// SignTransactions signs transactions in the group.
	// indexesToSign specifies which transactions in the group to sign.
	// Returns signed transaction bytes for each index, nil for unsigned ones.
	SignTransactions(ctx context.Context, txns [][]byte, indexesToSign []int) ([][]byte, error)
}

// FacilitatorAvmSigner defines facilitator operations for AVM
// Supports multiple signers for load balancing, key rotation, and high availability
type FacilitatorAvmSigner interface {
	// GetAddresses returns all addresses this facilitator can use as fee payers for a network
	GetAddresses(ctx context.Context, network string) []string

	// SignTransaction signs a single transaction with the signer matching senderAddress
	// Returns the signed transaction bytes (msgpack-encoded SignedTxn)
	SignTransaction(ctx context.Context, txnBytes []byte, senderAddress string, network string) ([]byte, error)

	// GetAlgodClient returns the Algod client for the given network
	GetAlgodClient(network string) (*algodClient.Client, error)

	// SimulateTransactions simulates a group of transactions (signed or unsigned)
	// Returns error if simulation fails
	SimulateTransactions(ctx context.Context, txns [][]byte, network string) error

	// SendTransactions sends signed transactions to the network
	// Returns the transaction ID of the first transaction
	SendTransactions(ctx context.Context, signedTxns [][]byte, network string) (string, error)

	// WaitForConfirmation waits for a transaction to be confirmed
	WaitForConfirmation(ctx context.Context, txID string, network string, waitRounds uint64) error
}

// AssetInfo contains information about an Algorand Standard Asset
type AssetInfo struct {
	ID       string // ASA ID
	Symbol   string // Token symbol (e.g., "USDC")
	Decimals int    // Token decimals
}

// NetworkConfig contains network-specific configuration
type NetworkConfig struct {
	Name         string    // Network name
	CAIP2        string    // CAIP-2 identifier
	GenesisHash  string    // Base64-encoded genesis hash
	AlgodURL     string    // Default Algod API URL
	AlgodToken   string    // Default Algod API token
	DefaultAsset AssetInfo // Default stablecoin
}

// ClientConfig contains optional client configuration
type ClientConfig struct {
	AlgodURL   string // Custom Algod API URL
	AlgodToken string // Custom Algod API token
}

// DecodedTransactionInfo contains decoded transaction information for verification
type DecodedTransactionInfo struct {
	Txn    types.Transaction
	Signed *types.SignedTxn // nil if unsigned
	Raw    []byte           // Original raw bytes
}

// ToMap converts an ExactAvmPayload to a map for JSON marshaling
func (p *ExactAvmPayload) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"paymentGroup": p.PaymentGroup,
		"paymentIndex": p.PaymentIndex,
	}
}

// PayloadFromMap creates an ExactAvmPayload from a map
func PayloadFromMap(data map[string]interface{}) (*ExactAvmPayload, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload data: %w", err)
	}

	var payload ExactAvmPayload
	if err := json.Unmarshal(jsonBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if len(payload.PaymentGroup) == 0 {
		return nil, fmt.Errorf("missing paymentGroup field in payload")
	}

	if payload.PaymentIndex < 0 || payload.PaymentIndex >= len(payload.PaymentGroup) {
		return nil, fmt.Errorf("paymentIndex %d out of range [0, %d)", payload.PaymentIndex, len(payload.PaymentGroup))
	}

	return &payload, nil
}

// IsValidNetwork checks if the network is supported for Algorand
func IsValidNetwork(network string) bool {
	if _, ok := NetworkConfigs[network]; ok {
		return true
	}
	if _, ok := V1ToV2NetworkMap[network]; ok {
		return true
	}
	return false
}

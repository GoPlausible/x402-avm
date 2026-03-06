package avm

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	algocrypto "github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/types"

	x402avm "github.com/GoPlausible/x402-avm/go/mechanisms/avm"
)

// ClientSigner implements x402avm.ClientAvmSigner using an Ed25519 private key.
// This provides client-side transaction signing for creating payment payloads.
type ClientSigner struct {
	privateKey ed25519.PrivateKey
	address    string
}

// NewClientSignerFromPrivateKey creates a client signer from a base64-encoded private key.
// The private key should be 64 bytes: 32-byte seed + 32-byte public key (standard Algorand format).
//
// Args:
//
//	privateKeyBase64: Base64-encoded 64-byte Algorand private key
//
// Returns:
//
//	ClientAvmSigner implementation ready for use with avm.NewExactAvmScheme()
//	Error if private key is invalid
//
// Example:
//
//	signer, err := avm.NewClientSignerFromPrivateKey("base64EncodedKey...")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	client := x402.Newx402Client().
//	    Register("algorand:*", avm.NewExactAvmScheme(signer))
func NewClientSignerFromPrivateKey(privateKeyBase64 string) (x402avm.ClientAvmSigner, error) {
	// Decode base64 private key
	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid private key base64: %w", err)
	}

	if len(keyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d bytes, got %d", ed25519.PrivateKeySize, len(keyBytes))
	}

	privateKey := ed25519.PrivateKey(keyBytes)

	// Derive address from public key (last 32 bytes)
	account, err := algocrypto.AccountFromPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive account from private key: %w", err)
	}

	return &ClientSigner{
		privateKey: privateKey,
		address:    account.Address.String(),
	}, nil
}

// Address returns the Algorand address of the signer (58-character base32).
func (s *ClientSigner) Address() string {
	return s.address
}

// SignTransactions signs specified transactions in a group.
// indexesToSign specifies which transactions to sign (0-based indices).
// Returns signed transaction bytes for signed indices, nil for unsigned ones.
//
// Args:
//
//	ctx: Context for cancellation and timeout control
//	txns: Array of msgpack-encoded unsigned transactions
//	indexesToSign: Which indices in the group to sign
//
// Returns:
//
//	Array of signed transaction bytes (nil for unsigned indices)
//	Error if signing fails
func (s *ClientSigner) SignTransactions(ctx context.Context, txns [][]byte, indexesToSign []int) ([][]byte, error) {
	// Build a set for fast lookup
	signSet := make(map[int]bool, len(indexesToSign))
	for _, idx := range indexesToSign {
		signSet[idx] = true
	}

	results := make([][]byte, len(txns))

	for i, txnBytes := range txns {
		if !signSet[i] {
			// Not signing this transaction
			results[i] = nil
			continue
		}

		// Decode the unsigned transaction
		var txn types.Transaction
		if err := msgpack.Decode(txnBytes, &txn); err != nil {
			return nil, fmt.Errorf("failed to decode transaction %d: %w", i, err)
		}

		// Sign using the SDK's SignTransaction which handles the "TX" prefix
		_, signedBytes, err := algocrypto.SignTransaction(s.privateKey, txn)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction %d: %w", i, err)
		}

		results[i] = signedBytes
	}

	return results, nil
}

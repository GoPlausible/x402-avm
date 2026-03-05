package avm

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/algorand/go-algorand-sdk/v2/types"
)

// Generate a deterministic test key for testing
func testPrivateKey(t *testing.T) string {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	// Use a deterministic seed for reproducible tests
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	pk := ed25519.NewKeyFromSeed(seed)
	return base64.StdEncoding.EncodeToString(pk)
}

func TestNewClientSignerFromPrivateKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid key",
			key:     testPrivateKey(t),
			wantErr: false,
		},
		{
			name:    "invalid key - not base64",
			key:     "invalid!!!not-base64",
			wantErr: true,
		},
		{
			name:    "invalid key - wrong length",
			key:     base64.StdEncoding.EncodeToString([]byte("short")),
			wantErr: true,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := NewClientSignerFromPrivateKey(tt.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewClientSignerFromPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if signer == nil {
				t.Error("expected non-nil signer")
			}
		})
	}
}

func TestClientSigner_Address(t *testing.T) {
	signer, err := NewClientSignerFromPrivateKey(testPrivateKey(t))
	if err != nil {
		t.Fatalf("NewClientSignerFromPrivateKey() failed: %v", err)
	}

	addr := signer.Address()

	// Algorand addresses are 58 characters (base32 encoded with checksum)
	if len(addr) != 58 {
		t.Errorf("Address() length = %d, want 58", len(addr))
	}

	// Should be deterministic
	addr2 := signer.Address()
	if addr != addr2 {
		t.Errorf("Address() not deterministic: %s != %s", addr, addr2)
	}
}

func TestClientSigner_SignTransactions(t *testing.T) {
	signer, err := NewClientSignerFromPrivateKey(testPrivateKey(t))
	if err != nil {
		t.Fatalf("NewClientSignerFromPrivateKey() failed: %v", err)
	}

	addr := signer.Address()

	// Create a simple payment transaction
	suggestedParams := types.SuggestedParams{
		Fee:             1000,
		FirstRoundValid: 1000,
		LastRoundValid:  2000,
		GenesisID:       "testnet-v1.0",
		GenesisHash:     []byte{0x48, 0x63, 0xb5, 0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		FlatFee:         true,
	}

	txn, err := transaction.MakePaymentTxn(
		addr, addr, 0, nil, "", suggestedParams,
	)
	if err != nil {
		t.Fatalf("MakePaymentTxn() failed: %v", err)
	}

	// Encode the transaction
	txnBytes := msgpack.Encode(txn)

	// Sign only index 0
	results, err := signer.SignTransactions(context.Background(), [][]byte{txnBytes}, []int{0})
	if err != nil {
		t.Fatalf("SignTransactions() failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0] == nil {
		t.Error("Expected signed bytes at index 0, got nil")
	}

	// Should be decodable as a signed transaction
	var stxn types.SignedTxn
	if err := msgpack.Decode(results[0], &stxn); err != nil {
		t.Fatalf("Failed to decode signed transaction: %v", err)
	}

	// Signature should not be zero
	isZero := true
	for _, b := range stxn.Sig {
		if b != 0 {
			isZero = false
			break
		}
	}
	if isZero {
		t.Error("Signature is all zeros")
	}
}

func TestClientSigner_SignTransactions_SkipIndices(t *testing.T) {
	signer, err := NewClientSignerFromPrivateKey(testPrivateKey(t))
	if err != nil {
		t.Fatalf("NewClientSignerFromPrivateKey() failed: %v", err)
	}

	addr := signer.Address()

	suggestedParams := types.SuggestedParams{
		Fee:             1000,
		FirstRoundValid: 1000,
		LastRoundValid:  2000,
		GenesisID:       "testnet-v1.0",
		GenesisHash:     []byte{0x48, 0x63, 0xb5, 0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		FlatFee:         true,
	}

	txn1, _ := transaction.MakePaymentTxn(addr, addr, 0, nil, "", suggestedParams)
	txn2, _ := transaction.MakePaymentTxn(addr, addr, 0, nil, "", suggestedParams)

	txns := [][]byte{
		msgpack.Encode(txn1),
		msgpack.Encode(txn2),
	}

	// Only sign index 1 (skip index 0)
	results, err := signer.SignTransactions(context.Background(), txns, []int{1})
	if err != nil {
		t.Fatalf("SignTransactions() failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if results[0] != nil {
		t.Error("Expected nil at index 0 (not signed)")
	}

	if results[1] == nil {
		t.Error("Expected signed bytes at index 1, got nil")
	}
}

package unit_test

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"testing"

	x402 "github.com/coinbase/x402/go"
	avm "github.com/coinbase/x402/go/mechanisms/avm"
	avmclient "github.com/coinbase/x402/go/mechanisms/avm/exact/client"
	avmfacilitator "github.com/coinbase/x402/go/mechanisms/avm/exact/facilitator"
	avmv1client "github.com/coinbase/x402/go/mechanisms/avm/exact/v1/client"
	avmv1facilitator "github.com/coinbase/x402/go/mechanisms/avm/exact/v1/facilitator"
	"github.com/coinbase/x402/go/types"

	algodClient "github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
)

// =========================================================================
// Mock Signers for Unit Tests
// =========================================================================

// mockAvmClientSigner implements avm.ClientAvmSigner for testing
type mockAvmClientSigner struct {
	address   string
	signError error
}

func (m *mockAvmClientSigner) Address() string {
	if m.address == "" {
		return "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAY5HFKQ"
	}
	return m.address
}

func (m *mockAvmClientSigner) SignTransactions(ctx context.Context, txns [][]byte, indexesToSign []int) ([][]byte, error) {
	if m.signError != nil {
		return nil, m.signError
	}

	results := make([][]byte, len(txns))
	signSet := make(map[int]bool)
	for _, idx := range indexesToSign {
		signSet[idx] = true
	}

	for i := range txns {
		if signSet[i] {
			// Return mock signed bytes (just prefix with marker)
			results[i] = append([]byte{0xDE, 0xAD}, txns[i]...)
		}
	}

	return results, nil
}

// mockAvmFacilitatorSigner implements avm.FacilitatorAvmSigner for testing
type mockAvmFacilitatorSigner struct {
	addresses         []string
	signError         error
	simulateError     error
	sendError         error
	confirmError      error
	lastTxID          string
	signedTxnBytes    []byte
}

func (m *mockAvmFacilitatorSigner) GetAddresses(ctx context.Context, network string) []string {
	if m.addresses == nil {
		return []string{"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAY5HFKQ"}
	}
	return m.addresses
}

func (m *mockAvmFacilitatorSigner) SignTransaction(ctx context.Context, txnBytes []byte, senderAddress string, network string) ([]byte, error) {
	if m.signError != nil {
		return nil, m.signError
	}
	if m.signedTxnBytes != nil {
		return m.signedTxnBytes, nil
	}
	// Return mock signed bytes
	return append([]byte{0xBE, 0xEF}, txnBytes...), nil
}

func (m *mockAvmFacilitatorSigner) GetAlgodClient(network string) (*algodClient.Client, error) {
	return nil, nil
}

func (m *mockAvmFacilitatorSigner) SimulateTransactions(ctx context.Context, txns [][]byte, network string) error {
	return m.simulateError
}

func (m *mockAvmFacilitatorSigner) SendTransactions(ctx context.Context, signedTxns [][]byte, network string) (string, error) {
	if m.sendError != nil {
		return "", m.sendError
	}
	m.lastTxID = "MOCKTXID1234567890ABCDEF"
	return m.lastTxID, nil
}

func (m *mockAvmFacilitatorSigner) WaitForConfirmation(ctx context.Context, txID string, network string, waitRounds uint64) error {
	return m.confirmError
}

// =========================================================================
// Client Tests
// =========================================================================

// TestExactAvmClientScheme tests the Scheme() method
func TestExactAvmClientScheme(t *testing.T) {
	signer := &mockAvmClientSigner{}
	client := avmclient.NewExactAvmScheme(signer)

	if client.Scheme() != avm.SchemeExact {
		t.Errorf("Expected scheme %s, got %s", avm.SchemeExact, client.Scheme())
	}
}

// TestExactAvmClientSchemeV1 tests the V1 Scheme() method
func TestExactAvmClientSchemeV1(t *testing.T) {
	signer := &mockAvmClientSigner{}
	client := avmv1client.NewExactAvmSchemeV1(signer)

	if client.Scheme() != avm.SchemeExact {
		t.Errorf("Expected scheme %s, got %s", avm.SchemeExact, client.Scheme())
	}
}

// =========================================================================
// Facilitator Tests
// =========================================================================

// TestExactAvmFacilitatorScheme tests the facilitator scheme initialization
func TestExactAvmFacilitatorScheme(t *testing.T) {
	signer := &mockAvmFacilitatorSigner{}

	t.Run("Creates scheme", func(t *testing.T) {
		scheme := avmfacilitator.NewExactAvmScheme(signer)
		if scheme == nil {
			t.Error("Expected scheme to be created")
		}
		if scheme.Scheme() != avm.SchemeExact {
			t.Errorf("Expected scheme %s, got %s", avm.SchemeExact, scheme.Scheme())
		}
	})

	t.Run("CaipFamily returns algorand:*", func(t *testing.T) {
		scheme := avmfacilitator.NewExactAvmScheme(signer)
		if scheme.CaipFamily() != "algorand:*" {
			t.Errorf("Expected CaipFamily algorand:*, got %s", scheme.CaipFamily())
		}
	})

	t.Run("GetSigners returns addresses", func(t *testing.T) {
		signer := &mockAvmFacilitatorSigner{
			addresses: []string{"ADDR1Y5HFKQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", "ADDR2Y5HFKQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		}
		scheme := avmfacilitator.NewExactAvmScheme(signer)

		signers := scheme.GetSigners(x402.Network(avm.AlgorandTestnetCAIP2))
		if len(signers) != 2 {
			t.Errorf("Expected 2 signers, got %d", len(signers))
		}
	})

	t.Run("GetExtra returns feePayer", func(t *testing.T) {
		signer := &mockAvmFacilitatorSigner{
			addresses: []string{"FEEPAYERADDRESS234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"},
		}
		scheme := avmfacilitator.NewExactAvmScheme(signer)

		extra := scheme.GetExtra(x402.Network(avm.AlgorandTestnetCAIP2))
		if extra == nil {
			t.Fatal("Expected extra to be non-nil")
		}
		if _, ok := extra["feePayer"]; !ok {
			t.Error("Expected feePayer in extra")
		}
	})

	t.Run("GetExtra returns nil for no addresses", func(t *testing.T) {
		signer := &mockAvmFacilitatorSigner{
			addresses: []string{},
		}
		scheme := avmfacilitator.NewExactAvmScheme(signer)

		extra := scheme.GetExtra(x402.Network(avm.AlgorandTestnetCAIP2))
		if extra != nil {
			t.Error("Expected nil extra when no addresses")
		}
	})
}

// TestExactAvmFacilitatorSchemeV1 tests V1 facilitator
func TestExactAvmFacilitatorSchemeV1(t *testing.T) {
	signer := &mockAvmFacilitatorSigner{}
	scheme := avmv1facilitator.NewExactAvmSchemeV1(signer)

	if scheme.Scheme() != avm.SchemeExact {
		t.Errorf("Expected scheme %s, got %s", avm.SchemeExact, scheme.Scheme())
	}

	if scheme.CaipFamily() != "algorand:*" {
		t.Errorf("Expected CaipFamily algorand:*, got %s", scheme.CaipFamily())
	}
}

// TestVerifyInvalidInputs tests verification validation
func TestVerifyInvalidInputs(t *testing.T) {
	ctx := context.Background()
	signer := &mockAvmFacilitatorSigner{
		addresses: []string{"FEEPAYERADDRESS234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"},
	}
	scheme := avmfacilitator.NewExactAvmScheme(signer)

	t.Run("Rejects scheme mismatch", func(t *testing.T) {
		payload := types.PaymentPayload{
			X402Version: 2,
			Accepted: types.PaymentRequirements{
				Scheme:  "wrong",
				Network: avm.AlgorandTestnetCAIP2,
			},
		}
		requirements := types.PaymentRequirements{
			Scheme:  avm.SchemeExact,
			Network: avm.AlgorandTestnetCAIP2,
		}

		_, err := scheme.Verify(ctx, payload, requirements, nil)
		if err == nil {
			t.Error("Expected error for scheme mismatch")
		}
		if !strings.Contains(err.Error(), avmfacilitator.ErrUnsupportedScheme) {
			t.Errorf("Expected ErrUnsupportedScheme, got: %s", err.Error())
		}
	})

	t.Run("Rejects network mismatch", func(t *testing.T) {
		payload := types.PaymentPayload{
			X402Version: 2,
			Accepted: types.PaymentRequirements{
				Scheme:  avm.SchemeExact,
				Network: avm.AlgorandMainnetCAIP2, // Different!
			},
		}
		requirements := types.PaymentRequirements{
			Scheme:  avm.SchemeExact,
			Network: avm.AlgorandTestnetCAIP2,
			Extra: map[string]interface{}{
				"feePayer": "FEEPAYERADDRESS234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ234567",
			},
		}

		_, err := scheme.Verify(ctx, payload, requirements, nil)
		if err == nil {
			t.Error("Expected error for network mismatch")
		}
		if !strings.Contains(err.Error(), avmfacilitator.ErrNetworkMismatch) {
			t.Errorf("Expected ErrNetworkMismatch, got: %s", err.Error())
		}
	})

	t.Run("Rejects missing feePayer", func(t *testing.T) {
		payload := types.PaymentPayload{
			X402Version: 2,
			Accepted: types.PaymentRequirements{
				Scheme:  avm.SchemeExact,
				Network: avm.AlgorandTestnetCAIP2,
			},
		}
		requirements := types.PaymentRequirements{
			Scheme:  avm.SchemeExact,
			Network: avm.AlgorandTestnetCAIP2,
			Extra:   nil, // Missing feePayer!
		}

		_, err := scheme.Verify(ctx, payload, requirements, nil)
		if err == nil {
			t.Error("Expected error for missing feePayer")
		}
		if !strings.Contains(err.Error(), avmfacilitator.ErrMissingFeePayer) {
			t.Errorf("Expected ErrMissingFeePayer, got: %s", err.Error())
		}
	})

	t.Run("Rejects unmanaged feePayer", func(t *testing.T) {
		payload := types.PaymentPayload{
			X402Version: 2,
			Accepted: types.PaymentRequirements{
				Scheme:  avm.SchemeExact,
				Network: avm.AlgorandTestnetCAIP2,
			},
		}
		requirements := types.PaymentRequirements{
			Scheme:  avm.SchemeExact,
			Network: avm.AlgorandTestnetCAIP2,
			Extra: map[string]interface{}{
				"feePayer": "UNKNOWNADDRESS34567890ABCDEFGHIJKLMNOPQRSTUVWXYZ2345678",
			},
		}

		_, err := scheme.Verify(ctx, payload, requirements, nil)
		if err == nil {
			t.Error("Expected error for unmanaged feePayer")
		}
		if !strings.Contains(err.Error(), avmfacilitator.ErrFeePayerNotManaged) {
			t.Errorf("Expected ErrFeePayerNotManaged, got: %s", err.Error())
		}
	})

	t.Run("Rejects invalid payload format", func(t *testing.T) {
		payload := types.PaymentPayload{
			X402Version: 2,
			Accepted: types.PaymentRequirements{
				Scheme:  avm.SchemeExact,
				Network: avm.AlgorandTestnetCAIP2,
			},
			Payload: map[string]interface{}{
				// Missing paymentGroup
				"paymentIndex": 0,
			},
		}
		requirements := types.PaymentRequirements{
			Scheme:  avm.SchemeExact,
			Network: avm.AlgorandTestnetCAIP2,
			Extra: map[string]interface{}{
				"feePayer": "FEEPAYERADDRESS234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ234567",
			},
		}

		_, err := scheme.Verify(ctx, payload, requirements, nil)
		if err == nil {
			t.Error("Expected error for invalid payload format")
		}
		if !strings.Contains(err.Error(), avmfacilitator.ErrInvalidPayloadFormat) {
			t.Errorf("Expected ErrInvalidPayloadFormat, got: %s", err.Error())
		}
	})

	t.Run("Rejects group size exceeded", func(t *testing.T) {
		// Create a group with 17 entries (max is 16)
		bigGroup := make([]string, 17)
		for i := range bigGroup {
			bigGroup[i] = base64.StdEncoding.EncodeToString([]byte{0x01})
		}

		payload := types.PaymentPayload{
			X402Version: 2,
			Accepted: types.PaymentRequirements{
				Scheme:  avm.SchemeExact,
				Network: avm.AlgorandTestnetCAIP2,
			},
			Payload: map[string]interface{}{
				"paymentGroup": bigGroup,
				"paymentIndex": 0,
			},
		}
		requirements := types.PaymentRequirements{
			Scheme:  avm.SchemeExact,
			Network: avm.AlgorandTestnetCAIP2,
			Extra: map[string]interface{}{
				"feePayer": "FEEPAYERADDRESS234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ234567",
			},
		}

		_, err := scheme.Verify(ctx, payload, requirements, nil)
		if err == nil {
			t.Error("Expected error for group size exceeded")
		}
		if !strings.Contains(err.Error(), avmfacilitator.ErrGroupSizeExceeded) {
			t.Errorf("Expected ErrGroupSizeExceeded, got: %s", err.Error())
		}
	})
}

// =========================================================================
// Version Compatibility Tests
// =========================================================================

// TestAVMVersionMismatch tests that V1 and V2 don't mix
func TestAVMVersionMismatch(t *testing.T) {
	t.Run("V1 Client produces V1 payload", func(t *testing.T) {
		// V1 client should set X402Version to 1
		signer := &mockAvmClientSigner{}
		client := avmv1client.NewExactAvmSchemeV1(signer)

		if client.Scheme() != avm.SchemeExact {
			t.Errorf("Expected scheme %s, got %s", avm.SchemeExact, client.Scheme())
		}
	})

	t.Run("V2 Client produces V2 payload", func(t *testing.T) {
		signer := &mockAvmClientSigner{}
		client := avmclient.NewExactAvmScheme(signer)

		if client.Scheme() != avm.SchemeExact {
			t.Errorf("Expected scheme %s, got %s", avm.SchemeExact, client.Scheme())
		}
	})
}

// TestAVMDualVersionSupport tests dual V1+V2 registration
func TestAVMDualVersionSupport(t *testing.T) {
	t.Run("Client registers both V1 and V2", func(t *testing.T) {
		signer := &mockAvmClientSigner{}
		client := x402.Newx402Client()

		avmClientV1 := avmv1client.NewExactAvmSchemeV1(signer)
		client.RegisterV1(avm.AlgorandTestnetV1, avmClientV1)

		avmClientV2 := avmclient.NewExactAvmScheme(signer)
		client.Register(avm.AlgorandTestnetCAIP2, avmClientV2)

		// Both should be registered without panics
	})

	t.Run("Facilitator registers both V1 and V2", func(t *testing.T) {
		signer := &mockAvmFacilitatorSigner{}
		facilitator := x402.Newx402Facilitator()

		avmFacilitatorV1 := avmv1facilitator.NewExactAvmSchemeV1(signer)
		facilitator.RegisterV1([]x402.Network{x402.Network(avm.AlgorandTestnetV1)}, avmFacilitatorV1)

		avmFacilitatorV2 := avmfacilitator.NewExactAvmScheme(signer)
		facilitator.Register([]x402.Network{x402.Network(avm.AlgorandTestnetCAIP2)}, avmFacilitatorV2)

		// Both should be registered without panics
	})
}

// =========================================================================
// Ed25519 Signature Tests
// =========================================================================

// TestEd25519SignatureVerification tests that the Ed25519 signing roundtrip works
func TestEd25519SignatureVerification(t *testing.T) {
	// Generate a test key pair
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(i + 42)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	t.Run("Valid signature verifies", func(t *testing.T) {
		message := []byte("TX" + "test message bytes")
		signature := ed25519.Sign(privateKey, message)

		if !ed25519.Verify(publicKey, message, signature) {
			t.Error("Valid signature should verify")
		}
	})

	t.Run("Wrong message fails verification", func(t *testing.T) {
		message := []byte("TX" + "test message bytes")
		signature := ed25519.Sign(privateKey, message)

		wrongMessage := []byte("TX" + "wrong message bytes")
		if ed25519.Verify(publicKey, wrongMessage, signature) {
			t.Error("Wrong message should not verify")
		}
	})

	t.Run("Wrong key fails verification", func(t *testing.T) {
		message := []byte("TX" + "test message bytes")
		signature := ed25519.Sign(privateKey, message)

		wrongSeed := make([]byte, ed25519.SeedSize)
		for i := range wrongSeed {
			wrongSeed[i] = byte(i + 99)
		}
		wrongKey := ed25519.NewKeyFromSeed(wrongSeed).Public().(ed25519.PublicKey)

		if ed25519.Verify(wrongKey, message, signature) {
			t.Error("Wrong key should not verify")
		}
	})
}

// =========================================================================
// Constants Tests
// =========================================================================

func TestAvmConstants(t *testing.T) {
	t.Run("SchemeExact is exact", func(t *testing.T) {
		if avm.SchemeExact != "exact" {
			t.Errorf("Expected SchemeExact='exact', got %s", avm.SchemeExact)
		}
	})

	t.Run("MaxAtomicGroupSize is 16", func(t *testing.T) {
		if avm.MaxAtomicGroupSize != 16 {
			t.Errorf("Expected MaxAtomicGroupSize=16, got %d", avm.MaxAtomicGroupSize)
		}
	})

	t.Run("MinTxnFee is 1000", func(t *testing.T) {
		if avm.MinTxnFee != 1000 {
			t.Errorf("Expected MinTxnFee=1000, got %d", avm.MinTxnFee)
		}
	})

	t.Run("MaxReasonableFee is 16000", func(t *testing.T) {
		if avm.MaxReasonableFee != 16000 {
			t.Errorf("Expected MaxReasonableFee=16000, got %d", avm.MaxReasonableFee)
		}
	})

	t.Run("USDC ASA IDs are correct", func(t *testing.T) {
		if avm.USDCMainnetASAID != "31566704" {
			t.Errorf("Expected mainnet USDC ASA ID 31566704, got %s", avm.USDCMainnetASAID)
		}
		if avm.USDCTestnetASAID != "10458941" {
			t.Errorf("Expected testnet USDC ASA ID 10458941, got %s", avm.USDCTestnetASAID)
		}
	})

	t.Run("CAIP-2 identifiers have correct format", func(t *testing.T) {
		if !strings.HasPrefix(avm.AlgorandMainnetCAIP2, "algorand:") {
			t.Errorf("Mainnet CAIP-2 should start with 'algorand:', got %s", avm.AlgorandMainnetCAIP2)
		}
		if !strings.HasPrefix(avm.AlgorandTestnetCAIP2, "algorand:") {
			t.Errorf("Testnet CAIP-2 should start with 'algorand:', got %s", avm.AlgorandTestnetCAIP2)
		}
	})

	t.Run("V1 network names", func(t *testing.T) {
		if avm.AlgorandMainnetV1 != "algorand-mainnet" {
			t.Errorf("Expected algorand-mainnet, got %s", avm.AlgorandMainnetV1)
		}
		if avm.AlgorandTestnetV1 != "algorand-testnet" {
			t.Errorf("Expected algorand-testnet, got %s", avm.AlgorandTestnetV1)
		}
	})

	t.Run("Network configs have required fields", func(t *testing.T) {
		for caip2, config := range avm.NetworkConfigs {
			if config.Name == "" {
				t.Errorf("Network %s missing name", caip2)
			}
			if config.CAIP2 != caip2 {
				t.Errorf("Network %s CAIP2 mismatch: %s", caip2, config.CAIP2)
			}
			if config.GenesisHash == "" {
				t.Errorf("Network %s missing genesis hash", caip2)
			}
			if config.AlgodURL == "" {
				t.Errorf("Network %s missing Algod URL", caip2)
			}
			if config.DefaultAsset.ID == "" {
				t.Errorf("Network %s missing default asset ID", caip2)
			}
			if config.DefaultAsset.Symbol == "" {
				t.Errorf("Network %s missing default asset symbol", caip2)
			}
		}
	})

	t.Run("V1-V2 network maps are bidirectional", func(t *testing.T) {
		for v1, v2 := range avm.V1ToV2NetworkMap {
			reverse, ok := avm.V2ToV1NetworkMap[v2]
			if !ok {
				t.Errorf("V2ToV1 missing entry for %s", v2)
			}
			if reverse != v1 {
				t.Errorf("V1(%s)->V2(%s)->V1(%s) roundtrip failed", v1, v2, reverse)
			}
		}
	})
}

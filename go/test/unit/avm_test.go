// Package unit_test contains unit tests for the AVM mechanism
package unit_test

import (
	"testing"

	x402 "github.com/coinbase/x402/go"
	avm "github.com/coinbase/x402/go/mechanisms/avm"
	avmserver "github.com/coinbase/x402/go/mechanisms/avm/exact/server"
)

// TestAlgorandServerPriceParsing tests V2 server price parsing
func TestAlgorandServerPriceParsing(t *testing.T) {
	server := avmserver.NewExactAvmScheme()
	network := x402.Network(avm.AlgorandTestnetCAIP2)

	tests := []struct {
		name          string
		price         x402.Price
		expectedAsset string
		shouldError   bool
	}{
		{
			name:          "Simple decimal",
			price:         "0.10",
			expectedAsset: avm.USDCTestnetASAID,
			shouldError:   false,
		},
		{
			name:          "Dollar sign",
			price:         "$0.10",
			expectedAsset: avm.USDCTestnetASAID,
			shouldError:   false,
		},
		{
			name:          "With currency",
			price:         "0.10 USDC",
			expectedAsset: avm.USDCTestnetASAID,
			shouldError:   false,
		},
		{
			name:          "Float",
			price:         float64(0.10),
			expectedAsset: avm.USDCTestnetASAID,
			shouldError:   false,
		},
		{
			name:          "Integer",
			price:         1,
			expectedAsset: avm.USDCTestnetASAID,
			shouldError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.ParsePrice(tt.price, network)
			if tt.shouldError && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !tt.shouldError {
				if result.Asset != tt.expectedAsset {
					t.Errorf("Expected asset %s, got %s", tt.expectedAsset, result.Asset)
				}
				if result.Amount == "" {
					t.Error("Expected non-empty amount")
				}
			}
		})
	}
}

// TestAlgorandUtilities tests utility functions
func TestAlgorandUtilities(t *testing.T) {
	t.Run("NormalizeNetwork", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
			isError  bool
		}{
			{avm.AlgorandMainnetV1, avm.AlgorandMainnetCAIP2, false},
			{avm.AlgorandTestnetV1, avm.AlgorandTestnetCAIP2, false},
			{avm.AlgorandMainnetCAIP2, avm.AlgorandMainnetCAIP2, false},
			{avm.AlgorandTestnetCAIP2, avm.AlgorandTestnetCAIP2, false},
			{"invalid", "", true},
		}

		for _, tt := range tests {
			result, err := avm.NormalizeNetwork(tt.input)
			if tt.isError && err == nil {
				t.Errorf("Expected error for input %s", tt.input)
			}
			if !tt.isError && err != nil {
				t.Errorf("Unexpected error for input %s: %v", tt.input, err)
			}
			if !tt.isError && result != tt.expected {
				t.Errorf("For input %s, expected %s, got %s", tt.input, tt.expected, result)
			}
		}
	})

	t.Run("ValidateAlgorandAddress", func(t *testing.T) {
		validAddresses := []string{
			"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAY5HFKQ", // Zero address
		}

		invalidAddresses := []string{
			"",
			"invalid",
			"0x1234567890123456789012345678901234567890", // EVM address
			"123",
			"short",
		}

		for _, addr := range validAddresses {
			if !avm.ValidateAlgorandAddress(addr) {
				t.Errorf("Expected %s to be valid", addr)
			}
		}

		for _, addr := range invalidAddresses {
			if avm.ValidateAlgorandAddress(addr) {
				t.Errorf("Expected %s to be invalid", addr)
			}
		}
	})

	t.Run("ParseAmount", func(t *testing.T) {
		tests := []struct {
			amount   string
			decimals int
			expected uint64
		}{
			{"1", 6, 1000000},
			{"0.1", 6, 100000},
			{"0.01", 6, 10000},
			{"1.5", 6, 1500000},
			{"100", 6, 100000000},
			{"0.000001", 6, 1},
		}

		for _, tt := range tests {
			result, err := avm.ParseAmount(tt.amount, tt.decimals)
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.amount, err)
			}
			if result != tt.expected {
				t.Errorf("For %s with %d decimals, expected %d, got %d", tt.amount, tt.decimals, tt.expected, result)
			}
		}
	})

	t.Run("FormatAmount", func(t *testing.T) {
		tests := []struct {
			amount   uint64
			decimals int
			expected string
		}{
			{1000000, 6, "1"},
			{100000, 6, "0.1"},
			{10000, 6, "0.01"},
			{1500000, 6, "1.5"},
			{100000000, 6, "100"},
			{0, 6, "0"},
			{1, 6, "0.000001"},
		}

		for _, tt := range tests {
			result := avm.FormatAmount(tt.amount, tt.decimals)
			if result != tt.expected {
				t.Errorf("For %d with %d decimals, expected %s, got %s", tt.amount, tt.decimals, tt.expected, result)
			}
		}
	})
}

// TestAlgorandIsValidNetwork tests network validation
func TestAlgorandIsValidNetwork(t *testing.T) {
	validNetworks := []string{
		avm.AlgorandMainnetCAIP2,
		avm.AlgorandTestnetCAIP2,
		avm.AlgorandMainnetV1,
		avm.AlgorandTestnetV1,
	}

	invalidNetworks := []string{
		"ethereum",
		"solana",
		"invalid:network",
		"",
	}

	for _, network := range validNetworks {
		if !avm.IsValidNetwork(network) {
			t.Errorf("Expected %s to be valid", network)
		}
	}

	for _, network := range invalidNetworks {
		if avm.IsValidNetwork(network) {
			t.Errorf("Expected %s to be invalid", network)
		}
	}
}

// TestAlgorandGetNetworkConfig tests network config retrieval
func TestAlgorandGetNetworkConfig(t *testing.T) {
	tests := []struct {
		input        string
		expectedCAIP string
		shouldError  bool
	}{
		{avm.AlgorandMainnetV1, avm.AlgorandMainnetCAIP2, false},
		{avm.AlgorandMainnetCAIP2, avm.AlgorandMainnetCAIP2, false},
		{avm.AlgorandTestnetV1, avm.AlgorandTestnetCAIP2, false},
		{avm.AlgorandTestnetCAIP2, avm.AlgorandTestnetCAIP2, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		config, err := avm.GetNetworkConfig(tt.input)
		if tt.shouldError && err == nil {
			t.Errorf("Expected error for %s", tt.input)
		}
		if !tt.shouldError && err != nil {
			t.Errorf("Unexpected error for %s: %v", tt.input, err)
		}
		if !tt.shouldError {
			if config.CAIP2 != tt.expectedCAIP {
				t.Errorf("Expected CAIP2 %s, got %s", tt.expectedCAIP, config.CAIP2)
			}
		}
	}
}

// TestAlgorandGetAssetInfo tests asset info retrieval
func TestAlgorandGetAssetInfo(t *testing.T) {
	t.Run("Default asset by ID", func(t *testing.T) {
		info, err := avm.GetAssetInfo(avm.AlgorandTestnetCAIP2, avm.USDCTestnetASAID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if info.ID != avm.USDCTestnetASAID {
			t.Errorf("Expected ID %s, got %s", avm.USDCTestnetASAID, info.ID)
		}
		if info.Decimals != 6 {
			t.Errorf("Expected decimals 6, got %d", info.Decimals)
		}
	})

	t.Run("Numeric ASA ID", func(t *testing.T) {
		info, err := avm.GetAssetInfo(avm.AlgorandTestnetCAIP2, "12345")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if info.ID != "12345" {
			t.Errorf("Expected ID 12345, got %s", info.ID)
		}
	})

	t.Run("Default fallback", func(t *testing.T) {
		info, err := avm.GetAssetInfo(avm.AlgorandTestnetCAIP2, "unknown")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if info.ID != avm.USDCTestnetASAID {
			t.Errorf("Expected default asset %s, got %s", avm.USDCTestnetASAID, info.ID)
		}
	})
}

// TestAlgorandGetGenesisHash tests genesis hash retrieval
func TestAlgorandGetGenesisHash(t *testing.T) {
	tests := []struct {
		network  string
		expected string
		isError  bool
	}{
		{avm.AlgorandMainnetCAIP2, avm.AlgorandMainnetGenesisHash, false},
		{avm.AlgorandTestnetCAIP2, avm.AlgorandTestnetGenesisHash, false},
		{avm.AlgorandMainnetV1, avm.AlgorandMainnetGenesisHash, false},
		{avm.AlgorandTestnetV1, avm.AlgorandTestnetGenesisHash, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		result, err := avm.GetGenesisHashForNetwork(tt.network)
		if tt.isError && err == nil {
			t.Errorf("Expected error for %s", tt.network)
		}
		if !tt.isError && err != nil {
			t.Errorf("Unexpected error for %s: %v", tt.network, err)
		}
		if !tt.isError && result != tt.expected {
			t.Errorf("For %s, expected %s, got %s", tt.network, tt.expected, result)
		}
	}
}

// TestAlgorandIsAlgorandNetwork tests IsAlgorandNetwork
func TestAlgorandIsAlgorandNetwork(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{avm.AlgorandMainnetCAIP2, true},
		{avm.AlgorandTestnetCAIP2, true},
		{"algorand:custom", true},
		{"algorand-mainnet", true},
		{"algorand-testnet", true},
		{"eip155:84532", false},
		{"solana:mainnet", false},
		{"", false},
	}

	for _, tt := range tests {
		result := avm.IsAlgorandNetwork(tt.input)
		if result != tt.expected {
			t.Errorf("IsAlgorandNetwork(%s) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

// TestAlgorandIsTestnetNetwork tests IsTestnetNetwork
func TestAlgorandIsTestnetNetwork(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{avm.AlgorandTestnetCAIP2, true},
		{avm.AlgorandTestnetV1, true},
		{avm.AlgorandMainnetCAIP2, false},
		{avm.AlgorandMainnetV1, false},
		{"invalid", false},
	}

	for _, tt := range tests {
		result := avm.IsTestnetNetwork(tt.input)
		if result != tt.expected {
			t.Errorf("IsTestnetNetwork(%s) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

// TestAlgorandPayloadFromMap tests payload serialization/deserialization
func TestAlgorandPayloadFromMap(t *testing.T) {
	t.Run("Valid payload", func(t *testing.T) {
		original := &avm.ExactAvmPayload{
			PaymentGroup: []string{"dHhuMQ==", "dHhuMg=="},
			PaymentIndex: 1,
		}

		m := original.ToMap()
		decoded, err := avm.PayloadFromMap(m)
		if err != nil {
			t.Fatalf("PayloadFromMap() error: %v", err)
		}

		if len(decoded.PaymentGroup) != 2 {
			t.Errorf("Expected 2 items in group, got %d", len(decoded.PaymentGroup))
		}
		if decoded.PaymentIndex != 1 {
			t.Errorf("Expected paymentIndex 1, got %d", decoded.PaymentIndex)
		}
	})

	t.Run("Empty group", func(t *testing.T) {
		m := map[string]interface{}{
			"paymentGroup": []string{},
			"paymentIndex": 0,
		}
		_, err := avm.PayloadFromMap(m)
		if err == nil {
			t.Error("Expected error for empty paymentGroup")
		}
	})

	t.Run("Invalid payment index", func(t *testing.T) {
		m := map[string]interface{}{
			"paymentGroup": []string{"dHhuMQ=="},
			"paymentIndex": 5,
		}
		_, err := avm.PayloadFromMap(m)
		if err == nil {
			t.Error("Expected error for out-of-range paymentIndex")
		}
	})

	t.Run("Missing paymentGroup", func(t *testing.T) {
		m := map[string]interface{}{
			"paymentIndex": 0,
		}
		_, err := avm.PayloadFromMap(m)
		if err == nil {
			t.Error("Expected error for missing paymentGroup")
		}
	})
}

// TestAlgorandTransactionEncoding tests transaction encoding/decoding roundtrip
func TestAlgorandTransactionEncoding(t *testing.T) {
	t.Run("EncodeSignedTransactionBytes roundtrip", func(t *testing.T) {
		original := []byte{0x01, 0x02, 0x03, 0x04}
		encoded := avm.EncodeSignedTransactionBytes(original)
		decoded, err := avm.DecodeRawTransaction(encoded)
		if err != nil {
			t.Fatalf("DecodeRawTransaction() error: %v", err)
		}
		if len(decoded) != len(original) {
			t.Errorf("Expected %d bytes, got %d", len(original), len(decoded))
		}
		for i := range original {
			if decoded[i] != original[i] {
				t.Errorf("Byte %d mismatch: expected %d, got %d", i, original[i], decoded[i])
			}
		}
	})

	t.Run("DecodeRawTransaction invalid base64", func(t *testing.T) {
		_, err := avm.DecodeRawTransaction("not valid base64!!!")
		if err == nil {
			t.Error("Expected error for invalid base64")
		}
	})
}

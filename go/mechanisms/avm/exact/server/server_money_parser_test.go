package server

import (
	"fmt"
	"testing"

	x402 "github.com/coinbase/x402/go"
	"github.com/coinbase/x402/go/mechanisms/avm"
)

// TestRegisterMoneyParser_SingleCustomParser tests a single custom money parser
func TestRegisterMoneyParser_SingleCustomParser(t *testing.T) {
	server := NewExactAvmScheme()

	// Register custom parser: large amounts use custom token
	server.RegisterMoneyParser(func(amount float64, network x402.Network) (*x402.AssetAmount, error) {
		if amount > 100 {
			return &x402.AssetAmount{
				Amount: fmt.Sprintf("%.0f", amount*1e6),
				Asset:  "99999999", // Custom ASA ID
				Extra: map[string]interface{}{
					"token": "CUSTOM",
					"tier":  "large",
				},
			}, nil
		}
		return nil, nil // Use default for small amounts
	})

	// Test large amount - should use custom parser
	result1, err := server.ParsePrice(150.0, x402.Network(avm.AlgorandTestnetCAIP2))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedAmount1 := fmt.Sprintf("%.0f", 150*1e6)
	if result1.Amount != expectedAmount1 {
		t.Errorf("Expected amount %s, got %s", expectedAmount1, result1.Amount)
	}

	if result1.Asset != "99999999" {
		t.Errorf("Expected custom ASA, got %s", result1.Asset)
	}

	if result1.Extra["token"] != "CUSTOM" {
		t.Errorf("Expected token='CUSTOM', got %v", result1.Extra["token"])
	}

	// Test small amount - should fall back to default (USDC)
	result2, err := server.ParsePrice(50.0, x402.Network(avm.AlgorandTestnetCAIP2))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedAmount2 := "50000000" // 50 * 1e6 (USDC has 6 decimals)
	if result2.Amount != expectedAmount2 {
		t.Errorf("Expected amount %s, got %s", expectedAmount2, result2.Amount)
	}

	if result2.Asset != avm.USDCTestnetASAID {
		t.Errorf("Expected USDC ASA ID, got %s", result2.Asset)
	}
}

// TestRegisterMoneyParser_MultipleInChain tests multiple money parsers in chain
func TestRegisterMoneyParser_MultipleInChain(t *testing.T) {
	server := NewExactAvmScheme()

	// Parser 1: Premium tier (> 1000)
	server.RegisterMoneyParser(func(amount float64, network x402.Network) (*x402.AssetAmount, error) {
		if amount > 1000 {
			return &x402.AssetAmount{
				Amount: fmt.Sprintf("%.0f", amount*1e6),
				Asset:  "88888888",
				Extra:  map[string]interface{}{"tier": "premium"},
			}, nil
		}
		return nil, nil
	})

	// Parser 2: Large tier (> 100)
	server.RegisterMoneyParser(func(amount float64, network x402.Network) (*x402.AssetAmount, error) {
		if amount > 100 {
			return &x402.AssetAmount{
				Amount: fmt.Sprintf("%.0f", amount*1e6),
				Asset:  "77777777",
				Extra:  map[string]interface{}{"tier": "large"},
			}, nil
		}
		return nil, nil
	})

	network := x402.Network(avm.AlgorandTestnetCAIP2)

	// Test premium tier (first parser)
	result1, err := server.ParsePrice(2000.0, network)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result1.Extra["tier"] != "premium" {
		t.Errorf("Expected tier='premium', got %v", result1.Extra["tier"])
	}

	// Test large tier (second parser)
	result2, err := server.ParsePrice(200.0, network)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result2.Extra["tier"] != "large" {
		t.Errorf("Expected tier='large', got %v", result2.Extra["tier"])
	}

	// Test default (no parser matches)
	result3, err := server.ParsePrice(50.0, network)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result3.Asset != avm.USDCTestnetASAID {
		t.Errorf("Expected USDC, got %s", result3.Asset)
	}
}

// TestRegisterMoneyParser_StringPrices tests parsing with string prices
func TestRegisterMoneyParser_StringPrices(t *testing.T) {
	server := NewExactAvmScheme()

	server.RegisterMoneyParser(func(amount float64, network x402.Network) (*x402.AssetAmount, error) {
		if amount > 50 {
			return &x402.AssetAmount{
				Amount: fmt.Sprintf("%.0f", amount*1e6),
				Asset:  "99999999",
			}, nil
		}
		return nil, nil
	})

	network := x402.Network(avm.AlgorandTestnetCAIP2)

	tests := []struct {
		name          string
		price         string
		expectedAsset string
	}{
		{"Dollar format", "$100", "99999999"},                // > 50
		{"Plain decimal", "25.50", avm.USDCTestnetASAID},     // <= 50 (USDC)
		{"Large amount", "75", "99999999"},                   // > 50
		{"Small amount", "10", avm.USDCTestnetASAID},         // <= 50
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.ParsePrice(tt.price, network)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if result.Asset != tt.expectedAsset {
				t.Errorf("Expected asset %s, got %s", tt.expectedAsset, result.Asset)
			}
		})
	}
}

// TestRegisterMoneyParser_Chainability tests that RegisterMoneyParser returns the service for chaining
func TestRegisterMoneyParser_Chainability(t *testing.T) {
	server := NewExactAvmScheme()

	result := server.
		RegisterMoneyParser(func(amount float64, network x402.Network) (*x402.AssetAmount, error) {
			return nil, nil
		}).
		RegisterMoneyParser(func(amount float64, network x402.Network) (*x402.AssetAmount, error) {
			return nil, nil
		})

	if result != server {
		t.Error("Expected RegisterMoneyParser to return server for chaining")
	}
}

// TestRegisterMoneyParser_NoCustomParsers tests default behavior with no custom parsers
func TestRegisterMoneyParser_NoCustomParsers(t *testing.T) {
	server := NewExactAvmScheme()

	result, err := server.ParsePrice(10.0, x402.Network(avm.AlgorandTestnetCAIP2))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Asset != avm.USDCTestnetASAID {
		t.Errorf("Expected default USDC, got %s", result.Asset)
	}

	expectedAmount := "10000000" // 10 * 1e6
	if result.Amount != expectedAmount {
		t.Errorf("Expected amount %s, got %s", expectedAmount, result.Amount)
	}
}

// TestParsePrice_PreParsedObject tests pre-parsed price objects
func TestParsePrice_PreParsedObject(t *testing.T) {
	server := NewExactAvmScheme()
	network := x402.Network(avm.AlgorandTestnetCAIP2)

	priceObj := map[string]interface{}{
		"amount": "500000",
		"asset":  "12345678",
	}

	result, err := server.ParsePrice(priceObj, network)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Amount != "500000" {
		t.Errorf("Expected amount 500000, got %s", result.Amount)
	}

	if result.Asset != "12345678" {
		t.Errorf("Expected asset 12345678, got %s", result.Asset)
	}
}

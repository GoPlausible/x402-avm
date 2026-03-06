package server

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	x402 "github.com/GoPlausible/x402-avm/go"
	"github.com/GoPlausible/x402-avm/go/mechanisms/avm"
	"github.com/GoPlausible/x402-avm/go/types"
)

// ExactAvmScheme implements the SchemeNetworkServer interface for AVM (Algorand) exact payments (V2)
type ExactAvmScheme struct {
	moneyParsers []x402.MoneyParser
}

// NewExactAvmScheme creates a new ExactAvmScheme
func NewExactAvmScheme() *ExactAvmScheme {
	return &ExactAvmScheme{
		moneyParsers: []x402.MoneyParser{},
	}
}

// Scheme returns the scheme identifier
func (s *ExactAvmScheme) Scheme() string {
	return avm.SchemeExact
}

// RegisterMoneyParser registers a custom money parser in the parser chain.
func (s *ExactAvmScheme) RegisterMoneyParser(parser x402.MoneyParser) *ExactAvmScheme {
	s.moneyParsers = append(s.moneyParsers, parser)
	return s
}

// ParsePrice parses a price and converts it to an asset amount (V2)
func (s *ExactAvmScheme) ParsePrice(price x402.Price, network x402.Network) (x402.AssetAmount, error) {
	networkStr := string(network)

	config, err := avm.GetNetworkConfig(networkStr)
	if err != nil {
		return x402.AssetAmount{}, err
	}

	// Handle pre-parsed price object (with amount and asset)
	if priceMap, ok := price.(map[string]interface{}); ok {
		if amountVal, hasAmount := priceMap["amount"]; hasAmount {
			amountStr, ok := amountVal.(string)
			if !ok {
				return x402.AssetAmount{}, errors.New(ErrAmountMustBeString)
			}

			asset := config.DefaultAsset.ID
			if assetVal, hasAsset := priceMap["asset"]; hasAsset {
				if assetStr, ok := assetVal.(string); ok {
					asset = assetStr
				}
			}

			extra := make(map[string]interface{})
			if extraVal, hasExtra := priceMap["extra"]; hasExtra {
				if extraMap, ok := extraVal.(map[string]interface{}); ok {
					extra = extraMap
				}
			}

			return x402.AssetAmount{
				Amount: amountStr,
				Asset:  asset,
				Extra:  extra,
			}, nil
		}
	}

	// Parse Money to decimal number
	decimalAmount, err := s.parseMoneyToDecimal(price)
	if err != nil {
		return x402.AssetAmount{}, err
	}

	// Try each custom money parser in order
	for _, parser := range s.moneyParsers {
		result, err := parser(decimalAmount, network)
		if err != nil {
			continue
		}
		if result != nil {
			return *result, nil
		}
	}

	// All custom parsers returned nil, use default conversion
	return s.defaultMoneyConversion(decimalAmount, config)
}

// parseMoneyToDecimal converts Money (string | number) to decimal amount
func (s *ExactAvmScheme) parseMoneyToDecimal(price x402.Price) (float64, error) {
	if priceStr, ok := price.(string); ok {
		cleanPrice := strings.TrimSpace(priceStr)
		cleanPrice = strings.TrimPrefix(cleanPrice, "$")
		cleanPrice = strings.TrimSpace(cleanPrice)

		parts := strings.Fields(cleanPrice)
		if len(parts) >= 1 {
			amount, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return 0, fmt.Errorf(ErrFailedToParsePrice+": '%s': %w", priceStr, err)
			}
			return amount, nil
		}
	}

	switch v := price.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	}

	return 0, fmt.Errorf(ErrInvalidPriceFormat+": %v", price)
}

// defaultMoneyConversion converts decimal amount to USDC AssetAmount
func (s *ExactAvmScheme) defaultMoneyConversion(amount float64, config *avm.NetworkConfig) (x402.AssetAmount, error) {
	amountStr := fmt.Sprintf("%.6f", amount)
	parsedAmount, err := avm.ParseAmount(amountStr, config.DefaultAsset.Decimals)
	if err != nil {
		return x402.AssetAmount{}, fmt.Errorf(ErrFailedToConvertAmount+": %w", err)
	}

	return x402.AssetAmount{
		Amount: strconv.FormatUint(parsedAmount, 10),
		Asset:  config.DefaultAsset.ID,
		Extra:  make(map[string]interface{}),
	}, nil
}

// EnhancePaymentRequirements adds scheme-specific enhancements to V2 payment requirements
func (s *ExactAvmScheme) EnhancePaymentRequirements(
	ctx context.Context,
	requirements types.PaymentRequirements,
	supportedKind types.SupportedKind,
	extensionKeys []string,
) (types.PaymentRequirements, error) {
	_ = ctx

	networkStr := string(requirements.Network)
	config, err := avm.GetNetworkConfig(networkStr)
	if err != nil {
		return requirements, err
	}

	// Get asset info
	var assetInfo *avm.AssetInfo
	if requirements.Asset != "" {
		assetInfo, err = avm.GetAssetInfo(networkStr, requirements.Asset)
		if err != nil {
			return requirements, err
		}
	} else {
		assetInfo = &config.DefaultAsset
		requirements.Asset = assetInfo.ID
	}

	// Ensure amount is in the correct format (smallest unit)
	if requirements.Amount != "" && strings.Contains(requirements.Amount, ".") {
		amount, err := avm.ParseAmount(requirements.Amount, assetInfo.Decimals)
		if err != nil {
			return requirements, fmt.Errorf(ErrFailedToParseAmount+": %w", err)
		}
		requirements.Amount = strconv.FormatUint(amount, 10)
	}

	// Initialize extra map if needed
	if requirements.Extra == nil {
		requirements.Extra = make(map[string]interface{})
	}

	// Add feePayer from supportedKind.extra to payment requirements
	if supportedKind.Extra != nil {
		if feePayer, ok := supportedKind.Extra["feePayer"]; ok {
			requirements.Extra["feePayer"] = feePayer
		}
	}

	// Copy extensions from supportedKind if provided
	if supportedKind.Extra != nil {
		for _, key := range extensionKeys {
			if val, ok := supportedKind.Extra[key]; ok {
				requirements.Extra[key] = val
			}
		}
	}

	return requirements, nil
}

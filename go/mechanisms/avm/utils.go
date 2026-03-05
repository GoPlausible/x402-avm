package avm

import (
	"encoding/base64"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/types"
)

var (
	// Algorand address regex (58-character base32 string)
	algorandAddressRegex = regexp.MustCompile(`^[A-Z2-7]{58}$`)
)

// NormalizeNetwork converts V1 network names to CAIP-2 format
func NormalizeNetwork(network string) (string, error) {
	if strings.Contains(network, ":") {
		if _, ok := NetworkConfigs[network]; ok {
			return network, nil
		}
		return "", fmt.Errorf("unsupported Algorand network: %s", network)
	}

	caip2Network, ok := V1ToV2NetworkMap[network]
	if !ok {
		return "", fmt.Errorf("unsupported Algorand network: %s", network)
	}

	return caip2Network, nil
}

// GetNetworkConfig returns the configuration for a network
func GetNetworkConfig(network string) (*NetworkConfig, error) {
	caip2Network, err := NormalizeNetwork(network)
	if err != nil {
		return nil, err
	}

	config, ok := NetworkConfigs[caip2Network]
	if !ok {
		return nil, fmt.Errorf("network configuration not found: %s", network)
	}

	return &config, nil
}

// GetAssetInfo returns information about an asset on a network
func GetAssetInfo(network string, assetIDOrSymbol string) (*AssetInfo, error) {
	config, err := GetNetworkConfig(network)
	if err != nil {
		return nil, err
	}

	// Check if it matches the default asset ID
	if assetIDOrSymbol == config.DefaultAsset.ID {
		return &config.DefaultAsset, nil
	}

	// Check if it's a numeric ASA ID
	if _, err := strconv.ParseUint(assetIDOrSymbol, 10, 64); err == nil {
		return &AssetInfo{
			ID:       assetIDOrSymbol,
			Symbol:   "UNKNOWN",
			Decimals: DefaultDecimals,
		}, nil
	}

	// Default to the network's default asset
	return &config.DefaultAsset, nil
}

// ValidateAlgorandAddress checks if a string is a valid Algorand address
func ValidateAlgorandAddress(address string) bool {
	if !algorandAddressRegex.MatchString(address) {
		return false
	}

	_, err := types.DecodeAddress(address)
	return err == nil
}

// GetGenesisHashForNetwork returns the expected genesis hash for a CAIP-2 network
func GetGenesisHashForNetwork(network string) (string, error) {
	caip2, err := NormalizeNetwork(network)
	if err != nil {
		return "", err
	}

	hash, ok := CAIP2ToGenesisHash[caip2]
	if !ok {
		return "", fmt.Errorf("unknown genesis hash for network: %s", network)
	}

	return hash, nil
}

// ParseAmount converts a decimal string amount to token smallest units
func ParseAmount(amount string, decimals int) (uint64, error) {
	amount = strings.TrimSpace(amount)

	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid amount format: %s", amount)
	}

	intPart, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer part: %s", parts[0])
	}

	decPart := uint64(0)
	if len(parts) == 2 && parts[1] != "" {
		decStr := parts[1]
		if len(decStr) > decimals {
			decStr = decStr[:decimals]
		} else {
			decStr += strings.Repeat("0", decimals-len(decStr))
		}

		decPart, err = strconv.ParseUint(decStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid decimal part: %s", parts[1])
		}
	}

	multiplier := uint64(math.Pow10(decimals))
	result := intPart*multiplier + decPart

	return result, nil
}

// FormatAmount converts an amount in smallest units to a decimal string
func FormatAmount(amount uint64, decimals int) string {
	if amount == 0 {
		return "0"
	}

	divisor := uint64(math.Pow10(decimals))
	quotient := amount / divisor
	remainder := amount % divisor

	decStr := fmt.Sprintf("%0*d", decimals, remainder)
	decStr = strings.TrimRight(decStr, "0")

	if decStr == "" {
		return fmt.Sprintf("%d", quotient)
	}

	return fmt.Sprintf("%d.%s", quotient, decStr)
}

// EncodeTransaction encodes a transaction to base64 msgpack
func EncodeTransaction(txn types.Transaction) string {
	return base64.StdEncoding.EncodeToString(msgpack.Encode(txn))
}

// EncodeSignedTransaction encodes a signed transaction to base64 msgpack
func EncodeSignedTransaction(stxn types.SignedTxn) string {
	return base64.StdEncoding.EncodeToString(msgpack.Encode(stxn))
}

// EncodeSignedTransactionBytes encodes raw signed transaction bytes to base64
func EncodeSignedTransactionBytes(raw []byte) string {
	return base64.StdEncoding.EncodeToString(raw)
}

// DecodeRawTransaction decodes a base64-encoded transaction to raw bytes
func DecodeRawTransaction(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// DecodeSignedTransaction decodes a base64-encoded msgpack signed transaction
func DecodeSignedTransaction(encoded string) (*types.SignedTxn, error) {
	rawBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	var stxn types.SignedTxn
	if err := msgpack.Decode(rawBytes, &stxn); err != nil {
		return nil, fmt.Errorf("failed to decode signed transaction: %w", err)
	}

	return &stxn, nil
}

// DecodeUnsignedTransaction decodes a base64-encoded msgpack unsigned transaction
func DecodeUnsignedTransaction(encoded string) (*types.Transaction, error) {
	rawBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	var txn types.Transaction
	if err := msgpack.Decode(rawBytes, &txn); err != nil {
		return nil, fmt.Errorf("failed to decode unsigned transaction: %w", err)
	}

	return &txn, nil
}

// DecodeTransactionAuto decodes a base64-encoded transaction, trying signed first then unsigned.
// Returns the transaction, whether it has a signature, and the raw bytes.
func DecodeTransactionAuto(encoded string) (*types.Transaction, bool, []byte, error) {
	rawBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, false, nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Try decoding as signed transaction first
	var stxn types.SignedTxn
	if err := msgpack.Decode(rawBytes, &stxn); err == nil {
		hasSig := !isZeroSignature(stxn.Sig) || !isZeroMultisig(stxn.Msig) || len(stxn.Lsig.Logic) > 0
		return &stxn.Txn, hasSig, rawBytes, nil
	}

	// Fall back to unsigned transaction
	var txn types.Transaction
	if err := msgpack.Decode(rawBytes, &txn); err != nil {
		return nil, false, nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	return &txn, false, rawBytes, nil
}

// GetTransactionSender returns the sender address of a transaction
func GetTransactionSender(txn *types.Transaction) string {
	return txn.Sender.String()
}

// GetTransactionType returns the type of a transaction as a string
func GetTransactionType(txn *types.Transaction) string {
	return string(txn.Type)
}

// HasSignature checks if a signed transaction has any form of signature
func HasSignature(stxn *types.SignedTxn) bool {
	return !isZeroSignature(stxn.Sig) || !isZeroMultisig(stxn.Msig) || len(stxn.Lsig.Logic) > 0
}

// isZeroSignature checks if a signature is all zeros
func isZeroSignature(sig types.Signature) bool {
	for _, b := range sig {
		if b != 0 {
			return false
		}
	}
	return true
}

// isZeroMultisig checks if a multisig is empty/zero
func isZeroMultisig(msig types.MultisigSig) bool {
	return msig.Version == 0 && msig.Threshold == 0 && len(msig.Subsigs) == 0
}

// isZeroDigest checks if a digest is all zeros
func isZeroDigest(d types.Digest) bool {
	for _, b := range d {
		if b != 0 {
			return false
		}
	}
	return true
}

// IsAlgorandNetwork checks if a network string is an Algorand network
func IsAlgorandNetwork(network string) bool {
	return strings.HasPrefix(network, "algorand:") || strings.HasPrefix(network, "algorand-")
}

// IsTestnetNetwork checks if a network is a testnet
func IsTestnetNetwork(network string) bool {
	return network == AlgorandTestnetCAIP2 || network == AlgorandTestnetV1
}

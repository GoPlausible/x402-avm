package avm

import "os"

const (
	// SchemeExact is the scheme identifier for exact payments
	SchemeExact = "exact"

	// DefaultDecimals is the default token decimals for USDC
	DefaultDecimals = 6

	// CAIP-2 network identifiers (V2)
	// Format: algorand:<genesis-hash-base64>
	AlgorandMainnetCAIP2 = "algorand:wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8="
	AlgorandTestnetCAIP2 = "algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI="

	// Genesis hashes (base64-encoded)
	AlgorandMainnetGenesisHash = "wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8="
	AlgorandTestnetGenesisHash = "SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI="

	// V1 network names
	AlgorandMainnetV1 = "algorand-mainnet"
	AlgorandTestnetV1 = "algorand-testnet"

	// USDC ASA IDs
	USDCMainnetASAID = "31566704"
	USDCTestnetASAID = "10458941"

	// Algod API fallback endpoints (AlgoNode)
	FallbackAlgodMainnet = "https://mainnet-api.algonode.cloud"
	FallbackAlgodTestnet = "https://testnet-api.algonode.cloud"

	// Transaction limits
	MaxAtomicGroupSize = 16
	MinTxnFee          = 1000 // microAlgos
	MaxReasonableFee   = 16000 // microAlgos - sanity check for fee payer

	// Address validation
	AlgorandAddressLength = 58
)

// getAlgodURL returns the Algod URL for a network, checking env vars first
func getAlgodURL(envVar, fallback string) string {
	if url := os.Getenv(envVar); url != "" {
		return url
	}
	return fallback
}

var (
	// NetworkConfigs maps CAIP-2 identifiers to network configurations
	NetworkConfigs = map[string]NetworkConfig{
		AlgorandMainnetCAIP2: {
			Name:        "Algorand Mainnet",
			CAIP2:       AlgorandMainnetCAIP2,
			GenesisHash: AlgorandMainnetGenesisHash,
			AlgodURL:    getAlgodURL("ALGOD_MAINNET_URL", FallbackAlgodMainnet),
			DefaultAsset: AssetInfo{
				ID:       USDCMainnetASAID,
				Symbol:   "USDC",
				Decimals: DefaultDecimals,
			},
		},
		AlgorandTestnetCAIP2: {
			Name:        "Algorand Testnet",
			CAIP2:       AlgorandTestnetCAIP2,
			GenesisHash: AlgorandTestnetGenesisHash,
			AlgodURL:    getAlgodURL("ALGOD_TESTNET_URL", FallbackAlgodTestnet),
			DefaultAsset: AssetInfo{
				ID:       USDCTestnetASAID,
				Symbol:   "USDC",
				Decimals: DefaultDecimals,
			},
		},
	}

	// V1ToV2NetworkMap maps V1 network names to CAIP-2 identifiers
	V1ToV2NetworkMap = map[string]string{
		AlgorandMainnetV1: AlgorandMainnetCAIP2,
		AlgorandTestnetV1: AlgorandTestnetCAIP2,
	}

	// V2ToV1NetworkMap maps CAIP-2 identifiers to V1 network names
	V2ToV1NetworkMap = map[string]string{
		AlgorandMainnetCAIP2: AlgorandMainnetV1,
		AlgorandTestnetCAIP2: AlgorandTestnetV1,
	}

	// CAIP2ToGenesisHash maps CAIP-2 identifiers to genesis hashes
	CAIP2ToGenesisHash = map[string]string{
		AlgorandMainnetCAIP2: AlgorandMainnetGenesisHash,
		AlgorandTestnetCAIP2: AlgorandTestnetGenesisHash,
	}
)

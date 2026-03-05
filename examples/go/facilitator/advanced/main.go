package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

/**
 * Advanced Facilitator Examples
 *
 * This package demonstrates advanced patterns for production-ready x402 facilitators:
 *
 * - all-networks: Facilitator with all supported networks
 * - bazaar: Facilitator with bazaar discovery extension
 * - payment-identifier: Facilitator with payment identifier idempotency
 *
 * Usage:
 *   go run . all-networks
 *   go run . bazaar
 *   go run . payment-identifier
 */

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	pattern := "all-networks"
	if len(os.Args) > 1 {
		pattern = os.Args[1]
	}

	fmt.Printf("\n🚀 Running facilitator example: %s\n\n", pattern)

	// Get configuration
	evmPrivateKey := os.Getenv("EVM_PRIVATE_KEY")
	svmPrivateKey := os.Getenv("SVM_PRIVATE_KEY")
	avmPrivateKey := os.Getenv("AVM_PRIVATE_KEY")

	// Validate at least one private key is provided
	if evmPrivateKey == "" && svmPrivateKey == "" && avmPrivateKey == "" {
		fmt.Println("❌ At least one of EVM_PRIVATE_KEY, SVM_PRIVATE_KEY, or AVM_PRIVATE_KEY is required")
		os.Exit(1)
	}

	// Run the selected example
	switch pattern {
	case "all-networks":
		if err := runAllNetworksExample(evmPrivateKey, svmPrivateKey, avmPrivateKey); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

	case "bazaar":
		if err := runBazaarExample(evmPrivateKey, svmPrivateKey, avmPrivateKey); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

	case "payment-identifier":
		if err := runPaymentIdentifierExample(evmPrivateKey, svmPrivateKey, avmPrivateKey); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("❌ Unknown pattern: %s\n", pattern)
		fmt.Println("Available patterns: all-networks, bazaar, payment-identifier")
		os.Exit(1)
	}
}

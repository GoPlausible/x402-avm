package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	x402 "github.com/coinbase/x402/go"
	x402http "github.com/coinbase/x402/go/http"
	ginmw "github.com/coinbase/x402/go/http/gin"
	avm "github.com/coinbase/x402/go/mechanisms/avm"
	avmserver "github.com/coinbase/x402/go/mechanisms/avm/exact/server"
	evm "github.com/coinbase/x402/go/mechanisms/evm/exact/server"
	svm "github.com/coinbase/x402/go/mechanisms/svm/exact/server"
	ginfw "github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const (
	DefaultPort = "4021"
)

func main() {
	godotenv.Load()

	avmAddress := os.Getenv("AVM_PAYEE_ADDRESS")
	evmAddress := os.Getenv("EVM_PAYEE_ADDRESS")
	svmAddress := os.Getenv("SVM_PAYEE_ADDRESS")

	if avmAddress == "" && evmAddress == "" && svmAddress == "" {
		fmt.Println("❌ At least one of AVM_PAYEE_ADDRESS, EVM_PAYEE_ADDRESS, or SVM_PAYEE_ADDRESS is required")
		os.Exit(1)
	}

	facilitatorURL := os.Getenv("FACILITATOR_URL")
	if facilitatorURL == "" {
		fmt.Println("❌ FACILITATOR_URL environment variable is required")
		fmt.Println("   Example: https://x402.org/facilitator")
		os.Exit(1)
	}

	// Network configuration
	avmNetwork := x402.Network(avm.AlgorandTestnetCAIP2)                  // Algorand Testnet
	evmNetwork := x402.Network("eip155:84532")                            // Base Sepolia
	svmNetwork := x402.Network("solana:EtWTRABZaYq6iMfeYKouRu166VU2xqa1") // Solana Devnet

	fmt.Printf("🚀 Starting Gin x402 server...\n")
	if avmAddress != "" {
		fmt.Printf("   AVM Payee address: %s\n", avmAddress)
		fmt.Printf("   AVM Network: %s\n", avmNetwork)
	}
	if evmAddress != "" {
		fmt.Printf("   EVM Payee address: %s\n", evmAddress)
		fmt.Printf("   EVM Network: %s\n", evmNetwork)
	}
	if svmAddress != "" {
		fmt.Printf("   SVM Payee address: %s\n", svmAddress)
		fmt.Printf("   SVM Network: %s\n", svmNetwork)
	}
	fmt.Printf("   Facilitator: %s\n", facilitatorURL)

	// Create Gin router
	r := ginfw.Default()

	// Create HTTP facilitator client
	facilitatorClient := x402http.NewHTTPFacilitatorClient(&x402http.FacilitatorConfig{
		URL: facilitatorURL,
	})

	/**
	 * Configure x402 payment middleware
	 *
	 * This middleware protects specific routes with payment requirements.
	 * When a client accesses a protected route without payment, they receive
	 * a 402 Payment Required response with payment details.
	 */

	// Build accepts array dynamically based on configured addresses
	paymentOptions := x402http.PaymentOptions{}
	if avmAddress != "" {
		paymentOptions = append(paymentOptions, x402http.PaymentOption{
			Scheme:  "exact",
			Price:   "$0.001",
			Network: avmNetwork,
			PayTo:   avmAddress,
		})
	}
	if evmAddress != "" {
		paymentOptions = append(paymentOptions, x402http.PaymentOption{
			Scheme:  "exact",
			Price:   "$0.001",
			Network: evmNetwork,
			PayTo:   evmAddress,
		})
	}
	if svmAddress != "" {
		paymentOptions = append(paymentOptions, x402http.PaymentOption{
			Scheme:  "exact",
			Price:   "$0.001",
			Network: svmNetwork,
			PayTo:   svmAddress,
		})
	}

	routes := x402http.RoutesConfig{
		"GET /weather": {
			Accepts:     paymentOptions,
			Description: "Get weather data for a city",
			MimeType:    "application/json",
		},
	}

	// Build scheme config dynamically based on configured addresses
	schemes := []ginmw.SchemeConfig{}
	if avmAddress != "" {
		schemes = append(schemes, ginmw.SchemeConfig{
			Network: avmNetwork,
			Server:  avmserver.NewExactAvmScheme(),
		})
	}
	if evmAddress != "" {
		schemes = append(schemes, ginmw.SchemeConfig{
			Network: evmNetwork,
			Server:  evm.NewExactEvmScheme(),
		})
	}
	if svmAddress != "" {
		schemes = append(schemes, ginmw.SchemeConfig{
			Network: svmNetwork,
			Server:  svm.NewExactSvmScheme(),
		})
	}

	// Apply x402 payment middleware
	r.Use(ginmw.X402Payment(ginmw.Config{
		Routes:      routes,
		Facilitator: facilitatorClient,
		Schemes:     schemes,
		Timeout:     30 * time.Second,
	}))

	/**
	 * Protected endpoint - requires $0.001 USDC payment
	 *
	 * Clients must provide a valid x402 payment to access this endpoint.
	 * The payment is verified and settled before the endpoint handler runs.
	 */
	r.GET("/weather", func(c *ginfw.Context) {
		city := c.DefaultQuery("city", "San Francisco")

		weatherData := map[string]map[string]interface{}{
			"San Francisco": {"weather": "foggy", "temperature": 60},
			"New York":      {"weather": "cloudy", "temperature": 55},
			"London":        {"weather": "rainy", "temperature": 50},
			"Tokyo":         {"weather": "clear", "temperature": 65},
		}

		data, exists := weatherData[city]
		if !exists {
			data = map[string]interface{}{"weather": "sunny", "temperature": 70}
		}

		c.JSON(http.StatusOK, ginfw.H{
			"city":        city,
			"weather":     data["weather"],
			"temperature": data["temperature"],
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	})

	/**
	 * Health check endpoint - no payment required
	 *
	 * This endpoint is not protected by x402 middleware.
	 */
	r.GET("/health", func(c *ginfw.Context) {
		c.JSON(http.StatusOK, ginfw.H{
			"status":  "ok",
			"version": "2.0.0",
		})
	})

	fmt.Printf("   Server listening on http://localhost:%s\n\n", DefaultPort)

	if err := r.Run(":" + DefaultPort); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}

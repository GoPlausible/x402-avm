// Package integration_test contains integration tests for the x402 Go SDK.
// This file tests the AVM (Algorand) mechanism integration with V2 implementation.
// These tests make REAL on-chain transactions using private keys from environment variables.
package integration_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	algodClient "github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	algocrypto "github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/types"

	x402 "github.com/GoPlausible/x402-avm/go"
	avm "github.com/GoPlausible/x402-avm/go/mechanisms/avm"
	avmclient "github.com/GoPlausible/x402-avm/go/mechanisms/avm/exact/client"
	avmfacilitator "github.com/GoPlausible/x402-avm/go/mechanisms/avm/exact/facilitator"
	avmserver "github.com/GoPlausible/x402-avm/go/mechanisms/avm/exact/server"
	avmsigners "github.com/GoPlausible/x402-avm/go/signers/avm"
	x402types "github.com/GoPlausible/x402-avm/go/types"
)

// newRealClientAvmSigner creates a client signer using the helper
func newRealClientAvmSigner(privateKeyBase64 string) (avm.ClientAvmSigner, error) {
	return avmsigners.NewClientSignerFromPrivateKey(privateKeyBase64)
}

// Real Algorand facilitator signer for integration tests
type realFacilitatorAvmSigner struct {
	privateKey  []byte
	address     string
	algodURL    string
	algodToken  string
	algodClient *algodClient.Client
}

func newRealFacilitatorAvmSigner(privateKeyBase64 string, algodURL string, algodToken string) (*realFacilitatorAvmSigner, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	account, err := algocrypto.AccountFromPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to derive account: %w", err)
	}

	client, err := algodClient.MakeClient(algodURL, algodToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create algod client: %w", err)
	}

	return &realFacilitatorAvmSigner{
		privateKey:  keyBytes,
		address:     account.Address.String(),
		algodURL:    algodURL,
		algodToken:  algodToken,
		algodClient: client,
	}, nil
}

func (s *realFacilitatorAvmSigner) GetAddresses(ctx context.Context, network string) []string {
	return []string{s.address}
}

func (s *realFacilitatorAvmSigner) SignTransaction(ctx context.Context, txnBytes []byte, senderAddress string, network string) ([]byte, error) {
	if senderAddress != s.address {
		return nil, fmt.Errorf("no signer for address %s, have %s", senderAddress, s.address)
	}

	var txn types.Transaction
	if err := msgpack.Decode(txnBytes, &txn); err != nil {
		// Try as signed transaction
		var stxn types.SignedTxn
		if err2 := msgpack.Decode(txnBytes, &stxn); err2 != nil {
			return nil, fmt.Errorf("failed to decode transaction: %w", err)
		}
		txn = stxn.Txn
	}

	_, signedBytes, err := algocrypto.SignTransaction(s.privateKey, txn)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return signedBytes, nil
}

func (s *realFacilitatorAvmSigner) GetAlgodClient(network string) (*algodClient.Client, error) {
	return s.algodClient, nil
}

func (s *realFacilitatorAvmSigner) SimulateTransactions(ctx context.Context, txns [][]byte, network string) error {
	// For integration tests, we skip simulation since we're testing the full flow
	return nil
}

func (s *realFacilitatorAvmSigner) SendTransactions(ctx context.Context, signedTxns [][]byte, network string) (string, error) {
	// Concatenate all signed transactions
	var combined []byte
	for _, stxn := range signedTxns {
		combined = append(combined, stxn...)
	}

	txID, err := s.algodClient.SendRawTransaction(combined).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to send transactions: %w", err)
	}

	return txID, nil
}

func (s *realFacilitatorAvmSigner) WaitForConfirmation(ctx context.Context, txID string, network string, waitRounds uint64) error {
	status, err := s.algodClient.Status().Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	lastRound := status.LastRound
	for i := uint64(0); i < waitRounds; i++ {
		info, _, err := s.algodClient.PendingTransactionInformation(txID).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to get pending transaction: %w", err)
		}

		if info.ConfirmedRound > 0 {
			return nil
		}

		lastRound++
		s.algodClient.StatusAfterBlock(lastRound).Do(ctx)
	}

	return fmt.Errorf("transaction %s not confirmed after %d rounds", txID, waitRounds)
}

// Local facilitator client for testing
type localAvmFacilitatorClient struct {
	facilitator *x402.X402Facilitator
	signer      *realFacilitatorAvmSigner
}

func (l *localAvmFacilitatorClient) Verify(
	ctx context.Context,
	payloadBytes []byte,
	requirementsBytes []byte,
) (*x402.VerifyResponse, error) {
	return l.facilitator.Verify(ctx, payloadBytes, requirementsBytes)
}

func (l *localAvmFacilitatorClient) Settle(
	ctx context.Context,
	payloadBytes []byte,
	requirementsBytes []byte,
) (*x402.SettleResponse, error) {
	return l.facilitator.Settle(ctx, payloadBytes, requirementsBytes)
}

func (l *localAvmFacilitatorClient) GetSupported(ctx context.Context) (x402.SupportedResponse, error) {
	return l.facilitator.GetSupported(), nil
}

// TestAVMIntegrationV2 tests the full V2 AVM payment flow with real on-chain transactions
func TestAVMIntegrationV2(t *testing.T) {
	// Skip if environment variables not set
	clientPrivateKey := os.Getenv("AVM_CLIENT_PRIVATE_KEY")
	facilitatorPrivateKey := os.Getenv("AVM_FACILITATOR_PRIVATE_KEY")
	facilitatorAddress := os.Getenv("AVM_FACILITATOR_ADDRESS")
	resourceServerAddress := os.Getenv("AVM_RESOURCE_SERVER_ADDRESS")
	algodServer := os.Getenv("ALGOD_SERVER")
	algodToken := os.Getenv("ALGOD_TOKEN")

	if clientPrivateKey == "" || facilitatorPrivateKey == "" || facilitatorAddress == "" || resourceServerAddress == "" {
		t.Skip("Skipping AVM integration test: AVM_CLIENT_PRIVATE_KEY, AVM_FACILITATOR_PRIVATE_KEY, AVM_FACILITATOR_ADDRESS, and AVM_RESOURCE_SERVER_ADDRESS must be set")
	}

	if algodServer == "" {
		algodServer = "https://testnet-api.algonode.cloud"
	}

	t.Run("AVM V2 Flow - x402Client / x402ResourceServer / x402Facilitator", func(t *testing.T) {
		ctx := context.Background()

		// Create real client signer
		clientSigner, err := newRealClientAvmSigner(clientPrivateKey)
		if err != nil {
			t.Fatalf("Failed to create client signer: %v", err)
		}

		// Setup client with AVM v2 scheme
		client := x402.Newx402Client()
		avmClient := avmclient.NewExactAvmScheme(clientSigner, &avm.ClientConfig{
			AlgodURL:   algodServer,
			AlgodToken: algodToken,
		})
		client.Register(avm.AlgorandTestnetCAIP2, avmClient)

		// Create real facilitator signer
		facilitatorSigner, err := newRealFacilitatorAvmSigner(facilitatorPrivateKey, algodServer, algodToken)
		if err != nil {
			t.Fatalf("Failed to create facilitator signer: %v", err)
		}

		// Setup facilitator with AVM v2 scheme
		facilitator := x402.Newx402Facilitator()
		avmFacilitator := avmfacilitator.NewExactAvmScheme(facilitatorSigner)
		facilitator.Register([]x402.Network{x402.Network(avm.AlgorandTestnetCAIP2)}, avmFacilitator)

		// Create facilitator client wrapper
		facilitatorClient := &localAvmFacilitatorClient{
			facilitator: facilitator,
			signer:      facilitatorSigner,
		}

		// Setup resource server with AVM v2
		avmServer := avmserver.NewExactAvmScheme()
		server := x402.Newx402ResourceServer(
			x402.WithFacilitatorClient(facilitatorClient),
		)
		server.Register(x402.Network(avm.AlgorandTestnetCAIP2), avmServer)

		// Initialize server to fetch supported kinds
		err = server.Initialize(ctx)
		if err != nil {
			t.Fatalf("Failed to initialize server: %v", err)
		}

		// Server - builds PaymentRequired response
		accepts := []x402types.PaymentRequirements{
			{
				Scheme:  avm.SchemeExact,
				Network: avm.AlgorandTestnetCAIP2,
				Asset:   avm.USDCTestnetASAID,
				Amount:  "1000", // 0.001 USDC (6 decimals)
				PayTo:   resourceServerAddress,
				Extra: map[string]interface{}{
					"feePayer": facilitatorAddress,
				},
			},
		}
		resource := &x402types.ResourceInfo{
			URL:         "https://api.example.com/premium",
			Description: "Premium API Access",
			MimeType:    "application/json",
		}
		paymentRequiredResponse := server.CreatePaymentRequiredResponse(accepts, resource, "", nil)

		// Verify it's V2
		if paymentRequiredResponse.X402Version != 2 {
			t.Errorf("Expected X402Version 2, got %d", paymentRequiredResponse.X402Version)
		}

		// Verify feePayer is in requirements
		if len(paymentRequiredResponse.Accepts) == 0 {
			t.Fatal("Expected at least one payment requirement")
		}

		firstAccept := paymentRequiredResponse.Accepts[0]
		if firstAccept.Extra == nil || firstAccept.Extra["feePayer"] == nil {
			t.Fatal("Expected feePayer in payment requirements extra")
		}

		// Client - selects payment requirement (V2 typed)
		selected, err := client.SelectPaymentRequirements(paymentRequiredResponse.Accepts)
		if err != nil {
			t.Fatalf("Failed to select payment requirements: %v", err)
		}

		// Client - creates payment payload (V2 typed)
		paymentPayload, err := client.CreatePaymentPayload(ctx, selected, paymentRequiredResponse.Resource, paymentRequiredResponse.Extensions)
		if err != nil {
			t.Fatalf("Failed to create payment payload: %v", err)
		}

		// Verify payload is V2
		if paymentPayload.X402Version != 2 {
			t.Errorf("Expected payload X402Version 2, got %d", paymentPayload.X402Version)
		}

		// Verify payload structure
		if paymentPayload.Accepted.Scheme != avm.SchemeExact {
			t.Errorf("Expected scheme %s, got %s", avm.SchemeExact, paymentPayload.Accepted.Scheme)
		}

		avmPayload, err := avm.PayloadFromMap(paymentPayload.Payload)
		if err != nil {
			t.Fatalf("Failed to parse AVM payload: %v", err)
		}

		if len(avmPayload.PaymentGroup) == 0 {
			t.Error("Expected transactions in payment group")
		}

		// Server - finds matching requirements (typed)
		accepted := server.FindMatchingRequirements(accepts, paymentPayload)
		if accepted == nil {
			t.Fatal("No matching payment requirements found")
		}

		// Server - verifies payment (typed)
		verifyResponse, err := server.VerifyPayment(ctx, paymentPayload, *accepted)
		if err != nil {
			t.Fatalf("Failed to verify payment: %v", err)
		}

		if !verifyResponse.IsValid {
			t.Fatalf("Payment verification failed: %s", verifyResponse.InvalidReason)
		}

		if verifyResponse.Payer != clientSigner.Address() {
			t.Errorf("Expected payer %s, got %s", clientSigner.Address(), verifyResponse.Payer)
		}

		// Server - settles payment (REAL ON-CHAIN TRANSACTION, typed)
		settleResponse, err := server.SettlePayment(ctx, paymentPayload, *accepted)
		if err != nil {
			t.Fatalf("Failed to settle payment: %v", err)
		}

		if !settleResponse.Success {
			t.Fatalf("Payment settlement failed: %s", settleResponse.ErrorReason)
		}

		// Verify the transaction ID
		if settleResponse.Transaction == "" {
			t.Error("Expected transaction ID in settlement response")
		}

		if string(settleResponse.Network) != avm.AlgorandTestnetCAIP2 {
			t.Errorf("Expected network %s, got %s", avm.AlgorandTestnetCAIP2, settleResponse.Network)
		}

		if settleResponse.Payer != clientSigner.Address() {
			t.Errorf("Expected payer %s, got %s", clientSigner.Address(), settleResponse.Payer)
		}

		t.Logf("Settlement successful! TX ID: %s", settleResponse.Transaction)
	})
}

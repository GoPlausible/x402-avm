package main

import (
	x402 "github.com/GoPlausible/x402-avm/go"
	avm "github.com/GoPlausible/x402-avm/go/mechanisms/avm"
	avmclient "github.com/GoPlausible/x402-avm/go/mechanisms/avm/exact/client"
	evm "github.com/GoPlausible/x402-avm/go/mechanisms/evm/exact/client"
	svm "github.com/GoPlausible/x402-avm/go/mechanisms/svm/exact/client"
	avmsigners "github.com/GoPlausible/x402-avm/go/signers/avm"
	evmsigners "github.com/GoPlausible/x402-avm/go/signers/evm"
	svmsigners "github.com/GoPlausible/x402-avm/go/signers/svm"
)

/**
 * Mechanism Helper Registration Client
 *
 * This demonstrates a convenient pattern using mechanism helpers with wildcard
 * network registration for clean, readable client configuration.
 *
 * This approach is simpler than the builder pattern when you want to register
 * all networks of a particular type with the same signer.
 */

func createMechanismHelperRegistrationClient(evmPrivateKey, svmPrivateKey, avmPrivateKey string) (*x402.X402Client, error) {
	// Start with a new client
	client := x402.Newx402Client()

	// Register AVM scheme if key is provided (alphabetic order)
	if avmPrivateKey != "" {
		avmSigner, err := avmsigners.NewClientSignerFromPrivateKey(avmPrivateKey)
		if err != nil {
			return nil, err
		}
		client.Register("algorand:*", avmclient.NewExactAvmScheme(avmSigner, &avm.ClientConfig{}))
	}

	// Register EVM scheme if key is provided
	if evmPrivateKey != "" {
		evmSigner, err := evmsigners.NewClientSignerFromPrivateKey(evmPrivateKey)
		if err != nil {
			return nil, err
		}
		client.Register("eip155:*", evm.NewExactEvmScheme(evmSigner))
	}

	// Register SVM scheme if key is provided
	if svmPrivateKey != "" {
		svmSigner, err := svmsigners.NewClientSignerFromPrivateKey(svmPrivateKey)
		if err != nil {
			return nil, err
		}
		client.Register("solana:*", svm.NewExactSvmScheme(svmSigner))
	}

	return client, nil
}


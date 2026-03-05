package main

import (
	x402 "github.com/coinbase/x402/go"
	avmclient "github.com/coinbase/x402/go/mechanisms/avm/exact/client"
	evm "github.com/coinbase/x402/go/mechanisms/evm/exact/client"
	svm "github.com/coinbase/x402/go/mechanisms/svm/exact/client"
	avmsigners "github.com/coinbase/x402/go/signers/avm"
	evmsigners "github.com/coinbase/x402/go/signers/evm"
	svmsigners "github.com/coinbase/x402/go/signers/svm"

	avm "github.com/coinbase/x402/go/mechanisms/avm"
)

/**
 * Builder Pattern Client
 *
 * This demonstrates the basic way to configure an x402 client by chaining
 * Register() calls to map network patterns to scheme clients.
 *
 * This approach gives you fine-grained control over which networks use
 * which signers and schemes.
 */

func createBuilderPatternClient(evmPrivateKey, svmPrivateKey, avmPrivateKey string) (*x402.X402Client, error) {
	// Create client and register schemes using builder pattern
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


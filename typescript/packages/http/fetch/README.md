# @x402-avm/fetch

A utility package that extends the native `fetch` API to automatically handle 402 Payment Required responses using the x402 payment protocol v2. This package enables seamless integration of payment functionality into your applications when making HTTP requests, with first-class support for the Algorand Virtual Machine (AVM).

## Installation

```bash
pnpm install @x402-avm/fetch
```

## Quick Start

```typescript
import { wrapFetchWithPaymentFromConfig } from "@x402-avm/fetch";
import { ExactAvmScheme, ClientAvmSigner } from "@x402-avm/avm";
import algosdk from "algosdk";

// Create an AVM signer
const secretKey = Buffer.from(process.env.AVM_PRIVATE_KEY!, "base64");
const address = algosdk.encodeAddress(secretKey.slice(32));
const avmSigner: ClientAvmSigner = {
  address,
  signTransactions: async (txns: Uint8Array[], indexesToSign?: number[]) => {
    return txns.map((txn, i) => {
      if (indexesToSign && !indexesToSign.includes(i)) return null;
      const decoded = algosdk.decodeUnsignedTransaction(txn);
      const signed = algosdk.signTransaction(decoded, secretKey);
      return signed.blob;
    });
  },
};

// Wrap the fetch function with payment handling
const fetchWithPayment = wrapFetchWithPaymentFromConfig(fetch, {
  schemes: [
    {
      network: "algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=", // Algorand Testnet
      client: new ExactAvmScheme(avmSigner),
    },
  ],
});

// Make a request that may require payment
const response = await fetchWithPayment("https://api.example.com/paid-endpoint", {
  method: "GET",
});

const data = await response.json();
```

## API

### `wrapFetchWithPayment(fetch, client)`

Wraps the native fetch API to handle 402 Payment Required responses automatically.

#### Parameters

- `fetch`: The fetch function to wrap (typically `globalThis.fetch`)
- `client`: An x402Client instance with registered payment schemes

### `wrapFetchWithPaymentFromConfig(fetch, config)`

Convenience wrapper that creates an x402Client from a configuration object.

#### Parameters

- `fetch`: The fetch function to wrap (typically `globalThis.fetch`)
- `config`: Configuration object with the following properties:
  - `schemes`: Array of scheme registrations, each containing:
    - `network`: Network identifier (e.g., `'algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI='`, `'eip155:8453'`, `'algorand:*'` for wildcards)
    - `client`: The scheme client implementation (e.g., `ExactAvmScheme`, `ExactEvmScheme`, `ExactSvmScheme`)
    - `x402Version`: Optional protocol version (defaults to 2, set to 1 for legacy support)
  - `paymentRequirementsSelector`: Optional function to select payment requirements from multiple options

#### Returns

A wrapped fetch function that automatically handles 402 responses by:
1. Making the initial request
2. If a 402 response is received, parsing the payment requirements
3. Creating a payment header using the configured scheme client
4. Retrying the request with the payment header

## Examples

### Basic Usage with AVM

```typescript
import { config } from "dotenv";
import { wrapFetchWithPaymentFromConfig, decodePaymentResponseHeader } from "@x402-avm/fetch";
import { ExactAvmScheme, ClientAvmSigner } from "@x402-avm/avm";
import algosdk from "algosdk";

config();

const { AVM_PRIVATE_KEY, API_URL } = process.env;

// Create an AVM signer
const secretKey = Buffer.from(AVM_PRIVATE_KEY!, "base64");
const address = algosdk.encodeAddress(secretKey.slice(32));
const avmSigner: ClientAvmSigner = {
  address,
  signTransactions: async (txns: Uint8Array[], indexesToSign?: number[]) => {
    return txns.map((txn, i) => {
      if (indexesToSign && !indexesToSign.includes(i)) return null;
      const decoded = algosdk.decodeUnsignedTransaction(txn);
      const signed = algosdk.signTransaction(decoded, secretKey);
      return signed.blob;
    });
  },
};

const fetchWithPayment = wrapFetchWithPaymentFromConfig(fetch, {
  schemes: [
    {
      network: "algorand:*", // Support all Algorand networks
      client: new ExactAvmScheme(avmSigner),
    },
  ],
});

// Make a request to a paid API endpoint
fetchWithPayment(API_URL, {
  method: "GET",
})
  .then(async response => {
    const data = await response.json();

    // Optionally decode the payment response header
    const paymentResponse = response.headers.get("PAYMENT-RESPONSE");
    if (paymentResponse) {
      const decoded = decodePaymentResponseHeader(paymentResponse);
      console.log("Payment details:", decoded);
    }

    console.log("Response data:", data);
  })
  .catch(error => {
    console.error(error);
  });
```

### Using Builder Pattern

For more control, you can use the builder pattern to register multiple schemes:

```typescript
import { wrapFetchWithPayment, x402Client } from "@x402-avm/fetch";
import { ExactAvmScheme, ClientAvmSigner } from "@x402-avm/avm";
import { ExactEvmScheme } from "@x402-avm/evm/exact/client";
import { ExactSvmScheme } from "@x402-avm/svm/exact/client";
import algosdk from "algosdk";
import { privateKeyToAccount } from "viem/accounts";
import { createKeyPairSignerFromBytes } from "@solana/kit";
import { base58 } from "@scure/base";

// Create AVM signer
const secretKey = Buffer.from(process.env.AVM_PRIVATE_KEY!, "base64");
const address = algosdk.encodeAddress(secretKey.slice(32));
const avmSigner: ClientAvmSigner = {
  address,
  signTransactions: async (txns: Uint8Array[], indexesToSign?: number[]) => {
    return txns.map((txn, i) => {
      if (indexesToSign && !indexesToSign.includes(i)) return null;
      const decoded = algosdk.decodeUnsignedTransaction(txn);
      const signed = algosdk.signTransaction(decoded, secretKey);
      return signed.blob;
    });
  },
};

// Create EVM and SVM signers
const evmSigner = privateKeyToAccount("0xYourPrivateKey");
const svmSigner = await createKeyPairSignerFromBytes(base58.decode("YourSvmPrivateKey"));

// Build client with multiple schemes (AVM first)
const client = new x402Client()
  .register("algorand:*", new ExactAvmScheme(avmSigner))
  .register("eip155:*", new ExactEvmScheme(evmSigner))
  .register("solana:*", new ExactSvmScheme(svmSigner));

// Wrap fetch with the client
const fetchWithPayment = wrapFetchWithPayment(fetch, client);
```

### Multi-Chain Support

```typescript
import { wrapFetchWithPaymentFromConfig } from "@x402-avm/fetch";
import { ExactAvmScheme } from "@x402-avm/avm";
import { ExactEvmScheme } from "@x402-avm/evm";
import { ExactSvmScheme } from "@x402-avm/svm";

const fetchWithPayment = wrapFetchWithPaymentFromConfig(fetch, {
  schemes: [
    // AVM chains
    {
      network: "algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=", // Algorand Testnet
      client: new ExactAvmScheme(avmSigner),
    },
    // EVM chains
    {
      network: "eip155:8453", // Base
      client: new ExactEvmScheme(evmAccount),
    },
    // SVM chains
    {
      network: "solana:EtWTRABZaYq6iMfeYKouRu166VU2xqa1", // Solana devnet
      client: new ExactSvmScheme(svmSigner),
    },
  ],
});
```

### Custom Payment Requirements Selector

```typescript
import { wrapFetchWithPaymentFromConfig, type SelectPaymentRequirements } from "@x402-avm/fetch";
import { ExactAvmScheme } from "@x402-avm/avm";

// Custom selector that prefers the cheapest option
const selectCheapestOption: SelectPaymentRequirements = (version, accepts) => {
  if (!accepts || accepts.length === 0) {
    throw new Error("No payment options available");
  }

  // Sort by amount and return the cheapest
  const sorted = [...accepts].sort((a, b) =>
    BigInt(a.amount) - BigInt(b.amount)
  );

  return sorted[0];
};

const fetchWithPayment = wrapFetchWithPaymentFromConfig(fetch, {
  schemes: [
    {
      network: "algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=",
      client: new ExactAvmScheme(avmSigner),
    },
  ],
  paymentRequirementsSelector: selectCheapestOption,
});
```


# @x402-avm/axios

A utility package that extends Axios to automatically handle 402 Payment Required responses using the x402 payment protocol v2. This package enables seamless integration of payment functionality into your applications when making HTTP requests, with first-class support for the Algorand Virtual Machine (AVM).

## Installation

```bash
pnpm install @x402-avm/axios
```

## Quick Start

```typescript
import axios from "axios";
import { wrapAxiosWithPaymentFromConfig } from "@x402-avm/axios";
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

// Wrap the axios instance with payment handling
const api = wrapAxiosWithPaymentFromConfig(axios.create(), {
  schemes: [
    {
      network: "algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=", // Algorand Testnet
      client: new ExactAvmScheme(avmSigner),
    },
  ],
});

// Make a request that may require payment
const response = await api.get("https://api.example.com/paid-endpoint");

const data = response.data;
```

## API

### `wrapAxiosWithPayment(axiosInstance, client)`

Wraps an Axios instance to handle 402 Payment Required responses automatically.

#### Parameters

- `axiosInstance`: The Axios instance to wrap (typically from `axios.create()`)
- `client`: An x402Client instance with registered payment schemes

### `wrapAxiosWithPaymentFromConfig(axiosInstance, config)`

Convenience wrapper that creates an x402Client from a configuration object.

#### Parameters

- `axiosInstance`: The Axios instance to wrap (typically from `axios.create()`)
- `config`: Configuration object with the following properties:
  - `schemes`: Array of scheme registrations, each containing:
    - `network`: Network identifier (e.g., `'algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI='`, `'eip155:8453'`, `'algorand:*'` for wildcards)
    - `client`: The scheme client implementation (e.g., `ExactAvmScheme`, `ExactEvmScheme`, `ExactSvmScheme`)
    - `x402Version`: Optional protocol version (defaults to 2, set to 1 for legacy support)
  - `paymentRequirementsSelector`: Optional function to select payment requirements from multiple options

#### Returns

A wrapped Axios instance that automatically handles 402 responses by:
1. Making the initial request
2. If a 402 response is received, parsing the payment requirements
3. Creating a payment header using the configured scheme client
4. Retrying the request with the payment header

## Examples

### Basic Usage with AVM

```typescript
import { config } from "dotenv";
import axios from "axios";
import { wrapAxiosWithPaymentFromConfig, decodePaymentResponseHeader } from "@x402-avm/axios";
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

const api = wrapAxiosWithPaymentFromConfig(axios.create(), {
  schemes: [
    {
      network: "algorand:*", // Support all Algorand networks
      client: new ExactAvmScheme(avmSigner),
    },
  ],
});

// Make a request to a paid API endpoint
api.get(API_URL)
  .then(response => {
    const data = response.data;

    // Optionally decode the payment response header
    const paymentResponse = response.headers["payment-response"];
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
import axios from "axios";
import { wrapAxiosWithPayment, x402Client } from "@x402-avm/axios";
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

// Wrap axios with the client
const api = wrapAxiosWithPayment(axios.create(), client);
```

### Multi-Chain Support

```typescript
import axios from "axios";
import { wrapAxiosWithPaymentFromConfig } from "@x402-avm/axios";
import { ExactAvmScheme } from "@x402-avm/avm";
import { ExactEvmScheme } from "@x402-avm/evm";
import { ExactSvmScheme } from "@x402-avm/svm";

const api = wrapAxiosWithPaymentFromConfig(axios.create(), {
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
import axios from "axios";
import { wrapAxiosWithPaymentFromConfig, type SelectPaymentRequirements } from "@x402-avm/axios";
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

const api = wrapAxiosWithPaymentFromConfig(axios.create(), {
  schemes: [
    {
      network: "algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=",
      client: new ExactAvmScheme(avmSigner),
    },
  ],
  paymentRequirementsSelector: selectCheapestOption,
});
```


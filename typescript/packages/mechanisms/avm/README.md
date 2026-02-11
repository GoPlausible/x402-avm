# @x402-avm/avm

AVM (Algorand Virtual Machine) implementation of the x402 payment protocol using the **Exact** payment scheme with ASA (Algorand Standard Asset) transfers.

## Installation

```bash
npm install @x402-avm/avm
```

## Overview

This package provides three main components for handling x402 payments on Algorand:

- **Client** - For applications that need to make payments (have wallets/signers)
- **Facilitator** - For payment processors that verify and execute on-chain transactions
- **Service** - For resource servers that accept payments and build payment requirements

## Package Exports

### Main Package (`@x402-avm/avm`)

**V2 Protocol Support** - Modern x402 protocol with CAIP-2 network identifiers

**Client** (`@x402-avm/avm` or `@x402-avm/avm/exact/client`):
- `ExactAvmScheme` - V2 client implementation using ASA transfers
- `ClientAvmSigner` - TypeScript interface for client signers (implement with `algosdk`)

**Facilitator** (`@x402-avm/avm/exact/facilitator`):
- `ExactAvmScheme` - V2 facilitator for payment verification and settlement
- `FacilitatorAvmSigner` - TypeScript interface for facilitator signers (implement with `algosdk`)

**Server** (`@x402-avm/avm/exact/server`):
- `ExactAvmScheme` - V2 service for building payment requirements

### V1 Package (`@x402-avm/avm/v1`)

**V1 Protocol Support** - Legacy x402 protocol with simple network names

**Exports:**
- `ExactAvmSchemeV1` - V1 scheme implementation (client and facilitator)
- `NETWORKS` - Array of all supported V1 network names

**Supported V1 Networks:**
```typescript
[
  "algorand-mainnet",  // Mainnet
  "algorand-testnet"   // Testnet
]
```

## Version Differences

### V2 (Main Package)
- Network format: CAIP-2 (`algorand:wGHE2Pwdvd7S12BL5FaOP20EGYesN73k`)
- Wildcard support: Yes (`algorand:*`)
- Payload structure: Partial (core wraps with metadata)
- Extensions: Full support
- Transaction: Atomic group with optional fee payer

### V1 (V1 Package)
- Network format: Simple names (`algorand-testnet`)
- Wildcard support: No (fixed list)
- Payload structure: Complete
- Extensions: Limited
- Transaction: Atomic group with optional fee payer

## Usage Patterns

### 1. Direct Registration

```typescript
import { x402Client } from "@x402-avm/core/client";
import { ExactAvmScheme } from "@x402-avm/avm";
import { ExactAvmSchemeV1 } from "@x402-avm/avm/v1";

const client = new x402Client()
  .register("algorand:*", new ExactAvmScheme(signer))
  .registerV1("algorand-testnet", new ExactAvmSchemeV1(signer))
  .registerV1("algorand-mainnet", new ExactAvmSchemeV1(signer));
```

### 2. Using Config (Flexible)

```typescript
import { x402Client } from "@x402-avm/core/client";
import { ExactAvmScheme } from "@x402-avm/avm";

const client = x402Client.fromConfig({
  schemes: [
    { network: "algorand:*", client: new ExactAvmScheme(signer) },
    {
      network: "algorand-testnet",
      client: new ExactAvmSchemeV1(signer),
      x402Version: 1
    }
  ]
});
```

## Supported Networks

**V2 Networks** (via CAIP-2):
- `algorand:wGHE2Pwdvd7S12BL5FaOP20EGYesN73k` - Mainnet
- `algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDe` - Testnet
- `algorand:*` - Wildcard (matches all Algorand networks)

**V1 Networks** (simple names):
- `algorand-mainnet` - Mainnet
- `algorand-testnet` - Testnet

## Signer Implementation

This package exports `ClientAvmSigner` and `FacilitatorAvmSigner` as TypeScript interfaces. You implement them directly using `algosdk`:

### Client Signer

```typescript
import algosdk from "algosdk";

// Decode Base64 private key (64 bytes: 32-byte seed + 32-byte public key)
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
```

### Facilitator Signer

See [facilitator example](../../examples/typescript/facilitator/) for a full `FacilitatorAvmSigner` implementation.

## Asset Support

Supports Algorand Standard Assets (ASA):
- USDC (primary)
- Any ASA with proper opt-in

## Transaction Structure

The exact payment scheme uses atomic transaction groups with:
- Payment transaction (ASA transfer or ALGO payment)
- Optional fee payer transaction (gasless transactions)
- Transaction simulation for validation

## Development

```bash
# Build
pnpm build

# Test
pnpm test

# Integration tests
pnpm test:integration

# Lint & Format
pnpm lint
pnpm format
```

## Related Packages

- `@x402-avm/core` - Core protocol types and client
- `@x402-avm/fetch` - HTTP wrapper with automatic payment handling
- `@x402-avm/evm` - EVM/Ethereum implementation
- `@x402-avm/svm` - Solana/SVM implementation
- `algosdk` - Algorand JavaScript SDK (peer dependency)

# @x402-avm/paywall

Modular paywall UI for the x402 payment protocol with support for Algorand, EVM, and Solana networks.

## Features

- Pre-built paywall UI out of the box
- Wallet connection (Pera, Defly, Lute, MetaMask, Phantom, etc.)
- USDC balance checking
- Multi-network support (AVM + EVM + Solana)
- Tree-shakeable - only bundle what you need
- Fully customizable via builder pattern

## Installation

```bash
pnpm add @x402-avm/paywall
```

## Bundle Sizes

Choose the import that matches your needs:

| Import | Size | Networks | Use Case |
|--------|------|----------|----------|
| `@x402-avm/paywall` | ~4MB | AVM + EVM + Solana | Multi-network apps |
| `@x402-avm/paywall/avm` | ~1MB | AVM only | Algorand apps |
| `@x402-avm/paywall/evm` | 3.4MB | EVM only | Base, Ethereum, Polygon, etc. |
| `@x402-avm/paywall/svm` | 1.0MB | Solana only | Solana apps |

## Usage

### Option 1: AVM Only (Algorand)

```typescript
import { createPaywall } from '@x402-avm/paywall';
import { avmPaywall } from '@x402-avm/paywall/avm';

const paywall = createPaywall()
  .withNetwork(avmPaywall)
  .withConfig({
    appName: 'My Algorand App',
    testnet: true
  })
  .build();

// Use with Express
app.use(paymentMiddleware(routes, facilitators, schemes, undefined, paywall));
```

### Option 2: EVM Only

```typescript
import { createPaywall } from '@x402-avm/paywall';
import { evmPaywall } from '@x402-avm/paywall/evm';

const paywall = createPaywall()
  .withNetwork(evmPaywall)
  .withConfig({
    appName: 'My App',
    testnet: true
  })
  .build();
```

### Option 3: Solana Only

```typescript
import { createPaywall } from '@x402-avm/paywall';
import { svmPaywall } from '@x402-avm/paywall/svm';

const paywall = createPaywall()
  .withNetwork(svmPaywall)
  .withConfig({
    appName: 'My Solana App',
    testnet: true
  })
  .build();
```

### Option 4: Multi-Network

```typescript
import { createPaywall } from '@x402-avm/paywall';
import { avmPaywall } from '@x402-avm/paywall/avm';
import { evmPaywall } from '@x402-avm/paywall/evm';
import { svmPaywall } from '@x402-avm/paywall/svm';

const paywall = createPaywall()
  .withNetwork(avmPaywall)   // First-match priority
  .withNetwork(evmPaywall)   // Fallback option
  .withNetwork(svmPaywall)   // Fallback option
  .withConfig({
    appName: 'Multi-chain App',
    testnet: true
  })
  .build();
```

## Configuration

### PaywallConfig Options

```typescript
interface PaywallConfig {
  appName?: string;              // App name shown in wallet connection
  appLogo?: string;              // App logo URL
  currentUrl?: string;           // URL of protected resource
  testnet?: boolean;             // Use testnet (default: true)
}
```

## How It Works

### First-Match Selection

When multiple networks are registered, the paywall uses **first-match selection**:

1. Iterates through `paymentRequired.accepts` array
2. Finds the first payment requirement that has a registered handler
3. Uses that handler to generate the HTML

**Example:**
```typescript
// Server returns multiple options
{
  "accepts": [
    { "network": "algorand:SGO1GKS...", ... },  // First
    { "network": "eip155:8453", ... },           // Second
    { "network": "solana:5eykt...", ... }        // Third
  ]
}

// If all handlers registered, Algorand is selected (it's first in accepts)
const paywall = createPaywall()
  .withNetwork(avmPaywall)
  .withNetwork(evmPaywall)
  .withNetwork(svmPaywall)
  .build();
```

### Supported Networks

**AVM Networks** (via `avmPaywall`):
- CAIP-2: `algorand:*` (e.g., `algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=` for testnet)
- Wallets: Pera, Defly, Lute (via @txnlab/use-wallet)

**EVM Networks** (via `evmPaywall`):
- CAIP-2: `eip155:*` (e.g., `eip155:8453` for Base, `eip155:84532` for Base Sepolia)

**Solana Networks** (via `svmPaywall`):
- CAIP-2: `solana:*` (e.g., `solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp` for mainnet)

## With HTTP Middleware

### Express

```typescript
import express from 'express';
import { paymentMiddleware } from '@x402-avm/express';
import { createPaywall } from '@x402-avm/paywall';
import { avmPaywall } from '@x402-avm/paywall/avm';

const app = express();

const paywall = createPaywall()
  .withNetwork(avmPaywall)
  .withConfig({ appName: 'My API' })
  .build();

app.use(paymentMiddleware(
  { "/api/premium": { price: "$0.10", network: "algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=", payTo: "ALGO_ADDRESS..." } },
  facilitators,
  schemes,
  undefined,
  paywall
));
```

### Automatic Detection

If you provide `paywallConfig` without a custom paywall, `@x402-avm/core` automatically:
1. Tries to load `@x402-avm/paywall` if installed
2. Falls back to basic HTML if not installed

```typescript
// Simple usage - auto-detects @x402-avm/paywall
app.use(paymentMiddleware(routes, facilitators, schemes, {
  appName: 'My App',
  testnet: true
}));
```

## Custom Network Handlers

You can create custom handlers for new networks:

```typescript
import { createPaywall, type PaywallNetworkHandler } from '@x402-avm/paywall';
import { avmPaywall } from '@x402-avm/paywall/avm';
import { evmPaywall } from '@x402-avm/paywall/evm';

const customPaywall: PaywallNetworkHandler = {
  supports: (req) => req.network.startsWith('custom:'),
  generateHtml: (req, paymentRequired, config) => {
    return `<!DOCTYPE html>...`;  // Your custom network paywall
  }
};

const paywall = createPaywall()
  .withNetwork(avmPaywall)
  .withNetwork(evmPaywall)
  .withNetwork(customPaywall)  // Custom handler
  .build();
```

## Development

### Build

```bash
pnpm build:paywall  # Generate HTML templates
pnpm build          # Build TypeScript
```

### Test

```bash
pnpm test           # Run unit tests
```
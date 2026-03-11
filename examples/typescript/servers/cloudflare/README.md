# x402 Cloudflare Worker Example Server

A multi-chain x402 resource server running on [Cloudflare Workers](https://workers.cloudflare.com/), built with [Hono](https://hono.dev/) and the `@x402-avm` SDK. It provides pay-per-request API endpoints protected by the x402 payment protocol across **Algorand (AVM)**, **Ethereum / Base Sepolia (EVM)**, and **Solana Devnet (SVM)**.

A live instance is deployed at **[example.x402.goplausible.xyz](https://example.x402.goplausible.xyz)**.

## What it does

This worker demonstrates how to build an x402-protected resource server on Cloudflare's edge network. Every protected endpoint:

1. Returns **402 Payment Required** with a `PAYMENT-REQUIRED` header describing accepted payment options.
2. Accepts a signed payment via the `PAYMENT-SIGNATURE` request header.
3. Verifies and settles the payment on-chain through the [GoPlausible facilitator](https://facilitator.goplausible.xyz).
4. Serves the protected content only after successful settlement.

No API keys, sessions, or accounts are needed — payment is the authentication.

## Endpoints

### Landing page

| Route   | Description                                                                          |
| ------- | ------------------------------------------------------------------------------------ |
| `GET /` | Interactive landing page listing all available endpoints with copy-to-clipboard URLs |

### Protected API (JSON response)

Each weather endpoint returns a JSON weather report. All three accept payments on any of the three networks, but present their primary network first in the `accepts` array.

| Route              | Primary network  | Price  | Response                                                  |
| ------------------ | ---------------- | ------ | --------------------------------------------------------- |
| `GET /avm/weather` | Algorand Testnet | $0.001 | `{ "report": { "weather": "sunny", "temperature": 70 } }` |
| `GET /evm/weather` | Base Sepolia     | $0.001 | Same                                                      |
| `GET /svm/weather` | Solana Devnet    | $0.001 | Same                                                      |

### Protected page (HTML response)

Each protected endpoint returns a styled HTML page confirming payment verification.

| Route                | Primary network  | Price  | Response                  |
| -------------------- | ---------------- | ------ | ------------------------- |
| `GET /avm/protected` | Algorand Testnet | $0.001 | Payment confirmation HTML |
| `GET /evm/protected` | Base Sepolia     | $0.001 | Same                      |
| `GET /svm/protected` | Solana Devnet    | $0.001 | Same                      |

### Bazaar discovery

The weather endpoints include [Bazaar discovery extensions](https://github.com/GoPlausible/x402-avm/tree/main/typescript/packages/extensions) in their 402 response, providing input/output schemas so AI agents and automated clients can discover what the API expects and returns.

## How it works

The server registers three payment schemes — `ExactAvmScheme`, `ExactEvmScheme`, and `ExactSvmScheme` — with an `x402ResourceServer` from `@x402-avm/hono`. A `paymentMiddleware` intercepts all requests, checks the route configuration, and either returns 402 or verifies the attached payment before passing through to the route handler.

```
Client                  Cloudflare Worker                Facilitator
  |                          |                               |
  |--- GET /avm/weather ---->|                               |
  |<--- 402 + PAYMENT-REQ ---|                               |
  |                          |                               |
  | (sign tx off-chain)      |                               |
  |                          |                               |
  |--- GET + PAYMENT-SIG --->|                               |
  |                          |--- verify + settle ---------->|
  |                          |<-- settlement receipt ---------|
  |<--- 200 + JSON ----------|                               |
```

## Setup

### Prerequisites

- Node.js >= 18
- [pnpm](https://pnpm.io/) >= 10.7.0
- A [Cloudflare account](https://dash.cloudflare.com/sign-up) (free tier works)
- Wallet addresses for at least one network (AVM, EVM, or SVM) to receive payments

### Install

From the `examples/typescript/` workspace root:

```bash
pnpm install
```

### Configure

Copy the example environment file and fill in your values:

```bash
cp .dev.vars.example .dev.vars
```

Edit `.dev.vars`:

```ini
EVM_ADDRESS=0xYourEvmAddress
SVM_ADDRESS=YourSolanaAddress
AVM_ADDRESS=YourAlgorandAddress
FACILITATOR_URL=https://facilitator.goplausible.xyz
```

For production, set these as Cloudflare Worker secrets:

```bash
wrangler secret put EVM_ADDRESS
wrangler secret put SVM_ADDRESS
wrangler secret put AVM_ADDRESS
wrangler secret put FACILITATOR_URL
```

### Run locally

```bash
pnpm --filter cloudflare-x402-example dev
```

The worker starts at `http://localhost:8787`.

### Deploy

```bash
pnpm --filter cloudflare-x402-example deploy
```

## Testing with an x402 client

Use any x402-compatible client to test the endpoints. For example, with the `@x402-avm/fetch` client:

```typescript
import { wrapFetch } from "@x402-avm/fetch";
import { ExactAvmScheme } from "@x402-avm/avm/exact/client";

const payFetch = wrapFetch(fetch, {
  avmPrivateKey: process.env.AVM_PRIVATE_KEY,
  schemes: [new ExactAvmScheme()],
});

const res = await payFetch("https://example.x402.goplausible.xyz/avm/weather");
const data = await res.json();
console.log(data);
// { report: { weather: "sunny", temperature: 70 } }
```

Or test manually with `curl` to see the 402 flow:

```bash
# Step 1: Get payment requirements
curl -i https://example.x402.goplausible.xyz/avm/weather
# Returns 402 with PAYMENT-REQUIRED header

# Step 2: Send with payment (construct PAYMENT-SIGNATURE from your client)
curl -i -H "PAYMENT-SIGNATURE: <base64-payment-json>" \
  https://example.x402.goplausible.xyz/avm/weather
```

## Project structure

```
cloudflare/
  src/
    index.ts          # Hono app with payment middleware, routes, and landing page
  public/
    favicon.ico       # Site favicon
    goPlausible-logo-type-h.png   # GoPlausible logo
  .dev.vars.example   # Template for local environment variables
  package.json        # Dependencies and scripts
  tsconfig.json       # TypeScript configuration
  wrangler.jsonc      # Cloudflare Workers configuration
```

## Links

- [x402 Protocol](https://github.com/coinbase/x402)
- [GoPlausible](https://goplausible.com)
- [x402-avm (Algorand + multi-chain)](https://github.com/GoPlausible/x402-avm/tree/branch-algorand-v2-typescript-algokit-pr-publish)
- [x402 on Algorand documentation](https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/README.md)
- [GoPlausible Facilitator](https://facilitator.goplausible.xyz)
- [Live demo](https://x402.goplausible.xyz)

## License

Apache-2.0 — see the repository [LICENSE](https://github.com/GoPlausible/x402-avm/blob/main/LICENSE) for details.

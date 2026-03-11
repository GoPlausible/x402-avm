# x402 SST Example Server

An example x402 resource server deployed as an **AWS Lambda** function via **[SST v3 (Ion)](https://sst.dev)**. It demonstrates how to protect API endpoints and HTML pages with pay-per-request payments using the x402 protocol, supporting Algorand (AVM), Ethereum (EVM), and Solana (SVM) networks simultaneously.

## What It Does

This server deploys a Hono application to AWS Lambda behind an API Gateway v2 (HTTP API). Every protected route requires a micro-payment before the resource is served. The payment flow follows the x402 protocol:

1. Client makes a request to a protected endpoint
2. Server responds `402 Payment Required` with a `PAYMENT-REQUIRED` header describing what to pay, to whom, and on which network
3. Client constructs and signs a transaction (Algorand, EVM, or Solana)
4. Client retries the request with the signed payment in the `PAYMENT-SIGNATURE` header
5. Server verifies and settles the payment via the GoPlausible facilitator, then serves the response

No accounts, no sessions, no API keys — just HTTP and on-chain payments.

## Example Resource Endpoints

All endpoints cost **$0.001** per request and run on **testnet** networks.

### Protected API — JSON Response

| Endpoint           | Primary Network    | Description                                                            |
| ------------------ | ------------------ | ---------------------------------------------------------------------- |
| `GET /avm/weather` | Algorand Testnet   | Returns `{ report: { weather, temperature } }` — AVM payment preferred |
| `GET /evm/weather` | Base Sepolia (EVM) | Returns `{ report: { weather, temperature } }` — EVM payment preferred |
| `GET /svm/weather` | Solana Devnet      | Returns `{ report: { weather, temperature } }` — SVM payment preferred |

Each weather endpoint accepts payments from all three networks. The order in `accepts[]` signals the server's preference, but any supported network is valid.

### Protected URL — HTML Response

| Endpoint             | Primary Network    | Description                                                   |
| -------------------- | ------------------ | ------------------------------------------------------------- |
| `GET /avm/protected` | Algorand Testnet   | Returns an "Access Granted" HTML page — AVM payment preferred |
| `GET /evm/protected` | Base Sepolia (EVM) | Returns an "Access Granted" HTML page — EVM payment preferred |
| `GET /svm/protected` | Solana Devnet      | Returns an "Access Granted" HTML page — SVM payment preferred |

### Landing Page

`GET /` — Unprotected landing page listing all endpoints with copy-URL buttons.

## Architecture

```
API Gateway v2 (HTTP API)
      │
      ▼
AWS Lambda (Node.js 20)
      │
      ├── Logging middleware (all routes)
      ├── x402 paymentMiddleware (all protected routes)
      │       │
      │       └── HTTPFacilitatorClient → GoPlausible Facilitator
      │
      ├── GET /                    → Landing page (HTML)
      ├── GET /avm/weather         → Weather JSON  [x402 protected]
      ├── GET /evm/weather         → Weather JSON  [x402 protected]
      ├── GET /svm/weather         → Weather JSON  [x402 protected]
      ├── GET /avm/protected       → Access HTML   [x402 protected]
      ├── GET /evm/protected       → Access HTML   [x402 protected]
      └── GET /svm/protected       → Access HTML   [x402 protected]
```

Deployed and managed by SST v3 (Ion) — infrastructure is defined in `sst.config.ts` using the `sst.aws.ApiGatewayV2` construct. SST bundles the Lambda handler with esbuild automatically.

## Prerequisites

- [AWS account](https://aws.amazon.com) with credentials configured (`aws configure` or environment variables)
- [SST CLI](https://sst.dev/docs/reference/cli/) — installed automatically via `sst` devDependency
- [Node.js](https://nodejs.org) >= 18
- [pnpm](https://pnpm.io) >= 10.7

## Setup

### 1. Install dependencies

From the examples workspace root:

```bash
cd examples/typescript
pnpm install
```

Or from this directory:

```bash
pnpm install
```

### 2. Configure environment variables

```bash
cp .env.example .env
```

Edit `.env`:

```env
EVM_ADDRESS=0xYourEthereumAddress
SVM_ADDRESS=YourSolanaAddress
AVM_ADDRESS=YourAlgorandAddress
FACILITATOR_URL=https://facilitator.goplausible.xyz
```

| Variable          | Description                                                         |
| ----------------- | ------------------------------------------------------------------- |
| `EVM_ADDRESS`     | Ethereum address (with `0x` prefix) that receives EVM payments      |
| `SVM_ADDRESS`     | Solana public key that receives SVM payments                        |
| `AVM_ADDRESS`     | Algorand address that receives AVM payments                         |
| `FACILITATOR_URL` | x402 facilitator URL (defaults to GoPlausible's public facilitator) |

### 3. Run locally

```bash
pnpm dev
```

SST's `dev` mode runs your Lambda function locally via a live development tunnel connected to real AWS resources. The API Gateway URL is printed to the console.

### 4. Deploy to AWS

```bash
pnpm deploy
```

The deployed API Gateway URL is printed at the end of deployment. For production deployments use:

```bash
pnpm deploy --stage production
```

### 5. Remove stack

```bash
pnpm remove
```

## Production Secrets

For production, set secrets via SST instead of `.env`:

```bash
npx sst secret set EVM_ADDRESS 0xYourAddress
npx sst secret set SVM_ADDRESS YourSolanaAddress
npx sst secret set AVM_ADDRESS YourAlgorandAddress
npx sst secret set FACILITATOR_URL https://facilitator.goplausible.xyz
```

## Testing

Test with the x402 CLI client or any x402-compatible HTTP client:

```bash
# Without payment — expect 402
curl -i https://<your-api-id>.execute-api.<region>.amazonaws.com/avm/weather

# With x402 client
npx x402-fetch https://<your-api-id>.execute-api.<region>.amazonaws.com/avm/weather
```

Live testnet endpoints are available at [example.x402.goplausible.xyz](https://example.x402.goplausible.xyz).

## Related

- [x402 Protocol](https://github.com/coinbase/x402) — Core protocol specification
- [x402-avm](https://x402.goplausible.xyz) — Algorand x402 resources
- [Algorand Facilitator](https://facilitator.goplausible.xyz) — Public Algorand facilitator with Base and Solana support
- [Cloudflare Example](../cloudflare/) — Same server running on Cloudflare Workers
- [Hono Example](../hono/) — Same server running on Node.js with Hono
- [SST Docs](https://sst.dev/docs) — SST v3 documentation

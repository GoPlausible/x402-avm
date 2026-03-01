# AGENTS.md

## X402 Development

x402 is an HTTP-native payment protocol built on the HTTP 402 "Payment Required" status code. Three components work together: **Client** requests a protected resource, **Server** responds with 402 and structured payment requirements, and **Facilitator** verifies and settles the payment on-chain. The client signs a transaction, retries the request with an `PAYMENT-SIGNATURE` header, and the server forwards it to the facilitator for verification and settlement before granting access.

Algorand (AVM) is a **first-class citizen** alongside EVM (Ethereum) and SVM (Solana) — never conditional, always registered unconditionally. It uses CAIP-2 network identifiers and supports fee abstraction (facilitator pays transaction fees), ASA payments (USDC, ALGO), atomic transaction groups, and 3.3-second finality.

### Payment Flow

```
Client                  Resource Server           Facilitator           Algorand
  |                          |                        |                    |
  | 1. GET /api/data         |                        |                    |
  |------------------------->|                        |                    |
  | 2. 402 + requirements    |                        |                    |
  |<-------------------------|                        |                    |
  | 3. Build + sign txn      |                        |                    |
  | 4. GET + PAYMENT-SIGNATURE header|                        |                    |
  |------------------------->| 5. verify(payload)     |                    |
  |                          |----------------------->| 6. simulate_group  |
  |                          |                        |------------------->|
  |                          |                        |<-------------------|
  |                          |<-----------------------| {isValid: true}    |
  |                          | 7. settle(payload)     |                    |
  |                          |----------------------->| 8. sign + send     |
  |                          |                        |------------------->|
  |                          |                        |<-------------------| txId
  |                          |<-----------------------|                    |
  | 9. 200 + data            |                        |                    |
  |<-------------------------|                        |                    |
```

### CAIP-2 Network Identifiers

| Network          | Identifier                                              |
| ---------------- | ------------------------------------------------------- |
| Algorand Testnet | `algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=` |
| Algorand Mainnet | `algorand:wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8=` |

### X402 Components

**Client** — Wraps HTTP clients (fetch/axios/httpx/requests) to automatically handle 402 responses. Builds Algorand transaction groups using `ClientAvmSigner`, signs them, encodes as `PAYMENT-SIGNATURE` header, and retries the request.

**Resource Server** — Middleware for Express/Hono/Next.js/FastAPI/Flask that intercepts requests to protected routes. Returns 402 with `PaymentRequirements` (scheme, network, payTo, price, asset) when no payment header is present. Forwards valid payments to facilitator for verification and settlement.

**Facilitator** — On-chain payment processor implementing `FacilitatorAvmSigner` protocol: `simulate_group()` to verify transaction structure, `sign_group()` for fee abstraction (signs fee payer transactions), `send_group()` to submit atomic groups, and `confirm_transaction()` to wait for finality. Public facilitator: `https://facilitator.goplausible.xyz`

**Paywall** — Browser UI component (`@x402-avm/paywall`) for manual payment when automatic client payment isn't available. Renders a payment form and handles wallet interaction.

**Bazaar Extension** — Discovery extension (`@x402-avm/extensions` / `x402-avm[extensions]`) that registers facilitator and server capabilities for API cataloging. Clients can query available facilitators, their supported networks, and pricing through standardized endpoints.

### Signer Protocols (Core Architecture)

Protocol definitions live in the SDK; implementations are provided by users/examples:

| Protocol               | Role        | Key Methods                                                                                              |
| ---------------------- | ----------- | -------------------------------------------------------------------------------------------------------- |
| `ClientAvmSigner`      | Client-side | `address`, `sign_transactions(unsigned_txns, indexes_to_sign)`                                           |
| `FacilitatorAvmSigner` | Facilitator | `get_addresses`, `sign_transaction`, `sign_group`, `simulate_group`, `send_group`, `confirm_transaction` |

### Packages

**TypeScript** (npm):

| Package                | Purpose                                            |
| ---------------------- | -------------------------------------------------- |
| `@x402-avm/core`       | Base protocol (client, server, facilitator, types) |
| `@x402-avm/avm`        | Algorand mechanism + constants                     |
| `@x402-avm/evm`        | Ethereum mechanism                                 |
| `@x402-avm/svm`        | Solana mechanism                                   |
| `@x402-avm/express`    | Express.js server middleware                       |
| `@x402-avm/hono`       | Hono server middleware                             |
| `@x402-avm/next`       | Next.js middleware (paymentProxy)                  |
| `@x402-avm/fetch`      | Fetch API client wrapper                           |
| `@x402-avm/axios`      | Axios client wrapper                               |
| `@x402-avm/paywall`    | Browser paywall UI                                 |
| `@x402-avm/extensions` | Extensions (Bazaar discovery)                      |

**Python** (pip) — single package `x402-avm` with extras:

| Install Extra          | Purpose                        |
| ---------------------- | ------------------------------ |
| `x402-avm[avm]`        | Algorand mechanism + constants |
| `x402-avm[evm]`        | Ethereum mechanism             |
| `x402-avm[svm]`        | Solana mechanism               |
| `x402-avm[fastapi]`    | FastAPI server middleware      |
| `x402-avm[flask]`      | Flask server middleware        |
| `x402-avm[httpx]`      | Async HTTP client (httpx)      |
| `x402-avm[requests]`   | Sync HTTP client (requests)    |
| `x402-avm[extensions]` | Extensions (Bazaar discovery)  |
| `x402-avm[all]`        | Everything                     |

### X402 Skills

Skills are aggregated into two parent skills with nested references and examples:

| Skill                        | Language   | Covers                                                                                           |
| ---------------------------- | ---------- | ------------------------------------------------------------------------------------------------ |
| `algorand-x402-typescript`   | TypeScript | Clients (fetch/axios), servers (Express/Hono), Next.js, facilitators, Bazaar, paywalls, core/AVM |
| `algorand-x402-python`       | Python     | Clients (httpx/requests), servers (FastAPI/Flask), facilitators, Bazaar, core/AVM                 |

Each parent skill contains nested references organized by component. Every component has three files:
- **`{name}.md`** — Step-by-step implementation guide
- **`{name}-reference.md`** — API details and type signatures
- **`{name}-examples.md`** — Complete, runnable code samples

**TypeScript references** (under `algorand-x402-typescript`):

| Reference prefix                          | Topic                                   |
| ----------------------------------------- | --------------------------------------- |
| `explain-algorand-x402-typescript`        | Package ecosystem, signer interfaces    |
| `create-typescript-x402-client`           | Fetch/Axios clients with 402 handling   |
| `create-typescript-x402-server`           | Express/Hono server middleware          |
| `create-typescript-x402-nextjs`           | Next.js fullstack integration           |
| `create-typescript-x402-facilitator`      | Facilitator + Bazaar discovery          |
| `create-typescript-x402-paywall`          | Browser paywall UI components           |
| `use-typescript-x402-core-avm`            | Low-level core/AVM SDK usage            |

**Python references** (under `algorand-x402-python`):

| Reference prefix                          | Topic                                   |
| ----------------------------------------- | --------------------------------------- |
| `explain-algorand-x402-python`            | Package ecosystem, encoding boundaries  |
| `create-python-x402-client`               | httpx/requests clients with 402 handling|
| `create-python-x402-server`               | FastAPI/Flask server middleware          |
| `create-python-x402-facilitator`          | Facilitator + Bazaar discovery          |
| `use-python-x402-core-avm`               | Low-level core/AVM SDK usage            |

### Building X402 Applications

1. **Understand**: Load `algorand-x402-typescript` or `algorand-x402-python` — each includes explanation references covering the protocol, components, and payment flow
2. **Choose components**: Client, server, facilitator, paywall — or a subset
3. **Pick language**: TypeScript (`@x402-avm/*` packages) or Python (`x402-avm[extras]`)
4. **Navigate references**: Each parent skill lists nested references by component — read the `{name}.md` guide, consult `{name}-reference.md` for APIs, use `{name}-examples.md` for runnable code
5. **Implement signers**: `ClientAvmSigner` for clients, `FacilitatorAvmSigner` for facilitators — protocol definitions are in the SDK, implementations in your code

### Environment Variables

| Variable          | Purpose                                                |
| ----------------- | ------------------------------------------------------ |
| `AVM_PRIVATE_KEY` | Base64-encoded 64-byte key (32 seed + 32 pub)          |
| `ALGOD_SERVER`    | Custom Algod node URL (optional, defaults to AlgoNode) |
| `ALGOD_TOKEN`     | Algod node API token (optional)                        |
| `PAY_TO`          | Algorand address to receive payments (server)          |

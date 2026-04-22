# @x402/elysia

Elysia plugin for the [x402 Payment Protocol](https://x402.org). Protect any route behind an Algorand (or multi-chain) micropayment with a single `.use()` call.

Part of the [`x402-avm`](https://github.com/GoPlausible/x402-avm) ecosystem.

## Installation

```bash
pnpm add @x402/elysia
```

## Quick Start

```typescript
import Elysia from "elysia";
import { paymentMiddleware, x402ResourceServer } from "@x402/elysia";
import { ExactAvmScheme } from "@x402/avm/exact/server";
import { HTTPFacilitatorClient } from "@x402/core/server";
import { ALGORAND_TESTNET_CAIP2 } from "@x402/avm";

const facilitatorClient = new HTTPFacilitatorClient({
  url: "https://facilitator.x402.goplausible.xyz",
});

const resourceServer = new x402ResourceServer(facilitatorClient).register(
  ALGORAND_TESTNET_CAIP2,
  new ExactAvmScheme(),
);

const app = new Elysia()
  .use(
    paymentMiddleware(
      {
        "GET /api/premium": {
          accepts: {
            scheme: "exact",
            price: "$0.01",
            network: ALGORAND_TESTNET_CAIP2,
            payTo: "YOUR_ALGORAND_ADDRESS",
          },
          description: "Premium data access",
        },
      },
      resourceServer,
    ),
  )
  .get("/api/premium", () => ({ data: "secret content" }))
  .listen(3000);

console.log("Server running on http://localhost:3000");
```

## Multi-Chain Example

Accept payment from EVM, SVM, or AVM:

```typescript
import { ExactAvmScheme } from "@x402/avm/exact/server";
import { ExactEvmScheme } from "@x402/evm/exact/server";
import { ALGORAND_TESTNET_CAIP2 } from "@x402/avm";

const resourceServer = new x402ResourceServer(facilitatorClient)
  .register("eip155:84532", new ExactEvmScheme())
  .register(ALGORAND_TESTNET_CAIP2, new ExactAvmScheme());

app.use(
  paymentMiddleware(
    {
      "GET /api/data": {
        accepts: [
          {
            scheme: "exact",
            price: "$0.01",
            network: "eip155:84532",
            payTo: "0xYourEvmAddress",
          },
          {
            scheme: "exact",
            price: "$0.01",
            network: ALGORAND_TESTNET_CAIP2,
            payTo: "YOUR_ALGORAND_ADDRESS",
          },
        ],
        description: "Data endpoint",
      },
    },
    resourceServer,
  ),
);
```

## Dynamic Pricing

Use a function to compute `payTo` or `price` at request time:

```typescript
paymentMiddleware(
  {
    "GET /api/data/:tier": {
      accepts: {
        scheme: "exact",
        network: ALGORAND_TESTNET_CAIP2,
        payTo: "YOUR_ALGORAND_ADDRESS",
        price: ctx => (ctx.path.includes("premium") ? "$0.10" : "$0.01"),
      },
    },
  },
  resourceServer,
);
```

## Advanced: Custom HTTP Server

For full control (lifecycle hooks, custom paywall provider):

```typescript
import {
  paymentMiddlewareFromHTTPServer,
  x402HTTPResourceServer,
  x402ResourceServer,
} from "@x402/elysia";

const httpServer = new x402HTTPResourceServer(resourceServer, routes);

httpServer.server.onAfterSettle(async ctx => {
  console.log("Settled:", ctx.result.transaction);
});

app.use(paymentMiddlewareFromHTTPServer(httpServer, paywallConfig));
```

## Settlement Overrides

You can pass optional overrides into the settlement call from inside a route handler:

```typescript
import { setSettlementOverrides } from "@x402/elysia";

app.get("/api/data", ctx => {
  setSettlementOverrides(ctx.set, { myField: "value" });
  return { data: "secret" };
});
```

## API Reference

### `paymentMiddleware(routes, server, paywallConfig?, paywall?, syncFacilitatorOnStart?)`

Creates an Elysia plugin that protects matching routes.

| Parameter | Type | Description |
|-----------|------|-------------|
| `routes` | `RoutesConfig` | Map of route patterns to payment config |
| `server` | `x402ResourceServer` | Resource server with registered schemes |
| `paywallConfig` | `PaywallConfig?` | Optional paywall UI config (app name/logo) |
| `paywall` | `PaywallProvider?` | Custom paywall HTML provider |
| `syncFacilitatorOnStart` | `boolean` | Sync with facilitator on startup (default: `true`) |

### `paymentMiddlewareFromHTTPServer(httpServer, paywallConfig?, paywall?, syncFacilitatorOnStart?)`

Lower-level variant that accepts a pre-configured `x402HTTPResourceServer`.

### `paymentMiddlewareFromConfig(routes, facilitatorClients?, schemes?, paywallConfig?, paywall?, syncFacilitatorOnStart?)`

Convenience variant that creates the `x402ResourceServer` internally.

### `ElysiaAdapter`

Implements `HTTPAdapter` from `@x402/core`. Wraps an Elysia `Context` for use with x402 core internals.

### `setSettlementOverrides(set, overrides)`

Attaches settlement overrides to the outgoing response. The middleware strips the header before forwarding to the client.

## How It Works

```
Request → onBeforeHandle checks PAYMENT-SIGNATURE header
  ├── Route not protected → pass through
  ├── No header → return 402 with PaymentRequired JSON/HTML
  ├── Invalid payment → return 402
  └── Valid payment → pass to handler
       └── onAfterHandle → settle on-chain → attach PAYMENT-RESPONSE header
```

## License

Apache-2.0

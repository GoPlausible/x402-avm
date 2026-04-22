---
"@x402/elysia": minor
---

Add Elysia framework adapter for the x402 payment protocol.

New package `@x402/elysia` providing `paymentMiddleware`, `paymentMiddlewareFromHTTPServer`, `paymentMiddlewareFromConfig`, and `ElysiaAdapter` — matching the API surface of `@x402/hono` and supporting Algorand (AVM) and multi-chain payment schemes.

Also ships `setSettlementOverrides` / `SETTLEMENT_OVERRIDES_HEADER` / `SettlementOverrides` as forward-compat shims for settlement override support (mirroring `x402-foundation/express` v2.9.0). These are currently no-ops in this fork's `@x402/core` and will activate automatically when core adopts the upstream field.

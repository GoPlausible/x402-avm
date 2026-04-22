# @x402/elysia Changelog

## 2.6.0

### Minor Changes

- Initial release: Elysia framework adapter for the x402 payment protocol.
  - `paymentMiddleware` — Elysia plugin protecting routes with x402 payment checks.
  - `paymentMiddlewareFromHTTPServer` — Lower-level variant accepting a pre-configured `x402HTTPResourceServer`.
  - `paymentMiddlewareFromConfig` — Convenience variant that builds the resource server internally.
  - `ElysiaAdapter` — Framework adapter implementing `HTTPAdapter` from `@x402/core`.
  - `setSettlementOverrides` / `SETTLEMENT_OVERRIDES_HEADER` / `SettlementOverrides` — Forward-compat shims for settlement overrides, mirroring `x402-foundation/express` v2.9.0. Currently no-op in this fork's `@x402/core`; will activate automatically when core exports the constant and reads `responseHeaders` in `processSettlement`.
  - Updated dependencies
    - @x402/core@2.6.0
    - @x402/extensions@2.6.0

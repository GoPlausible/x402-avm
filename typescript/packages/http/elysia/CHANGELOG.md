# @x402/elysia Changelog

## 2.6.0

### Minor Changes

- Initial release: Elysia framework adapter for the x402 payment protocol.
  - `paymentMiddleware` — Elysia plugin protecting routes with x402 payment checks.
  - `paymentMiddlewareFromHTTPServer` — Lower-level variant accepting a pre-configured `x402HTTPResourceServer`.
  - `paymentMiddlewareFromConfig` — Convenience variant that builds the resource server internally.
  - `ElysiaAdapter` — Framework adapter implementing `HTTPAdapter` from `@x402/core`.
  - `setSettlementOverrides` — Helper to attach settlement overrides from route handlers.
  - Updated dependencies
    - @x402/core@2.6.0
    - @x402/extensions@2.6.0

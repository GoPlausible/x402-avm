import Elysia, { type Context } from "elysia";
import {
  x402HTTPResourceServer,
  x402ResourceServer,
  RouteConfigurationError,
  type RoutesConfig,
  type PaywallConfig,
  type PaywallProvider,
  type HTTPRequestContext,
  type HTTPTransportContext,
  type FacilitatorClient,
} from "@x402/core/server";
import type { Network, SchemeNetworkServer } from "@x402/core/types";
import { ElysiaAdapter } from "./adapter";
import { elysiaResponseToSettlementBuffer } from "./response-buffer";
import { flattenHeadersForSettlement, stripSettlementOverridesHeader } from "./settlement-headers";

/**
 * Configuration for registering a payment scheme with a specific network.
 */
export interface SchemeRegistration {
  /**
   * The network identifier (e.g. "algorand:mainnet-v1.0").
   */
  network: Network;

  /**
   * The scheme server implementation for this network.
   */
  server: SchemeNetworkServer;
}

type X402VerifiedPayload = {
  paymentPayload: unknown;
  paymentRequirements: unknown;
  declaredExtensions: unknown;
  context: HTTPRequestContext;
  httpServer: x402HTTPResourceServer;
};

/**
 * Returns true if any route in the configuration declares bazaar extensions.
 *
 * @param routes - The route configuration to inspect.
 * @returns True if at least one route has extensions.bazaar defined.
 */
function checkIfBazaarNeeded(routes: RoutesConfig): boolean {
  if ("accepts" in routes) {
    return !!(routes.extensions && "bazaar" in routes.extensions);
  }
  return Object.values(routes as Record<string, { extensions?: Record<string, unknown> }>).some(
    routeConfig => !!(routeConfig.extensions && "bazaar" in routeConfig.extensions),
  );
}

/**
 * Reads the numeric status code from an Elysia set object.
 *
 * @param ctx - A partial context object containing the set property.
 * @param ctx.set - The Elysia set object holding the mutable response status.
 * @returns The numeric HTTP status code, defaulting to 200 if unset or unparseable.
 */
function getStatusCode(ctx: { set: Context["set"] }): number {
  const s = ctx.set.status;
  if (typeof s === "number") return s;
  if (typeof s === "string") {
    const parsed = parseInt(s, 10);
    return isNaN(parsed) ? 200 : parsed;
  }
  return 200;
}

/**
 * Sends a 502 error response for facilitator boundary failures.
 *
 * @param ctx - The current Elysia context.
 * @param message - The error message to include in the response body.
 * @returns A plain object error body; Elysia serialises it as JSON.
 */
function sendFacilitatorBoundaryError(ctx: Context, message: string): Record<string, unknown> {
  ctx.set.status = 502;
  ctx.set.headers["content-type"] = "application/json";
  return { error: message };
}

/**
 * Creates an Elysia x402 payment plugin from a pre-configured x402HTTPResourceServer.
 *
 * Use this when you need full control over the HTTP server instance, for example
 * to attach lifecycle hooks or inject a custom paywall provider.
 *
 * @param httpServer - Pre-configured x402HTTPResourceServer instance.
 * @param paywallConfig - Optional paywall UI configuration (app name, logo, etc.).
 * @param paywall - Optional custom paywall HTML provider.
 * @param syncFacilitatorOnStart - Whether to sync with the facilitator on startup (default: true).
 * @returns An Elysia plugin that guards protected routes with x402 payment checks.
 *
 * @example
 * ```typescript
 * import { paymentMiddlewareFromHTTPServer, x402HTTPResourceServer } from "@x402/elysia";
 *
 * const httpServer = new x402HTTPResourceServer(resourceServer, routes)
 *   .onAfterSettle(async ctx => console.log("Settled:", ctx.result.transaction));
 *
 * app.use(paymentMiddlewareFromHTTPServer(httpServer, paywallConfig));
 * ```
 */
export function paymentMiddlewareFromHTTPServer(
  httpServer: x402HTTPResourceServer,
  paywallConfig?: PaywallConfig,
  paywall?: PaywallProvider,
  syncFacilitatorOnStart = true,
) {
  if (paywall) {
    httpServer.registerPaywallProvider(paywall);
  }

  let initPromise: Promise<void> | null = syncFacilitatorOnStart ? httpServer.initialize() : null;
  let isInitialized = false;
  let bazaarPromise: Promise<void> | null = null;

  /**
   * Waits for the httpServer initialization promise to resolve, then marks initialized.
   *
   * @returns A promise that resolves once initialization is complete.
   */
  async function initializeHttpServer(): Promise<void> {
    if (!syncFacilitatorOnStart || isInitialized) {
      return;
    }
    if (!initPromise) {
      initPromise = httpServer.initialize();
    }
    try {
      await initPromise;
      isInitialized = true;
      initPromise = null;
    } catch (error) {
      initPromise = null;
      throw error;
    }
  }

  if (checkIfBazaarNeeded(httpServer.routes) && !httpServer.server.hasExtension("bazaar")) {
    bazaarPromise = import("@x402/extensions/bazaar")
      .then(({ bazaarResourceServerExtension }) => {
        httpServer.server.registerExtension(bazaarResourceServerExtension);
      })
      .catch(err => {
        console.error("Failed to load bazaar extension:", err);
      });
  }

  return new Elysia({ name: "@x402/elysia" })
    .state("__x402", undefined as X402VerifiedPayload | undefined)
    .onBeforeHandle({ as: "global" }, async ctx => {
      const adapter = new ElysiaAdapter(ctx as unknown as Context);
      const context: HTTPRequestContext = {
        adapter,
        path: adapter.getPath(),
        method: adapter.getMethod(),
        paymentHeader: adapter.getHeader("payment-signature") || adapter.getHeader("x-payment"),
      };

      if (!httpServer.requiresPayment(context)) {
        return;
      }

      if (syncFacilitatorOnStart && !isInitialized) {
        try {
          await initializeHttpServer();
        } catch (error) {
          if (error instanceof RouteConfigurationError) {
            throw error;
          }
          return sendFacilitatorBoundaryError(
            ctx as Context,
            error instanceof Error ? error.message : String(error),
          );
        }
      }

      if (bazaarPromise) {
        await bazaarPromise;
        bazaarPromise = null;
      }

      let result: Awaited<ReturnType<typeof httpServer.processHTTPRequest>>;
      try {
        result = await httpServer.processHTTPRequest(context, paywallConfig);
      } catch (error) {
        if (error instanceof RouteConfigurationError) {
          throw error;
        }
        return sendFacilitatorBoundaryError(
          ctx as Context,
          error instanceof Error ? error.message : String(error),
        );
      }

      switch (result.type) {
        case "no-payment-required":
          return;

        case "payment-error": {
          const { response } = result;
          Object.entries(response.headers).forEach(([key, value]) => {
            ctx.set.headers[key] = value;
          });
          ctx.set.status = response.status;
          if (response.isHtml) {
            ctx.set.headers["content-type"] = "text/html";
            return response.body;
          } else {
            return response.body ?? {};
          }
        }

        case "payment-verified": {
          const { paymentPayload, paymentRequirements, declaredExtensions } = result;
          ctx.store.__x402 = {
            paymentPayload,
            paymentRequirements,
            declaredExtensions,
            context,
            httpServer,
          };
          return;
        }
      }
    })
    .onAfterHandle({ as: "global" }, async ctx => {
      const verified = ctx.store.__x402;
      if (!verified) return;

      const {
        paymentPayload,
        paymentRequirements,
        declaredExtensions,
        context,
        httpServer: server,
      } = verified;

      if (getStatusCode(ctx) >= 400) return;

      try {
        const responseBody = await elysiaResponseToSettlementBuffer(ctx.response);

        const responseHeaders = flattenHeadersForSettlement(
          ctx.set.headers as Record<string, unknown>,
        );
        stripSettlementOverridesHeader(ctx.set.headers as Record<string, unknown>);

        // responseHeaders is passed for parity with x402-foundation/express v2.9.0.
        // @x402/core's HTTPTransportContext does not yet include this field;
        // cast it through to pass it opaquely until the upstream core type is updated.
        const transportContext = {
          request: context,
          responseBody,
          responseHeaders,
        } as unknown as HTTPTransportContext;

        const settleResult = await server.processSettlement(
          paymentPayload as Parameters<typeof server.processSettlement>[0],
          paymentRequirements as Parameters<typeof server.processSettlement>[1],
          declaredExtensions as Parameters<typeof server.processSettlement>[2],
          transportContext,
        );

        if (!settleResult.success) {
          const { response } = settleResult;
          Object.entries(response.headers).forEach(([key, value]) => {
            ctx.set.headers[key] = value;
          });
          ctx.set.status = response.status;
          if (response.isHtml) {
            ctx.set.headers["content-type"] = "text/html";
            ctx.response = response.body ?? "";
          } else {
            ctx.set.headers["content-type"] = "application/json";
            ctx.response = response.body ?? {};
          }
        } else {
          Object.entries(settleResult.headers).forEach(([key, value]) => {
            ctx.set.headers[key] = value;
          });
        }
      } catch (error) {
        // Mirror x402-foundation/express: settlement infrastructure errors surface as 502.
        // FacilitatorResponseError is not yet exported by @x402/core, so all caught
        // errors are treated as facilitator boundary failures.
        console.error("[x402/elysia] Settlement error:", error);
        ctx.set.status = 502;
        ctx.set.headers["content-type"] = "application/json";
        ctx.response = { error: error instanceof Error ? error.message : String(error) };
      }
    });
}

/**
 * Creates an Elysia x402 payment plugin.
 *
 * Protects matching routes by requiring a valid x402 payment header.
 * Returns a 402 response with payment instructions when payment is missing or invalid.
 * Settles payment on-chain after a successful response.
 *
 * @param routes - Route configurations mapping path patterns to payment options.
 * @param server - Pre-configured x402ResourceServer with registered payment schemes.
 * @param paywallConfig - Optional paywall UI configuration for browser requests.
 * @param paywall - Optional custom paywall HTML provider.
 * @param syncFacilitatorOnStart - Whether to sync with facilitator on startup (default: true).
 * @returns An Elysia plugin that guards protected routes with x402 payment checks.
 *
 * @example
 * ```typescript
 * import Elysia from "elysia";
 * import { paymentMiddleware, x402ResourceServer } from "@x402/elysia";
 *
 * const server = new x402ResourceServer(facilitatorClient)
 *   .register(NETWORK, new ExactAvmScheme());
 *
 * app.use(paymentMiddleware({ "GET /api/data": { accepts: ... } }, server));
 * ```
 */
export function paymentMiddleware(
  routes: RoutesConfig,
  server: x402ResourceServer,
  paywallConfig?: PaywallConfig,
  paywall?: PaywallProvider,
  syncFacilitatorOnStart = true,
) {
  const httpServer = new x402HTTPResourceServer(server, routes);
  return paymentMiddlewareFromHTTPServer(
    httpServer,
    paywallConfig,
    paywall,
    syncFacilitatorOnStart,
  );
}

/**
 * Creates an Elysia x402 payment plugin from a declarative configuration object.
 *
 * Convenience wrapper that creates the x402ResourceServer internally.
 * Suitable for simple setups where you do not need a reference to the resource server.
 *
 * @param routes - Route configurations for protected endpoints.
 * @param facilitatorClients - Optional facilitator client(s) for payment verification.
 * @param schemes - Optional array of scheme registrations for server-side payment processing.
 * @param paywallConfig - Optional paywall UI configuration.
 * @param paywall - Optional custom paywall provider.
 * @param syncFacilitatorOnStart - Whether to sync with facilitator on startup (default: true).
 * @returns An Elysia plugin that guards protected routes with x402 payment checks.
 *
 * @example
 * ```typescript
 * import { paymentMiddlewareFromConfig } from "@x402/elysia";
 *
 * app.use(paymentMiddlewareFromConfig(
 *   routes,
 *   myFacilitatorClient,
 *   [{ network: "algorand:mainnet-v1.0", server: avmSchemeServer }],
 * ));
 * ```
 */
export function paymentMiddlewareFromConfig(
  routes: RoutesConfig,
  facilitatorClients?: FacilitatorClient | FacilitatorClient[],
  schemes?: SchemeRegistration[],
  paywallConfig?: PaywallConfig,
  paywall?: PaywallProvider,
  syncFacilitatorOnStart = true,
) {
  const ResourceServer = new x402ResourceServer(facilitatorClients);
  if (schemes) {
    schemes.forEach(({ network, server: schemeServer }) => {
      ResourceServer.register(network, schemeServer);
    });
  }
  return paymentMiddleware(routes, ResourceServer, paywallConfig, paywall, syncFacilitatorOnStart);
}

export { RouteConfigurationError };

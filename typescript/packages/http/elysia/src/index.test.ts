import { describe, it, expect, vi, beforeEach } from "vitest";
import Elysia from "elysia";
import type {
  HTTPProcessResult,
  x402HTTPResourceServer,
  PaywallProvider,
  FacilitatorClient,
} from "@x402/core/server";
import {
  x402ResourceServer,
  x402HTTPResourceServer as HTTPResourceServer,
} from "@x402/core/server";
import type { PaymentPayload, PaymentRequirements, SchemeNetworkServer } from "@x402/core/types";
import { paymentMiddleware, paymentMiddlewareFromConfig, type SchemeRegistration } from "./index";

// --- Test Fixtures ---
const mockRoutes = {
  "GET /api/*": {
    accepts: {
      scheme: "exact",
      payTo: "ALGORAND_ADDRESS",
      price: "$0.01",
      network: "algorand:mainnet-v1.0",
    },
  },
} as const;

const mockPaymentPayload = {
  scheme: "exact",
  network: "algorand:mainnet-v1.0",
  payload: { signature: "abc" },
} as unknown as PaymentPayload;

const mockPaymentRequirements = {
  scheme: "exact",
  network: "algorand:mainnet-v1.0",
  maxAmountRequired: "1000",
  payTo: "ALGORAND_ADDRESS",
} as unknown as PaymentRequirements;

// --- Mock setup ---
let mockProcessHTTPRequest: ReturnType<typeof vi.fn>;
let mockProcessSettlement: ReturnType<typeof vi.fn>;
let mockRegisterPaywallProvider: ReturnType<typeof vi.fn>;
let mockRequiresPayment: ReturnType<typeof vi.fn>;

vi.mock("@x402/core/server", () => ({
  x402ResourceServer: vi.fn().mockImplementation(() => ({
    initialize: vi.fn().mockResolvedValue(undefined),
    registerExtension: vi.fn(),
    register: vi.fn(),
    hasExtension: vi.fn().mockReturnValue(false),
  })),
  x402HTTPResourceServer: vi.fn().mockImplementation((server, routes) => ({
    initialize: vi.fn().mockResolvedValue(undefined),
    processHTTPRequest: mockProcessHTTPRequest,
    processSettlement: mockProcessSettlement,
    registerPaywallProvider: mockRegisterPaywallProvider,
    requiresPayment: mockRequiresPayment,
    routes: routes,
    server: server || {
      hasExtension: vi.fn().mockReturnValue(false),
      registerExtension: vi.fn(),
    },
  })),
  RouteConfigurationError: class RouteConfigurationError extends Error {},
}));

/**
 * Configures the mock HTTP server to return the given request and settlement results.
 *
 * @param processResult - The result to return from processHTTPRequest.
 * @param settlementResult - The result to return from processSettlement.
 */
function setupMockHttpServer(
  processResult: HTTPProcessResult,
  settlementResult:
    | { success: true; headers: Record<string, string> }
    | {
        success: false;
        errorReason: string;
        headers: Record<string, string>;
        response: {
          status: number;
          headers: Record<string, string>;
          body?: unknown;
          isHtml?: boolean;
        };
      } = {
    success: true,
    headers: {},
  },
): void {
  mockProcessHTTPRequest.mockResolvedValue(processResult);
  mockProcessSettlement.mockResolvedValue(settlementResult);
}

/**
 * Builds a minimal Elysia test app with the payment plugin and a single GET route.
 *
 * @param path - The route path to register.
 * @param handler - The route handler function.
 * @returns An Elysia app ready to handle requests.
 */
function buildApp(path = "/api/test", handler: () => unknown = () => ({ ok: true })): Elysia {
  return new Elysia()
    .use(
      paymentMiddleware(
        mockRoutes as never,
        {} as unknown as x402ResourceServer,
        undefined,
        undefined,
        false,
      ),
    )
    .get(path, handler);
}

// --- Tests ---
describe("paymentMiddleware", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockProcessHTTPRequest = vi.fn();
    mockProcessSettlement = vi.fn();
    mockRegisterPaywallProvider = vi.fn();
    mockRequiresPayment = vi.fn().mockReturnValue(true);

    vi.mocked(HTTPResourceServer).mockImplementation(
      (server, routes) =>
        ({
          processHTTPRequest: mockProcessHTTPRequest,
          processSettlement: mockProcessSettlement,
          registerPaywallProvider: mockRegisterPaywallProvider,
          requiresPayment: mockRequiresPayment,
          routes: routes,
          server: server || {
            hasExtension: vi.fn().mockReturnValue(false),
            registerExtension: vi.fn(),
          },
        }) as unknown as x402HTTPResourceServer,
    );
  });

  it("passes through when no-payment-required", async () => {
    setupMockHttpServer({ type: "no-payment-required" });

    const app = buildApp();
    const res = await app.handle(new Request("http://localhost/api/test"));

    expect(res.status).toBe(200);
    expect(mockProcessHTTPRequest).toHaveBeenCalled();
  });

  it("returns 402 JSON for payment-error", async () => {
    setupMockHttpServer({
      type: "payment-error",
      response: {
        status: 402,
        body: { error: "Payment required" },
        headers: {},
        isHtml: false,
      },
    });

    const app = buildApp();
    const res = await app.handle(new Request("http://localhost/api/test"));

    expect(res.status).toBe(402);
    const body = await res.json();
    expect(body).toEqual({ error: "Payment required" });
  });

  it("returns 402 HTML for payment-error with isHtml", async () => {
    setupMockHttpServer({
      type: "payment-error",
      response: {
        status: 402,
        body: "<html>Paywall</html>",
        headers: {},
        isHtml: true,
      },
    });

    const app = buildApp();
    const res = await app.handle(new Request("http://localhost/api/test"));

    expect(res.status).toBe(402);
    const text = await res.text();
    expect(text).toContain("Paywall");
    expect(res.headers.get("content-type")).toContain("text/html");
  });

  it("sets custom headers from payment-error response", async () => {
    setupMockHttpServer({
      type: "payment-error",
      response: {
        status: 402,
        body: { error: "Payment required" },
        headers: { "X-Custom-Header": "custom-value" },
        isHtml: false,
      },
    });

    const app = buildApp();
    const res = await app.handle(new Request("http://localhost/api/test"));

    expect(res.headers.get("X-Custom-Header")).toBe("custom-value");
  });

  it("settles and attaches PAYMENT-RESPONSE header on success", async () => {
    setupMockHttpServer(
      {
        type: "payment-verified",
        paymentPayload: mockPaymentPayload,
        paymentRequirements: mockPaymentRequirements,
        declaredExtensions: undefined,
      },
      { success: true, headers: { "PAYMENT-RESPONSE": "settled-encoded" } },
    );

    const app = buildApp();
    const res = await app.handle(new Request("http://localhost/api/test"));

    expect(res.status).toBe(200);
    expect(mockProcessSettlement).toHaveBeenCalledWith(
      mockPaymentPayload,
      mockPaymentRequirements,
      undefined,
      expect.objectContaining({
        request: expect.objectContaining({
          path: "/api/test",
          method: "GET",
        }),
        responseBody: expect.any(Buffer),
      }),
    );
    expect(res.headers.get("PAYMENT-RESPONSE")).toBe("settled-encoded");
  });

  it("skips settlement when handler returns >= 400", async () => {
    setupMockHttpServer(
      {
        type: "payment-verified",
        paymentPayload: mockPaymentPayload,
        paymentRequirements: mockPaymentRequirements,
        declaredExtensions: undefined,
      },
      { success: true, headers: {} },
    );

    const app = new Elysia()
      .use(
        paymentMiddleware(
          mockRoutes as never,
          {} as unknown as x402ResourceServer,
          undefined,
          undefined,
          false,
        ),
      )
      .get("/api/test", ({ set }) => {
        set.status = 500;
        return { error: "server error" };
      });

    await app.handle(new Request("http://localhost/api/test"));

    expect(mockProcessSettlement).not.toHaveBeenCalled();
  });

  it("returns 402 when settlement returns success: false", async () => {
    setupMockHttpServer(
      {
        type: "payment-verified",
        paymentPayload: mockPaymentPayload,
        paymentRequirements: mockPaymentRequirements,
        declaredExtensions: undefined,
      },
      {
        success: false,
        errorReason: "Insufficient funds",
        headers: { "PAYMENT-RESPONSE": "failure-encoded" },
        response: {
          status: 402,
          headers: { "PAYMENT-RESPONSE": "failure-encoded" },
          body: { error: "Settlement failed" },
          isHtml: false,
        },
      },
    );

    const app = buildApp();
    const res = await app.handle(new Request("http://localhost/api/test"));

    expect(res.status).toBe(402);
    expect(res.headers.get("PAYMENT-RESPONSE")).toBe("failure-encoded");
    const body = await res.json();
    expect(body).toEqual({ error: "Settlement failed" });
  });

  it("passes paywallConfig to processHTTPRequest", async () => {
    setupMockHttpServer({ type: "no-payment-required" });
    const paywallConfig = { appName: "test-app" };

    const app = new Elysia()
      .use(
        paymentMiddleware(
          mockRoutes as never,
          {} as unknown as x402ResourceServer,
          paywallConfig,
          undefined,
          false,
        ),
      )
      .get("/api/test", () => ({ ok: true }));

    await app.handle(new Request("http://localhost/api/test"));

    expect(mockProcessHTTPRequest).toHaveBeenCalledWith(expect.anything(), paywallConfig);
  });

  it("registers custom paywall provider", () => {
    setupMockHttpServer({ type: "no-payment-required" });
    const paywall: PaywallProvider = { generateHtml: vi.fn() };

    paymentMiddleware(
      mockRoutes as never,
      {} as unknown as x402ResourceServer,
      undefined,
      paywall,
      false,
    );

    expect(mockRegisterPaywallProvider).toHaveBeenCalledWith(paywall);
  });
});

describe("paymentMiddlewareFromConfig", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockProcessHTTPRequest = vi.fn();
    mockProcessSettlement = vi.fn();
    mockRegisterPaywallProvider = vi.fn();
    mockRequiresPayment = vi.fn().mockReturnValue(true);

    vi.mocked(HTTPResourceServer).mockImplementation(
      (server, routes) =>
        ({
          initialize: vi.fn().mockResolvedValue(undefined),
          processHTTPRequest: mockProcessHTTPRequest,
          processSettlement: mockProcessSettlement,
          registerPaywallProvider: mockRegisterPaywallProvider,
          requiresPayment: mockRequiresPayment,
          routes: routes,
          server: server || {
            hasExtension: vi.fn().mockReturnValue(false),
            registerExtension: vi.fn(),
          },
        }) as unknown as x402HTTPResourceServer,
    );

    vi.mocked(x402ResourceServer).mockImplementation(
      () =>
        ({
          initialize: vi.fn().mockResolvedValue(undefined),
          registerExtension: vi.fn(),
          register: vi.fn(),
        }) as unknown as x402ResourceServer,
    );
  });

  it("creates x402ResourceServer with facilitator clients", () => {
    setupMockHttpServer({ type: "no-payment-required" });
    const facilitator = { verify: vi.fn(), settle: vi.fn() } as unknown as FacilitatorClient;

    paymentMiddlewareFromConfig(mockRoutes as never, facilitator);

    expect(x402ResourceServer).toHaveBeenCalledWith(facilitator);
  });

  it("registers scheme servers for each network", () => {
    setupMockHttpServer({ type: "no-payment-required" });
    const schemeServer = { verify: vi.fn(), settle: vi.fn() } as unknown as SchemeNetworkServer;
    const schemes: SchemeRegistration[] = [
      { network: "algorand:mainnet-v1.0", server: schemeServer },
      { network: "eip155:8453", server: schemeServer },
    ];

    paymentMiddlewareFromConfig(mockRoutes as never, undefined, schemes);

    const serverInstance = vi.mocked(x402ResourceServer).mock.results[0].value;
    expect(serverInstance.register).toHaveBeenCalledTimes(2);
    expect(serverInstance.register).toHaveBeenCalledWith("algorand:mainnet-v1.0", schemeServer);
    expect(serverInstance.register).toHaveBeenCalledWith("eip155:8453", schemeServer);
  });

  it("returns a working Elysia plugin", async () => {
    setupMockHttpServer({ type: "no-payment-required" });

    const plugin = paymentMiddlewareFromConfig(mockRoutes as never);
    const app = new Elysia().use(plugin).get("/api/test", () => ({ ok: true }));
    const res = await app.handle(new Request("http://localhost/api/test"));

    expect(res.status).toBe(200);
    expect(mockProcessHTTPRequest).toHaveBeenCalled();
  });
});

describe("ElysiaAdapter (via middleware integration)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockProcessHTTPRequest = vi.fn();
    mockProcessSettlement = vi.fn();
    mockRegisterPaywallProvider = vi.fn();
    mockRequiresPayment = vi.fn().mockReturnValue(true);

    vi.mocked(HTTPResourceServer).mockImplementation(
      (server, routes) =>
        ({
          processHTTPRequest: mockProcessHTTPRequest,
          processSettlement: mockProcessSettlement,
          registerPaywallProvider: mockRegisterPaywallProvider,
          requiresPayment: mockRequiresPayment,
          routes: routes,
          server: server || {
            hasExtension: vi.fn().mockReturnValue(false),
            registerExtension: vi.fn(),
          },
        }) as unknown as x402HTTPResourceServer,
    );
  });

  it("extracts path and method from request", async () => {
    setupMockHttpServer({ type: "no-payment-required" });

    const app = new Elysia()
      .use(
        paymentMiddleware(
          mockRoutes as never,
          {} as unknown as x402ResourceServer,
          undefined,
          undefined,
          false,
        ),
      )
      .post("/api/weather", () => ({ ok: true }));

    await app.handle(new Request("http://localhost/api/weather", { method: "POST" }));

    expect(mockProcessHTTPRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        path: "/api/weather",
        method: "POST",
      }),
      undefined,
    );
  });

  it("extracts payment-signature header", async () => {
    setupMockHttpServer({ type: "no-payment-required" });

    const app = buildApp();
    await app.handle(
      new Request("http://localhost/api/test", {
        headers: { "payment-signature": "sig-data" },
      }),
    );

    expect(mockProcessHTTPRequest).toHaveBeenCalledWith(
      expect.objectContaining({ paymentHeader: "sig-data" }),
      undefined,
    );
  });

  it("extracts x-payment header", async () => {
    setupMockHttpServer({ type: "no-payment-required" });

    const app = buildApp();
    await app.handle(
      new Request("http://localhost/api/test", {
        headers: { "x-payment": "payment-data" },
      }),
    );

    expect(mockProcessHTTPRequest).toHaveBeenCalledWith(
      expect.objectContaining({ paymentHeader: "payment-data" }),
      undefined,
    );
  });

  it("prefers payment-signature over x-payment", async () => {
    setupMockHttpServer({ type: "no-payment-required" });

    const app = buildApp();
    await app.handle(
      new Request("http://localhost/api/test", {
        headers: { "payment-signature": "sig-data", "x-payment": "x-payment-data" },
      }),
    );

    expect(mockProcessHTTPRequest).toHaveBeenCalledWith(
      expect.objectContaining({ paymentHeader: "sig-data" }),
      undefined,
    );
  });

  it("passes undefined paymentHeader when no payment headers present", async () => {
    setupMockHttpServer({ type: "no-payment-required" });

    const app = buildApp();
    await app.handle(new Request("http://localhost/api/test"));

    expect(mockProcessHTTPRequest).toHaveBeenCalledWith(
      expect.objectContaining({ paymentHeader: undefined }),
      undefined,
    );
  });
});

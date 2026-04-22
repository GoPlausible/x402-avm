import { describe, it, expect } from "vitest";
import type { Context } from "elysia";
import { ElysiaAdapter } from "./adapter";

/**
 * Creates a minimal mock Elysia Context backed by a real Web API Request.
 *
 * @param options - Configuration options for the mock context.
 * @param options.url - The full request URL.
 * @param options.method - The HTTP method.
 * @param options.headers - Request headers as a plain record.
 * @param options.body - Optional request body (will be JSON-serialized).
 * @returns A partial Context that ElysiaAdapter can operate on.
 */
function createMockContext(
  options: {
    url?: string;
    method?: string;
    headers?: Record<string, string>;
    body?: unknown;
  } = {},
): Context {
  const url = options.url || "https://example.com/api/test";
  const method = options.method || "GET";
  const headers = new Headers(options.headers);

  const init: RequestInit = { method, headers };
  if (options.body !== undefined && method !== "GET" && method !== "HEAD") {
    init.body = JSON.stringify(options.body);
  }

  return {
    request: new Request(url, init),
    set: { headers: {}, status: undefined },
  } as unknown as Context;
}

describe("ElysiaAdapter", () => {
  describe("getHeader", () => {
    it("returns header value when present", () => {
      const ctx = createMockContext({ headers: { "x-payment": "test-payment" } });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getHeader("x-payment")).toBe("test-payment");
    });

    it("returns undefined for missing headers", () => {
      const ctx = createMockContext();
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getHeader("x-missing")).toBeUndefined();
    });

    it("is case-insensitive", () => {
      const ctx = createMockContext({ headers: { "payment-signature": "sig-abc" } });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getHeader("Payment-Signature")).toBe("sig-abc");
    });
  });

  describe("getMethod", () => {
    it("returns the HTTP method", () => {
      const ctx = createMockContext({ method: "POST" });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getMethod()).toBe("POST");
    });

    it("defaults to GET", () => {
      const ctx = createMockContext();
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getMethod()).toBe("GET");
    });
  });

  describe("getPath", () => {
    it("returns the pathname without query string", () => {
      const ctx = createMockContext({ url: "https://example.com/api/weather?city=NYC" });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getPath()).toBe("/api/weather");
    });

    it("returns root path for bare domain", () => {
      const ctx = createMockContext({ url: "https://example.com/" });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getPath()).toBe("/");
    });
  });

  describe("getUrl", () => {
    it("returns the full URL including query string", () => {
      const ctx = createMockContext({ url: "https://example.com/api/test?foo=bar" });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getUrl()).toBe("https://example.com/api/test?foo=bar");
    });
  });

  describe("getAcceptHeader", () => {
    it("returns Accept header when present", () => {
      const ctx = createMockContext({ headers: { accept: "text/html" } });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getAcceptHeader()).toBe("text/html");
    });

    it("returns empty string when missing", () => {
      const ctx = createMockContext();
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getAcceptHeader()).toBe("");
    });
  });

  describe("getUserAgent", () => {
    it("returns User-Agent header when present", () => {
      const ctx = createMockContext({ headers: { "user-agent": "Mozilla/5.0" } });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getUserAgent()).toBe("Mozilla/5.0");
    });

    it("returns empty string when missing", () => {
      const ctx = createMockContext();
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getUserAgent()).toBe("");
    });
  });

  describe("getQueryParams", () => {
    it("returns all query parameters as a record", () => {
      const ctx = createMockContext({
        url: "https://example.com/api/test?foo=bar&baz=qux",
      });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getQueryParams()).toEqual({ foo: "bar", baz: "qux" });
    });

    it("returns repeated param values as an array", () => {
      const ctx = createMockContext({
        url: "https://example.com/api/test?tag=a&tag=b",
      });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getQueryParams()).toEqual({ tag: ["a", "b"] });
    });

    it("returns empty object when no query params", () => {
      const ctx = createMockContext({ url: "https://example.com/api/test" });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getQueryParams()).toEqual({});
    });
  });

  describe("getQueryParam", () => {
    it("returns string value for single param", () => {
      const ctx = createMockContext({
        url: "https://example.com/api/test?city=NYC",
      });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getQueryParam("city")).toBe("NYC");
    });

    it("returns array for repeated param", () => {
      const ctx = createMockContext({
        url: "https://example.com/api/test?tag=a&tag=b",
      });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getQueryParam("tag")).toEqual(["a", "b"]);
    });

    it("returns undefined for missing param", () => {
      const ctx = createMockContext({ url: "https://example.com/api/test" });
      const adapter = new ElysiaAdapter(ctx);
      expect(adapter.getQueryParam("missing")).toBeUndefined();
    });
  });

  describe("getBody", () => {
    it("returns parsed JSON body", async () => {
      const body = { data: "test" };
      const ctx = createMockContext({
        url: "https://example.com/api/test",
        method: "POST",
        headers: { "content-type": "application/json" },
        body,
      });
      const adapter = new ElysiaAdapter(ctx);
      expect(await adapter.getBody()).toEqual(body);
    });

    it("returns undefined when body is not valid JSON", async () => {
      const request = new Request("https://example.com/api/test", {
        method: "POST",
        body: "not json {{",
        headers: { "content-type": "text/plain" },
      });
      const ctx = { request, set: { headers: {}, status: undefined } } as unknown as Context;
      const adapter = new ElysiaAdapter(ctx);
      expect(await adapter.getBody()).toBeUndefined();
    });

    it("returns undefined for GET request with no body", async () => {
      const ctx = createMockContext({ method: "GET" });
      const adapter = new ElysiaAdapter(ctx);
      expect(await adapter.getBody()).toBeUndefined();
    });
  });
});

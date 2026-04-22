import type { Context } from "elysia";
import type { HTTPAdapter } from "@x402/core/server";

/**
 * Elysia HTTP adapter for x402 payment middleware.
 *
 * Wraps an Elysia Context to implement the framework-agnostic HTTPAdapter interface,
 * allowing x402 core to read request data without coupling to Elysia internals.
 */
export class ElysiaAdapter implements HTTPAdapter {
  /**
   * Creates a new ElysiaAdapter instance.
   *
   * @param ctx - The Elysia context object for the current request.
   */
  constructor(private ctx: Context) {}

  /**
   * Gets a header value from the request.
   *
   * @param name - The header name (case-insensitive per HTTP spec).
   * @returns The header value or undefined if not present.
   */
  getHeader(name: string): string | undefined {
    return this.ctx.request.headers.get(name) ?? undefined;
  }

  /**
   * Gets the HTTP method of the request.
   *
   * @returns The HTTP method string (e.g. "GET", "POST").
   */
  getMethod(): string {
    return this.ctx.request.method;
  }

  /**
   * Gets the pathname of the request URL.
   *
   * @returns The URL pathname without query string.
   */
  getPath(): string {
    return new URL(this.ctx.request.url).pathname;
  }

  /**
   * Gets the full URL of the request including query string.
   *
   * @returns The full request URL string.
   */
  getUrl(): string {
    return this.ctx.request.url;
  }

  /**
   * Gets the Accept header from the request.
   *
   * @returns The Accept header value or an empty string if not present.
   */
  getAcceptHeader(): string {
    return this.ctx.request.headers.get("accept") ?? "";
  }

  /**
   * Gets the User-Agent header from the request.
   *
   * @returns The User-Agent header value or an empty string if not present.
   */
  getUserAgent(): string {
    return this.ctx.request.headers.get("user-agent") ?? "";
  }

  /**
   * Gets all query parameters from the request URL.
   *
   * @returns A record mapping parameter names to their value(s).
   */
  getQueryParams(): Record<string, string | string[]> {
    const url = new URL(this.ctx.request.url);
    const result: Record<string, string | string[]> = {};
    url.searchParams.forEach((value, key) => {
      const existing = result[key];
      if (existing === undefined) {
        result[key] = value;
      } else if (Array.isArray(existing)) {
        existing.push(value);
      } else {
        result[key] = [existing, value];
      }
    });
    return result;
  }

  /**
   * Gets a single query parameter value by name.
   *
   * @param name - The query parameter name.
   * @returns The parameter value, an array of values if repeated, or undefined if absent.
   */
  getQueryParam(name: string): string | string[] | undefined {
    const url = new URL(this.ctx.request.url);
    const values = url.searchParams.getAll(name);
    if (values.length === 0) return undefined;
    return values.length === 1 ? values[0] : values;
  }

  /**
   * Gets the parsed JSON body of the request.
   *
   * @returns The parsed body or undefined if parsing fails.
   */
  async getBody(): Promise<unknown> {
    try {
      return await this.ctx.request.clone().json();
    } catch {
      return undefined;
    }
  }
}

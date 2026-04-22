import { ElysiaCustomStatusResponse } from "elysia/error";

/**
 * Derives the byte payload used for x402 settlement from an Elysia handler return value.
 *
 * Mirrors Hono's `Buffer.from(await res.clone().arrayBuffer())` pattern while
 * handling the full range of types that Elysia's internal `mapResponse` may produce,
 * including strings, JSON objects, streams, Blobs, and Elysia status wrappers.
 *
 * @param value - The raw value returned by an Elysia route handler.
 * @returns A Buffer containing the serialized response body bytes.
 */
export async function elysiaResponseToSettlementBuffer(value: unknown): Promise<Buffer> {
  if (value === undefined || value === null) {
    return Buffer.alloc(0);
  }

  if (value instanceof ElysiaCustomStatusResponse) {
    return elysiaResponseToSettlementBuffer(value.response);
  }

  if (value instanceof Response) {
    return Buffer.from(await value.clone().arrayBuffer());
  }

  if (typeof value === "string") {
    return Buffer.from(value, "utf8");
  }

  if (typeof value === "number" || typeof value === "boolean") {
    return Buffer.from(String(value), "utf8");
  }

  if (value instanceof Uint8Array) {
    return Buffer.from(value);
  }

  if (value instanceof ArrayBuffer) {
    return Buffer.from(value);
  }

  if (typeof Blob !== "undefined" && value instanceof Blob) {
    return Buffer.from(await value.arrayBuffer());
  }

  if (value instanceof ReadableStream) {
    const reader = value.getReader();
    const chunks: Buffer[] = [];
    try {
      while (true) {
        const { done, value: chunk } = await reader.read();
        if (done) break;
        if (chunk) chunks.push(Buffer.from(chunk));
      }
    } finally {
      reader.releaseLock();
    }
    return chunks.length ? Buffer.concat(chunks) : Buffer.alloc(0);
  }

  if (value instanceof FormData) {
    return Buffer.from(await new Response(value).arrayBuffer());
  }

  const ctor = (value as { constructor?: { name?: string } }).constructor?.name;
  if (ctor === "Array" || ctor === "Object" || Array.isArray(value)) {
    return Buffer.from(JSON.stringify(value), "utf8");
  }

  if (typeof value === "object" && value !== null) {
    try {
      return Buffer.from(JSON.stringify(value), "utf8");
    } catch {
      /* fall through */
    }
  }

  return Buffer.from(String(value), "utf8");
}

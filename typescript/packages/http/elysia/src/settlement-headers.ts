/**
 * Forward-compat shim — settlement overrides helpers.
 *
 * `@x402/core` in this fork does not yet consume the `Settlement-Overrides`
 * header nor the `responseHeaders` field on `HTTPTransportContext`. Calls to
 * `setSettlementOverrides` write a header that the middleware reads and strips
 * before the response ships, but the value is never forwarded to
 * `processSettlement`. The API surface is kept stable for forward compatibility
 * with `x402-foundation/express` v2.9.0 and will activate automatically once
 * upstream core exports `SETTLEMENT_OVERRIDES_HEADER` and reads `responseHeaders`
 * in `processSettlement`.
 *
 * When upstream exports the constant, replace this file with:
 *   export { SETTLEMENT_OVERRIDES_HEADER, type SettlementOverrides } from "@x402/core/server";
 */

/**
 * Header used to pass partial settlement overrides, matching the x402-foundation/x402 convention.
 *
 * Defined locally because @x402/core does not yet export this constant.
 * Forward-compat shim — no-op in x402-avm today; see file-level note above.
 */
export const SETTLEMENT_OVERRIDES_HEADER = "Settlement-Overrides";

/**
 * A map of optional override values for the settlement call.
 *
 * Forward-compat shim — no-op in x402-avm today; see file-level note above.
 */
export type SettlementOverrides = Record<string, string | number | undefined>;

/**
 * Attaches settlement overrides to the outgoing response headers before settlement runs.
 *
 * The Elysia middleware strips this header from the wire response after reading it,
 * so it is never exposed to the client. The value is currently not forwarded to
 * `processSettlement` by this fork's `@x402/core`.
 *
 * Forward-compat shim — no-op in x402-avm today; see file-level note above.
 *
 * @param set - The Elysia set object containing mutable response headers.
 * @param set.headers - The mutable headers record on the set object.
 * @param overrides - The settlement override values to attach.
 */
export function setSettlementOverrides(
  set: { headers: Record<string, unknown> },
  overrides: SettlementOverrides,
): void {
  set.headers[SETTLEMENT_OVERRIDES_HEADER] = JSON.stringify(overrides);
}

/**
 * Flattens an Elysia headers record into a plain string-to-string map for use in settlement.
 *
 * @param headers - The raw headers record from `ctx.set.headers`.
 * @returns A plain `Record<string, string>` suitable for the settlement transport context.
 */
export function flattenHeadersForSettlement(
  headers: Record<string, unknown> | undefined,
): Record<string, string> {
  const out: Record<string, string> = {};
  if (!headers) return out;
  for (const [k, v] of Object.entries(headers)) {
    if (v === undefined || v === null) continue;
    out[k] = Array.isArray(v) ? v.join(", ") : String(v);
  }
  return out;
}

/**
 * Removes the Settlement-Overrides header from the Elysia headers record.
 *
 * Called after the overrides have been read so the header is not forwarded to the client.
 *
 * @param headers - The mutable headers record from `ctx.set.headers`.
 */
export function stripSettlementOverridesHeader(headers: Record<string, unknown>): void {
  const match = SETTLEMENT_OVERRIDES_HEADER.toLowerCase();
  for (const k of Object.keys(headers)) {
    if (k.toLowerCase() === match) {
      delete headers[k];
    }
  }
}

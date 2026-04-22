/**
 * Header used to pass partial settlement overrides, matching the x402-foundation/x402 convention.
 *
 * Defined locally because @x402/core does not yet export this constant.
 * When upstream exports it, replace with:
 *   export { SETTLEMENT_OVERRIDES_HEADER, type SettlementOverrides } from "@x402/core/server";
 */
export const SETTLEMENT_OVERRIDES_HEADER = "Settlement-Overrides";

/**
 * A map of optional override values for the settlement call.
 */
export type SettlementOverrides = Record<string, string | number | undefined>;

/**
 * Attaches settlement overrides to the outgoing response headers before settlement runs.
 *
 * The Elysia middleware strips this header from the wire response after reading it
 * for `processSettlement`, so it is never exposed to the client.
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

import { paymentProxyFromConfig } from "@x402-avm/next";
import { HTTPFacilitatorClient } from "@x402-avm/core/server";
import { ExactEvmScheme } from "@x402-avm/evm/exact/server";
import { ExactSvmScheme } from "@x402-avm/svm/exact/server";
import { ExactAvmScheme } from "@x402-avm/avm/exact/server";
import { NextRequest, NextResponse } from "next/server";
import { createPaywall } from "@x402-avm/paywall";
import { evmPaywall } from "@x402-avm/paywall/evm";
import { svmPaywall } from "@x402-avm/paywall/svm";
import { avmPaywall } from "@x402-avm/paywall/avm";
import { createThemedPaywall } from "./themed-paywall";

const evmPayeeAddress = process.env.RESOURCE_EVM_ADDRESS as `0x${string}`;
const svmPayeeAddress = process.env.RESOURCE_SVM_ADDRESS as string;
const avmPayeeAddress = process.env.RESOURCE_AVM_ADDRESS || process.env.RESOURCE_WALLET_ADDRESS as string;
const facilitatorUrl = process.env.FACILITATOR_URL as string;

const EVM_NETWORK = "eip155:84532" as const; // Base Sepolia
const SVM_NETWORK = "solana:EtWTRABZaYq6iMfeYKouRu166VU2xqa1" as const; // Solana Devnet
const AVM_NETWORK = "algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=" as const; // Algorand Testnet

// List of blocked countries and regions
const BLOCKED_COUNTRIES = [
  "KP", // North Korea
  "IR", // Iran
  "CU", // Cuba
  "SY", // Syria
];

// List of blocked regions within specific countries
const BLOCKED_REGIONS = {
  UA: ["43", "14", "09"],
};

// Validate required environment variables
if (!facilitatorUrl) {
  console.error("❌ FACILITATOR_URL environment variable is required");
}

// Create HTTP facilitator client
const facilitatorClient = new HTTPFacilitatorClient({ url: facilitatorUrl });

// Build the paywall provider with landing-page dark theme
const basePaywall = createPaywall()
  .withNetwork(avmPaywall)
  .withNetwork(evmPaywall)
  .withNetwork(svmPaywall)
  .withConfig({
    appName: "x402 Demo",
    appLogo: "/logos/x402-examples.png",
  })
  .build();

const paywall = createThemedPaywall(basePaywall);

const accepts = [
  {
    payTo: avmPayeeAddress,
    scheme: "exact",
    price: "$0.01",
    network: AVM_NETWORK,
  },
  {
    payTo: evmPayeeAddress,
    scheme: "exact",
    price: "$0.01",
    network: EVM_NETWORK,
  },
  {
    payTo: svmPayeeAddress,
    scheme: "exact",
    price: "$0.01",
    network: SVM_NETWORK,
  },
];

const schemes = [
  { network: AVM_NETWORK, server: new ExactAvmScheme() },
  { network: EVM_NETWORK, server: new ExactEvmScheme() },
  { network: SVM_NETWORK, server: new ExactSvmScheme() },
];

const x402PaymentProxy = paymentProxyFromConfig(
  {
    "/protected": {
      accepts,
      description: "Access to protected content",
    },
    "/examples/weather": {
      accepts,
      description: "Access to protected weather API",
    },
  },
  facilitatorClient,
  schemes,
  undefined, // paywallConfig
  paywall, // paywall provider
);

const geolocationProxy = async (req: NextRequest) => {
  // Get the country and region from Vercel's headers
  const country = req.headers.get("x-vercel-ip-country") || "US";
  const region = req.headers.get("x-vercel-ip-country-region");

  const isCountryBlocked = BLOCKED_COUNTRIES.includes(country);
  const isRegionBlocked =
    region && BLOCKED_REGIONS[country as keyof typeof BLOCKED_REGIONS]?.includes(region);

  if (isCountryBlocked || isRegionBlocked) {
    return new NextResponse("Access denied: This service is not available in your region", {
      status: 451,
      headers: {
        "Content-Type": "text/plain",
      },
    });
  }

  return null;
};

export const proxy = async (req: NextRequest) => {
  const geolocationResponse = await geolocationProxy(req);
  if (geolocationResponse) {
    return geolocationResponse;
  }
  const delegate = x402PaymentProxy as unknown as (
    request: NextRequest,
  ) => ReturnType<typeof x402PaymentProxy>;
  return delegate(req);
};

// Configure which paths the proxy should run on
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (metadata files)
     */
    "/((?!_next/static|_next/image|favicon.ico).*)",
    "/", // Include the root path explicitly
  ],
};

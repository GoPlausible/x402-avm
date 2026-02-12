import Link from "next/link";
import Image from "next/image";
import {
  BoltIcon,
  CloudIcon,
  MusicalNoteIcon,
  CheckIcon,
  DocumentTextIcon,
  CodeBracketIcon,
  BookOpenIcon,
  Squares2X2Icon,
  ServerIcon,
  WrenchScrewdriverIcon,
  CurrencyDollarIcon,
  CpuChipIcon,
  SparklesIcon,
  LockOpenIcon,
  ArrowPathIcon,
  CommandLineIcon,
  UserGroupIcon,
} from "@heroicons/react/24/outline";

import { FeatureItem } from "./components/FeatureItem";
import AlgoIcon from "./assets/algorand-logomark-white-RGB.svg";
import { Section } from "./components/Section";
import { BackgroundVideo } from "./components/BackgroundVideo";
import NavBar from "./components/NavBar";

const whatIsItFeatures = [
  {
    title: "No fees",
    description: "x402 as a protocol has 0 fees for either the customer or the merchant.",
    icon: <CheckIcon className="w-5 h-5 text-indigo-400" />,
  },
  {
    title: "Instant settlement",
    description:
      "Accept payments at the speed of the blockchain. Money in your wallet in 2 seconds, not T+2.",
    icon: <CheckIcon className="w-5 h-5 text-indigo-400" />,
  },
  {
    title: "Blockchain Agnostic",
    description:
      "x402 is not tied to any specific blockchain or token, its a neutral standard open to integration by all.",
    icon: <CheckIcon className="w-5 h-5 text-indigo-400" />,
  },
  {
    title: "Frictionless",
    description:
      "As little as 1 line of middleware code or configuration in your existing web server stack and you can start accepting payments. Customers and agents aren't required to create an account or provide any personal information.",
    icon: <CheckIcon className="w-5 h-5 text-indigo-400" />,
  },

  {
    title: "Security & trust via an open standard",
    description:
      "Anyone can implement or extend x402. It's not tied to any centralized provider, and encourages broad community participation.",
    icon: <CheckIcon className="w-5 h-5 text-indigo-400" />,
  },
  {
    title: "Web native",
    description:
      "Activates the dormant 402 HTTP status code and works with any HTTP stack. It works simply via headers and status codes on your existing HTTP server.",
    icon: <CheckIcon className="w-5 h-5 text-indigo-400" />,
  },
];
const whyItMattersFeatures = [
  {
    title: "AI Agents",
    description: "Agents can use the x402 Protocol to pay for API requests in real-time.",
    icon: <BoltIcon className="w-5 h-5 text-indigo-400" />,
  },
  {
    title: "Cloud Storage Providers",
    description:
      "Using x402, customers can easily access storage services without account creation.",
    icon: <CloudIcon className="w-5 h-5 text-indigo-400" />,
  },
  {
    title: "Content Creators",
    description: "x402 unlocks instant transactions, enabling true micropayments for content.",
    icon: <MusicalNoteIcon className="w-5 h-5 text-indigo-400" />,
  },
];
const useCaseFeatures = [
  {
    title: "Micropayments",
    description:
      "Real-time, low-fee processing that makes high-volume, fractional-cent pricing economically viable.",
    icon: <CurrencyDollarIcon className="w-5 h-5 text-indigo-400" />,
  },
  {
    title: "Machine-to-machine payments (M2M)",
    description:
      "High-velocity payments between automated systems (e.g., IOT), enabling devices and services to independently exchange value.",
    icon: <CpuChipIcon className="w-5 h-5 text-indigo-400" />,
  },
  {
    title: "Agentic payments",
    description:
      "Autonomous software agents can initiate, negotiate, and settle transactions programmatically, enabling commerce between intelligent systems without human intervention.",
    icon: <SparklesIcon className="w-5 h-5 text-indigo-400" />,
  },
  {
    title: "Pay-as-you-go access",
    description:
      "Autonomous agents can unlock metered access to premium content, media, data, or services, billed per use instead of fixed subscriptions.",
    icon: <LockOpenIcon className="w-5 h-5 text-indigo-400" />,
  },
  {
    title: "Subscriptions and renewals",
    description:
      "Autonomous subscription payment renewals executed by agents on behalf of users.",
    icon: <ArrowPathIcon className="w-5 h-5 text-indigo-400" />,
  },
  {
    title: "Pay-per-API call",
    description:
      "Replace API keys, subscription tiers, and quota management with programmable payment per request.",
    icon: <CommandLineIcon className="w-5 h-5 text-indigo-400" />,
  },
];

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-gray-800 to-black text-white relative overflow-hidden">
      {/* Video Background */}
      <div className="fixed w-full z-0">
        <div className="fixed w-full bg-gradient-to-t from-black" />
        <BackgroundVideo src="/banner.mp4" />
      </div>

      <div className="relative z-10">
        {/* Top nav */}
        <NavBar />
        {/* Hero Section */}
        <section className="max-w-6xl mx-auto px-4 py-20 lg:py-28">
          <div className="text-center">
            <div className="mb-6 flex items-center justify-center gap-4">
              <Image
                src="/x402-logo.png"
                alt="x402 logo"
                width={320}
                height={160}
                className="inline-block"
              />
              <Image
                src="/algorand-logomark-white-RGB.png"
                alt="Algorand logo"
                width={100}
                height={100}
                className="inline-block"
              />
            </div>

            <p className="text-xl text-gray-400 mb-8 font-mono">
              x402 internet-native protocol for Algorand Blockchain
            </p>
            <div className="flex flex-wrap gap-4 mb-6 justify-center">
              <Link
                href="/protected"
                className="px-6 py-4 bg-blue-600 hover:bg-blue-700 rounded-lg font-mono transition-colors flex items-center gap-2 text-lg"
              >
                <CodeBracketIcon className="w-5 h-5 mr-1" />
                Protected Page (TestNet)
              </Link>
              <Link
                href="/examples/price"
                className="px-6 py-4 bg-indigo-600 hover:bg-indigo-700 rounded-lg font-mono transition-colors flex items-center gap-2 text-lg"
              >
                <CloudIcon className="w-5 h-5 mr-1" />
                Protected API (TestNet)
              </Link>
            </div>
            <div className="flex flex-wrap gap-4 mb-8 justify-center">
              <Link
                href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/README.md"
                target="_blank"
                rel="noopener noreferrer"
                className="px-4 py-3 border-2 border-transparent hover:border-blue-600 rounded-lg font-mono transition-colors flex items-center gap-2 text-sm"
              >
                <DocumentTextIcon className="w-5 h-5 mr-1" />
                Documentation
              </Link>
              <Link
                href="https://facilitator.goplausible.xyz"
                target="_blank"
                rel="noopener noreferrer"
                className="px-4 py-3 border-2 border-transparent hover:border-blue-600 rounded-lg font-mono transition-colors flex items-center gap-2 text-sm"
              >
                <ServerIcon className="w-5 h-5 mr-1" />
                X402 Facilitator
              </Link>
              <Link
                href="https://facilitator.goplausible.xyz/docs"
                target="_blank"
                rel="noopener noreferrer"
                className="px-4 py-3 border-2 border-transparent hover:border-blue-600 rounded-lg font-mono transition-colors flex items-center gap-2 text-sm"
              >
                <WrenchScrewdriverIcon className="w-5 h-5 mr-1" />
                Facilitator API Docs
              </Link>
              <Link
                href="https://github.com/algorand-devrel/algorand-agent-skills"
                target="_blank"
                rel="noopener noreferrer"
                className="px-4 py-3 border-2 border-transparent hover:border-blue-600 rounded-lg font-mono transition-colors flex items-center gap-2 text-sm"
              >
                <UserGroupIcon className="w-5 h-5 mr-1" />
                x402 Agent Skills
              </Link>
            </div>
          </div>
        </section>

        <Section>
          {/* What is it? */}
          <div className="relative">
            <div className="flex items-center justify-center gap-4 mb-6">
              <h3 className="text-3xl font-bold text-blue-400 text-center">
                The best way to accept digital payments on Web 3.0
              </h3>
            </div>
            <div className="bg-gray-800/30 rounded-2xl p-8 backdrop-blur-2xl border border-gray-700/50">
              <p className="text-gray-300 leading-relaxed text-xl mb-8">
                Built by Coinbase around the{" "}
                <Link
                  href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/402"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-400 hover:text-blue-500"
                >
                  HTTP 402
                </Link>{" "}
                status code,{" "}
                <span className="font-bold">
                  x402 protocol enables users to pay for resources via API
                </span>{" "}
                without registration, emails, OAuth, or complex signatures. Algorand support is
                provided by Algorand Foundation & GoPlausible.
              </p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-10 text-gray-400">
                {whatIsItFeatures.map((feature, index) => (
                  <FeatureItem key={index} {...feature} />
                ))}
              </div>
            </div>
          </div>
        </Section>

        <Section>
          {/* Why it matters */}
          <div className="relative">
            <div className="bg-gray-800/30 rounded-2xl p-8  backdrop-blur-xl border border-gray-700/50">
              <div className="flex items-center gap-4 mb-6">
                <h3 className="text-3xl font-bold text-indigo-400">
                  Powering Next-Gen Digital Commerce on Algorand and other blockchains
                </h3>
              </div>
              <p className="text-gray-300 leading-relaxed text-xl mb-8">
                <span className="font-bold">x402 unlocks new monetization models,</span> offering
                developers and content creators a frictionless way to earn revenue from small
                transactions without forcing subscriptions or showing ads.
              </p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {whyItMattersFeatures.map((feature, index) => (
                  <FeatureItem key={index} {...feature} iconBgColor="bg-indigo-500/10" />
                ))}
              </div>
              <div className="mt-10">
                <h4 className="text-2xl font-bold text-indigo-300 mb-6">Use Cases</h4>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {useCaseFeatures.map((feature, index) => (
                    <FeatureItem key={index} {...feature} iconBgColor="bg-indigo-500/10" />
                  ))}
                </div>
              </div>
            </div>
          </div>
        </Section>

        <Section>
          {/* How it works */}

          <div className="relative">
            <div className="bg-gray-800/30 rounded-2xl p-8 backdrop-blur-xl border border-gray-700/50">
              <div className="flex items-center gap-4 mb-6">
                <h3 className="text-3xl font-bold text-indigo-400">The x402 flow:</h3>
              </div>
              <p className="text-gray-300 leading-relaxed text-lg">
                x402 allows any web developer to accept crypto payments without the complexity of
                having to interact with the blockchain.
              </p>
              <br />

              <p className="text-gray-300 leading-relaxed text-lg mb-8">
                For a payment-requiring endpoint or resource (URL), if a request arrives without
                payment, the server responds with HTTP 402, prompting the client that payment is
                required, accompanied by payment requirements information.
              </p>
              <div className="mb-8">
                <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-gray-300 relative overflow-hidden">
                  <pre className="syntax-highlight">
                    <span className="text-purple-400">HTTP</span>
                    <span className="text-gray-300">/1.1 </span>
                    <span className="text-amber-300">402</span>
                    <span className="text-gray-300"> Payment Required</span>
                  </pre>
                </div>
              </div>
              <p className="text-gray-300 leading-relaxed text-lg mb-8">
                The client then constructs a payment transaction for the required amount and sends
                it back to the server in the X-PAYMENT request header.
              </p>
              <p className="text-gray-300 leading-relaxed text-lg mb-8">
                The server forwards the payment to a facilitator, which verifies and settles it on
                the Algorand blockchain, then signals back to the server with the results.
              </p>
              <p className="text-gray-300 leading-relaxed text-lg mb-8">
                Finally, the server validates the payment results and, if successful, serves the
                original request with a 200 OK response along with the requested resource.
              </p>
              <div className="mb-8">
                <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-gray-300 relative overflow-hidden">
                  <pre className="syntax-highlight">
                    <span className="text-purple-400">HTTP</span>
                    <span className="text-gray-300">/1.1 </span>
                    <span className="text-amber-300">200</span>
                    <span className="text-gray-300"> OK</span>
                  </pre>
                </div>
              </div>
            </div>
          </div>
        </Section>
        <Section>
          {/* Packages and links */}
          <div className="relative">
            <div className="bg-gray-800/30 rounded-2xl p-8 backdrop-blur-xl border border-gray-700/50">
              <div className="flex items-center gap-4 mb-6">
                <h3 className="text-3xl font-bold text-indigo-400">x402 Algorand Resources</h3>
              </div>
              <p className="text-gray-300 leading-relaxed text-xl mb-6">
                Algorand x402 Main Resources:
              </p>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/coinbase/x402/blob/main/specs/schemes/exact/scheme_exact_algo.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 Algorand Exact Spec
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/README.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 Algorand Guides, Examples and docs
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/x402-avm/tree/branch-v2-algorand-publish"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 Algorand Implementation Repository
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://facilitator.goplausible.xyz"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    X402 Facilitator
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://facilitator.goplausible.xyz/docs"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    Facilitator API Docs
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://facilitator.goplausible.xyz/supported"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    Facilitator Supported Networks
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/x402-avm/tree/branch-v2-algorand-publish/examples/typescript"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    TypeScript Example Packages
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/x402-avm/tree/branch-v2-algorand-publish/examples/python"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    Python Example Packages
                  </a>
                </div>
              </div>
              <br />
              <p className="text-gray-300 leading-relaxed text-xl mb-6">
                x402 Algorand TypeScript Packages (npm):
              </p>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://www.npmjs.com/package/@x402-avm/core"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    @x402-avm/core
                  </a>
                </div>

                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://www.npmjs.com/package/@x402-avm/avm"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    @x402-avm/avm
                  </a>
                </div>

                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://www.npmjs.com/package/@x402-avm/express"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    @x402-avm/express
                  </a>
                </div>

                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://www.npmjs.com/package/@x402-avm/hono"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    @x402-avm/hono
                  </a>
                </div>

                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://www.npmjs.com/package/@x402-avm/next"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    @x402-avm/next
                  </a>
                </div>

                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://www.npmjs.com/package/@x402-avm/fetch"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    @x402-avm/fetch
                  </a>
                </div>

                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://www.npmjs.com/package/@x402-avm/axios"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    @x402-avm/axios
                  </a>
                </div>

                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://www.npmjs.com/package/@x402-avm/paywall"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    @x402-avm/paywall
                  </a>
                </div>

                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://www.npmjs.com/package/@x402-avm/extensions"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    @x402-avm/extensions
                  </a>
                </div>
              </div>

              <br />
              <p className="text-gray-300 leading-relaxed text-xl mb-6">
                x402 Algorand Python Package (pip):
              </p>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://pypi.org/project/x402-avm/"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402-avm (PyPI)
                  </a>
                </div>
              </div>

              <br />
              <p className="text-gray-300 leading-relaxed text-xl mb-6">
                TypeScript Code Examples:
              </p>

              <p className="text-gray-400 text-sm mb-3 font-mono">Core and Mechanism Examples:</p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/typescript/x402-avm-core-examples.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 Core Package Examples
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/typescript/x402-avm-avm-examples.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 AVM (Algorand) Mechanism Examples
                  </a>
                </div>
              </div>

              <p className="text-gray-400 text-sm mb-3 font-mono">Extensions and Paywall Examples:</p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/typescript/x402-avm-extensions-examples.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 Extensions Examples
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/typescript/x402-avm-paywall-examples.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 Paywall UI Examples
                  </a>
                </div>
              </div>

              <p className="text-gray-400 text-sm mb-3 font-mono">Back-end Framework-Specific Middleware Examples:</p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/typescript/x402-avm-express-examples.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 Express Middleware Examples
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/typescript/x402-avm-hono-examples.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 Hono Middleware Examples
                  </a>
                </div>
              </div>

              <p className="text-gray-400 text-sm mb-3 font-mono">Fullstack Framework-Specific Examples:</p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/typescript/x402-avm-next-examples.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 Next.js Middleware Examples
                  </a>
                </div>
              </div>

              <p className="text-gray-400 text-sm mb-3 font-mono">HTTP Client Examples:</p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/typescript/x402-avm-fetch-examples.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 Fetch Client Examples
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/typescript/x402-avm-axios-examples.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 Axios Client Examples
                  </a>
                </div>
              </div>

              <br />
              <p className="text-gray-300 leading-relaxed text-xl mb-6">
                Python Code Examples:
              </p>

              <p className="text-gray-400 text-sm mb-3 font-mono">Core and Mechanism Examples:</p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/python/x402-avm-avm-examples-python.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 AVM (Algorand) Mechanism Examples (Python)
                  </a>
                </div>
              </div>

              <p className="text-gray-400 text-sm mb-3 font-mono">Extensions Examples:</p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/python/x402-avm-extensions-examples-python.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 Extensions Examples (Python)
                  </a>
                </div>
              </div>

              <p className="text-gray-400 text-sm mb-3 font-mono">Back-end Framework-Specific Middleware Examples:</p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/python/x402-avm-fastapi-examples-python.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 FastAPI Middleware Examples (Python)
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/python/x402-avm-flask-examples-python.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 Flask Middleware Examples (Python)
                  </a>
                </div>
              </div>

              <p className="text-gray-400 text-sm mb-3 font-mono">HTTP Client Examples:</p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/python/x402-avm-httpx-examples-python.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 HTTPX Client Examples (Python)
                  </a>
                </div>
                <div className="bg-black/50 rounded-lg p-4">
                  <a
                    href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/python/x402-avm-requests-examples-python.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-400 font-bold block mb-2"
                  >
                    x402 V2 Requests Client Examples (Python)
                  </a>
                </div>
              </div>
            </div>
          </div>
        </Section>
        <Section>
          {/* Algorand x402 packages */}
          <div className="relative">
            <div className="bg-gray-800/30 rounded-2xl p-8 backdrop-blur-xl border border-gray-700/50">
              <div className="flex items-center gap-4 mb-6">
                <h3 className="text-3xl font-bold text-indigo-400">
                  x402 Algorand packages make integration easy for developers
                </h3>
              </div>
              <p className="text-gray-300 leading-relaxed text-xl mb-8">
                x402 protocol comes with exact specifications for each blockchain virtual machine
                type, <span className="font-bold">EVM, AVM, SVM</span>, as well as a core
                package and per-framework integration packages published as scoped npm packages
                and a Python package on PyPI.
              </p>
              <div className="mb-8">
                <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-gray-300 relative overflow-hidden">
                  <pre className="syntax-highlight">
                    <span className="text-green-400">npm install @x402-avm/core @x402-avm/avm</span>
                    {"\n"}
                    <span className="text-gray-400">
                      // Core protocol + Algorand mechanism
                    </span>
                  </pre>
                </div>
              </div>
              <p className="text-gray-300 leading-relaxed text-xl mb-8">
                Built atop @x402-avm/core, integration packages exist for the most popular backend
                and frontend frameworks:{" "}
                <span className="font-bold">Express, Hono, Next.js, Fetch, and Axios</span>.
              </p>
              <p className="text-gray-300 leading-relaxed text-xl mb-8">
                {" "}
                <span className="font-bold">Backends and SSR:</span>
              </p>
              <div className="mb-8">
                <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-gray-300 relative overflow-hidden">
                  <pre className="syntax-highlight">
                    <span className="text-green-400">npm install @x402-avm/express</span>
                    {"\n"}
                    <span className="text-gray-400">
                      // x402 payment middleware for Express.js
                    </span>
                  </pre>
                </div>
              </div>
              <div className="mb-8">
                <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-gray-300 relative overflow-hidden">
                  <pre className="syntax-highlight">
                    <span className="text-green-400">npm install @x402-avm/hono</span>
                    {"\n"}
                    <span className="text-gray-400">
                      // x402 payment middleware for Hono (Cloudflare Workers, Deno)
                    </span>
                  </pre>
                </div>
              </div>
              <div className="mb-8">
                <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-gray-300 relative overflow-hidden">
                  <pre className="syntax-highlight">
                    <span className="text-green-400">npm install @x402-avm/next</span>
                    {"\n"}
                    <span className="text-gray-400">
                      // x402 payment middleware for Next.js (App Router)
                    </span>
                  </pre>
                </div>
              </div>
              <p className="text-gray-300 leading-relaxed text-xl mb-8">
                {" "}
                <span className="font-bold">Frontends and clients:</span>
              </p>
              <div className="mb-8">
                <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-gray-300 relative overflow-hidden">
                  <pre className="syntax-highlight">
                    <span className="text-green-400">npm install @x402-avm/fetch</span>
                    {"\n"}
                    <span className="text-gray-400">
                      // x402 client wrapper for the Fetch API
                    </span>
                  </pre>
                </div>
              </div>
              <div className="mb-8">
                <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-gray-300 relative overflow-hidden">
                  <pre className="syntax-highlight">
                    <span className="text-green-400">npm install @x402-avm/axios</span>
                    {"\n"}
                    <span className="text-gray-400">
                      // x402 client wrapper for Axios
                    </span>
                  </pre>
                </div>
              </div>
              <p className="text-gray-300 leading-relaxed text-xl mb-8">
                {" "}
                <span className="font-bold">Python:</span>
              </p>
              <div className="mb-8">
                <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-gray-300 relative overflow-hidden">
                  <pre className="syntax-highlight">
                    <span className="text-green-400">pip install &quot;x402-avm[avm,fastapi]&quot;</span>
                    {"\n"}
                    <span className="text-gray-400">
                      // Python package with extras: [avm], [fastapi], [flask], [httpx], [requests], [extensions], [all]
                    </span>
                  </pre>
                </div>
              </div>
            </div>
          </div>
        </Section>

        {/* x402 Button Section */}
        <div className="relative z-10 text-center py-12">
          <Image
            src="/x402-button-large.png"
            alt="x402 button"
            width={320}
            height={160}
            className="bg-white mx-auto"
            style={{ borderRadius: 10 }}
          />
        </div>
      </div>
      <footer className="relative z-10 py-8 text-center text-sm text-gray-400">
        By using this site, you agree to be bound by the GoPlausible's{" "}
        <a
          href="https://goplausible.com/terms"
          target="_blank"
          rel="noopener noreferrer"
          className="text-blue-500"
        >
          Terms of Service
        </a>{" "}
        and{" "}
        <a
          href="https://goplausible.com/privacy"
          target="_blank"
          rel="noopener noreferrer"
          className="text-blue-500"
        >
          Privacy Policy
        </a>
        .
      </footer>
    </div>
  );
}

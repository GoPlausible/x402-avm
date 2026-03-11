import Link from "next/link";
import { ChevronLeftIcon } from "@heroicons/react/24/solid";
import { CheckCircleIcon, MusicalNoteIcon } from "@heroicons/react/24/outline";
import NavBar from "../components/NavBar";
import { BackgroundVideo } from "../components/BackgroundVideo";

export default function ProtectedPage() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-gray-800 to-black text-white relative overflow-hidden">
      {/* Video Background */}
      <div className="fixed w-full z-0">
        <div className="fixed w-full bg-gradient-to-t from-black" />
        <BackgroundVideo src="/banner.mp4" />
      </div>

      <div className="relative z-10">
        <NavBar />

        <div className="max-w-4xl mx-auto px-4 py-8">
          <Link
            href="/"
            className="mb-8 inline-flex items-center text-white hover:text-gray-300 transition-colors font-mono relative z-20"
          >
            <ChevronLeftIcon className="w-5 h-5 mr-2" />
            Back to Main Page
          </Link>

          <div className="bg-gray-800/30 rounded-2xl p-8 backdrop-blur-2xl border border-gray-700/50">
            <div className="flex items-center gap-3 mb-6">
              <div className="p-2 rounded-lg bg-green-500/10">
                <CheckCircleIcon className="w-8 h-8 text-green-400" />
              </div>
              <div>
                <h1 className="text-3xl font-bold text-green-400">
                  Payment Successful
                </h1>
                <p className="text-gray-400 font-mono text-sm">
                  x402 Protected Content Unlocked
                </p>
              </div>
            </div>

            <div className="bg-black/30 rounded-xl p-6 border border-gray-700/30 mb-6">
              <div className="flex items-center gap-2 mb-4">
                <MusicalNoteIcon className="w-5 h-5 text-indigo-400" />
                <h2 className="text-lg font-semibold text-gray-200">
                  Enjoy this banger song
                </h2>
              </div>
              <div className="rounded-lg overflow-hidden">
                <iframe
                  width="100%"
                  height="300"
                  scrolling="no"
                  frameBorder="no"
                  allow="autoplay"
                  src="https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/2044190296&color=%23ff5500&auto_play=true&hide_related=false&show_comments=true&show_user=true&show_reposts=false&show_teaser=true&visual=true"
                />
              </div>
              <div className="mt-2 text-xs text-gray-500 font-mono">
                <a
                  href="https://soundcloud.com/dan-kim-675678711"
                  title="danXkim"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-gray-400 transition-colors"
                >
                  danXkim
                </a>{" "}
                &middot;{" "}
                <a
                  href="https://soundcloud.com/dan-kim-675678711/x402"
                  title="x402 (DJ Reppel Remix)"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-gray-400 transition-colors"
                >
                  x402 (DJ Reppel Remix)
                </a>
              </div>
            </div>

            <div className="bg-black/20 rounded-lg p-4 border border-gray-700/20">
              <p className="text-gray-400 text-sm font-mono leading-relaxed">
                This content was served after a successful x402 payment on Algorand TestNet.
                The payment was verified by the facilitator and settled on-chain before
                delivering this page.
              </p>
            </div>
          </div>
        </div>
      </div>

      <footer className="relative z-10 py-8 text-center text-sm text-gray-400">
        By using this site, you agree to be bound by GoPlausible&apos;s{" "}
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

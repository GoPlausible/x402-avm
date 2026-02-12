'use client';

import Link from "next/link";
import { ChevronDownIcon } from "@heroicons/react/24/outline";
import GithubIcon from "../assets/github.svg";
import LogoIcon from "../assets/logo.svg";
import AlgoIcon from "../assets/algorand-logomark-white-RGB.svg";
import { useState, useRef, useEffect } from "react";

const NavBar = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [isAlgoOpen, setIsAlgoOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const algoMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
      if (algoMenuRef.current && !algoMenuRef.current.contains(event.target as Node)) {
        setIsAlgoOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <section className="max-w-6xl mx-auto px-4 pt-4 relative z-20">
      <div className="flex gap-4 md:gap-8 justify-between sm:justify-end items-center">
        <Link
          href="https://goplausible.com"
          target="_blank"
          rel="noopener noreferrer"
          className="font-mono hover:text-blue-400 transition-colors flex items-center gap-1 text-sm"
        >
          <LogoIcon className="w-4 h-4 mr-1" />
          GoPlausible
        </Link>

        {/* Coinbase x402 dropdown */}
        <div className="relative" ref={menuRef}>
          <button
            onClick={() => setIsOpen(!isOpen)}
            className="font-mono hover:text-blue-400 transition-colors flex items-center gap-1 text-sm cursor-pointer"
          >
            Coinbase x402
            <ChevronDownIcon
              className={`w-3.5 h-3.5 transition-transform ${isOpen ? "rotate-180" : ""}`}
            />
          </button>
          {isOpen && (
            <div className="absolute right-0 mt-2 w-56 bg-gray-900/95 backdrop-blur-xl border border-gray-700/50 rounded-lg shadow-xl overflow-hidden">
              <Link
                href="https://x402.gitbook.io/x402"
                target="_blank"
                rel="noopener noreferrer"
                className="block px-4 py-3 text-sm font-mono text-gray-300 hover:text-blue-400 hover:bg-gray-800/50 transition-colors"
                onClick={() => setIsOpen(false)}
              >
                Coinbase x402 docs
              </Link>
              <Link
                href="/x402.pdf"
                target="_blank"
                rel="noopener noreferrer"
                className="block px-4 py-3 text-sm font-mono text-gray-300 hover:text-blue-400 hover:bg-gray-800/50 transition-colors border-t border-gray-700/30"
                onClick={() => setIsOpen(false)}
              >
                x402 one pager
              </Link>
              <Link
                href="/x402-whitepaper.pdf"
                target="_blank"
                rel="noopener noreferrer"
                className="block px-4 py-3 text-sm font-mono text-gray-300 hover:text-blue-400 hover:bg-gray-800/50 transition-colors border-t border-gray-700/30"
                onClick={() => setIsOpen(false)}
              >
                x402 white paper
              </Link>
            </div>
          )}
        </div>

        <Link
          href="https://github.com/GoPlausible"
          target="_blank"
          rel="noopener noreferrer"
          className="font-mono hover:text-blue-400 transition-colors flex items-center gap-2 text-sm"
        >
          <GithubIcon className="w-4 h-4 mr-1" fill="currentColor" />
          GitHub
        </Link>
        {/* Algorand x402 dropdown */}
        <div className="relative" ref={algoMenuRef}>
          <button
            onClick={() => setIsAlgoOpen(!isAlgoOpen)}
            className="font-mono hover:text-blue-400 transition-colors flex items-center gap-1 text-sm cursor-pointer"
          >
            <AlgoIcon className="w-10 h-10 mr-1" />
            Algorand
            <ChevronDownIcon
              className={`w-3.5 h-3.5 transition-transform ${isAlgoOpen ? "rotate-180" : ""}`}
            />
          </button>
          {isAlgoOpen && (
            <div className="absolute right-0 mt-2 w-64 bg-gray-900/95 backdrop-blur-xl border border-gray-700/50 rounded-lg shadow-xl overflow-hidden">
              <Link
                href="https://algorand.co"
                target="_blank"
                rel="noopener noreferrer"
                className="block px-4 py-3 text-sm font-mono text-gray-300 hover:text-blue-400 hover:bg-gray-800/50 transition-colors"
                onClick={() => setIsAlgoOpen(false)}
              >
                Algorand
              </Link>
              <Link
                href="https://algorand.co/agentic-commerce/x402"
                target="_blank"
                rel="noopener noreferrer"
                className="block px-4 py-3 text-sm font-mono text-gray-300 hover:text-blue-400 hover:bg-gray-800/50 transition-colors border-t border-gray-700/30"
                onClick={() => setIsAlgoOpen(false)}
              >
                Algorand x402
              </Link>
              <Link
                href="https://algorand.co/blog/x402-unlocking-the-agentic-commerce-era"
                target="_blank"
                rel="noopener noreferrer"
                className="block px-4 py-3 text-sm font-mono text-gray-300 hover:text-blue-400 hover:bg-gray-800/50 transition-colors border-t border-gray-700/30"
                onClick={() => setIsAlgoOpen(false)}
              >
                Read Algorand x402 blog post
              </Link>
              <Link
                href="https://algorand.co/agentic-commerce/x402/developers"
                target="_blank"
                rel="noopener noreferrer"
                className="block px-4 py-3 text-sm font-mono text-gray-300 hover:text-blue-400 hover:bg-gray-800/50 transition-colors border-t border-gray-700/30"
                onClick={() => setIsAlgoOpen(false)}
              >
                Algorand x402 post for devs
              </Link>
            </div>
          )}
        </div>
      </div>
    </section>
  );
};

export default NavBar;

import type { PaywallProvider, PaymentRequired, PaywallConfig } from "@x402-avm/paywall";

/**
 * Wraps an existing PaywallProvider to inject dark theme styling and
 * a navigation header that matches the x402-avm landing page look & feel.
 */
export function createThemedPaywall(basePaywall: PaywallProvider): PaywallProvider {
  return {
    generateHtml(paymentRequired: PaymentRequired, config?: PaywallConfig): string {
      const html = basePaywall.generateHtml(paymentRequired, config);

      const darkThemeCss = `
        <style>
          :root {
            --background-color: transparent;
            --container-background-color: rgba(31, 41, 55, 0.3);
            --text-color: #f3f4f6;
            --secondary-text-color: #9ca3af;
            --details-background-color: rgba(0, 0, 0, 0.3);
            --details-background-color-hover: rgba(0, 0, 0, 0.4);
            --button-primary-color: #2563eb;
            --button-primary-hover-color: #1d4ed8;
            --button-secondary-color: rgba(55, 65, 81, 0.5);
            --button-secondary-hover-color: rgba(55, 65, 81, 0.7);
            --button-positive-color: #059669;
            --button-positive-hover-color: #047857;
            --button-error-color: #ef4444;
            --button-error-hover-color: #dc2626;
          }

          body {
            background: linear-gradient(to bottom, #1f2937, #000000);
            min-height: 100vh;
            font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace;
          }

          .container {
            backdrop-filter: blur(24px);
            -webkit-backdrop-filter: blur(24px);
            border: 1px solid rgba(55, 65, 81, 0.5);
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
          }

          .title {
            color: #818cf8;
          }

          .container p,
          .container .subtitle,
          .header p {
            color: #f3f4f6;
          }

          .input {
            background-color: rgba(0, 0, 0, 0.3);
            border-color: rgba(55, 65, 81, 0.5);
            color: #f3f4f6;
          }

          .input:focus {
            border-color: #2563eb;
            box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.3);
          }

          .input option {
            background-color: #1f2937;
            color: #f3f4f6;
          }

          .payment-details {
            border: 1px solid rgba(55, 65, 81, 0.3);
          }

          .payment-label {
            color: #9ca3af;
          }

          .payment-value {
            color: #f3f4f6;
          }

          .status {
            color: #9ca3af;
          }

          .balance-button {
            color: #f3f4f6;
          }

          a {
            color: #60a5fa;
          }

          a:hover {
            color: #93bbfc;
          }

          .instructions {
            color: #9ca3af;
          }

          .spinner > div {
            border-color: rgba(55, 65, 81, 0.5);
            border-top-color: #60a5fa;
          }

          /* Nav header */
          .x402-nav {
            max-width: 48rem;
            margin: 0 auto;
            padding: 1rem 1.5rem 0;
            display: flex;
            justify-content: space-between;
            align-items: center;
          }

          .x402-nav a {
            color: #f3f4f6;
            text-decoration: none;
            font-size: 0.875rem;
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            transition: color 150ms;
          }

          .x402-nav a:hover {
            color: #9ca3af;
          }

          .x402-nav .nav-links {
            display: flex;
            gap: 1.5rem;
          }

          .x402-nav .nav-links a {
            font-size: 0.8rem;
          }
        </style>
      `;

      const navHeader = `
        <div class="x402-nav">
          <a href="/">
            <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M11.78 5.22a.75.75 0 0 1 0 1.06L8.06 10l3.72 3.72a.75.75 0 1 1-1.06 1.06l-4.25-4.25a.75.75 0 0 1 0-1.06l4.25-4.25a.75.75 0 0 1 1.06 0Z" clip-rule="evenodd" />
            </svg>
            Back to Main Page
          </a>
          <div class="nav-links">
            <a href="https://goplausible.com" target="_blank" rel="noopener noreferrer">
              <svg width="16" height="16" viewBox="0 0 1135 1135" style="display:inline-block;vertical-align:middle;margin-right:4px;">
                <defs>
                  <linearGradient id="gp1" gradientUnits="objectBoundingBox" x1="0.5" y1="0" x2="0.5" y2="1" gradientTransform="matrix(0.69,-0.66,0.96,0.84,-0.06,0.49)">
                    <stop offset="0" stop-color="#6ee831"/><stop offset="1" stop-color="#42a112"/>
                  </linearGradient>
                  <linearGradient id="gp2" gradientUnits="objectBoundingBox" x1="0.5" y1="0" x2="0.5" y2="1" gradientTransform="matrix(0.24,-0.96,1.30,0.32,-0.50,1.16)">
                    <stop offset="0" stop-color="#6ee831"/><stop offset="1" stop-color="#42a112"/>
                  </linearGradient>
                </defs>
                <path d="M634.17 171.81L206.1 919.98l680.85-1.27-160.05-273.1-110.51 196.89-271.83 0 340.42-590.66z" fill="url(#gp1)" transform="matrix(1.17,0,0,1.17,-185.44,-75.94)"/>
                <path d="M731.98 330.59L481.74 767.55l90.19-1.27 157.51-273.1 250.24 425.53 85.1 0z" fill="url(#gp2)" transform="matrix(1.17,0,0,1.17,-185.44,-75.94)"/>
              </svg>
              GoPlausible
            </a>
            <a href="https://github.com/GoPlausible/.github/blob/main/profile/algorand-x402-documentation/README.md" target="_blank" rel="noopener noreferrer">Documentation</a>
          </div>
        </div>
      `;

      // Inject dark theme CSS before </head>
      let themed = html.replace("</head>", `${darkThemeCss}\n</head>`);

      // Inject nav header after <body> (before the root div)
      themed = themed.replace('<div id="root">', `${navHeader}\n<div id="root">`);

      // Replace USDC faucet link with Circle faucet
      themed = themed.replace(
        /https:\/\/dispenser\.testnet\.aws\.algodev\.network\//g,
        "https://faucet.circle.com/",
      );

      // Add "Need Testnet Algo?" link after the USDC faucet line
      // The bundled HTML contains the compiled React output; we inject via script
      themed = themed.replace("</body>", `
        <script>
          (function() {
            var observer = new MutationObserver(function(mutations, obs) {
              var instructions = document.querySelectorAll('.instructions');
              if (instructions.length > 0) {
                var last = instructions[instructions.length - 1];
                if (!document.getElementById('algo-faucet-link')) {
                  var p = document.createElement('p');
                  p.className = 'instructions';
                  p.id = 'algo-faucet-link';
                  p.innerHTML = 'Need Testnet Algo? <a href="https://lora.algokit.io/testnet/fund" target="_blank" rel="noopener noreferrer">Get some <u>here</u>.</a>';
                  last.parentNode.insertBefore(p, last.nextSibling);
                }
              }
            });
            observer.observe(document.body, { childList: true, subtree: true });
          })();
        </script>
      </body>`);

      return themed;
    },
  };
}

import { paymentMiddleware, x402ResourceServer } from "@x402-avm/hono";
import { ExactEvmScheme } from "@x402-avm/evm/exact/server";
import { ExactSvmScheme } from "@x402-avm/svm/exact/server";
import { ExactAvmScheme } from "@x402-avm/avm/exact/server";
import { HTTPFacilitatorClient } from "@x402-avm/core/server";
import { declareDiscoveryExtension } from "@x402-avm/extensions/bazaar";
import { Hono, type Context } from "hono";

type Bindings = {
  EVM_ADDRESS: string;
  SVM_ADDRESS: string;
  AVM_ADDRESS: string;
  FACILITATOR_URL: string;
};

const app = new Hono<{ Bindings: Bindings }>();

// Lazy-initialized middleware per isolate (Cloudflare Workers reuse isolates)
let cachedMiddleware: ReturnType<typeof paymentMiddleware> | null = null;
let cachedBindingsKey = "";

function getPaymentMiddleware(env: Bindings) {
  // Re-create if bindings changed (different env in preview vs production)
  const key = `${env.FACILITATOR_URL}:${env.AVM_ADDRESS}:${env.EVM_ADDRESS}:${env.SVM_ADDRESS}`;
  if (cachedMiddleware && cachedBindingsKey === key) {
    console.log("[x402] Payment middleware cache hit (reusing existing)");
    return cachedMiddleware;
  }

  console.log("[x402] Initializing payment middleware...");
  console.log("[x402]   Facilitator URL:", env.FACILITATOR_URL);
  console.log("[x402]   AVM payTo:", env.AVM_ADDRESS);
  console.log("[x402]   EVM payTo:", env.EVM_ADDRESS);
  console.log("[x402]   SVM payTo:", env.SVM_ADDRESS);

  const facilitatorClient = new HTTPFacilitatorClient({ url: env.FACILITATOR_URL });
  console.log("[x402] HTTPFacilitatorClient created");

  const avm = {
    scheme: "exact",
    price: "$0.001",
    network: "algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=" as const,
    payTo: env.AVM_ADDRESS,
  };
  const evm = {
    scheme: "exact",
    price: "$0.001",
    network: "eip155:84532" as const,
    payTo: env.EVM_ADDRESS,
  };
  const svm = {
    scheme: "exact",
    price: "$0.001",
    network: "solana:EtWTRABZaYq6iMfeYKouRu166VU2xqa1" as const,
    payTo: env.SVM_ADDRESS,
  };

  // Each network-prefixed route has its own network first in accepts
  const avmAccepts = [avm, evm, svm];
  const evmAccepts = [evm, avm, svm];
  const svmAccepts = [svm, avm, evm];

  const weatherExtensions = {
    ...declareDiscoveryExtension({
      input: { city: "San Francisco" },
      inputSchema: {
        properties: { city: { type: "string", description: "City name" } },
        required: ["city"],
      },
      output: {
        example: { report: { weather: "sunny", temperature: 70 } },
      },
    }),
  };

  const server = new x402ResourceServer(facilitatorClient)
    .register("eip155:84532", new ExactEvmScheme())
    .register("solana:EtWTRABZaYq6iMfeYKouRu166VU2xqa1", new ExactSvmScheme())
    .register("algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=", new ExactAvmScheme());

  console.log("[x402] Registered schemes: ExactEvmScheme, ExactSvmScheme, ExactAvmScheme");

  const routes = {
    "GET /avm/weather": {
      accepts: avmAccepts,
      description: "Weather data",
      mimeType: "application/json",
      extensions: weatherExtensions,
    },
    "GET /evm/weather": {
      accepts: evmAccepts,
      description: "Weather data",
      mimeType: "application/json",
      extensions: weatherExtensions,
    },
    "GET /svm/weather": {
      accepts: svmAccepts,
      description: "Weather data",
      mimeType: "application/json",
      extensions: weatherExtensions,
    },
    "GET /avm/protected": {
      accepts: avmAccepts,
      description: "Protected content",
      mimeType: "text/html",
    },
    "GET /evm/protected": {
      accepts: evmAccepts,
      description: "Protected content",
      mimeType: "text/html",
    },
    "GET /svm/protected": {
      accepts: svmAccepts,
      description: "Protected content",
      mimeType: "text/html",
    },
  };

  console.log("[x402] Protected routes:", Object.keys(routes).join(", "));

  cachedMiddleware = paymentMiddleware(routes, server);
  cachedBindingsKey = key;

  console.log("[x402] Payment middleware initialized successfully");
  return cachedMiddleware;
}

// --- Global request/response logging middleware ---
app.use("/*", async (c, next) => {
  const start = Date.now();
  const method = c.req.method;
  const path = c.req.path;
  const url = c.req.url;
  const requestId = crypto.randomUUID().slice(0, 8);

  // Log incoming request
  console.log(`[${requestId}] --> ${method} ${path}`);
  console.log(`[${requestId}]     URL: ${url}`);
  console.log(`[${requestId}]     User-Agent: ${c.req.header("user-agent") ?? "(none)"}`);
  console.log(`[${requestId}]     Accept: ${c.req.header("accept") ?? "(none)"}`);

  // Log payment-related headers
  const paymentSig = c.req.header("payment-signature");
  const xPayment = c.req.header("x-payment");
  if (paymentSig) {
    console.log(`[${requestId}]     PAYMENT-SIGNATURE: ${paymentSig.slice(0, 80)}...`);
  } else if (xPayment) {
    console.log(`[${requestId}]     X-PAYMENT: ${xPayment.slice(0, 80)}...`);
  } else {
    console.log(`[${requestId}]     Payment header: (none)`);
  }

  try {
    await next();
  } catch (err) {
    const elapsed = Date.now() - start;
    console.error(`[${requestId}] !!! ${method} ${path} threw after ${elapsed}ms`);
    console.error(`[${requestId}]     Error: ${err instanceof Error ? err.message : String(err)}`);
    if (err instanceof Error && err.stack) {
      console.error(`[${requestId}]     Stack: ${err.stack}`);
    }
    throw err;
  }

  const elapsed = Date.now() - start;
  const status = c.res.status;

  // Log response
  if (status === 402) {
    console.log(`[${requestId}] <-- ${status} Payment Required (${elapsed}ms)`);
    const paymentRequired = c.res.headers.get("PAYMENT-REQUIRED");
    if (paymentRequired) {
      console.log(`[${requestId}]     PAYMENT-REQUIRED header set (${paymentRequired.length} chars)`);
    }
  } else if (status >= 400) {
    console.warn(`[${requestId}] <-- ${status} ${method} ${path} (${elapsed}ms)`);
    console.warn(`[${requestId}]     Content-Type: ${c.res.headers.get("content-type") ?? "(none)"}`);
  } else {
    console.log(`[${requestId}] <-- ${status} OK (${elapsed}ms)`);
  }

  // Log settlement header if present
  const paymentResponse = c.res.headers.get("PAYMENT-RESPONSE");
  if (paymentResponse) {
    console.log(`[${requestId}]     PAYMENT-RESPONSE header set — settlement succeeded`);
  }
});

// --- Payment middleware (applied after logging) ---
app.use("/*", async (c, next) => {
  console.log(`[x402] Running payment middleware for ${c.req.method} ${c.req.path}`);
  const middleware = getPaymentMiddleware(c.env);
  return middleware(c, next);
});

// --- Route handlers ---

app.get("/", (c) => {
  console.log("[route] GET / — serving landing page");
  return c.html(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>x402 AVM Testnet</title>
  <link rel="icon" href="/favicon.ico" type="image/x-icon">
  <style>
    *{margin:0;padding:0;box-sizing:border-box}
    body{
      min-height:100vh;display:flex;align-items:center;justify-content:center;
      font-family:system-ui,-apple-system,sans-serif;
      background:#050505;color:#e4e4e7;
    }
    .page{max-width:640px;width:92%;padding:48px 0}

    /* header */
    .logo-img{height:128px;margin-bottom:20px;opacity:.9}
    .logo{
      font-size:13px;font-weight:600;letter-spacing:1.5px;text-transform:uppercase;
      color:#818cf8;margin-bottom:8px;
    }
    h1{
      font-size:28px;font-weight:700;line-height:1.3;margin-bottom:16px;
      background:linear-gradient(135deg,#c7d2fe 0%,#818cf8 50%,#a78bfa 100%);
      -webkit-background-clip:text;-webkit-text-fill-color:transparent;
    }
    .intro{color:#a1a1aa;font-size:15px;line-height:1.7;margin-bottom:32px}
    .intro strong{color:#c7d2fe}

    /* network pills */
    .networks{display:flex;gap:8px;margin-bottom:36px;flex-wrap:wrap}
    .pill{
      padding:5px 14px;border-radius:999px;font-size:12px;font-weight:500;
      border:1px solid #262626;color:#a1a1aa;letter-spacing:.3px;
    }
    .pill.avm{border-color:#1e3a5f;color:#60a5fa}
    .pill.evm{border-color:#2d2044;color:#a78bfa}
    .pill.svm{border-color:#1a3a2a;color:#4ade80}

    /* section */
    .section{margin-bottom:32px}
    .section-title{
      font-size:11px;font-weight:600;letter-spacing:1.2px;text-transform:uppercase;
      color:#52525b;margin-bottom:14px;padding-bottom:10px;
      border-bottom:1px solid #1a1a1a;
    }

    /* endpoint grid */
    .grid{display:grid;grid-template-columns:1fr 1fr 1fr;gap:8px}
    .endpoint{
      position:relative;display:block;padding:14px 16px;border-radius:10px;
      border:1px solid #1a1a1a;text-decoration:none;
      transition:border-color .15s,background .15s;
    }
    .endpoint:hover{border-color:#333;background:#0f0f0f}
    .endpoint .net{
      font-size:11px;font-weight:600;letter-spacing:.8px;text-transform:uppercase;
      margin-bottom:4px;
    }
    .endpoint .net.avm{color:#60a5fa}
    .endpoint .net.evm{color:#a78bfa}
    .endpoint .net.svm{color:#4ade80}
    .endpoint .path{font-size:13px;color:#a1a1aa;font-family:monospace}
    .endpoint .price{font-size:11px;color:#52525b;margin-top:4px}
    .copy-btn{
      position:absolute;top:10px;right:10px;
      background:none;border:1px solid #262626;border-radius:6px;
      padding:4px 6px;cursor:pointer;color:#52525b;
      transition:color .15s,border-color .15s;line-height:1;
    }
    .copy-btn:hover{color:#a1a1aa;border-color:#3f3f46}
    .copy-btn.copied{color:#4ade80;border-color:#166534}
    .copy-btn svg{width:14px;height:14px;display:block}

    /* footer */
    .divider{height:1px;background:#1a1a1a;margin:32px 0 24px}
    .footer{display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:12px}
    .footer-text{font-size:13px;color:#3f3f46}
    .footer a{color:#818cf8;text-decoration:none}
    .footer a:hover{text-decoration:underline}
    .badge{
      padding:4px 12px;border-radius:999px;font-size:11px;font-weight:500;
      letter-spacing:.5px;border:1px solid #262626;color:#52525b;
    }

    @media(max-width:520px){
      .grid{grid-template-columns:1fr}
    }
  </style>
</head>
<body>
  <div class="page">
    <a href="https://www.goplausible.com"><img class="logo-img" src="/goPlausible-logo-type-h.png" alt="GoPlausible"></a>
    <div class="logo">x402 Protocol</div>
    <h1>Pay-per-Request API Testing Endpoints</h1>
    <p class="intro">
      <strong>x402</strong> is an open protocol for machine-to-machine payments over HTTP.
      It extends the <strong>402 Payment Required</strong> status code with a standardised
      header-based flow &mdash; clients receive payment requirements, sign a transaction,
      and attach it as a header on retry. The server verifies and settles on-chain before
      serving the response. No accounts, no sessions, no API keys.
    </p>

    <div class="networks">
      <span class="pill avm">Algorand Testnet(AVM)</span>
      <span class="pill evm">Ethereum / Base Sepolia (EVM)</span>
      <span class="pill svm">Solana Devnet (SVM)</span>
    </div>

    <p class="intro" style="margin-bottom:36px">
      All endpoints below run on <strong>testnet</strong> networks and cost
      <strong>$0.001</strong> per request. Payment verification and settlement is handled
      by the <a href="https://www.goplausible.xyz" style="color:#818cf8;text-decoration:none">GoPlausible</a>
      facilitator. Each route accepts all three networks but presents its primary network first.
    </p>

    <div class="section">
      <div class="section-title">Protected API &mdash; JSON response</div>
      <div class="grid">
        <a class="endpoint" href="/avm/weather">
          <button class="copy-btn" data-path="/avm/weather" onclick="copyUrl(event)"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg></button>
          <div class="net avm">Algorand</div>
          <div class="path">/avm/weather</div>
          <div class="price">$0.001</div>
        </a>
        <a class="endpoint" href="/evm/weather">
          <button class="copy-btn" data-path="/evm/weather" onclick="copyUrl(event)"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg></button>
          <div class="net evm">Ethereum</div>
          <div class="path">/evm/weather</div>
          <div class="price">$0.001</div>
        </a>
        <a class="endpoint" href="/svm/weather">
          <button class="copy-btn" data-path="/svm/weather" onclick="copyUrl(event)"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg></button>
          <div class="net svm">Solana</div>
          <div class="path">/svm/weather</div>
          <div class="price">$0.001</div>
        </a>
      </div>
    </div>

    <div class="section">
      <div class="section-title">Protected URL &mdash; HTML response</div>
      <div class="grid">
        <a class="endpoint" href="/avm/protected">
          <button class="copy-btn" data-path="/avm/protected" onclick="copyUrl(event)"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg></button>
          <div class="net avm">Algorand</div>
          <div class="path">/avm/protected</div>
          <div class="price">$0.001</div>
        </a>
        <a class="endpoint" href="/evm/protected">
          <button class="copy-btn" data-path="/evm/protected" onclick="copyUrl(event)"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg></button>
          <div class="net evm">Ethereum</div>
          <div class="path">/evm/protected</div>
          <div class="price">$0.001</div>
        </a>
        <a class="endpoint" href="/svm/protected">
          <button class="copy-btn" data-path="/svm/protected" onclick="copyUrl(event)"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg></button>
          <div class="net svm">Solana</div>
          <div class="path">/svm/protected</div>
          <div class="price">$0.001</div>
        </a>
      </div>
    </div>

    <div class="divider"></div>
    <div class="footer">
      <span class="footer-text">
        Facilitator by <a href="https://www.goplausible.xyz">GoPlausible</a>
        &middot; Read more about  <a href="https://facilitator.goplausible.xyz"> The x402 on Algorand</a>
      </span>
      <span class="badge">TESTNET</span>
    </div>
  </div>
  <script>
    function copyUrl(e){
      e.preventDefault();e.stopPropagation();
      var btn=e.currentTarget,path=btn.getAttribute('data-path');
      var url=location.origin+path;
      navigator.clipboard.writeText(url).then(function(){
        btn.classList.add('copied');
        btn.innerHTML='<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>';
        setTimeout(function(){
          btn.classList.remove('copied');
          btn.innerHTML='<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>';
        },1500);
      });
    }
  </script>
</body>
</html>`);
});

const weatherHandler = (c: Context<{ Bindings: Bindings }>) => {
  console.log(`[route] ${c.req.method} ${c.req.path} — payment verified, serving weather JSON`);
  return c.json({
    report: {
      weather: "sunny",
      temperature: 70,
    },
  });
};

app.get("/avm/weather", weatherHandler);
app.get("/evm/weather", weatherHandler);
app.get("/svm/weather", weatherHandler);

const protectedHandler = (c: Context<{ Bindings: Bindings }>) => {
  console.log(`[route] ${c.req.method} ${c.req.path} — payment verified, serving protected HTML`);
  return c.html(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Access Granted</title>
  <link rel="icon" href="/favicon.ico" type="image/x-icon">
  <style>
    *{margin:0;padding:0;box-sizing:border-box}
    body{
      min-height:100vh;
      display:flex;align-items:center;justify-content:center;
      font-family:system-ui,-apple-system,sans-serif;
      background:#0a0a0a;color:#fafafa;
    }
    .card{
      max-width:480px;width:90%;
      border:1px solid #262626;border-radius:16px;
      padding:48px 40px;text-align:center;
      background:linear-gradient(145deg,#111 0%,#0a0a0a 100%);
      box-shadow:0 0 80px rgba(99,102,241,.08);
    }
    .logo-img{height:28px;margin-bottom:20px;opacity:.9}
    .icon{font-size:48px;margin-bottom:20px}
    h1{font-size:24px;font-weight:600;margin-bottom:12px;
       background:linear-gradient(135deg,#818cf8,#a78bfa);
       -webkit-background-clip:text;-webkit-text-fill-color:transparent}
    p{color:#a1a1aa;line-height:1.6;font-size:15px}
    .badge{
      display:inline-block;margin-top:24px;
      padding:6px 16px;border-radius:999px;
      font-size:12px;font-weight:500;letter-spacing:.5px;
      border:1px solid #262626;color:#818cf8;
    }
    .divider{
      height:1px;background:#262626;margin:24px 0;
    }
    .footer{font-size:13px;color:#52525b}
    .footer a{color:#818cf8;text-decoration:none}
  </style>
</head>
<body>
  <div class="card">
    <a href="https://www.goplausible.xyz"><img class="logo-img" src="/goPlausible-logo-type-h.png" alt="GoPlausible"></a>
    <div class="icon">\u2713</div>
    <h1>Payment Verified</h1>
    <p>
      Congratulations! Your x402 payment was successfully processed
      and settled on-chain. You now have access to this protected resource.
    </p>
    <div class="divider"></div>
    <p>
      This content is served by a Cloudflare Worker using the
      <strong style="color:#e4e4e7">x402</strong> payment protocol
      with support for Algorand, Ethereum, and Solana.
    </p>
    <span class="badge">x402 PROTOCOL</span>
    <div class="divider"></div>
    <p class="footer">
      Powered by <a href="https://www.goplausible.xyz">GoPlausible</a>
    </p>
  </div>
</body>
</html>`);
};

app.get("/avm/protected", protectedHandler);
app.get("/evm/protected", protectedHandler);
app.get("/svm/protected", protectedHandler);

export default app;

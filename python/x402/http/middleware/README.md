# x402 HTTP Middleware with Algorand Integration

Server-side middleware for protecting routes with x402 payments, including first-class Algorand (AVM) support.

## FastAPI (Async)

```bash
uv add x402-avm[fastapi]
```

### Basic Usage

```python
from fastapi import FastAPI
from x402 import x402ResourceServer
from x402.http import HTTPFacilitatorClient, FacilitatorConfig, PaymentOption
from x402.http.middleware import PaymentMiddlewareASGI
from x402.http.types import RouteConfig
from x402.mechanisms.avm.exact import ExactAvmServerScheme
from x402.mechanisms.evm.exact import ExactEvmServerScheme

app = FastAPI()

# Configure server
facilitator = HTTPFacilitatorClient(FacilitatorConfig(url="https://x402.org/facilitator"))
server = x402ResourceServer(facilitator)
server.register("algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=", ExactAvmServerScheme())
server.register("eip155:84532", ExactEvmServerScheme())

# Define protected routes
routes = {
    "GET /api/weather/*": RouteConfig(
        accepts=[
            PaymentOption(
                scheme="exact",
                pay_to="ALGO_ADDRESS...",
                price="$0.01",
                network="algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=",
            ),
            PaymentOption(
                scheme="exact",
                pay_to="0x...",
                price="$0.01",
                network="eip155:84532",
            ),
        ],
        description="Weather API",
    ),
}

# Add middleware
app.add_middleware(PaymentMiddlewareASGI, routes=routes, server=server)
```

### Function-Based Middleware

```python
from x402.http.middleware import payment_middleware

@app.middleware("http")
async def x402_middleware(request, call_next):
    return await payment_middleware(routes, server)(request, call_next)
```

### Accessing Payment Info

```python
@app.get("/api/weather")
async def weather(request: Request):
    payload = request.state.payment_payload
    requirements = request.state.payment_requirements
    return {"weather": "sunny"}
```

## Flask (Sync)

```bash
uv add x402-avm[flask]
```

### Basic Usage

```python
from flask import Flask, g
from x402 import x402ResourceServerSync
from x402.http import HTTPFacilitatorClientSync, FacilitatorConfig, PaymentOption
from x402.http.middleware import PaymentMiddleware
from x402.http.types import RouteConfig
from x402.mechanisms.avm.exact import ExactAvmServerScheme
from x402.mechanisms.evm.exact import ExactEvmServerScheme

app = Flask(__name__)

# Configure server (sync variant)
facilitator = HTTPFacilitatorClientSync(FacilitatorConfig(url="https://x402.org/facilitator"))
server = x402ResourceServerSync(facilitator)
server.register("algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=", ExactAvmServerScheme())
server.register("eip155:84532", ExactEvmServerScheme())

# Define routes
routes = {
    "GET /api/weather/*": RouteConfig(
        accepts=[
            PaymentOption(
                scheme="exact",
                pay_to="ALGO_ADDRESS...",
                price="$0.01",
                network="algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=",
            ),
        ],
    ),
}

# Add middleware
PaymentMiddleware(app, routes, server)

@app.route("/api/weather")
def weather():
    payload = g.payment_payload
    return {"weather": "sunny"}
```

### Convenience Function

```python
from x402.http.middleware import payment_middleware

payment_middleware(app, routes, server, paywall_config={"appName": "My API"})
```

## Sync/Async Matching

| Framework | Server                   | Facilitator Client          |
| --------- | ------------------------ | --------------------------- |
| FastAPI   | `x402ResourceServer`     | `HTTPFacilitatorClient`     |
| Flask     | `x402ResourceServerSync` | `HTTPFacilitatorClientSync` |

Using async components with Flask raises `TypeError`.

## Route Patterns

| Pattern                | Matches                     |
| ---------------------- | --------------------------- |
| `GET /api/weather`     | Only GET to /api/weather    |
| `/api/users/*`         | Any method to /api/users/\* |
| `POST /api/users/[id]` | POST to /api/users/123      |
| `* /api/*`             | Any method to /api/\*       |

## Paywall Configuration

```python
paywall_config = {
    "appName": "My API",
    "appLogo": "/logo.png",
    "testnet": True,
}

app.add_middleware(PaymentMiddlewareASGI, routes=routes, server=server, paywall_config=paywall_config)
```

Browser requests to protected routes show an HTML paywall. API requests receive 402 with `PAYMENT-REQUIRED` header.

## Custom Paywall

```python
from x402.http import PaywallProvider

class MyPaywall(PaywallProvider):
    def generate_html(self, payment_required, config):
        return "<html>Custom paywall...</html>"

app.add_middleware(
    PaymentMiddlewareASGI, routes=routes, server=server, paywall_provider=MyPaywall()
)
```

## Exports

### FastAPI

| Export                             | Description                |
| ---------------------------------- | -------------------------- |
| `payment_middleware()`             | Create middleware function |
| `payment_middleware_from_config()` | Create from config dict    |
| `PaymentMiddlewareASGI`            | ASGI middleware class      |
| `FastAPIAdapter`                   | HTTPAdapter for FastAPI    |

### Flask

| Export                             | Description             |
| ---------------------------------- | ----------------------- |
| `PaymentMiddleware`                | WSGI middleware class   |
| `payment_middleware()`             | Convenience function    |
| `payment_middleware_from_config()` | Create from config dict |
| `FlaskAdapter`                     | HTTPAdapter for Flask   |

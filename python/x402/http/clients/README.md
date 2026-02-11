# x402 HTTP Clients with Algorand Integration

HTTP client wrappers with automatic 402 payment handling, including first-class Algorand (AVM) support.

## httpx (Async)

```bash
uv add x402-avm[httpx]
```

### Transport Wrapper

```python
from x402 import x402Client
from x402.http.clients import x402_httpx_transport
from x402.mechanisms.avm.exact import ExactAvmScheme
import httpx

client = x402Client()
client.register("algorand:*", ExactAvmScheme(signer=avm_signer))

async with httpx.AsyncClient(
    transport=x402_httpx_transport(client)
) as http:
    response = await http.get("https://api.example.com/paid")
```

### Convenience Wrapper

```python
from x402.http.clients import wrapHttpxWithPayment

async with wrapHttpxWithPayment(client) as http:
    response = await http.get("https://api.example.com/paid")
```

### From Config

```python
from x402 import x402ClientConfig, SchemeRegistration
from x402.http.clients import wrapHttpxWithPaymentFromConfig

config = x402ClientConfig(
    schemes=[
        SchemeRegistration(
            network="algorand:*",
            client=ExactAvmScheme(signer=avm_signer),
        ),
    ],
)

async with wrapHttpxWithPaymentFromConfig(config) as http:
    response = await http.get("https://api.example.com/paid")
```

### Client Class

```python
from x402.http.clients import x402HttpxClient

async with x402HttpxClient(client) as http:
    response = await http.get("https://api.example.com/paid")
```

## requests (Sync)

```bash
uv add x402-avm[requests]
```

### Session Wrapper

```python
from x402 import x402ClientSync
from x402.http.clients import wrapRequestsWithPayment
from x402.mechanisms.avm.exact import ExactAvmScheme
import requests

client = x402ClientSync()
client.register("algorand:*", ExactAvmScheme(signer=avm_signer))

session = wrapRequestsWithPayment(requests.Session(), client)
response = session.get("https://api.example.com/paid")
```

### HTTP Adapter

```python
from x402.http.clients import x402_http_adapter
import requests

session = requests.Session()
adapter = x402_http_adapter(client)
session.mount("https://", adapter)
session.mount("http://", adapter)

response = session.get("https://api.example.com/paid")
```

### Convenience Function

```python
from x402.http.clients import x402_requests

session = x402_requests(client)
response = session.get("https://api.example.com/paid")
```

### From Config

```python
from x402.http.clients import wrapRequestsWithPaymentFromConfig

session = wrapRequestsWithPaymentFromConfig(requests.Session(), config)
```

## Multi-Chain Registration

Register multiple blockchain mechanisms:

```python
from x402.mechanisms.avm.exact import ExactAvmScheme
from x402.mechanisms.evm.exact import ExactEvmScheme
from x402.mechanisms.svm.exact import ExactSvmScheme

client = x402Client()
client.register("algorand:*", ExactAvmScheme(signer=avm_signer))
client.register("eip155:*", ExactEvmScheme(signer=evm_signer))
client.register("solana:*", ExactSvmScheme(signer=svm_signer))
```

## Sync/Async Matching

| HTTP Client     | x402 Client             |
| --------------- | ----------------------- |
| httpx (async)   | `x402Client` (async)    |
| requests (sync) | `x402ClientSync` (sync) |

Using mismatched variants raises `TypeError`.

## Exports

### httpx

| Export                             | Description              |
| ---------------------------------- | ------------------------ |
| `x402_httpx_transport()`           | Create async transport   |
| `wrapHttpxWithPayment()`           | Wrap existing client     |
| `wrapHttpxWithPaymentFromConfig()` | Create from config       |
| `x402HttpxClient`                  | Convenience client class |

### requests

| Export                                | Description                     |
| ------------------------------------- | ------------------------------- |
| `x402_http_adapter()`                 | Create HTTP adapter             |
| `wrapRequestsWithPayment()`           | Wrap session                    |
| `wrapRequestsWithPaymentFromConfig()` | Create from config              |
| `x402_requests()`                     | Create new session with payment |

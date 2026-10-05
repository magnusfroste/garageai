# GarageAI

**Your garage. Europe's AI engine.**

GarageAI is a marketplace for local AI inference. Garage owners offer GPUs they already own, buyers call the models through one OpenAI-compatible API, and the owner gets paid for every token served. Local, sovereign, and the money goes to the people who own the hardware.

- 🌐 **Site:** [www.garageai.eu](https://www.garageai.eu)
- 🧑‍🔧 **Offer your GPU:** [app.garageai.eu](https://app.garageai.eu/auth?intent=operator)
- 🔌 **Use AI:** [app.garageai.eu](https://app.garageai.eu/auth?intent=buyer)

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## How it works

Two sides, one marketplace. One account can do both, and can own several garages.

```
Buyers ──HTTPS──▶ LiteLLM gateway (EU) ──private WireGuard mesh──▶ garages
                  llm.garageai.eu/v1                                Ollama · LM Studio · llama.cpp · vLLM · SGLang · Paddock
```

### Use AI

1. **Sign up and add credits** at [app.garageai.eu](https://app.garageai.eu/auth?intent=buyer) — prepaid, pay per token, no subscription.
2. **Create an API key.** The API at `https://llm.garageai.eu/v1` is OpenAI-compatible: change the base URL and the OpenAI SDKs, agents and coding tools just work.
3. **Pick a model and how to buy it** — **Pool** (cheaper, load-balanced across every garage offering that model, with automatic failover) or a **Specific garage** (a named garage, for example when location matters).

### Offer your GPU

1. **Start the "Offer your GPU" wizard** at [app.garageai.eu](https://app.garageai.eu/auth?intent=operator) — macOS or Linux (Windows coming soon) with Ollama, LM Studio, llama.cpp, vLLM, SGLang or Paddock (beta).
2. **Run one command.** [`scripts/garageai-connect.sh`](scripts/garageai-connect.sh) installs the [NetBird](https://netbird.io) client, joins your machine to the encrypted mesh, registers your models and installs a heartbeat that keeps the offered models in sync. No inbound ports are opened.
3. **Pass the acceptance test.** Every model is tested through the gateway with a real streamed request — time-to-first-token and tokens per second are measured before it can be sold.
4. **Go live and earn** the token price buyers pay for every request your garage serves. No platform fee during launch.

### Under the hood

- **The mesh** — self-hosted NetBird gives every garage a private WireGuard connection to the gateway. Garages are never exposed to the public internet.
- **The gateway** — [LiteLLM](https://github.com/BerriAI/litellm) on an EU VPS routes each request over the mesh, counts tokens per key and model, and handles retries and failover. Setup: [`infra/gateway/`](infra/gateway/README.md).

## Repository

| Path | What it is |
|---|---|
| [`src/`](src) | The Astro site at [www.garageai.eu](https://www.garageai.eu) |
| [`infra/gateway/`](infra/gateway/README.md) | The core: NetBird mesh + LiteLLM on one VPS — setup, DNS, firewall, backups |
| [`scripts/garageai-connect.sh`](scripts/garageai-connect.sh) | Connects a garage node to the mesh |
| [`docs/inference-engines.md`](docs/inference-engines.md) | Support matrix for local inference engines |

```bash
npm install
npm run dev     # the site on localhost
```

## Contributing & security

Contributions are welcome — see [CONTRIBUTING.md](docs/CONTRIBUTING.md). To report a vulnerability, see [SECURITY.md](docs/SECURITY.md).

## License

MIT — open source, made in Sweden, built for Europe.

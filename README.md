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
3. **Pick a model and how to buy it** — **Pool** (one model name across every garage offering it; each request goes to one healthy garage and is billed at that garage's price, with automatic failover) or a **Specific garage** (a named garage, for example when location matters; you pay its price). Each garage sets its own price; GarageAI has no price list of its own.

### Offer your GPU

1. **Start the "Offer your GPU" wizard** at [app.garageai.eu](https://app.garageai.eu/auth?intent=operator) — macOS or Linux with Ollama, LM Studio, llama.cpp, vLLM, SGLang or Paddock (beta). Create your garage and set your price per million tokens.
2. **Run one command.** The wizard gives you a command that installs [GarageAI Bridge](cli/README.md), one small open-source program (`garageai`), and runs `garageai connect`. Bridge installs the [NetBird](https://netbird.io) client, joins your machine to the encrypted mesh, checks that your runtime is reachable, registers your models and installs a heartbeat service that keeps the offered models in sync. No inbound ports are opened. It asks for your password once. Every step is reported to the wizard, so if something stops you see what to fix; `garageai doctor` shows the same on the machine, and `garageai uninstall` removes it again.

   **Windows is in beta:** [`scripts/garageai-connect.ps1`](scripts/garageai-connect.ps1) supports Ollama and LM Studio and runs in PowerShell as administrator. It is tested in CI but not yet on a real GPU machine, and the wizard does not offer Windows yet. To try it, email powerup@garageai.eu with the subject "Windows beta" and your GPU and runtime.
3. **Pass the acceptance test.** Every model is tested through the gateway with a real streamed request — time-to-first-token and tokens per second are measured before it can be sold.
4. **Go live and earn** the price you set for every token your garage serves, also for pool requests. No platform fee during launch.

### Under the hood

- **The mesh** — self-hosted NetBird gives every garage a private WireGuard connection to the gateway. Garages are never exposed to the public internet.
- **The gateway** — [LiteLLM](https://github.com/BerriAI/litellm) on an EU VPS routes each request over the mesh, counts tokens per key and model, and handles retries and failover. Setup: [`infra/gateway/`](infra/gateway/README.md).

## Repository

| Path | What it is |
|---|---|
| [`src/`](src) | The Astro site at [www.garageai.eu](https://www.garageai.eu) |
| [`infra/gateway/`](infra/gateway/README.md) | The core: NetBird mesh + LiteLLM on one VPS — setup, DNS, firewall, backups |
| [`cli/`](cli/README.md) | GarageAI Bridge (`garageai`): connects a garage node to the mesh, the default since October 2026 |
| [`scripts/garageai-connect.sh`](scripts/garageai-connect.sh) | The previous connect script, kept as a backup and reference (Windows: `garageai-connect.ps1`) |
| [`docs/bridge.md`](docs/bridge.md) | Bridge's vision and roadmap |
| [`docs/inference-engines.md`](docs/inference-engines.md) | Support matrix for local inference engines |

```bash
npm install
npm run dev     # the site on localhost
```

## Contributing & security

Contributions are welcome — see [CONTRIBUTING.md](docs/CONTRIBUTING.md). To report a vulnerability, see [SECURITY.md](docs/SECURITY.md).

Offering your GPU? [Security for garage owners](docs/garage-security.md) explains in plain words what Bridge installs and who can reach your machine.

## License

MIT — open source, made in Sweden, built for Europe.

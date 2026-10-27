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
Buyers ──HTTPS──▶ LiteLLM gateway (EU) ──WireGuard tunnel (Bridge)──▶ garages
                  api.garageai.eu/v1                                Ollama · LM Studio · llama.cpp · vLLM · SGLang · Xyntetik Runner · Paddock
```

### Use AI

1. **Sign up and add credits** at [app.garageai.eu](https://app.garageai.eu/auth?intent=buyer) — prepaid, pay per token, no subscription.
2. **Create an API key.** The API at `https://api.garageai.eu/v1` is OpenAI-compatible: change the base URL and the OpenAI SDKs, agents and coding tools just work.
3. **Pick a model and how to buy it** — **Pool** (one model name across every garage offering it; each request goes to one healthy garage and is billed at that garage's price, with automatic failover) or a **Specific garage** (a named garage, for example when location matters; you pay its price). Each garage sets its own price; GarageAI has no price list of its own.

### Offer your GPU

1. **Start the "Offer your GPU" wizard** at [app.garageai.eu](https://app.garageai.eu/auth?intent=operator) — macOS, Linux or Windows with Ollama, LM Studio, llama.cpp, vLLM, SGLang, Xyntetik Runner or Paddock (beta). No engine yet? `garageai runtime install` sets up Xyntetik Runner, a one-file engine made in Sweden, with a model that fits the machine. Create your garage and set your price per million tokens.
2. **Run one command.** The wizard gives you a command that installs [GarageAI Bridge](cli/README.md), one small open-source program (`garageai`), and runs `garageai connect`. Bridge opens its own encrypted WireGuard tunnel to the gateway (no VPN app, and your runtime can stay on localhost; [NetBird](https://netbird.io) is the fallback for networks that block UDP), checks your runtime, registers your models and installs a heartbeat service that keeps the offered models in sync. No inbound ports are opened. It asks for your password once. Every step is reported to the wizard, so if something stops you see what to fix; `garageai doctor` shows the same on the machine, and `garageai uninstall` removes it again.

   **Windows** (Bridge 1.0): the same, from PowerShell opened as Administrator. The tunnel runs as a Windows service and the heartbeat as a scheduled task, so the garage keeps serving after a reboot without anyone logged in.
3. **Pass the acceptance test.** Every model is tested through the gateway with a real streamed request — time-to-first-token and tokens per second are measured before it can be sold.
4. **Go live and earn** the price you set for every token your garage serves, also for pool requests. No platform fee during launch.

### Under the hood

- **The tunnel** — Bridge carries a WireGuard tunnel in user space from each garage out to a hub on the gateway ([design](docs/bridge-tunnel.md)); a self-hosted NetBird mesh is the fallback. Garages are never exposed to the public internet.
- **The gateway** — [LiteLLM](https://github.com/BerriAI/litellm) on an EU VPS routes each request over the mesh, counts tokens per key and model, and handles retries and failover. Setup: [`infra/gateway/`](infra/gateway/README.md).
- **Bridge** — the `garageai` binary on each garage ([`cli/`](cli/README.md)): doctor, connect (browser link, no tokens to paste), the tunnel, a heartbeat that reports models and context windows, self-update, and models fetched on the portal's request ([vision and phases](docs/bridge.md)).
- **Where logic lives** — the portal decides, the gateway executes within seconds, Bridge sees the machine, the Operations Center measures ([architecture](docs/architecture.md)). Security for garage owners: [what runs on your machine](docs/garage-security.md).
- **For engine builders** — any runtime that speaks the OpenAI API is a garage: `GET /v1/models`, `POST /v1/chat/completions` with SSE streaming, usage in the last chunk (`prompt_tokens_details.cached_tokens` lets buyers pay the cache price), and ideally `/metrics` so Bridge can tell engines apart. Listening on `127.0.0.1` is enough: the tunnel reaches it there. The support matrix and the engines under evaluation: [docs/inference-engines.md](docs/inference-engines.md).

## Repository

| Path | What it is |
|---|---|
| [`cli/`](cli/README.md) | GarageAI Bridge (`garageai`): connects a garage with its own WireGuard tunnel, heartbeat, self-update, the menu-bar icon, `garageai runtime` |
| [`infra/gateway/`](infra/gateway/README.md) | The core on one EU VPS: LiteLLM gateway and its routing rules, the Bridge tunnel hub, Operations Center, health checks, backups (NetBird only as the fallback for networks that block UDP) |
| [`www/`](www) | The Astro site at [www.garageai.eu](https://www.garageai.eu): page copy in [`www/src/content/`](www/src/content), legal texts in [`www/src/legal/`](www/src/legal) |
| [`scripts/garageai-connect.sh`](scripts/garageai-connect.sh) | The previous connect script, kept as a backup and reference (Windows: `garageai-connect.ps1`) |
| [`docs/www-and-portal.md`](docs/www-and-portal.md) | Who does what: the site (explain, persuade, search) and the portal (do, live data), and who owns each touchpoint |
| [`docs/architecture.md`](docs/architecture.md) | Where logic lives: portal, gateway, Bridge, Operations Center; the control plane later |
| [`docs/bridge.md`](docs/bridge.md) | Bridge's vision and roadmap |
| [`docs/garage-troubleshooting.md`](docs/garage-troubleshooting.md) | What each problem Bridge reports means and how to fix it |
| [`docs/inference-engines.md`](docs/inference-engines.md) | Support matrix for local inference engines |
| [`docs/my-machine.md`](docs/my-machine.md) | My machine: your own garage, private and free, from anywhere |

The app at [app.garageai.eu](https://app.garageai.eu) (accounts, keys, credits, chat, My garages) is its own repository, [garageai-portal](https://github.com/magnusfroste/garageai-portal).

```bash
cd www
npm install
npm run dev     # the site on localhost
```

## Contributing & security

Contributions are welcome — see [CONTRIBUTING.md](docs/CONTRIBUTING.md). To report a vulnerability, see [SECURITY.md](docs/SECURITY.md).

Offering your GPU? [Security for garage owners](docs/garage-security.md) explains in plain words what Bridge installs and who can reach your machine. Or let your own AI check: [AUDIT.md](AUDIT.md) lists every security claim with the code that proves it.

## License

MIT — open source, made in Sweden, built for Europe.

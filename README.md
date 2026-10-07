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
                                          ┌─ private WireGuard mesh ─▶ garages (home and office GPUs:
Buyers ──HTTPS──▶ LiteLLM gateway (EU) ───┤                            Ollama · LM Studio · llama.cpp · vLLM · SGLang · Paddock)
                  llm.garageai.eu/v1      └─ HTTPS ──────────────────▶ verified providers (datacenter endpoints)
```

Supply comes from two kinds of operators, listed side by side in the same catalogue and sold
through the same API:

- **Garages** — people and companies with a GPU they already own. The connect script joins the
  machine to the mesh; no inbound ports, no public address. Location is measured from the mesh
  connection and shown as a flag (never stored as an IP).
- **Verified providers** — companies with a public OpenAI-compatible endpoint (their own
  datacenter or cluster). Added by GarageAI admins after an agreement; marked "Provider" in the
  catalogue.

### Use AI

1. **Sign up and add credits** at [app.garageai.eu](https://app.garageai.eu/auth?intent=buyer) — prepaid, pay per token, no subscription.
2. **Create an API key.** The API at `https://llm.garageai.eu/v1` is OpenAI-compatible: change the base URL and the OpenAI SDKs, agents and coding tools just work.
3. **Pick a model and how to buy it.** Model names follow the OpenRouter convention, `creator/model` (for example `deepseek/deepseek-v4-flash`, `qwen/qwen3.8-27b`), so an OpenRouter config moves over by changing the base URL. **Pool** = that name, load-balanced across every garage and provider offering the model, with automatic failover. **Specific garage** = `garage/<name>/<model>`, when it matters who serves you and where. An **EU only** filter shows only supply with a visible EU/EEA location.

### Offer your GPU

1. **Start the "Offer your GPU" wizard** at [app.garageai.eu](https://app.garageai.eu/auth?intent=operator) — macOS or Linux (Windows coming soon) with Ollama, LM Studio, llama.cpp, vLLM, SGLang or Paddock (beta). You accept the operator terms (never log, store or read buyers' prompts) before you get your command.
2. **Run one command.** [`scripts/garageai-connect.sh`](scripts/garageai-connect.sh) installs the [NetBird](https://netbird.io) client, joins your machine to the encrypted mesh, registers your models and installs a heartbeat that keeps your model inventory in sync. No inbound ports are opened.
3. **Choose what to offer.** The heartbeat reports every installed model; nothing is sold until you switch it on. Each offered model passes an **acceptance test** through the gateway (a real streamed request measuring time-to-first-token and tokens per second) before it can be sold, and is probed hourly after that.
4. **Go live and earn** the token price buyers pay for every request your garage serves, pause whenever you like, and watch your **Reliability** grade and **Availability** on your garage page. No platform fee during launch.

### Under the hood

- **The mesh** — self-hosted NetBird gives every garage a private WireGuard connection to the gateway. Only the gateway can reach a garage's runtime port; garages cannot reach each other or the internet through the mesh.
- **The gateway** — [LiteLLM](https://github.com/BerriAI/litellm) on an EU VPS routes each request over the mesh (or over HTTPS to a provider), counts tokens per key and model, and handles retries and failover. Responses carry nothing that reveals which garage or provider served them. Setup: [`infra/gateway/`](infra/gateway/README.md).
- **The portal** — accounts, credits, keys, chat, the operator wizard, the catalogue, acceptance tests, reliability grades, revenue statements. The portal is the source of truth; it writes to LiteLLM and NetBird. Lives in [magnusfroste/garageai-portal](https://github.com/magnusfroste/garageai-portal).
- **Health** — a service on the gateway checks every garage every minute (tunnel up? runtime answering?) and the portal delists and relists within a minute, without any action from the operator.

## Repository

| Path | What it is |
|---|---|
| [`src/`](src) | The Astro site at [www.garageai.eu](https://www.garageai.eu) |
| [`infra/gateway/`](infra/gateway/README.md) | The core: NetBird mesh + LiteLLM on one VPS — setup, lockdown, health service, backups, security checklist |
| [`scripts/garageai-connect.sh`](scripts/garageai-connect.sh) | Connects a garage to the mesh (macOS, Linux); [`garageai-connect.ps1`](scripts/garageai-connect.ps1) is the Windows port in progress |
| [`scripts/tests/`](scripts/tests) | Tests for the connect scripts, run in CI on Ubuntu, macOS and Windows |
| [`docs/garage-security.md`](docs/garage-security.md) | What the script installs, what GarageAI can reach, the operator terms |
| [`docs/inference-engines.md`](docs/inference-engines.md) | Support matrix for local inference engines |

```bash
npm install
npm run dev     # the site on localhost
```

## Contributing & security

Contributions are welcome — see [CONTRIBUTING.md](docs/CONTRIBUTING.md). To report a vulnerability, see [SECURITY.md](docs/SECURITY.md).

Offering your GPU? [Security for garage owners](docs/garage-security.md) explains in plain words what the connect script installs and who can reach your machine.

## License

MIT — open source, made in Sweden, built for Europe.

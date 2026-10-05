# GarageAI

**Your garage. Europe's AI engine.**

GarageAI is a marketplace for local AI inference. Garage owners offer GPUs they already own, buyers call the models through one OpenAI-compatible API, and the owner gets paid for every token served. Local, sovereign, and the money goes to the people who own the hardware.

- 🌐 **Site:** [www.garageai.eu](https://www.garageai.eu)
- 🧑‍🔧 **Offer your GPU:** [app.garageai.eu](https://app.garageai.eu/auth?intent=operator)
- 🔌 **Use AI:** [app.garageai.eu](https://app.garageai.eu/auth?intent=buyer)

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## How it works

```
Buyers ──HTTPS──▶ LiteLLM gateway (EU VPS) ──private WireGuard mesh──▶ garages
                  one OpenAI-compatible API                            Ollama · LM Studio · llama.cpp · vLLM · …
```

- **The mesh** — every garage joins a private WireGuard mesh run on self-hosted [NetBird](https://netbird.io). Garages open no inbound ports and are never exposed to the public internet.
- **The gateway** — [LiteLLM](https://github.com/BerriAI/litellm) on an EU VPS is the API buyers call. It routes each request over the mesh to a garage, counts tokens per key and model, and handles retries and failover.
- **The garage** — operators keep their own macOS or Linux machine and whichever runtime they like. A heartbeat reports the models currently loaded, so what GarageAI sells follows what the garage runs.

## Connect a garage

Sign up as an operator at [app.garageai.eu](https://app.garageai.eu/auth?intent=operator) to get a setup key, then on the machine that serves your models:

```bash
./scripts/garageai-connect.sh --setup-key <KEY> --management-url <URL> --runtime ollama
```

The script installs the NetBird client, joins the mesh, checks that your runtime answers on its mesh IP, registers the models you offer, and installs the heartbeat. Run it with `--help` for every option.

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

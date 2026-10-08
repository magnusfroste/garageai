# GarageAI core: NetBird mesh + LiteLLM on one VPS

The GarageAI core runs on a single EU VPS:

- **NetBird** (self-hosted, fully open source, NetBird GmbH, Berlin) gives every garage a
  private WireGuard connection to the gateway. Garages open no inbound ports and are never
  exposed to the public internet, and no US-run tunnel service sits in the path.
- **LiteLLM** is the OpenAI-compatible API buyers call. It routes each request over the mesh
  to a garage, counts tokens per key and model, and handles retries and failover.

```
                           ┌──────────────── EU VPS (Docker) ─────────────────┐
 Buyers ──HTTPS──▶ :443 ──▶│ Traefik ─┬─▶ llm.garageai.eu     → LiteLLM ─▶ Postgres
                           │          └─▶ netbird.garageai.eu → NetBird server + dashboard
 Garages ─UDP 3478 (STUN)─▶│                                                     │
                           │ NetBird client on the host (wt0) ◀── LiteLLM routes │
                           └──────────────────────────┬──────────────────────────┘
                                   private WireGuard mesh
                      ┌───────────────┬───────────────┴───────────┐
                  garage-lund     garage-malmo                garage-…
                  (Ollama)        (llama.cpp)                 (Paddock)
```

## Docker or directly on the VPS?

| Component | Where | Why |
|---|---|---|
| NetBird server + dashboard | Docker | The official installer is Docker-based and sets up Traefik and Let's Encrypt for you |
| LiteLLM + Postgres | Docker | LiteLLM's documented deployment. Pinned image tags make upgrades and rollbacks one command |
| Traefik | Docker | Comes with NetBird's installer. LiteLLM joins it via labels, so there is one proxy and one set of certificates |
| **NetBird client** (the gateway's own peer) | **Directly on the host** | It creates the `wt0` WireGuard interface. The host routes the mesh range through it, and Docker containers, LiteLLM included, reach garages via the host's routes |

## 0. VPS, DNS and firewall

- **VPS:** Hetzner CX22 (2 vCPU, 4 GB, ~€4/month) is enough for a PoC; CX32 gives headroom.
  Ubuntu 24.04, location Falkenstein, Nuremberg or Helsinki.
- **DNS:** two A records pointing at the VPS: `netbird.garageai.eu` and `llm.garageai.eu`.
- **Hetzner Cloud Firewall, inbound:**

  | Port | Protocol | Purpose |
  |---|---|---|
  | 22 | TCP | SSH, restricted to your own IP |
  | 80, 443 | TCP | Traefik (Let's Encrypt, NetBird, LiteLLM) |
  | 3478 | UDP | STUN, so peers can discover their public address |
  | 51820 | UDP | Optional: lets garages connect to the gateway peer directly instead of via relay |

```bash
curl -fsSL https://get.docker.com | sh
```

## 1. NetBird server

Run NetBird's official installer unattended. Option `0` installs the built-in Traefik:

```bash
mkdir -p /opt/garageai/netbird && cd /opt/garageai/netbird
curl -fsSL -o getting-started.sh https://github.com/netbirdio/netbird/releases/latest/download/getting-started.sh
sudo NETBIRD_DOMAIN=netbird.garageai.eu \
     NETBIRD_LETSENCRYPT_EMAIL=you@example.com \
     NETBIRD_REVERSE_PROXY_TYPE=0 \
     NETBIRD_ENABLE_PROXY=false \
     NETBIRD_ENABLE_CROWDSEC=false \
     NETBIRD_NON_INTERACTIVE=true \
     bash getting-started.sh
```

The installer writes `docker-compose.yml`, `config.yaml` and `dashboard.env` to this
directory and starts the stack. Note the Docker network name; LiteLLM joins it in step 4:

```bash
docker network ls | grep netbird      # expected: netbird_netbird (directory name + "_netbird")
```

Open `https://netbird.garageai.eu` and create the admin account.

> **Back up `/opt/garageai/netbird/config.yaml`.** It holds the datastore encryption key.
> Without it, the NetBird database cannot be read.

## 2. Groups, access policy and setup keys (NetBird dashboard)

1. **Groups:** create `gateway` and `garages`.
2. **Access control — important:** delete the *Default* policy. It lets every peer reach every
   other peer, so garages could reach each other. Add one policy instead:

   | Source    | Destination | Protocol | Ports |
   |-----------|-------------|----------|-------|
   | `gateway` | `garages`   | TCP      | 11434, 1234, 8080, 8000 (+ any custom runtime port) |

3. **Setup keys:**
   - one *one-off* key with auto-group `gateway`, for this VPS;
   - one key per operator with auto-group `garages`, one-off or with a short expiry. Revoking
     it in the dashboard cuts that operator off.

## 3. Join the VPS to the mesh as the gateway peer

Installed directly on the host, not in Docker:

```bash
curl -fsSL https://pkgs.netbird.io/install.sh | sh
sudo env NB_SETUP_KEY=<gateway-setup-key> netbird up --management-url https://netbird.garageai.eu
netbird status        # should show "Connected" and a 100.x mesh IP; peer in group "gateway"
```

## 4. LiteLLM

```bash
git clone https://github.com/magnusfroste/garageai /opt/garageai/repo
cp -r /opt/garageai/repo/infra/gateway/litellm /opt/garageai/litellm
cd /opt/garageai/litellm
cp .env.example .env
nano .env             # LITELLM_DOMAIN, NETBIRD_NETWORK, keys and password (see comments)
docker compose up -d
curl https://llm.garageai.eu/health/liveliness
```

`config.yaml` has no models: garages are added at runtime (step 6). Its router settings retry
on another garage and take a failing garage out of rotation for 60 seconds, because residential
nodes drop out more often than datacenter ones. Streaming passes through Traefik without
buffering; the NetBird installer already disables Traefik's timeouts for long-lived streams.

## 5. Onboard a garage

The operator starts their runtime (Ollama, LM Studio, llama.cpp, vLLM, Paddock, Unsloth),
then runs:

```bash
curl -fsSLO https://raw.githubusercontent.com/magnusfroste/garageai/main/scripts/garageai-connect.sh
bash garageai-connect.sh --setup-key <their-key> --management-url https://netbird.garageai.eu --runtime ollama
```

The script joins the mesh, checks that the runtime answers on the mesh IP, explains how to
rebind it if it only listens on localhost, and prints the node's details.

## 6. Register the garage in LiteLLM

From anywhere that can reach the LiteLLM API:

```bash
export LITELLM_URL=https://llm.garageai.eu LITELLM_MASTER_KEY=sk-...
infra/gateway/register-node.sh add garage-lund 100.92.1.7 11434 qwen3:32b gemma4:27b
infra/gateway/register-node.sh list
infra/gateway/register-node.sh remove garage-lund qwen3:32b     # when they stop sharing
```

If the runtime requires an API key (for example vLLM started with `--api-key`), pass it in
`NODE_API_KEY`. LiteLLM stores it encrypted and buyers never see it:

```bash
NODE_API_KEY=sk-... infra/gateway/register-node.sh add garage-lund 100.92.1.7 8000 glm-5.3-flash
```

First end-to-end test:

```bash
curl https://llm.garageai.eu/v1/chat/completions \
  -H "Authorization: Bearer $LITELLM_MASTER_KEY" -H "Content-Type: application/json" \
  -d '{"model": "garage/garage-lund/qwen3:32b", "stream": true,
       "messages": [{"role": "user", "content": "Hej från GarageAI!"}]}'
```

If LiteLLM can't reach the garage but the host can (`curl http://<garage-mesh-ip>:11434/v1/models`),
test from inside the container:

```bash
docker compose -f /opt/garageai/litellm/docker-compose.yml exec litellm python -c \
  "import urllib.request; print(urllib.request.urlopen('http://<garage-mesh-ip>:11434/v1/models', timeout=5).read()[:200])"
```

## 7. The two marketplace tiers

`register-node.sh` registers every model twice:

| Model name                     | Routes to                     | Sold as |
|--------------------------------|-------------------------------|---------|
| `garage/garage-lund/qwen3:32b` | that one garage only          | Booked, single-tenant access to a named garage |
| `qwen3:32b`                    | any garage offering the model | Cheaper pool access, load-balanced |

Give a buyer a LiteLLM virtual key limited to what they bought:

```bash
curl -X POST "$LITELLM_URL/key/generate" \
  -H "Authorization: Bearer $LITELLM_MASTER_KEY" -H "Content-Type: application/json" \
  -d '{"models": ["garage/garage-lund/qwen3:32b"], "max_budget": 25, "duration": "30d",
       "metadata": {"buyer": "acme-ab", "booking": "b_123"}}'
```

LiteLLM's spend tracking per key and model is the basis for operator payouts.

## Operations

- **Back up:** `/opt/garageai/netbird/config.yaml`, `/opt/garageai/litellm/.env`, and the Docker
  volumes `netbird_data` and `pgdata`.
- **Upgrade LiteLLM:** change `LITELLM_VERSION` in `.env`, then `docker compose pull && docker compose up -d`.
- **Scale out later:** the same files move unchanged to a bigger VPS, or LiteLLM and NetBird can
  be split onto separate machines.

## Locked-down admin API

Traefik only exposes the LLM API publicly (`/v1/chat/completions`, `/v1/completions`,
`/v1/embeddings`, `/v1/models`, `/v1/messages`, `/v1/responses`, the non-`/v1` aliases and
`/health/liveliness|readiness`), rate limited per client IP. Every other route (admin UI, `/sso`,
`/openapi.json`, `/model/*`, `/key/*`, `/user/*`, `/spend/*`, `/health`) is only routed when the
request carries the master key, and returns 403 otherwise. Buyer keys therefore cannot read
`/model/info` (garage addresses) or call `/health` (real inference on every garage).
After rotating the master key, recreate the container (`docker compose up -d`) so the rule follows.

Responses are stripped of everything that would tell a buyer which garage or provider served
them: the `x-litellm-*` headers (upstream `api_base`, deployment id, runtime model name, timing)
and the upstream server's own headers that LiteLLM forwards as `llm_provider-*` (Cloudflare,
uvicorn, rate-limit headers). Only `x-litellm-call-id` (for support) and
`x-litellm-response-cost` remain. Traefik has no wildcard header removal, so the
`llm_provider-*` list is explicit; after adding a provider, check one response with
`curl -D -` and extend the list if a new header appears.

## Agent traffic

Most traffic comes from coding agents (Claude Code on `/v1/messages`, Codex on
`/v1/responses`) with prompts of 100k+ tokens. Three settings in `litellm/` exist for them:

- `enable_pre_call_checks`: LiteLLM counts the prompt first and answers 400 when it does not
  fit the model's window, instead of letting the garage break the stream.
- `garageai_callbacks.py` (mounted into the container, registered under
  `litellm_settings.callbacks`): adds `"items": {}` to array parameters in tool schemas that
  lack it. Claude Code sends such schemas; without the fix LiteLLM's token counter fails and
  the pre-call check is silently skipped.
- `stream_timeout: 300`, `request_timeout: 600`: prefilling an agent-sized prompt takes
  minutes on home GPUs and busy providers; the stream timeout includes the wait for the
  first token.

## Garage health checks (levels 1 and 2)

`health/garageai-health.py` (installed to `/usr/local/bin/garageai-health`, run every minute by
`garageai-health.timer`) checks every garage from the gateway, which is the only place that can
reach the mesh:

| Level | Check | Cost |
|---|---|---|
| 1. Tunnel | NetBird peer connected to the management server | none |
| 2. Runtime | `GET /v1/models` on the garage's runtime over the mesh (5 s timeout) | none, no generation |
| 3. Model | Acceptance test and hourly probe, run by the portal through LiteLLM | GPU time |

It fetches its targets from the portal (`/functions/v1/gateway-targets`) and reports to
`/functions/v1/gateway-health-report`, authenticating with the LiteLLM master key. Until those
endpoints exist it reads `/etc/garageai/health-targets.json` and only writes
`/var/lib/garageai-health/last.json`. Config: `/etc/garageai/gateway-health.env` (root, 0600).

The portal uses levels 1 and 2 to delist and relist within a minute without any token: relisting
needs only the stored runtime key and the garage's offered models. Level 3 sets the grade and
removes a model only after repeated real failures.

## Backups

`backup/garageai-backup` (installed to `/usr/local/sbin`, run nightly by `garageai-backup.timer`)
writes the LiteLLM database, `.env`, NetBird `config.yaml`, the NetBird data volume and the
certificates to `/var/backups/garageai`, keeping 7 days. This only protects against corruption
and mistakes: copy the archives off-site (encrypted) as well.

## Security notes

- The LiteLLM master key stays on the gateway side. Garage nodes never see it, and
  `garageai-connect.sh` does not need it.
- Postgres has no published port and is only on LiteLLM's internal network.
- Prompts are processed on the operator's machine. Mesh encryption protects them in transit,
  not on the garage itself. Verified operators and a data processing agreement are needed
  before selling to buyers with sensitive data.
- Registration goes through the portal's `register-node` endpoint with a per-garage token;
  the master key never leaves the gateway and the portal.

## Security checklist

Run after every LiteLLM or NetBird upgrade, after adding a provider, and before giving an
external tester a key. All commands from the gateway; `$BUYER` is an ordinary buyer key.

1. **Nothing but the LLM API is public.** Expect 403 for the admin UI and routes, 401 for
   anonymous `/v1/models`:
   ```bash
   for p in /ui /openapi.json /model/info /health /key/info; do printf "%-14s " $p; curl -s -o /dev/null -w "%{http_code}\n" https://llm.garageai.eu$p -H "Authorization: Bearer $BUYER"; done
   curl -s -o /dev/null -w "%{http_code}\n" https://llm.garageai.eu/v1/models
   ```
2. **Responses do not reveal the garage or provider.** Read every header of one response per
   tier (pool and `garage/...`) and per provider. Only `x-litellm-call-id` and
   `x-litellm-response-cost` may remain; no `x-litellm-model-*`, no `llm_provider-*`:
   ```bash
   curl -s -D - -o /dev/null https://llm.garageai.eu/v1/chat/completions -H "Authorization: Bearer $BUYER" \
     -H "Content-Type: application/json" -d '{"model":"<model>","max_tokens":5,"messages":[{"role":"user","content":"hi"}]}' | grep -i "^x-\|^llm"
   ```
   A new header means the `llm_provider-*` list in `docker-compose.yml` needs that name.
   (Found 2026-10-06: `x-litellm-model-api-base` carried the provider's URL to every buyer.)
3. **Mesh policy is one-directional.** Only `gateway -> garages` (and named garage routes);
   no `Default` policy; garages cannot reach each other or the gateway:
   ```bash
   curl -s -H "Authorization: Token $NB" https://netbird.garageai.eu/api/policies | jq -c '.[] | {name, src:[.rules[].sources[].name], dst:[.rules[].destinations[].name]}'
   ```
4. **No live setup keys.** Every key `valid: false` or `revoked: true`, and every peer pinned
   to its own `garage-<name>` group:
   ```bash
   curl -s -H "Authorization: Token $NB" https://netbird.garageai.eu/api/setup-keys | jq -c '.[] | {name, valid, revoked}'
   ```
5. **Portal exposes nothing anonymously** beyond the public RPCs (`garage_public_*`): anon reads
   of `garages`, `garage_tokens`, `garage_runtime_secrets`, `profiles` return `[]` or a
   permission error, and `gateway-targets` / `gateway-health-report` answer 401 without
   `x-gateway-key`.
6. **Health service and backups are running:** `systemctl status garageai-health.timer
   garageai-backup.timer`, and `/var/backups/garageai` has today's archive.
7. **Secrets are where they should be and nowhere else:** `.env` and `config.yaml` are
   0600/root, nothing secret in the repo (`git grep -I sk-`), and any key pasted in a chat or
   ticket has been rotated.

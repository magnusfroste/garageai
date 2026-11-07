# GarageAI core: LiteLLM + Bridge tunnel hub on one VPS

The GarageAI core runs on a single EU VPS:

- **LiteLLM** is the OpenAI-compatible API buyers call. It routes each request through the
  tunnel to a garage, counts tokens per key and model, and handles retries and failover.
- **The Bridge tunnel hub** (`garageai-hub`) is the gateway's end of every garage's tunnel.
  GarageAI Bridge on the garage carries a WireGuard tunnel in user space and dials out to the
  hub, so garages open no inbound ports and are never exposed to the public internet. The hub
  is plain WireGuard in the kernel; no tunnel service sits in the path. Design:
  [docs/bridge-tunnel.md](../../docs/bridge-tunnel.md).

```
                           ┌──────────────── EU VPS ──────────────────────────────┐
 Buyers ──HTTPS──▶ :443 ──▶│ Traefik ─┬─▶ api.garageai.eu / llm  → LiteLLM ─▶ Postgres
                           │ (Docker) └─▶ ops.garageai.eu        → Operations Center
                           │                                         │ http://10.66.x.y:<port>/v1
 Garages ─UDP 51830 (out)─▶│ wg-hub 10.66.0.1/16 (kernel, host) ◀────┘
                           │ garageai-hub: peers from the portal, nftables one way
                           └──────────────────────────┬───────────────────────────┘
                                 WireGuard, one tunnel per garage
                      ┌───────────────┬───────────────┴───────────┐
                  garage-lund     garage-malmo                garage-…
                  (Ollama)        (llama.cpp)                 (vLLM)
```

## Components

| Component | Where | What |
|---|---|---|
| Traefik | Docker, `traefik/` (`/opt/garageai/traefik`) | The one reverse proxy and set of Let's Encrypt certificates. LiteLLM and Ops join its network `garageai_proxy` via labels. Runbook and cut-over: [traefik/README.md](traefik/README.md) |
| LiteLLM + Postgres | Docker, `litellm/` (`/opt/garageai/litellm`) | LiteLLM's documented deployment. Pinned image tags make upgrades and rollbacks one command |
| **Tunnel hub** (`garageai-hub`) | **Directly on the host**, `hub/` | Kernel WireGuard interface `wg-hub` (10.66.0.1/16, UDP 51830), peers from the portal's `gateway-targets`, nftables table `inet garageai_tunnel`. Docker containers, LiteLLM included, reach garages through the host's route to 10.66.0.0/16 |
| Guard | Host, `guard/` | Blocks clients that guess API keys or probe the gateway |
| Health | Host, `health/` | Levels 1 and 2 health checks for every garage, reported to the portal |
| Operations Center | Host collector + Docker nginx, `ops/` | The dashboard at ops.garageai.eu |
| Backup | Host, `backup/` | Nightly, encrypted, off-site |

## 0. VPS, DNS and firewall

- **VPS:** Hetzner CX22 (2 vCPU, 4 GB, ~€4/month) is enough for a PoC; CX32 gives headroom.
  Ubuntu 24.04, location Falkenstein, Nuremberg or Helsinki.
- **DNS:** A records pointing at the VPS: `api.garageai.eu` (the public API name in the docs since 2026-10-12), `llm.garageai.eu` (the original name; it keeps working for existing clients, and it alone carries the master key's admin API) and `ops.garageai.eu`.
- **Hetzner Cloud Firewall, inbound:**

  | Port | Protocol | Purpose |
  |---|---|---|
  | 22 | TCP | SSH, restricted to your own IP |
  | 80, 443 | TCP | Traefik (Let's Encrypt, LiteLLM, Operations Center) |
  | 51830 | UDP | The Bridge tunnel hub (`wg-hub`): garages dial out to it |

```bash
curl -fsSL https://get.docker.com | sh
sudo apt install -y wireguard-tools nftables
git clone https://github.com/magnusfroste/garageai /opt/garageai/repo
```

## 1. Traefik

```bash
sudo install -d /opt/garageai/traefik
sudo cp /opt/garageai/repo/infra/gateway/traefik/docker-compose.yml /opt/garageai/traefik/
cd /opt/garageai/traefik && sudo docker compose up -d
```

It creates the Docker network `garageai_proxy` and the volume `garageai_traefik_letsencrypt`.
Streaming passes through without buffering: the entrypoint's read, write and idle timeouts are
off, because agent-sized prompts stream for minutes.

## 2. The Bridge tunnel hub

`hub/garageai-hub.py` (installed to `/usr/local/sbin/garageai-hub`, run as root by
`garageai-hub.service`):

- creates `wg-hub`, a kernel WireGuard interface, 10.66.0.1/16, listening on UDP 51830;
- every 10 s reads the garages with transport `tunnel` from the portal (`gateway-targets`:
  `tunnel_public_key`, `tunnel_ip`) and adds, updates and removes peers so `wg-hub` matches.
  Each peer is a public key and its /32; there are no endpoints, the hub learns each garage's
  address from its handshakes, whatever NAT it is behind. If the portal cannot be read it keeps
  the peers it has and never empties the hub because of a failed read;
- loads `table inet garageai_tunnel`: on `wg-hub` only replies to connections the gateway opened
  are accepted (`ct state established,related`), so a garage can reach neither the gateway nor
  another garage;
- writes `/var/lib/garageai-hub/status.json` (latest handshake, endpoint, bytes per garage) for
  the health check and the Operations Center.

Config: `/etc/garageai/gateway-health.env` (portal URL and gateway key, shared with the health
service) and the hub's private key `/etc/wireguard/garageai-hub.key` (root, 0600; AppArmor
confines `wg` to `/etc/wireguard`). The key never leaves the gateway; its public half is the
`hub_public_key` the portal hands to every garage. Lose it and every garage must connect again,
so it is in the nightly backup. `/etc/garageai/hub-peers.json` can hold static test peers.

```bash
sudo sh -c 'umask 077; wg genkey > /etc/wireguard/garageai-hub.key'
sudo wg pubkey < /etc/wireguard/garageai-hub.key        # the portal's hub_public_key
sudo install -m 0755 /opt/garageai/repo/infra/gateway/hub/garageai-hub.py /usr/local/sbin/garageai-hub
sudo install -m 0644 /opt/garageai/repo/infra/gateway/hub/garageai-hub.service /etc/systemd/system/
sudo systemctl daemon-reload && sudo systemctl enable --now garageai-hub
sudo wg show wg-hub                                       # one peer per tunnel garage
```

LiteLLM reaches a garage at `http://<tunnel ip>:<port>/v1`. On the garage, Bridge answers on
its tunnel address only for the runtime's port, as an HTTP proxy with a path allowlist, and
forwards to the runtime on `127.0.0.1`.

## 3. LiteLLM

```bash
cp -r /opt/garageai/repo/infra/gateway/litellm /opt/garageai/litellm
cd /opt/garageai/litellm
cp .env.example .env
nano .env             # LITELLM_DOMAIN, PROXY_NETWORK (garageai_proxy), keys and password (see comments)
docker compose up -d
curl https://api.garageai.eu/health/liveliness
```

`config.yaml` has no models: garages are added at runtime (step 5). Its router settings retry
on another garage and take a failing garage out of rotation for 60 seconds, because residential
nodes drop out more often than datacenter ones.

## 4. Onboard a garage

The operator starts their runtime (Ollama, LM Studio, llama.cpp, vLLM, Xyntetik Runner, Unsloth)
and follows the wizard in the portal, which gives them one command: it installs
[GarageAI Bridge](../../cli/README.md) and runs `garageai connect`. Bridge creates its tunnel
key on the machine, sends only the public key, gets its tunnel IP and the hub's key and
endpoint from the portal, and registers the garage. Within 10 s the hub adds it as a peer.

## 5. Register the garage in LiteLLM

The portal registers garages (`register-node`, with a per-garage token). By hand, from
anywhere that can reach the LiteLLM API:

```bash
export LITELLM_URL=https://llm.garageai.eu LITELLM_MASTER_KEY=sk-...   # admin API: llm only, never api
infra/gateway/register-node.sh add garage-lund 10.66.0.7 11434 qwen3:32b gemma4:27b
infra/gateway/register-node.sh list
infra/gateway/register-node.sh remove garage-lund qwen3:32b     # when they stop sharing
```

If the runtime requires an API key (for example vLLM started with `--api-key`), pass it in
`NODE_API_KEY`. LiteLLM stores it encrypted and buyers never see it:

```bash
NODE_API_KEY=sk-... infra/gateway/register-node.sh add garage-lund 10.66.0.7 8000 glm-5.3-flash
```

First end-to-end test:

```bash
curl https://api.garageai.eu/v1/chat/completions \
  -H "Authorization: Bearer $LITELLM_MASTER_KEY" -H "Content-Type: application/json" \
  -d '{"model": "garage/garage-lund/qwen3:32b", "stream": true,
       "messages": [{"role": "user", "content": "Hej från GarageAI!"}]}'
```

If LiteLLM can't reach the garage but the host can (`curl http://<tunnel-ip>:11434/v1/models`),
test from inside the container:

```bash
docker compose -f /opt/garageai/litellm/docker-compose.yml exec litellm python -c \
  "import urllib.request; print(urllib.request.urlopen('http://<tunnel-ip>:11434/v1/models', timeout=5).read()[:200])"
```

If the host cannot reach it either, `sudo wg show wg-hub` shows whether the garage's latest
handshake is recent (at most about 25 s old: Bridge's keepalive).

## 6. The two marketplace tiers

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

- **Back up:** the nightly backup covers it all (see Backups below).
- **Upgrade LiteLLM:** pull the new image first, run `tests/test_callbacks.py` inside it (`docker run --rm -i -v /opt/garageai/litellm/garageai_callbacks.py:/app/garageai_callbacks.py:ro --entrypoint python ghcr.io/berriai/litellm:vX - /app/garageai_callbacks.py < infra/gateway/litellm/tests/test_callbacks.py`), then in a quiet window change `LITELLM_VERSION` in `.env` and `docker compose up -d litellm` (about 50 s). Afterwards check every routed model answers, /v1/messages with a system message mid-conversation, streaming, and that no `x-litellm-*` or `llm_provider-*` header reaches the client. Rollback: set the old version back; the image is still cached.
- **Upgrade Traefik:** change the pinned tag in `/opt/garageai/traefik/docker-compose.yml`, then
  `sudo docker compose pull && sudo docker compose up -d` there at a quiet time (a few seconds
  without HTTPS).
- **Scale out later:** the same files move unchanged to a bigger VPS. More hubs (one per region)
  belong to the control plane (docs/architecture.md).

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
  `litellm_settings.callbacks`):
  - adds `"items": {}` to array parameters in tool schemas that lack it (Claude Code sends
    such schemas; without it LiteLLM's token counter fails and its pre-call check is skipped);
  - checks prompt + requested `max_tokens` against 95 % of the model's window before
    routing (runtimes such as vLLM count both; the margin covers tokenizer differences);
  - answers in the words agents act on: OpenAI format `code: "context_length_exceeded"` and
    "maximum context length is N tokens", Anthropic format "prompt is too long: N tokens >
    M maximum" or "input length and `max_tokens` exceed context limit"; upstream context
    errors that slip through are rewritten the same way, without internal names;
  - gives garages that wait a start (rule 11): a market garage serving the same model as an
    auto entry, but not on the auto list, with fewer than 10 market requests in 24 h, is put first
    for 25 % of small (≤ 16k-token) `garageai/auto` requests that leave room for the answer
    (no `max_tokens` below 1000: the first garages run a thinking model). It must fit the prompt and start it
    in time; a failure pauses it for 30 min and the retry goes down the auto list. Look for
    `garageai: start:` in the LiteLLM log.
  - checks the runtime's token count (rule 12): LiteLLM bills from the `usage` the runtime
    reports and the spend log keeps no answer text, so the gateway counts every market answer
    itself. Reported completion tokens (less hidden reasoning tokens) may exceed that count by
    30 % + 64 (a tool call 2× + 200: its structure is not in the answer, and a runtime that drops
    a malformed call still reports its tokens), the prompt by 2× + 500. Beyond that, `garageai: usage:` names the garage, the
    call id and both counts; a garage is routed around for 60 min, a provider only logged. The
    spend row is not changed: an admin corrects it from the line. Hidden reasoning cannot be
    counted, so the line also shows the reported reasoning tokens. The same pass logs tool calls
    that came back without arguments as `garageai: tool:` (measured only, so far).
  - switches thinking off unless asked for (rule 13): a market request that says nothing about
    reasoning goes to an Ollama or vLLM deployment with `reasoning_effort: "none"`; a buyer who
    wants reasoning sets `reasoning_effort` (or `thinking` on /v1/messages). Providers, Runner,
    private machines and probes are untouched. Look for `garageai: thinking off:` in the log.
  - a conversation that has a deployment keeps it (rule 7, tightened): rule 11 never takes a
    request from a conversation whose prefix is cached elsewhere.
- `stream_timeout: 300`, `request_timeout: 600`: prefilling an agent-sized prompt takes
  minutes on home GPUs and busy providers; the stream timeout includes the wait for the
  first token.

## Garage health checks (levels 1 and 2)

`health/garageai-health.py` (installed to `/usr/local/bin/garageai-health`, run every minute by
`garageai-health.timer`) checks every garage from the gateway, which is the only place that can
reach the garages through their tunnels:

| Level | Check | Cost |
|---|---|---|
| 1. Tunnel | The garage's latest handshake with the hub (`/var/lib/garageai-hub/status.json`) is under 3 minutes old | none |
| 2. Runtime | `GET /v1/models` on the garage's runtime through the tunnel (5 s timeout) | none, no generation |
| 3. Model | Acceptance test and hourly probe, run by the portal through LiteLLM | GPU time |

It fetches its targets from the portal (`/functions/v1/gateway-targets`) and reports to
`/functions/v1/gateway-health-report`, authenticating with the LiteLLM master key. Until those
endpoints exist it reads `/etc/garageai/health-targets.json` and only writes
`/var/lib/garageai-health/last.json`. Config: `/etc/garageai/gateway-health.env` (root, 0600).
Providers (a public HTTPS endpoint) skip level 1; every other target is a tunnel garage.

The portal uses levels 1 and 2 to delist and relist within a minute without any token: relisting
needs only the stored runtime key and the garage's offered models. Level 3 sets the grade and
removes a model only after repeated real failures.

## Operations Center

`ops/` is a read-only dashboard for running the platform, separate from the portal's admin
(which handles garages, models, money and keys). It shows the gateway host, containers,
timers, backups and certificates; the Bridge tunnel hub (each garage's latest handshake,
address and traffic); every garage and provider (tunnel, runtime, uptime, vLLM load,
provider DNS and TLS); gateway traffic per model (TTFT p50/p95, tokens, spend) and error
classes; the portal and site seen from outside; and a list of current alerts.

- `garageai-ops-collect` (installed to `/usr/local/bin`, run every minute by
  `garageai-ops.timer`) writes `/var/lib/garageai-ops/www/ops.json`. It reads the health
  service's config and output, the hub's status, Docker and LiteLLM's database. The file holds no
  keys, tokens or prompts.
- `ops/docker-compose.yml` serves the page with nginx behind Traefik at
  `https://ops.garageai.eu` (DNS A record to the gateway, grey cloud), with basic auth and
  a rate limit. The password is generated on the server into `/etc/garageai/ops-password`
  (root, 0600); read it once over SSH and keep it in a password manager. The hash lives in
  `/opt/garageai/ops/.env`.
- `ops/index.html` is a single static page with a left menu and one page per area:
  Dashboard (status and KPIs, each linking to its page), Supply, Traffic, Tunnel, Server,
  Updates and Logs. Each alert carries the page it belongs to, so the menu shows a count per
  page. The sidebar shows what is running (main, LiteLLM, Traefik, host OS); the top
  bar has refresh, pending updates and the signed-in user. `ops/nginx.conf` answers `/whoami`
  with the user Traefik passes in `X-Ops-User`; signing out makes the browser forget the
  basic-auth credentials.

### Planned stops

When an operator stops a garage on purpose, or a known outage is being fixed, silence it. Its alerts then become info with the reason: no Telegram, no reminders, and no "Resolved" when the silence starts. Supply shows it as *planned stop*. The silence ends by itself, and both ends are events.

```bash
sudo garageai-ops-collect --silence autoversio 36h "DGX Spark reinstall"
sudo garageai-ops-collect --unsilence autoversio
```

It matches the exact garage name: silencing `autoversio` does not silence `autoversio-south`. Later, Bridge's `garageai pause` will set this from the operator's side (docs/bridge.md, phase 1).

### Onboarding

The *Onboarding* page lists garages created in the portal but not registered yet. For each it shows:
- the last step the connect script reported, and its message,
- the problem codes from the garage profile (`--doctor --json`),
- whether the garage's tunnel is up,
- how long it has waited.

A garage is stuck when its last step failed or waited for the operator for 30 minutes. That is a warning, so it reaches Telegram once. A garage that never reported a step is noted after a day. The full profile and the fixes are on the garage's page in the portal admin.

### Business

The *Business* page reads the portal's `ops-summary` every 5 minutes. It authenticates like `gateway-targets`, with `x-gateway-key`, and returns aggregates only: no e-mail addresses, names, keys or Stripe ids. The page shows:
- buyers,
- top-ups and credit outstanding,
- Stripe status (last top-up, failed webhooks if tracked, checkouts not completed),
- earnings per operator and whether they have accepted the terms,
- garages created but never registered, with their last onboarding step,
- recent admin activity.

Alerts:
- failed Stripe webhooks,
- 3 or more checkouts not completed,
- 5 or more active buyers under $1.

The *Models* page also shows each deployment's cache-read price and warns when one is missing. *Quality* shows the share of input tokens each garage served from its prefix cache.

### Chat

The *Chat* page talks to any public model, or to one garage through its dedicated route, the way a buyer would. Each reply shows time to first token, tokens per second, tokens in and out (and reasoning), with the reasoning collapsible.

Quick prompts check a newly loaded model:
- identity,
- exact instruction following,
- Swedish,
- reasoning,
- code,
- JSON mode,
- a tool call,
- output speed (about 400 words),
- a system message mid-conversation.

The browser never sees a key. nginx forwards only `/llm/v1/chat/completions` and `/llm/v1/models` to the LiteLLM container on the shared Docker network, and adds `OPS_CHAT_KEY`:
- That is a LiteLLM key called `ops-chat`, with a budget of $5 per 30 days.
- It lives in `/opt/garageai/ops/.env` and is filled into `nginx.conf.template` when the container starts.
- Every other path under `/llm/` returns 404.
- The page sits behind the same basic auth as the rest.

### Models and quality

The *Models* page lists every LiteLLM deployment buyers can call: public name, tier (pool or dedicated), garage, runtime model, the context window and output cap the gateway enforces, and the price. Each deployment is checked against its garage. The gateway holds the tunnels, so it reads each runtime's own `/v1/models` every 10 minutes, which the portal cannot do. It alerts on:
- no context window (the gateway cannot reject over-long prompts),
- a gateway window larger than the runtime's,
- a routed model the runtime no longer serves,
- no price.

*Supply* gets *Quality per garage* from LiteLLM's spend logs (probes excluded), for the last 24 h and as 7-day sparklines:
- success rate,
- time to first token (p50, p95),
- output speed: tokens per second after the first token, median over replies of 20+ tokens.

Further alerts:
- below 90 % success over at least 20 requests,
- a vLLM garage with requests waiting 5 minutes in a row, meaning the garage is full.

A garage that is still onboarding does not count as downtime.

### History and events

The collector keeps history in `/var/lib/garageai-ops/history.db` (SQLite):
- **Samples:** one per minute for the host and every garage, kept for 30 days.
- **Events:** kept for 180 days. Logged automatically:
  - alerts opened and resolved, with how long they lasted,
  - gateway reboots and container restarts,
  - version changes and files deployed on the gateway,
  - merges to `main`,
  - garages added to or removed from supply,
  - clients blocked by the guard,
  - notes.

Every 5 minutes it writes `history.json` for the page. Traffic history comes straight from
LiteLLM's spend logs (30 days).

The page shows:
- **Traffic:** requests and time to first token per hour. Single spikes are clipped and marked.
- **Supply:** uptime per garage and hour, and vLLM load.
- **Server:** memory, disk and load.
- **Events:** the timeline.

Add a note, for example for a maintenance window:
```bash
sudo garageai-ops-collect --note "Maintenance: LiteLLM upgrade"
```

### Telegram alerts

The collector sends alerts to one Telegram chat. It is one-way: the bot only sends, and never
acts on messages it receives.

- Critical alerts are sent after 2 minutes in a row and repeated every 4 hours while still open.
- Warnings are sent after 15 minutes in a row.
- When an alert that was sent clears, a "Resolved" message follows.
- A morning report (status, supply, last 24 h traffic and spend, backup, updates) is sent at
  07:00 Europe/Stockholm.

A send that fails is retried the next minute, and the Operations Center shows the failure.

Setup:

1. In Telegram, ask @BotFather for `/newbot` and copy the token.
2. On the gateway, store the token without echoing it:
   ```bash
   sudo install -m 600 /dev/null /etc/garageai/telegram.env
   read -rsp "Bot token: " T && printf 'TELEGRAM_BOT_TOKEN=%s\n' "$T" | sudo tee /etc/garageai/telegram.env >/dev/null; unset T; echo
   ```
3. Send `/start` to the new bot from your own account, then save that chat and get a test message:
   ```bash
   sudo garageai-ops-collect --telegram-setup
   ```

`sudo garageai-ops-collect --telegram-test` sends another test message.

## Guard: blocking key guessing

`guard/garageai-guard.py` (installed to `/usr/local/sbin/garageai-guard`, run every minute by
`garageai-guard.timer`) reads Traefik's access log. It blocks a source IP for 24 h, or 7 days if
it comes back within a week, when within 2 minutes it gets either:
- 30 invalid API keys (401), or
- 60 requests to paths the gateway does not serve (403).

The block is an nftables set with timeouts in `table inet garageai_guard`, dropped before
Docker's NAT. It never blocks private (10.0.0.0/8 covers the tunnel) or loopback addresses, the gateway itself, or
`GUARD_ALLOW` in `/etc/garageai/guard.env`. Blocks and dropped packets show in the Operations
Center (Traffic), and each block is an event.

On 2026-10-03 two Azure VMs sent 93,026 requests with invalid keys at about 15/s, which is under
the 20/s rate limit. Replaying that log, the guard would have blocked both within a minute, and
it finds no false positives in the normal traffic since.

```bash
sudo garageai-guard --replay 2026-10-03T21:48:00Z 2026-10-03T23:35:00Z   # who would have been blocked
sudo garageai-guard --unban 203.0.113.7
```

Install:
```bash
sudo install -m 0755 guard/garageai-guard.py /usr/local/sbin/garageai-guard
sudo install -m 0644 guard/garageai-guard.service guard/garageai-guard.timer /etc/systemd/system/
sudo systemctl daemon-reload && sudo systemctl enable --now garageai-guard.timer
```

## Backups

`backup/garageai-backup` (installed to `/usr/local/sbin`, run nightly by `garageai-backup.timer`)
writes the LiteLLM database, its `.env`, `config.yaml` and compose file, the hub's key
(`/etc/wireguard/garageai-hub.key`), `/etc/garageai`, the ops `.env`, the Traefik compose file
and its Let's Encrypt volume (`garageai_traefik_letsencrypt`) to `/var/backups/garageai`, keeping
7 days. With `BACKUP_PASSPHRASE` and `BACKUP_S3_REMOTE` in `/etc/garageai/backup.env` the archive
is encrypted and copied off-site (see the script's header; restore: docs/architecture.md).

## Security notes

- The LiteLLM master key stays on the gateway side. Garages never see it, and Bridge does not
  need it.
- Postgres has no published port and is only on LiteLLM's internal network.
- Prompts are processed on the operator's machine. The tunnel's encryption protects them in transit,
  not on the garage itself. Verified operators and a data processing agreement are needed
  before selling to buyers with sensitive data.
- Registration goes through the portal's `register-node` endpoint with a per-garage token;
  the master key never leaves the gateway and the portal.

## Security checklist

Run after every LiteLLM or Traefik upgrade, after adding a provider, and before giving an
external tester a key. All commands from the gateway; `$BUYER` is an ordinary buyer key.

1. **Nothing but the LLM API is public.** Expect 403 for the admin UI and routes, 401 for
   anonymous `/v1/models`:
   ```bash
   for h in api llm; do for p in /ui /openapi.json /model/info /health /key/info; do printf "%-4s %-14s " $h $p; curl -s -o /dev/null -w "%{http_code}\n" https://$h.garageai.eu$p -H "Authorization: Bearer $BUYER"; done; done
   curl -s -o /dev/null -w "%{http_code}\n" https://api.garageai.eu/v1/models
   ```
2. **Responses do not reveal the garage or provider.** Read every header of one response per
   tier (pool and `garage/...`) and per provider. Only `x-litellm-call-id` and
   `x-litellm-response-cost` may remain; no `x-litellm-model-*`, no `llm_provider-*`:
   ```bash
   curl -s -D - -o /dev/null https://api.garageai.eu/v1/chat/completions -H "Authorization: Bearer $BUYER" \
     -H "Content-Type: application/json" -d '{"model":"<model>","max_tokens":5,"messages":[{"role":"user","content":"hi"}]}' | grep -i "^x-\|^llm"
   ```
   A new header means the `llm_provider-*` list in `docker-compose.yml` needs that name.
   (Found 2026-10-06: `x-litellm-model-api-base` carried the provider's URL to every buyer.)
3. **The tunnel is one-directional.** On `wg-hub` only replies to the gateway's own
   connections are accepted; anything a garage starts is dropped, and every peer allows only its
   own /32:
   ```bash
   sudo nft list table inet garageai_tunnel     # established,related accept, then counter drop, in input and forward
   sudo wg show wg-hub allowed-ips              # one /32 in 10.66.0.0/16 per peer, nothing wider
   ```
4. **Portal exposes nothing anonymously** beyond the public RPCs (`garage_public_*`): anon reads
   of `garages`, `garage_tokens`, `garage_runtime_secrets`, `profiles` return `[]` or a
   permission error, and `gateway-targets` / `gateway-health-report` answer 401 without
   `x-gateway-key`.
5. **Hub, health service and backups are running:** `systemctl status garageai-hub
   garageai-health.timer garageai-backup.timer`, and `/var/backups/garageai` has today's archive.
6. **Secrets are where they should be and nowhere else:** `.env` and `config.yaml` are
   0600/root, nothing secret in the repo (`git grep -I sk-`), and any key pasted in a chat or
   ticket has been rotated.

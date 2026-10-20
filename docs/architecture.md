# Where logic lives

GarageAI has four parts: the **portal** (app.garageai.eu, Lovable/Supabase, repo
`garageai-portal`), the **gateway** (LiteLLM, Traefik and NetBird on the Hetzner VPS), **Bridge**
on each garage (`cli/`), and the **Operations Center** (`infra/gateway/ops/`). This page says which
part decides what. It applies to every change, whoever makes it: here, in Lovable, or on the
website.

Decided 2026-10-11, when `garageai/auto` (ordered failover across a curated list) was designed.

## The rule

| Kind of logic | Lives in | Pace | Examples |
|---|---|---|---|
| **What should hold**: decisions by people, business terms | Portal | minutes to days | which models are sold and at what price, the auto list and its order, garages, owners, terms, earnings, accounts and keys, onboarding (link, wizard) |
| **Carrying it out right now** | Gateway | milliseconds to seconds | picking a deployment, failover to the next order, cooldown after failures, context checks before the call, request fixes (callbacks), blocking attacks (guard), health checks every minute |
| **What a garage sees about itself** | Bridge | per run, every 5 min | doctor, the heartbeat with models, contexts and its own report, the knowledge base |
| **Measuring and telling** | Operations Center | 5 s to 1 min | Live, alerts and Telegram, time to live, drift between main and what is deployed |

**In short:** anything that must react within seconds lives in the gateway. Anything a person
decides lives in the portal. Anything the machine knows about itself lives in Bridge. Anything
measured lives in the Operations Center.

## How they connect

- **Portal → gateway:** `syncModels` turns the portal's desired state (curated models, garages,
  auto list, prices) into LiteLLM deployments (`/model/new`, `/model/delete`). LiteLLM keeps them in
  its own database (`STORE_MODEL_IN_DB`).
- **Gateway → portal:** the health service reports every minute (`gateway-health-report`). The
  portal marks models healthy or unhealthy and runs the daily acceptance tests.
- **Bridge → portal:** `register-node`, `node-heartbeat`, `onboarding-report` and the browser link.
- **Portal and gateway → Operations Center:** `gateway-targets` and `ops-summary` (the portal,
  with the gateway key), LiteLLM's database and API, NetBird's API, the host itself.

## Why this split survives failures

- **The portal is down:** the gateway keeps serving with the last routing it was given, and failover
  keeps working. Only changes stop.
- **The gateway restarts:** LiteLLM reads its routing back from its database. A decision in the
  portal must never be made from gateway data taken while the gateway itself is down. On
  2026-10-11 the daily acceptance test of garage-m2 ran during a 30-second planned restart, failed
  because of the restart and took the garage out of routing. Fixed: a test that fails because the
  gateway is unreachable is inconclusive.
- **A garage goes down:** the gateway skips it within seconds (cooldown, the next order). The
  portal takes it out of routing on the next health report. The Operations Center alerts.

## Rules for changes

- Logic in the portal that the gateway depends on (routing, prices, health) gets tests, asked for
  in every Lovable request. The code is in `garageai-portal` on GitHub and is read before it is
  relied on.
- A fast reaction (failover, retries, timeouts, blocking) is never built as a portal function that
  runs on a schedule.
- The gateway does not hold business decisions. Its configuration (`config.yaml`, callbacks) holds
  how to carry out requests, never which models are for sale or at what price.
- Everything that runs on the gateway is in this repo, and the Operations Center reports drift
  between main and what is deployed.

## The pool: one name, several garages

Decided 2026-10-12, after the OpenRouter comparison. Operators have no idea what to charge, and a
pool price set by GarageAI would be administration that grows with the catalogue. So:

- **Every garage model goes into the pool.** The public name (`qwen/qwen3-4b`) is the pool; the
  dedicated name (`garage/<garage>/<model>`) is an address for pinning one garage, at that garage's
  own price. There is no opt-in. A garage that wants out of the pool pauses.
- **The price is prefilled** at onboarding from the market (OpenRouter's price for the same model,
  or its size class). The operator can change it or go on. The portal sets each pool deployment's
  `weight` from the price: cheaper garages get a larger share.
- **The gateway distributes** each request among the garages that are up, in proportion to their
  weights (`simple-shuffle`), skips garages whose context window does not fit the prompt, retries
  once on another garage, and cools down a failing one. Not latency-based: that starved the slower
  garages, and one slow request skewed it. Not strictly cheapest-first: that starves everyone but
  one. Weighted distribution gives every garage that is up traffic and pay.
- **The buyer sees** what OpenRouter shows: "from" (the cheapest garage), the expected price (the
  weighted mean), and the cap (the most expensive garage), with the list of garages and their
  prices. The buyer pays the price of the garage that answered.
- `garageai/auto` is something else: a curated, ordered list across models, not a pool.

## Two kinds of supply: garages and providers

Decided by Magnus, deliberately: GarageAI serves two kinds of supply in one catalogue and one API.

- **Garages** run GarageAI Bridge: a machine at home or in an office, reached through Bridge's own
  WireGuard tunnel (`connection_type: mesh`). Bridge reports a heartbeat every 5 minutes, the
  tunnel's state, the runtime's models and planned absences (`garageai pause`, a laptop on
  battery). A garage that goes away is taken out of routing by the portal within minutes, and a
  buyer who names its model gets "temporarily unavailable" (503) from the gateway.
- **Providers** are datacenter or partner APIs reached over HTTPS with an API key
  (`connection_type: endpoint`): operators GarageAI does not control, who would find installing
  Bridge hard or impossible. No Bridge, no heartbeat, no tunnel state. What GarageAI knows about a
  provider comes only from requests: buyers' requests, the hourly acceptance probe, and Ops'
  checks. So the gateway has to handle a failing provider itself, at request time.

### What happens when one fails (measured 2026-10-18)

Tested with a healthy garage and a failing deployment in one model group, 20 requests each:

| Failure | Mixed group, before the change | Only the failing one, before |
|---|---|---|
| Provider rejects our key (401) | 20/20 served by the other | Buyer got 401, as if their own key were wrong |
| Provider error (500) | 19/20 | 500 after 3 s |
| Provider refuses connections | 18/20 | 500 after 2 s |
| Provider hangs | 4/6; the others waited until the client gave up | Hangs up to 600 s |
| Garage offline | taken out of routing by the garage logic | 503 "temporarily unavailable", auto suggested |

The rules since (`infra/gateway/litellm/config.yaml`, `garageai_callbacks.py`):

- **One retry, on another deployment.** A provider that answers 401/403, 404 or times out is
  cooled down at its first failure (60 s), so the retry and the next buyers go elsewhere at once.
  Other errors cool a deployment down after 2 failures.
- **A 400 never cools a deployment down:** it is the buyer's request (too long a prompt, a bad
  parameter), not a broken garage.
- **Providers get 180 s, garages 600 s** before a request is given up. Garages need the 600 s for
  100k-token prefills on home GPUs; a provider that hangs would otherwise hold every request for
  ten minutes.
- **The buyer sees 503 "temporarily unavailable"** with Retry-After 60 and a pointer to
  `garageai/auto` when a deployment rejects our key or every deployment of the model is cooling
  down, never the provider's own 401.
- **Ops raises a critical alert** when a provider rejects GarageAI's key (autoversio went 13 hours
  without one, 2026-10-17/18).

### `garageai/auto` across both kinds

`garageai/auto` holds entries from garages and providers alike. A failing entry is skipped by the
retry and then by the cooldown, so auto keeps answering as long as one entry is up. The portal's
order for auto entries is not yet in the gateway (the portal shows order 1, 2, 3; the gateway
picks by weight): until the portal sets `order` on each auto deployment, auto spreads requests
over its entries instead of using a first choice and fallbacks.

### Big prompts: how long a deployment needs before the first token

A window says what fits, not how long reading it takes. On 2026-10-18 a 160k-token Hermes prompt
failed over from the DeepSeek provider to dgx-city, which fits 524k tokens but reads ~1,800
tokens/s when cold: minutes before the first token, and the agent gave up and retried for 26
minutes. Two gateway rules (garageai_callbacks.py, rules 6 and 7) apply to every model with more
than one deployment, `garageai/auto` included:

- **Time to start, measured.** Each deployment's prefill speed comes from its own streamed requests
  (new tokens / time to first token, median of the latest 20; seeded from three days of spend logs
  after a restart). A deployment that would need more than 120 s for a request's new tokens is
  skipped while another can start in time; one that would need more than 540 s (the stream timeout
  is 600 s) is never used, and if none can, the buyer gets a 503 saying so at once. Nothing is a
  fixed cap: a faster garage takes bigger prompts by itself. Unmeasured providers are not limited;
  an unmeasured garage counts as 800 tokens/s.
- **A conversation stays where its cache is.** Runtimes keep the prefix they just processed, so the
  next turn on the same deployment reads only the new tokens (dgx-city: 188k-token turns in 1-11 s
  when warm). The gateway remembers the deployment per conversation (the buyer's key and the first
  system and user message) for 10 minutes and sends the next turn there while it is healthy and
  the new tokens fit; this wins over `order`. A failure forgets it, so the retry goes elsewhere.

## Known gateway quirks, and where they are handled

- **Ollama's response ids collide (fixed 2026-10-13 with a database trigger).** LiteLLM keys a
  spend-log row by the provider's response id, Ollama answers `chatcmpl-<0..999>`, and LiteLLM
  skips duplicate rows silently: 46 % of requests to Ollama garages left no spend-log row (payouts,
  Ops Live and the quality figures read those rows; budgets read the daily tables, which were
  complete). `infra/gateway/litellm/sql/spendlogs_unique_request_id.sql` rewrites the id on
  insert. Found by a pool load test (60 requests, 35 rows). Reported upstream: [BerriAI/litellm#46403](https://github.com/BerriAI/litellm/issues/46403) (key spend logs on `litellm_call_id`) and [ollama/ollama#18996](https://github.com/ollama/ollama/issues/18996) (unique ids); remove the trigger when both ship and the gateway runs those versions.

## Backups

`garageai-backup.timer` (03:30 UTC nightly, `infra/gateway/backup/`) takes what cannot be
recreated: LiteLLM's Postgres (keys, deployments, spend), LiteLLM's `.env` (master and salt key),
the tunnel hub's key, `/etc/garageai`, the ops `.env`, NetBird's config and volumes, the
certificates. Encrypted with the passphrase in `/etc/garageai/backup.env` and copied to the S3
bucket in `BACKUP_S3_REMOTE` once that is set (decision 2026-10-13: Hetzner snapshots of the VPS
until there are more customers, then the bucket); 7 days kept locally. The Operations Center warns
when the newest is older than 26 h or a configured off-site copy failed. **The passphrase lives in Magnus's password manager
too**: without it the off-site copies are noise.

Restore on a new server, in short: install Docker and the gateway from this repo
(`infra/gateway/`), `gpg -d garageai-<stamp>.tgz.gpg | tar -xz`, put `litellm.env` and
`garageai-hub.key` and `etc-garageai.tgz` back where they came from, start Postgres and
`pg_restore -U litellm -d litellm litellm.pgdump`, start the rest. Garages reconnect by themselves:
the hub key and the gateway's names are the same. Point the DNS names at the new address.

## Later: a control plane

Today `syncModels`, a portal function, translates the portal's desired state into the gateway's
configuration. That is enough for one gateway and a handful of garages. When the platform grows,
move this translation into its own **control plane**:
- **when:** more than one gateway or region, or hundreds of garages;
- **what it does:** reads the desired state from the portal, and reconciles every gateway against
  it continuously (like a Kubernetes controller). It also owns the acceptance tests and health
  decisions now spread over portal functions;
- **what stays in the portal:** the people's decisions and the data. The portal stops talking to
  LiteLLM directly.

Part of the move: LiteLLM's admin API (the master key's `/key/*`, `/model/*`, `/ui`) leaves the
public name. Today it is reachable on `llm.garageai.eu` (Traefik passes it only with the exact
master-key header), because the portal's functions call it there. In the control plane it is reached
only from inside (the control plane, the Operations Center), and `llm.garageai.eu` becomes a plain
alias of `api.garageai.eu` for old clients. Noted 2026-10-12.

Until then, keep the boundaries above, so that the move is a relocation and not a rewrite.

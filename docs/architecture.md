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

## Backups

`garageai-backup.timer` (03:30 UTC nightly, `infra/gateway/backup/`) takes what cannot be
recreated: LiteLLM's Postgres (keys, deployments, spend), LiteLLM's `.env` (master and salt key),
the tunnel hub's key, `/etc/garageai`, the ops `.env`, NetBird's config and volumes, the
certificates. Encrypted with the passphrase in `/etc/garageai/backup.env` and copied to the S3
bucket in `BACKUP_S3_REMOTE`; 7 days kept locally. The Operations Center warns when the newest is
older than 26 h or never left the server. **The passphrase lives in Magnus's password manager
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

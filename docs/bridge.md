# GarageAI Bridge — vision and roadmap

**Bridge** is the bridge between an operator's GPU and the GarageAI network. It is one binary,
`garageai`, on the operator's machine.

A bridge goes both ways:
- **Outward:** the garage sells inference to the network. This is what it does today.
- **Inward:** later, the operator also uses the network from the garage: their own GPU first, the
  network when it is not enough, shared with whom they choose.

**North star:** someone who has never run a model, and only has good hardware they want to earn
from, opens Bridge. Bridge sees the machine, says what it can earn and with which model, installs
and configures everything, and the garage is live. No terminal, no flags, no guessing.

Status: Bridge is the default way to connect a garage since 2026-10-10 (bridge-v0.2.0). The connect
script stays as an internal backup and reference. Phase 0 is done and phase 1 is in progress. Everything else is the agreed direction
(2026-10-10), not a commitment to dates.

## Principles

- **One contract.**
  - The garage profile (schema 1), the onboarding reports and the heartbeat payload (with
    `contexts`) are the interface to the portal and the Operations Center.
  - Bridge and `garageai-connect.sh` speak it identically: `scripts/tests/parity-doctor.sh`
    proves it.
  - The script keeps working until Bridge has parity, so the portal and ops never notice the
    switch.
- **Deterministic discovery, judgment on top.** Bridge itself sees the machine: ports, runtimes,
  models, GPU, firewall. The agent (phase 3) explains and chooses fixes on top of that profile.
  Detection does not depend on a model.
- **Read by default, act after a yes.** Diagnosing never changes anything. Restarting a runtime,
  changing a firewall or loading a model asks first, unless the operator has set it to automatic.
- **Codes always, text only with consent.** What Bridge does is product data: the problem codes
  doctor finds and whether they are gone by the next heartbeat, which knowledge-base sections are
  looked up for which code, searches that found nothing (counted per code, not their words), and
  which tools are called. That shows which fixes work and where the knowledge base has gaps, and
  it feeds time to live. The operator's own words (questions to Bridge's agent) only with their
  consent, asked once and changeable, redacted on the machine (tokens, keys, paths with user names,
  e-mail) and kept 90 days. The operator's conversation with their own agent over MCP is never
  ours: we see our tool calls only.
- **Never prompts or keys.** Bridge sees metadata about the machine and the runtime, never buyers'
  prompts or replies. Keys and whole command lines are never reported.
- **One binary, stdlib first.**
  - Linux, macOS and Windows (amd64 and arm64) from one Go codebase.
  - Installed with `curl … | sh` (no signing needed for that path) and on Windows with a
    PowerShell one-liner.
  - Signing comes later, Windows first.

## Phases

### 0. Profile — done (PR #67, #68, #71)
- `garageai doctor` and `garageai doctor --json` (Go).
- `garageai-connect.sh --doctor --json` (bash) gives the same profile, and parity is tested on
  Linux and macOS.
- The profile travels with the onboarding reports. It shows in the portal admin, the operator's
  wizard and the Operations Center's Onboarding page.

### 1. Connect and keep alive — in progress (PR: bridge-connect)
Done: `garageai connect` (the six steps, same options, payloads and onboarding reports as the
script, root through sudo with secrets on stdin), `garageai run` as a systemd timer or launchd
daemon with `contexts` in every heartbeat, `garageai uninstall`, migration of the script's old
heartbeat, self-update (v0.5.0: the portal's stable version, SHA256SUMS, run-before-swap, on by
default, `--no-auto-update` to opt out). Not yet: Windows service, pause, the firewall rule.
- `garageai connect`: join the mesh, find the runtime, register. Same onboarding reports, now
  with `version`.
- `garageai run` as a service (systemd, launchd, Windows service):
  - heartbeat with `contexts`, with backoff,
  - self-healing: restart a crashed runtime, notice VRAM out of memory,
  - the firewall rule.
- **Planned stop:**
  - `garageai pause`, and automatically when another process (a game, a render) takes the GPU.
  - The network then shows *paused*, not *down*: no critical alert and no reminders. This closes
    the gap seen when autoversio North stopped on 2026-10-10.
- `garageai self-update`, so garages do not get stuck on an old heartbeat. dgx-city needed
  "New command" to get `contexts`.
- A release flow: goreleaser in GitHub Actions, SHA256SUMS, a small readable install script like
  norrivaagent's.
- To explore: embed the NetBird client (Go, BSD-3) so connecting needs no separate NetBird install.

### 1.5. Bridge tunnel — no NetBird on the garage, no sudo (decided 2026-10-10, spike first)

Today a garage needs the NetBird client: a separate install, a system service with a login that
can be lost (a macOS upgrade logged garage-m2 out), a tun device that needs root, and a runtime
that listens on all addresses so the mesh can reach it. The model we want is cloudflared's: one
program, an outbound tunnel, a local forward.

- **Bridge terminates the tunnel itself:** WireGuard in user space (wireguard-go with its
  netstack) inside Bridge. It connects outward to the gateway, accepts the gateway's requests on
  the garage's mesh address inside its own stack, and forwards them to `127.0.0.1:<port>`.
- **What that removes:** the NetBird install and login, the tun device and root, `OLLAMA_HOST`
  and any exposure of the runtime to the home network, the firewall rule, and the difference
  between Linux, macOS and Windows (netstack is pure Go). The heartbeat can run as a user
  service. Onboarding becomes: run one command, no password.
- **Hub, not mesh:** traffic only ever goes gateway ↔ garage, so the gateway runs one WireGuard
  end with a public UDP port and every garage connects outward through its NAT. No STUN, signal
  or relay. The portal hands out each garage's key pair and mesh address with the connect
  command, and a blocked garage is a removed peer. A TCP fallback over 443 comes later for
  networks that block UDP.
- **What replaces NetBird's pieces:** peer status and last seen come from the WireGuard
  handshake and Bridge's heartbeat; access rules are trivial (a garage sees only the gateway);
  the dashboard is the Operations Center.
- **Migration without a cut-over:** the gateway reaches old garages through NetBird and new ones
  through the hub at the same time; a garage moves with one "New command". NetBird's management,
  signal, relay and dashboard are switched off when the last garage has moved.
- **Two ways were weighed.** Embedding NetBird's own client engine (netstack mode plus a
  forwarder) keeps NetBird's server and dashboard but binds Bridge to internal APIs that change
  often and a 40–60 MB binary. An own WireGuard hub is less code to depend on, less to run, and
  fully ours. The hub was chosen as the end state.
- **Spike (2–3 days) before building it:** wireguard-go netstack in Bridge against a WireGuard
  end on the gateway, forward to a local runtime, measure throughput and reconnect behaviour,
  and connect one real garage (the Mac) that way. The spike decides; if it holds, this is the
  next big step and phase 2 follows on top of it.

### 2. Run the runtime — the right engine, configured right

**Decided 2026-10-16 (Magnus): Bridge bundles Xyntetik Runner as the engine for operators who
have none.** `garageai runtime install` (offered by connect when no runtime answers, after a
question) fetches Runner from its releases, verifies SHA256SUMS, installs it root-only like Bridge,
runs it as a system service (`--serve --port 8080 --parallel 1 -c 8192`) and owns its lifecycle:
updates, restarts, model changes through the wish list, `runner --fit` as the fit check before a
pull. Ollama, LM Studio, vLLM and the rest stay supported; Runner is the path for whoever has
nothing. Why, although Runner measured ~45 % of Ollama's decode on Apple silicon on
2026-10-16 (M1: 8 vs 20 tok/s; M2: 12 vs 27; prefill ~95 tok/s on M2; prefix cache and
cached_tokens excellent, 2.3 s load, 0.35 s warm first token): the engine gets faster when its
author sees real installations and our nightly measurements per machine, and a one-file engine
made in Sweden is what "install Bridge, done" needs. Asks to Joakim: a pull-to-cache command,
`/v1/models` listing cached models, unique response ids, Apple-silicon speed. Ops gets "engine"
as a dimension in the acceptance figures so Runner vs Ollama is a curve, not a number. First
release: v0.7.0.

- **Recommend per hardware** (the engine-diversity principle, done by Bridge):
  - Ollama on a Mac, vLLM on NVIDIA with enough VRAM, llama.cpp for small cards.
  - Each with the right flags from the start: context window, prefix cache,
    `--enable-prompt-tokens-details` (so cached tokens bill at the cache price), bind to the mesh.
- `garageai runtime install|start|stop`, and `garageai models pull|load|unload`.
- **Models chosen in the portal, fetched by Bridge (next, decided 2026-10-13).** The heartbeat
  answer carries the garage's wanted models; Bridge asks Ollama to pull them (`/api/pull`, no sudo,
  no terminal), reports progress in the heartbeat ("downloading, 2.1 of 4.7 GB"), and the model is
  normalized and priced as today once it is there. Removal the same way. Ollama first (every Mac
  garage runs it); LM Studio and vLLM later. Three callers of the same mechanism:
  1. **The operator** picks from "models that fit your machine" on My garages (GPU and memory from
     the heartbeat, market price, "N garages have it already"), and clicks Add. No model names to
     know at onboarding.
  2. **GarageAI suggests**: "we would like qwen3-4b on garage-m4" is a proposal the operator
     accepts with one click. Their disk and bandwidth, so never without a yes.
  3. **GarageAI Smart (on from the start, can be turned off):** "GarageAI manages my models." GarageAI picks models that
     are in demand and fit the machine, fetches them, puts them in the pool at the market price,
     and tells the operator why ("180 requests/day, two garages were not enough"). This is what
     makes GarageAI a marketplace that can steer supply towards demand, which OpenRouter's
     providers cannot do for it.
     - **Macro, not micro.** Decisions come from days of demand (requests per model, waits and
       503s when a pool is full, pool depth against traffic), not from the last hour, and a
       garage's set of models changes at most once a day and only when it matters: every pull
       costs the operator bandwidth and wears the disk, and a model that keeps being swapped
       serves nobody. A model goes out only after 14 days without traffic.
     - **Guard rails set once by the operator:** how much disk GarageAI may use (e.g. 60 GB);
       models the operator added themselves are never touched. On from the start, in plain words: "we
       download on your bandwidth and disk, you get traffic".
     - **A set that fits in memory at once.** A garage offering 14 models that cannot all be
       loaded (Erik's Mac mini, 2026-10-13) reloads on every switch, and buyers wait for the
       load. Smart chooses a set that fits together (1 to 3 on a Mac) and takes the rest out of
       the pool (still on disk, still a dedicated address). `OLLAMA_KEEP_ALIVE` of hours, so a
       loaded model stays loaded between requests.
     - **Prefer warm garages (gateway side, not Bridge):** the heartbeat already reports which
       models are loaded; the portal gives a pool deployment whose model is cold a lower weight,
       so requests go to a garage that has the model in memory and a cold one gets traffic when
       the warm ones are full or down. No lock on a garage: a lock takes its other models out
       of the pool and only moves the reload to the next window.
     - First test: garage-m4 (16 GB, gemma4 12B only today) gets qwen3-4b, and that pool has
       three garages.
  Build order: Bridge first (wanted list in the heartbeat answer, pull with progress, report);
  the contract is the same whoever fills the list. Then the portal: the Add button, then the
  Smart choice with the demand rule as a nightly function.
  **Smart's first day downloads nothing (Magnus, 2026-10-13).** The garage goes live with the
  model it has, and earns from the first minute if that model has demand. Downloads come later,
  from days of demand, at the macro pace: never a question without an answer at onboarding, never
  a 5 GB download as the first impression. First get going, then optimise and curate.
  **Smart is on from the start: like a Mac, it works at once with sensible choices, and a pro
  can change everything.** So the page after the first heartbeat is not a question but a notice,
  with the Smart switch and the disk limit visible for whoever wants them off.
  **Bridge side built 2026-10-13 (v0.6.0), the contract:** the heartbeat answer carries
  `wanted_models: [{"model": "qwen3:4b", "action": "pull" | "remove", "by": "operator" | "garageai"}]`
  in the runtime's naming. Bridge reports `agent.models: [{model, status: downloading | ready |
  failed | removed | refused, by, completed_mb, total_mb, error, at}]` with every heartbeat and
  keeps the state in `/etc/garageai/models.json`. One pull per heartbeat run (a pull blocks the
  run; a heartbeat is re-sent every 2 minutes meanwhile), a failed pull is retried after an hour,
  10 GB are kept free, a model the operator added is never removed (`refused`), Ollama only.
  Doctor: `model_pull_failed`. `garageai models` lists the state.
- **Onboarding: three moments (decided 2026-10-13).** We cannot give good advice before we have
  seen the machine, so the teaching happens after the first heartbeat, not at signup.
  1. **Signup (30 seconds):** one sentence on what a garage is ("your computer answers AI
     requests from others when it is idle, and you are paid per token") and one action: connect
     a computer. A chat key is created silently here, so the chat just works later.
  2. **The first heartbeat (the big moment):** now we know the machine (16 GB, Ollama, gemma4
     12B). A page appears once and stays under My garages as "What should my garage offer?":
     *your machine* in two lines ("16 GB: models up to ~14B at good speed"); *recommended for
     you*, 2 to 3 models chosen on hardware and demand, with "N garages have it, X requests a
     day"; *how the pool works* in one sentence; and the notice that Smart is on: what it will
     do and why, with the switch and the disk limit beside it for whoever wants to take over.
     This is where the setting lives, because this is where it means something. The page is
     the wizard, without steps: whoever comes back sees the current recommendation for the
     current machine.
  3. **The chat:** not teaching, proof. "Your gemma4 answered in 1.2 s." House default
     preselected, key already there. Nobody is sent there from onboarding.
  Erik's 14 models are not wrong, they are not a product: with Smart on, "3 of your 14 fit at
  once and 2 have demand: we offer qwen3-coder-30b and gemma4-26b in the pool, the rest stay on
  disk". Smart means we do the work and say why; that is what separates GarageAI from a list.
  Smart is a success factor, not a feature among others (Magnus, 2026-10-13).
- **Measured profile:**
  - Bridge benchmarks prefill and decode, the practical maximum prompt, and the time to first
    token at typical agent prompt sizes, then reports them.
  - This is phase 1 of the "GarageAI för agenter" design doc. The gateway can then route and warn
    on facts, for example long prompts to a slow garage.
- **Compliance check:**
  - the runtime does not log prompts (operator terms), the firewall is on, versions are current.
  - The portal can show a *verified* badge.
- **Location:** confirm the garage is in the EU, backing the sovereignty promise.

### Security backlog (from the 2026-10-13 review; fixed items in docs/garage-security.md)
- **Signed releases:** sign SHA256SUMS (minisign/ed25519) with a key kept off CI, embed the
  public key in Bridge, verify before writing the new binary. Until then a GitHub account or a
  compromised Actions workflow can ship a release that the portal's `bridge_stable` then rolls
  out. Two-person rule for releases.
- **Tunnel as its own user:** `garageai` system user, `User=`/`UserName`, `NoNewPrivileges`,
  `ProtectSystem=strict`, `ProtectHome`, no capabilities. Nothing in the tunnel needs root.
- ~~Inference-only forward~~ done in v0.6.2 (`cli/internal/tunnel/forward.go`): an HTTP reverse
  proxy with a path allowlist and the Host rewritten to the runtime's own address.
- **Windows:** `bridge.json` and `tunnel.key` need an explicit DACL (Administrators + SYSTEM);
  do not write them before the service exists.
- **install.sh:** when `/usr/local/bin` is user-writable (Intel Homebrew), still `sudo install`,
  and have the root half warn when it runs from a user-writable file.
- **Managed models:** check free space on the filesystem Ollama's model directory is on (not `/`),
  and cap the total with the operator's disk limit from the portal.

### 3. Help — the agent in Bridge
- `garageai login` (the norrivaagent pattern):
  - the browser page hands over the operator's session and model access;
  - "the agent is you", so RLS applies to it too;
  - inference goes through the GarageAI gateway, possibly on the operator's own garage.
- `garageai ask "why is my garage not live?"`:
  - it reads the profile, runtime logs and onboarding reports, explains, and proposes the fix
    from a runbook per runtime;
  - it applies the fix after a yes, then verifies.
- `garageai mcp`: the same tools for the operator's own agent (Claude Code, opencode, …).
  **First (2026-10-11):** read-only tools (doctor, explain_problem, search_knowledge,
  garage_status) and the knowledge base. Bridge's own agent (`garageai ask`/`chat`) comes next as
  one more client of the same tools.
- The knowledge base is `docs/garage-troubleshooting.md`, one section per doctor problem code,
  built into the binary: it matches that version's doctor and works when the garage reaches
  nothing. No vector database: a few hundred answers fit in one prompt, and doctor's codes make the
  lookup exact. When operators' free-text questions run into thousands, pgvector in the Supabase we
  already have, and a `kb.json` Bridge can fetch to update answers between releases.
- Fleet knowledge: profiles from every onboarding show which setups exist out there, and feed the
  runbooks and the defaults.
- Name: the agent is a function of Bridge. A separate name, for example "Mechanic", is kept for
  later in case there is a reason to split.

### 4. Earn smart
- **Machine telemetry and the garage's own request timings (Bridge v0.6.3, 2026-10-14).** The
  heartbeat carries `agent.machine` (`load_pct`, `mem_total_mb`, `mem_free_mb`, `disk_free_gb`,
  `power` = ac | battery | "", `battery_pct`, `user_active`, and on NVIDIA `gpu_util_pct`,
  `vram_used_mb`, `vram_total_mb`) and `agent.requests` (the last 5 minutes as the forward saw
  them: `requests`, `inference`, `errors` (5xx), `streamed`, `cold_starts`, `ttfb_p50_ms`,
  `ttfb_p95_ms`, `total_p50_ms`). Doctor prints a "Now" line. Ops shows both in "From the
  garage" and treats a tunnel that drops on a laptop on battery as *away* (info), not down.
  Next: the portal and the pool use them (prefer idle and warm garages, skip a laptop on battery,
  Smart's fit check), and the gateway's timings next to the garage's separate network from model.
- **From the vLLM Panel sync (2026-10-14, Joakim's session on dgx-city), in order:**
  1. *Unified memory:* when nvidia-smi reports N/A (DGX Spark / GB10), treat the GPU as unified
     and let `machine.mem_free_mb` (MemAvailable) be the headroom; add `gpu_temp_c` and
     `gpu_power_w` (work on GB10, feed the energy idea).
  2. *Lifecycle states:* `garageai pause --reason "redeploy" --for 15m` / `garageai resume` →
     `agent.state = {kind: maintenance|draining, reason, until}`; the portal drains the garage
     instead of failing requests, the owner page says "redeploying since 21:11", Ops does not
     page. A derived `starting` state when the port is closed or /health is 503 while the runtime
     process exists (a 2-node vLLM start takes 6–13 min). Two sources of state: `garageai
     pause/resume` (humans and tools with sudo; writes under /etc/garageai) and a **state file**
     a runtime manager writes without sudo, agreed with vLLM Panel 1.28.0 (2026-10-14):
     `~<user>/.local/state/vllm-panel/state.json`, schema 1, `{panel, updated, state {kind:
     serving|starting|maintenance|stopped|failed, reason, since, until}, model {served_name,
     max_model_len}, nodes [{name, role, ok, container, gpu_temp_c, gpu_power_w,
     mem_available_mb}]}`, rewritten at least every 60 s. Bridge reads it as root with care
     (regular file only, ≤ 64 KB, `updated` within 180 s, else ignored and said in doctor) and
     honours only states that reduce traffic (maintenance → maintenance, starting → starting,
     stopped → away; failed → a doctor problem with the panel's reason). The tunnel service polls
     the file every 10 s and runs the heartbeat at once on a change; the portal takes the garage
     out of routing on that heartbeat (admin-pause path), in-flight requests finish (the drain).
     **Measured 2026-10-15 (dgx-city, vLLM Panel 1.28.2, Bridge v0.6.5, a real MiMo redeploy):**
     maintenance written 11:30:01 → Bridge saw it 11:30:03 → heartbeat 11:30:11 → both
     deployments gone 11:30:14 (13 s, 20 s before the container stopped); starting 11:31:02;
     serving 11:42:24 → heartbeat 11:42:32 → deployments back 11:43:06 (42 s, of which ~30 s an
     acceptance test the portal will now skip for a model that was live before). No alert, no
     request affected. The first test the night before ran on the slow path (the tunnel service
     still had the old binary, fixed by v0.6.6): out after 84 s via the health check, back after
     3 min via the 5-minute heartbeat.
  3. *Engine facts* as a compact `agent.engine` from the runtime's /metrics when it has one:
     prefix-cache hit rate, measured prefill tok/s (the practical max prompt), inter-token
     latency, preemptions, finished-reason counts (length/abort/error), max_num_seqs and KV
     capacity in tokens. Owner-vs-garage traffic = engine totals minus what the forward saw.
  4. *Later:* multi-node (read `nodes[]` from a panel endpoint on 127.0.0.1 when present) and
     energy (tokens per kWh from power.draw; the spot-price rule in phase 4).
- **A laptop is a garage while it is plugged in (backlog 2026-10-13).** garage-m2 was away
  53 minutes when Magnus unplugged it and took it along; Ops paged "tunnel down" within three
  minutes, as it does every time a laptop leaves. Bridge knows the power state (`pmset -g batt`
  on macOS, `/sys/class/power_supply` on Linux) and reports it in the heartbeat; the portal and
  Ops treat a laptop on battery as *away*, not *down*: no alert, the pool skips it, and the
  garage page says "away since 10:10 (on battery)". When it is plugged in again it is a garage
  again. The operator sees it as the deal it is: a garage while charging.
- **Schedule:** sell when the machine is idle (nights, workdays), not while the operator games or
  works.
- **Electricity price:** sell only when the spot price is below what the tokens pay. This is a
  Nordic/European angle: the garage sells when power is cheap.
- `garageai stats`: tokens served, earnings, uptime, GPU temperature, locally and live.
- Payout and earnings in the same view as in the portal.

### 5. Both directions — use and share
- **A local endpoint for the operator:**
  - `http://localhost:…/v1`, the own GPU first, the network when it is busy or lacks the model;
  - one key, and the operator's own agents just work.
- **Private garage:**
  - share with family, colleagues or a team by invitation, with no public sale;
  - this is the "Private garage" backlog item, and Bridge is its local half.
- **Several machines:** one Bridge for a home lab or a small cluster (several GPUs, several nodes,
  like the DGX cluster).

### North star: the control panel for a beginner
- **The same binary serves a local web UI** (`garageai ui` → `http://localhost:…`, embedded in the
  Go binary). No Electron, nothing else to install. A tray icon or desktop app can come later.
- **The first visit:**
  1. Bridge detects the hardware.
  2. It says "your RTX 4090 can run *model X* at about N tok/s; at today's prices that is about
     €Y per month when idle at night".
  3. Start earning: one click installs the runtime and the model with the right settings,
     connects to the mesh, and runs the acceptance test. The garage is live.
- **After that:**
  - a dashboard with earnings, load, temperature and the schedule;
  - pause and resume;
  - model suggestions;
  - the agent in a chat box when something is wrong.
- Everything an expert does in the terminal, a beginner does with the panel. Both are the same
  Bridge.

## Where Bridge meets the rest of GarageAI

| Bridge | Portal | Gateway / Operations Center |
|---|---|---|
| profile, onboarding reports | onboarding_events (+ profile), wizard progress, admin | Onboarding page, stuck alerts |
| heartbeat with contexts | garage_models.context_length → LiteLLM limits | Models page checks against the runtime |
| pause / planned stop | garage status *paused* | info, not critical; no reminders |
| Bridge tunnel (hub) | key pair + mesh address per garage in the connect command | WireGuard end on the gateway; peer status from handshakes |
| measured profile | routing facts, "verified" badge | quality per garage |
| local endpoint, private garage | keys, sharing, billing | routing to private pools |

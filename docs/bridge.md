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
heartbeat. Not yet: Windows service, pause, self-update, the firewall rule, embedded NetBird.
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
- **Recommend per hardware** (the engine-diversity principle, done by Bridge):
  - Ollama on a Mac, vLLM on NVIDIA with enough VRAM, llama.cpp for small cards.
  - Each with the right flags from the start: context window, prefix cache,
    `--enable-prompt-tokens-details` (so cached tokens bill at the cache price), bind to the mesh.
- `garageai runtime install|start|stop`, and `garageai models pull|load|unload`.
- **Demand-driven model suggestions:** "the network needs qwen3.8-27b, you have VRAM for it, load
  it?". Supply follows what buyers actually call.
- **Measured profile:**
  - Bridge benchmarks prefill and decode, the practical maximum prompt, and the time to first
    token at typical agent prompt sizes, then reports them.
  - This is phase 1 of the "GarageAI för agenter" design doc. The gateway can then route and warn
    on facts, for example long prompts to a slow garage.
- **Compliance check:**
  - the runtime does not log prompts (operator terms), the firewall is on, versions are current.
  - The portal can show a *verified* badge.
- **Location:** confirm the garage is in the EU, backing the sovereignty promise.

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
- Fleet knowledge: profiles from every onboarding show which setups exist out there, and feed the
  runbooks and the defaults.
- Name: the agent is a function of Bridge. A separate name, for example "Mechanic", is kept for
  later in case there is a reason to split.

### 4. Earn smart
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

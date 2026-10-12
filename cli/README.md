# GarageAI Bridge (`garageai`)

**Bridge** is the bridge between an operator's GPU and the GarageAI network, in both directions:
the garage sells to the network, and later the operator uses the network from the garage. It is one
binary, `garageai`, and since 2026-10-10 the default way to connect a garage: the portal's
command uses it. `scripts/garageai-connect.sh` stays as an internal backup and reference, kept in
step with Bridge by `scripts/tests/parity-doctor.sh`. The vision and the phases are in
[docs/bridge.md](../docs/bridge.md).

```sh
curl -fsSL https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.sh | sh
garageai connect
```

`garageai connect` without options links the machine in the browser: Bridge shows a code and opens
app.garageai.eu/bridge/link, the operator checks that it is their machine, chooses the garage and
approves, and Bridge picks up the garage's credentials itself. No command with a token is copied or
shown. Bridge finds the runtime (one with models, preferably reachable over the mesh) unless
`--runtime`/`--port` say otherwise, and suggests the garage name from the computer's name. Every
report, registration and heartbeat carries a machine id (the OS's id, hashed), so the portal warns
when a garage suddenly reports from another machine.

## Commands

```sh
garageai connect [options]   # join the mesh, check the runtime, register, install the heartbeat
garageai doctor [--json]     # what runs on this machine, what is wrong, and how to fix it
garageai run [--once]        # the heartbeat (the installed service runs this every 5 minutes)
garageai uninstall           # remove the heartbeat and Bridge's configuration
garageai version
```

`connect` takes the same options and `GARAGEAI_*` environment variables as
`scripts/garageai-connect.sh`, and sends the same payloads to the portal (`register-node`,
`node-heartbeat`, `onboarding-report` with the garage profile). So the portal's command only swaps
`bash garageai-connect.sh` for `garageai connect`. It needs root for joining the mesh and installing
the heartbeat: run as a normal user it re-runs itself through `sudo` and passes the secrets on
stdin, never on the command line.

What `connect` does, like the script:
1. NetBird installed (installs the CLI client if not, after asking).
2. Join the mesh with the setup key, or keep the current connection. Without a setup key and not
   on the mesh, it says to get a new command from the portal.
3. The runtime answers an OpenAI-compatible API (with the API key when it needs one).
4. Reachable over the mesh: the runtime listens on the network, streamed replies carry usage, and
   for Ollama the loaded context window (reported as it is; the gateway enforces it).
5. Register; the portal runs the acceptance test.
6. Install the heartbeat as a service (systemd timer or launchd daemon) running `garageai run --once`
   from its own root-only copy of Bridge (`/usr/local/lib/garageai/garageai` on Linux,
   `/Library/Application Support/GarageAI/garageai` on macOS): the service runs as root, so it never
   runs the copy in the operator's home, which their account could change.
   The connect script's old heartbeat is removed, so there is one reporter with one token.

The heartbeat also carries Bridge's own report (`agent`): its version, the OS and what `doctor` finds
right now (severity, code and message; never keys). The portal keeps the latest one, and the
Operations Center shows it per garage, so a garage is seen from the inside, not only from the mesh.

Every step is reported to the portal (`onboarding-report`), with the garage profile at the start,
on a failure and when done, so a stuck connect is visible in the wizard, the admin and the
Operations Center.

**Updating:** run the install command again. When the machine has the heartbeat service, the
installer also runs `sudo garageai install-service`, which puts the new version in the service's
root-only place, so the heartbeat runs it from the next report on. `doctor` reports
`service_binary_unsafe` (and the Operations Center shows it) when a root service runs a binary a
normal user can change, as Bridge before v0.2.4 could set up.

What each problem means and how to fix it: [docs/garage-troubleshooting.md](../docs/garage-troubleshooting.md).

## The Bridge tunnel: `garageai connect --tunnel bridge`

Bridge can carry the garage's tunnel to the gateway itself ([docs/bridge-tunnel.md](../docs/bridge-tunnel.md)):
WireGuard in user space inside Bridge, no NetBird, no network interface, no routes, and the runtime
stays on `127.0.0.1`. Bridge makes a key pair on the machine and sends only the public key to the
portal (`tunnel-assign`), gets its tunnel address (10.66.x.y) and the hub's key back, dials out to
the hub on UDP 51830 (any NAT works), and forwards the tunnel to the runtime locally. The tunnel runs
as its own service (`garageai tunnel`: `garageai-tunnel.service` or `eu.garageai.tunnel`) and
reconnects by itself after sleep or a network change. `garageai connect` without `--tunnel` keeps
using NetBird; a garage can switch either way by running connect again.

## Your own agent: `garageai mcp`

Bridge's tools as an MCP server over stdio, for the agent you already use (Claude Code, Claude
Desktop, ChatGPT, opencode, ...). The tools only read; the agent tells you what to run.

| Tool | What it gives the agent |
|---|---|
| `doctor` | The garage profile: machine, GPU, NetBird, heartbeat, runtimes and models, problems with fixes |
| `explain_problem` | The knowledge base's section for a problem code |
| `search_knowledge` | The best sections for a question in plain words |
| `garage_status` | How the garage is connected, which binary the heartbeat runs, its latest log lines (no secrets) |

The knowledge base is [docs/garage-troubleshooting.md](../docs/garage-troubleshooting.md), built
into the binary (`cli/internal/kb`; a test keeps the copy equal to the doc and checks that every
problem code doctor can report has a section).

```sh
claude mcp add garageai -- garageai mcp          # Claude Code
```

Claude Desktop and others: a server with `"command": "garageai", "args": ["mcp"]` (the full path,
e.g. `/Users/you/.local/bin/garageai`, when the app does not see your shell's PATH).

## Implementation notes

- Standard library only. One codebase for Linux, macOS and Windows (amd64 and arm64). The
  heartbeat service is not implemented on Windows yet.
- **Linux:** reads `/proc` directly, no `ss` or `lsof` needed. **macOS:** `lsof`. **Windows:**
  `netstat` and `tasklist`.
- Mesh state comes from the NetBird service when it answers, else from the tunnel interface
  (a normal user may not be allowed to ask the service on macOS).
- Never reports keys or whole command lines.

## Build and test

```sh
cd cli
go test ./...
go build -o garageai .
../scripts/tests/parity-doctor.sh "$PWD/garageai"   # same profile as the script?
```

CI ("Bridge (Go)") tests on Linux, macOS and Windows and runs the parity test on Linux and macOS.
A tag `bridge-vX.Y.Z` publishes the six binaries plus SHA256SUMS as a pre-release; `install.sh`
picks the newest `bridge-v*` release and verifies the checksum.

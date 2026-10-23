# Audit GarageAI yourself (or let your AI do it)

You are about to let a program from us run on your computer and let strangers' AI requests reach
it. You should not have to take our word for what it does. Everything that runs on a garage is in
this repository, and this file says what we claim and where in the code each claim can be checked.

**For an AI agent** (Claude Code, Codex, Cursor, ChatGPT with the repository, ...): read this file
first, then verify every claim below against the code it points to. Report each claim as *holds*,
*does not hold* or *could not verify*, with the file and line. Then report anything else you would
want the owner of the computer to know before installing GarageAI Bridge. Be critical. We would
rather hear it from you than from an incident.

**For a person:** paste this into your own AI agent:

> Review https://github.com/magnusfroste/garageai as a security reviewer for someone about to
> install GarageAI Bridge on their own computer. Start with AUDIT.md: check every claim against the
> code it points to, say which hold and which do not, and tell me what else worries you.

## What runs on a garage

| Part | Where | Runs as |
|---|---|---|
| Installer | [`cli/install.sh`](cli/install.sh), [`cli/install.ps1`](cli/install.ps1) | the user, asks for sudo/Administrator once |
| Bridge (`garageai`) | [`cli/`](cli) (Go, no cgo) | root / SYSTEM for the two services; the user for `doctor` and the menu-bar icon |
| Tunnel service | [`cli/internal/tunnel/`](cli/internal/tunnel), [`cli/tunnel_cmd.go`](cli/tunnel_cmd.go) | root / SYSTEM |
| Heartbeat service | [`cli/run.go`](cli/run.go), [`cli/internal/portal/`](cli/internal/portal) | root / SYSTEM, every 5 minutes |
| Menu-bar icon (optional) | [`cli/internal/tray/`](cli/internal/tray), [`cli/tray_cmd.go`](cli/tray_cmd.go) | the logged-in user |
| Xyntetik Runner (optional, only with `garageai runtime install`) | a third-party engine, [`cli/runtime_cmd.go`](cli/runtime_cmd.go) installs a pinned release | a service |

## Claims to verify

1. **Nothing on the computer accepts connections.** Bridge's WireGuard tunnel is in user space
   (wireguard-go with a gVisor netstack): no network interface, no change to routing, and no TCP
   listener on the host. Its listeners exist only inside the tunnel's own netstack. The tunnel
   holds one UDP socket on a random port for its outgoing connection to the gateway; WireGuard
   answers nothing on it except a valid handshake from the gateway's key, so to anyone else it
   looks closed.
   → `cli/internal/tunnel/tunnel.go` (`conn.NewDefaultBind`, `ListenTCPAddrPort` on the netstack).
2. **The tunnel has exactly one peer: the gateway's key, allowed one address.** Nobody else can
   complete a handshake, and garages cannot reach each other.
   → `cli/internal/tunnel/tunnel.go` (the `public_key=` / `allowed_ip=<hub>/32` peer configuration).
3. **Only inference paths reach the runtime.** Requests through the tunnel are checked against a
   fixed list (chat, completions, embeddings, models, Anthropic messages, responses, Ollama's
   inference and read-only endpoints, metrics, health). Everything else, such as Ollama's pull and
   delete, gets 403 and never reaches the runtime.
   → `cli/internal/tunnel/forward.go` (`allowedPrefixes`, `Allowed`, `ForwardHandler`), with tests in `forward_test.go`.
4. **The heartbeat sends no prompts and no conversations.** It sends the garage name, runtime,
   port, model list, context windows, Bridge version, doctor's findings (short messages), request
   timings (counts and milliseconds), token counts and the requested model name per request, compute
   power (watts, for tokens per kWh) and machine health (load, memory, disk, power, battery, GPU).
   To count tokens, the tunnel looks at the first and last 8 KB of each response in memory for the
   runtime's `usage` numbers and at the request's `"model"`; it keeps only those numbers and the name,
   never content. **One key is sent:** the runtime's own API key, *if the owner configured one*
   (for example a vLLM `--api-key`), because the gateway needs it to call the runtime.
   → `cli/run.go` (`heartbeat`, `agent`, `writeStatus`), `cli/internal/portal/portal.go` (`Registration`, `Agent`), `cli/internal/machine/telemetry*.go`, `cli/internal/tunnel/stats.go`, `cli/internal/tunnel/usage.go`, `cli/internal/energy/`.
5. **Secrets are readable by root only.** The configuration with the portal token and the
   tunnel's private key are 0600 (root) on Linux and macOS. On Windows they are under
   `%ProgramData%\GarageAI`, restricted to Administrators and SYSTEM with icacls. The few files
   meant for the user (`bridge-public.json`, `status.json`, `pause.json`, request timings) hold
   no secrets.
   → `cli/internal/service/service.go`, `windows.go`, `binary_unix.go`, `binary_windows.go`, `status.go`.
6. **Self-update only installs releases of this repository, checked against their checksum.**
   The portal can say which version is stable, but not where to download it from outside this
   repository's GitHub releases. Every download is checked against the release's `SHA256SUMS`
   before it replaces the running binary.
   → `cli/update.go` (`downloadVerified`, `downloadAsset`), `cli/run.go` (`maybeSelfUpdate`). It can be turned off: `garageai connect --no-auto-update`.
7. **The installers do the same:** they download from this repository's releases and verify
   `SHA256SUMS`.
   → `cli/install.sh`, `cli/install.ps1`.
8. **Releases are built in public from a tag.** GitHub Actions builds every binary from the tagged
   source and publishes them with `SHA256SUMS`. Anyone can read the build log.
   → `.github/workflows/cli.yml` (`build` and `release` jobs).
9. **The menu-bar icon only reads.** It runs as the user, reads only the secret-free files from
   claim 5, opens no port, makes no network request, and its actions are fixed: open the portal,
   run `garageai doctor`, open at login on or off, hide.
   → `cli/internal/tray/model.go` (package comment and `Do`), `platform_*.go`, `run_*.go`.
10. **The gateway does not store prompts or answers.** It keeps per-request records (tokens,
    timing, cost, which deployment), not content.
    → `infra/gateway/litellm/config.yaml` (no prompt storage is enabled), `infra/gateway/litellm/garageai_callbacks.py`.

## Rebuild a release yourself

```sh
git clone https://github.com/magnusfroste/garageai && cd garageai
git checkout bridge-v1.2.0
cd cli
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w -X main.version=v1.2.0" -o garageai-darwin-arm64 .
shasum -a 256 garageai-darwin-arm64     # compare with the release's SHA256SUMS
```

Use the Go version shown in that release's build log. With the same version the binary should be
identical. If it is not, tell us.

## What you are trusting us with (known limits)

These are true today. They are also on the [security page](https://www.garageai.eu/security):

- **The requests.** The gateway decides what reaches your runtime. A compromised gateway could
  send your runtime arbitrary *inference* requests (claim 3 still limits the paths).
- **Our GitHub account.** Self-update trusts this repository's releases. Releases are not yet
  signed with a key kept off GitHub; that is planned. Until then, `--no-auto-update` is the way to
  update only by hand.
- **Root/Administrator.** The two services need it to install themselves. Their code is the
  `cli/` directory in this repository, nothing else.
- **Your runtime.** Ollama, LM Studio, vLLM, llama.cpp, Runner and the others are not our code.
  Bridge only forwards inference to them.
- **The app** at app.garageai.eu is a separate repository,
  [garageai-portal](https://github.com/magnusfroste/garageai-portal), and runs on Supabase.

## Found something?

Write to security@garageai.eu (see [docs/SECURITY.md](docs/SECURITY.md)). We answer within two
working days and credit you if you want. An AI review that finds a real problem counts too.

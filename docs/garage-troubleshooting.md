# Garage troubleshooting

What to do when `garageai doctor` (or the Operations Center's "From the garage" column, or the
garage page in the portal) reports a problem. One section per problem code, with what it means,
how to fix it, and what we learned getting real garages online. This is the knowledge base for
Bridge: when a fix text in `cli/internal/profile/build.go` changes, change it here too, and the
other way round.

Start with:

```sh
garageai doctor
```

It prints the machine, NetBird, the heartbeat, every runtime it finds and its models, then the
problems with a fix for each. `garageai doctor --json` is the same as JSON (the profile that
connect also reports to the portal).

Severity: **error** means buyers cannot be served, **warning** means they are served worse or will
not be after a restart, **info** is worth knowing.

## The machine

### `mac_sleeps` (info)
The Mac sleeps some minutes after its display turns off. A sleeping Mac drops off the mesh, and the
gateway sees the garage as down (on garage-m2 this was an 8-minute "tunnel down" alert).

- **MacBook (Air, Pro):** System Settings → Battery → **Options…** → turn on *Prevent automatic
  sleeping on power adapter when the display is off*. It only holds **on the power adapter**: on
  battery a MacBook sleeps anyway, so a MacBook garage is a garage while it is plugged in.
  *Wake for network access* can stay on *Only on Power Adapter* or *Always*.
- **Mac Studio, Mac mini, iMac (no battery):** System Settings → Energy → turn on *Prevent
  automatic sleeping when the display is off*. Also turn on *Start up automatically after a power
  failure*, so a power cut does not take the garage offline until someone presses the button.
- The display may still turn off: that does not matter, only sleep does.
- Check: `pmset -g` shows `sleep 0` (on a MacBook: while plugged in), and doctor no longer reports it.

### `no_gpu` (info)
No GPU found (nvidia-smi, rocm-smi, Apple silicon). Fine for a test; a CPU garage is too slow to sell.

## NetBird (the mesh)

### `netbird_missing` (error)
The NetBird client is not installed. Run `garageai connect`: it installs the CLI client (no desktop
app needed).

Learned: before Bridge v0.2.3 the heartbeat (run by launchd with a minimal PATH) reported this
falsely on Macs. The Operations Center ignores it while the gateway sees the garage on the mesh.

### `netbird_disconnected` (error)
Installed but not connected. `sudo netbird up`; if it says *NeedsLogin*, the machine's mesh
registration was removed: get a new command from the portal (My garages → New command, with a
setup key) and run it.

Learned: a setup key is used once. Re-running an old command fails with "No setup key"; the portal
makes a new one.

### `other_vpn` (info)
Tailscale (or ZeroTier) is installed. Operators worry that NetBird breaks it; it does not. GarageAI's
mesh is `100.83.0.0/16`, and NetBird only routes that range. Tailscale uses addresses in
`100.64.0.0/10` and keeps working: your tailnet, MagicDNS and exit nodes are untouched.

- **The one exception:** a device of yours whose Tailscale address starts with `100.83.` is out of
  reach from this machine while NetBird is on (the more specific route wins). Check with
  `tailscale status`. If you have one, tell us: GarageAI can move its range.
- **Exposing the runtime:** "listen on the network" (`OLLAMA_HOST=0.0.0.0`, "Expose Ollama to the
  network") also makes it reachable from your tailnet and your local network, not only from GarageAI.
  That is usually fine at home; if not, a firewall rule or binding to the NetBird address only
  (it is in `garageai doctor`) narrows it.
- Learned: the first operator with Tailscale (2026-10-11) stopped onboarding at the NetBird step
  for this reason. Bridge before v0.3.2 could also take Tailscale's address for the mesh address
  when it could not ask NetBird; it now only accepts addresses in `100.83.0.0/16`.
- Later, Bridge's own tunnel (phase 1.5 in docs/bridge.md) adds no system route or interface, and
  this question goes away.

## The Bridge tunnel

Since v0.4.2 the default (v0.4.0: opt-in with `--tunnel bridge`), a garage carries its own tunnel in Bridge (
docs/bridge-tunnel.md): no NetBird, no network interface, no sudo for the network, and the runtime
stays on `127.0.0.1`. Bridge dials out to the gateway over UDP 51830 and forwards the tunnel to the
runtime locally.

### `tunnel_missing` / `tunnel_stopped` (error)
The garage is set up for the Bridge tunnel, but the tunnel service is not installed or not running,
so the gateway cannot reach it. Restart it (`sudo systemctl restart garageai-tunnel` on Linux,
`sudo launchctl kickstart -k system/eu.garageai.tunnel` on macOS), or run
`garageai connect --tunnel bridge` again. The service log is `/var/log/garageai-tunnel.log` on
macOS and `journalctl -u garageai-tunnel` on Linux.

- **No handshake with the hub:** the network blocks UDP out (some offices and hotspots). A TCP
  fallback is planned; until then use NetBird: `garageai connect --tunnel netbird`.
- **After sleep or a network change** the tunnel reconnects by itself within a few seconds.
- **A new tunnel garage's first handshakes can go unanswered for a few seconds:** the hub learns
  about it from the portal on its next round (every 10 s). Bridge waits up to a minute and says so.
  Before v0.4.1 it gave up after 20 s ("No answer from the gateway's tunnel hub"); running the
  command again worked, because the hub knew the garage by then.
- **"Could not install the tunnel service: launchctl bootstrap ... Bootstrap failed: 5: Input/output
  error"** (macOS, connect run again over a running tunnel): launchd was still tearing the old
  service down when Bridge registered the new one. Run
  `sudo launchctl bootstrap system /Library/LaunchDaemons/eu.garageai.tunnel.plist` once the
  tunnel is down for good, or run connect again. Since v0.5.1 Bridge waits for the teardown and
  retries. Learned: garage-m1 and garage-m4 both hit it on 2026-10-12 when updating to v0.5.0 by
  running the install command.
- **The runtime on localhost is correct** for a tunnel garage: doctor does not report
  `localhost_only` for the port Bridge forwards.

## The runtime

### `no_runtime` (error)
No OpenAI-compatible runtime answers on this machine. Start it (Ollama, LM Studio, vLLM,
llama.cpp, ...) and run doctor again.

Learned: macOS's AirPlay receiver answers on port 5000 with 403 and looks like a server; doctor
ignores it.

### `localhost_only` (error)
The runtime listens on 127.0.0.1 only, so the gateway cannot reach it over the mesh.
- **Ollama app:** Settings → *Expose Ollama to the network*.
- **Ollama with Homebrew or a terminal:** start it with `OLLAMA_HOST=0.0.0.0:11434`.
- **llama.cpp:** `llama-server --host 0.0.0.0`. **vLLM / SGLang:** `--host 0.0.0.0`.
- **LM Studio:** Developer → Settings → *Serve on Local Network*.

The mesh is private (only the gateway can reach a garage), so listening on all interfaces is what
we want. Learned: Ollama's internal runner has its own localhost port while a model is loaded. It
is a process called `ollama` or `llama-server`, and its model is a file in `~/.ollama/models/blobs/`.
Doctor before v0.2.6 could report it as a llama.cpp on localhost: ignore that, it is not a problem.

### `ollama_host_not_persistent` (warning, macOS)
Ollama listens on the network now, but will not after a restart. Use the Ollama app's *Expose
Ollama to the network* setting, or start Ollama with `OLLAMA_HOST` set the way you always start it.

Learned: `launchctl setenv` does not reach Ollama run by `brew services`; set it in the service's
own environment, or use the app.

### `small_context` (warning, Ollama)
Ollama runs the model with a small context window (default 4096 tokens) and silently cuts longer
prompts. Coding agents send far longer ones. Set `OLLAMA_CONTEXT_LENGTH` (doctor says how much this
machine can take, e.g. 8192 on 8 GB) the same way as `OLLAMA_HOST`, and restart Ollama. The
heartbeat reports the new window within 5 minutes; no need to connect again. Until then the gateway
knows the real window and rejects longer prompts instead of letting them be cut.

### `port_mismatch` (error)
The garage is registered with one port, but the runtime answers on another. Start the runtime on
the registered port, or run the connect command again with `--port`.

### `needs_api_key` (info)
The runtime wants an API key (vLLM and SGLang usually run with one, since they listen on the
network by default; Ollama and LM Studio usually have none). Since v0.4.4 connect finds the key
itself where the runtime was started (its command line `--api-key`, its environment, or the Docker
container's command and environment) and otherwise asks for it with hidden input. Never paste it in
a chat or a ticket. With the Bridge tunnel the runtime does not need to listen on the network at
all: start vLLM with `--host 127.0.0.1` and the key becomes a second lock instead of the only one.

### `runtime_in_terminal` (warning)
The runtime runs in a terminal window (`ollama serve` typed into a shell, for example). It stops
when that window is closed or the machine restarts, and the garage goes offline until someone
starts it again. Run it as a service: the Ollama app keeps Ollama running (macOS), Linux has the
`ollama` systemd service, vLLM runs as a systemd unit or in Docker with `--restart unless-stopped`.
Learned: garage-m2 served buyers for two days from a terminal window nobody remembered.

### `runtime_user_session` (info, macOS)
The runtime runs as a LaunchAgent in your login session (Homebrew's `brew services`, or an agent you
made), on a Mac with several user accounts. It stops when you log out. Another user logging in with
fast user switching is fine, as long as your account stays logged in. For a shared Mac, run the
runtime as a system service (a LaunchDaemon). Learned: garage-m4 went offline when another family
member logged the operator out.

### Xyntetik Runner
Recognised since v0.6.2 (`runner` process, `runner_` counters on `/metrics`), default port 8080.
It serves localhost only by design and rejects requests whose `Host` is not loopback, so it works
with the Bridge tunnel (the default), never with the NetBird fallback. Start it with
`runner -m <model.gguf> --serve --port 8080 --parallel 2`; it reports cached prompt tokens by
itself. Under evaluation as a recommended engine (docs/inference-engines.md §8).

### `no_cached_token_report` (info, vLLM)
vLLM does not report cached prompt tokens, so buyers pay full price for prefix-cache hits. Add
`--enable-prompt-tokens-details` to the vLLM command.

## Models managed by GarageAI

Since v0.6.0 the portal can ask a garage to fetch or remove a model (My garages → Add, or
GarageAI Smart choosing in-demand models for the machine; docs/bridge.md). The heartbeat carries
the wish list, Bridge asks Ollama to pull (`/api/pull`), reports progress ("2.1 of 4.7 GB") and,
when it is there, the next heartbeat offers the model. `garageai models` shows the list and how
each went. Ollama only for now. A model the operator added themselves is never removed by Bridge
("refused" in the list); one Bridge fetched can be. Bridge keeps 10 GB free on the disk and
refuses a pull that would not.

### `model_pull_failed` (warning)
GarageAI asked this garage to fetch a model and Ollama could not (the reason is in the message:
no route to ollama.com, not enough disk, a model name that does not exist). Bridge tries again
after an hour, not every 5 minutes. Check the disk (`df -h`) and that `ollama pull <model>` works
by hand; if it does, the next retry will too. `garageai models` shows the state.

## The heartbeat

### `heartbeat_missing` / `heartbeat_stopped` (warning)
The heartbeat reports the garage's models every 5 minutes; without it, a model you load or remove
is not picked up. Run `garageai connect` again. The Operations Center also warns when a garage
answers but has sent no heartbeat for 20 minutes.

### `heartbeat_rejected` (error)
The portal rejects the heartbeat's token: it was replaced (a new command was made for this garage)
or the garage was deleted. My garages → New command, run it here.

Learned: do not paste the connect command (it contains the garage's token) into chats or tickets.
If you did, make a new command: the old token stops working.

### `service_binary_unsafe` (error)
The heartbeat runs as root, from a file a normal user account can change: anything running as that
user could get root. Run the install command again, which updates the service:

```sh
curl -fsSL https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.sh | sh
```

or `sudo garageai install-service`. The service then runs its own root-only copy
(`/Library/Application Support/GarageAI/garageai` on macOS, `/usr/local/lib/garageai/garageai` on
Linux). Learned: Bridge rc1 set up services that ran the copy in the operator's home.

## Updating Bridge

Since v0.5.0 Bridge updates itself: the portal tells the heartbeat which version garages should run
(the "stable" version, promoted by GarageAI after it has run on our own garages), and Bridge
fetches that release, checks it against SHA256SUMS, runs the new binary once, and only then swaps
the service's copy and restarts the tunnel. The heartbeat runs the new version from its next run.
Your own `garageai` command is refreshed too when it is in `/usr/local/bin` or `~/.local/bin`.
`garageai connect --no-auto-update` keeps a version until you run `garageai update` yourself.

### `update_failed` (warning)
The last self-update did not go through (the reason is in the message: no route to github.com, a
checksum mismatch, or a binary that did not run). Run `garageai update`; if it fails the same way,
check the machine's internet access to github.com. Nothing was replaced: Bridge keeps running the
version it had.

## Mix-ups we have seen

- **`garageai: command not found` right after installing.** Before 2026-10-13 the installer put
  Bridge in `~/.local/bin` whenever `/usr/local/bin` was not writable (always, on a Mac), and that
  directory is not on a Mac's PATH. Either add it
  (`echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc`) or run the install
  command again: it now installs to `/usr/local/bin` with sudo. The heartbeat and the tunnel were
  never affected; they run the root-only copy.
- **A register token left in the shell.** `export GARAGEAI_REGISTER_TOKEN=…` from an old connect
  command stays in that terminal. Bridge then skips the browser link and uses the old token. Since
  v0.4.1 it at least uses GarageAI's own portal address with it; to link afresh, open a new
  terminal or `unset GARAGEAI_REGISTER_TOKEN GARAGEAI_REGISTER_URL GARAGEAI_SETUP_KEY GARAGEAI_MANAGEMENT_URL`.
- **The terminal waits for a password behind the browser.** After approving in the browser, the
  terminal asked for the sudo password, out of sight. Since v0.4.1 Bridge asks for it first, before
  the browser opens.

- **One garage name, two machines.** A garage was first set up on one laptop, whose settings were
  then deleted, and later connected from another laptop under the same name. The token and the
  setup key belonged to the first machine, so the second had to be paired again. Give every
  machine its own garage (`garage-m2`, `garage-m3`, ...). Planned: Bridge sends a machine id, and
  the portal warns when a garage suddenly comes from another machine.
- **A garage that is "offline" though everything runs:** check that its models are not disabled
  by an admin (Admin → Models); routing needs both.

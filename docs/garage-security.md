# Security for garage owners

What happens on your machine when you offer your GPU on GarageAI, in plain words: what the
connect command (GarageAI Bridge) installs, who can reach your machine, and what they can and cannot do.

## The short version

The plain-language version for operators is at https://www.garageai.eu/security ("Is it safe to share from my computer?"). This document is the technical one.

- Your machine makes one **outgoing** encrypted connection to GarageAI. You open **no ports on
  your router**, and nothing on your machine becomes reachable from the internet.
- Exactly **one party** can connect to your machine through that tunnel: the GarageAI gateway.
  Other garages cannot reach you, and you cannot reach them.
- The gateway can reach **one thing**: the port of your inference runtime (Ollama, LM Studio,
  vLLM and so on). It cannot reach your files, your other programs or other devices at home.
- You can leave at any time with one command, and the tunnel is gone.

## How it works

```
 Buyer ──HTTPS──▶ GarageAI gateway (EU) ══ encrypted tunnel ══▶ your runtime (one port)
                                                 ▲
                             your machine opened this tunnel, outwards
```

Since Bridge v0.4.2 the tunnel is Bridge's own: WireGuard, run inside the `garageai` program
(no network interface, no VPN client, no route on your machine). Bridge dials out to the
gateway's hub on UDP 51830, and inside that tunnel it accepts connections on one port and passes
them to your runtime on `127.0.0.1`. Your runtime keeps listening on localhost only. The tunnel
is keyed: only the hub's key is accepted, so nothing but the gateway can send anything through
it. Bridge does keep a UDP socket open for the tunnel, but WireGuard answers nothing it cannot
authenticate, so from the internet it looks like nothing at all. Because your machine is the one
that connects out, it works behind a home router, behind NAT, on a laptop that changes network.

The earlier setup, a [NetBird](https://netbird.io) mesh with a network interface, is still
available as a fallback (`garageai connect --tunnel netbird`), with the same one rule: **the
gateway may connect to garages on the runtime ports**, nothing else.

## Why not port forwarding, ngrok or a personal VPN?

| | What it means for you |
|---|---|
| **Port forwarding on your router** | Your runtime is open to the whole internet. Ollama has no password, so anyone who finds the port can use your GPU. |
| **A public tunnel URL (ngrok and similar)** | Your runtime gets a public address that anyone with the link can call. Free addresses change or stop working, and then your garage is offline. |
| **A personal VPN account (Tailscale and similar)** | Works well for your own devices, but it is your private network. A marketplace needs a network where the buyer side can reach only one port on your machine and nothing else. |
| **GarageAI's mesh** | No public address exists. Only the gateway can reach the runtime port, and the connection is set up and kept alive for you. |

## What the connect command installs

The command installs **GarageAI Bridge**, one small program called `garageai`, and runs
`garageai connect`. Bridge is open source: read [`cli/`](../cli/README.md) before you run it.
The install downloads a build for your system and checks it against the release's checksum.
Bridge asks for your password once (`sudo`) because these things need administrator rights:

1. **The tunnel service** (`garageai tunnel`, a systemd service on Linux, a launchd daemon on
   macOS) that keeps the tunnel up and reconnects by itself after sleep or a network change. Its
   key is in `/etc/garageai/tunnel.key`, readable only by administrators, and never leaves the
   machine (only the public half is sent to GarageAI).
2. **A heartbeat service** (`garageai run`, every 5 minutes) that tells GarageAI which models your
   runtime serves and their context windows, and reports what `garageai doctor` finds (so support
   can see what you see). It sends your garage's name, the runtime type, the port, the model names
   and Bridge's own version. It does not send prompts, files or anything else. Its settings,
   including your garage's token, are in `/etc/garageai/bridge.json`, readable only by
   administrators (`bridge-public.json` holds the same without secrets, for doctor). The service
   runs its own copy of the program in a directory only root can write (`/Library/Application
   Support/GarageAI` on macOS, `/usr/local/lib/garageai` on Linux), never the file in your PATH,
   so nothing running as your user can change what runs as root.
3. **Progress reports.** While connecting, Bridge tells GarageAI which step it is on and, when
   a step fails, what it found on the machine: operating system, GPU, memory, which inference
   runtimes answer on which ports and addresses, firewall state. Never keys, prompts or command
   lines; file paths are reduced to file names.

Bridge does not change your router, firewall, Docker or other software. `garageai uninstall`
removes the services and Bridge's settings.

## What GarageAI can and cannot do on your machine

**Can:**

- **Send requests to your runtime's port** through the tunnel. In practice the gateway calls two
  things: the model list, and chat completions on behalf of buyers.
- **Update Bridge** (since v0.5.0). The portal names a version it has run on GarageAI's own
  garages first; Bridge fetches it from this repository's GitHub releases only (no other address
  is accepted, whatever the portal says), checks it against the release's checksums, runs it
  once, and only then replaces the service's copy. `garageai connect --no-auto-update` turns this
  off; `garageai update` does it by hand.
- **Fetch or remove models in Ollama** (since v0.6.0), when you add a model under My garages or
  let GarageAI choose. Only from Ollama's library or Hugging Face, one at a time, with 10 GB kept
  free on the disk; a model you installed yourself is never removed by Bridge. `garageai models`
  shows what was asked and what happened.

**Cannot:** log in to your machine, run commands, read your files, see your screen, reach other
ports, or reach other devices on your home network. Bridge runs nothing the portal sends: the
portal can only choose among the actions above, and each has its own limits in Bridge's code.

Two things you should know, because they follow from how runtimes work:

- **Only inference passes through the tunnel (since Bridge v0.6.2).** Bridge forwards the
  OpenAI-style inference paths (`/v1/chat/completions`, `/v1/models`, embeddings, the Anthropic
  and Responses variants), the read-only Ollama endpoints the health check uses, and `/metrics`.
  Ollama's management commands (pull, delete, create) and anything else on the port answer 403
  from Bridge itself and never reach the runtime. This is a property of your machine, not a
  promise about the gateway. (With the NetBird fallback the runtime is reached directly, so there
  the promise still applies.)
- **With the NetBird fallback the runtime listens on `0.0.0.0`**, which also opens it to your own
  home network (not to the internet). With the Bridge tunnel (the default) it stays on localhost.

## Security review

Bridge was reviewed on 2026-10-13 (an independent read of all of `cli/`, threat model: a
compromised portal, a local user on the machine, a network attacker, a buyer, and the operator's
own secrets). Fixed the same day (v0.6.1):

- The portal could name the address Bridge fetched updates from, and the checksums came from the
  same address: a compromised portal could have run anything as root on every garage. Now only
  this repository's GitHub releases are accepted.
- After an update, Bridge (as root) refreshed the `garageai` command in every user's `~/.local/bin`
  by following whatever was there, including a symlink a local user could have planted. It now
  touches only `/usr/local/bin/garageai`, and only when that file and directory are root's.
- As root, Bridge searched Homebrew's user-owned directories for tools such as `nvidia-smi`; a
  file planted there by anything running as your user would have run as root every 5 minutes.
  Root now only uses directories nobody else can write in.
- The runtime's API key was offered to any local port that answered "unauthorized"; now only to
  the configured runtime port, and it is no longer placed in the environment where child
  programs would inherit it. The NetBird setup key no longer appears on a command line.
- A link the portal sends is only opened in the browser when it is GarageAI's own https address.
- Portal addresses must be https; redirects to plain http are refused.
- Model fetches are limited to Ollama's library and Hugging Face.

Still open, in order: signing the release checksums with a key kept off GitHub (so a compromised
GitHub account cannot ship a release); running the tunnel service as its own unprivileged user;
Windows file permissions (Bridge's Windows service is not finished); an inference-only path
filter in the tunnel. The findings and their status are tracked in docs/bridge.md.

## Other people's prompts run on your machine

When you sell inference, buyers' prompts are processed by your runtime. They exist in plain
text in your machine's memory while the model computes the answer; there is no way around
that. What you do with them is a matter of rules, not technology, so the rule is strict:

> **Operators must not log, store, read, forward or analyse buyers' prompts or responses.**
> This is part of the operator terms. A garage found doing so is removed from the
> marketplace and its account is closed. The only exception is GarageAI's own acceptance
> test and hourly probe, whose prompts are generated by the platform.

GarageAI cannot see from the gateway whether a runtime logs prompts, so this relies on you.
Buyers are told that prompts are processed on the operator's machine and are advised not to
send sensitive data to community garages; verified operators and a data-processing agreement
are the path for buyers who need more than that.

### Does my runtime log prompts?

By default, none of the supported runtimes write prompts to disk. Several can be switched
to do so with a flag; do not use those flags on a garage.

| Runtime | Default | Logs prompts when |
|---|---|---|
| vLLM | request id, token counts and latency only | `--enable-log-requests` (older: `--log-requests`) or `VLLM_LOGGING_LEVEL=DEBUG` |
| SGLang | request metadata only | `--log-requests` |
| Ollama | endpoint, status and timing only | not in normal operation; `OLLAMA_DEBUG=1` adds detail |
| llama.cpp (`llama-server`) | request metadata only | `--verbose` |
| LM Studio | nothing on disk; the app's developer log shows request content while open | the developer log is on screen |
| Unsloth, MLX, Lemonade | as the vLLM / llama.cpp server they wrap | same flags as the underlying server |

If you need request logging to debug your own setup, pause the garage in the portal first
and turn it off again before resuming.

## Checking and leaving

```bash
garageai doctor      # tunnel, runtime, heartbeat: what works and what to fix
garageai uninstall   # remove the heartbeat and Bridge's settings (then: sudo netbird down)
```

Stopping your runtime or shutting down the machine also takes your garage off the market
within minutes. Nothing breaks; it comes back when the runtime is up again.

## Reporting a problem

See [SECURITY.md](SECURITY.md) for how to report a vulnerability.

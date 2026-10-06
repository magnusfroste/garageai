# Security for garage owners

What happens on your machine when you offer your GPU on GarageAI, in plain words: what the
connect script installs, who can reach your machine, and what they can and cannot do.

## The short version

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

Your machine joins a private WireGuard network using [NetBird](https://netbird.io), an open
source project from Berlin. GarageAI runs its own NetBird server in the EU, so no third-party
tunnel service sits in the path. Because your machine is the one that connects out, it works
behind a home router, behind NAT and with an IP address that changes.

The network has one rule: **the gateway may connect to garages on the runtime ports**. There
is no rule that lets a garage connect to another garage or to the gateway, so those connections
are dropped.

## Why not port forwarding, ngrok or a personal VPN?

| | What it means for you |
|---|---|
| **Port forwarding on your router** | Your runtime is open to the whole internet. Ollama has no password, so anyone who finds the port can use your GPU. |
| **A public tunnel URL (ngrok and similar)** | Your runtime gets a public address that anyone with the link can call. Free addresses change or stop working, and then your garage is offline. |
| **A personal VPN account (Tailscale and similar)** | Works well for your own devices, but it is your private network. A marketplace needs a network where the buyer side can reach only one port on your machine and nothing else. |
| **GarageAI's mesh** | No public address exists. Only the gateway can reach the runtime port, and the connection is set up and kept alive for you. |

## What the connect script installs

The script is open source: read [`scripts/garageai-connect.sh`](../scripts/garageai-connect.sh)
(or `garageai-connect.ps1` on Windows) before you run it. It needs administrator rights for
two things:

1. **The NetBird client**, which creates the tunnel and a network interface for it.
2. **A heartbeat**, a small scheduled job that tells GarageAI every 5 minutes which models your
   runtime serves. It sends your garage's name, the runtime type, the port and the model names.
   It does not send prompts, files or anything else. Its settings, including your garage's
   token, are stored in a file only administrators can read.

It does not change your router, Docker or other software. The NetBird client does two things
on the machine itself: it adds firewall rules for its own interface (`wt0`) so that only the
gateway can reach the runtime port, and it may register a DNS resolver for the mesh's
`.netbird.selfhosted` names with systemd-resolved. Both are removed by `--uninstall`.

## What GarageAI can and cannot do on your machine

**Can:** send requests to your runtime's port through the tunnel. In practice GarageAI calls
two things: the model list, and chat completions on behalf of buyers.

**Cannot:** log in to your machine (remote shell over the tunnel is turned off), read your
files, see your screen, reach other ports, or reach other devices on your home network.

Two things you should know, because they follow from how runtimes work:

- **The gateway can reach everything your runtime serves on that port.** For Ollama, that port
  also carries its model management commands (pull, delete). GarageAI does not call them; the
  portal and the gateway code are open source so this can be checked. If you want a hard
  guarantee, run a runtime that only serves inference on that port, such as `llama-server` or
  vLLM.
- **Listening on `0.0.0.0` also opens the runtime to your own home network** (not to the
  internet). On a shared or office network, bind the runtime to your mesh address instead, or
  restrict the port with your firewall. `--doctor` shows your mesh address. The Windows script
  adds a firewall rule that only allows the mesh.

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
bash garageai-connect.sh --doctor      # tunnel, runtime, heartbeat: what works and what to fix
bash garageai-connect.sh --uninstall   # remove the heartbeat and leave the mesh
```

Stopping your runtime or shutting down the machine also takes your garage off the market
within minutes. Nothing breaks; it comes back when the runtime is up again.

## Reporting a problem

See [SECURITY.md](SECURITY.md) for how to report a vulnerability.

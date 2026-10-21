## The short answer

Yes, with three things you should know. Your computer opens no ports and accepts no connections.
Bridge opens one outgoing, encrypted tunnel to the GarageAI gateway, and the only thing that comes
back through it is inference requests for the model you offer. Nothing else on your computer or
your home network can be reached. What you should know: you are trusting GarageAI with the
requests, Bridge updates itself from our releases, and your AI runtime is a program we did not
write. Each is explained below, with what we do about it.

## What runs on your computer

GarageAI Bridge is one small open-source program, `garageai`. `garageai connect` installs two
services:

- **The tunnel.** A WireGuard tunnel that your computer opens to the gateway at
  api.garageai.eu. Bridge carries it itself: no VPN app, no network interface, no change to
  your routing. The tunnel accepts exactly one peer, the gateway's key. Anyone else on the
  internet cannot even start a handshake.
- **The heartbeat.** Every 5 minutes it reports to the portal which models your runtime
  serves, their context windows, and how your machine is doing (load, memory, disk, battery).
  Never prompts, never conversations, never command lines. The one key it sends is your
  runtime's own API key, and only if you set one (for example vLLM's `--api-key`): the gateway
  needs it to call your runtime.

Both run as root (Administrator on Windows), because installing services needs it. Their
configuration and the tunnel key are readable by root only. `garageai uninstall` removes
everything.

## What can reach your computer through the tunnel

One thing: your inference runtime, on the port you chose, on localhost. And not all of it:

- Bridge forwards the tunnel to the runtime through an HTTP proxy with a fixed list of allowed
  paths: chat and completions, embeddings, the model list, health. The runtime's management
  calls (download, delete, create a model) are refused with 403 before they reach it. We tested
  this: a download request sent through the tunnel is rejected.
- The proxy points at one address, 127.0.0.1 and your runtime's port. The tunnel routes no
  traffic: your files, your other programs, other ports and other devices on your network are
  not reachable from it.
- Other garages cannot reach you at all, and buyers never see which garage served them or where
  it is.
- Your runtime can keep listening on localhost only, closed to your home network.

## What GarageAI sees

- Your account details and your garage's settings.
- The heartbeat: models, context windows, machine health, Bridge version, problems found by
  `garageai doctor`.
- Usage: how many requests your garage served, tokens, and what you earned.
- The public address your tunnel connects from (needed to keep it up), like any server you
  connect to.

GarageAI does not store prompts or answers. They pass through the gateway in memory and are
processed on your machine. The privacy policy lists everything we keep.

## What you should know

**1. You are trusting GarageAI with the requests.** The gateway decides what is sent to your
runtime. That is the point of a marketplace, and it also means that whoever controls the gateway
can use your GPU: send it requests, keep it busy. Not read your files, not reach your network,
but use the capacity you offered. We run the gateway on our own server in the EU, with the
configuration public in the repository. An outside watchdog checks the gateway, the portal and
this website from GitHub every 5 minutes, and signed releases are next, to make a takeover harder
and visible.

**2. Bridge updates itself from our releases.** A new version is downloaded from our GitHub
releases, checked against its checksum, run once as a test, and only then swapped in. New
versions go to our own canary garages for 24 hours before they become the stable version other
garages fetch. The honest remaining risk: if our GitHub account were taken over, an attacker
could publish a release. Signed releases, with a key that does not live on GitHub, are on our
list. Until then you can turn self-update off (`garageai connect --no-auto-update`) and update by
hand with `garageai update`.

**3. Your runtime is a program we did not write.** Ollama, LM Studio, llama.cpp and the others
are third-party programs. Bridge limits what reaches them to inference requests, but a
vulnerability in how a runtime handles a request is outside Bridge. Keep your runtime updated;
`garageai doctor` tells you what runs.

## Check it yourself

You do not have to take our word for it:

- `garageai doctor` shows what Bridge found: the tunnel, the runtime, what listens where, and
  every problem with a fix.
- Bridge is open source under the MIT licence; the forward with its allowed paths is in
  `cli/internal/tunnel/forward.go`.
- `sudo lsof -i -P | grep garageai` (macOS, Linux) or `netstat -ano` (Windows) shows that Bridge
  has no TCP listener and one UDP socket, for its tunnel to the gateway. WireGuard answers nothing
  on it except the gateway's key.

### Let your own AI check it

Bridge's code is public, so you can ask your own AI agent (Claude Code, Codex, Cursor, ChatGPT with
the repository) to review it before you install. Our [AUDIT.md](https://github.com/magnusfroste/garageai/blob/main/AUDIT.md)
lists every claim on this page with the file where it can be checked, and how to rebuild a release
and compare its checksum. Paste this into your agent:

> Review https://github.com/magnusfroste/garageai as a security reviewer for someone about to
> install GarageAI Bridge on their own computer. Start with AUDIT.md: check every claim against the
> code it points to, say which hold and which do not, and tell me what else worries you.

If it finds something, tell us (below).
- `garageai pause` takes your garage off the market while you need the machine;
  `garageai uninstall` removes Bridge and its services.

## The rules you agree to

Other people's prompts run on your machine. The operator terms say you may not log, store, read,
forward or analyse them, and the supported runtimes do not write prompts to disk unless you turn
logging on. A garage found doing so is removed from the marketplace.

## Reporting a problem

Found something? Write to security@garageai.eu. We answer within two working days and credit
you if you want.

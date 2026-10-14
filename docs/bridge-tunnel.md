# Bridge tunnel (phase 1.5): design

Bridge carries the garage's tunnel to the gateway itself, in user space. There is no NetBird
client, no network interface on the garage, and no system routes. The runtime stays on
`127.0.0.1`. This page is the design. Read it before the gateway hub (step C) is built.

Status 2026-10-12, evening: **the default for new garages** (Bridge v0.4.2). garage-m1, garage-m2,
garage-m4 and an outside operator's Mac mini (with Tailscale, Ollama on localhost) run on it, and so
does dgx-city. NetBird stays as the fallback (`--tunnel netbird`) until the TCP fallback exists.

Status 2026-10-12, morning:
- **The spike holds over the internet** (`spikes/bridge-tunnel/README.md`). It was 16 ms vs
  NetBird's 28 ms, streamed the same, worked through carrier NAT, and came back ≤ 19 s after a Wi-Fi
  switch.
- **Why now:** onboarding showed the cost of NetBird. One operator stopped, worried about Tailscale.
  We had the address mix-up with Tailscale. NetBird is an extra app and a sudo step, and it was slow
  to recover after network changes.

## Shape

```
garage (Bridge, user space)                       gateway
┌─────────────────────────────┐   UDP 51830    ┌──────────────────────────────────────┐
│ runtime 127.0.0.1:11434     │◀── wireguard ──│ wg-hub 10.66.0.1/16 (kernel)          │
│ Bridge: netstack 10.66.x.y  │   (garage dials│   ▲  LiteLLM api_base                  │
│   listens :11434 in tunnel, │    out; NAT ok)│   │  http://10.66.x.y:11434/v1         │
│   forwards to 127.0.0.1     │                │ garageai-hub: peers from the portal    │
└─────────────────────────────┘                └──────────────────────────────────────┘
```

- **Addresses:** `10.66.0.0/16`. The hub is `10.66.0.1`, and each garage gets one `/32`, assigned
  by the portal and stable for the garage's lifetime. 10.66/16 overlaps neither NetBird's
  100.83/16 nor Tailscale's 100.64/10.
- **The garage dials out to the hub's public UDP port 51830**, and its keepalive (25 s) holds the
  NAT mapping. The hub has no endpoint for a garage: it learns it from each handshake, so any NAT,
  any network change and any new address work without configuration.
- **On the garage**, Bridge listens on its tunnel address, inside its own netstack, on the runtime's
  port, and forwards to `127.0.0.1:<port>`. LiteLLM's `api_base` keeps the mesh shape
  (`http://<tunnel ip>:<port>/v1`). Only the host part differs from NetBird.
- **On the gateway, the hub is a kernel WireGuard interface (`wg-hub`)**, not user space. LiteLLM's
  container reaches `10.66.0.0/16` through the host the same way it reaches NetBird's `wt0` today,
  and the kernel is faster and survives our own process restarts. The protocol is the same, so a
  user-space garage talks to a kernel hub unchanged.

## Keys: no secret leaves the machine

1. Bridge creates its key pair on the garage (`~/.config/garageai/tunnel.key` for the operator, or
   `/etc/garageai/tunnel.key` (0600) for the service).
2. With the browser link (`bridge-link-start`) or `register-node`, Bridge sends its **public** key.
3. The portal stores `tunnel_public_key` and assigns `tunnel_ip`, and returns
   `{tunnel_ip, hub_public_key, hub_endpoint: "api.garageai.eu:51830"}`.
4. If the key changes (a reinstall), the portal replaces the old one. The hub then drops the old
   peer and adds the new one. One garage has one key at a time.

The hub's private key lives only on the gateway (`/etc/garageai/hub.key`, 0600).

## The hub service (gateway, `garageai-hub`)

- Every 30 s it reads the garages that use the tunnel from `gateway-targets` (new fields
  `tunnel_public_key` and `tunnel_ip`), and reconciles `wg-hub` with `wg set` and `wg syncconf`.
  It adds new peers, removes deleted or disabled garages, and changes keys. There are no
  endpoints, and `allowed-ips` is the garage's `/32`.
- It reports each peer's latest handshake and transfer to the Operations Center: the Mesh page
  and the "tunnel down" signal. This is the same model as today's NetBird peer status.
- If the portal cannot be reached, it keeps the last peer set. It never empties the hub because of
  a failed read.

## Security

**The forward is an HTTP proxy, not a TCP copy (v0.6.2).** Bridge accepts connections on its
tunnel address and proxies them to the runtime on 127.0.0.1 with a path allowlist (inference
paths, `/v1/models`, the read-only Ollama endpoints the health check uses, `/metrics`); anything
else, such as Ollama's pull/delete, gets a 403 from Bridge. The Host header is rewritten to the
runtime's own address, which runtimes that accept loopback authorities only (Xyntetik Runner)
require. Streaming passes through unbuffered; the wait for the first byte allows a long prefill.
`cli/internal/tunnel/forward.go`.


- **Only the gateway may open connections:** a garage cannot reach the gateway or another
  garage. nftables on `wg-hub`:
  - **inbound** from 10.66/16 to the host: only replies to connections the gateway opened
    (`ct state established,related`), everything else dropped;
  - **forward** between garages: dropped.

  This is the same rule as NetBird's unidirectional "gateway-to-garages" policy.
- **A garage only sees** the gateway's requests to its runtime port. On the garage, Bridge
  forwards one port to `127.0.0.1` and nothing else. The garage's own network is never reachable
  through the tunnel.
- **UDP 51830** is open in the Hetzner firewall. The guard can rate-limit handshakes per source
  later if needed (WireGuard already ignores unauthenticated packets cheaply).
- **Key theft on a garage** lets the thief be that garage, and nothing more. The fix is the same as
  for a stolen register token: a new command or link replaces the key.

## Bridge (garage end)

- **`cli/internal/tunnel`:** wireguard-go and the gVisor netstack. This ends "stdlib only" for
  Bridge. The tunnel cannot be built without them, and the binary grows from about 6 to about
  13 MB.
- **`garageai tunnel`** runs it in the foreground for tests. As a service it runs as its own
  long-running daemon next to the heartbeat (`eu.garageai.tunnel` / `garageai-tunnel.service`,
  KeepAlive), because the heartbeat is a 5-minute one-shot.
- **Network changes:** Bridge watches its addresses. On a change it rebinds its UDP socket and
  sends a handshake at once, instead of waiting for the next keepalive. The aim is about 1 s after
  the new network is up, against ≤ 25 s in the spike.
- **`garageai connect --tunnel bridge`** opts in: the mesh steps use the tunnel instead of NetBird.
  NetBird stays the default until the tunnel has run on three garages.
- **doctor** reports the tunnel: up, latest handshake, hub reachable over UDP. A new problem,
  `tunnel_udp_blocked`, comes with the TCP fallback below.

## Portal (Lovable)

- `garages`: `tunnel_public_key`, `tunnel_ip` (unique, assigned from 10.66.0.2 upwards) and
  `transport` (`netbird` or `tunnel`).
- `bridge-link-start` and `register-node` accept `tunnel_public_key`, and the answers carry
  `tunnel_ip`, `hub_public_key` and `hub_endpoint`.
- `gateway-targets` gains `transport`, `tunnel_ip` and `tunnel_public_key`. For a tunnel garage,
  `host` is its `tunnel_ip`, so routing (`syncModels`), health and the Ops Center work unchanged.

## Later

- **TCP fallback** for networks that block UDP: WireGuard over a TCP/443 stream (udp-over-tcp),
  terminated by the hub service.
- **More hubs:** one per region. This belongs to the control plane (docs/architecture.md).

## Steps

| Step | What | Touches production |
|---|---|---|
| A | This design | no |
| B | `cli/internal/tunnel` + `garageai tunnel`, tested against the spike hub | no |
| C | `wg-hub` + `garageai-hub` + nftables on the gateway | yes: maintenance window, Magnus's OK |
| D | Portal: keys, tunnel IPs, gateway-targets | yes (Lovable) |
| E | Pilot: garage-m4 from NetBird to the tunnel; then three garages; then the default for new garages | yes |

**Rejecting the tunnel** is possible at any step without harm. NetBird keeps running for every
garage until E, and a tunnel garage can go back to NetBird with `garageai connect`.

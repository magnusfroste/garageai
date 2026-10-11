# Spike: Bridge terminates the tunnel itself

Question (docs/bridge.md, phase 1.5): can GarageAI Bridge replace the NetBird client on a garage
the way cloudflared works for Easypanel? One program, an outbound tunnel, a forward to the
runtime on localhost. No tun device, no root, no NetBird install.

## What was built

`main.go`, about 250 lines, with two ends. Both run WireGuard in user space: wireguard-go and its
gVisor netstack.

- **garage**
  - connects outward to the hub and keeps the NAT mapping alive (keepalive 25 s);
  - accepts TCP on its tunnel address *inside its own stack*;
  - forwards each connection to the runtime on `127.0.0.1`.
- **hub** (the gateway's end)
  - listens on UDP and has no endpoint configured for the garage: it learns it from the
    garage's handshake, whatever NAT the garage is behind;
  - calls the runtime through the tunnel.

`fake_runtime.py` stands in for a runtime bound to `127.0.0.1` only.

## Results (2026-10-10, gateway VPS, 2 vCPU, both ends on the same machine over UDP)

| Check | Result |
|---|---|
| Runs as a normal user (uid 1000), no sudo | yes |
| Runtime listens on localhost only, still reachable through the tunnel | yes |
| First `/v1/models` after start | 5.3 s (the hub waits for the garage's first handshake; in operation the tunnel is always up) |
| `/v1/models` once connected | 2–14 ms |
| Streamed reply, time to first byte | 42 ms |
| Throughput, both ends in user space | 335–371 Mbit/s |
| Garage restarted with a new UDP source port mid-test | next request answered in 6 ms; the hub learned the new endpoint from the handshake |
| Garage end idle | 11 MB RSS, 0 % CPU |
| Builds for Linux, macOS (arm64) and Windows | yes, 13.6 MB |

**Verdict: it holds.** Throughput is far beyond what token streams need: a long reply is a few
kbit/s, and a 100k-token prompt is under 1 MB. Reconnects need no configuration, and the
garage needs no privileges.

## Not tested yet (the next steps, need the operator's OK)

1. **A real garage over the internet.** This needs a UDP port open on the gateway in the Hetzner
   firewall, and the garage end run on the MacBook (garage-m2) behind a home NAT.
2. **The gateway end as a kernel WireGuard interface** (`wg-hub`, 10.66.0.0/16), so LiteLLM's
   container can route to garages the way it reaches the NetBird mesh today. The protocol is the
   same, so a user-space garage talks to a kernel hub unchanged. This is a network change on the
   gateway, so it is done in a maintenance step.
3. **The portal's side:** a key pair and a tunnel address per garage, delivered with the connect
   command; removing a garage removes its peer.
4. **Networks that block UDP:** a TCP fallback over 443.

If these hold, the garage half moves into `cli/` as Bridge's tunnel, and the NetBird client is
no longer needed on new garages.

## Run it

```sh
go build -o spike .
eval "$(./spike keys | sed 's/^/G_/')"; eval "$(./spike keys | sed 's/^/H_/')"
python3 fake_runtime.py 18001 &
./spike garage -key $G_private -hub-key $H_public -hub 127.0.0.1:51830 -forward 8001=127.0.0.1:18001 &
./spike hub -key $H_private -garage-key $G_public -listen 51830 -port 8001 -bulk-mb 300
```

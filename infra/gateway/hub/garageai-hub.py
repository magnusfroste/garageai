#!/usr/bin/env python3
"""GarageAI Bridge tunnel hub (docs/bridge-tunnel.md): the gateway's end of the garages' tunnels.

`wg-hub` is a kernel WireGuard interface, 10.66.0.1/16, UDP 51830. Each garage that uses the Bridge
tunnel is one peer: its public key and its /32. There are no endpoints: the hub learns each garage's
address from its handshakes, whatever NAT it is behind. LiteLLM reaches garages at
http://<tunnel ip>:<port>/v1 through the host, as it reaches NetBird's wt0.

Every 10 s this service:
- reads the garages with transport "tunnel" from the portal (gateway-targets);
- adds, updates and removes peers so that wg-hub matches. If the portal cannot be read, it keeps
  the peers it has and never empties the hub because of a failed read;
- writes /var/lib/garageai-hub/status.json (latest handshake, endpoint, bytes per garage) for the
  health check and the Operations Center.

Only the gateway may open connections: nftables (table inet garageai_tunnel) drops anything a
garage starts towards the gateway or another garage, and lets replies through.

Config: /etc/garageai/gateway-health.env (portal URL and gateway key, as the health service) and
/etc/wireguard/garageai-hub.key (root, 0600; AppArmor confines wg to /etc/wireguard). Run as root by garageai-hub.service.
"""
import json
import os
import subprocess
import sys
import time
import urllib.request

IFACE = "wg-hub"
HUB_IP = "10.66.0.1/16"
PORT = 51830
KEY = "/etc/wireguard/garageai-hub.key"   # AppArmor lets wg read keys only in /etc/wireguard
ENV = "/etc/garageai/gateway-health.env"
STATUS = "/var/lib/garageai-hub/status.json"
EXTRA = "/etc/garageai/hub-peers.json"   # optional static peers {pubkey: {"ip":..,"garage":..}}, for tests
EVERY = 10   # a new tunnel garage waits for this before its first handshake is accepted
NFT = f"""
table inet garageai_tunnel {{
  chain input {{
    type filter hook input priority 0; policy accept;
    iifname "{IFACE}" ct state established,related accept
    iifname "{IFACE}" counter drop
  }}
  chain forward {{
    type filter hook forward priority 0; policy accept;
    iifname "{IFACE}" ct state established,related accept
    iifname "{IFACE}" counter drop
  }}
}}
"""


def sh(*args, input=None, check=True):
    r = subprocess.run(args, input=input, capture_output=True, text=True)
    if check and r.returncode != 0:
        raise RuntimeError(f"{' '.join(args)}: {r.stderr.strip() or r.stdout.strip()}")
    return r.stdout


def log(msg):
    print(f"hub: {msg}", flush=True)


def setup():
    """Idempotent: the interface, its key, port and address, and the firewall table."""
    if subprocess.run(["ip", "link", "show", IFACE], capture_output=True).returncode != 0:
        sh("ip", "link", "add", IFACE, "type", "wireguard")
        log(f"created {IFACE}")
    sh("wg", "set", IFACE, "listen-port", str(PORT), "private-key", KEY)
    if HUB_IP.split("/")[0] not in sh("ip", "-4", "addr", "show", IFACE):
        sh("ip", "addr", "add", HUB_IP, "dev", IFACE)
    sh("ip", "link", "set", IFACE, "up")
    subprocess.run(["nft", "delete", "table", "inet", "garageai_tunnel"], capture_output=True)
    sh("nft", "-f", "-", input=NFT)


def load_env():
    env = {}
    for line in open(ENV):
        line = line.strip()
        if "=" in line and not line.startswith("#"):
            k, v = line.split("=", 1)
            env[k] = v.strip().strip("\"'")
    return env


def desired(env):
    """{public key: {"ip", "garage"}} for tunnel garages, or None when the portal cannot be read."""
    url = env.get("GARAGEAI_PORTAL_URL", "").rstrip("/") + "/functions/v1/gateway-targets"
    req = urllib.request.Request(url, headers={"x-gateway-key": env.get("GARAGEAI_GATEWAY_KEY", ""), "User-Agent": "garageai-hub/1"})
    try:
        with urllib.request.urlopen(req, timeout=15) as r:
            data = json.load(r)
    except Exception as e:  # noqa: BLE001
        log(f"portal not readable ({e}); keeping the current peers")
        return None
    targets = data if isinstance(data, list) else data.get("targets", [])
    want = {}
    for t in targets:
        k, ip = t.get("tunnel_public_key"), t.get("tunnel_ip")
        if t.get("transport") == "tunnel" and k and ip:
            want[k] = {"ip": ip, "garage": t.get("garage")}
    if os.path.exists(EXTRA):
        try:
            want.update(json.load(open(EXTRA)))
        except ValueError:
            log(f"{EXTRA} is not valid JSON; ignored")
    return want


def current():
    """{public key: allowed ip} on wg-hub."""
    out = {}
    for line in sh("wg", "show", IFACE, "allowed-ips").splitlines():
        parts = line.split()
        if parts:
            out[parts[0]] = parts[1].split("/")[0] if len(parts) > 1 and parts[1] != "(none)" else None
    return out


def reconcile(want):
    have = current()
    for k, v in want.items():
        if have.get(k) != v["ip"]:
            sh("wg", "set", IFACE, "peer", k, "allowed-ips", f"{v['ip']}/32")
            log(f"peer {v.get('garage') or k[:8]} -> {v['ip']}")
    for k in set(have) - set(want):
        sh("wg", "set", IFACE, "peer", k, "remove")
        log(f"removed peer {k[:8]}")


def write_status(names):
    peers = {}
    lines = sh("wg", "show", IFACE, "dump").splitlines()[1:]   # the first line is the interface
    for line in lines:
        f = line.split("\t")
        if len(f) < 8:
            continue
        key, endpoint, allowed, hs, rx, tx = f[0], f[2], f[3], int(f[4]), int(f[5]), int(f[6])
        peers[key] = {"garage": (names.get(key) or {}).get("garage"), "ip": allowed.split("/")[0],
                      "endpoint": None if endpoint == "(none)" else endpoint,
                      "latest_handshake": hs or None, "rx_bytes": rx, "tx_bytes": tx}
    os.makedirs(os.path.dirname(STATUS), exist_ok=True)
    tmp = STATUS + ".tmp"
    with open(tmp, "w") as f:
        json.dump({"at": int(time.time()), "interface": IFACE, "port": PORT, "peers": peers}, f)
    os.chmod(tmp, 0o644)
    os.replace(tmp, STATUS)


def main():
    setup()
    once = "--once" in sys.argv
    names = {}
    env = load_env()
    while True:
        try:
            want = desired(env)
            if want is not None:
                reconcile(want)
                names = want
            write_status(names)
        except Exception as e:  # noqa: BLE001 - keep serving the peers we have
            log(f"error: {e}")
        if once:
            return
        time.sleep(EVERY)


if __name__ == "__main__":
    main()

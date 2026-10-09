#!/usr/bin/env python3
"""garageai-guard: block clients that guess API keys or probe the gateway.

Runs every minute on the gateway (garageai-guard.timer). It reads the last few minutes of
Traefik's access log for llm.garageai.eu and blocks a source IP for 24 h (7 days when it
comes back within a week) when, within 2 minutes, it gets:
  - 30 or more 401 answers (invalid API key), or
  - 60 or more 403 answers (paths the gateway does not serve: LiteLLM admin and pass-through routes).
A real buyer with a wrong key sees a few 401s, not 30 in two minutes.

Blocking is an nftables set with timeouts (table inet garageai_guard), dropped before Docker's
NAT so it covers every container behind Traefik. Bans survive a reboot: the state file is
replayed into the set on every run. Never blocked: private, loopback and mesh addresses, the
gateway itself, and GUARD_ALLOW (comma-separated IPs) in /etc/garageai/guard.env.

State and status: /var/lib/garageai-guard/ (status.json is shown in the Operations Center).
  garageai-guard                     run once (what the timer does)
  garageai-guard --replay FROM TO    show who would have been blocked in that window (ISO times), block nothing
  garageai-guard --unban IP          lift a block
"""
import ipaddress
import json
import os
import re
import subprocess
import sys
import time
from collections import defaultdict
from datetime import datetime, timezone

DIR = "/var/lib/garageai-guard"
STATE = f"{DIR}/state.json"
STATUS = f"{DIR}/status.json"
ENV = "/etc/garageai/guard.env"
TRAEFIK = "netbird-traefik"
ROUTERS = ("litellm@docker", "litellm-admin@docker", "litellm-deny@docker")
WINDOW = 120                      # seconds
RULES = {401: (30, "invalid API keys"), 403: (60, "blocked paths")}
BAN, REPEAT_BAN = 24 * 3600, 7 * 24 * 3600
TABLE = "garageai_guard"
LINE = re.compile(r'^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) (\S+)[^"]*" (\d{3}) .*"([^"]+@docker)" ')


def sh(cmd, check=False):
    r = subprocess.run(cmd, capture_output=True, text=True)
    if check and r.returncode:
        raise RuntimeError(f"{' '.join(cmd[:3])}: {r.stderr.strip()}")
    return r


def protected():
    nets = [ipaddress.ip_network(n) for n in ("10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8",
                                              "100.64.0.0/10", "::1/128", "fc00::/7", "fe80::/10")]
    own = sh(["hostname", "-I"]).stdout.split()
    allow = []
    try:
        for line in open(ENV):
            if line.startswith("GUARD_ALLOW="):
                allow = [x.strip() for x in line.split("=", 1)[1].strip().strip("'\"").split(",") if x.strip()]
    except OSError:
        pass
    singles = {ipaddress.ip_address(x) for x in own + allow if x}
    return lambda ip: ip in singles or any(ip in n for n in nets)


def offenders(since, until=None):
    """{ip: {status: [timestamps]}} from Traefik's access log."""
    cmd = ["docker", "logs", TRAEFIK, "--since", since] + (["--until", until] if until else [])
    r = subprocess.run(cmd, capture_output=True, text=True)
    hits = defaultdict(lambda: defaultdict(list))
    for line in (r.stdout + r.stderr).splitlines():
        m = LINE.match(line)
        if not m or m.group(6) not in ROUTERS:
            continue
        status = int(m.group(5))
        if status not in RULES:
            continue
        ts = datetime.strptime(m.group(2), "%d/%b/%Y:%H:%M:%S %z").timestamp()
        hits[m.group(1)][status].append(ts)
    return hits


def verdicts(hits, is_protected):
    """[(ip, count, reason)] for clients over a limit within any WINDOW."""
    out = []
    for ip, by_status in hits.items():
        try:
            if is_protected(ipaddress.ip_address(ip)):
                continue
        except ValueError:
            continue
        for status, times in by_status.items():
            limit, reason = RULES[status]
            times.sort()
            j, best = 0, 0
            for i, t in enumerate(times):
                while t - times[j] > WINDOW:
                    j += 1
                best = max(best, i - j + 1)
            if best >= limit:
                out.append((ip, best, reason, len(times)))
                break
    return out


def nft_ensure():
    if sh(["nft", "list", "table", "inet", TABLE]).returncode == 0:
        return
    rules = f"""table inet {TABLE} {{
  set banned4 {{ type ipv4_addr; flags timeout; }}
  set banned6 {{ type ipv6_addr; flags timeout; }}
  chain pre {{
    type filter hook prerouting priority -300; policy accept;
    ip saddr @banned4 tcp dport {{ 80, 443 }} counter drop
    ip6 saddr @banned6 tcp dport {{ 80, 443 }} counter drop
  }}
}}
"""
    r = subprocess.run(["nft", "-f", "-"], input=rules, capture_output=True, text=True)
    if r.returncode:
        raise RuntimeError(f"nft: {r.stderr.strip()}")


def nft_set(ip, seconds):
    s = "banned6" if ":" in ip else "banned4"
    sh(["nft", "delete", "element", "inet", TABLE, s, f"{{ {ip} }}"])
    if seconds > 0:
        sh(["nft", "add", "element", "inet", TABLE, s, f"{{ {ip} timeout {int(seconds)}s }}"], check=True)


def load():
    try:
        return json.load(open(STATE))
    except (OSError, ValueError):
        return {"bans": {}, "history": []}


def save(state):
    os.makedirs(DIR, exist_ok=True)
    json.dump(state, open(STATE, "w"))
    now = time.time()
    iso = lambda t: datetime.fromtimestamp(t, timezone.utc).isoformat(timespec="seconds")
    active = [{"ip": ip, **{k: (iso(v) if k in ("since", "until") else v) for k, v in b.items()}}
              for ip, b in sorted(state["bans"].items(), key=lambda x: -x[1]["since"])]
    recent = [{**h, "since": iso(h["since"]), "until": iso(h["until"])} for h in state["history"][-50:]][::-1]
    tmp = STATUS + ".tmp"
    json.dump({"checked_at": iso(now), "active": active, "recent": recent,
               "rules": f"block 24 h (7 d if repeated) at {RULES[401][0]} invalid keys or {RULES[403][0]} blocked paths within {WINDOW // 60} min"},
              open(tmp, "w"))
    os.chmod(tmp, 0o644)
    os.replace(tmp, STATUS)


def run():
    state, now = load(), time.time()
    nft_ensure()
    for ip, b in list(state["bans"].items()):          # expire, and replay into nft (after a reboot the set is empty)
        if b["until"] <= now:
            del state["bans"][ip]
        else:
            nft_set(ip, b["until"] - now)
    for ip, count, reason, total in verdicts(offenders("3m"), protected()):
        if ip in state["bans"]:
            continue
        repeat = any(h["ip"] == ip and h["since"] > now - 7 * 86400 for h in state["history"])
        b = {"since": now, "until": now + (REPEAT_BAN if repeat else BAN), "count": count, "reason": reason, "repeat": repeat}
        nft_set(ip, b["until"] - now)
        state["bans"][ip] = b
        state["history"] = (state["history"] + [{"ip": ip, **b}])[-500:]
        print(f"blocked {ip}: {count} {reason} within {WINDOW // 60} min ({'7 d, repeated' if repeat else '24 h'})")
    save(state)


def main():
    if os.geteuid() != 0:
        sys.exit("run as root")
    if len(sys.argv) == 4 and sys.argv[1] == "--replay":
        found = verdicts(offenders(sys.argv[2], sys.argv[3]), protected())
        for ip, count, reason, total in found:
            print(f"would block {ip}: {count} {reason} within {WINDOW // 60} min ({total} in the window)")
        print(f"{len(found)} client(s) over the limits")
    elif len(sys.argv) == 3 and sys.argv[1] == "--unban":
        state = load()
        state["bans"].pop(sys.argv[2], None)
        nft_ensure()
        nft_set(sys.argv[2], 0)
        save(state)
        print(f"unblocked {sys.argv[2]}")
    else:
        run()


if __name__ == "__main__":
    main()

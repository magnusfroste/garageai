#!/usr/bin/env python3
"""GarageAI Operations Center collector.

Runs every minute on the gateway (garageai-ops.timer) and writes one JSON file with
everything an operator needs to see: the gateway host, containers, certificates,
backups, the NetBird mesh, every garage and provider, gateway traffic and errors, the
portal seen from outside, and a list of current alerts. The page in index.html renders
it. The file holds no secrets: no keys, no tokens, no prompts.

Config: /etc/garageai/gateway-health.env (same as the health service).
Output: /var/lib/garageai-ops/www/ops.json; state in /var/lib/garageai-ops/state.json.
"""
import json
import os
import re
import shutil
import socket
import ssl
import subprocess
import time
import urllib.error
import urllib.request
from datetime import datetime, timezone

OUT_DIR = "/var/lib/garageai-ops/www"
STATE = "/var/lib/garageai-ops/state.json"
HEALTH_LAST = "/var/lib/garageai-health/last.json"
BACKUP_DIR = "/var/backups/garageai"
LITELLM_DIR = "/opt/garageai/litellm"
CERT_HOSTS = ["llm.garageai.eu", "netbird.garageai.eu", "ops.garageai.eu", "app.garageai.eu", "www.garageai.eu"]
TIMERS = ["garageai-health.timer", "garageai-backup.timer", "garageai-ops.timer"]
EXPECTED_POLICIES = {"gateway-to-garages"}
UA = "garageai-ops/1 (+https://garageai.eu)"
HISTORY_MINUTES = 24 * 60


def now():
    return datetime.now(timezone.utc)


def sh(cmd, timeout=20):
    try:
        return subprocess.run(cmd, capture_output=True, text=True, timeout=timeout).stdout
    except Exception:  # noqa: BLE001
        return ""


def http(url, headers=None, timeout=8):
    req = urllib.request.Request(url, headers={"User-Agent": UA, **(headers or {})})
    t0 = time.time()
    try:
        with urllib.request.urlopen(req, timeout=timeout) as r:
            body = r.read()
            return r.status, body, round((time.time() - t0) * 1000)
    except urllib.error.HTTPError as e:
        return e.code, b"", round((time.time() - t0) * 1000)
    except Exception:  # noqa: BLE001
        return None, b"", round((time.time() - t0) * 1000)


def load_env(path="/etc/garageai/gateway-health.env"):
    env = {}
    try:
        for line in open(path):
            line = line.strip()
            if line and not line.startswith("#") and "=" in line:
                k, v = line.split("=", 1)
                env[k] = v.strip().strip('"').strip("'")
    except OSError:
        pass
    return env


def read_json(path, default):
    try:
        return json.load(open(path))
    except Exception:  # noqa: BLE001
        return default


# ---------------------------------------------------------------- host
def host_info():
    load = open("/proc/loadavg").read().split()[:3]
    mem = {}
    for line in open("/proc/meminfo"):
        k, v = line.split(":")
        mem[k] = int(v.split()[0]) // 1024
    disk = shutil.disk_usage("/")
    uptime_h = float(open("/proc/uptime").read().split()[0]) / 3600
    return {
        "load": [float(x) for x in load],
        "cpus": os.cpu_count(),
        "mem_total_mb": mem.get("MemTotal"),
        "mem_available_mb": mem.get("MemAvailable"),
        "swap_used_mb": mem.get("SwapTotal", 0) - mem.get("SwapFree", 0),
        "disk_used_pct": round(disk.used / disk.total * 100, 1),
        "disk_free_gb": round(disk.free / 1e9, 1),
        "uptime_hours": round(uptime_h, 1),
        "reboot_required": os.path.exists("/var/run/reboot-required"),
    }


def pending_updates(state):
    cached = state.get("updates")
    if cached and time.time() - cached["at"] < 3600:
        return cached["count"]
    out = sh(["apt", "list", "--upgradable"], timeout=60)
    count = sum(1 for line in out.splitlines() if "upgradable" in line)
    state["updates"] = {"at": time.time(), "count": count}
    return count


def timers():
    rows = []
    for unit in TIMERS:
        active = sh(["systemctl", "is-active", unit]).strip()
        svc = unit.replace(".timer", ".service")
        show = dict(line.split("=", 1) for line in sh(["systemctl", "show", svc, "-p", "ExecMainStatus", "-p", "ExecMainExitTimestamp"]).splitlines() if "=" in line)
        rows.append({"unit": unit, "active": active, "last_exit": show.get("ExecMainStatus"), "last_run": show.get("ExecMainExitTimestamp") or None})
    return rows


def backup_info():
    try:
        files = sorted((os.path.join(BACKUP_DIR, f) for f in os.listdir(BACKUP_DIR) if f.endswith(".tgz")), key=os.path.getmtime)
    except OSError:
        return {"newest": None}
    if not files:
        return {"newest": None}
    newest = files[-1]
    return {"newest": os.path.basename(newest), "age_hours": round((time.time() - os.path.getmtime(newest)) / 3600, 1),
            "size_mb": round(os.path.getsize(newest) / 1e6, 1), "count": len(files)}


def cert_days(host):
    try:
        ctx = ssl.create_default_context()
        with socket.create_connection((host, 443), timeout=6) as s, ctx.wrap_socket(s, server_hostname=host) as t:
            not_after = t.getpeercert()["notAfter"]
        expires = datetime.strptime(not_after, "%b %d %H:%M:%S %Y %Z").replace(tzinfo=timezone.utc)
        return {"host": host, "days_left": (expires - now()).days, "expires": expires.date().isoformat()}
    except Exception as e:  # noqa: BLE001
        return {"host": host, "days_left": None, "error": type(e).__name__}


def containers():
    out = sh(["docker", "ps", "-a", "--format", "{{json .}}"])
    rows = []
    for line in out.splitlines():
        try:
            c = json.loads(line)
        except ValueError:
            continue
        rows.append({"name": c.get("Names"), "image": c.get("Image"), "state": c.get("State"), "status": c.get("Status")})
    return sorted(rows, key=lambda r: r["name"] or "")


# ---------------------------------------------------------------- mesh
def mesh(env):
    local = {}
    try:
        st = json.loads(sh(["netbird", "status", "--json"]) or "{}")
        local = {
            "management": (st.get("management") or {}).get("connected"),
            "signal": (st.get("signal") or {}).get("connected"),
            "relays_available": sum(1 for r in (st.get("relays") or {}).get("details", []) if r.get("available")),
            "relays_total": len((st.get("relays") or {}).get("details", [])),
            "ip": st.get("netbirdIp"),
            "version": st.get("daemonVersion"),
        }
        links = {}
        for p in (st.get("peers") or {}).get("details", []):
            links[(p.get("fqdn") or "").split(".")[0]] = {
                "status": p.get("status"), "type": p.get("connectionType"),
                "latency_ms": round((p.get("latency") or 0) / 1e6, 1) if isinstance(p.get("latency"), (int, float)) else p.get("latency"),
                "handshake": p.get("lastWireguardHandshake"),
            }
    except ValueError:
        links = {}
    api = env.get("NETBIRD_API_URL", "").rstrip("/")
    token = env.get("NETBIRD_API_TOKEN", "")
    peers, valid_keys, policies = [], None, []
    if api and token:
        hdr = {"Authorization": f"Token {token}", "Accept": "application/json"}
        code, body, _ = http(f"{api}/peers", hdr)
        if code == 200:
            for p in json.loads(body):
                name = p.get("name")
                peers.append({
                    "name": name, "connected": p.get("connected"), "last_seen": p.get("last_seen"),
                    "version": p.get("version"), "os": p.get("os"), "country": p.get("country_code"),
                    "groups": [g.get("name") for g in p.get("groups") or [] if g.get("name") != "All"],
                    "link": links.get(name),
                })
        code, body, _ = http(f"{api}/setup-keys", hdr)
        if code == 200:
            valid_keys = sum(1 for k in json.loads(body) if k.get("valid") and not k.get("revoked"))
        code, body, _ = http(f"{api}/policies", hdr)
        if code == 200:
            for p in json.loads(body):
                rules = p.get("rules") or []
                policies.append({
                    "name": p.get("name"), "enabled": p.get("enabled"),
                    "sources": sorted({g.get("name") for r in rules for g in r.get("sources") or []}),
                    "destinations": sorted({g.get("name") for r in rules for g in r.get("destinations") or []}),
                    "ports": sorted({pt for r in rules for pt in r.get("ports") or []}),
                    "bidirectional": any(r.get("bidirectional") for r in rules),
                })
    return {"local": local, "peers": peers, "valid_setup_keys": valid_keys, "policies": policies}


# ---------------------------------------------------------------- garages and providers
def portal_targets(env):
    url = env.get("GARAGEAI_PORTAL_URL", "").rstrip("/")
    key = env.get("GARAGEAI_GATEWAY_KEY", "")
    if not url or not key:
        return None, None, None
    code, body, ms = http(f"{url}/functions/v1/gateway-targets", {"x-gateway-key": key}, timeout=15)
    if code != 200:
        return code, ms, None
    data = json.loads(body)
    return code, ms, data.get("targets", data) if isinstance(data, dict) else data


def vllm_metrics(host, port, key, state, name):
    hdr = {"Authorization": f"Bearer {key}"} if key else {}
    code, body, _ = http(f"http://{host}:{port}/metrics", hdr, timeout=5)
    if code != 200:
        return None
    vals = {}
    for line in body.decode(errors="ignore").splitlines():
        m = re.match(r"^(vllm:[a-z_]+)(\{[^}]*\})?\s+([0-9.eE+-]+)$", line)
        if m:
            vals[m.group(1)] = vals.get(m.group(1), 0.0) + float(m.group(3))
    prev = state.setdefault("vllm", {}).get(name)
    t = time.time()
    out = {
        "running": vals.get("vllm:num_requests_running"),
        "waiting": vals.get("vllm:num_requests_waiting"),
        "kv_cache_pct": round(100 * (vals.get("vllm:kv_cache_usage_perc", vals.get("vllm:gpu_cache_usage_perc", 0.0))), 1),
    }
    cur = {"t": t, "prompt": vals.get("vllm:prompt_tokens_total", 0.0), "gen": vals.get("vllm:generation_tokens_total", 0.0)}
    if prev and t > prev["t"]:
        dt = t - prev["t"]
        out["prompt_tok_s"] = round(max(0.0, cur["prompt"] - prev["prompt"]) / dt, 1)
        out["gen_tok_s"] = round(max(0.0, cur["gen"] - prev["gen"]) / dt, 1)
    state["vllm"][name] = cur
    return out


def provider_checks(target):
    url = target.get("url") or ""
    host = re.sub(r"^https?://", "", url).split("/")[0].split(":")[0]
    info = {"host": host}
    try:
        ips = sorted({a[4][0] for a in socket.getaddrinfo(host, 443)})
        info["ips"] = ips
        info["private_ip"] = any(ip.startswith(("10.", "192.168.", "127.", "169.254.", "100.64.", "fc", "fd")) or re.match(r"^172\.(1[6-9]|2\d|3[01])\.", ip) for ip in ips)
    except OSError:
        info["ips"] = []
    info["tls"] = cert_days(host) if host else None
    return info


def garages(env, state, targets):
    health = read_json(HEALTH_LAST, {})
    by_name = {r.get("garage"): r for r in health.get("results", [])}
    hist = state.setdefault("history", {})
    minute = int(time.time() // 60)
    rows, providers = [], []
    for t in targets or []:
        name = t.get("garage")
        h = by_name.get(name, {})
        up = bool(h.get("runtime_ok")) and (h.get("mesh_connected") is not False)
        series = hist.setdefault(name, [])
        if not series or series[-1][0] != minute:
            series.append([minute, 1 if up else 0])
        hist[name] = series[-HISTORY_MINUTES:]
        last60 = [s[1] for s in hist[name] if s[0] > minute - 60]
        last24 = [s[1] for s in hist[name]]
        row = {
            "garage": name, "type": "provider" if t.get("endpoint") else "mesh",
            "mesh": h.get("mesh_connected"), "runtime_ok": h.get("runtime_ok"), "error": h.get("runtime_error"),
            "models": len(h.get("models") or []), "checked_at": h.get("checked_at"),
            "up_1h_pct": round(100 * sum(last60) / len(last60), 1) if last60 else None,
            "up_24h_pct": round(100 * sum(last24) / len(last24), 1) if last24 else None,
            "minutes_tracked": len(last24),
        }
        if t.get("endpoint"):
            row["provider"] = provider_checks(t)
            providers.append(name)
        elif t.get("host") and t.get("port"):
            row["vllm"] = vllm_metrics(t["host"], t["port"], t.get("runtime_api_key"), state, name)
        rows.append(row)
    for name in list(hist):
        if name not in {t.get("garage") for t in targets or []}:
            del hist[name]
    return rows


# ---------------------------------------------------------------- traffic and errors
def psql(sql):
    out = sh(["docker", "compose", "-f", f"{LITELLM_DIR}/docker-compose.yml", "exec", "-T", "postgres",
              "psql", "-U", "litellm", "-d", "litellm", "-AtF", "\t", "-c", sql], timeout=30)
    return [line.split("\t") for line in out.splitlines() if line.strip()]


def traffic():
    sql = """
      select coalesce(nullif(split_part(model_id, '__', 1), ''), '?') garage, model_group, count(*),
             count(*) filter (where status <> 'success'),
             coalesce(sum(prompt_tokens),0), coalesce(sum(completion_tokens),0), round(coalesce(sum(spend),0)::numeric,4),
             round((percentile_cont(0.5) within group (order by extract(epoch from ("completionStartTime"-"startTime"))))::numeric, 2),
             round((percentile_cont(0.95) within group (order by extract(epoch from ("completionStartTime"-"startTime"))))::numeric, 2),
             round(avg(prompt_tokens)::numeric, 0)
      from "LiteLLM_SpendLogs"
      where "startTime" > now() - interval '{iv}' and model_group not like 'probe/%' and model_group not like 'garage-probe%'
      group by 1, 2 order by 3 desc"""
    cols = ["garage", "model", "requests", "failed", "prompt_tokens", "completion_tokens", "spend_usd", "ttft_p50_s", "ttft_p95_s", "avg_prompt"]
    res = {}
    for label, iv in (("last_hour", "1 hour"), ("last_24h", "24 hours")):
        rows = []
        for r in psql(sql.format(iv=iv)):
            row = dict(zip(cols, r))
            for k in cols[2:]:
                try:
                    row[k] = float(row[k]) if "." in row[k] else int(row[k])
                except (ValueError, TypeError):
                    row[k] = None
            rows.append(row)
        res[label] = rows
    return res


def gateway_log_counts():
    out = sh(["docker", "compose", "-f", f"{LITELLM_DIR}/docker-compose.yml", "logs", "--since", "60m", "litellm"], timeout=60)
    endpoints = {}
    for m in re.finditer(r'"(POST|GET) (/v1/[a-z_/]+)[^"]*" (\d{3})', out):
        key = f"{m.group(2)} {m.group(3)}"
        endpoints[key] = endpoints.get(key, 0) + 1
    def count(pattern):
        return len(re.findall(pattern, out))
    return {
        "endpoints": dict(sorted(endpoints.items(), key=lambda kv: -kv[1])),
        "invalid_key": count(r"Key not found in database"),
        "context_rejected": count(r"garageai: context check rejected"),
        "timeouts_408": count(r'" 408'),
        "mid_stream_failures": count(r"MidStreamFallbackError"),
        "cooldown_429": count(r"All deployments for selected model are in cooldown"),
        "server_errors_5xx": count(r'" 5\d\d'),
        "token_count_failures": count(r"failed to count tokens"),
    }


# ---------------------------------------------------------------- outside view
def outside():
    rows = []
    for name, url in (("gateway", "https://llm.garageai.eu/health/liveliness"), ("portal", "https://app.garageai.eu/"),
                      ("site", "https://www.garageai.eu/"), ("netbird", "https://netbird.garageai.eu/")):
        code, _, ms = http(url, timeout=10)
        rows.append({"name": name, "url": url, "status": code, "ms": ms})
    return rows


# ---------------------------------------------------------------- alerts
def alerts(d):
    a = []
    def add(level, text):
        a.append({"level": level, "text": text})
    h = d["host"]
    if h["disk_used_pct"] >= 90:
        add("critical", f"Disk {h['disk_used_pct']} % full")
    elif h["disk_used_pct"] >= 80:
        add("warning", f"Disk {h['disk_used_pct']} % full")
    if (h["mem_available_mb"] or 0) < 300:
        add("warning", f"Only {h['mem_available_mb']} MB memory available")
    if h["swap_used_mb"] > 1024:
        add("warning", f"Swap in use: {h['swap_used_mb']} MB")
    if h["reboot_required"]:
        add("warning", "The gateway needs a reboot for installed updates")
    for c in d["containers"]:
        if c["state"] != "running":
            add("critical", f"Container {c['name']} is {c['state']}")
    for t in d["timers"]:
        if t["active"] != "active":
            add("critical", f"{t['unit']} is {t['active']}")
        elif t["last_exit"] not in (None, "", "0"):
            add("warning", f"{t['unit']}: last run exited with {t['last_exit']}")
    b = d["backup"]
    if not b.get("newest"):
        add("critical", "No backup found")
    elif b["age_hours"] > 26:
        add("warning", f"Newest backup is {b['age_hours']} h old")
    for c in d["certs"]:
        if c.get("days_left") is None:
            add("warning", f"Could not read the certificate for {c['host']}")
        elif c["days_left"] < 3:
            add("critical", f"Certificate for {c['host']} expires in {c['days_left']} days")
        elif c["days_left"] < 14:
            add("warning", f"Certificate for {c['host']} expires in {c['days_left']} days")
    m = d["mesh"]
    if m["local"].get("management") is False or m["local"].get("signal") is False:
        add("critical", "The gateway's NetBird client is not connected to management/signal")
    if m.get("valid_setup_keys"):
        add("warning", f"{m['valid_setup_keys']} NetBird setup key(s) are still valid")
    for p in m.get("policies", []):
        if p["bidirectional"]:
            add("critical", f"NetBird policy {p['name']} is bidirectional: garages could reach the gateway")
        if p["name"] not in EXPECTED_POLICIES:
            add("info", f"Extra NetBird policy: {p['name']}")
    if d["portal"]["targets_status"] != 200:
        add("critical", f"Portal gateway-targets answered {d['portal']['targets_status']}")
    for g in d["garages"]:
        if not g["runtime_ok"] or g["mesh"] is False:
            add("critical", f"{g['garage']}: {'tunnel down' if g['mesh'] is False else 'runtime not answering'}"
                            + (f" ({g['error']})" if g.get("error") else ""))
        p = g.get("provider") or {}
        if p.get("private_ip"):
            add("critical", f"Provider {g['garage']} now resolves to a private address {p.get('ips')}")
        tls = p.get("tls") or {}
        if tls.get("days_left") is not None and tls["days_left"] < 14:
            add("warning", f"Provider {g['garage']}: TLS certificate expires in {tls['days_left']} days")
    for o in d["outside"]:
        if o["status"] != 200:
            add("critical", f"{o['name']} ({o['url']}) answered {o['status']}")
    e = d["errors_last_hour"]
    if e["invalid_key"] > 50:
        add("warning", f"{e['invalid_key']} requests with an unknown API key in the last hour")
    if e["timeouts_408"] > 5:
        add("warning", f"{e['timeouts_408']} timeouts (408) in the last hour")
    if e["mid_stream_failures"]:
        add("warning", f"{e['mid_stream_failures']} streams broke mid-way in the last hour")
    if e["server_errors_5xx"] > 5:
        add("warning", f"{e['server_errors_5xx']} server errors (5xx) in the last hour")
    if d["host"]["updates_pending"]:
        add("info", f"{d['host']['updates_pending']} package updates pending (unattended-upgrades handles security updates)")
    order = {"critical": 0, "warning": 1, "info": 2}
    return sorted(a, key=lambda x: order[x["level"]])


def main():
    env = load_env()
    state = read_json(STATE, {})
    t0 = time.time()
    code, ms, targets = portal_targets(env)
    data = {
        "generated_at": now().isoformat(timespec="seconds"),
        "host": {**host_info(), "updates_pending": pending_updates(state)},
        "timers": timers(),
        "backup": backup_info(),
        "certs": [cert_days(h) for h in CERT_HOSTS],
        "containers": containers(),
        "mesh": mesh(env),
        "portal": {"targets_status": code, "targets_ms": ms, "garages_listed": len(targets or [])},
        "garages": garages(env, state, targets),
        "traffic": traffic(),
        "errors_last_hour": gateway_log_counts(),
        "outside": outside(),
    }
    data["alerts"] = alerts(data)
    data["collect_seconds"] = round(time.time() - t0, 1)
    os.makedirs(OUT_DIR, exist_ok=True)
    tmp = os.path.join(OUT_DIR, ".ops.json.tmp")
    with open(tmp, "w") as f:
        json.dump(data, f, separators=(",", ":"))
    os.chmod(tmp, 0o644)
    os.replace(tmp, os.path.join(OUT_DIR, "ops.json"))
    json.dump(state, open(STATE, "w"))
    crit = sum(1 for x in data["alerts"] if x["level"] == "critical")
    warn = sum(1 for x in data["alerts"] if x["level"] == "warning")
    print(f"ops: {len(data['garages'])} garages, {crit} critical, {warn} warning, {data['collect_seconds']} s")


if __name__ == "__main__":
    main()

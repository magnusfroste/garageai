#!/usr/bin/env python3
"""GarageAI Operations Center collector.

Runs every minute on the gateway (garageai-ops.timer) and writes one JSON file with
everything an operator needs to see: the gateway host, containers, certificates,
backups, the NetBird mesh, every garage and provider, gateway traffic and errors, the
portal seen from outside, and a list of current alerts. The page in index.html renders
it. The file holds no secrets: no keys, no tokens, no prompts.

Config: /etc/garageai/gateway-health.env (same as the health service).
Output: /var/lib/garageai-ops/www/ops.json; state in /var/lib/garageai-ops/state.json.

History and events: /var/lib/garageai-ops/history.db (SQLite). Samples every minute (30 days),
events (180 days): alerts opened and resolved, reboots, container restarts, version changes,
deployed files, merges to main, mesh and supply changes, and notes. Every 5 minutes the page's
history.json is written from it (and from LiteLLM's spend logs for hourly traffic).
  garageai-ops-collect --note "text"      add a note to the event log (e.g. a maintenance window)

Telegram notifications (optional): /etc/garageai/telegram.env (root, 0600) with
TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID. One-way: the bot only sends to that chat.
  garageai-ops-collect --telegram-setup   find the chat that messaged the bot and save its id
  garageai-ops-collect --telegram-test    send a test message
"""
import hashlib
import json
import os
import platform
import re
import shutil
import socket
import sqlite3
import ssl
import subprocess
import sys
import time
import urllib.error
import urllib.request
from datetime import datetime, timedelta, timezone
from zoneinfo import ZoneInfo

OUT_DIR = "/var/lib/garageai-ops/www"
STATE = "/var/lib/garageai-ops/state.json"
HEALTH_LAST = "/var/lib/garageai-health/last.json"
BACKUP_DIR = "/var/backups/garageai"
LITELLM_DIR = "/opt/garageai/litellm"
CERT_HOSTS = ["llm.garageai.eu", "netbird.garageai.eu", "ops.garageai.eu", "app.garageai.eu", "www.garageai.eu"]
TIMERS = ["garageai-health.timer", "garageai-backup.timer", "garageai-ops.timer", "garageai-guard.timer"]
GUARD_STATUS = "/var/lib/garageai-guard/status.json"
EXPECTED_POLICIES = {"gateway-to-garages"}
UA = "garageai-ops/1 (+https://garageai.eu)"
HISTORY_MINUTES = 24 * 60
REPO = "magnusfroste/garageai"
REPO_DIR = "/home/garageai/garageai"
# Deployed file -> its source in the repository (compared with main on GitHub).
DEPLOYED = {
    "/opt/garageai/litellm/config.yaml": "infra/gateway/litellm/config.yaml",
    "/opt/garageai/litellm/docker-compose.yml": "infra/gateway/litellm/docker-compose.yml",
    "/opt/garageai/litellm/garageai_callbacks.py": "infra/gateway/litellm/garageai_callbacks.py",
    "/usr/local/bin/garageai-health": "infra/gateway/health/garageai-health.py",
    "/usr/local/sbin/garageai-backup": "infra/gateway/backup/garageai-backup",
    "/usr/local/bin/garageai-ops-collect": "infra/gateway/ops/garageai-ops-collect.py",
    "/var/lib/garageai-ops/www/index.html": "infra/gateway/ops/index.html",
    "/opt/garageai/ops/docker-compose.yml": "infra/gateway/ops/docker-compose.yml",
    "/opt/garageai/ops/nginx.conf.template": "infra/gateway/ops/nginx.conf.template",
    "/usr/local/sbin/garageai-guard": "infra/gateway/guard/garageai-guard.py",
}
LOG_ALERT_MB = 400
OPS_URL = "https://ops.garageai.eu"
HISTORY_DB = "/var/lib/garageai-ops/history.db"
SAMPLE_DAYS = 30
EVENT_DAYS = 180
HISTORY_EVERY = 300           # seconds between history.json writes
LITELLM_URL = "https://llm.garageai.eu"
CONTEXT_EVERY = 600           # seconds between reading each runtime's /v1/models for context windows
QUEUE_ALERT_MINUTES = 5       # a vLLM garage with requests waiting this long in a row is full
TELEGRAM_ENV = "/etc/garageai/telegram.env"
# Minutes in a row an alert must be present before it is sent: short blips stay quiet, and a
# deploy that briefly differs from main does not page anyone. Info alerts are never sent.
NOTIFY_AFTER = {"critical": 2, "warning": 15}
REMIND_MINUTES = 240          # repeat open critical alerts this often
DAILY_REPORT = (7, "Europe/Stockholm")


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
    peers, valid_keys, key_names, policies = [], None, [], []
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
            valid = [k for k in json.loads(body) if k.get("valid") and not k.get("revoked")]
            valid_keys, key_names = len(valid), sorted(k.get("name", "?") for k in valid)
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
    return {"local": local, "peers": peers, "valid_setup_keys": valid_keys, "valid_setup_key_names": key_names, "policies": policies}


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
        onboarding = not t.get("endpoint") and not (t.get("host") and t.get("port"))
        if not onboarding and (not series or series[-1][0] != minute):   # onboarding is not downtime
            series.append([minute, 1 if up else 0])
        hist[name] = series[-HISTORY_MINUTES:]
        last60 = [s[1] for s in hist[name] if s[0] > minute - 60]
        last24 = [s[1] for s in hist[name]]
        row = {
            "garage": name, "type": "provider" if t.get("endpoint") else "mesh",
            "mesh": h.get("mesh_connected"), "runtime_ok": h.get("runtime_ok"), "error": h.get("runtime_error"),
            # Not registered yet: the operator is still onboarding (joined the mesh or not, no host and port in
            # the portal). That is not an outage.
            "onboarding": not t.get("endpoint") and not (t.get("host") and t.get("port")),
            # The connect script's latest step, from the portal (gateway-targets), while onboarding.
            "onboarding_step": t.get("onboarding"),
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
            streaks = state.setdefault("queue_streak", {})
            streaks[name] = streaks.get(name, 0) + 1 if (row["vllm"] or {}).get("waiting") else 0
            row["queue_minutes"] = streaks[name]
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
             round((percentile_cont(0.5) within group (order by extract(epoch from ("completionStartTime"-"startTime"))) filter (where "completionStartTime" < "endTime"))::numeric, 2),
             round((percentile_cont(0.95) within group (order by extract(epoch from ("completionStartTime"-"startTime"))) filter (where "completionStartTime" < "endTime"))::numeric, 2),
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


# ---------------------------------------------------------------- guard
def guard():
    """Clients blocked by garageai-guard, and how many packets the block has dropped."""
    st = read_json(GUARD_STATUS, None)
    if st is None:
        return None
    dropped = 0
    for line in sh(["nft", "list", "chain", "inet", "garageai_guard", "pre"]).splitlines():
        m = re.search(r"counter packets (\d+)", line)
        if m:
            dropped += int(m.group(1))
    return {**st, "dropped_packets": dropped}


# ---------------------------------------------------------------- versions, source, logs
def cached(state, key, seconds, fn):
    entry = state.setdefault("cache", {}).get(key)
    if entry and time.time() - entry["at"] < seconds:
        return entry["value"]
    value = fn()
    if value is not None:
        state["cache"][key] = {"at": time.time(), "value": value}
    return value if value is not None else (entry or {}).get("value")


def github_json(path):
    code, body, _ = http(f"https://api.github.com/{path}", {"Accept": "application/vnd.github+json"}, timeout=10)
    return json.loads(body) if code == 200 else None


def image_of(conts, name):
    return next((c["image"] for c in conts if c["name"] == name), "")


def versions(env, state, conts):
    rows = []
    litellm = image_of(conts, "garageai-litellm-litellm-1").rsplit(":", 1)[-1]
    latest = cached(state, "litellm_latest", 3600, lambda: (github_json("repos/BerriAI/litellm/releases/latest") or {}).get("tag_name"))
    rows.append({"component": "LiteLLM (gateway)", "running": litellm, "latest": latest,
                 "update": bool(latest and litellm and latest != litellm),
                 "how": "set LITELLM_VERSION in /opt/garageai/litellm/.env, then: cd /opt/garageai/litellm && sudo docker compose pull && sudo docker compose up -d"})
    nb = {}
    api, token = env.get("NETBIRD_API_URL", "").rstrip("/"), env.get("NETBIRD_API_TOKEN", "")
    if api and token:
        code, body, _ = http(f"{api}/instance/version", {"Authorization": f"Token {token}"})
        nb = json.loads(body) if code == 200 else {}
    server_img = image_of(conts, "netbird-server").rsplit(":", 1)[-1]
    rows.append({"component": "NetBird management/relay", "running": nb.get("management_current_version") or server_img,
                 "latest": nb.get("management_available_version"), "update": bool(nb.get("management_update_available")),
                 "pinned": server_img != "latest",
                 "how": "change the netbird-server tag in /opt/garageai/netbird/docker-compose.yml, then: sudo docker compose pull && sudo docker compose up -d (quiet time)"})
    dash_img = image_of(conts, "netbird-dashboard").rsplit(":", 1)[-1]
    dash_avail = nb.get("dashboard_available_version")
    rows.append({"component": "NetBird dashboard", "running": dash_img, "latest": dash_avail and f"v{dash_avail}",
                 "update": bool(dash_avail and dash_img not in (f"v{dash_avail}", dash_avail, "latest")), "pinned": dash_img != "latest",
                 "how": "change the dashboard tag in /opt/garageai/netbird/docker-compose.yml, then: sudo docker compose up -d"})
    client = sh(["netbird", "version"]).strip()
    rows.append({"component": "NetBird client (gateway peer)", "running": client, "latest": nb.get("management_available_version"),
                 "update": bool(client and nb.get("management_available_version") and client != nb.get("management_available_version")),
                 "how": "sudo apt update && sudo apt install --only-upgrade netbird"})
    traefik = image_of(conts, "netbird-traefik").rsplit(":", 1)[-1]
    t_latest = cached(state, "traefik_latest", 3600, lambda: (github_json("repos/traefik/traefik/releases/latest") or {}).get("tag_name"))
    rows.append({"component": "Traefik", "running": traefik, "latest": t_latest,
                 "update": bool(t_latest and not t_latest.startswith(traefik)),
                 "how": "minor/patch updates arrive with the floating tag on recreate; a new minor needs the tag changed in /opt/garageai/netbird/docker-compose.yml"})
    os_name = ""
    try:
        os_name = dict(l.strip().split("=", 1) for l in open("/etc/os-release") if "=" in l).get("PRETTY_NAME", "").strip('"')
    except OSError:
        pass
    rows.append({"component": "Host OS", "running": f"{os_name} · kernel {platform.release()}", "latest": None, "update": False,
                 "how": "security updates install automatically (unattended-upgrades); reboot when the alert says so: sudo reboot"})
    return rows


def sha256_file(path):
    try:
        return hashlib.sha256(open(path, "rb").read()).hexdigest()
    except OSError:
        return None


def source(state):
    main = cached(state, "repo_main", 600, lambda: (github_json(f"repos/{REPO}/commits/main") or {}).get("sha"))
    local = sh(["git", "-c", "safe.directory=*", "-C", REPO_DIR, "rev-parse", "HEAD"]).strip() or None
    branch = sh(["git", "-c", "safe.directory=*", "-C", REPO_DIR, "rev-parse", "--abbrev-ref", "HEAD"]).strip() or None
    files = []
    remote = state.setdefault("cache", {}).setdefault("main_files", {"sha": None, "hashes": {}})
    if main and remote.get("sha") != main:
        hashes = {}
        for repo_path in DEPLOYED.values():
            code, body, _ = http(f"https://raw.githubusercontent.com/{REPO}/{main}/{repo_path}", timeout=10)
            hashes[repo_path] = hashlib.sha256(body).hexdigest() if code == 200 else None
        remote.update({"sha": main, "hashes": hashes})
    for deployed, repo_path in DEPLOYED.items():
        want, have = remote["hashes"].get(repo_path), sha256_file(deployed)
        files.append({"deployed": deployed, "source": repo_path,
                      "status": "missing" if have is None else "unknown" if want is None else "in sync" if want == have else "differs from main"})
    return {"repo": REPO, "main": main, "main_short": (main or "")[:7], "checkout": local and local[:7], "checkout_branch": branch,
            "files": files, "netbird_compose": "/opt/garageai/netbird/docker-compose.yml (installer-generated, not in the repo)"}


def logs(conts):
    rows = []
    for c in conts:
        info = sh(["docker", "inspect", c["name"], "--format", "{{.LogPath}}|{{json .HostConfig.LogConfig.Config}}"]).strip()
        path, _, opts = info.partition("|")
        try:
            size = os.path.getsize(path) if path else 0
        except OSError:
            size = 0
        try:
            cfg = json.loads(opts or "{}") or {}
        except ValueError:
            cfg = {}
        rows.append({"container": c["name"], "log_mb": round(size / 1e6, 1), "max_size": cfg.get("max-size"), "max_file": cfg.get("max-file")})
    journal = re.search(r"take up ([0-9.]+[KMGT]?)", sh(["journalctl", "--disk-usage"]))
    df = sh(["docker", "system", "df", "--format", "{{.Type}}|{{.Size}}|{{.Reclaimable}}"])
    docker_df = [dict(zip(["type", "size", "reclaimable"], l.split("|"))) for l in df.splitlines() if "|" in l]
    return {"containers": rows, "journald": journal.group(1) if journal else None, "docker": docker_df}


# ---------------------------------------------------------------- alerts
def alerts(d):
    a = []
    # page: the Operations Center page the alert belongs to (supply, traffic, mesh, server, updates, logs).
    def add(level, text, page):
        a.append({"level": level, "text": text, "page": page})
    onboarding = {g["garage"] for g in d["garages"] if g.get("onboarding")}
    h = d["host"]
    if h["disk_used_pct"] >= 90:
        add("critical", f"Disk {h['disk_used_pct']} % full", "server")
    elif h["disk_used_pct"] >= 80:
        add("warning", f"Disk {h['disk_used_pct']} % full", "server")
    if (h["mem_available_mb"] or 0) < 300:
        add("warning", f"Only {h['mem_available_mb']} MB memory available", "server")
    if h["swap_used_mb"] > 1024:
        add("warning", f"Swap in use: {h['swap_used_mb']} MB", "server")
    if h["reboot_required"]:
        add("warning", "The gateway needs a reboot for installed updates", "server")
    for c in d["containers"]:
        if c["state"] != "running":
            add("critical", f"Container {c['name']} is {c['state']}", "server")
        if str(c.get("image", "")).endswith(":latest"):
            add("warning", f"Container {c['name']} runs an unpinned image ({c['image']}); a restart may upgrade it", "updates")
    for t in d["timers"]:
        if t["active"] != "active":
            add("critical", f"{t['unit']} is {t['active']}", "server")
        elif t["last_exit"] not in (None, "", "0"):
            add("warning", f"{t['unit']}: last run exited with {t['last_exit']}", "server")
    b = d["backup"]
    if not b.get("newest"):
        add("critical", "No backup found", "server")
    elif b["age_hours"] > 26:
        add("warning", f"Newest backup is {b['age_hours']} h old", "server")
    for c in d["certs"]:
        if c.get("days_left") is None:
            add("warning", f"Could not read the certificate for {c['host']}", "server")
        elif c["days_left"] < 3:
            add("critical", f"Certificate for {c['host']} expires in {c['days_left']} days", "server")
        elif c["days_left"] < 14:
            add("warning", f"Certificate for {c['host']} expires in {c['days_left']} days", "server")
    m = d["mesh"]
    if m["local"].get("management") is False or m["local"].get("signal") is False:
        add("critical", "The gateway's NetBird client is not connected to management/signal", "mesh")
    if m.get("valid_setup_keys"):
        # A key handed to an operator who is onboarding right now is expected; any other valid key is not.
        names = m.get("valid_setup_key_names") or []
        stray = [n for n in names if n not in onboarding] if names else None
        if stray is None or stray:
            add("warning", f"{m['valid_setup_keys']} NetBird setup key(s) are still valid", "mesh")
        else:
            add("info", f"Setup key waiting for onboarding: {', '.join(names)}", "mesh")
    for p in m.get("policies", []):
        if p["bidirectional"]:
            add("critical", f"NetBird policy {p['name']} is bidirectional: garages could reach the gateway", "mesh")
        if p["name"] not in EXPECTED_POLICIES:
            add("info", f"Extra NetBird policy: {p['name']}", "mesh")
    if d["portal"]["targets_status"] != 200:
        add("critical", f"Portal gateway-targets answered {d['portal']['targets_status']}", "supply")
    for g in d["garages"]:
        if g.get("onboarding"):
            st = g.get("onboarding_step") or {}
            if st.get("status") in ("failed", "stopped"):
                add("info", f"{g['garage']} is onboarding, stuck at {' '.join(str(st.get('step', '?')).split())}: {st.get('message', '')}", "supply")
            else:
                add("info", f"{g['garage']} is onboarding: " + (f"{' '.join(str(st.get('step')).split())} ({st.get('status')})" if st.get("step")
                    else f"{'joined the mesh, ' if g['mesh'] else ''}not registered with the portal yet"), "supply")
            continue
        if not g["runtime_ok"] or g["mesh"] is False:
            add("critical", f"{g['garage']}: {'tunnel down' if g['mesh'] is False else 'runtime not answering'}"
                            + (f" ({g['error']})" if g.get("error") else ""), "supply")
        p = g.get("provider") or {}
        if p.get("private_ip"):
            add("critical", f"Provider {g['garage']} now resolves to a private address {p.get('ips')}", "supply")
        tls = p.get("tls") or {}
        if tls.get("days_left") is not None and tls["days_left"] < 14:
            add("warning", f"Provider {g['garage']}: TLS certificate expires in {tls['days_left']} days", "supply")
    md = d.get("models") or {}
    if md.get("error"):
        add("warning", md["error"], "models")
    seen = {}
    for r in md.get("deployments", []):
        for c in r["checks"]:
            if c["level"] == "info":
                continue
            seen.setdefault((r["public"], r["garage"], c["level"], c["text"]), []).append(r["tier"])
    for (public, garage, level, text), tiers in seen.items():
        add(level, f"{public} on {garage} ({'/'.join(sorted(set(tiers)))}): {text}", "models")
    for g in d["garages"]:
        if (g.get("queue_minutes") or 0) >= QUEUE_ALERT_MINUTES:
            add("warning", f"{g['garage']} has had requests waiting for {g['queue_minutes']} min: the garage is full", "supply")
    for g, q in (d.get("quality_24h") or {}).items():
        if (q.get("requests") or 0) >= 20 and q.get("success_pct") is not None and q["success_pct"] < 90:
            add("warning", f"{g}: only {q['success_pct']} % of {q['requests']} requests succeeded in the last 24 h", "supply")
    for v in d.get("versions", []):
        if v.get("update"):
            add("info", f"Update available: {v['component']} {v['running']} -> {v['latest']}", "updates")
    for f in d.get("source", {}).get("files", []):
        if f["status"] in ("differs from main", "missing"):
            add("warning", f"{f['deployed']} {f['status']} ({f['source']})", "updates")
    for l in d.get("logs", {}).get("containers", []):
        if l["log_mb"] > LOG_ALERT_MB:
            add("warning", f"Container log for {l['container']} is {l['log_mb']} MB", "logs")
        elif not l["max_size"] and l["log_mb"] > 100:
            add("warning", f"Container log for {l['container']} is {l['log_mb']} MB and has no size limit", "logs")
    gd = d.get("guard")
    if gd is None:
        add("warning", "garageai-guard has not run (no status file)", "traffic")
    elif (now() - datetime.fromisoformat(gd["checked_at"])).total_seconds() > 300:
        add("warning", f"garageai-guard last ran {gd['checked_at']}", "traffic")
    for o in d["outside"]:
        if o["status"] != 200:
            add("critical", f"{o['name']} ({o['url']}) answered {o['status']}", "server")
    e = d["errors_last_hour"]
    if e["invalid_key"] > 1000:
        add("critical", f"{e['invalid_key']} requests with an unknown API key in the last hour: someone is guessing keys", "traffic")
    elif e["invalid_key"] > 50:
        add("warning", f"{e['invalid_key']} requests with an unknown API key in the last hour", "traffic")
    if e["timeouts_408"] > 5:
        add("warning", f"{e['timeouts_408']} timeouts (408) in the last hour", "traffic")
    if e["mid_stream_failures"]:
        add("warning", f"{e['mid_stream_failures']} streams broke mid-way in the last hour", "traffic")
    if e["server_errors_5xx"] > 5:
        add("warning", f"{e['server_errors_5xx']} server errors (5xx) in the last hour", "traffic")
    if d["host"]["updates_pending"]:
        add("info", f"{d['host']['updates_pending']} package updates pending (unattended-upgrades handles security updates)", "updates")
    order = {"critical": 0, "warning": 1, "info": 2}
    return sorted(a, key=lambda x: order[x["level"]])


# ---------------------------------------------------------------- models: the sellable catalogue, checked
def runtime_contexts(t, state, served):
    """{runtime model id: context window} read from the runtime itself. The gateway is on the mesh, so it
    can ask a garage directly, which the portal (outside the mesh) cannot. Cached; re-read when the
    runtime's model list changes."""
    cache = state.setdefault("contexts", {})
    name = t.get("garage")
    entry = cache.get(name)
    sig = ",".join(sorted(served or []))
    if entry and entry.get("sig") == sig and time.time() - entry["at"] < CONTEXT_EVERY:
        return entry["value"]
    key = t.get("runtime_api_key")
    hdr = {"Authorization": f"Bearer {key}"} if key else {}
    url = f"{t['url'].rstrip('/')}/models" if t.get("endpoint") and t.get("url") else (
          f"http://{t['host']}:{t['port']}/v1/models" if t.get("host") and t.get("port") else None)
    value = {}
    if url:
        code, body, _ = http(url, hdr, timeout=8)
        if code == 200:
            try:
                for m in json.loads(body).get("data", []):
                    n = m.get("max_model_len") or m.get("context_length") or m.get("context_window") or (m.get("meta") or {}).get("n_ctx_train")
                    if isinstance(m.get("id"), str) and isinstance(n, (int, float)) and n > 0:
                        value[m["id"]] = int(n)
            except ValueError:
                pass
        elif entry:
            value = entry["value"]
    cache[name] = {"at": time.time(), "sig": sig, "value": value}
    return value


def models(env, state, targets):
    """Every LiteLLM deployment with what buyers get (public name, tier, price, context, output cap)
    and checks that it is consistent with the garage behind it."""
    key = env.get("GARAGEAI_GATEWAY_KEY", "")
    code, body, _ = http(f"{LITELLM_URL}/v1/model/info", {"Authorization": f"Bearer {key}"}, timeout=15)
    if code != 200:
        return {"error": f"LiteLLM /v1/model/info answered {code}"}
    health = {r.get("garage"): r for r in read_json(HEALTH_LAST, {}).get("results", [])}
    by_garage = {t.get("garage"): t for t in targets or []}
    contexts = {g: runtime_contexts(t, state, (health.get(g) or {}).get("models")) for g, t in by_garage.items()}
    rows, probes = [], 0
    for dep in json.loads(body).get("data", []):
        mi, lp = dep.get("model_info") or {}, dep.get("litellm_params") or {}
        public = dep.get("model_name", "")
        if public.startswith("probe/") or public.startswith("garage-probe") or (mi.get("garage_tier") == "probe"):
            probes += 1
            continue
        garage = mi.get("garage") or "?"
        runtime_model = str(lp.get("model", "")).split("/", 1)[-1]
        cin = mi.get("input_cost_per_token", lp.get("input_cost_per_token"))
        cout = mi.get("output_cost_per_token", lp.get("output_cost_per_token"))
        served = (health.get(garage) or {}).get("models")
        rctx = (contexts.get(garage) or {}).get(runtime_model)
        max_in, max_out = mi.get("max_input_tokens"), mi.get("max_output_tokens")
        checks = []
        if not max_in:
            checks.append(("warning", "no context window: the gateway cannot reject over-long prompts"))
        elif rctx and max_in > rctx:
            checks.append(("warning", f"gateway allows {int(max_in):,} tokens, the runtime only {rctx:,}"))
        elif rctx and max_in < rctx * 0.5:
            checks.append(("info", f"gateway limits to {int(max_in):,} tokens, the runtime has {rctx:,}"))
        if not max_out:
            checks.append(("info", "no output cap"))
        if served is not None and runtime_model not in served:
            checks.append(("warning", f"routed to {runtime_model}, which the runtime does not serve now"))
        if not cin or not cout:
            checks.append(("warning", "no price: requests are free"))
        rows.append({"public": public, "garage": garage, "tier": mi.get("garage_tier") or str(mi.get("id", "")).rsplit("__", 1)[-1],
                     "runtime_model": runtime_model, "context": int(max_in) if max_in else None, "runtime_context": rctx,
                     "max_output": int(max_out) if max_out else None,
                     "price_in": round(cin * 1e6, 4) if cin else None, "price_out": round(cout * 1e6, 4) if cout else None,
                     "up": bool((health.get(garage) or {}).get("runtime_ok")),
                     "checks": [{"level": l, "text": x} for l, x in checks]})
    rows.sort(key=lambda r: (r["public"], r["tier"], r["garage"]))
    return {"deployments": rows, "public_models": len({r["public"] for r in rows}), "probes": probes}


def quality():
    """Per garage: last 24 h and each of the last 7 days. Success rate, time to first token and
    output speed (tokens per second after the first token) from LiteLLM's spend logs."""
    base = """
      with s as (
        select coalesce(nullif(split_part(model_id, '__', 1), ''), '?') g, "startTime" st, status,
               -- LiteLLM sets completionStartTime = endTime when a reply is not streamed: no first token to time
               case when "completionStartTime" < "endTime" then extract(epoch from ("completionStartTime" - "startTime")) end ttft,
               case when completion_tokens >= 20 and "endTime" > "completionStartTime"
                    then completion_tokens / extract(epoch from ("endTime" - "completionStartTime")) end tps
        from "LiteLLM_SpendLogs"
        where "startTime" > now() - interval '{iv}' and model_group not like 'probe/%' and model_group not like 'garage-probe%'
          and call_type not like '/%')
      select g, {bucket}, count(*), count(*) filter (where status = 'success'),
             round((percentile_cont(0.5) within group (order by ttft) filter (where status = 'success'))::numeric, 2),
             round((percentile_cont(0.95) within group (order by ttft) filter (where status = 'success'))::numeric, 2),
             round((percentile_cont(0.5) within group (order by tps) filter (where status = 'success'))::numeric, 1)
      from s where g <> '?' group by 1, 2 order by 2"""
    num = lambda x: float(x) if x not in (None, "") else None
    def row(n, ok, p50, p95, tps):
        n, ok = int(n), int(ok)
        return {"requests": n, "success_pct": round(100 * ok / n, 1) if n else None,
                "ttft_p50": num(p50), "ttft_p95": num(p95), "tps_p50": num(tps)}
    out = {}
    for g, day, *vals in psql(base.format(iv="7 days", bucket="to_char(date_trunc('day', st), 'YYYY-MM-DD')")):
        out.setdefault(g, {"days": [], "last_24h": None})["days"].append({"day": day, **row(*vals)})
    for g, _, *vals in psql(base.format(iv="24 hours", bucket="'24h'")):
        out.setdefault(g, {"days": [], "last_24h": None})["last_24h"] = row(*vals)
    return out

# ---------------------------------------------------------------- history and events
def db_open():
    db = sqlite3.connect(HISTORY_DB, timeout=10)
    db.executescript("""
      create table if not exists host_samples (ts integer primary key, mem_pct real, disk_pct real, load1 real,
                                               crit integer, warn integer);
      create table if not exists garage_samples (ts integer, garage text, up integer, running integer, waiting integer,
                                                 kv_pct real, gen_tok_s real, primary key (ts, garage));
      create table if not exists events (id integer primary key autoincrement, ts integer, level text, kind text,
                                         page text, text text);
      create index if not exists events_ts on events (ts);
    """)
    return db


def add_event(db, level, kind, page, text, ts=None):
    db.execute("insert into events (ts, level, kind, page, text) values (?, ?, ?, ?, ?)",
               (int(ts or time.time()), level, kind, page, text))


def record_samples(db, d, state):
    ts = int(time.time() // 60 * 60)
    h = d["host"]
    mem = round(100 * (1 - h["mem_available_mb"] / h["mem_total_mb"]), 1) if h.get("mem_total_mb") else None
    db.execute("insert or replace into host_samples values (?, ?, ?, ?, ?, ?)",
               (ts, mem, h["disk_used_pct"], h["load"][0],
                sum(a["level"] == "critical" for a in d["alerts"]), sum(a["level"] == "warning" for a in d["alerts"])))
    for g in d["garages"]:
        if g.get("onboarding"):
            continue
        v = g.get("vllm") or {}
        db.execute("insert or replace into garage_samples values (?, ?, ?, ?, ?, ?, ?)",
                   (ts, g["garage"], 1 if g["runtime_ok"] and g["mesh"] is not False else 0,
                    v.get("running"), v.get("waiting"), v.get("kv_cache_pct"), v.get("gen_tok_s")))
    if not state.get("history_backfilled"):
        # The first 24 h of uptime were kept in state.json before the database existed.
        for name, series in state.get("history", {}).items():
            db.executemany("insert or ignore into garage_samples (ts, garage, up) values (?, ?, ?)",
                           [(m * 60, name, up) for m, up in series])
        state["history_backfilled"] = True
    old = time.time() - SAMPLE_DAYS * 86400
    db.execute("delete from host_samples where ts < ?", (old,))
    db.execute("delete from garage_samples where ts < ?", (old,))
    db.execute("delete from events where ts < ?", (time.time() - EVENT_DAYS * 86400,))


def container_starts():
    ids = sh(["docker", "ps", "-aq"]).split()
    if not ids:
        return {}
    out = sh(["docker", "inspect", "--format", "{{.Name}} {{.State.StartedAt}}", *ids])
    return {n.lstrip("/"): t for n, _, t in (line.partition(" ") for line in out.splitlines()) if n}


def duration(seconds):
    m = int(seconds // 60)
    return f"{m} min" if m < 120 else f"{m // 60} h {m % 60} min" if m < 2880 else f"{m // 1440} d {m // 60 % 24} h"


def detect_events(db, d, state):
    """Compare this run with the previous one and log what changed."""
    t = time.time()
    boot = int(t - float(open("/proc/uptime").read().split()[0]))
    cur = {
        "alerts": {alert_key(a): [a["level"], a["text"], a.get("page")] for a in d["alerts"]},
        "boot": boot,
        "containers": container_starts(),
        "versions": {v["component"]: v["running"] for v in d.get("versions", []) if v.get("running")},
        "files": {f: sha256_file(f) for f in DEPLOYED},
        "main": (d.get("source") or {}).get("main"),
        "peers": sorted(p["name"] for p in d["mesh"].get("peers", [])),
        "garages": sorted(g["garage"] for g in d["garages"]),
        "setup_keys": d["mesh"].get("valid_setup_keys"),
        "blocked": {b["ip"]: b["since"] for b in (d.get("guard") or {}).get("active", [])},
    }
    prev = state.get("events_prev")
    since = state.setdefault("alert_since", {})
    state["events_prev"] = cur
    if prev is None:
        since.update({k: t for k in cur["alerts"]})
        return
    for k, (level, text, page) in cur["alerts"].items():
        if k not in prev["alerts"]:
            since[k] = t
            add_event(db, level, "alert", page, text)
        elif prev["alerts"][k][0] != level:
            add_event(db, level, "alert", page, f"{text} (was {prev['alerts'][k][0]})")
    for k, (level, text, page) in prev["alerts"].items():
        if k not in cur["alerts"]:
            started = since.pop(k, None)
            add_event(db, "ok", "resolved", page, f"Resolved: {text}" + (f" (after {duration(t - started)})" if started else ""))
    if prev.get("boot") and abs(boot - prev["boot"]) > 120:
        add_event(db, "info", "reboot", "server", "Gateway rebooted", ts=boot)
    for name, started in cur["containers"].items():
        before = prev["containers"].get(name)
        if before is None:
            add_event(db, "info", "restart", "server", f"Container {name} created")
        elif started != before and t - boot > 300:   # after a reboot every container restarts: one event is enough
            add_event(db, "info", "restart", "server", f"Container {name} restarted")
    for name in set(prev["containers"]) - set(cur["containers"]):
        add_event(db, "info", "restart", "server", f"Container {name} removed")
    for comp, running in cur["versions"].items():
        before = prev["versions"].get(comp)
        if before and before != running:
            add_event(db, "info", "deploy", "updates", f"{comp}: {before} -> {running}")
    changed = [f for f, h in cur["files"].items() if h and prev["files"].get(f) and h != prev["files"][f]]
    if changed:
        add_event(db, "info", "deploy", "updates", "Deployed on the gateway: " + ", ".join(changed))
    if cur["main"] and prev.get("main") and cur["main"] != prev["main"]:
        msg = ((github_json(f"repos/{REPO}/commits/{cur['main']}") or {}).get("commit") or {}).get("message", "")
        add_event(db, "info", "source", "updates", f"main -> {cur['main'][:7]}: {msg.splitlines()[0] if msg else ''}".rstrip(": "))
    for name in sorted(set(cur["peers"]) - set(prev["peers"])):
        add_event(db, "warning", "mesh", "mesh", f"New NetBird peer: {name}")
    for name in sorted(set(prev["peers"]) - set(cur["peers"])):
        add_event(db, "info", "mesh", "mesh", f"NetBird peer removed: {name}")
    for name in sorted(set(cur["garages"]) - set(prev["garages"])):
        add_event(db, "info", "supply", "supply", f"Added to supply: {name}")
    for name in sorted(set(prev["garages"]) - set(cur["garages"])):
        add_event(db, "info", "supply", "supply", f"Removed from supply: {name}")
    for ip, since in cur["blocked"].items():
        if prev.get("blocked", {}).get(ip) != since:
            b = next(x for x in d["guard"]["active"] if x["ip"] == ip)
            add_event(db, "warning", "guard", "traffic", f"Blocked {ip} for {'7 days (repeated)' if b.get('repeat') else '24 h'}: "
                                                         f"{b['count']} {b['reason']} within 2 min")
    keys, before = cur["setup_keys"], prev.get("setup_keys")
    if keys is not None and before is not None and keys > before:
        add_event(db, "warning", "mesh", "mesh", f"NetBird setup key created ({keys} valid)")


def recent_events(db, limit, since=0):
    rows = db.execute("select ts, level, kind, page, text from events where ts >= ? order by ts desc, id desc limit ?",
                      (since, limit)).fetchall()
    return [{"ts": datetime.fromtimestamp(r[0], timezone.utc).isoformat(timespec="seconds"),
             "level": r[1], "kind": r[2], "page": r[3], "text": r[4]} for r in rows]


def traffic_hourly():
    probes = "model_group not like 'probe/%' and model_group not like 'garage-probe%'"
    by_garage = psql(f"""
      select extract(epoch from date_trunc('hour', "startTime"))::bigint, coalesce(nullif(split_part(model_id, '__', 1), ''), '?'),
             count(*), count(*) filter (where status <> 'success'),
             coalesce(sum(prompt_tokens), 0), coalesce(sum(completion_tokens), 0), round(coalesce(sum(spend), 0)::numeric, 4)
      from "LiteLLM_SpendLogs" where "startTime" > now() - interval '7 days' and {probes} group by 1, 2""")
    latency = psql(f"""
      select extract(epoch from date_trunc('hour', "startTime"))::bigint,
             round((percentile_cont(0.5) within group (order by extract(epoch from ("completionStartTime" - "startTime"))) filter (where "completionStartTime" < "endTime"))::numeric, 2),
             round((percentile_cont(0.95) within group (order by extract(epoch from ("completionStartTime" - "startTime"))) filter (where "completionStartTime" < "endTime"))::numeric, 2)
      from "LiteLLM_SpendLogs" where "startTime" > now() - interval '7 days' and {probes} and status = 'success' group by 1""")
    hour = int(time.time() // 3600 * 3600)
    rows = {h: {"t": h, "requests": 0, "failed": 0, "tok_in": 0, "tok_out": 0, "spend": 0.0, "p50": None, "p95": None, "garages": {}}
            for h in range(hour - 167 * 3600, hour + 1, 3600)}
    for h, g, n, f, ti, to, sp in by_garage:
        r = rows.get(int(h))
        if r:
            r["requests"] += int(n); r["failed"] += int(f); r["tok_in"] += int(ti); r["tok_out"] += int(to)
            r["spend"] = round(r["spend"] + float(sp), 4); r["garages"][g] = r["garages"].get(g, 0) + int(n)
    for h, p50, p95 in latency:
        r = rows.get(int(h))
        if r:
            r["p50"] = float(p50) if p50 else None; r["p95"] = float(p95) if p95 else None
    return list(rows.values())


def write_history(db, state):
    t = time.time()
    path = os.path.join(OUT_DIR, "history.json")
    if os.path.exists(path) and t - state.get("history_written", 0) < HISTORY_EVERY:
        return
    week, month = t - 7 * 86400, t - 30 * 86400
    host = db.execute("""select ts / 900 * 900, round(avg(mem_pct), 1), round(avg(disk_pct), 1), round(avg(load1), 2)
                         from host_samples where ts >= ? group by 1 order by 1""", (week,)).fetchall()
    uptime, daily, load = {}, {}, {}
    for g, b, pct, n in db.execute("""select garage, ts / 3600 * 3600, round(avg(up) * 100, 1), count(*)
                                      from garage_samples where ts >= ? group by 1, 2 order by 2""", (week,)):
        uptime.setdefault(g, []).append([b, pct, n])
    for g, b, pct in db.execute("""select garage, ts / 86400 * 86400, round(avg(up) * 100, 2)
                                   from garage_samples where ts >= ? group by 1, 2 order by 2""", (month,)):
        daily.setdefault(g, []).append([b, pct])
    for g, b, run, wait, kv, gen in db.execute("""select garage, ts / 900 * 900, round(avg(running), 2), max(waiting),
                                                 round(avg(kv_pct), 1), round(avg(gen_tok_s), 1)
                                                 from garage_samples where ts >= ? and running is not null group by 1, 2 order by 2""", (week,)):
        load.setdefault(g, []).append([b, run, wait, kv, gen])
    q = quality()
    state["quality"] = q
    out = {"generated_at": now().isoformat(timespec="seconds"), "host": host, "uptime_hourly": uptime,
           "uptime_daily": daily, "garage_load": load, "traffic_hourly": traffic_hourly(),
           "events": recent_events(db, 1000, since=month), "quality": q}
    tmp = path + ".tmp"
    with open(tmp, "w") as f:
        json.dump(out, f, separators=(",", ":"))
    os.chmod(tmp, 0o644)
    os.replace(tmp, path)
    state["history_written"] = t


def history(d, state):
    """Samples, events and history.json. A failure here must never stop ops.json."""
    try:
        db = db_open()
        try:
            record_samples(db, d, state)
            detect_events(db, d, state)
            db.commit()
            d["events_recent"] = recent_events(db, 15)
            write_history(db, state)
        finally:
            db.close()
    except Exception as e:  # noqa: BLE001
        d["alerts"].append({"level": "warning", "page": "server", "text": f"History and event log failed: {type(e).__name__}: {e}"})


# ---------------------------------------------------------------- Telegram notifications
def telegram(cfg, text):
    """Send one HTML message; returns None on success, else a short error (never the token)."""
    body = json.dumps({"chat_id": cfg["TELEGRAM_CHAT_ID"], "text": text, "parse_mode": "HTML",
                       "disable_web_page_preview": True}).encode()
    req = urllib.request.Request(f"https://api.telegram.org/bot{cfg['TELEGRAM_BOT_TOKEN']}/sendMessage", data=body,
                                 headers={"Content-Type": "application/json", "User-Agent": UA})
    try:
        with urllib.request.urlopen(req, timeout=10) as r:
            return None if r.status == 200 else f"HTTP {r.status}"
    except urllib.error.HTTPError as e:
        return f"HTTP {e.code}"
    except Exception as e:  # noqa: BLE001
        return type(e).__name__


def html(s):
    return str(s).replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")


def alert_key(a):
    # Numbers change from run to run ("backup is 27.3 h old"); the alert is the same.
    return a.get("page", "") + ":" + re.sub(r"\d+(?:\.\d+)?", "#", a["text"])


ICON = {"critical": "\U0001F534", "warning": "\U0001F7E0", "resolved": "\u2705", "reminder": "\U0001F501"}


def daily_report(d):
    crit = sum(a["level"] == "critical" for a in d["alerts"])
    warn = sum(a["level"] == "warning" for a in d["alerts"])
    g = d["garages"]
    down = [x["garage"] for x in g if not x["runtime_ok"] or x["mesh"] is False]
    t = d["traffic"]["last_24h"]
    req = sum(r["requests"] for r in t)
    failed = sum(r["failed"] for r in t)
    tokens = sum(r["prompt_tokens"] + r["completion_tokens"] for r in t)
    spend = sum(r["spend_usd"] or 0 for r in t)
    p95 = max([r["ttft_p95_s"] or 0 for r in t] or [0])
    updates = [f"{v['component']} {v['latest']}" for v in d.get("versions", []) if v.get("update")]
    status = (f"{ICON['critical']} {crit} critical, {warn} warning" if crit else
              f"{ICON['warning']} {warn} warning" if warn else f"{ICON['resolved']} All systems operational")
    lines = [f"<b>GarageAI morning report</b>", status, "",
             f"Supply: {len(g) - len(down)}/{len(g)} online" + (f" (down: {html(', '.join(down))})" if down else ""),
             f"Last 24 h: {req:,} requests, {failed} failed, {tokens / 1e6:.1f}M tokens, ${spend:.2f}",
             f"First token p95 (worst model): {p95} s",
             f"Backup: {d['backup'].get('age_hours', '?')} h old · disk {d['host']['disk_used_pct']} %"]
    blocked = [b for b in (d.get("guard") or {}).get("recent", [])
               if datetime.fromisoformat(b["since"]) > now() - timedelta(hours=24)]
    if blocked:
        lines.append(f"Blocked clients, last 24 h: {len(blocked)} ({html(', '.join(b['ip'] for b in blocked[:5]))})")
    if updates:
        lines.append(f"Updates available: {html(', '.join(updates))}")
    lines.append(f'<a href="{OPS_URL}/">Operations Center</a>')
    return "\n".join(lines)


def notify(d, state):
    """Send new, escalated, repeated and resolved alerts, and the morning report. A failed send is
    retried on the next run (the alert is only marked as sent when Telegram accepted it)."""
    cfg = load_env(TELEGRAM_ENV)
    configured = bool(cfg.get("TELEGRAM_BOT_TOKEN") and cfg.get("TELEGRAM_CHAT_ID"))
    ns = state.setdefault("notify", {})
    open_ = ns.setdefault("open", {})
    t = time.time()
    current = {alert_key(a): a for a in d["alerts"] if a["level"] in NOTIFY_AFTER}
    for k, a in current.items():
        e = open_.setdefault(k, {"seen": 0, "notified": None, "notified_level": None})
        e.update(level=a["level"], text=a["text"], page=a.get("page"), seen=e["seen"] + 1)
    items, sent_keys, resolved = [], [], []
    for k, e in list(open_.items()):
        if k not in current:
            if e.get("notified"):
                items.append(("resolved", e)); resolved.append(k)
            else:
                del open_[k]
            continue
        rank = {"warning": 1, "critical": 2}
        if e["seen"] >= NOTIFY_AFTER[e["level"]] and (not e["notified"] or rank[e["level"]] > rank.get(e["notified_level"], 0)):
            items.append((e["level"], e)); sent_keys.append(k)
        elif e["notified"] and e["level"] == "critical" and t - e["notified"] >= REMIND_MINUTES * 60:
            items.append(("reminder", e)); sent_keys.append(k)

    error = None
    if configured and items:
        order = {"critical": 0, "reminder": 1, "warning": 2, "resolved": 3}
        lines = []
        for kind, e in sorted(items, key=lambda x: order[x[0]]):
            label = {"critical": "Critical", "warning": "Warning", "reminder": "Still open", "resolved": "Resolved"}[kind]
            lines.append(f'{ICON[kind]} <b>{label}</b> · <a href="{OPS_URL}/#{e.get("page") or "dashboard"}">{html((e.get("page") or "").title())}</a>\n{html(e["text"])}')
        error = telegram(cfg, "\n\n".join(lines))
        if not error:
            for k in sent_keys:
                open_[k].update(notified=t, notified_level=open_[k]["level"])
            for k in resolved:
                open_.pop(k, None)
            ns["last_sent"] = now().isoformat(timespec="seconds")

    hour, zone = DAILY_REPORT
    local = datetime.now(ZoneInfo(zone))
    if configured and local.hour >= hour and ns.get("daily") != local.date().isoformat():
        err = telegram(cfg, daily_report(d))
        if not err:
            ns["daily"] = local.date().isoformat()
            ns["last_sent"] = now().isoformat(timespec="seconds")
        error = error or err
    if error:
        ns["last_error"] = {"at": now().isoformat(timespec="seconds"), "error": error}
    elif items or not configured:
        ns.pop("last_error", None)

    d["notify"] = {"channel": "telegram", "configured": configured, "last_sent": ns.get("last_sent"),
                   "last_error": ns.get("last_error"), "sent_open": sum(1 for e in open_.values() if e.get("notified")),
                   "rules": f"critical after {NOTIFY_AFTER['critical']} min, warning after {NOTIFY_AFTER['warning']} min, "
                            f"critical repeated every {REMIND_MINUTES // 60} h, resolved when cleared",
                   "daily_report": f"{hour:02d}:00 {zone}"}
    if not configured:
        d["alerts"].append({"level": "info", "text": "Telegram notifications are not configured", "page": "server"})
    elif ns.get("last_error"):
        d["alerts"].append({"level": "warning", "page": "server",
                            "text": f"Telegram notifications failing since {ns['last_error']['at']} ({ns['last_error']['error']})"})
    d["alerts"].sort(key=lambda a: {"critical": 0, "warning": 1, "info": 2}[a["level"]])


def telegram_cli(arg):
    cfg = load_env(TELEGRAM_ENV)
    if not cfg.get("TELEGRAM_BOT_TOKEN"):
        sys.exit(f"No TELEGRAM_BOT_TOKEN in {TELEGRAM_ENV}")
    if arg == "--telegram-setup":
        code, body, _ = http(f"https://api.telegram.org/bot{cfg['TELEGRAM_BOT_TOKEN']}/getUpdates", timeout=15)
        if code != 200:
            sys.exit(f"Telegram getUpdates answered {code}: check the bot token")
        chats = {}
        for u in json.loads(body).get("result", []):
            c = (u.get("message") or {}).get("chat") or {}
            if c.get("type") == "private":
                chats[c["id"]] = c.get("username") or c.get("first_name") or "?"
        if len(chats) != 1:
            sys.exit(f"Expected exactly one private chat with the bot, found {len(chats)}: "
                     "send /start to the bot from your own Telegram account and run this again")
        chat_id, who = next(iter(chats.items()))
        lines = [l for l in open(TELEGRAM_ENV).read().splitlines() if not l.startswith("TELEGRAM_CHAT_ID=")]
        with open(TELEGRAM_ENV, "w") as f:
            f.write("\n".join(lines + [f"TELEGRAM_CHAT_ID={chat_id}"]) + "\n")
        os.chmod(TELEGRAM_ENV, 0o600)
        cfg["TELEGRAM_CHAT_ID"] = str(chat_id)
        print(f"Saved the chat with {who}.")
    if not cfg.get("TELEGRAM_CHAT_ID"):
        sys.exit("No TELEGRAM_CHAT_ID yet: run --telegram-setup")
    err = telegram(cfg, f'{ICON["resolved"]} <b>GarageAI Operations Center</b> is connected. Alerts will arrive here.\n<a href="{OPS_URL}/">Open</a>')
    sys.exit(f"Sending failed: {err}" if err else 0)


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
        "guard": guard(),
    }
    data["models"] = models(env, state, targets)
    data["quality_24h"] = {g: q["last_24h"] for g, q in (state.get("quality") or {}).items() if q.get("last_24h")}
    data["versions"] = versions(env, state, data["containers"])
    data["source"] = source(state)
    data["logs"] = logs(data["containers"])
    data["alerts"] = alerts(data)
    notify(data, state)
    history(data, state)
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
    if len(sys.argv) > 1 and sys.argv[1] in ("--telegram-setup", "--telegram-test"):
        telegram_cli(sys.argv[1])
    elif len(sys.argv) == 3 and sys.argv[1] == "--note":
        with db_open() as db:
            add_event(db, "info", "note", "server", sys.argv[2])
        print("Note added.")
    else:
        main()

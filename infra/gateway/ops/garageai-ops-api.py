#!/usr/bin/env python3
"""GarageAI Operations Center API: the one thing the page may change, its own silences.

Listens on 127.0.0.1:8781 and on the Docker bridge (docker0) on the host; nginx in the ops
container proxies /api/ to it, behind Traefik's basic auth, which also passes the signed-in user in
X-Ops-User. nginx adds a shared key (X-Ops-Api-Key, from /etc/garageai/ops-api.key via the
container's OPS_API_KEY), so no other container can call it with a made-up user. It writes /var/lib/garageai-ops/
silences.json, the file garageai-ops-collect --silence writes too; the collector applies it on
its next run (within a minute).

  GET  /api/silences                    {garage: {until, reason, by}}
  POST /api/silence   {garage, hours, reason}
  POST /api/unsilence {garage}
"""
import json
import os
import re
import subprocess
import sys
import threading
from datetime import datetime, timedelta, timezone
from http.server import BaseHTTPRequestHandler, HTTPServer

SILENCES = "/var/lib/garageai-ops/silences.json"
KEY_FILE = "/etc/garageai/ops-api.key"
OPS = "/var/lib/garageai-ops/www/ops.json"
NAME = re.compile(r"^[a-z0-9][a-z0-9-]{1,40}$")


def read_silences():
    try:
        with open(SILENCES) as f:
            v = json.load(f)
        return v if isinstance(v, dict) else {}
    except (OSError, ValueError):
        return {}


def write_silences(sil):
    tmp = SILENCES + ".tmp"
    with open(tmp, "w") as f:
        json.dump(sil, f, indent=1)
    os.chmod(tmp, 0o644)
    os.replace(tmp, SILENCES)


def known_names():
    try:
        with open(OPS) as f:
            d = json.load(f)
        return {g["garage"] for g in d.get("garages", [])} | set((d.get("silences") or {}).keys())
    except (OSError, ValueError, KeyError):
        return set()


class H(BaseHTTPRequestHandler):
    def reply(self, code, body):
        data = json.dumps(body).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        self.wfile.write(data)

    def user(self):
        if self.headers.get("X-Ops-Api-Key", "") != API_KEY:
            return None
        u = self.headers.get("X-Ops-User", "").strip()
        return u if re.fullmatch(r"[\w.@-]{1,64}", u) else None

    def do_GET(self):
        if self.path != "/api/silences":
            return self.reply(404, {"error": "not found"})
        if not self.user():
            return self.reply(403, {"error": "no signed-in user"})
        self.reply(200, read_silences())

    def do_POST(self):
        user = self.user()
        if not user:
            return self.reply(403, {"error": "no signed-in user"})
        try:
            body = json.loads(self.rfile.read(int(self.headers.get("Content-Length") or 0) or b"{}"))
        except ValueError:
            return self.reply(400, {"error": "invalid JSON"})
        garage = str(body.get("garage", "")).strip()
        if not NAME.match(garage) or garage not in known_names():
            return self.reply(400, {"error": "unknown garage"})
        sil = read_silences()
        if self.path == "/api/unsilence":
            sil.pop(garage, None)
            write_silences(sil)
            print(f"api: {user} ended the silence of {garage}", flush=True)
            return self.reply(200, {"ok": True, "silences": sil})
        if self.path != "/api/silence":
            return self.reply(404, {"error": "not found"})
        try:
            hours = float(body.get("hours", 0))
        except (TypeError, ValueError):
            hours = 0
        if not 0.25 <= hours <= 336:
            return self.reply(400, {"error": "hours must be between 0.25 and 336"})
        reason = re.sub(r"\s+", " ", str(body.get("reason", ""))).strip()[:200] or "planned stop"
        until = (datetime.now(timezone.utc) + timedelta(hours=hours)).isoformat(timespec="seconds")
        sil[garage] = {"until": until, "reason": f"{reason} ({user})", "by": user}
        write_silences(sil)
        print(f"api: {user} silenced {garage} for {hours} h: {reason}", flush=True)
        self.reply(200, {"ok": True, "silences": sil})

    def log_message(self, *a):
        pass


def docker_bridge_ip():
    out = subprocess.run(["ip", "-4", "-o", "addr", "show", "docker0"], capture_output=True, text=True).stdout
    m = re.search(r"inet (\d+\.\d+\.\d+\.\d+)", out)
    return m.group(1) if m else None


if __name__ == "__main__":
    API_KEY = open(KEY_FILE).read().strip()
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8781
    addrs = ["127.0.0.1"] + ([docker_bridge_ip()] if docker_bridge_ip() else [])
    servers = [HTTPServer((a, port), H) for a in addrs]
    for srv in servers[1:]:
        threading.Thread(target=srv.serve_forever, daemon=True).start()
    print(f"api: listening on {', '.join(f'{a}:{port}' for a in addrs)}", flush=True)
    servers[0].serve_forever()

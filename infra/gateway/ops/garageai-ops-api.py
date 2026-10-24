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

Sign-in (2026-10-24, replaces Traefik's basic auth, which browsers forget on every restart):
  GET  /api/auth   nginx's auth_request: 200 + X-Ops-User when the ops_session cookie is valid, else 401
  GET  /login      the sign-in form (a real form, so password managers fill it in)
  POST /login      checks the password (/etc/garageai/ops-password), sets ops_session for 30 days
  POST /logout     clears it
The cookie is user.expiry.HMAC-SHA256 with /etc/garageai/ops-session.key (root, 0600). Five wrong
passwords from one address lock that address out for 15 minutes.
"""
import hashlib
import hmac
import html
import json
import os
import re
import time
import urllib.parse
import subprocess
import sys
import threading
from datetime import datetime, timedelta, timezone
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer as HTTPServer

SILENCES = "/var/lib/garageai-ops/silences.json"
KEY_FILE = "/etc/garageai/ops-api.key"
OPS = "/var/lib/garageai-ops/www/ops.json"
NAME = re.compile(r"^[a-z0-9][a-z0-9-]{1,40}$")
PASSWORD_FILE = "/etc/garageai/ops-password"
SESSION_KEY_FILE = "/etc/garageai/ops-session.key"
SESSION_DAYS = 30
USER = "garageai"
_failures = {}  # address -> [times of wrong passwords]
_lock = threading.Lock()


def session_cookie(user, key, now=None):
    exp = int((now or time.time()) + SESSION_DAYS * 86400)
    msg = f"{user}.{exp}"
    sig = hmac.new(key, msg.encode(), hashlib.sha256).hexdigest()
    return f"{msg}.{sig}"


def session_user(cookie_header, key, now=None):
    """The signed-in user from the ops_session cookie, or None."""
    for part in (cookie_header or "").split(";"):
        name, _, value = part.strip().partition("=")
        if name != "ops_session":
            continue
        try:
            user, exp, sig = value.rsplit(".", 2)
            want = hmac.new(key, f"{user}.{exp}".encode(), hashlib.sha256).hexdigest()
            if hmac.compare_digest(sig, want) and int(exp) > (now or time.time()) and re.fullmatch(r"[\w.@-]{1,64}", user):
                return user
        except ValueError:
            pass
    return None


def locked_out(addr, now=None):
    now = now or time.time()
    with _lock:
        recent = [t for t in _failures.get(addr, []) if now - t < 900]
        _failures[addr] = recent
        return len(recent) >= 5


def note_failure(addr, now=None):
    with _lock:
        _failures.setdefault(addr, []).append(now or time.time())


LOGIN_PAGE = """<!doctype html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1"><title>GarageAI Operations Center</title>
<style>:root{color-scheme:light dark}body{font:16px -apple-system,system-ui,sans-serif;display:grid;place-items:center;min-height:100vh;margin:0;background:Canvas;color:CanvasText}
form{display:grid;gap:12px;width:min(320px,90vw)}h1{font-size:20px;margin:0 0 8px}input,button{font:inherit;padding:10px 12px;border-radius:8px;border:1px solid #8886}
button{background:#2f6fdf;color:#fff;border:0;cursor:pointer}.err{color:#c43c3c;font-size:14px}</style></head>
<body><form method="post" action="/login"><h1>GarageAI Operations Center</h1>%s
<input name="username" autocomplete="username" value="garageai" aria-label="User">
<input name="password" type="password" autocomplete="current-password" placeholder="Password" aria-label="Password" autofocus required>
<button type="submit">Sign in</button><small>Stays signed in on this device for 30 days.</small></form></body></html>"""


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

    def page(self, code, body, headers=()):
        data = body.encode()
        self.send_response(code)
        for k, v in headers:
            self.send_header(k, v)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(data)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        self.wfile.write(data)

    def client(self):
        # Traefik's X-Forwarded-For, as nginx passes it on: the first address is the browser.
        return (self.headers.get("X-Forwarded-For", "") or "unknown").split(",")[0].strip()

    def keyed(self):
        return hmac.compare_digest(self.headers.get("X-Ops-Api-Key", ""), API_KEY)

    def do_GET(self):
        if self.path == "/api/auth":
            if not self.keyed():
                return self.reply(403, {"error": "forbidden"})
            user = session_user(self.headers.get("Cookie"), SESSION_KEY)
            if not user:
                return self.reply(401, {"error": "not signed in"})
            self.send_response(200)
            self.send_header("X-Ops-User", user)
            self.send_header("Content-Length", "0")
            self.end_headers()
            return
        if self.path == "/api/whoami":
            return self.reply(200, {"user": self.user() or ""})
        if self.path.split("?")[0] == "/login":
            return self.page(200, LOGIN_PAGE % "")
        if self.path != "/api/silences":
            return self.reply(404, {"error": "not found"})
        if not self.user():
            return self.reply(403, {"error": "no signed-in user"})
        self.reply(200, read_silences())

    def do_POST(self):
        if self.path in ("/login", "/logout") and not self.keyed():
            return self.reply(403, {"error": "forbidden"})
        if self.path == "/logout":
            return self.page(303, "", [("Location", "/login"), ("Set-Cookie", "ops_session=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Strict")])
        if self.path == "/login":
            addr = self.client()
            if locked_out(addr):
                print(f"api: sign-in refused for {addr}: too many wrong passwords", flush=True)
                return self.page(429, LOGIN_PAGE % '<p class="err">Too many wrong passwords. Try again in 15 minutes.</p>')
            form = urllib.parse.parse_qs(self.rfile.read(min(int(self.headers.get("Content-Length") or 0), 4096)).decode(errors="replace"))
            given = (form.get("password") or [""])[0]
            if not given or not hmac.compare_digest(given.encode(), PASSWORD.encode()):
                note_failure(addr)
                print(f"api: wrong password from {addr}", flush=True)
                return self.page(401, LOGIN_PAGE % '<p class="err">Wrong password.</p>')
            print(f"api: {USER} signed in from {addr}", flush=True)
            cookie = session_cookie(USER, SESSION_KEY)
            return self.page(303, "", [("Location", "/"), ("Set-Cookie", f"ops_session={cookie}; Path=/; Max-Age={SESSION_DAYS * 86400}; HttpOnly; Secure; SameSite=Strict")])
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
    PASSWORD = open(PASSWORD_FILE).read().strip()
    SESSION_KEY = open(SESSION_KEY_FILE).read().strip().encode()
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8781
    addrs = ["127.0.0.1"] + ([docker_bridge_ip()] if docker_bridge_ip() else [])
    servers = [HTTPServer((a, port), H) for a in addrs]
    for srv in servers[1:]:
        threading.Thread(target=srv.serve_forever, daemon=True).start()
    print(f"api: listening on {', '.join(f'{a}:{port}' for a in addrs)}", flush=True)
    servers[0].serve_forever()

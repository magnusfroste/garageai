#!/usr/bin/env python3
"""GarageAI Operations Center: the managed-models chain, tested end to end every night.

The chain (docs/bridge.md "Models chosen in the portal, fetched by Bridge"): a wanted model in the
portal → Bridge fetches it in Ollama → the heartbeat reports it → the portal offers and tests it →
the gateway routes to it → a buyer gets an answer → the portal asks for its removal → Bridge
removes it → the gateway stops routing. Every step is something we have seen fail once; this
runs the whole chain on one of our own garages with a small model, times each step, and writes
chain-test.json for the Ops page (the collector alerts when the last run failed or is old).

Also a failure path each night: a model name that does not exist must end as `failed` with a
reason, not hang.

Runs as a systemd timer (garageai-chain-test.timer, 04:00 UTC). Portal side: the
managed-models-ops edge function (want / cancel / status, gateway key). Needs the gateway key
(GARAGEAI_GATEWAY_KEY) and portal URL (GARAGEAI_PORTAL_URL) from /etc/garageai/gateway-health.env.
Options: CHAIN_TEST_GARAGE (default garage-m1), CHAIN_TEST_MODEL (default qwen3:0.6b).
"""
import importlib.machinery
import importlib.util
import json
import os
import sys
import time
import urllib.request
from datetime import datetime, timezone

loader = importlib.machinery.SourceFileLoader("collect", os.environ.get("GARAGEAI_OPS_COLLECT", "/usr/local/bin/garageai-ops-collect"))
spec = importlib.util.spec_from_loader("collect", loader)
collect = importlib.util.module_from_spec(spec)
loader.exec_module(collect)

GARAGE = os.environ.get("CHAIN_TEST_GARAGE", "garage-m1")
MODEL = os.environ.get("CHAIN_TEST_MODEL", "qwen3:0.6b")
BAD_MODEL = "garageai-does-not-exist:1b"
OUT = os.path.join(collect.OUT_DIR, "chain-test.json")
HISTORY = os.path.join(os.path.dirname(collect.OUT_DIR), "chain-tests.json")
KEEP = 60
POLL = 30
# Generous: a heartbeat is 5 minutes, a pull a few, the acceptance test a minute, the gateway sync a minute.
LIMITS = {"pull": 20 * 60, "offered": 12 * 60, "routed": 6 * 60, "removed": 15 * 60, "unrouted": 6 * 60, "failed": 15 * 60}


def log(msg):
    print(f"{datetime.now(timezone.utc).strftime('%H:%M:%S')} {msg}", flush=True)


class Portal:
    def __init__(self, env):
        self.url = env.get("GARAGEAI_PORTAL_URL", "").rstrip("/") + "/functions/v1/managed-models-ops"
        self.key = env.get("GARAGEAI_GATEWAY_KEY", "")

    def call(self, body):
        # An edge function that has been idle starts cold (the first want timed out at 20 s on
        # 2026-10-14): a long timeout, and one retry for timeouts and 5xx.
        last = (None, {"error": "no attempt"})
        for attempt in range(2):
            req = urllib.request.Request(self.url, data=json.dumps(body).encode(), method="POST",
                                         headers={"Content-Type": "application/json", "x-gateway-key": self.key, "User-Agent": collect.UA})
            try:
                with urllib.request.urlopen(req, timeout=60) as r:
                    return r.status, json.loads(r.read() or b"{}")
            except urllib.error.HTTPError as e:
                last = (e.code, {"error": e.read().decode(errors="replace")[:300]})
                if e.code < 500:
                    return last
            except Exception as e:  # noqa: BLE001
                last = (None, {"error": str(e)[:300]})
            if attempt == 0:
                time.sleep(5)
        return last

    def want(self, model, action):
        return self.call({"op": "want", "garage": GARAGE, "model": model, "action": action})

    def cancel(self, model):
        return self.call({"op": "cancel", "garage": GARAGE, "model": model})

    def status(self):
        code, body = self.call({"op": "status", "garage": GARAGE})
        return body if code == 200 else {}


def fresh(row):
    """A failed/refused report counts only when Bridge reported it after this request was made:
    Bridge reports its state before it acts on the wish list, so the first heartbeat after a new
    request can still carry a stale verdict from an earlier one (2026-10-14, run 4)."""
    rep, req = row.get("reported_at"), row.get("requested_at")
    return bool(rep and req and str(rep) > str(req))


def managed(status, model):
    for m in status.get("managed", status.get("managed_models", []) or []):
        if m.get("model") == model:
            return m
    return {}


def garage_model(status, model):
    """The portal's garage_models row for the model: {installed, offered, status} (the status op
    nests it under "garage_model"; runs 7 and 8 read the outer entry and never saw "offered")."""
    for m in status.get("models", status.get("garage_models", []) or []):
        if m.get("model") == model:
            return m.get("garage_model") or {}
    return {}


def deployments(env, model):
    """The gateway's deployments for GARAGE serving `model` (by Ollama tag), as [(public name, id)]."""
    code, body, _ = collect.http(f"{collect.LITELLM_URL}/v1/model/info", {"Authorization": f"Bearer {env.get('GARAGEAI_GATEWAY_KEY', '')}"}, timeout=15)
    out = []
    if code != 200:
        return out
    for d in json.loads(body).get("data", []):
        mi, lp = d.get("model_info") or {}, d.get("litellm_params") or {}
        if str(mi.get("id", "")).startswith(GARAGE + "__") and str(lp.get("model", "")).endswith("/" + model):
            out.append((d.get("model_name"), mi.get("id")))
    return out


def ask(env, public_name):
    """One short completion through the public API, pinned to this garage. (status, seconds)."""
    body = json.dumps({"model": public_name, "max_tokens": 8, "messages": [{"role": "user", "content": "Chain test: say ok"}]}).encode()
    req = urllib.request.Request("https://api.garageai.eu/v1/chat/completions", data=body, method="POST",
                                 headers={"Content-Type": "application/json", "Authorization": f"Bearer {env.get('GARAGEAI_GATEWAY_KEY', '')}", "User-Agent": collect.UA})
    t0 = time.time()
    try:
        with urllib.request.urlopen(req, timeout=120) as r:
            r.read()
            return r.status, round(time.time() - t0, 1)
    except urllib.error.HTTPError as e:
        return e.code, round(time.time() - t0, 1)
    except Exception:  # noqa: BLE001
        return None, round(time.time() - t0, 1)


def wait_for(name, check, limit, give_up=None):
    """Polls check() until it returns a truthy detail or the limit passes. give_up() returning a
    reason ends the wait at once as a failure (a pull refused by Bridge will not become ready)."""
    t0 = time.time()
    while True:
        detail = check()
        if detail:
            s = round(time.time() - t0)
            log(f"  ✓ {name} after {s} s: {detail}")
            return {"name": name, "ok": True, "seconds": s, "detail": detail}
        if give_up and (why := give_up()):
            log(f"  ✗ {name}: {why}")
            return {"name": name, "ok": False, "seconds": round(time.time() - t0), "detail": why}
        if time.time() - t0 > limit:
            log(f"  ✗ {name}: not within {limit // 60} min")
            return {"name": name, "ok": False, "seconds": round(time.time() - t0), "detail": f"not within {limit // 60} min"}
        time.sleep(POLL)


def away(env):
    """Why the test garage is off on purpose, or None: paused in the portal (gateway-targets sends
    paused garages with paused: true) or a planned stop silenced in Ops. A paused garage cannot
    fetch a model, so the night would fail and say nothing about the chain (garage-m1 paused and
    asleep, 2026-10-18/19)."""
    until = (collect.read_silences().get(GARAGE) or {}).get("until")
    if until and until > datetime.now(timezone.utc).isoformat():
        return f"planned stop in Ops until {until[:16]}"
    try:
        _, _, targets = collect.portal_targets(env)
    except Exception:  # noqa: BLE001 - without the list, run the test and let it say what it finds
        return None
    t = next((x for x in targets or [] if x.get("garage") == GARAGE), None)
    if t and t.get("paused"):
        return "paused in the portal"
    return None


def run(env):
    portal = Portal(env)
    steps = []
    why = away(env)
    if why:
        log(f"chain test skipped: {GARAGE} is {why}")
        return {"ok": None, "skipped": f"{GARAGE} is {why}", "steps": steps}
    log(f"chain test on {GARAGE} with {MODEL}")
    st = portal.status()
    # A leftover is a model still on the garage (installed) or a request still in flight; a
    # finished history row (removed, failed, refused) is not one (run 6 skipped itself on one).
    if managed(st, MODEL).get("status") in ("wanted", "downloading", "ready") or garage_model(st, MODEL).get("installed"):
        # Left over from a failed night: clean up first, then test from a clean state next night.
        log(f"  {MODEL} is already on {GARAGE} (left over?): asking for removal, test skipped")
        portal.want(MODEL, "remove")
        return {"ok": False, "skipped": "leftover model, removal requested", "steps": steps}

    code, body = portal.want(MODEL, "pull")
    steps.append({"name": "want", "ok": code == 200, "seconds": 0, "detail": body.get("error") or f"row {body.get('status', 'created')}"})
    if code != 200:
        return {"ok": False, "steps": steps, "error": f"want: HTTP {code} {body.get('error', '')}"}

    steps.append(wait_for("pull", lambda: (lambda m: m.get("status") == "ready" and f"ready, {m.get('total_mb') or '?'} MB")(managed(portal.status(), MODEL)), LIMITS["pull"],
                          give_up=lambda: (lambda m: m.get("status") in ("failed", "refused") and fresh(m) and f"{m.get('status')}: {m.get('error') or '-'}")(managed(portal.status(), MODEL))))
    if not steps[-1]["ok"]:
        m = managed(portal.status(), MODEL)
        steps[-1]["detail"] += f" (status {m.get('status')}: {m.get('error') or '-'})"
        portal.cancel(MODEL)
        return {"ok": False, "steps": steps, "error": "pull"}

    # Offered = the portal decided to sell it (any test status; "routed" below checks the gateway).
    # The gateway having a deployment is the portal's decision carried out, so it counts as offered too.
    steps.append(wait_for("offered", lambda: (lambda g: g.get("offered") and f"offered, status {g.get('status')}")(garage_model(portal.status(), MODEL))
                          or (deployments(env, MODEL) and "offered (the gateway has its deployment)"), LIMITS["offered"]))
    if not steps[-1]["ok"]:
        steps[-1]["detail"] += f" (garage_models row: {garage_model(portal.status(), MODEL) or 'none'})"
    steps.append(wait_for("routed", lambda: (lambda d: d and f"{len(d)} deployment(s): " + ", ".join(n for n, _ in d))(deployments(env, MODEL)), LIMITS["routed"]))
    if steps[-1]["ok"]:
        name = next((n for n, _ in deployments(env, MODEL) if n.startswith("garage/")), None) or deployments(env, MODEL)[0][0]
        code, secs = ask(env, name)
        steps.append({"name": "answered", "ok": code == 200, "seconds": secs, "detail": f"{name}: HTTP {code} in {secs} s"})
        log(f"  {'✓' if code == 200 else '✗'} answered: {name} HTTP {code} in {secs} s")

    code, body = portal.want(MODEL, "remove")
    steps.append({"name": "want removal", "ok": code == 200, "seconds": 0, "detail": body.get("error") or "requested"})
    steps.append(wait_for("removed", lambda: (lambda m: m.get("status") == "removed" and "removed")(managed(portal.status(), MODEL)), LIMITS["removed"]))
    steps.append(wait_for("unrouted", lambda: (not deployments(env, MODEL)) and "no deployment left", LIMITS["unrouted"]))

    # The failure path: a name that does not exist must fail with a reason, not hang.
    code, body = portal.want(BAD_MODEL, "pull")
    if code == 200:
        steps.append(wait_for("bad name fails", lambda: (lambda m: m.get("status") in ("failed", "refused") and f"{m.get('status')}: {(m.get('error') or '')[:80]}")(managed(portal.status(), BAD_MODEL)), LIMITS["failed"]))
        portal.cancel(BAD_MODEL)
    else:
        steps.append({"name": "bad name fails", "ok": False, "seconds": 0, "detail": f"want: HTTP {code}"})

    return {"ok": all(s["ok"] for s in steps), "steps": steps}


def main():
    env = collect.load_env()
    t0 = time.time()
    try:
        result = run(env)
    except Exception as e:  # noqa: BLE001
        result = {"ok": False, "steps": [], "error": f"{type(e).__name__}: {e}"}
    result.update({"at": datetime.now(timezone.utc).isoformat(timespec="seconds"), "garage": GARAGE, "model": MODEL,
                   "total_seconds": round(time.time() - t0)})
    history = []
    try:
        history = json.load(open(HISTORY))
    except (OSError, ValueError):
        pass
    history = (history + [result])[-KEEP:]
    for path, data in ((HISTORY, history), (OUT, {"last": result, "runs": len(history), "passed": sum(1 for r in history if r.get("ok")),
                                                   "history": [{"at": r["at"], "ok": r.get("ok"), "skipped": bool(r.get("skipped")), "total_seconds": r.get("total_seconds")} for r in history]})):
        tmp = path + ".tmp"
        with open(tmp, "w") as f:
            json.dump(data, f, separators=(",", ":"))
        os.chmod(tmp, 0o644)
        os.replace(tmp, path)
    if result.get("ok") is None:
        log(f"chain test skipped ({result.get('skipped')})")
        return 0   # a skip is not a failed run: the timer stays green, Ops says how long nothing was tested
    log(f"chain test {'passed' if result['ok'] else 'FAILED'} in {result['total_seconds']} s")
    return 0 if result["ok"] else 1


if __name__ == "__main__":
    sys.exit(main())

#!/usr/bin/env python3
"""GarageAI Operations Center, live: who uses what, right now.

Every 5 seconds it reads the gateway's latest requests (LiteLLM's spend logs) and writes
/var/lib/garageai-ops/www/live.json for the Live page: the last requests, which key went to which
model and garage in the last 15 minutes, how busy each garage is, and today's top keys. Probes are
left out. A key's owner is shown as a masked e-mail (e…n@gmail.com); prompts and replies are
never read.

Runs as a systemd service (garageai-ops-live.service). It reuses the collector's helpers.
"""
import importlib.machinery
import importlib.util
import json
import os
import re
import sys
import time
from datetime import datetime, timezone

loader = importlib.machinery.SourceFileLoader("collect", os.environ.get("GARAGEAI_OPS_COLLECT", "/usr/local/bin/garageai-ops-collect"))
spec = importlib.util.spec_from_loader("collect", loader)
collect = importlib.util.module_from_spec(spec)
loader.exec_module(collect)

EVERY = 5            # seconds between updates
WINDOW = 15          # minutes of flows
RECENT = 40          # requests in the ticker
OWNERS_EVERY = 300   # seconds between re-reading key owners
OUT = os.path.join(collect.OUT_DIR, "live.json")


def mask(email):
    """e…n@gmail.com: enough to recognise someone you know, not to harvest addresses."""
    if not email or "@" not in email:
        return None
    name, domain = email.split("@", 1)
    return (name[0] + "…" + name[-1] if len(name) > 2 else name[0] + "…") + "@" + domain


def owners():
    """{hashed key: {alias, owner}} from LiteLLM's keys and users."""
    rows = collect.psql(
        'select t.token, coalesce(t.key_alias, \'\'), coalesce(u.user_email, \'\'), coalesce(t.user_id, \'\') '
        'from "LiteLLM_VerificationToken" t left join "LiteLLM_UserTable" u on u.user_id = t.user_id')
    out = {}
    for r in rows:
        if len(r) < 4:
            continue
        token, alias, email, uid = r[:4]
        out[token] = {"alias": alias.strip() or None, "owner": mask(email) or ("created in LiteLLM" if not uid or not email else None)}
    return out


def garage_of(model_id, api_base):
    if model_id and model_id.startswith("auto__"):
        model_id = model_id[len("auto__"):]   # auto__<curated model id>: the garage is the next part
    if model_id and "__" in model_id:
        return model_id.split("__", 1)[0]
    m = re.search(r"//([^/:]+)", api_base or "")
    return m.group(1) if m else "?"


def auto_entries(env):
    """{deployment id: {order, garage, model}} for garageai/auto's deployments, from LiteLLM."""
    code, body, _ = collect.http(f"{collect.LITELLM_URL}/v1/model/info", {"Authorization": f"Bearer {env.get('GARAGEAI_GATEWAY_KEY', '')}"}, timeout=15)
    out = {}
    if code != 200:
        return out
    for d in json.loads(body).get("data", []):
        mi, lp = d.get("model_info") or {}, d.get("litellm_params") or {}
        if mi.get("garage_tier") == "auto" or str(mi.get("id", "")).startswith("auto__"):
            out[mi.get("id")] = {"order": mi.get("order", lp.get("order")), "garage": mi.get("garage"), "public": d.get("model_name"),
                                 "model": str(lp.get("model", "")).split("/", 1)[-1]}
    return out


def ms(a, b):
    try:
        return round((datetime.fromisoformat(b) - datetime.fromisoformat(a)).total_seconds() * 1000)
    except (TypeError, ValueError):
        return None


def snapshot(keys, autos=None):
    now = time.time()
    rows = collect.psql(
        'select to_char("startTime" at time zone \'UTC\', \'YYYY-MM-DD"T"HH24:MI:SS.MS\'), '
        'to_char("completionStartTime" at time zone \'UTC\', \'YYYY-MM-DD"T"HH24:MI:SS.MS\'), '
        'to_char("endTime" at time zone \'UTC\', \'YYYY-MM-DD"T"HH24:MI:SS.MS\'), '
        'api_key, coalesce(model_group, \'\'), coalesce(model_id, \'\'), coalesce(api_base, \'\'), '
        'coalesce(prompt_tokens, 0), coalesce(completion_tokens, 0), coalesce(status, \'\'), coalesce(spend, 0), '
        "coalesce(metadata->>'user_api_key_alias', '') "
        'from "LiteLLM_SpendLogs" '
        f'where "startTime" > now() - interval \'{WINDOW} minutes\' and coalesce(model_group, \'\') not like \'probe/%\' '
        'order by "startTime" desc limit 2000')
    reqs = []
    for r in rows:
        if len(r) < 12:
            continue
        start, first, end, key, model, mid, base, pin, pout, status, spend, alias = r[:12]
        k = keys.get(key, {})
        if not k and key.endswith("master_key"):
            k = {"alias": "master key", "owner": "gateway admin / ops"}
        reqs.append({
            "auto_order": (autos or {}).get(mid, {}).get("order"),
            "t": start + "Z", "key": k.get("alias") or alias.strip() or ("key …" + key[-6:] if key else "?"),
            "owner": k.get("owner"), "model": model, "garage": garage_of(mid, base),
            "tier": "dedicated" if model.startswith("garage/") else "pool",
            "in": int(float(pin)), "out": int(float(pout)), "ok": status != "failure",
            "ttft_ms": ms(start, first) if first else None, "ms": ms(start, end) if end else None,
            "spend": round(float(spend), 6),
        })
    age = lambda r: now - datetime.fromisoformat(r["t"][:-1]).replace(tzinfo=timezone.utc).timestamp()
    flows, garages, keyagg = {}, {}, {}
    for r in reqs:
        a = age(r)
        f = flows.setdefault((r["key"], r["model"], r["garage"]), {"key": r["key"], "owner": r["owner"], "model": r["model"],
                                                                   "garage": r["garage"], "n": 0, "n_1m": 0, "fail": 0, "tokens": 0})
        f["n"] += 1
        f["n_1m"] += a < 60
        f["fail"] += not r["ok"]
        f["tokens"] += r["in"] + r["out"]
        g = garages.setdefault(r["garage"], {"n": 0, "n_1m": 0, "out_tokens": 0, "last_s": None, "ttft": []})
        g["n"] += 1
        g["n_1m"] += a < 60
        g["out_tokens"] += r["out"]
        g["last_s"] = round(a) if g["last_s"] is None else min(g["last_s"], round(a))
        if r["ttft_ms"] is not None and r["ok"]:
            g["ttft"].append(r["ttft_ms"])
        kk = keyagg.setdefault(r["key"], {"key": r["key"], "owner": r["owner"], "n": 0, "tokens": 0, "spend": 0.0, "last_s": None})
        kk["n"] += 1
        kk["tokens"] += r["in"] + r["out"]
        kk["spend"] += r["spend"]
        kk["last_s"] = round(a) if kk["last_s"] is None else min(kk["last_s"], round(a))
    for k in keyagg.values():
        k["spend"] = round(k["spend"], 4)
    for g in garages.values():
        t = sorted(g.pop("ttft"))
        g["ttft_p50_ms"] = t[len(t) // 2] if t else None
    last_min = [r for r in reqs if age(r) < 60]
    # garageai/auto: which place in the list answered, and how often it had to fall back.
    auto = None
    if autos:
        best = min((a["order"] for a in autos.values() if a.get("order") is not None), default=None)
        served = [r for r in reqs if r["auto_order"] is not None]
        entries = []
        for i, a in sorted(autos.items(), key=lambda kv: (kv[1].get("order") or 99)):
            mine = [r for r in served if r["auto_order"] == a.get("order") and r["garage"] == a.get("garage")]
            entries.append({"id": i, "order": a.get("order"), "garage": a.get("garage"), "model": a.get("model"), "public": a.get("public"),
                            "n": len(mine), "last_s": round(min(age(r) for r in mine)) if mine else None})
        auto = {"entries": entries, "requests": len(served), "fallbacks": sum(r["auto_order"] != best for r in served),
                "first_choice": best}
    return {
        "generated_at": datetime.now(timezone.utc).isoformat(timespec="seconds"),
        "window_minutes": WINDOW,
        "now": {"requests_1m": len(last_min), "tokens_1m": sum(r["in"] + r["out"] for r in last_min),
                "keys_1m": len({r["key"] for r in last_min}), "garages_1m": len({r["garage"] for r in last_min}),
                "requests_15m": len(reqs), "failed_15m": sum(not r["ok"] for r in reqs)},
        "recent": reqs[:RECENT],
        "flows": sorted(flows.values(), key=lambda f: -f["n"]),
        "garages": garages,
        "keys": sorted(keyagg.values(), key=lambda k: -k["n"]),
        "auto": auto,
    }


def write(data):
    tmp = OUT + ".tmp"
    with open(tmp, "w") as f:
        json.dump(data, f, separators=(",", ":"))
    os.chmod(tmp, 0o644)
    os.replace(tmp, OUT)


def main():
    keys, keys_at, autos, autos_at = {}, 0, {}, 0
    env = collect.load_env()
    once = "--once" in sys.argv
    while True:
        t0 = time.time()
        try:
            if time.time() - keys_at > OWNERS_EVERY:
                keys, keys_at = owners(), time.time()
            if time.time() - autos_at > 60:
                autos, autos_at = auto_entries(env), time.time()
            write(snapshot(keys, autos))
        except Exception as e:  # noqa: BLE001 - keep the live view running; the next round tries again
            print(f"live: {e}", file=sys.stderr, flush=True)
        if once:
            return
        time.sleep(max(1.0, EVERY - (time.time() - t0)))


if __name__ == "__main__":
    main()

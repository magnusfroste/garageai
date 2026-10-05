#!/usr/bin/env python3
"""garageai-health — level 1 and 2 health checks for every garage, run on the gateway.

Level 1 (tunnel):  is the garage's NetBird peer connected to the management server?
Level 2 (runtime): does the runtime answer GET /v1/models over the mesh, and with which models?
No tokens are generated. Level 3 (does the model actually generate) stays with the portal's
acceptance test and hourly probe.

Every run:
  1. targets  = GET  {portal}/functions/v1/gateway-targets      (falls back to a local file)
  2. level 1  = NetBird management API /peers (connected flag), by pinned peer id
  3. level 2  = GET http://<host>:<port>/v1/models over the mesh, 5 s timeout
  4. results  = POST {portal}/functions/v1/gateway-health-report (skipped in observe mode)
The gateway authenticates to the portal with the LiteLLM master key, a secret both already hold.
The last results are always written to /var/lib/garageai-health/last.json.

Config (environment, e.g. /etc/garageai/gateway-health.env):
  GARAGEAI_PORTAL_URL        https://<project>.supabase.co
  GARAGEAI_GATEWAY_KEY       the LiteLLM master key
  NETBIRD_API_URL            https://netbird.example.eu/api
  NETBIRD_API_TOKEN          NetBird token (read peers only is enough)
  GARAGEAI_TARGETS_FILE      optional local targets, used when the portal endpoint is missing
"""
import json, os, sys, time, urllib.error, urllib.request
from datetime import datetime, timezone

STATE_DIR = "/var/lib/garageai-health"
TIMEOUT = 5


def http(method, url, headers=None, body=None, timeout=TIMEOUT):
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(url, data=data, method=method, headers=headers or {})
    if data is not None:
        req.add_header("Content-Type", "application/json")
    with urllib.request.urlopen(req, timeout=timeout) as r:
        raw = r.read()
        return r.status, (json.loads(raw) if raw else None)


def load_targets(portal, gw_key, targets_file):
    """[{garage, peer_id, host, port, runtime_api_key?}] — from the portal, else the local file."""
    if portal and gw_key:
        try:
            _, data = http("GET", f"{portal}/functions/v1/gateway-targets",
                           {"x-gateway-key": gw_key}, timeout=15)
            return data.get("targets", []), "portal"
        except urllib.error.HTTPError as e:
            if e.code != 404:
                print(f"gateway-targets: HTTP {e.code}", file=sys.stderr)
        except Exception as e:  # noqa: BLE001 — portal down must not stop local checks
            print(f"gateway-targets: {e}", file=sys.stderr)
    if targets_file and os.path.exists(targets_file):
        with open(targets_file) as f:
            return json.load(f), "file"
    return [], "none"


def peers_connected(nb_url, nb_token):
    """{peer_id: connected, name: connected} from the NetBird management API, or None."""
    try:
        _, peers = http("GET", f"{nb_url}/peers", {"Authorization": f"Token {nb_token}",
                                                   "Accept": "application/json"}, timeout=15)
    except Exception as e:  # noqa: BLE001
        print(f"netbird peers: {e}", file=sys.stderr)
        return None
    out = {}
    for p in peers:
        out[p["id"]] = bool(p.get("connected"))
        out.setdefault("name:" + p.get("name", ""), bool(p.get("connected")))
    return out


def check_runtime(host, port, key):
    headers = {"Authorization": f"Bearer {key}"} if key else {}
    try:
        _, data = http("GET", f"http://{host}:{port}/v1/models", headers)
        return True, None, sorted(m["id"] for m in data.get("data", []))
    except urllib.error.HTTPError as e:
        # The runtime answered: it is up, but we cannot list models (e.g. wrong key).
        return True, f"http_{e.code}", None
    except Exception as e:  # noqa: BLE001
        reason = getattr(e, "reason", e)
        return False, type(reason).__name__ if not isinstance(reason, str) else reason, None


def main():
    portal = os.environ.get("GARAGEAI_PORTAL_URL", "").rstrip("/")
    gw_key = os.environ.get("GARAGEAI_GATEWAY_KEY", "")
    nb_url = os.environ.get("NETBIRD_API_URL", "").rstrip("/")
    nb_token = os.environ.get("NETBIRD_API_TOKEN", "")
    targets, source = load_targets(portal, gw_key, os.environ.get("GARAGEAI_TARGETS_FILE"))
    connected = peers_connected(nb_url, nb_token) if nb_url and nb_token else None

    now = datetime.now(timezone.utc).isoformat(timespec="seconds")
    results = []
    for t in targets:
        peer = t.get("peer_id")
        mesh = None if connected is None else connected.get(peer, connected.get("name:" + t["garage"], False))
        if mesh is False:
            runtime_ok, err, models = False, "mesh_disconnected", None
        else:
            runtime_ok, err, models = check_runtime(t["host"], t["port"], t.get("runtime_api_key"))
        results.append({"garage": t["garage"], "checked_at": now, "mesh_connected": mesh,
                        "runtime_ok": runtime_ok, "runtime_error": err, "models": models})

    os.makedirs(STATE_DIR, exist_ok=True)
    with open(f"{STATE_DIR}/last.json", "w") as f:
        json.dump({"checked_at": now, "source": source, "results": results}, f, indent=2)

    if source == "portal":
        try:
            http("POST", f"{portal}/functions/v1/gateway-health-report",
                 {"x-gateway-key": gw_key}, {"results": results}, timeout=15)
        except Exception as e:  # noqa: BLE001
            print(f"gateway-health-report: {e}", file=sys.stderr)

    for r in results:
        print(f"{r['garage']}: mesh={r['mesh_connected']} runtime={r['runtime_ok']}"
              f"{' (' + r['runtime_error'] + ')' if r['runtime_error'] else ''}"
              f" models={','.join(r['models'] or []) or '-'}")
    print(f"source={source} garages={len(results)}")


if __name__ == "__main__":
    main()

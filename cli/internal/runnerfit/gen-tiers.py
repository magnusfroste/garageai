#!/usr/bin/env python3
"""Regenerate tiers.json, Bridge's snapshot of Runner's docs/model-tiers.json at a release.

    cd cli/internal/runnerfit && python3 gen-tiers.py v1.15.0 > tiers.json

Fetches the roster with `gh` (or reads it from a local file given as the second argument) and
keeps tier1 and tier2 rows with exactly one file and at least one backend whose recorded gate
passed; only those backends are kept. fails_fidelity is whether the row's claim says its file
misses the fidelity bar (FAILS below). Python 3 stdlib only.
"""
import base64
import json
import re
import subprocess
import sys

REPO = "Joakimpalm-Zen/xyntetik-runner"

# How a claim says the row's own file misses the bar against its parent. "fails the fidelity bar"
# is the usual wording; Runner has also written "runs below the fidelity bar" (1.12), "fails the
# 0.05 bar" and "Nx the bar" (a mean KLD multiple, 1.15). Not "below the bar" alone: a passing
# row uses it for its siblings.
FAILS = re.compile(r"fails the (fidelity|0\.05) bar|runs below the fidelity bar|\b\d+x the bar\b")


def upstream(version):
    out = subprocess.run(
        ["gh", "api", f"repos/{REPO}/contents/docs/model-tiers.json?ref={version}", "--jq", ".content"],
        check=True, capture_output=True, text=True).stdout
    return json.loads(base64.b64decode(out))


def snapshot(version, roster):
    rows = []
    for r in roster["rows"]:
        if r["tier"] not in ("tier1", "tier2") or len(r.get("files") or []) != 1:
            continue
        passes = {b: {"date": c["date"], "device_class": c["device_class"], "runner_version": c["runner_version"]}
                  for b, c in (r.get("backends") or {}).items()
                  if c and c.get("kind") == "gate" and c.get("result") == "pass"}
        if not passes:
            continue
        f = r["files"][0]
        rows.append({
            "id": r["id"],
            "model": r["model"],
            "tier": r["tier"],
            "quantization": r["quantization"],
            "features": r["envelope"]["features"],
            "licence": r["licence"]["id"],
            "fails_fidelity": bool(FAILS.search(r["claim"])),
            "repo": f["repo"],
            "file": f["name"],
            "bytes": f["bytes"],
            "sha256": f["sha256"],
            "pass": passes,
        })
    return {
        "source": f"{REPO} docs/model-tiers.json at {version} (tier1 and tier2, single-file rows with a recorded "
                  "pass, passing backends only; licence and whether the row's claim says it fails the fidelity bar)",
        "schema": roster["schema"],
        "rows": rows,
    }


if __name__ == "__main__":
    if len(sys.argv) not in (2, 3):
        sys.exit(__doc__)
    version = sys.argv[1]
    roster = json.load(open(sys.argv[2])) if len(sys.argv) == 3 else upstream(version)
    if roster["schema"] != "xyntetik.runner.model-tiers.v1":
        sys.exit(f"unknown schema {roster['schema']}: check the filter before trusting the snapshot")
    sys.stdout.write(json.dumps(snapshot(version, roster), indent=1, ensure_ascii=False))

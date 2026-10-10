#!/usr/bin/env bash
# Parity: the Go binary (cli/) and garageai-connect.sh --doctor --json must describe the same
# machine the same way. Runs both against a fake runtime (bound to localhost, then to the network)
# and compares what matters: the runtimes found (port, kind, binds, network, models, windows) and
# the problem codes. Usage: tests/parity-doctor.sh PATH_TO_GARAGEAI_BINARY
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1
BIN="$1"
PORT="$(python3 -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1])')"
fails=0
norm='{runtimes: [.runtimes[] | {port, kind, binds, network, models: [.models[] | {id, context}]}],
       problems: ([.problems[].code] | sort)}'
for bind in 127.0.0.1 0.0.0.0; do
  python3 tests/fake_runtime.py "$bind" "$PORT" 2>/dev/null & RT=$!
  for _ in $(seq 1 20); do curl -fsS --max-time 1 "http://127.0.0.1:$PORT/v1/models" >/dev/null 2>&1 && break; sleep 0.5; done
  a="$(bash ./garageai-connect.sh --doctor --json 2>/dev/null | jq -S "$norm")"
  b="$("$BIN" doctor --json 2>/dev/null | jq -S "$norm")"
  kill "$RT" 2>/dev/null; wait "$RT" 2>/dev/null
  if [ -n "$a" ] && [ "$a" = "$b" ]; then
    echo "  ok    same profile with the runtime bound to $bind"
  else
    echo "  FAIL  profiles differ with the runtime bound to $bind"
    diff <(printf '%s\n' "$a") <(printf '%s\n' "$b") | sed 's/^/          /'
    fails=$((fails + 1))
  fi
done
[ "$fails" -eq 0 ] && echo "parity: all checks passed" || { echo "parity: $fails check(s) failed"; exit 1; }

#!/usr/bin/env bash
# Tests for garageai-connect.sh that need no NetBird, no GPU and no network:
# syntax (also under macOS's bash 3.2), option parsing and --doctor against a fake runtime.
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1
SCRIPT=./garageai-connect.sh
PORT=18000
fails=0
check() { # check "name" "expected text" "actual output"
  if printf '%s' "$3" | grep -qF -- "$2"; then printf '  ok    %s\n' "$1"
  else printf '  FAIL  %s\n        expected: %s\n        got:\n' "$1" "$2"; printf '%s\n' "$3" | sed 's/^/          /'; fails=$((fails + 1)); fi
}
start_runtime() {
  python3 tests/fake_runtime.py "$1" & RT=$!
  for _ in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do
    curl -fsS --max-time 1 "http://127.0.0.1:$PORT/v1/models" >/dev/null 2>&1 && return 0
    sleep 0.5
  done
  echo "  FAIL  fake runtime did not start on $1:$PORT"; fails=$((fails + 1))
}
stop_runtime() { kill "$RT" 2>/dev/null; wait "$RT" 2>/dev/null; }

echo "bash $(bash --version | head -n 1)"
bash -n "$SCRIPT" && echo "  ok    syntax" || { echo "  FAIL  syntax"; fails=$((fails + 1)); }

out="$(bash "$SCRIPT" --help 2>&1)"; check "--help lists --doctor" "--doctor" "$out"
out="$(bash "$SCRIPT" --no-such-option 2>&1)"; check "unknown option is reported" "Unknown option" "$out"
out="$(bash "$SCRIPT" --runtime nope 2>&1)"; check "unknown runtime is rejected" "Unknown runtime" "$out"
out="$(bash "$SCRIPT" --runtime paddock --name x 2>&1)"; check "paddock requires a key" "requires an API key" "$out"

out="$(bash "$SCRIPT" --doctor --runtime other --port $PORT 2>&1)"
check "doctor: no runtime" "nothing answers on port $PORT" "$out"

start_runtime 127.0.0.1
out="$(bash "$SCRIPT" --doctor --runtime ollama --port $PORT 2>&1)"
check "doctor: lists models" "answers on port $PORT with 2 model(s)" "$out"
check "doctor: localhost-only is a problem" "only listens on 127.0.0.1" "$out"
check "doctor: gives the Ollama hint" "OLLAMA_HOST=0.0.0.0:$PORT" "$out"
stop_runtime

start_runtime 0.0.0.0
out="$(bash "$SCRIPT" --doctor --runtime ollama --port $PORT 2>&1)"
check "doctor: network bind is fine" "listens on the network" "$out"
stop_runtime

out="$(printf 'n\n' | bash "$SCRIPT" --uninstall 2>&1)"; check "uninstall asks first" "Aborted" "$out"

[ "$fails" -eq 0 ] && echo "all tests passed" || { echo "$fails test(s) failed"; exit 1; }

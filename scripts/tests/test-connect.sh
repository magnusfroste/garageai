#!/usr/bin/env bash
# Tests for garageai-connect.sh that need no NetBird, no GPU and no network:
# syntax (also under macOS's bash 3.2), option parsing and --doctor against a fake runtime.
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1
SCRIPT=./garageai-connect.sh
# A free port, so the test does not depend on what the machine already runs.
PORT="$(python3 -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1])')"
fails=0
check() { # check "name" "expected text" "actual output"
  if printf '%s' "$3" | grep -qF -- "$2"; then printf '  ok    %s\n' "$1"
  else printf '  FAIL  %s\n        expected: %s\n        got:\n' "$1" "$2"; printf '%s\n' "$3" | sed 's/^/          /'; fails=$((fails + 1)); fi
}
start_runtime() {
  python3 tests/fake_runtime.py "$1" "$PORT" 2>/tmp/fake_runtime.err & RT=$!
  for _ in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do
    curl -fsS --max-time 1 "http://127.0.0.1:$PORT/v1/models" >/dev/null 2>&1 && return 0
    sleep 0.5
  done
  echo "  FAIL  fake runtime did not start on $1:$PORT"; sed 's/^/          /' /tmp/fake_runtime.err; fails=$((fails + 1))
}
stop_runtime() { kill "$RT" 2>/dev/null; wait "$RT" 2>/dev/null; }

echo "bash $(bash --version | head -n 1)"
bash -n "$SCRIPT" && echo "  ok    syntax" || { echo "  FAIL  syntax"; fails=$((fails + 1)); }

out="$(bash "$SCRIPT" --help 2>&1)"; check "--help lists --doctor" "--doctor" "$out"
out="$(bash "$SCRIPT" --no-such-option 2>&1)"; check "unknown option is reported" "Unknown option" "$out"
out="$(bash "$SCRIPT" --runtime nope 2>&1)"; check "unknown runtime is rejected" "Unknown runtime" "$out"
out="$(bash "$SCRIPT" --runtime paddock --name x 2>&1)"; check "paddock requires a key" "requires an API key" "$out"
out="$(GARAGEAI_RUNTIME_API_KEY='sk-abcGARAGEAI_MESH_WAIT_SECONDS=90' bash "$SCRIPT" --runtime vllm --name x 2>&1)"
check "key with another variable pasted in is rejected" "something else was pasted into it" "$out"
out="$(GARAGEAI_RUNTIME_API_KEY='sk-abc def' bash "$SCRIPT" --runtime vllm --name x 2>&1)"
check "key with a space is rejected" "something else was pasted into it" "$out"

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
check "doctor: a 4,096-token Ollama window is a problem" "context window only 4096 tokens" "$out"
out="$(GARAGEAI_OLLAMA_CONTEXT=4096 bash "$SCRIPT" --doctor --runtime ollama --port $PORT 2>&1)"
check "doctor: a window at the target is fine" "context window 4096 tokens" "$out"
stop_runtime

out="$(printf 'n\n' | bash "$SCRIPT" --uninstall 2>&1)"; check "uninstall asks first" "Aborted" "$out"

# Onboarding report: a failed connect tells the portal which step stopped and why.
REPORTS="$(mktemp)"
FAKE_REPORT_LOG="$REPORTS" start_runtime 127.0.0.1
REG="http://127.0.0.1:$PORT/functions/v1/register-node"
out="$(GARAGEAI_RUNTIME_API_KEY='sk-secret "quoted" value' bash "$SCRIPT" --runtime vllm --name rep-test \
       --register-url "$REG" --register-token tok-123 2>&1)"
check "report: the script still fails as before" "something else was pasted into it" "$out"
check "report: start is reported" '"status": "started"' "$(python3 -c 'import json,sys; [print(json.dumps(json.loads(json.loads(l)["body"]))) for l in open(sys.argv[1])]' "$REPORTS" 2>&1)"
check "report: the failure and its message are reported" '"status": "failed", "message": "The runtime API key contains spaces' \
      "$(python3 -c 'import json,sys; [print(json.dumps(json.loads(json.loads(l)["body"]))) for l in open(sys.argv[1])]' "$REPORTS" 2>&1)"
check "report: every report is valid JSON with the machine facts" "valid" \
      "$(python3 -c 'import json,sys; b=[json.loads(json.loads(l)["body"]) for l in open(sys.argv[1])]; print("valid" if b and all({"step","status","os","arch","node_name"} <= set(x) for x in b) else b)' "$REPORTS" 2>&1)"
check "report: authenticated with the register token" "Bearer tok-123" "$(cat "$REPORTS")"
check "report: no key in the body" "no secret" "$(python3 -c 'import json,sys; print("no secret" if not any("sk-secret" in json.loads(l)["body"] or "tok-123" in json.loads(l)["body"] for l in open(sys.argv[1])) else "LEAK")' "$REPORTS")"
: > "$REPORTS"
out="$(GARAGEAI_REPORT=0 GARAGEAI_RUNTIME_API_KEY='a b' bash "$SCRIPT" --runtime vllm --name x --register-url "$REG" --register-token t 2>&1)"
check "report: GARAGEAI_REPORT=0 sends nothing" "0" "$(wc -l < "$REPORTS" | tr -d ' ')"
out="$(bash "$SCRIPT" --doctor --runtime other --port $PORT --register-url "$REG" --register-token t 2>&1)"
check "report: --doctor sends nothing" "0" "$(wc -l < "$REPORTS" | tr -d ' ')"
stop_runtime
DEAD="$(python3 -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1])')"
out="$(GARAGEAI_RUNTIME_API_KEY='a b' bash "$SCRIPT" --runtime vllm --name x --register-url "http://127.0.0.1:$DEAD/register-node" --register-token t 2>&1)"
check "report: an unreachable portal does not change the outcome" "something else was pasted into it" "$out"
rm -f "$REPORTS"

# Context windows: the filters, then the heartbeat end to end against the fake runtime.
CTX_JQ="$(sed -n "s/^CONTEXTS_JQ='\(.*\)'$/\1/p" "$SCRIPT" | head -n 1)"
OLLAMA_JQ="$(sed -n "s/^OLLAMA_CONTEXTS_JQ='\(.*\)'$/\1/p" "$SCRIPT" | head -n 1)"
check "contexts: vLLM max_model_len" '{"m":524288}' "$(echo '{"data":[{"id":"m","max_model_len":524288}]}' | jq -c "$CTX_JQ")"
check "contexts: llama.cpp n_ctx_train" '{"g":8192}' "$(echo '{"data":[{"id":"g","meta":{"n_ctx_train":8192}}]}' | jq -c "$CTX_JQ")"
check "contexts: context_length, and models without one are left out" '{"a":4096}' "$(echo '{"data":[{"id":"a","context_length":4096},{"id":"b"}]}' | jq -c "$CTX_JQ")"
check "contexts: Ollama /api/ps" '{"qwen3:4b":16384}' "$(echo '{"models":[{"name":"qwen3:4b","context_length":16384}]}' | jq -c "$OLLAMA_JQ")"
HB_DIR="$(mktemp -d)"; REPORTS="$HB_DIR/reports"
sed -n "/<<'HEARTBEAT'$/,/^HEARTBEAT$/p" "$SCRIPT" | sed '1d;$d' > "$HB_DIR/heartbeat"
FAKE_REPORT_LOG="$REPORTS" start_runtime 127.0.0.1
for rt in vllm ollama; do
  : > "$REPORTS"
  printf '%s\n' "GARAGEAI_HEARTBEAT_URL=http://127.0.0.1:$PORT/functions/v1/node-heartbeat" "GARAGEAI_REGISTER_TOKEN=tok" \
    "GARAGEAI_NODE_NAME=hb-test" "GARAGEAI_RUNTIME=$rt" "GARAGEAI_PORT=$PORT" "GARAGEAI_RUNTIME_API_KEY=" "GARAGEAI_MESH_IP=" > "$HB_DIR/env"
  GARAGEAI_HEARTBEAT_CONF="$HB_DIR/env" bash "$HB_DIR/heartbeat" >/dev/null 2>&1
  body="$(python3 -c 'import json,sys; print(json.loads(open(sys.argv[1]).readline())["body"])' "$REPORTS" 2>&1)"
  [ "$rt" = vllm ] && check "heartbeat (vllm) sends contexts" '"contexts":{"qwen3:4b":32768}' "$body"
  [ "$rt" = ollama ] && check "heartbeat (ollama) uses the loaded window" '"contexts":{"qwen3:4b":4096}' "$body"
done
check "heartbeat still sends the models" '"models":["qwen3:4b","nomic-embed-text:latest"]' "$body"
stop_runtime
rm -rf "$HB_DIR"

[ "$fails" -eq 0 ] && echo "all tests passed" || { echo "$fails test(s) failed"; exit 1; }

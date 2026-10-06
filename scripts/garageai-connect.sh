#!/usr/bin/env bash
# garageai-connect.sh — connect a garage inference node to the GarageAI mesh.
#
# Run this on the machine in your garage that serves models. It:
#   1. installs the NetBird client if it is missing (open source, NetBird GmbH, Berlin)
#   2. joins the private GarageAI WireGuard mesh using your setup key
#   3. checks that your inference runtime answers an OpenAI-compatible API
#      (Ollama, LM Studio, llama.cpp, vLLM, Paddock, Unsloth, ...)
#   4. checks that the runtime is reachable on this node's mesh IP
#   5. lists the models you offer and registers the node with GarageAI
#      (or prints the details for manual registration)
#   6. installs a small heartbeat (systemd timer on Linux, launchd on macOS) that
#      reports the runtime's current models every 5 minutes, so loading or removing a
#      model updates what GarageAI sells without re-running this script
#
# Your runtime is never exposed to the public internet: only the GarageAI
# gateway can reach it, over the encrypted mesh. This script never needs the
# gateway's LiteLLM master key.
#
# Usage:
#   ./garageai-connect.sh --setup-key KEY --management-url https://netbird.example.eu \
#       [--runtime ollama|lmstudio|llamacpp|vllm|sglang|paddock|unsloth|mlx|lemonade|other] \
#       [--port PORT] \
#       [--name NODE_NAME] [--runtime-api-key KEY] [--register-url URL --register-token TOKEN] \
#       [--models MODEL[,MODEL...]] [--skip-install] [--no-heartbeat] [--yes]
#   ./garageai-connect.sh --doctor       check this garage and say exactly what to fix
#   ./garageai-connect.sh --uninstall    remove the heartbeat, the Ollama login item and leave the mesh
#   ./garageai-connect.sh --remove-heartbeat
#
# --runtime-api-key is for runtimes started with an API key (e.g. vLLM --api-key). It is
# used to query the runtime and is sent to GarageAI with the registration, so the gateway
# can call the runtime. It is never shown to buyers.
#
# Every option can also be given as an environment variable:
#   GARAGEAI_SETUP_KEY, GARAGEAI_MANAGEMENT_URL, GARAGEAI_RUNTIME, GARAGEAI_PORT,
#   GARAGEAI_NODE_NAME, GARAGEAI_RUNTIME_API_KEY, GARAGEAI_REGISTER_URL, GARAGEAI_REGISTER_TOKEN,
#   GARAGEAI_MODELS
#
# --models chooses which of the runtime's models to offer at registration (comma-separated).
# Without it every chat model the runtime lists is offered; embedding and reranker models are
# never offered. Later changes (offer, pause) are made in the GarageAI portal, under My garages.
#
# Supported: Linux and macOS. Windows: garageai-connect.ps1 (Ollama and LM Studio).

set -euo pipefail

SETUP_KEY="${GARAGEAI_SETUP_KEY:-}"
MANAGEMENT_URL="${GARAGEAI_MANAGEMENT_URL:-}"
RUNTIME="${GARAGEAI_RUNTIME:-ollama}"
PORT="${GARAGEAI_PORT:-}"
NODE_NAME="${GARAGEAI_NODE_NAME:-$(hostname -s 2>/dev/null || hostname)}"
RUNTIME_API_KEY="${GARAGEAI_RUNTIME_API_KEY:-}"
OFFER_MODELS="${GARAGEAI_MODELS:-}"
REGISTER_URL="${GARAGEAI_REGISTER_URL:-}"
REGISTER_TOKEN="${GARAGEAI_REGISTER_TOKEN:-}"
SKIP_INSTALL=0
HEARTBEAT=1
REMOVE_HEARTBEAT=0
DOCTOR=0
UNINSTALL=0
RUNTIME_GIVEN=0
[ -n "${GARAGEAI_RUNTIME:-}" ] && RUNTIME_GIVEN=1
PORT_GIVEN=0
[ -n "${GARAGEAI_PORT:-}" ] && PORT_GIVEN=1
ASSUME_YES=0
MESH_WAIT_SECONDS="${GARAGEAI_MESH_WAIT_SECONDS:-30}"

bold() { printf '\033[1m%s\033[0m\n' "$*"; }
info() { printf '  %s\n' "$*"; }
ok()   { printf '  \033[32m✓\033[0m %s\n' "$*"; }
warn() { printf '  \033[33m!\033[0m %s\n' "$*" >&2; }
die()  { printf '  \033[31m✗\033[0m %s\n' "$*" >&2; exit 1; }

usage() { sed -n '2,43p' "$0" | sed 's/^# \{0,1\}//'; exit "${1:-0}"; }

while [ $# -gt 0 ]; do
  case "$1" in
    --setup-key)      SETUP_KEY="${2:-}"; shift 2 ;;
    --management-url) MANAGEMENT_URL="${2:-}"; shift 2 ;;
    --runtime)        RUNTIME="${2:-}"; RUNTIME_GIVEN=1; shift 2 ;;
    --port)           PORT="${2:-}"; PORT_GIVEN=1; shift 2 ;;
    --name)           NODE_NAME="${2:-}"; shift 2 ;;
    --runtime-api-key) RUNTIME_API_KEY="${2:-}"; shift 2 ;;
    --models)         OFFER_MODELS="${2:-}"; shift 2 ;;
    --register-url)   REGISTER_URL="${2:-}"; shift 2 ;;
    --register-token) REGISTER_TOKEN="${2:-}"; shift 2 ;;
    --skip-install)   SKIP_INSTALL=1; shift ;;
    --no-heartbeat)   HEARTBEAT=0; shift ;;
    --remove-heartbeat) REMOVE_HEARTBEAT=1; shift ;;
    --doctor)         DOCTOR=1; shift ;;
    --uninstall)      UNINSTALL=1; shift ;;
    --yes|-y)         ASSUME_YES=1; shift ;;
    -h|--help)        usage 0 ;;
    *) warn "Unknown option: $1"; usage 1 ;;
  esac
done

default_port() {
  case "$1" in
    ollama)   echo 11434 ;;
    lmstudio) echo 1234 ;;
    llamacpp) echo 8080 ;;
    vllm)     echo 8000 ;;
    sglang)   echo 30000 ;;
    paddock)  echo 11540 ;;
    unsloth)  echo 8888 ;;
    mlx)      echo 8080 ;;
    lemonade) echo 13305 ;;
    *)        echo "" ;;
  esac
}

if [ "$DOCTOR" -eq 1 ] && [ -e /etc/garageai/heartbeat.env ]; then
  # Root-only file: read it only if sudo needs no password; otherwise keep the defaults.
  HB="$(sudo -n cat /etc/garageai/heartbeat.env 2>/dev/null || true)"
  hb_get() { printf '%s\n' "$HB" | sed -n "s/^$1=//p" | head -n 1; }
  [ "$RUNTIME_GIVEN" -eq 1 ] || { v="$(hb_get GARAGEAI_RUNTIME)"; [ -z "$v" ] || RUNTIME="$v"; }
  [ "$PORT_GIVEN" -eq 1 ] || { v="$(hb_get GARAGEAI_PORT)"; [ -z "$v" ] || PORT="$v"; }
  [ -n "$RUNTIME_API_KEY" ] || RUNTIME_API_KEY="$(hb_get GARAGEAI_RUNTIME_API_KEY)"
fi

case "$RUNTIME" in
  ollama|lmstudio|llamacpp|vllm|sglang|paddock|unsloth|mlx|lemonade|other) ;;
  *) die "Unknown runtime '$RUNTIME' (use ollama, lmstudio, llamacpp, vllm, sglang, paddock, unsloth, mlx, lemonade or other)" ;;
esac

# Paddock creates and requires an API key whenever it listens beyond localhost.
case "$RUNTIME_API_KEY" in \<*\>|YOUR_*|'<YOUR'*) die "Replace the placeholder in --runtime-api-key / GARAGEAI_RUNTIME_API_KEY with your runtime's real API key (or leave it out)." ;; esac
if [ "$RUNTIME" = paddock ] && [ -z "$RUNTIME_API_KEY" ] && [ "$DOCTOR" -eq 0 ] && [ "$UNINSTALL" -eq 0 ]; then
  die "Paddock requires an API key on network binds. Pass the same key with --runtime-api-key."
fi

[ -n "$PORT" ] || PORT="$(default_port "$RUNTIME")"
[ -n "$PORT" ] || die "Runtime '$RUNTIME' has no default port — pass --port with the port its OpenAI-compatible server listens on."
case "$PORT" in *[!0-9]*|'') die "Invalid port: $PORT" ;; esac

# Print how to start a runtime so that the mesh can reach it.
runtime_hint() {
  local bind="$1"
  case "$RUNTIME" in
    ollama)
      info "Ollama listens on 127.0.0.1 by default. Start it bound to the mesh:"
      info "    OLLAMA_HOST=${bind}:${PORT} OLLAMA_NUM_PARALLEL=4 ollama serve"
      info "  On Linux with the systemd service: sudo systemctl edit ollama, add"
      info "    [Service]"
      info "    Environment=\"OLLAMA_HOST=${bind}:${PORT}\""
      info "    Environment=\"OLLAMA_NUM_PARALLEL=4\""
      info "  then: sudo systemctl restart ollama"
      info "  On macOS: run this script again and accept the offer to make it permanent, or:"
      info "    launchctl setenv OLLAMA_HOST ${bind}:${PORT}"
      info "  then restart Ollama: Ollama app → quit from the menu bar and open it again;"
      info "  Homebrew → brew services restart ollama"
      info "  (Install: macOS → the Ollama app from ollama.com/download, or brew install ollama;"
      info "   Linux → curl -fsSL https://ollama.com/install.sh | sh)"
      info "  Ollama has no API key; only the gateway can reach it over the mesh." ;;
    lmstudio)
      info "LM Studio: Developer tab → start the server on port ${PORT} and enable"
      info "  \"Serve on Local Network\" so it is not bound to 127.0.0.1 only."
      info "  Headless: lms server start --bind ${bind} --port ${PORT}"
      info "  Optional: Settings → Require Authentication, then pass --runtime-api-key." ;;
    llamacpp)
      info "llama.cpp:"
      info "    llama-server -m /path/to/model.gguf --host ${bind} --port ${PORT} -np 4 --jinja [--api-key KEY]" ;;
    vllm)
      info "vLLM:"
      info "    vllm serve <model> --host ${bind} --port ${PORT} [--api-key KEY] [--served-model-name NAME]" ;;
    sglang)
      info "SGLang:"
      info "    python -m sglang.launch_server --model-path <model> --host ${bind} --port ${PORT} [--api-key KEY]" ;;
    paddock)
      info "Paddock (beta):"
      info "    paddock-runner --model /path/to/model.gguf --host ${bind} --port ${PORT} --api-key KEY"
      info "  Pass the same key here with --runtime-api-key." ;;
    unsloth)
      info "Unsloth:"
      info "    unsloth run --model <repo>:<quant> -H ${bind} -p ${PORT} --disable-tools"
      info "  Create an API key in Settings → API and pass it with --runtime-api-key."
      info "  Or export the model to GGUF and serve it with --runtime llamacpp or ollama." ;;
    mlx)
      info "MLX (Apple Silicon):"
      info "    mlx_lm.server --model <model> --host ${bind} --port ${PORT}"
      info "  mlx-lm has no API key; only the gateway can reach it over the mesh." ;;
    lemonade)
      info "Lemonade (AMD):"
      info "    LEMONADE_API_KEY=KEY lemond --host ${bind} --port ${PORT}"
      info "  Pass the same key with --runtime-api-key." ;;
    other)
      info "Start your OpenAI-compatible server on port ${PORT}, bound to ${bind} instead of"
      info "  127.0.0.1 (see its documentation). It must serve /v1/models and /v1/chat/completions." ;;
  esac
}

as_root() {
  if [ "$(id -u)" -eq 0 ]; then "$@"; else sudo "$@"; fi
}

confirm() {
  [ "$ASSUME_YES" -eq 1 ] && return 0
  printf '  %s [y/N] ' "$1"
  local answer=""
  read -r answer || true
  case "$answer" in y|Y|yes|YES) return 0 ;; *) return 1 ;; esac
}

http_models() {
  # Prints one model id per line, or fails if no OpenAI-compatible API answers.
  local auth=()
  [ -n "$RUNTIME_API_KEY" ] && auth=(-H "Authorization: Bearer ${RUNTIME_API_KEY}")
  curl -fsS --max-time 5 ${auth[@]+"${auth[@]}"} "http://$1:${PORT}/v1/models" | jq -er '.data[].id'
}

HEARTBEAT_CONF=/etc/garageai/heartbeat.env
HEARTBEAT_BIN=/usr/local/bin/garageai-heartbeat
HEARTBEAT_PLIST=/Library/LaunchDaemons/eu.garageai.heartbeat.plist
HEARTBEAT_UNIT=/etc/systemd/system/garageai-heartbeat

remove_heartbeat() {
  case "$(uname -s)" in
    Darwin)
      as_root launchctl bootout system "$HEARTBEAT_PLIST" 2>/dev/null || true
      as_root rm -f "$HEARTBEAT_PLIST" ;;
    *)
      as_root systemctl disable --now garageai-heartbeat.timer 2>/dev/null || true
      as_root rm -f "${HEARTBEAT_UNIT}.service" "${HEARTBEAT_UNIT}.timer"
      as_root systemctl daemon-reload 2>/dev/null || true ;;
  esac
  as_root rm -f "$HEARTBEAT_BIN" "$HEARTBEAT_CONF"
}

install_heartbeat() {
  local url="${REGISTER_URL%/register-node}/node-heartbeat"
  as_root mkdir -p /etc/garageai /usr/local/bin
  # The register token and runtime key are secrets: root-only file.
  as_root sh -c "umask 077 && : > '$HEARTBEAT_CONF'"
  {
    printf 'GARAGEAI_HEARTBEAT_URL=%q\n' "$url"
    printf 'GARAGEAI_REGISTER_TOKEN=%q\n' "$REGISTER_TOKEN"
    printf 'GARAGEAI_NODE_NAME=%q\n' "$NODE_NAME"
    printf 'GARAGEAI_RUNTIME=%q\n' "$RUNTIME"
    printf 'GARAGEAI_PORT=%q\n' "$PORT"
    printf 'GARAGEAI_RUNTIME_API_KEY=%q\n' "$RUNTIME_API_KEY"
    printf 'GARAGEAI_MESH_IP=%q\n' "$MESH_IP"
  } | as_root tee "$HEARTBEAT_CONF" >/dev/null

  as_root tee "$HEARTBEAT_BIN" >/dev/null <<'HEARTBEAT'
#!/usr/bin/env bash
# garageai-heartbeat — reports this garage's current models to GarageAI.
# Installed by garageai-connect.sh and run every 5 minutes. Remove it with
# garageai-connect.sh --remove-heartbeat.
set -euo pipefail
# shellcheck disable=SC1090
. "${GARAGEAI_HEARTBEAT_CONF:-/etc/garageai/heartbeat.env}"
export PATH="/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin:${PATH:-}"

auth=()
[ -n "${GARAGEAI_RUNTIME_API_KEY:-}" ] && auth=(-H "Authorization: Bearer ${GARAGEAI_RUNTIME_API_KEY}")
models='[]'
for host in 127.0.0.1 "${GARAGEAI_MESH_IP:-}"; do
  [ -n "$host" ] || continue
  if out="$(curl -fsS --max-time 5 ${auth[@]+"${auth[@]}"} "http://${host}:${GARAGEAI_PORT}/v1/models" 2>/dev/null)" &&
     list="$(printf '%s' "$out" | jq -ec '[.data[].id]' 2>/dev/null)"; then
    models="$list"; break
  fi
done
# The heartbeat reports everything the runtime serves (the inventory); which models are
# offered is decided in the GarageAI portal. If the runtime does not answer, it reports no
# models so buyers are not routed here.

payload="$(jq -nc --arg name "$GARAGEAI_NODE_NAME" --argjson port "$GARAGEAI_PORT" \
  --arg runtime "$GARAGEAI_RUNTIME" --argjson models "$models" --arg key "${GARAGEAI_RUNTIME_API_KEY:-}" \
  '{name: $name, port: $port, runtime: $runtime, models: $models}
   + (if $key != "" then {runtime_api_key: $key} else {} end)')"
curl -fsS --max-time 180 -X POST "$GARAGEAI_HEARTBEAT_URL" \
  -H "Authorization: Bearer ${GARAGEAI_REGISTER_TOKEN}" \
  -H "Content-Type: application/json" -d "$payload"
echo
HEARTBEAT
  as_root chmod 755 "$HEARTBEAT_BIN"

  case "$(uname -s)" in
    Darwin)
      as_root tee "$HEARTBEAT_PLIST" >/dev/null <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>eu.garageai.heartbeat</string>
  <key>ProgramArguments</key><array><string>${HEARTBEAT_BIN}</string></array>
  <key>StartInterval</key><integer>300</integer>
  <key>RunAtLoad</key><true/>
  <key>StandardOutPath</key><string>/var/log/garageai-heartbeat.log</string>
  <key>StandardErrorPath</key><string>/var/log/garageai-heartbeat.log</string>
</dict>
</plist>
PLIST
      as_root launchctl bootout system "$HEARTBEAT_PLIST" 2>/dev/null || true
      as_root launchctl bootstrap system "$HEARTBEAT_PLIST" ;;
    *)
      as_root tee "${HEARTBEAT_UNIT}.service" >/dev/null <<UNIT
[Unit]
Description=GarageAI heartbeat (reports this garage's models)
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=${HEARTBEAT_BIN}
UNIT
      as_root tee "${HEARTBEAT_UNIT}.timer" >/dev/null <<UNIT
[Unit]
Description=Run the GarageAI heartbeat every 5 minutes

[Timer]
OnBootSec=1min
OnUnitActiveSec=5min

[Install]
WantedBy=timers.target
UNIT
      as_root systemctl daemon-reload
      as_root systemctl enable --now garageai-heartbeat.timer >/dev/null ;;
  esac
}

listen_addrs() {
  # Local addresses the runtime listens on for $PORT, one per line ("*" = all).
  if command -v ss >/dev/null 2>&1; then
    ss -ltnH "sport = :${PORT}" 2>/dev/null | awk '{print $4}' | sed 's/:[0-9]*$//; s/%.*//' | sort -u
  elif command -v lsof >/dev/null 2>&1; then
    lsof -nP -iTCP:"${PORT}" -sTCP:LISTEN 2>/dev/null | awk 'NR > 1 {print $9}' | sed 's/:[0-9]*$//' | sort -u
  fi
}

# Model ids that are embedding or reranker models: not sold as chat models.
NON_CHAT_RE='embed|bge-|bge:|e5-|minilm|rerank|colbert|gte-'

select_models() {
  # Reads model ids on stdin, prints the ones to offer.
  if [ -n "$OFFER_MODELS" ]; then
    local all; all="$(cat)"
    printf '%s\n' "$OFFER_MODELS" | tr ',' '\n' | sed 's/^ *//; s/ *$//' | while IFS= read -r want; do
      [ -n "$want" ] || continue
      if printf '%s\n' "$all" | grep -qxF -- "$want"; then printf '%s\n' "$want"
      else warn "--models: '${want}' is not served by the runtime; skipped"; fi
    done
  else
    grep -viE "$NON_CHAT_RE" || true
  fi
}

install_ollama_env_agent() {
  # macOS: make OLLAMA_HOST survive logout and reboot. The Ollama app reads it at start.
  local plist="$HOME/Library/LaunchAgents/eu.garageai.ollama-host.plist"
  mkdir -p "$HOME/Library/LaunchAgents"
  cat > "$plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>eu.garageai.ollama-host</string>
  <key>ProgramArguments</key>
  <array><string>/bin/launchctl</string><string>setenv</string><string>OLLAMA_HOST</string><string>0.0.0.0:${PORT}</string></array>
  <key>RunAtLoad</key><true/>
</dict>
</plist>
PLIST
  launchctl bootout "gui/$(id -u)" "$plist" 2>/dev/null || true
  launchctl bootstrap "gui/$(id -u)" "$plist" 2>/dev/null || launchctl load "$plist"
  launchctl setenv OLLAMA_HOST "0.0.0.0:${PORT}"
}

mesh_ip() {
  local ip=""
  ip="$(netbird status --json 2>/dev/null | jq -r '.netbirdIp // empty' 2>/dev/null || true)"
  if [ -z "$ip" ]; then
    ip="$(netbird status 2>/dev/null | awk -F': *' '/NetBird IP/ {print $2; exit}' || true)"
  fi
  printf '%s' "${ip%%/*}"
}

run_doctor() {
  local problems=0 ip listen models hb_active last
  bad() { printf '  \033[31m✗\033[0m %s\n' "$1"; shift; for l in "$@"; do info "    → $l"; done; problems=$((problems + 1)); }
  bold "GarageAI doctor — runtime ${RUNTIME}, port ${PORT}"

  for tool in curl jq; do
    command -v "$tool" >/dev/null 2>&1 || bad "'$tool' is not installed" "macOS: brew install $tool    Linux: sudo apt install $tool"
  done

  # Level 1: the tunnel
  if ! command -v netbird >/dev/null 2>&1; then
    bad "NetBird is not installed" "run the connect command from the GarageAI portal"
  else
    ip="$(mesh_ip)"
    if netbird status 2>/dev/null | grep -q '^Management: Connected' && [ -n "$ip" ]; then
      ok "Tunnel: connected to the GarageAI mesh as ${ip}"
    else
      bad "Tunnel: NetBird is installed but not connected" "sudo netbird up     (if that asks you to log in, get a new command from the portal)"
    fi
  fi

  # Level 2: the runtime
  if models="$(http_models 127.0.0.1 2>/dev/null)" || { [ -n "${ip:-}" ] && models="$(http_models "$ip" 2>/dev/null)"; }; then
    ok "Runtime: ${RUNTIME} answers on port ${PORT} with $(printf '%s\n' "$models" | grep -c .) model(s): $(printf '%s' "$models" | tr '\n' ' ')"
    listen="$(listen_addrs)"
    if [ -z "$listen" ]; then
      info "  (could not read which address it listens on; lsof or ss is missing)"
    elif printf '%s\n' "$listen" | grep -qxE "\\*|0\\.0\\.0\\.0|\\[::\\]|::|${ip:-none}"; then
      ok "Runtime: listens on the network ($(printf '%s' "$listen" | tr '\n' ' ')), so the gateway can reach it"
    else
      bad "Runtime: only listens on $(printf '%s' "$listen" | tr '\n' ' '), so the gateway cannot reach it" "see below"
      runtime_hint "0.0.0.0"
    fi
  else
    bad "Runtime: nothing answers on port ${PORT}" "start ${RUNTIME} (and check --runtime / --port if you use another one)"
  fi

  # Heartbeat
  if [ ! -e "$HEARTBEAT_BIN" ]; then
    bad "Heartbeat: not installed, so model changes and outages are noticed late" "run the connect command from the portal again (\"New command\")"
  else
    case "$(uname -s)" in
      Darwin) launchctl print system/eu.garageai.heartbeat >/dev/null 2>&1 && hb_active=1 || hb_active=0
              last="$(tail -n 1 /var/log/garageai-heartbeat.log 2>/dev/null || true)" ;;
      *)      systemctl is-active --quiet garageai-heartbeat.timer 2>/dev/null && hb_active=1 || hb_active=0
              last="$(journalctl -u garageai-heartbeat.service -n 1 -o cat 2>/dev/null || true)" ;;
    esac
    if [ "$hb_active" -eq 1 ]; then ok "Heartbeat: installed and scheduled"
    else bad "Heartbeat: installed but not running" "run the connect command from the portal again"; fi
    case "$last" in
      *'"ok":true'*) ok "Heartbeat: last report was accepted by GarageAI" ;;
      *401*|*nauthorized*) bad "Heartbeat: GarageAI rejects this garage's token (it was replaced or revoked)" "My garages → New command, and run that command here" ;;
      "") : ;;
      *) info "  last heartbeat output: $(printf '%s' "$last" | cut -c1-120)" ;;
    esac
  fi

  if [ "$(uname -s)" = Darwin ]; then
    if [ "$RUNTIME" = ollama ] && [ ! -e "$HOME/Library/LaunchAgents/eu.garageai.ollama-host.plist" ]; then
      info "  Note: OLLAMA_HOST is not set permanently; after a reboot Ollama listens on localhost again."
      info "    → run the connect command again and accept the offer to make it permanent"
    fi
    info "  Note: a sleeping Mac is offline for buyers. For a garage that should stay up:"
    info "    → System Settings → Battery/Energy → prevent automatic sleeping when the display is off"
  fi

  echo
  if [ "$problems" -eq 0 ]; then ok "No problems found. If the portal still shows the garage as offline, use Retest there."; return 0; fi
  warn "${problems} problem(s) found — fix the lines marked ✗ from the top down, then run --doctor again."
  return 1
}

if [ "$DOCTOR" -eq 1 ]; then
  run_doctor && exit 0 || exit 1
fi

if [ "$UNINSTALL" -eq 1 ]; then
  bold "Remove GarageAI from this machine"
  info "This removes the heartbeat and the Ollama login item and disconnects from the mesh."
  info "Your runtime and models are not touched. NetBird itself stays installed."
  confirm "Continue?" || die "Aborted."
  remove_heartbeat
  ok "Heartbeat removed"
  if [ "$(uname -s)" = Darwin ] && [ -e "$HOME/Library/LaunchAgents/eu.garageai.ollama-host.plist" ]; then
    launchctl bootout "gui/$(id -u)" "$HOME/Library/LaunchAgents/eu.garageai.ollama-host.plist" 2>/dev/null || true
    rm -f "$HOME/Library/LaunchAgents/eu.garageai.ollama-host.plist"
    launchctl unsetenv OLLAMA_HOST 2>/dev/null || true
    ok "Ollama login item removed (restart Ollama to listen on localhost only again)"
  fi
  if command -v netbird >/dev/null 2>&1; then
    as_root netbird down >/dev/null 2>&1 || true
    ok "Disconnected from the mesh"
    info "To remove NetBird too: macOS → sudo netbird service uninstall, then delete the app; Linux → sudo apt remove netbird"
  fi
  info "Finally, remove the garage under My garages in the portal so it is not offered again."
  exit 0
fi

if [ "$REMOVE_HEARTBEAT" -eq 1 ]; then
  remove_heartbeat
  ok "Heartbeat removed. The garage will show as offline on GarageAI after 15 minutes."
  exit 0
fi

bold "GarageAI node connect — ${NODE_NAME}"
echo

# 0. Prerequisites
command -v curl >/dev/null 2>&1 || die "'curl' is required (e.g. 'sudo apt install curl')."
if ! command -v jq >/dev/null 2>&1; then
  if command -v brew >/dev/null 2>&1 && confirm "'jq' is required. Install it with Homebrew now?"; then brew install jq
  elif command -v apt-get >/dev/null 2>&1 && confirm "'jq' is required. Install it with apt now?"; then as_root apt-get install -y jq
  elif command -v dnf >/dev/null 2>&1 && confirm "'jq' is required. Install it with dnf now?"; then as_root dnf install -y jq
  fi
  command -v jq >/dev/null 2>&1 || die "'jq' is required (macOS: brew install jq; Linux: sudo apt install jq)."
fi

# 1. NetBird client
bold "1/6  NetBird client"
if command -v netbird >/dev/null 2>&1; then
  ok "netbird is installed ($(netbird version 2>/dev/null || echo 'unknown version'))"
elif [ "$SKIP_INSTALL" -eq 1 ]; then
  die "netbird is not installed and --skip-install was given."
else
  info "NetBird is not installed. It will be installed from https://pkgs.netbird.io/install.sh"
  confirm "Install NetBird now?" || die "Aborted. Install NetBird yourself (https://netbird.io) and re-run."
  curl -fsSL https://pkgs.netbird.io/install.sh | sh
  command -v netbird >/dev/null 2>&1 || die "NetBird installation did not put 'netbird' on PATH."
  ok "netbird installed"
fi
echo

# 2. Join the mesh
bold "2/6  Join the GarageAI mesh"
MESH_IP="$(mesh_ip)"
if [ -n "$MESH_IP" ] && [ -z "$SETUP_KEY" ]; then
  ok "Already on the mesh (no setup key given, keeping the current connection)"
else
  [ -n "$SETUP_KEY" ] || die "No setup key. Pass --setup-key (you get one from GarageAI)."
  [ -n "$MANAGEMENT_URL" ] || die "No management URL. Pass --management-url (you get it from GarageAI)."
  # The setup key goes in the environment, not on the command line, so it does
  # not show up in the process list. The peer is named after the node, so the
  # gateway sees "garage-lund" rather than the machine's hostname.
  as_root env NB_SETUP_KEY="$SETUP_KEY" netbird up --management-url "$MANAGEMENT_URL" \
    --hostname "$NODE_NAME" >/dev/null
  waited=0
  MESH_IP="$(mesh_ip)"
  while [ -z "$MESH_IP" ] && [ "$waited" -lt "$MESH_WAIT_SECONDS" ]; do
    sleep 1; waited=$((waited + 1)); MESH_IP="$(mesh_ip)"
  done
fi
[ -n "$MESH_IP" ] || die "Could not get a mesh IP. Check 'netbird status' and that the setup key is valid."
ok "Mesh IP: ${MESH_IP}"
echo

# 3. Local runtime
bold "3/6  Inference runtime (${RUNTIME}, port ${PORT})"
if LOCAL_MODELS="$(http_models 127.0.0.1)"; then
  ok "OpenAI-compatible API answers on 127.0.0.1:${PORT}"
elif LOCAL_MODELS="$(http_models "$MESH_IP")"; then
  ok "OpenAI-compatible API answers on ${MESH_IP}:${PORT}"
else
  warn "No OpenAI-compatible API answered on port ${PORT}."
  runtime_hint "$MESH_IP"
  die "Start your runtime and run this script again."
fi
[ -n "$LOCAL_MODELS" ] || die "The runtime answered but lists no models. Load or pull a model first."
echo

# 4. Reachable over the mesh
# NetBird filters a node's traffic to its own mesh IP, so on many systems (macOS in
# particular) we cannot test the mesh path from here. Try it, and otherwise check which
# address the runtime listens on; the gateway's acceptance test then proves the path.
bold "4/6  Reachable over the mesh"
PROBE_HOST="$MESH_IP"
if MODELS="$(http_models "$MESH_IP" 2>/dev/null)"; then
  ok "Reachable on ${MESH_IP}:${PORT}"
else
  LISTEN="$(listen_addrs)"
  if printf '%s\n' "$LISTEN" | grep -qxE "\\*|0\\.0\\.0\\.0|\\[::\\]|::|${MESH_IP//./\\.}"; then
    PROBE_HOST=127.0.0.1
    printf '%s\n' "$LISTEN" | grep -qx "$MESH_IP" && PROBE_HOST="$MESH_IP"
    MODELS="$(http_models "$PROBE_HOST")" || die "The runtime stopped answering on ${PROBE_HOST}:${PORT}."
    ok "Listening on $(printf '%s' "$LISTEN" | tr '\n' ' ')(port ${PORT}); the gateway verifies the mesh path next"
    if [ "$RUNTIME" = ollama ] && [ "$(uname -s)" = Darwin ] && [ "$(id -u)" -ne 0 ] &&
       [ ! -e "$HOME/Library/LaunchAgents/eu.garageai.ollama-host.plist" ] &&
       confirm "Keep Ollama listening on the network after a reboot (adds a small login item that sets OLLAMA_HOST)?"; then
      install_ollama_env_agent
      ok "Done. OLLAMA_HOST is set at every login from now on."
    fi
  else
    warn "The runtime only listens on ${LISTEN:-127.0.0.1}, so the gateway cannot reach it."
    if [ "$RUNTIME" = ollama ] && [ "$(uname -s)" = Darwin ] && [ "$(id -u)" -ne 0 ] &&
       confirm "Make Ollama listen on the network permanently (a small login item that sets OLLAMA_HOST)?"; then
      install_ollama_env_agent
      ok "Done. Now restart Ollama so it picks this up, then run this script again:"
      if command -v brew >/dev/null 2>&1 && brew services list 2>/dev/null | grep -q '^ollama '; then
        info "  brew services restart ollama"
      else
        info "  quit Ollama from the menu bar and open it again"
      fi
      exit 0
    fi
    runtime_hint "0.0.0.0"
    info "0.0.0.0 makes it reachable over the mesh. Devices on your own LAN can reach it too;"
    info "  nothing on the internet can, unless your router forwards port ${PORT}."
    die "Restart the runtime bound to the mesh and run this script again."
  fi
fi
ALL_MODELS="$MODELS"
MODELS="$(printf '%s\n' "$ALL_MODELS" | select_models)"
printf '%s\n' "$ALL_MODELS" | while IFS= read -r m; do
  [ -n "$m" ] || continue
  if printf '%s\n' "$MODELS" | grep -qxF -- "$m"; then info "model: $m"
  elif [ -z "$OFFER_MODELS" ]; then info "model: $m (skipped: embedding/reranker model)"
  else info "model: $m (not offered)"; fi
done
[ -n "$MODELS" ] || die "No chat model to offer. Load one in ${RUNTIME} (or check --models) and run this script again."

# Buyers pay per token, so the runtime must report usage in streamed replies.
FIRST_MODEL="$(printf '%s\n' "$MODELS" | head -n 1)"
usage_auth=()
[ -n "$RUNTIME_API_KEY" ] && usage_auth=(-H "Authorization: Bearer ${RUNTIME_API_KEY}")
if curl -fsS -N --max-time 120 ${usage_auth[@]+"${usage_auth[@]}"} \
     -H "Content-Type: application/json" "http://${PROBE_HOST}:${PORT}/v1/chat/completions" \
     -d "$(jq -nc --arg m "$FIRST_MODEL" '{model: $m, max_tokens: 1, stream: true,
           stream_options: {include_usage: true}, messages: [{role: "user", content: "hi"}]}')" \
     2>/dev/null | grep -q '"usage" *: *{'; then
  ok "Token usage is reported (needed for per-token billing)"
else
  warn "No token usage in the streamed reply from ${FIRST_MODEL}. Billing may be incomplete;"
  warn "  check that the runtime supports stream_options.include_usage."
fi
echo

# 5. Register
bold "5/6  Register with GarageAI"
MODELS_JSON="$(printf '%s\n' "$MODELS" | jq -R . | jq -sc .)"
PAYLOAD="$(jq -nc \
  --arg name "$NODE_NAME" --arg mesh_ip "$MESH_IP" --argjson port "$PORT" \
  --arg runtime "$RUNTIME" --argjson models "$MODELS_JSON" --arg runtime_api_key "$RUNTIME_API_KEY" \
  '{name: $name, mesh_ip: $mesh_ip, port: $port, runtime: $runtime, models: $models}
   + (if $runtime_api_key != "" then {runtime_api_key: $runtime_api_key} else {} end)')"

if [ -n "$REGISTER_URL" ]; then
  [ -n "$REGISTER_TOKEN" ] || die "--register-url given without --register-token."
  info "Registering and running the acceptance test (a real request through the gateway)..."
  # register-node tests every model end to end before it is sold; this can take a minute.
  RESPONSE="$(curl -fsS --max-time 180 -X POST "$REGISTER_URL" \
    -H "Authorization: Bearer ${REGISTER_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD")" || die "Registration request to ${REGISTER_URL} failed."
  PASSED=0
  while IFS=$'\t' read -r model passed tps ttft err; do
    [ -n "$model" ] || continue
    if [ "$passed" = "true" ]; then
      PASSED=$((PASSED + 1))
      ok "${model}: passed (${tps} tok/s, first token after ${ttft} ms)"
    else
      warn "${model}: failed (${err})"
    fi
  done < <(printf '%s' "$RESPONSE" | jq -r '.acceptance[]? |
    [.model, (.passed|tostring), (.tokens_per_second // "?" | tostring),
     (.ttft_ms // "?" | tostring), (.error // "")] | @tsv')
  if [ "$PASSED" -gt 0 ]; then
    ok "Node registered — your garage is live on GarageAI."
  else
    warn "Registered, but no model passed the acceptance test, so nothing is for sale yet."
    warn "Check that the runtime answers on the mesh IP and that the model loads, then run this again."
  fi

  if [ "$HEARTBEAT" -eq 1 ]; then
    bold "6/6  Heartbeat"
    install_heartbeat
    ok "Installed: reports your models every 5 minutes. Load a new model and it shows up"
    info "  under My garages in the portal, where you choose to offer it. Remove with: $0 --remove-heartbeat"
  fi
else
  info "No --register-url given. Send these details to GarageAI to activate the node:"
  echo
  echo "$PAYLOAD" | jq 'del(.runtime_api_key)'
  echo
  info "Admin command (run by GarageAI on the gateway):"
  key_hint=""
  [ -n "$RUNTIME_API_KEY" ] && key_hint="NODE_API_KEY=<runtime API key> "
  # Word-splitting is intended: one argument per model id.
  # shellcheck disable=SC2086,SC2116
  info "  ${key_hint}infra/gateway/register-node.sh add ${NODE_NAME} ${MESH_IP} ${PORT} $(echo $MODELS)"
fi

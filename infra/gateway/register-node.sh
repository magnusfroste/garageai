#!/usr/bin/env bash
# register-node.sh — add or remove a garage node's models in the LiteLLM gateway.
#
# Each model is registered twice, which gives the two marketplace tiers:
#   dedicated  model_name "garage/<node>/<model>"  → only this garage (booked access)
#   pool       model_name "<model>"                → LiteLLM balances across every
#                                                    garage offering the same model
# Disable the pool entry with POOL=0.
#
# Usage:
#   register-node.sh add    <node> <mesh-ip> <port> <model> [<model> ...]
#   register-node.sh remove <node> <model> [<model> ...]
#   register-node.sh list
#
# Environment:
#   LITELLM_URL         Gateway base URL, e.g. https://llm.example.eu
#   LITELLM_MASTER_KEY  LiteLLM master key (keep it on the gateway side only)
#   POOL                1 (default) to also register the shared pool entry, 0 to skip
#
# LiteLLM must run with STORE_MODEL_IN_DB=True so models can be changed at runtime.

set -euo pipefail

: "${LITELLM_URL:?Set LITELLM_URL, e.g. https://llm.example.eu}"
: "${LITELLM_MASTER_KEY:?Set LITELLM_MASTER_KEY}"
POOL="${POOL:-1}"
LITELLM_URL="${LITELLM_URL%/}"

usage() { sed -n '2,22p' "$0" | sed 's/^# \{0,1\}//'; exit "${1:-0}"; }

api() {
  # api METHOD PATH [JSON]
  local method="$1" path="$2" body="${3:-}"
  if [ -n "$body" ]; then
    curl -fsS --max-time 20 -X "$method" "${LITELLM_URL}${path}" \
      -H "Authorization: Bearer ${LITELLM_MASTER_KEY}" \
      -H "Content-Type: application/json" -d "$body"
  else
    curl -fsS --max-time 20 -X "$method" "${LITELLM_URL}${path}" \
      -H "Authorization: Bearer ${LITELLM_MASTER_KEY}"
  fi
}

# Stable deployment id: <node>__<model>__<tier>, with characters LiteLLM ids and
# shells handle badly (/ : spaces) replaced. Model ids like "qwen3:32b" or
# "Qwen/Qwen3-32B" are common.
deployment_id() {
  printf '%s__%s__%s' "$1" "$(printf '%s' "$2" | tr -c 'A-Za-z0-9._-' '-')" "$3"
}

add_deployment() {
  local node="$1" ip="$2" port="$3" model="$4" tier="$5" model_name="$6"
  local id body
  id="$(deployment_id "$node" "$model" "$tier")"
  body="$(jq -nc \
    --arg model_name "$model_name" \
    --arg model "openai/${model}" \
    --arg api_base "http://${ip}:${port}/v1" \
    --arg id "$id" \
    '{model_name: $model_name,
      litellm_params: {model: $model, api_base: $api_base, api_key: "garage-node"},
      model_info: {id: $id}}')"
  # Re-registering after a node changes IP: drop the old entry first.
  api POST /model/delete "$(jq -nc --arg id "$id" '{id: $id}')" >/dev/null 2>&1 || true
  api POST /model/new "$body" >/dev/null
  printf '  ✓ %-9s %-40s → %s\n' "$tier" "$model_name" "http://${ip}:${port}/v1"
}

remove_deployment() {
  local id
  id="$(deployment_id "$1" "$2" "$3")"
  if api POST /model/delete "$(jq -nc --arg id "$id" '{id: $id}')" >/dev/null 2>&1; then
    printf '  ✓ removed %s\n' "$id"
  else
    printf '  - %s was not registered\n' "$id"
  fi
}

cmd="${1:-}"
[ $# -gt 0 ] && shift
case "$cmd" in
  add)
    [ $# -ge 4 ] || usage 1
    node="$1" ip="$2" port="$3"; shift 3
    for model in "$@"; do
      add_deployment "$node" "$ip" "$port" "$model" dedicated "garage/${node}/${model}"
      if [ "$POOL" = "1" ]; then
        add_deployment "$node" "$ip" "$port" "$model" pool "$model"
      fi
    done ;;
  remove)
    [ $# -ge 2 ] || usage 1
    node="$1"; shift
    for model in "$@"; do
      remove_deployment "$node" "$model" dedicated
      remove_deployment "$node" "$model" pool
    done ;;
  list)
    api GET /model/info | jq -r '
      .data[]
      | select((.model_info.id // "") | test("__(dedicated|pool)$"))
      | "\(.model_name)\t\(.litellm_params.api_base // "")"' ;;
  -h|--help|"") usage 0 ;;
  *) usage 1 ;;
esac

# Keys API for customer accounts

Live since 2026-11-25. A customer account (the first is Shipholm) creates and manages its own keys from its
server. The shape follows OpenRouter's key management API, so one client works against both by changing
the base URL. Amounts are in **credits** (1 credit = 1 USD today; treat it as a rate, not an identity).

Base URL: `https://api.garageai.eu`

## Two kinds of key

| | Management key `gai-mgmt-…` | Normal key `sk-…` |
|---|---|---|
| Does | creates, lists, changes, disables, deletes the account's keys; reads usage and balance | calls models |
| Calls models | no | yes |
| Sees | only its own account | only itself (`GET /v1/key`) |
| Where it is made | the portal, signed in as the account: Keys → Management key (shown once, optional expiry, rotate, revoke; at most 10 per account) | by the API (`POST /v1/keys`) or the portal |
| Where it lives | only on the customer's server, encrypted | the app or service that calls models |

Send either as `Authorization: Bearer <key>`.

## Endpoints

| Endpoint | Key | Answer |
|---|---|---|
| `GET /v1/keys` | management | `{"data": [Key…]}`, newest first, 100 per page; query `offset`, `include_disabled=true` |
| `POST /v1/keys` | management | 201 `{"data": Key, "key": "sk-…"}`; the full key is shown only here |
| `GET /v1/keys/{hash}` | management | `{"data": Key}` |
| `PATCH /v1/keys/{hash}` | management | `{"data": Key}` |
| `DELETE /v1/keys/{hash}` | management | `{"deleted": true}`; the key stops working at once |
| `GET /v1/keys/{hash}/activity?days=30` | management | `{"data": [{date, model, requests, prompt_tokens, completion_tokens, cached_tokens, cost}]}`, per UTC day and model, at most 30 days |
| `GET /v1/credits` | management | `{"data": {"total_credits", "total_usage"}}` for the whole account; balance = total_credits − total_usage |
| `GET /v1/key` | the normal key itself | `{"data": {label, limit, limit_remaining, limit_reset, usage, usage_daily, usage_weekly, usage_monthly, disabled}}` |

### Create and change

`POST /v1/keys` body (only `name` is required):

```json
{
  "name": "studio-acme",
  "limit": 10,
  "limit_reset": "monthly",
  "models": ["garageai/auto"],
  "rpm": 60,
  "metadata": {"tenant": "acme", "app": "studio"}
}
```

| Field | Rule |
|---|---|
| `name` | 1–80 characters |
| `limit` | credits ≥ 0, or `null` for no limit of its own (the account's balance still applies) |
| `limit_reset` | `"daily"`, `"weekly"`, `"monthly"` or `null`; the key's usage starts over each period |
| `models` | at most 50 names the catalogue knows: `garageai/auto`, a model name such as `deepseek/deepseek-v4.1-flash`, or one garage's `garage/<garage>/<model>`. Omitted = all models |
| `rpm` | requests per minute, 1–10000; omitted = the default (120) |
| `metadata` | any JSON object up to 4000 characters, returned on the key |

`PATCH /v1/keys/{hash}` takes any of these fields plus `"disabled": true | false`. Disabling blocks the key in
the gateway at once and `false` turns it back on; nothing else about the key changes.

### Key object

```json
{
  "hash": "…", "name": "studio-acme", "label": "sk-…AbCd", "disabled": false,
  "limit": 10, "limit_remaining": 9.2, "limit_reset": "monthly",
  "usage": 0.8, "usage_daily": 0.1, "usage_weekly": 0.5, "usage_monthly": 0.8,
  "created_at": "…", "updated_at": "…",
  "models": ["garageai/auto"], "rpm": 60, "metadata": {"tenant": "acme", "app": "studio"}
}
```

`usage` is the key's spend in the current limit period (since creation when there is no `limit_reset`).
`usage_daily` / `usage_weekly` / `usage_monthly` cover the last 1, 7 and 30 UTC days. `models`, `rpm` and
`metadata` are GarageAI's additions to OpenRouter's object. Chat keys made by the portal's own chat are not
listed.

## Calling models: cost per request

On `POST /v1/chat/completions` (and the other model paths) with a normal key:

- **Non-streamed:** `usage.cost` in the answer, in credits.
- **Streamed:** send `"stream_options": {"include_usage": true}`; the last chunk's `usage` carries `cost`.
- **Headers:** `x-litellm-response-cost` (final for non-streamed answers only), `x-litellm-key-spend`,
  `x-litellm-key-max-budget`, `x-litellm-call-id`.
- **Reconciliation:** store the answer's `id`. It is the id of the request's row in the gateway's ledger.
  Keep `x-litellm-call-id` next to it for support.

Usage figures in the keys API are built from the gateway's ledger and lag by a minute or so. For a figure at
once, add up `usage.cost` from the answers.

## Errors

| Code | When | Retry? |
|---|---|---|
| 400 | bad input (`{"error": {"code": 400, "message": "…"}}`), or the account already has 100 keys | no, fix the request |
| 401 | management calls: `Invalid management key`. Model calls: `Invalid API key.` (wrong or deleted key), `This key is disabled.` (PATCH `disabled: true`), `This key has expired.` | no; a disabled key works again after PATCH `disabled: false` |
| 402 | on a model call: the key's limit or the account's balance is spent (`"type": "budget_exceeded"`) | no, final until the limit resets or credits are bought |
| 404 | the hash is not a key of this account | no |
| 429 | more than 60 management calls per minute per management key, or the key's rpm on model calls | yes, after a short wait |
| 503 | on a model call: no garage can take it now; has `Retry-After` (usually 60 s) | yes, after `Retry-After` |
| 502 | the gateway could not be reached for a management call | yes, with backoff |

## Good practice

- One normal key per account or app, so one can be disabled without touching the others.
- Give each key a `limit` with a `limit_reset`; watch `GET /v1/credits` and alert before the account runs dry,
  because an empty account stops every key at once.
- Lock keys for apps with AI built in to the named models they need (`models`); for data that must stay in
  Sweden, use garages in Sweden, not `garageai/auto`.
- Rotate the management key before it expires: make a new one in the portal, switch over, revoke the old one.

## How it is built (for GarageAI)

- Traefik router `customer-keys` on the portal proxy (`infra/gateway/portal-proxy/docker-compose.yml`) sends
  `/v1/keys…`, `/v1/key`, `/v1/credits` on api.garageai.eu to the portal function `customer-keys`. Model
  paths stay on LiteLLM. On llm.garageai.eu these paths answer 403.
- The portal function (Lovable, migration `0103_customer_keys_api`) checks the management key by its hash,
  limits every lookup to that key's account, maps the calls to LiteLLM `/key/generate`, `/key/update`
  (`blocked`, `max_budget`, `budget_duration` 1d/7d/30d, `models`, `rpm_limit`), `/key/delete`, `/key/info`
  and `/user/daily/activity`, and writes every create, change, disable and delete to `customer_key_events`.
- The gateway (`litellm/config.yaml`, `garageai_callbacks.py`): `include_cost_in_streaming_usage`,
  `budget_exceeded_status_code: 402`, `add_usage_cost`; Traefik lets the four headers above through.

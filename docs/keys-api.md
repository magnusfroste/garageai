# Keys API for customer accounts (spec, agreed 2026-11-25; not built yet)

First customer: Shipholm (own GarageAI account, not Magnus's admin account). Shape copied from
OpenRouter's key management API so a client works against both by changing the base URL. Agreed
with Shipholm 2026-11-25.

## Endpoints, all on `api.garageai.eu`

Traefik sends these paths to a portal edge function. The URL then stays the same if the portal leaves
Lovable.

| Endpoint | Auth | Does |
|---|---|---|
| `GET /v1/keys` | management key | list the account's keys: id (hash), name, label, limit, limit_reset, usage (total, daily, weekly, monthly), disabled, models, rpm, metadata, created_at |
| `POST /v1/keys` | management key | create: name, limit, limit_reset (daily / weekly / monthly / none), models[], rpm, metadata {tenant, app}; the key value is returned once |
| `GET /v1/keys/{hash}` | management key | one key |
| `PATCH /v1/keys/{hash}` | management key | change name, limit, limit_reset, models, rpm, metadata, `disabled: true/false` |
| `DELETE /v1/keys/{hash}` | management key | delete |
| `GET /v1/keys/{hash}/activity` | management key | per day (optionally per hour), 30 days: model, requests, tokens in/out, cached, cost |
| `GET /v1/key` | the key itself | its own limit, usage and limit_remaining |
| `GET /v1/credits` | management key or any of the account's keys | account balance: total_credits, total_usage |

**Extensions beyond OpenRouter:** `models` and `rpm` per key. A key for an app with AI built in can
then be locked to one named model, for example a Swedish garage model for school apps (never auto).

## Management key

- One per account, created in the portal (Keys → Management key).
- It cannot call models; the gateway rejects it on /v1/chat etc.
- It has an expiry date and can be rotated.
- It only ever sees its own account's keys.

## Gateway side (my part)

1. **Streamed responses:** the last chunk's `usage` gets `cost`, in credits, like OpenRouter's
   `usage.cost`, when `stream_options.include_usage` is set. Non-streamed responses get
   `usage.cost` too.
2. **Error when a key's limit is reached:** HTTP **402** instead of LiteLLM's default 422. The body
   stays LiteLLM's: `{"error": {"message": "Budget has been exceeded! ...", "type": "budget_exceeded",
   "code": "402"}}`. The empty account balance gets the same treatment. Documented as
   final, not to be retried.
3. Traefik routes for `/v1/keys*`, `/v1/key` and `/v1/credits` to the edge function.
4. Response headers for buyers: `x-litellm-call-id`, `x-litellm-response-cost`, `x-litellm-key-spend` and
   `x-litellm-key-max-budget`. Until this change only the budget header reached buyers.

Items 1, 2 and 4: PR #272, which needs a LiteLLM restart in a quiet window.

## Portal side (Lovable)

The edge function behind the routes:
- It maps to LiteLLM `/key/generate`, `/key/update`, `/key/block` and `/key/unblock`, `/key/delete`,
  and `/user/daily/activity`, via the portal key. The Traefik list may need `/key/block`,
  `/key/unblock` and `/key/list`.
- `limit` and `limit_reset` map to LiteLLM `max_budget` and `budget_duration` (1d / 7d / 30d).
- The api_keys table stores metadata and only the last 4 characters of each key.
- It writes an audit trail: who created, changed or disabled a key, and when.

UI: Keys → Management key (create, show once, rotate, expiry).

## Until then

- Manual keys in the portal, one per active Studio account.
- Studio stores `x-litellm-call-id` per turn.
- No per-app keys until the API exists.

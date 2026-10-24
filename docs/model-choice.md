# Choosing between garages: what differs, shown and requestable

Spec, 2026-10-24. Status: decided, not built.

## Why

A provider such as Chutes serves a model the same way everywhere. On GarageAI the same model name
can run on an RTX 5090 at Q8 with a 64k window in one home, and on a MacBook at Q4 with an 8k
window in another. That difference is our weakness if it is hidden, and our strength if it is
visible and steerable: buyers can pick what they need, and good garages are seen to be good.

Most buyers do not browse; their agents call the API. So there are two parts: the app **shows**
the differences and filters on a few of them (step 1), and the API **accepts requirements** in the
request, which the gateway enforces while routing (step 2). Step 2 is where agents benefit: Hermes
can ask for at least 128k context and never land on a garage that cannot take its prompt.

## What differs per garage and model

| Property | Source today | Gap |
|---|---|---|
| Context window | heartbeat `contexts` → `max_input_tokens` | none |
| Speed: tokens/s and time to first token | acceptance test (`garage_model_tests`), gateway prefill measurements (gateway rule 6) | decode tokens/s should also come from real traffic, not only the test |
| Tool calling | acceptance test probe (`supports_tools`) | none |
| Quantization (Q4_K_M, Q8_0, FP8, BF16, …) | not stored | parse from the runtime model id (Ollama tag `…-q8_0`, GGUF file `…-Q4_K_M.gguf`, vLLM names `…-FP8`); **unknown when not stated, never guessed** |
| Location | `garages.measured_country` (ISO code only) | none |
| Garage or provider | `connection_type` (`mesh` vs `endpoint`) | none |
| Reliability | `garage_reliability` grade | none |
| Price | the garage's prices | none |
| Energy (tokens per kWh) | Bridge v1.3.0 reports `agent.energy` | stored and aggregated per garage and model (portal) |

## Step 1: show the differences and filter on a few (portal, Lovable)

**Model page: one row per garage serving the model**, like OpenRouter's provider table but with
our own columns:

| Garage | Country | Tokens/s | First token | Context | Quantization | Tools | Uptime 7 d | Price in/out | Energy |
|---|---|---|---|---|---|---|---|---|---|

- **Tokens/s and first token:** the median of the last 7 days of real requests where there are
  at least 10, else the latest acceptance test, labelled "test".
- **Quantization:** the parsed value, or "unknown".
- **Energy:** tokens per kWh when Bridge reports it. It is empty for garages without a power
  reading, never 0.
- **Providers** show "datacenter" in the Energy column and are marked as providers.
- **Garages offline right now** stay in the table, greyed, with "offline". This shows that
  supply exists.

**Catalogue filters: four, no more.** With a dozen garages, more filters mostly produce empty
results.

1. **Context at least:** 8k, 32k, 128k or 256k.
2. **Tool calling.**
3. **Speed at least:** 10, 30 or 60 tokens/s, by the best garage for the model.
4. **Where it runs:** Anywhere, EU only, or Garages only (no datacenters).

The filters apply to models: a model is listed when at least one garage meets them. The model
page then highlights the garages that meet them. The choice is kept in the URL so it can be
shared.

**Data the portal adds:**
- `garage_models.quantization`, parsed at registration and heartbeat. There is a shared parser
  with tests for the Ollama, GGUF and vLLM name forms.
- A daily aggregate of `agent.energy` per garage and model: Wh, busy Wh, output tokens, tokens
  per kWh.
- The model-page fields above, through the public garage RPCs. These RPCs already exclude
  private garages.

## Step 2: requirements in the API (gateway)

An optional object in the request body. OpenAI SDKs send it with `extra_body`. Every field is
optional.

```json
{
  "model": "garageai/auto",
  "messages": [...],
  "garageai": {
    "min_context": 131072,
    "min_tps": 30,
    "max_ttft_ms": 5000,
    "tools": true,
    "location": ["EU"],
    "supply": "garages",
    "quantization": ["q8", "fp8", "bf16"]
  }
}
```

| Field | Meaning |
|---|---|
| `min_context` | The deployment's window (`max_input_tokens`) must be at least this |
| `min_tps` | The deployment's measured decode tokens/s must be at least this. A deployment with no measurement does not qualify. |
| `max_ttft_ms` | The measured median time to first token must be at most this |
| `tools` | Only deployments whose model passed the tool probe |
| `location` | ISO country codes, or `"EU"` (the EU and EEA). Deployments with no known country do not qualify. |
| `supply` | `garages` (connection_type `mesh`), `providers` (`endpoint`) or `any` (default) |
| `quantization` | Allowed families: `q4`, `q5`, `q6`, `q8`, `fp8`, `bf16`, `fp16`. `unknown` does not qualify unless listed. |

**Rules:**
1. **The gateway removes the `garageai` object from the request** in the pre-call hook, before
   anything reaches a runtime, and keeps it in the request's metadata. It is never forwarded.
2. **Filtering happens in `async_filter_deployments`,** before gateway rules 6 (time to first
   token), 7 (stickiness) and 10 (owner first). A sticky deployment that fails a requirement is
   not used. The requirements always win.
3. **Data comes from `model_info`,** which the portal sets on every market deployment:
   `max_input_tokens` (exists), `connection_type` (exists), `country`, `quantization`,
   `supports_tools`, `tps_p50` and `ttft_p50_ms`. The portal refreshes the speed fields daily.
   The gateway reads, it does not measure.
4. **When nothing qualifies,** the answer is `400` with code `no_deployment_meets_requirements`.
   The message names the requirement that excluded the most candidates, for example "min_context
   131072: the largest window available for qwen/qwen3.8-27b right now is 65536". It is a 400 and
   not a 503, because retrying will not help unless the requirement changes.
5. **Unknown fields are refused with 400.** A typo must not be silently ignored.
6. **Logging:** one gateway log line per filtered request (the model and the requirements,
   never the prompt), and Ops counts requests by requirement and how many were refused.

**Documentation:** the API docs page gets a section "Choose where your request runs", with the
table above and examples for the OpenAI Python and JavaScript SDKs (`extra_body`), curl, and
Hermes and Claude Code configuration.

## Order and acceptance

Order:
1. Gateway step 2 with the fields that already exist: `min_context`, `supply`, `tools` (once
   the portal copies `supports_tools`) and `location` (once it copies `country`).
2. The portal's new `model_info` fields and the quantization parser.
3. The model-page table and the catalogue filters.
4. `min_tps`, `max_ttft_ms` and energy, as the measurements land.

Acceptance:
1. `garageai/auto` with `min_context: 200000` never routes to a deployment with a smaller window.
   With none available, the answer is 400 `no_deployment_meets_requirements` naming the largest
   window.
2. `supply: "garages"` never reaches a provider, and `supply: "providers"` never reaches a
   garage, checked over 50 requests.
3. `location: ["EU"]` never routes to a deployment whose country is outside the EU and EEA or
   unknown.
4. The `garageai` object is not present in what the runtime receives, checked with a mock
   deployment that echoes its request.
5. An unknown field (`min_contxt`) gets 400.
6. The model page lists each garage with country, speed, context, quantization ("unknown" where
   not parsed), tools, uptime, price and energy. The four catalogue filters narrow the list and
   survive a reload through the URL.
7. The quantization parser handles `qwen3-coder:30b-a3b-q8_0`, `granite-4.1-3b-Q8_0.gguf`,
   `Qwen3-30B-A3B-…-Q4_K_M.gguf`, `…-FP8` and `qwen3:4b` (unknown).

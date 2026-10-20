# Owner first: your machine is yours before it is the market's

Spec, 2026-10-20. Status: decided, not built. Point 1 of the owner-at-home review (the others are
under "Later" in [my-machine.md](my-machine.md) and in the roadmap).

## Why

My machine (`mine/<garage>/<model>`) lets an owner use their own garage from anywhere. But the
same runtime also serves the market, and today the gateway treats both alike. If a buyer's
160k-token prompt is being prefilled on the owner's DGX, the owner's "hi" waits behind it. On
Ollama and Runner, which serve one request at a time unless set up otherwise, the owner can wait
for minutes. The more a garage sells, the worse it gets for the person who owns it. That is the
opposite of the promise.

Owner first is the rule that fixes it: **while the owner is using their machine, the market
waits, not the owner.** It is also something to say out loud: no cloud chat can promise that the
hardware is yours first.

## What the owner sees

- In My garages, each garage has a setting **"When I use this machine myself"**:
  - **Pause the market while I use it** (the default). One line: "Buyers go to other garages while
    you chat, and come back a few minutes after your last message."
  - **Share it with the market.** One line: "You and buyers take turns; you may wait behind a
    buyer's long request."
- In the chat, when the owner's request had to wait for a buyer's request that was already
  running, the answer's footer says so: "waited 40 s for a buyer's request to finish". That way a
  slow answer is explained, not mysterious.
- Earnings show the cost honestly: My garages counts "paused for you" minutes per day next to the
  earnings.

## Rules

1. **Owner activity.** A garage is *owner-active* from the moment its owner sends a request to one
   of its `mine/` models until **5 minutes** after the owner's last request ends. The window
   slides with every new request. The length can be set per garage, from 1 to 30 minutes. The
   default is 5, which covers the time to read an answer and type the next question.
2. **Hold the market.** While a garage is owner-active and its setting is "Pause the market":
   - **New** market requests are not routed to that garage. This covers pools, `garageai/auto`
     and every other non-`mine/` model group. Other deployments take them as usual, and rule 6 of
     the gateway (time to first token) and `order` apply among the rest.
   - A buyer who asked for this garage specifically (`garage/<g>/<model>`) and has no other
     deployment gets `503` with `Retry-After` set to the remaining hold. The message says the
     garage is busy for a moment, not that it is gone. This is the same contract as a garage that
     is away.
   - Requests **already running** are never cut off. The gateway cannot take back a request a
     runtime is processing, and breaking a buyer's stream would be worse than a short wait.
3. **Own requests first at the runtime, where the runtime allows it.**
   - On **vLLM** started with `--scheduling-policy priority`, the gateway sends the owner's
     requests with `priority: -10` and market requests with `0`, so vLLM schedules the owner first
     even when a buyer's request is queued. Bridge's doctor recommends the flag on vLLM garages
     with "Pause the market".
   - **Ollama, LM Studio, llama.cpp and Runner** have no priorities. Rule 2 is what helps there,
     and doctor recommends a second parallel slot (`OLLAMA_NUM_PARALLEL=2`, llama.cpp `--parallel
     2`, Runner `--parallel 2` for a single model), so that the owner's request does not queue
     behind one long buyer request.
4. **Stickiness gives way.** A conversation that rule 7 of the gateway keeps on a garage that has
   become owner-active is routed elsewhere, as if the garage were unhealthy. Owner first wins over
   a warm cache.
5. **Only the owner counts.** Only requests to `mine/` models count as owner activity, and only
   from the owner's own keys (rule 8 already guarantees both). An owner who buys from their own
   garage through the market is a buyer like any other.
6. **Never a way to game the market.** A hold only takes a garage out of routing; it never adds
   traffic or earnings. Holds are logged, so a garage that is "owner-active" around the clock is
   visible in Ops.

## Implementation

| Part | Who | What |
|---|---|---|
| Owner activity | Gateway, `garageai_callbacks.py` | `async_pre_call_hook` on a `mine/` request marks the garage active until now + the garage's hold. `async_log_success_event` and `async_log_failure_event` move the end to request end + hold. The state lives in memory: the gateway is one process, and a restart simply ends every hold. |
| Hold | Gateway, `async_filter_deployments` | For any non-`mine/` model group, drop the deployments whose `model_info.garage` is owner-active and whose garage has `owner_priority: hold`. This runs before rules 6 and 7. If that leaves nothing, raise the 503 "busy for a moment" with `Retry-After` set to the remaining hold. Log one line per hold start and end: garage, minutes, market requests diverted. |
| Priority | Gateway, `async_pre_call_deployment_hook` | On a vLLM deployment (`model_info.runtime == "vllm"`) of a garage with the setting on, add `priority` to the request body: `-10` for `mine/` and `0` otherwise. Only send it when Bridge reports that vLLM was started with `--scheduling-policy priority` (heartbeat field, see Bridge). vLLM rejects the field otherwise. |
| Setting | Portal | `garages.owner_priority` (`hold` by default, or `share`) and `owner_hold_minutes` (default 5). Both are copied into `model_info` of the garage's market deployments, as `owner_priority` and `owner_hold_minutes`, so the gateway reads them without asking the portal. The setting and the "paused for you" minutes appear in My garages, and the chat footer note uses a response header from the gateway (below). Swedish strings. |
| Wait note | Gateway | When an owner's request is served by a garage that had market requests in flight at the time, the response carries `x-garageai-waited-for-market: <seconds>`, measured as the time to first token minus the garage's usual time to first token, only when it is above 5 s. Traefik passes it, and it is listed in CORS `Access-Control-Expose-Headers` for the chat. |
| Bridge | Bridge (`cli/`) | Doctor detects vLLM's `--scheduling-policy` and reports it in the heartbeat (`runtime_features: ["priority"]`). It recommends the flag on vLLM, and a second parallel slot on Ollama, llama.cpp and Runner, when the garage's setting is "Pause the market". |
| Ops | Ops collector | Per garage: hold minutes per day and market requests diverted. An alert when a garage is owner-active for more than 20 hours in a day. |

Order: gateway hold first (rules 1, 2, 4, 5), then the portal setting (until then every garage
counts as `hold` with 5 minutes), then vLLM priority with Bridge's report, then the wait note and
Ops.

## Acceptance

1. Magnus sends a message to `mine/dgx-city/...`. For the next 5 minutes, 50 `garageai/auto`
   requests go to the other auto entries (none to dgx-city), and the log shows one hold start.
   Five minutes after his last message, auto uses dgx-city again.
2. A buyer's `garage/dgx-city/xiaomi/mimo-v2.6-flash` during the hold gets 503 with a
   `Retry-After` of at most the remaining hold. After the hold, the same request succeeds.
3. A buyer request already streaming from dgx-city when Magnus writes finishes normally.
4. With the setting "Share it with the market", no hold happens.
5. On vLLM with `--scheduling-policy priority`: with one long buyer request queued, the owner's
   request starts first (its time to first token is about its usual, not the buyer's prefill on
   top).
6. Another account's market request to dgx-city does not start a hold. Magnus's request to the
   market route `garage/dgx-city/...` does not start one either.
7. A LiteLLM restart clears every hold, and nothing breaks.
8. Ops shows the hold minutes per garage for the day.

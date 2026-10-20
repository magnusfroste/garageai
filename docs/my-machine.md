# My machine: your own garage, private and free, from anywhere

Spec, 2026-10-20. Status: decided, not built. Point 1 of the chat review (the others are in
"Later" at the end of this document).

## Why

People who run local AI already have good chat apps at the desk: LM Studio, Ollama's app, Jan,
Open WebUI. What none of them gives without networking work (Tailscale, a tunnel, open ports) is
**the same machine from the phone or another computer, anywhere**. GarageAI already has that path:
Bridge's outgoing tunnel and the gateway. But today an owner reaches their own machine only as a
**buyer**. They pay credits to themselves, and only for models that passed the market's
acceptance test and are offered for sale. A garage paused in the portal cannot be reached at all.

"My machine" makes the owner's own garage a private model in the chat and the API: free, theirs
alone, available whether or not it is on the market. It is the reason a local-AI user picks
GarageAI over a do-it-yourself setup.

## What the owner sees

- In the chat's model picker, a section **Your machines** comes first. It has one row per model
  their garages serve, for example "garage-m4 · Granite 4.1 3B", with a state: *awake*, *asleep
  since 14:02* or *not answering*. There is no price, because it is free.
- In the API, the same models answer under `mine/<garage>/<model>`, with any of the owner's keys,
  and they appear in `/v1/models` for that owner only.
- In My garages, each garage has a choice **Who can use this machine**:
  - **Only me.** The garage is not on the market; the owner uses it privately.
  - **Me and the market**, the default. The garage is sold as today, and the owner also has the
    private, free route.
  - **Me, people I invite, and the market.** This is phase 2, with invitations; see Later.
- A garage that is paused in the portal stays usable by its owner. "Paused" means "off the
  market", not "off".

## Rules

1. **Private routes.** For every model a garage reports in its heartbeat, the gateway has a
   deployment `mine/<garage>/<model>`.
   - It needs no acceptance test and no Offer toggle. The owner's own machine is the owner's own
     business. It appears with the next heartbeat after the model does, and disappears when the
     model is gone.
   - Its context window is the one from the heartbeat (`contexts`), the same as the market route.
   - Fields: `model_info.garage_tier = "private"`, `operator_id` = the owner, and
     `connection_type = "mesh"`. `input_cost_per_token`, `output_cost_per_token` and
     `cache_read_input_token_cost` are all `0`.
2. **Only the owner.** The gateway refuses a `mine/...` request unless the key belongs to the
   garage's owner: the key's `user_id` must equal the deployment's `operator_id`. The refusal is
   `404 model_not_found`, the same as for a name that does not exist, so other people learn
   nothing. This check lives in `garageai_callbacks.py` (pre-call). It does not depend on the
   portal getting key permissions right.
3. **Hidden from everyone else.** `/v1/models`, the chat's model list and the public catalogue
   never show someone else's `mine/...` models. The gateway enforces this as well (see
   Implementation).
4. **Free, including with no credits.** Every price is 0, so LiteLLM skips its budget checks for
   these models (`_is_model_cost_zero`, LiteLLM 1.105). An owner whose balance is $0 can still use
   their own machine. Private use never produces earnings, which also closes the gap where an
   owner spends free starting credits on their own garage and earns them back.
5. **Paused and private garages stay reachable by their owner.** `gateway-targets` already keeps
   paused garages with `paused: true` (2026-10-19). The routing exclusion of paused garages must
   not apply to private routes. The health check keeps checking them. A runtime that is down
   still takes the private route out of routing exactly as today, and the chat then shows *not
   answering*.
6. **Privacy unchanged.** Prompts go buyer → gateway → tunnel → the owner's own machine. The
   gateway stores no prompts. Spend-log rows are kept, with tokens and a cost of 0, so the owner
   can see their own use. Ops counts private requests separately and never in revenue or in
   market quality figures.
7. **Never mixed into the market.** `mine/...` deployments are never members of `garageai/auto`,
   of a pool or of a dedicated market route. A market request can never land on a private
   deployment, and a private request can never land on someone else's garage.

## Implementation

| Part | Who | What |
|---|---|---|
| Deployments | Portal (edge functions that sync garages to LiteLLM) | Create, update and remove `mine/<garage>/<model>` from the heartbeat inventory per rule 1, independent of offer state, acceptance test and pause. Garage setting `visibility`: `private`, `market` (the default; the owner keeps the private route) or `shared` (phase 2). `private` removes the garage's market deployments and hides it from the catalogue. |
| Owner check | Gateway, `garageai_callbacks.py` | Rule 2 in `async_pre_call_hook`: the `mine/` prefix requires `user_api_key_dict.user_id == operator_id`, otherwise 404. Rule 7: the deployment filter drops any `garage_tier == "private"` deployment from every model group that is not `mine/...`. Tests in `tests/test_callbacks.py`. |
| Hidden listing | Gateway, `garageai_callbacks.py` | Built on LiteLLM 1.105's `async_filter_listed_models` hook, which narrows `/v1/models`, `/v1/models/{id}` (404), `/model/info` and `/model_group/info` per caller: a `mine/` name is listed only for a key whose `user_id` is one of its operators. If the filter fails, every private name is hidden. Access groups do not work for this: a key with `models: []` lists every model, access groups included (checked in `get_key_models`). |
| Health and routing | Gateway, health check | Paused garages (`paused: true`) keep their private routes in routing (rule 5). |
| Chat | Portal (Lovable, frontend) | The section "Your machines" first in the picker, with the state from the garage's latest heartbeat and health result. The awake, asleep and not-answering wording follows Bridge's own words: an *asleep since* time when the last heartbeat said `away`, or the machine reported battery or sleep. No price on these rows. |
| My garages | Portal | The choice "Who can use this machine" (Only me / Me and the market). A one-line explanation for each. Swedish strings. |
| Ops | Gateway, Ops collector | Private requests are counted per garage in their own column (requests, tokens). They are kept out of revenue, the buyer funnel and market quality. |

Order: owner check and filter in the gateway first, then the portal's deployments, then the
chat, then the visibility choice. The gateway part is safe to deploy before the portal creates
any `mine/` deployment.

## Acceptance

1. Magnus's key reaches `mine/garage-m4/ibm-granite/granite-4.1-3b` with a balance of $0. The
   spend-log row shows cost 0.
2. Another account's key gets `404 model_not_found` for the same name. That account's
   `/v1/models` does not list it.
3. `garageai/auto` and the market routes never select a private deployment (100 requests with
   only private deployments healthy give 503, never a private answer).
4. garage-m4 paused in the portal: the owner's private route still answers, and the market route
   does not.
5. Visibility "Only me": the garage disappears from the catalogue and the market within one
   heartbeat, and the private route keeps working.
6. A model loaded on the machine shows up in "Your machines" within one heartbeat (5 minutes)
   without an Offer toggle or acceptance test. A model removed from the machine disappears within
   one heartbeat.
7. The machine asleep: the chat row says *asleep since HH:MM* within 2 minutes, and sending to it
   gives a clear message, not a raw error.
8. Ops: private requests counted per garage, absent from revenue and from market quality.

## Later (from the same review, not in this spec)

- **2. Honest state and a safety net.** When the own machine does not answer, offer to continue
  on `garageai/auto`, clearly labelled ("answered by the network"). Every answer says where it
  ran (machine or network), the model and tokens/s.
- **3. Phone first.** An installable web app (PWA), fast start, voice input. The promise:
  "talk to the computer at home from the bus".
- **4. Chat history at home.** A third storage choice next to "this device" and "my account":
  Bridge stores the history on the owner's own machine, and the chat syncs there through the
  tunnel. Then neither prompts nor history leave the owner's hardware except in transit.
- **5. What Claude taught everyone to expect.** Projects with their own instructions, files (PDF,
  images for vision models), code and HTML previews in a side panel, web search.
- **Invitations (visibility "shared").** Share a machine with family or a team. The private route
  takes a list of allowed user ids, the invitees use it free (or at a price the owner sets), and
  the owner check in rule 2 accepts the list.

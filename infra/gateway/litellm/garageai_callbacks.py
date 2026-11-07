"""GarageAI's LiteLLM callbacks, loaded through `litellm_settings.callbacks` in config.yaml.

1. Tool schemas without "items". JSON Schema allows an array property without "items"
   (any item type), and Claude Code sends tool definitions like that. LiteLLM's token
   counter requires "items" and fails with KeyError('items'), after which the router's
   pre-call context check is silently skipped. Adding "items": {} means the same thing.

2. Our own context check, before routing. LiteLLM's check compares the prompt alone with
   the window, but runtimes count prompt + max_tokens (vLLM rejects 229k + 32k against a
   262k window). Counting is approximate (a generic tokenizer, not the model's own), so we
   keep a margin. A request that does not fit gets 400 at once, never a stream that breaks.

3. The error in the words agents act on. OpenAI-style clients (aider, LangChain, Cline)
   shorten the history when they see "maximum context length is N tokens" and code
   "context_length_exceeded"; Claude Code compacts or lowers max_tokens on Anthropic's
   "prompt is too long: N tokens > M maximum" and "input length and `max_tokens` exceed
   context limit". Upstream context errors that slip through are rewritten the same way,
   without LiteLLM's prefix or internal deployment names.

4. Response ids without internal names. On /v1/messages LiteLLM answers through its
   Responses bridge with an id that is base64 of "litellm:custom_llm_provider:...;
   model_id:<garage>__<runtime model>__<tier>;response_id:...", which tells a buyer which
   garage served them. Anthropic messages are not chained by id, so we replace it with a
   neutral "msg_" id derived from it (stable for the same response, reveals nothing).

5. System messages first, and only one. Many open models' chat templates (Qwen, among others)
   reject a system message anywhere but first ("System message must be at the beginning"),
   and some reject two. Agents send them later in a conversation (reminders, mode changes).
   When the order is irregular, every system/developer message is merged, in order, into a
   single system message at the start. A request that is already regular is not touched.
6. A garage gets only the prompts it can start on in time. A window says what fits, not how long
   reading it takes: dgx-city fits 524k tokens but prefills ~1,000 tokens/s, so a cold 160k-token
   prompt waited 2.5 min for its first token and agents gave up (agenthotels, 2026-10-18). Each
   deployment's prefill speed is measured from its own streamed requests (new tokens / time to
   first token); a deployment is skipped for a request when reading the prompt's *new* tokens
   would take longer than PREFER_S, as long as another can do it in time, and never sent one that
   could not start before the stream timeout cuts it (LIMIT_S); that buyer gets a quick 503 instead
   of a ten-minute wait. Measured 2026-10-20 (median, 3 days): DeepSeek provider 2,900 tokens/s,
   dgx-city 1,800, mac-mini-1cab 98. Faster garages get bigger prompts by themselves; nothing is a
   fixed cap. Providers are measured the same way; an unmeasured provider is not limited.

7. A conversation stays where its cache is. Runtimes keep the prompt prefix they just processed,
   so the next turn of the same conversation on the same deployment reads only the new tokens
   (dgx-city answered 188k-token turns in 1-11 s once warm). The gateway remembers which
   deployment served a conversation (the buyer's key + the first system and user message) and
   sends the next turn there for STICKY_S, when it is healthy and the new tokens fit rule 6. A
   failure forgets it, so the retry goes elsewhere.
8. My machine (docs/my-machine.md): `mine/<garage>/<model>` is the owner's own garage, private and
   free. Only the key of the garage's owner (user_id == the deployment's operator_id) may use or
   see such a name; anyone else gets 404 model_not_found, exactly as for a name that does not
   exist, and never finds it in /v1/models, /model/info or /model_group/info. The proxy admin
   (master key) sees every name in the listings, but may not use another owner's machine either. A private deployment
   (garage_tier "private") is never routed for any other model name, so a market request can never
   land on someone's private machine.
9. `garageai/auto` says which model answered. A buyer who asks for auto gets the canonical name of
   the model that served the request in the response's `model` field (for example
   `deepseek/deepseek-v4.1-flash`), streamed or not, as OpenRouter does for openrouter/auto. Never
   the garage: the name is the public catalogue name, which the catalogue already shows with its
   garages.
10. Owner first (docs/owner-priority.md): while an owner uses their machine through `mine/`, the
    market waits. From the owner's request until OWNER_HOLD minutes after their last one ends, new
    market requests skip that garage (others take them; a buyer of that garage alone gets 503 with
    Retry-After). Requests already running are never cut. A garage whose model_info says
    owner_priority "share" is not held. The hold lives in memory: a restart ends every hold.
11. A start for garages that wait (2026-11-04). `garageai/auto` is an ordered list with one garage per
    model, so a new garage that serves a model on that list (qwen3.8-27b on a home RTX 5090) got 6
    requests in two weeks while auto served 5,700: its operator sees nothing happen. Now a market
    garage that serves the same model as an auto entry, but is not itself on the list, and has
    served fewer than BOOST_TARGET market requests in the last 24 h, is put first for BOOST_SHARE of
    the small auto requests (prompt <= BOOST_MAX_PROMPT tokens, so an agent's 100k-token prompt never
    lands on an unmeasured home GPU), the one that has waited most first. Everything else still
    applies: its window (the router's pre-call check drops it when the prompt does not fit, and the
    request goes down the auto list as before), rule 6's prefill time, rule 10's owner hold. The next
    turn of a conversation it served goes back to it (rule 7). A failure takes the garage out of the
    start for BOOST_PAUSE_S, and the retry goes down the auto list. The buyer pays the garage's pool
    price and sees the model's name (rule 9). Counts live in memory: after a restart every garage
    counts as waiting until it has served again, which costs at most BOOST_SHARE of small requests.
    A request that caps its answer below BOOST_MIN_OUTPUT tokens gets no start (2026-11-05): the first
    garages run qwen3.8-27b with thinking on, which spent 242-400 tokens on a one-word answer, so a
    20-token cap came back empty (finish "length") where DeepSeek answered in 4.
12. The runtime's token count is checked (2026-11-05). LiteLLM bills from the `usage` the runtime
    reports (for a stream, its usage chunk wins), and the spend log keeps no answer text, so a
    modified runtime could report 100,000 completion tokens for a five-word answer: the buyer pays,
    the operator is credited, and nothing could show it afterwards. Now every market answer is
    counted here: the reported completion tokens (less reported reasoning tokens when the reasoning
    is not in the answer) may exceed our count of the answer by USAGE_RATIO plus USAGE_SLACK (a tool
    call by TOOL_RATIO plus TOOL_SLACK: measured 2026-11-06, a provider reported 108 tokens for a call
    whose arguments it then dropped, and 439 for one that counted 253), and the reported prompt tokens our count of the prompt by PROMPT_RATIO plus PROMPT_SLACK (images and
    tool schemas count differently per runtime). Beyond that the request is logged with its call id
    as `garageai: usage:` (the spend row stands; an admin corrects it from the line), and a garage
    is routed around for SUSPECT_S; a provider is only logged, since it carries auto. Hidden
    reasoning cannot be counted: a runtime can still pad `reasoning_tokens`, which the line shows.
"""
import asyncio
import hashlib
import random
import re
import statistics
import time
from collections import deque

import litellm
from fastapi import HTTPException
from litellm.integrations.custom_logger import CustomLogger
from litellm.proxy._types import ProxyException

MARGIN = 0.95  # share of the window we promise; covers tokenizer differences
SKIPPED_CALLS = ("embed", "image", "audio", "speech", "transcri", "rerank", "moderation")  # no prompt window to check
SKIP_KEYS = {"type", "role", "id", "tool_use_id", "call_id", "cache_control", "model", "signature"}


def _add_missing_items(node):
    if isinstance(node, dict):
        if node.get("type") == "array" and "items" not in node:
            node["items"] = {}
        for value in node.values():
            _add_missing_items(value)
    elif isinstance(node, list):
        for value in node:
            _add_missing_items(value)


def _strings(node, out):
    """Every text string in a request part, in any of the three API formats."""
    if isinstance(node, str):
        out.append(node)
    elif isinstance(node, dict):
        for key, value in node.items():
            if key not in SKIP_KEYS:
                _strings(value, out)
    elif isinstance(node, list):
        for value in node:
            _strings(value, out)


def _count_input(data):
    parts = []
    for key in ("system", "messages", "input", "instructions", "tools"):
        if data.get(key) is not None:
            _strings(data[key], parts)
    return litellm.token_counter(text="\n".join(parts)) if parts else 0


def _requested_output(data):
    for key in ("max_tokens", "max_completion_tokens", "max_output_tokens"):
        value = data.get(key)
        if isinstance(value, int) and value > 0:
            return value
    return 0


def _window(model):
    from litellm.proxy.proxy_server import llm_router  # set once the proxy has started
    if not llm_router or not model:
        return None
    try:
        info = llm_router.get_model_group_info(model_group=model)
    except Exception:  # noqa: BLE001 - an unknown model is the router's error to raise
        return None
    window = getattr(info, "max_input_tokens", None)  # LiteLLM reports it as a float
    return int(window) if isinstance(window, (int, float)) and window > 0 else None


def context_message(anthropic, prompt, output, limit):
    if anthropic:
        if prompt > limit:
            return f"prompt is too long: {prompt} tokens > {limit} maximum"
        return (f"input length and `max_tokens` exceed context limit: {prompt} + {output} > {limit}, "
                "decrease input length or `max_tokens` and try again")
    return (f"This model's maximum context length is {limit} tokens. However, you requested "
            f"{prompt + output} tokens ({prompt} in the messages, {output} in the completion). "
            "Please reduce the length of the messages or completion.")


RETRY_AFTER = 60              # seconds an agent should wait before trying a model that is away
KNOWN_FOR = "7 days"          # a model that answered within this is "away", not unknown
_served = {"at": 0.0, "names": set()}


async def recently_served(name):
    """True when `name` answered requests within KNOWN_FOR (from the spend logs, cached 10 min).
    Such a model is away for now (its garage is offline and the portal took it out of routing),
    not a typo, and deserves a 503 instead of "model not found"."""
    if time.time() - _served["at"] > 600:
        try:
            from litellm.proxy.proxy_server import prisma_client
            rows = await prisma_client.db.query_raw(
                'select distinct model_group from "LiteLLM_SpendLogs" '
                f"where \"startTime\" > now() - interval '{KNOWN_FOR}' and status = 'success' and model_group not like 'probe/%'")
            _served["names"] = {r.get("model_group") for r in rows if r.get("model_group")}
            _served["at"] = time.time()
        except Exception:  # noqa: BLE001 - without the list, answer as LiteLLM would
            return False
    return name in _served["names"]


def unavailable_message(model, auto_name=None):
    alt = f"use {auto_name}, which fails over automatically, or another model from /v1/models" if auto_name \
        else "or use another model from /v1/models"
    return (f"{model} is temporarily unavailable: the garage or provider that serves it is not answering. "
            f"Retry in a minute, {alt}.")


class ModelTemporarilyUnavailable(ProxyException):
    """503 with Retry-After for a model whose garage is offline: SDKs and agents retry a 503 and
    back off on Retry-After, while "model not found" (400) makes them give up."""

    def __init__(self, model, auto_name=None):
        super().__init__(message=unavailable_message(model, auto_name), type="service_unavailable", param="model",
                         code=503, headers={"Retry-After": str(RETRY_AFTER)})


# Providers (connection_type "endpoint": a datacenter or partner API reached over HTTPS) are not
# garages: no Bridge, no heartbeat, no tunnel state, so nothing takes a dead one out of routing
# within minutes. What the gateway does at request time is all there is:
# - a provider gets a shorter time limit than a home garage (600 s, for 100k-token prefills on home
#   GPUs): a provider that hangs otherwise holds every request it gets for ten minutes;
# - a provider that rejects our key answers the buyer as "temporarily unavailable" (503), not as an
#   authentication error the buyer would take for a problem with their own key.
PROVIDER_TIMEOUT = 180


def _deployment_info(kwargs):
    for key in ("model_info",):
        v = kwargs.get(key)
        if isinstance(v, dict) and v:
            return v
    for meta in ("litellm_metadata", "metadata"):
        m = kwargs.get(meta)
        if isinstance(m, dict) and isinstance(m.get("model_info"), dict):
            return m["model_info"]
    return {}


def is_provider(model_info):
    return isinstance(model_info, dict) and model_info.get("connection_type") == "endpoint"


def provider_time_limit(kwargs):
    """The request's time limit capped at PROVIDER_TIMEOUT when it goes to a provider; returns the
    (possibly) modified kwargs."""
    if not is_provider(_deployment_info(kwargs)):
        return kwargs
    for key in ("timeout", "stream_timeout"):
        current = kwargs.get(key)
        if not isinstance(current, (int, float)) or current > PROVIDER_TIMEOUT:
            kwargs[key] = PROVIDER_TIMEOUT
    return kwargs


def provider_unavailable(exception, request_data):
    """A 503 for the buyer when a deployment (not the buyer's key) failed in a way that only the
    garage or provider can fix: it rejected our key, or every deployment of the model is cooling
    down or unhealthy. None for everything else. It is an HTTPException because that is the only
    kind LiteLLM's proxy takes from a failure hook (anything else is logged and dropped: a
    ProxyException here was silently ignored on the first deploy, 2026-10-18)."""
    model = (request_data or {}).get("model") or "this model"
    text = str(exception)
    away = isinstance(exception, (litellm.AuthenticationError, litellm.PermissionDeniedError)) \
        or "No deployments available for selected model" in text \
        or "There are no healthy deployments" in text
    if not away:
        return None
    return HTTPException(status_code=503, headers={"Retry-After": str(RETRY_AFTER)},
                         detail=unavailable_message(model, "garageai/auto" if model != "garageai/auto" else None))


class ContextLengthExceeded(ProxyException):
    """400 with code "context_length_exceeded", as OpenAI returns it."""

    def __init__(self, message):
        super().__init__(message=message, type="invalid_request_error", param="messages", code=400)

    def to_dict(self):
        error = super().to_dict()
        error["code"] = "context_length_exceeded"
        return error


SYSTEM_ROLES = ("system", "developer")


def _merge_contents(contents):
    """Join message contents: plain text stays text; if any part is a list, keep the parts."""
    if all(c is None or isinstance(c, str) for c in contents):
        return "\n\n".join(c for c in contents if c)
    parts = []
    for c in contents:
        if isinstance(c, str) and c:
            parts.append({"type": "text", "text": c})
        elif isinstance(c, list):
            parts.extend(c)
    return parts


def hoist_system_messages(messages):
    """The messages with every system/developer message merged into one, first; None if they
    are already regular (no such message, or exactly one and it is first)."""
    if not isinstance(messages, list):
        return None
    idx = [i for i, m in enumerate(messages) if isinstance(m, dict) and m.get("role") in SYSTEM_ROLES]
    if not idx or idx == [0]:
        return None
    moved = set(idx)
    merged = {"role": "system", "content": _merge_contents([messages[i].get("content") for i in idx])}
    return [merged] + [m for i, m in enumerate(messages) if i not in moved]


def _as_blocks(content):
    if isinstance(content, list):
        return [c for c in content if c]
    return [{"type": "text", "text": content}] if content else []


def fold_into_anthropic_system(data):
    """An Anthropic request (/v1/messages) carries its system prompt in data["system"]; a
    system/developer message inside data["messages"] on top of that reaches the runtime as a second
    system message, which strict runtimes reject ("System message must be at the beginning", seen
    from Claude Code on qwen3.8 via /v1/responses). Fold such messages into data["system"], in
    order, after what is there. Returns how many were moved."""
    messages = data.get("messages")
    if not isinstance(messages, list):
        return 0
    idx = [i for i, m in enumerate(messages) if isinstance(m, dict) and m.get("role") in SYSTEM_ROLES]
    if not idx:
        return 0
    extra = [messages[i].get("content") for i in idx]
    current = data.get("system")
    if (current is None or isinstance(current, str)) and all(c is None or isinstance(c, str) for c in extra):
        data["system"] = "\n\n".join(c for c in [current, *extra] if c)
    else:
        blocks = _as_blocks(current)
        for c in extra:
            blocks.extend(_as_blocks(c))
        data["system"] = blocks
    data["messages"] = [m for i, m in enumerate(messages) if i not in set(idx)]
    return len(idx)


def _is_anthropic(call_type=None, request_data=None):
    if call_type == "anthropic_messages":
        return True
    url = str(((request_data or {}).get("proxy_server_request") or {}).get("url", ""))
    return "/v1/messages" in url


# ---------------------------------------------------------------- 6 + 7: prefill time and stickiness

PREFER_S = 120          # a deployment that needs longer than this for the new tokens is a last resort
LIMIT_S = 540           # never send a prompt that would not start before stream_timeout (600 s) cuts it
UNMEASURED_GARAGE_TPS = 800  # prefill tokens/s assumed for a garage with no measurements yet
MIN_NEW_TOKENS = 4000   # smaller prefills say more about queueing than about speed
SAMPLES = 20            # per deployment, the median of the latest
STICKY_S = 600          # how long a runtime is expected to keep a conversation's prefix
CHARS_PER_TOKEN = 3.3   # a slightly high token estimate: errs towards the faster deployment
_speeds = {}            # model_id -> deque of prefill tokens/s
_sticky = {}            # conversation -> (model_id, prompt tokens, time)
_seeded = {"done": False}
_too_slow = {}          # conversation (or prompt size) -> time no deployment could start in time


def record_prefill(model_id, prompt_tokens, cached_tokens, ttft_s):
    """One streamed request's prefill speed, when it says something."""
    if not model_id or not isinstance(prompt_tokens, (int, float)) or not isinstance(ttft_s, (int, float)):
        return
    new = prompt_tokens - (cached_tokens or 0)
    if new < MIN_NEW_TOKENS or ttft_s <= 0.2:
        return
    _speeds.setdefault(model_id, deque(maxlen=SAMPLES)).append(new / ttft_s)


def prefill_speed(deployment):
    """Measured prefill tokens/s of a deployment; None means not limited (an unmeasured provider)."""
    info = deployment.get("model_info") or {}
    samples = _speeds.get(info.get("id"))
    if samples and len(samples) >= 3:
        return statistics.median(samples)
    return None if is_provider(info) else UNMEASURED_GARAGE_TPS


def _meta(kwargs):
    out = {}
    for key in ("litellm_metadata", "metadata"):
        m = (kwargs or {}).get(key)
        if isinstance(m, dict):
            out.update(m)
    lp = (kwargs or {}).get("litellm_params")
    if isinstance(lp, dict) and isinstance(lp.get("metadata"), dict):
        out.update(lp["metadata"])
    return out


def _first_text(messages, roles):
    for m in messages or []:
        if isinstance(m, dict) and m.get("role") in roles:
            parts = []
            _strings(m.get("content"), parts)
            return "\n".join(parts)[:4000]
    return ""


def conversation_key(model_group, messages, kwargs):
    """The same conversation, turn after turn: the buyer's key, the model asked for, and the first
    system and user message (which do not change as the conversation grows). None without a user
    message."""
    user = _first_text(messages, ("user",))
    if not user:
        return None
    meta = _meta(kwargs)
    who = str(meta.get("user_api_key_hash") or meta.get("user_api_key") or "")
    system = _first_text(messages, SYSTEM_ROLES)
    if not system and isinstance((kwargs or {}).get("system"), (str, list)):
        parts = []
        _strings(kwargs["system"], parts)
        system = "\n".join(parts)[:4000]
    return hashlib.sha256("\x00".join([who, str(model_group), system, user]).encode()).hexdigest()


def estimate_tokens(messages, kwargs):
    parts = []
    _strings(messages, parts)
    for key in ("system", "tools"):
        if (kwargs or {}).get(key) is not None:
            _strings(kwargs[key], parts)
    return int(sum(len(p) for p in parts) / CHARS_PER_TOKEN)


def choose_deployments(deployments, prompt_tokens, conv, now=None):
    """Rules 6 and 7 on the router's healthy deployments. Returns the ones to route among (the
    router then applies `order` and weights) and a word on what was done, for the log."""
    now = time.time() if now is None else now
    if not isinstance(deployments, list) or len(deployments) < 2 or not prompt_tokens:
        return deployments, ""
    sticky = _sticky.get(conv) if conv else None
    if sticky and now - sticky[2] > STICKY_S:
        _sticky.pop(conv, None)
        sticky = None

    def seconds(d):
        speed = prefill_speed(d)
        if speed is None:
            return 0.0
        new = prompt_tokens
        if sticky and (d.get("model_info") or {}).get("id") == sticky[0]:
            new = max(0, prompt_tokens - sticky[1])
        return new / max(speed, 1.0)

    timed = [(seconds(d), d) for d in deployments]
    if sticky:
        for t, d in timed:
            if (d.get("model_info") or {}).get("id") == sticky[0] and t <= PREFER_S:
                return [d], "sticky"
    prefer = [d for t, d in timed if t <= PREFER_S]
    if prefer:
        return prefer, ("" if len(prefer) == len(deployments) else f"skipped {len(deployments) - len(prefer)} slow")
    possible = sorted([(t, d) for t, d in timed if t <= LIMIT_S], key=lambda x: x[0])
    if possible:
        return [possible[0][1]], f"only slow ones: fastest needs {int(possible[0][0])} s"
    return [], "none can start in time"


def remember(conv, model_id, prompt_tokens, now=None):
    if not conv or not model_id or not isinstance(prompt_tokens, (int, float)):
        return
    now = time.time() if now is None else now
    _sticky[conv] = (model_id, int(prompt_tokens), now)
    if len(_sticky) > 20000:
        cut = now - STICKY_S
        for k in [k for k, v in _sticky.items() if v[2] < cut]:
            _sticky.pop(k, None)


def too_slow_message(request_data):
    """The 503's words when rule 6 left no deployment for this request (within the last 30 s)."""
    data = request_data or {}
    msgs = data.get("messages")
    prompt = estimate_tokens(msgs, data)
    keys = [conversation_key(data.get("model"), msgs, data), f"size:{prompt}"]
    if not any(k and time.time() - _too_slow.get(k, 0) < 30 for k in keys):
        return None
    return (f"This prompt (about {prompt} tokens) would take more than {LIMIT_S // 60} minutes to start on the capacity "
            "that is available for this model right now. Retry in a minute, or shorten the conversation.")


async def seed_speeds():
    """Measurements from the last three days' spend logs, once after a start, so a restart does not
    forget what each deployment can do."""
    if _seeded["done"]:
        return
    _seeded["done"] = True
    try:
        from litellm.proxy.proxy_server import prisma_client
        rows = await prisma_client.db.query_raw(
            'select model_id, prompt_tokens, '
            "coalesce((metadata->'usage_object'->'prompt_tokens_details'->>'cached_tokens')::int, 0) as cached, "
            'extract(epoch from "completionStartTime" - "startTime") as ttft '
            'from "LiteLLM_SpendLogs" where "startTime" > now() - interval \'3 days\' and status = \'success\' '
            'and "completionStartTime" < "endTime" - interval \'1 second\' and prompt_tokens >= ' + str(MIN_NEW_TOKENS) +
            ' order by "startTime"')
        for r in rows:
            record_prefill(r.get("model_id"), r.get("prompt_tokens"), r.get("cached"), float(r.get("ttft") or 0))
        print(f"garageai: prefill speeds seeded for {len(_speeds)} deployment(s)", flush=True)
    except Exception as e:  # noqa: BLE001 - without history every deployment starts unmeasured
        print(f"garageai: could not seed prefill speeds: {e}", flush=True)


# ---------------------------------------------------------------- 8: my machine (private garages)

MINE_PREFIX = "mine/"
_owners_cache = {"at": 0.0, "router": None, "map": {}}


def is_private(deployment):
    info = (deployment or {}).get("model_info") or {}
    return info.get("garage_tier") == "private"


def private_owners(model_list):
    """{mine/... model name: {operator ids}} from the router's deployments."""
    out = {}
    for d in model_list or []:
        name = d.get("model_name") or ""
        info = d.get("model_info") or {}
        if name.startswith(MINE_PREFIX) and info.get("operator_id"):
            out.setdefault(name, set()).add(str(info["operator_id"]))
    return out


def _owners():
    from litellm.proxy.proxy_server import llm_router
    if llm_router is None:
        return {}
    c = _owners_cache
    # Rebuilt every 10 s: the portal adds and removes deployments at runtime.
    if c["router"] is not llm_router or time.time() - c["at"] > 10:
        c["map"], c["router"], c["at"] = private_owners(getattr(llm_router, "model_list", None)), llm_router, time.time()
    return c["map"]


def is_proxy_admin(user_api_key_dict):
    role = getattr(user_api_key_dict, "user_role", None)
    return getattr(role, "value", role) == "proxy_admin"


def visible_names(names, user_id, owners, admin=False):
    """The listing without other people's private machines. The proxy admin (the master key: the
    portal's sync, Ops) sees every name: the portal reconciles its deployments from the listing,
    and hiding them made it create the same private routes again and again (2026-10-20 21:00).
    Seeing is not using: the owner check in the pre-call hook applies to the admin too."""
    if admin:
        return list(names)
    uid = str(user_id) if user_id else None
    return [n for n in names if not (isinstance(n, str) and n.startswith(MINE_PREFIX)) or (uid and uid in owners.get(n, ()))]


def not_mine_message(model):
    return (f"The model `{model}` does not exist or you do not have access to it.")


class ModelNotFound(ProxyException):
    """404 model_not_found, as OpenAI answers an unknown model."""

    def __init__(self, model):
        super().__init__(message=not_mine_message(model), type="invalid_request_error", param="model", code=404)

    def to_dict(self):
        error = super().to_dict()
        error["code"] = "model_not_found"
        return error


async def used_by(user_id, name):
    """True when this user had successful requests to `name` within KNOWN_FOR: their own machine is
    away (asleep, runtime down), not a name they never had."""
    if not user_id:
        return False
    try:
        from litellm.proxy.proxy_server import prisma_client
        rows = await prisma_client.db.query_raw(
            'select 1 from "LiteLLM_SpendLogs" where "user" = $1 and model_group = $2 and status = \'success\' '
            f"and \"startTime\" > now() - interval '{KNOWN_FOR}' limit 1", str(user_id), name)
        return bool(rows)
    except Exception:  # noqa: BLE001 - unknown is not found
        return False


# ---------------------------------------------------------------- 9: auto says which model answered

AUTO_NAME = "garageai/auto"
RAW_NAME_KEY = "_complexity_router_return_raw_model_name"  # litellm.constants.RETURN_RAW_MODEL_NAME_METADATA_KEY


def keep_served_name(data, call_type):
    """Tell LiteLLM not to overwrite `model` with garageai/auto for an OpenAI-format auto request
    (it restamps every response to the requested name). Our hooks then set the served model's name.
    Anthropic-format requests keep LiteLLM's restamp: their stream goes a different way."""
    if data.get("model") != AUTO_NAME or _is_anthropic(call_type, data) or str(call_type) not in ("completion", "acompletion"):
        return False
    meta = data.get("metadata")
    if not isinstance(meta, dict):
        meta = data["metadata"] = {}
    meta[RAW_NAME_KEY] = True
    return True
_canon_cache = {"at": 0.0, "router": None, "map": {}}


def auto_canonical_names(model_list):
    """{auto deployment id: canonical model name}. An auto entry carries the garage and the runtime
    model; the same garage serves that runtime model under its catalogue name as a pool or a
    dedicated route (garage/<g>/<canonical>). model_info.canonical_model wins when the portal sets it."""
    by_key = {}
    for d in model_list or []:
        name = d.get("model_name") or ""
        if name == AUTO_NAME or name.startswith((MINE_PREFIX, "probe/")):
            continue
        info, lp = d.get("model_info") or {}, d.get("litellm_params") or {}
        key = (info.get("garage"), lp.get("model"))
        dedicated = name.startswith("garage/") and name.count("/") >= 2
        canon = name.split("/", 2)[2] if dedicated else name
        if key not in by_key or (by_key[key][1] and not dedicated):  # a pool name beats a dedicated one
            by_key[key] = (canon, dedicated)
    out = {}
    for d in model_list or []:
        if d.get("model_name") != AUTO_NAME:
            continue
        info, lp = d.get("model_info") or {}, d.get("litellm_params") or {}
        canon = info.get("canonical_model") or (by_key.get((info.get("garage"), lp.get("model"))) or (None,))[0]
        if canon and info.get("id"):
            out[info["id"]] = canon
    return out


def _auto_names():
    from litellm.proxy.proxy_server import llm_router
    if llm_router is None:
        return {}
    c = _canon_cache
    if c["router"] is not llm_router or time.time() - c["at"] > 10:
        c["map"], c["router"], c["at"] = auto_canonical_names(getattr(llm_router, "model_list", None)), llm_router, time.time()
    return c["map"]


def served_deployment_id(obj, request_data=None):
    hp = getattr(obj, "_hidden_params", None)
    if isinstance(hp, dict) and hp.get("model_id"):
        return hp["model_id"]
    for key in ("metadata", "litellm_metadata"):
        m = (request_data or {}).get(key)
        if isinstance(m, dict) and isinstance(m.get("model_info"), dict) and m["model_info"].get("id"):
            return m["model_info"]["id"]
    return None


def name_the_model(obj, request_data):
    """Puts the served model's canonical name in obj's `model` (a response or a stream chunk) when the
    buyer asked for garageai/auto. Returns obj."""
    if (request_data or {}).get("model") != AUTO_NAME:
        return obj
    # Always a public name: the served model's catalogue name, or garageai/auto when it is not known.
    # LiteLLM's own restamp is switched off for these requests (RAW_NAME_KEY), so this is the last word
    # on the field and a provider's internal model id can never reach the buyer.
    sid = served_deployment_id(obj, request_data)
    canon = _auto_names().get(sid) or _boost_ids.get(sid) or AUTO_NAME
    if isinstance(obj, (bytes, str)):
        old = '"model":"' + AUTO_NAME + '"'
        new = '"model":"' + canon + '"'
        if isinstance(obj, bytes):
            return obj.replace(old.encode(), new.encode()).replace(old.replace(":", ": ").encode(), new.encode())
        return obj.replace(old, new).replace(old.replace(":", ": "), new)
    if isinstance(obj, dict):
        obj["model"] = canon
        return obj
    try:
        obj.model = canon
    except Exception:  # noqa: BLE001 - leave the name rather than fail the response
        pass
    return obj


# ---------------------------------------------------------------- 10: owner first

OWNER_HOLD_MIN = 5          # minutes the market waits after the owner's last request
_owner_until = {}           # garage -> time until which its owner counts as active


def owner_hold_minutes(garage):
    """The garage's hold from its deployments' model_info (portal setting), else OWNER_HOLD_MIN."""
    try:
        from litellm.proxy.proxy_server import llm_router
        for d in getattr(llm_router, "model_list", None) or []:
            info = d.get("model_info") or {}
            if info.get("garage") == garage and isinstance(info.get("owner_hold_minutes"), (int, float)):
                return max(1, min(30, int(info["owner_hold_minutes"])))
    except Exception:  # noqa: BLE001
        pass
    return OWNER_HOLD_MIN


def mark_owner_active(garage, now=None, minutes=None):
    if not garage:
        return
    now = time.time() if now is None else now
    until = now + 60 * (minutes if minutes is not None else owner_hold_minutes(garage))
    was = _owner_until.get(garage, 0) > now
    _owner_until[garage] = max(_owner_until.get(garage, 0), until)
    if not was:
        print(f"garageai: owner first: {garage} held for the market (owner active)", flush=True)


def garage_of_mine(model):
    parts = str(model or "").split("/")
    return parts[1] if len(parts) >= 3 and parts[0] + "/" == MINE_PREFIX else None


def apply_owner_hold(deployments, now=None):
    """Market deployments of owner-active garages left out. Returns (kept, seconds until the first
    held garage is free, or 0 when nothing was held)."""
    now = time.time() if now is None else now
    kept, wait = [], 0
    for d in deployments:
        info = d.get("model_info") or {}
        until = _owner_until.get(info.get("garage"), 0)
        if until > now and info.get("owner_priority") != "share" and not is_private(d):
            wait = int(until - now) if not wait else min(wait, int(until - now))
            continue
        kept.append(d)
    return kept, wait


_held = {}                  # model group -> (time, retry-after seconds) when the hold left nothing


# ---------------------------------------------------------------- 11: a start for garages that wait

BOOST_SHARE = 0.25          # share of small auto requests offered to a garage that waits
BOOST_TARGET = 10           # market requests in BOOST_WINDOW_S below which a garage waits
BOOST_WINDOW_S = 86400
BOOST_MAX_PROMPT = 16000    # tokens: chat-sized requests only
BOOST_MIN_OUTPUT = 1000     # tokens: a lower max_tokens leaves a thinking model no room to answer
BOOST_PAUSE_S = 1800        # no start for a garage whose started request failed, for this long
BOOST_KEY = "garageai_boost"  # in the request's metadata: this request was offered a start once
_served_at = {}             # garage -> deque of times it served a market request
_boost_failed = {}          # garage -> time its started request failed
_boost_ids = {}             # deployment id -> canonical name, for requests rule 11 started there


def note_served(garage, now=None):
    if not garage:
        return
    now = time.time() if now is None else now
    q = _served_at.setdefault(garage, deque())
    q.append(now)
    while q and now - q[0] > BOOST_WINDOW_S:
        q.popleft()
    while len(q) > BOOST_TARGET:  # only "fewer than the target or not" matters
        q.popleft()


def served_recently(garage, now=None):
    now = time.time() if now is None else now
    return sum(1 for t in _served_at.get(garage, ()) if now - t <= BOOST_WINDOW_S)


def boost_candidates(model_list, now=None):
    """Market pool deployments of garages that wait and serve a model on the auto list."""
    now = time.time() if now is None else now
    auto = [d for d in model_list or [] if d.get("model_name") == AUTO_NAME]
    canon = set(auto_canonical_names(model_list).values())
    on_list = {(d.get("model_info") or {}).get("garage") for d in auto}
    out = []
    for d in model_list or []:
        info = d.get("model_info") or {}
        g = info.get("garage")
        if (info.get("garage_tier") != "pool" or is_private(d) or d.get("model_name") not in canon
                or not g or g in on_list or not info.get("id")):
            continue
        if now - _boost_failed.get(g, 0) < BOOST_PAUSE_S or served_recently(g, now) >= BOOST_TARGET:
            continue
        out.append(d)
    return out


def choose_boost(model_list, prompt_tokens, conv, now=None, roll=None, max_output=0):
    """Rule 11: the deployment to put first for this auto request, or None. max_output is the
    request's own answer cap, 0 when it sets none."""
    now = time.time() if now is None else now
    if not prompt_tokens or prompt_tokens > BOOST_MAX_PROMPT or 0 < max_output < BOOST_MIN_OUTPUT:
        return None
    cands, _ = apply_owner_hold(boost_candidates(model_list, now), now)
    cands = [d for d in cands if prompt_tokens / max(prefill_speed(d) or UNMEASURED_GARAGE_TPS, 1.0) <= PREFER_S]
    if not cands:
        return None
    sticky = _sticky.get(conv) if conv else None
    if sticky and now - sticky[2] <= STICKY_S:
        for d in cands:
            if (d.get("model_info") or {}).get("id") == sticky[0]:
                return d
    if (random.random() if roll is None else roll) >= BOOST_SHARE:
        return None
    least = min(served_recently((d.get("model_info") or {}).get("garage"), now) for d in cands)
    return random.choice([d for d in cands if served_recently((d.get("model_info") or {}).get("garage"), now) == least])


def first_in_line(deployment):
    """A copy of the deployment that the router's order filter puts before every auto entry."""
    return {**deployment, "litellm_params": {**(deployment.get("litellm_params") or {}), "order": 0},
            "model_info": {**(deployment.get("model_info") or {}), "order": 0}}


def _request_meta(request_kwargs):
    for key in ("litellm_metadata", "metadata"):
        m = (request_kwargs or {}).get(key)
        if isinstance(m, dict):
            return m
    return None


def _deployment_of_id(model_id):
    try:
        from litellm.proxy.proxy_server import llm_router
        for d in getattr(llm_router, "model_list", None) or []:
            if (d.get("model_info") or {}).get("id") == model_id:
                return d
    except Exception:  # noqa: BLE001
        pass
    return None


def _garage_of_id(model_id):
    return ((_deployment_of_id(model_id) or {}).get("model_info") or {}).get("garage")


# ---------------------------------------------------------------- 12: the runtime's token count is checked

USAGE_RATIO = 1.3           # reported answer tokens may exceed our count by this factor...
USAGE_SLACK = 64            # ...plus this many (chat template, stop token, tokenizer differences)
TOOL_RATIO = 2.0            # a tool call carries structure the answer does not show, and a runtime
TOOL_SLACK = 200            # that drops a malformed call still reports its tokens (measured 2026-11-06)
PROMPT_RATIO = 2.0          # the prompt is looser: images and tool schemas count differently per runtime
PROMPT_SLACK = 500
SUSPECT_S = 3600            # a garage whose count did not add up is routed around for this long
ANSWER_SKIP_KEYS = SKIP_KEYS | {"object", "created", "finish_reason", "usage", "index", "logprobs",
                                "system_fingerprint", "service_tier", "status", "stop_reason", "stop_sequence"}
_suspect = {}               # garage -> time its count last did not add up


def _as_dict(response):
    if hasattr(response, "model_dump"):
        try:
            response = response.model_dump()
        except Exception:  # noqa: BLE001
            response = dict(response) if isinstance(response, dict) else {}
    return response if isinstance(response, dict) else {}


def answer_text(response):
    """(every text the runtime returned as the answer, whether reasoning text was among it)."""
    parts = []
    body = {k: v for k, v in _as_dict(response).items() if k not in ANSWER_SKIP_KEYS}
    _collect_answer(body, parts)
    text = "\n".join(parts)
    return text, _has_key(body, ("reasoning_content", "reasoning", "thinking"))


def has_tool_call(response):
    return _has_key(_as_dict(response), ("tool_calls", "function_call"))


def _collect_answer(node, out):
    if isinstance(node, str):
        out.append(node)
    elif isinstance(node, dict):
        for key, value in node.items():
            if key not in ANSWER_SKIP_KEYS:
                _collect_answer(value, out)
    elif isinstance(node, list):
        for value in node:
            _collect_answer(value, out)


def _has_key(node, keys):
    if isinstance(node, dict):
        return any(k in node and node[k] for k in keys) or any(_has_key(v, keys) for v in node.values())
    if isinstance(node, list):
        return any(_has_key(v, keys) for v in node)
    return False


def usage_check(reported_completion, reported_prompt, counted_answer, counted_prompt, reasoning_tokens=0,
                reasoning_shown=False, tool_call=False):
    """None when the runtime's count adds up, else a short reason. counted_prompt 0 = not counted."""
    visible = reported_completion or 0
    if not reasoning_shown:
        visible -= reasoning_tokens or 0
    ratio, slack = (TOOL_RATIO, TOOL_SLACK) if tool_call else (USAGE_RATIO, USAGE_SLACK)
    if visible > counted_answer * ratio + slack:
        return (f"reported {reported_completion} completion tokens ({reasoning_tokens or 0} reasoning), "
                f"the answer counts {counted_answer}")
    if counted_prompt and reported_prompt and reported_prompt > counted_prompt * PROMPT_RATIO + PROMPT_SLACK:
        return f"reported {reported_prompt} prompt tokens, the prompt counts {counted_prompt}"
    return None


def drop_suspects(deployments, now=None):
    now = time.time() if now is None else now
    if not _suspect:
        return deployments
    return [d for d in deployments if now - _suspect.get((d.get("model_info") or {}).get("garage"), 0) >= SUSPECT_S]


async def check_usage(kwargs, response_obj):
    """Rule 12, after a market request succeeded: compare the runtime's usage with our own count."""
    slo = kwargs.get("standard_logging_object") or {}
    group = str(slo.get("model_group") or "")
    if not group or group.startswith(MINE_PREFIX) or group.startswith("probe/"):
        return None
    usage = ((slo.get("metadata") or {}).get("usage_object") or {})
    reasoning = ((usage.get("completion_tokens_details") or {}).get("reasoning_tokens")) or 0
    text, shown = answer_text(response_obj)
    if not text and not slo.get("completion_tokens"):
        return None
    counted_answer = await asyncio.to_thread(litellm.token_counter, text=text) if text else 0
    counted_prompt = 0
    if kwargs.get("messages"):
        try:
            counted_prompt = await asyncio.to_thread(estimate_tokens, kwargs.get("messages"), kwargs)
        except Exception:  # noqa: BLE001 - the prompt check is optional
            counted_prompt = 0
    why = usage_check(slo.get("completion_tokens"), slo.get("prompt_tokens"), counted_answer, counted_prompt,
                      reasoning, shown, tool_call=has_tool_call(response_obj))
    if not why:
        return None
    d = _deployment_of_id(slo.get("model_id")) or {}
    info = d.get("model_info") or {}
    garage = info.get("garage") or slo.get("model_id")
    if info.get("connection_type") == "endpoint":
        print(f"garageai: usage: provider {garage} model={group} id={slo.get('id')}: {why}", flush=True)
    else:
        _suspect[garage] = time.time()
        print(f"garageai: usage: garage {garage} model={group} id={slo.get('id')}: {why} "
              f"-> routed around for {SUSPECT_S // 60} min; check the spend row", flush=True)
    return why


class GarageAICallbacks(CustomLogger):
    async def async_pre_call_hook(self, user_api_key_dict, cache, data, call_type):
        model = data.get("model")
        if isinstance(model, str) and model.startswith(MINE_PREFIX):
            uid = getattr(user_api_key_dict, "user_id", None)
            owners = _owners().get(model)
            if owners is None:
                if await used_by(uid, model):
                    print(f"garageai: {model}: the owner's machine is away: 503", flush=True)
                    raise HTTPException(status_code=503, headers={"Retry-After": str(RETRY_AFTER)},
                                        detail=f"Your machine ({model.split('/')[1]}) is not answering right now: it may be "
                                               "asleep, offline or its runtime stopped. Retry when it is back, or use garageai/auto.")
                raise ModelNotFound(model)
            if not uid or str(uid) not in owners:
                print(f"garageai: {model}: not the owner's key: 404", flush=True)
                raise ModelNotFound(model)
            mark_owner_active(garage_of_mine(model))
        elif isinstance(model, str) and model and not model.startswith("probe/"):
            try:
                from litellm.proxy.proxy_server import llm_router
                names = set(llm_router.get_model_names()) if llm_router else None
            except Exception:  # noqa: BLE001 - never fail a request because we could not look
                names = None
            if names is not None and model not in names and await recently_served(model):
                auto = "garageai/auto" if "garageai/auto" in names else None
                print(f"garageai: {model} is away (not in routing, served within {KNOWN_FOR}): 503", flush=True)
                raise ModelTemporarilyUnavailable(model, auto)
        try:
            keep_served_name(data, call_type)
        except Exception:  # noqa: BLE001 - the name is a courtesy
            pass
        if data.get("tools"):
            _add_missing_items(data["tools"])
        if _is_anthropic(call_type, data):
            roles = ",".join(str(m.get("role", "?"))[:1] for m in data.get("messages") or [] if isinstance(m, dict))
            moved = fold_into_anthropic_system(data)
            if moved:
                print(f"garageai: folded {moved} system message(s) into the Anthropic system field "
                      f"model={data.get('model')} roles={roles[:80]}", flush=True)
        hoisted = hoist_system_messages(data.get("messages"))
        if hoisted is not None:
            moved = len(data["messages"]) - len(hoisted) + 1
            print(f"garageai: merged {moved} system message(s) into one at the start model={data.get('model')}", flush=True)
            data["messages"] = hoisted
        if any(part in str(call_type) for part in SKIPPED_CALLS):
            return data
        window = _window(data.get("model"))
        if not window:
            return data
        try:
            prompt = await asyncio.to_thread(_count_input, data)
        except Exception:  # noqa: BLE001 - never fail a request because we could not count it
            return data
        output = _requested_output(data)
        limit = int(window * MARGIN)
        if prompt > limit or prompt + output > limit:
            print(f"garageai: context check rejected model={data.get('model')} prompt={prompt} "
                  f"output={output} limit={limit}", flush=True)
            raise ContextLengthExceeded(context_message(_is_anthropic(call_type), prompt, output, limit))
        return data

    async def async_post_call_success_hook(self, data, user_api_key_dict, response):
        # The provider's raw response headers would reach the buyer as llm_provider-* headers
        # (OpenRouter's x-generation-id names the provider, which buyers must never see). Drop them
        # all here; Traefik strips the proxy's own x-litellm-* headers by name.
        hp = getattr(response, "_hidden_params", None)
        if isinstance(hp, dict) and hp.get("additional_headers"):
            hp["additional_headers"] = {}
        try:
            if ((data or {}).get("metadata") or {}).get(RAW_NAME_KEY):
                name_the_model(response, data)
        except Exception:  # noqa: BLE001 - the name is a courtesy, never a failure
            pass
        rid = response.get("id") if isinstance(response, dict) else getattr(response, "id", None)
        if isinstance(rid, str) and rid.startswith("resp_") and _is_anthropic(request_data=data):
            neutral = "msg_" + hashlib.sha256(rid.encode()).hexdigest()[:24]
            if isinstance(response, dict):
                response["id"] = neutral
            else:
                try:
                    response.id = neutral
                except Exception:  # noqa: BLE001 - leave the id rather than fail the response
                    pass
        return response

    async def async_filter_deployments(self, model, healthy_deployments, messages, request_kwargs=None,
                                       parent_otel_span=None):
        try:
            if isinstance(healthy_deployments, list) and not str(model).startswith(MINE_PREFIX):
                # Rule 8: someone's private machine is never routed for a market name.
                kept = [d for d in healthy_deployments if not is_private(d)]
                if len(kept) != len(healthy_deployments):
                    print(f"garageai: routing model={model}: left out {len(healthy_deployments) - len(kept)} private deployment(s)", flush=True)
                healthy_deployments = kept
                # Rule 12: a garage whose token count did not add up waits out SUSPECT_S.
                kept = drop_suspects(healthy_deployments)
                if len(kept) != len(healthy_deployments):
                    print(f"garageai: usage: model={model}: left out {len(healthy_deployments) - len(kept)} suspect deployment(s)", flush=True)
                healthy_deployments = kept
            if isinstance(healthy_deployments, list) and not str(model).startswith(MINE_PREFIX) and _owner_until:
                # Rule 10: the market waits while an owner uses their own machine.
                kept, wait = apply_owner_hold(healthy_deployments)
                if len(kept) != len(healthy_deployments):
                    print(f"garageai: owner first: model={model}: left out {len(healthy_deployments) - len(kept)} held deployment(s)", flush=True)
                    if not kept:
                        _held[model] = (time.time(), max(wait, 1))
                healthy_deployments = kept
            boost = None
            if model == AUTO_NAME and isinstance(healthy_deployments, list) and healthy_deployments:
                # Rule 11: once per request, so a retry goes down the auto list.
                meta = _request_meta(request_kwargs)
                if meta is not None and not meta.get(BOOST_KEY):
                    await seed_speeds()
                    from litellm.proxy.proxy_server import llm_router
                    msgs = messages if messages is not None else (request_kwargs or {}).get("messages")
                    d = choose_boost(getattr(llm_router, "model_list", None),
                                     estimate_tokens(msgs, request_kwargs), conversation_key(model, msgs, request_kwargs),
                                     max_output=_requested_output(request_kwargs or {}))
                    if d is not None:
                        meta[BOOST_KEY] = (d.get("model_info") or {}).get("garage")
                        _boost_ids[(d.get("model_info") or {}).get("id")] = d.get("model_name")
                        boost = first_in_line(d)
                        print(f"garageai: start: auto request offered to {meta[BOOST_KEY]} ({d.get('model_name')})", flush=True)
            if not isinstance(healthy_deployments, list) or len(healthy_deployments) < 2:
                return ([boost] if boost else []) + (healthy_deployments or [])
            await seed_speeds()
            msgs = messages if messages is not None else (request_kwargs or {}).get("messages")
            prompt = estimate_tokens(msgs, request_kwargs)
            conv = conversation_key(model, msgs, request_kwargs)
            chosen, why = choose_deployments(healthy_deployments, prompt, conv)
            if boost:
                return [boost] + (chosen or healthy_deployments)
            if why and why != "sticky":
                print(f"garageai: routing model={model} prompt~{prompt}: {why}", flush=True)
            if not chosen:
                _too_slow[conv or f"size:{prompt}"] = time.time()
            return chosen
        except Exception as e:  # noqa: BLE001 - never fail a request because of these rules
            print(f"garageai: routing rules skipped: {e}", flush=True)
            return healthy_deployments

    async def async_post_call_streaming_iterator_hook(self, user_api_key_dict, response, request_data):
        named = (request_data or {}).get("model") == AUTO_NAME and bool(((request_data or {}).get("metadata") or {}).get(RAW_NAME_KEY))
        async for chunk in response:
            if named:
                try:
                    chunk = name_the_model(chunk, request_data)
                except Exception:  # noqa: BLE001
                    pass
            yield chunk

    async def async_filter_listed_models(self, user_api_key_dict, model_names):
        try:
            return visible_names(list(model_names), getattr(user_api_key_dict, "user_id", None), _owners(),
                                 admin=is_proxy_admin(user_api_key_dict))
        except Exception:  # noqa: BLE001 - fail closed: hide every private name
            return [n for n in model_names if not (isinstance(n, str) and n.startswith(MINE_PREFIX))]

    async def async_log_success_event(self, kwargs, response_obj, start_time, end_time):
        try:
            slo = kwargs.get("standard_logging_object") or {}
            model_id, prompt = slo.get("model_id"), slo.get("prompt_tokens")
            usage = ((slo.get("metadata") or {}).get("usage_object") or {})
            cached = ((usage.get("prompt_tokens_details") or {}).get("cached_tokens")) or 0
            if slo.get("stream") and slo.get("completionStartTime") and slo.get("startTime"):
                record_prefill(model_id, prompt, cached, slo["completionStartTime"] - slo["startTime"])
            group = slo.get("model_group")
            conv = conversation_key(group, kwargs.get("messages"), kwargs)
            remember(conv, model_id, prompt)
            if str(group).startswith(MINE_PREFIX):
                mark_owner_active(garage_of_mine(group))
            elif group and not str(group).startswith("probe/"):
                note_served(_garage_of_id(model_id))
        except Exception:  # noqa: BLE001 - bookkeeping only
            pass
        try:
            await check_usage(kwargs, response_obj)
        except Exception as e:  # noqa: BLE001 - the check must never break logging
            print(f"garageai: usage check skipped: {e}", flush=True)

    async def async_log_failure_event(self, kwargs, response_obj, start_time, end_time):
        try:
            slo = kwargs.get("standard_logging_object") or {}
            if str(slo.get("model_group")).startswith(MINE_PREFIX):
                mark_owner_active(garage_of_mine(slo.get("model_group")))
            conv = conversation_key(slo.get("model_group"), kwargs.get("messages"), kwargs)
            if slo.get("model_id") in _boost_ids and slo.get("model_group") == AUTO_NAME:
                g = _garage_of_id(slo.get("model_id"))
                if g:
                    _boost_failed[g] = time.time()
                    print(f"garageai: start: {g} failed a started request; no start for {BOOST_PAUSE_S // 60} min", flush=True)
            if conv and _sticky.get(conv, (None,))[0] == slo.get("model_id"):
                _sticky.pop(conv, None)  # the retry and the next turn go elsewhere
        except Exception:  # noqa: BLE001 - bookkeeping only
            pass

    async def async_pre_call_deployment_hook(self, kwargs, call_type):
        try:
            return provider_time_limit(kwargs)
        except Exception:  # noqa: BLE001 - never fail a request because of the time limit
            return None

    async def async_post_call_failure_hook(self, request_data, original_exception, user_api_key_dict,
                                           traceback_str=None):
        if isinstance(original_exception, (ContextLengthExceeded, ModelTemporarilyUnavailable, ModelNotFound, HTTPException)):
            return None
        away = provider_unavailable(original_exception, request_data)
        held = _held.get((request_data or {}).get("model"))
        if away is not None and held and time.time() - held[0] < 30:
            print(f"garageai: {request_data.get('model')}: its garage is held for its owner -> 503", flush=True)
            raise HTTPException(status_code=503, headers={"Retry-After": str(held[1])},
                                detail=f"{request_data.get('model')} is busy for a moment: its owner is using the machine. "
                                       f"Retry in about {max(1, held[1] // 60)} minute(s), or use garageai/auto.")
        if away is not None and "No deployments available" in str(original_exception):
            slow = too_slow_message(request_data)
            if slow:
                print(f"garageai: {request_data.get('model')}: no deployment can start this prompt in time -> 503", flush=True)
                raise HTTPException(status_code=503, headers={"Retry-After": str(RETRY_AFTER)}, detail=slow)
        if away is not None:
            print(f"garageai: {request_data.get('model')}: {type(original_exception).__name__} from the deployment -> 503", flush=True)
            raise away
        text = str(original_exception)
        if not (isinstance(original_exception, litellm.ContextWindowExceededError)
                or "maximum context length" in text or "Max Input Tokens" in text):
            return None
        limit = re.search(r"maximum context length is (\d+)|Max Input Tokens=(\d+)", text)
        total = re.search(r"total of at least (\d+)|Got=(\d+)", text)
        output = _requested_output(request_data)
        if limit and total:
            limit_n = int(limit.group(1) or limit.group(2))
            total_n = int(total.group(1) or total.group(2))
            prompt_n = total_n - output if total.group(1) else total_n
            message = context_message(_is_anthropic(request_data=request_data), prompt_n, output, limit_n)
        else:
            message = "The prompt and max_tokens together exceed this model's context window."
        raise HTTPException(status_code=400, detail=message)


proxy_handler_instance = GarageAICallbacks()

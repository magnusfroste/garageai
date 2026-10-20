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
"""
import asyncio
import hashlib
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
    canon = _auto_names().get(served_deployment_id(obj, request_data))
    if not canon:
        return obj
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
            if not isinstance(healthy_deployments, list) or len(healthy_deployments) < 2:
                return healthy_deployments
            await seed_speeds()
            msgs = messages if messages is not None else (request_kwargs or {}).get("messages")
            prompt = estimate_tokens(msgs, request_kwargs)
            conv = conversation_key(model, msgs, request_kwargs)
            chosen, why = choose_deployments(healthy_deployments, prompt, conv)
            if why and why != "sticky":
                print(f"garageai: routing model={model} prompt~{prompt}: {why}", flush=True)
            if not chosen:
                _too_slow[conv or f"size:{prompt}"] = time.time()
            return chosen
        except Exception as e:  # noqa: BLE001 - never fail a request because of these rules
            print(f"garageai: routing rules skipped: {e}", flush=True)
            return healthy_deployments

    async def async_post_call_streaming_iterator_hook(self, user_api_key_dict, response, request_data):
        named = (request_data or {}).get("model") == AUTO_NAME
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
        except Exception:  # noqa: BLE001 - bookkeeping only
            pass

    async def async_log_failure_event(self, kwargs, response_obj, start_time, end_time):
        try:
            slo = kwargs.get("standard_logging_object") or {}
            conv = conversation_key(slo.get("model_group"), kwargs.get("messages"), kwargs)
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

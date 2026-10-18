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
"""
import asyncio
import hashlib
import re
import time

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
    down. None for everything else."""
    model = (request_data or {}).get("model") or "this model"
    text = str(exception)
    if isinstance(exception, (litellm.AuthenticationError, litellm.PermissionDeniedError)):
        return ModelTemporarilyUnavailable(model, "garageai/auto" if model != "garageai/auto" else None)
    if "No deployments available for selected model" in text:
        return ModelTemporarilyUnavailable(model, "garageai/auto" if model != "garageai/auto" else None)
    return None


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


class GarageAICallbacks(CustomLogger):
    async def async_pre_call_hook(self, user_api_key_dict, cache, data, call_type):
        model = data.get("model")
        if isinstance(model, str) and model and not model.startswith("probe/"):
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

    async def async_pre_call_deployment_hook(self, kwargs, call_type):
        try:
            return provider_time_limit(kwargs)
        except Exception:  # noqa: BLE001 - never fail a request because of the time limit
            return None

    async def async_post_call_failure_hook(self, request_data, original_exception, user_api_key_dict,
                                           traceback_str=None):
        if isinstance(original_exception, (ContextLengthExceeded, ModelTemporarilyUnavailable)):
            return None
        away = provider_unavailable(original_exception, request_data)
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

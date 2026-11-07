"""Unit tests for garageai_callbacks.py. They need LiteLLM, so run them in the LiteLLM container:
  sudo docker exec -i garageai-litellm-litellm-1 python - < infra/gateway/litellm/tests/test_callbacks.py
with the new garageai_callbacks.py copied in first (see infra/gateway/README.md), or after a deploy."""
import importlib.util
import sys
import time

spec = importlib.util.spec_from_file_location("cb", sys.argv[1] if len(sys.argv) > 1 else "/app/garageai_callbacks.py")
cb = importlib.util.module_from_spec(spec)
spec.loader.exec_module(cb)
h = cb.hoist_system_messages
fails = 0


def check(name, got, want):
    global fails
    ok = got == want
    fails += not ok
    print(("  ok    " if ok else "  FAIL  ") + name + ("" if ok else f"\n        got:  {got}\n        want: {want}"))


S, U, A = (lambda t: {"role": "system", "content": t}), (lambda t: {"role": "user", "content": t}), (lambda t: {"role": "assistant", "content": t})
check("regular: untouched", h([S("a"), U("b")]), None)
check("no system: untouched", h([U("b"), A("c")]), None)
check("not a list: untouched", h(None), None)
check("system later: moved first", h([U("q"), A("r"), S("remember")]), [S("remember"), U("q"), A("r")])
check("two systems: merged in order", h([S("a"), U("q"), S("b")]), [S("a\n\nb"), U("q")])
check("two leading systems: merged", h([S("a"), S("b"), U("q")]), [S("a\n\nb"), U("q")])
check("developer role counts as system", h([{"role": "developer", "content": "d"}, U("q"), S("s")]), [S("d\n\ns"), U("q")])
check("content parts kept", h([U("q"), {"role": "system", "content": [{"type": "text", "text": "x"}]}, S("y")]),
      [{"role": "system", "content": [{"type": "text", "text": "x"}, {"type": "text", "text": "y"}]}, U("q")])

m = cb.unavailable_message("qwen/qwen3.8-27b")
check("away: says temporarily unavailable", "temporarily unavailable" in m and "/v1/models" in m, True)
check("away: suggests auto when it exists", "garageai/auto" in cb.unavailable_message("x", "garageai/auto"), True)
e = cb.ModelTemporarilyUnavailable("qwen/qwen3.8-27b")
check("away: 503 with Retry-After", (str(e.code), e.headers.get("Retry-After")), ("503", "60"))
f = cb.fold_into_anthropic_system
d = {"system": "base", "messages": [U("q"), S("late")]}
check("anthropic: system message folded into system", (f(d), d), (1, {"system": "base\n\nlate", "messages": [U("q")]}))
d = {"system": [{"type": "text", "text": "a", "cache_control": {"type": "ephemeral"}}], "messages": [U("q"), S("b")]}
check("anthropic: blocks kept, cache_control too", (f(d), d["system"]),
      (1, [{"type": "text", "text": "a", "cache_control": {"type": "ephemeral"}}, {"type": "text", "text": "b"}]))
d = {"messages": [S("only"), U("q")]}
check("anthropic: no system field yet", (f(d), d), (1, {"system": "only", "messages": [U("q")]}))
d = {"system": "s", "messages": [U("q"), A("r")]}
check("anthropic: regular request untouched", (f(d), d), (0, {"system": "s", "messages": [U("q"), A("r")]}))
check("tool messages keep their order", h([S("a"), U("q"), {"role": "assistant", "tool_calls": [1]}, {"role": "tool", "content": "t"}, S("b")]),
      [S("a\n\nb"), U("q"), {"role": "assistant", "tool_calls": [1]}, {"role": "tool", "content": "t"}])
# Providers vs garages (2026-10-18): time limit and the 503 for a deployment's own failure.
prov = {"model_info": {"connection_type": "endpoint", "garage": "autoversio"}, "timeout": 600}
check("provider: time limit capped", cb.provider_time_limit(dict(prov))["timeout"], cb.PROVIDER_TIMEOUT)
check("provider: limit read from metadata too", cb.provider_time_limit({"metadata": {"model_info": {"connection_type": "endpoint"}}, "timeout": 600})["timeout"], cb.PROVIDER_TIMEOUT)
check("provider: a shorter limit stays", cb.provider_time_limit({"model_info": {"connection_type": "endpoint"}, "timeout": 30})["timeout"], 30)
check("garage: 600 s kept", cb.provider_time_limit({"model_info": {"connection_type": "mesh"}, "timeout": 600})["timeout"], 600)
import litellm
auth = litellm.AuthenticationError(message="401 Unauthorized", llm_provider="openai", model="x")
away = cb.provider_unavailable(auth, {"model": "deepseek/deepseek-v4.1-flash"})
check("provider 401: buyer gets 503 as an HTTPException (the only kind the proxy uses)", (type(away).__name__, away.status_code, away.headers.get("Retry-After")), ("HTTPException", 503, "60"))
check("provider 401: auto suggested", "garageai/auto" in away.detail, True)
cool = cb.provider_unavailable(Exception("No deployments available for selected model, Try again in 60 seconds"), {"model": "garageai/auto"})
check("all cooling down: 503", cool.status_code, 503)
check("auto does not suggest itself", "use garageai/auto" in cool.detail, False)
check("no healthy deployments: 503", cb.provider_unavailable(Exception("BadRequestError: There are no healthy deployments for this model"), {"model": "x"}).status_code, 503)
check("other errors: untouched", cb.provider_unavailable(ValueError("boom"), {"model": "x"}), None)


# ---- rules 6 and 7: prefill time and stickiness
cb._speeds.clear(); cb._sticky.clear()
prov = {"model_info": {"id": "prov", "connection_type": "endpoint"}}
dgx = {"model_info": {"id": "dgx", "connection_type": "mesh"}}
fast = {"model_info": {"id": "fast", "connection_type": "mesh"}}
for _ in range(5):
    cb.record_prefill("dgx", 150000, 0, 150.0)      # 1,000 tok/s
    cb.record_prefill("fast", 100000, 0, 10.0)      # 10,000 tok/s
cb.record_prefill("dgx", 3000, 0, 1.0)              # too small to count
check("speed: median of samples", round(cb.prefill_speed(dgx)), 1000)
check("speed: unmeasured provider is not limited", cb.prefill_speed(prov), None)
check("speed: unmeasured garage gets the default", cb.prefill_speed({"model_info": {"id": "new"}}), cb.UNMEASURED_GARAGE_TPS)
ids = lambda ds: [d["model_info"]["id"] for d in ds]
ch = lambda ds, p, c=None, now=None: (ids(cb.choose_deployments(ds, p, c, now)[0]), cb.choose_deployments(ds, p, c, now)[1])
check("small prompt: all kept", ch([prov, dgx], 20000)[0], ["prov", "dgx"])
check("160k cold: slow garage skipped while the provider is up", ch([prov, dgx], 160000)[0], ["prov"])
check("160k cold, only slow garages: the fastest is tried (160 s < limit)", ch([dgx, {"model_info": {"id": "x2", "connection_type": "mesh"}}], 160000)[0], ["dgx"])
check("600k cold on dgx with a slower peer: none can start in time", ch([dgx, {"model_info": {"id": "slow"}}], 600000)[0], [])
check("400k cold, dgx and a slower peer: dgx tried (400 s)", ch([dgx, {"model_info": {"id": "slow"}}], 400000)[0], ["dgx"])
check("faster garage takes the big prompt", ch([dgx, fast], 400000)[0], ["fast"])
check("single deployment: untouched", ch([dgx], 900000)[0], ["dgx"])
conv = cb.conversation_key("garageai/auto", [{"role": "system", "content": "sys"}, {"role": "user", "content": "hi"}], {"metadata": {"user_api_key_hash": "k"}})
conv2 = cb.conversation_key("garageai/auto", [{"role": "system", "content": "sys"}, {"role": "user", "content": "hi"}, {"role": "assistant", "content": "x"}, {"role": "user", "content": "more"}], {"metadata": {"user_api_key_hash": "k"}})
other = cb.conversation_key("garageai/auto", [{"role": "user", "content": "hi"}], {"metadata": {"user_api_key_hash": "someone else"}})
check("conversation: same key as it grows", conv, conv2)
check("conversation: another buyer is another conversation", conv == other, False)
check("conversation: none without a user message", cb.conversation_key("m", [{"role": "system", "content": "s"}], {}), None)
cb.remember(conv, "dgx", 158000, now=1000.0)
check("sticky: warm garage keeps the next 165k turn (7k new tokens)", ch([prov, dgx], 165000, conv, 1100.0), (["dgx"], "sticky"))
check("sticky: expires", ch([prov, dgx], 165000, conv, 1000.0 + cb.STICKY_S + 1)[0], ["prov"])
cb.remember(conv, "dgx", 10000, now=2000.0)
check("sticky: too many new tokens, normal rules", ch([prov, dgx], 300000, conv, 2010.0)[0], ["prov"])
cb.remember(conv, "gone", 158000, now=3000.0)
check("sticky: deployment not healthy, normal rules", ch([prov, dgx], 20000, conv, 3010.0)[0], ["prov", "dgx"])
check("estimate: chars / 3.3", cb.estimate_tokens([{"role": "user", "content": "x" * 3300}], {}), 1000)


# ---- rule 8: my machine
ml = [{"model_name": "mine/garage-m4/ibm-granite/granite-4.1-3b", "model_info": {"operator_id": "u-magnus", "garage_tier": "private"}},
      {"model_name": "garage/garage-m4/ibm-granite/granite-4.1-3b", "model_info": {"operator_id": "u-magnus", "garage_tier": "dedicated"}},
      {"model_name": "mine/dgx-city/x/y", "model_info": {"operator_id": "u-other", "garage_tier": "private"}}]
ow = cb.private_owners(ml)
check("owners: only mine/ names", sorted(ow), ["mine/dgx-city/x/y", "mine/garage-m4/ibm-granite/granite-4.1-3b"])
names = [d["model_name"] for d in ml] + ["garageai/auto"]
check("listing: owner sees own private, not others'", cb.visible_names(names, "u-magnus", ow),
      ["mine/garage-m4/ibm-granite/granite-4.1-3b", "garage/garage-m4/ibm-granite/granite-4.1-3b", "garageai/auto"])
check("listing: no user sees no private", cb.visible_names(names, None, ow), ["garage/garage-m4/ibm-granite/granite-4.1-3b", "garageai/auto"])
check("private deployment detected", (cb.is_private(ml[0]), cb.is_private(ml[1])), (True, False))
e = cb.ModelNotFound("mine/dgx-city/x/y")
check("not yours: 404 model_not_found", (str(e.code), e.to_dict().get("code")), ("404", "model_not_found"))


class _Role:
    value = "proxy_admin"
class _Key:
    def __init__(self, role): self.user_role = role
check("admin: proxy_admin recognised (enum)", cb.is_proxy_admin(_Key(_Role())), True)
check("admin: proxy_admin recognised (string)", cb.is_proxy_admin(_Key("proxy_admin")), True)
check("admin: internal user is not admin", cb.is_proxy_admin(_Key("internal_user")), False)
check("admin: listing shows every private name", cb.visible_names(names, "default_user_id", ow, admin=True), names)


# ---- rule 9: auto names the model
mlist = [
  {"model_name": "garageai/auto", "model_info": {"id": "auto__autoversio__kluster__pool", "garage": "autoversio"}, "litellm_params": {"model": "openai/kluster"}},
  {"model_name": "garageai/auto", "model_info": {"id": "auto__dgx-city__mimo__dedicated", "garage": "dgx-city"}, "litellm_params": {"model": "hosted_vllm/mimo-v2.6-flash"}},
  {"model_name": "garageai/auto", "model_info": {"id": "auto__x__y", "garage": "x", "canonical_model": "acme/set-by-portal"}, "litellm_params": {"model": "openai/y"}},
  {"model_name": "garage/autoversio/deepseek/deepseek-v4.1-flash", "model_info": {"id": "a1", "garage": "autoversio"}, "litellm_params": {"model": "openai/kluster"}},
  {"model_name": "deepseek/deepseek-v4.1-flash", "model_info": {"id": "a2", "garage": "autoversio"}, "litellm_params": {"model": "openai/kluster"}},
  {"model_name": "garage/dgx-city/xiaomi/mimo-v2.6-flash", "model_info": {"id": "d1", "garage": "dgx-city"}, "litellm_params": {"model": "hosted_vllm/mimo-v2.6-flash"}},
  {"model_name": "mine/dgx-city/xiaomi/mimo-v2.6-flash", "model_info": {"id": "m1", "garage": "dgx-city"}, "litellm_params": {"model": "hosted_vllm/mimo-v2.6-flash"}},
]
cn = cb.auto_canonical_names(mlist)
check("auto: pool name preferred", cn.get("auto__autoversio__kluster__pool"), "deepseek/deepseek-v4.1-flash")
check("auto: dedicated name stripped of garage", cn.get("auto__dgx-city__mimo__dedicated"), "xiaomi/mimo-v2.6-flash")
check("auto: portal's canonical_model wins", cn.get("auto__x__y"), "acme/set-by-portal")
class _R: pass
r = _R(); r.model = "garageai/auto"; r._hidden_params = {"model_id": "auto__dgx-city__mimo__dedicated"}
cb._canon_cache.update({"map": cn, "at": 9e18, "router": object()})
import types
orig = cb._auto_names; cb._auto_names = lambda: cn
check("auto: response named", cb.name_the_model(r, {"model": "garageai/auto"}).model, "xiaomi/mimo-v2.6-flash")
r2 = _R(); r2.model = "x"; r2._hidden_params = {"model_id": "auto__dgx-city__mimo__dedicated"}
check("other models untouched", cb.name_the_model(r2, {"model": "qwen/qwen3-4b"}).model, "x")
b = b'data: {"id":"1","model":"garageai/auto","choices":[]}'
check("auto: bytes chunk named", cb.name_the_model(b, {"model": "garageai/auto", "metadata": {"model_info": {"id": "auto__autoversio__kluster__pool"}}}),
      b'data: {"id":"1","model":"deepseek/deepseek-v4.1-flash","choices":[]}')
cb._auto_names = orig


# ---- rule 10: owner first
cb._owner_until.clear()
dA = {"model_info": {"id": "a", "garage": "dgx-city"}}
dB = {"model_info": {"id": "b", "garage": "autoversio"}}
dS = {"model_info": {"id": "s", "garage": "shared-g", "owner_priority": "share"}}
cb.mark_owner_active("dgx-city", now=1000.0, minutes=5)
cb.mark_owner_active("shared-g", now=1000.0, minutes=5)
kept, wait = cb.apply_owner_hold([dA, dB, dS], now=1100.0)
check("hold: owner-active garage left out, others and 'share' kept", [d["model_info"]["id"] for d in kept], ["b", "s"])
check("hold: retry-after is the remaining hold", wait, 200)
check("hold: over after the window", [d["model_info"]["id"] for d in cb.apply_owner_hold([dA, dB], now=1000.0 + 301)[0]], ["a", "b"])
cb.mark_owner_active("dgx-city", now=1250.0, minutes=5)
check("hold: slides with each request", cb.apply_owner_hold([dA], now=1500.0)[0], [])
check("garage of a mine/ name", (cb.garage_of_mine("mine/dgx-city/xiaomi/mimo"), cb.garage_of_mine("garage/dgx-city/x/y")), ("dgx-city", None))
check("hold: private deployments are not market", cb.apply_owner_hold([{"model_info": {"id": "p", "garage": "dgx-city", "garage_tier": "private"}}], now=1300.0)[0][0]["model_info"]["id"], "p")
cb._owner_until.clear()
# rule 9 flag
d1 = {"model": "garageai/auto", "metadata": {}}
check("auto name: flag set for OpenAI chat", (cb.keep_served_name(d1, "acompletion"), d1["metadata"].get(cb.RAW_NAME_KEY)), (True, True))
d2 = {"model": "garageai/auto"}
check("auto name: not for Anthropic", cb.keep_served_name(d2, "anthropic_messages"), False)
check("auto name: not for other models", cb.keep_served_name({"model": "qwen/qwen3-4b"}, "acompletion"), False)
cb._auto_names = lambda: {}
r3 = _R(); r3.model = "openai/kluster"; r3._hidden_params = {"model_id": "unknown"}
check("auto name: never a raw provider id (falls back to garageai/auto)", cb.name_the_model(r3, {"model": "garageai/auto"}).model, "garageai/auto")
cb._auto_names = orig

# ---- rule 11: a start for garages that wait
def dep(name, gid, garage, tier, order=None, canon=None, private=False):
    info = {"id": gid, "garage": garage, "garage_tier": "private" if private else tier}
    if canon:
        info["canonical_model"] = canon
    lp = {"model": "openai/" + gid}
    if order is not None:
        lp["order"] = order
    return {"model_name": name, "litellm_params": lp, "model_info": info}


ML = [dep("garageai/auto", "auto-av", "autoversio", "auto", 1, "deepseek/deepseek-v4.1-flash"),
      dep("garageai/auto", "auto-south", "autoversio-south", "auto", 3, "qwen/qwen3.8-27b"),
      dep("qwen/qwen3.8-27b", "pool-south", "autoversio-south", "pool"),
      dep("qwen/qwen3.8-27b", "pool-5090", "home-5090", "pool"),
      dep("qwen/qwen3.8-27b", "pool-mac", "mac-mini", "pool"),
      dep("qwen/qwen3-14b", "pool-cpu", "cpu-box", "pool"),
      dep("mine/home-5090/qwen/qwen3.8-27b", "mine-5090", "home-5090", "private", private=True)]
cb._served_at.clear(); cb._boost_failed.clear(); cb._owner_until.clear(); cb._sticky.clear(); cb._speeds.clear()
ids = lambda ds: sorted((d.get("model_info") or {}).get("id") for d in ds)
check("start: same model as an auto entry, not on the list, market only", ids(cb.boost_candidates(ML)), ["pool-5090", "pool-mac"])
for _ in range(cb.BOOST_TARGET):
    cb.note_served("mac-mini")
check("start: a garage that served enough no longer waits", ids(cb.boost_candidates(ML)), ["pool-5090"])
cb._served_at.clear()
check("start: not for a big prompt", cb.choose_boost(ML, 50000, None, roll=0.0), None)
check("start: not when the roll says no", cb.choose_boost(ML, 2000, None, roll=0.9), None)
check("start: not when max_tokens leaves a thinking model no room", cb.choose_boost(ML, 2000, None, roll=0.0, max_output=20), None)
check("start: a large or unset max_tokens still gets one", cb.choose_boost(ML, 2000, None, roll=0.0, max_output=4000) is not None and cb.choose_boost(ML, 2000, None, roll=0.0) is not None, True)
cb.note_served("mac-mini")
check("start: the garage that waited most first", cb.choose_boost(ML, 2000, None, roll=0.0)["model_info"]["id"], "pool-5090")
cb._boost_failed["home-5090"] = time.time()
check("start: a failed garage pauses", cb.choose_boost(ML, 2000, None, roll=0.0)["model_info"]["id"], "pool-mac")
cb._boost_failed.clear()
cb.mark_owner_active("home-5090")
check("start: an owner-held garage is skipped", cb.choose_boost(ML, 2000, None, roll=0.0)["model_info"]["id"], "pool-mac")
cb._owner_until.clear()
cb._speeds["pool-5090"] = cb.deque([10.0] * 5); cb._speeds["pool-mac"] = cb.deque([10.0] * 5)
check("start: not when prefill would be too slow", cb.choose_boost(ML, 2000, None, roll=0.0), None)
cb._speeds.clear()
cb._sticky["c1"] = ("pool-mac", 1500, time.time())
check("start: a conversation goes back to the garage that served it", cb.choose_boost(ML, 2000, "c1", roll=0.99)["model_info"]["id"], "pool-mac")
cb._sticky["c2"] = ("auto-av", 1500, time.time())
check("start: a conversation cached elsewhere is never taken", cb.choose_boost(ML, 2000, "c2", roll=0.0), None)
cb._sticky["c3"] = ("auto-av", 1500, time.time() - cb.STICKY_S - 1)
check("start: once its cache is stale it may be offered one", cb.choose_boost(ML, 2000, "c3", roll=0.0) is not None, True)
cb._sticky.clear()
f = cb.first_in_line(ML[3])
check("start: put first (order 0) on a copy", (f["litellm_params"]["order"], f["model_info"]["order"], "order" in ML[3]["litellm_params"]), (0, 0, False))
import litellm.utils as lu
check("start: the router's order filter keeps only it", ids(lu.get_order_filtered_deployments([f, ML[0], ML[1]])), ["pool-5090"])
check("start: when it is dropped, auto's order holds", ids(lu.get_order_filtered_deployments([ML[0], ML[1]])), ["auto-av"])
cb._boost_ids["pool-5090"] = "qwen/qwen3.8-27b"
orig = cb._auto_names; cb._auto_names = lambda: {}
r4 = _R(); r4.model = "garageai/auto"; r4._hidden_params = {"model_id": "pool-5090"}
check("start: the buyer sees the model's name", cb.name_the_model(r4, {"model": "garageai/auto"}).model, "qwen/qwen3.8-27b")
cb._auto_names = orig

# Rule 12 (2026-11-05): the runtime's token count is checked against our own.
uc = cb.usage_check
check("usage: an honest count adds up", uc(300, 1000, 280, 950), None)
check("usage: tokenizer slack allowed", uc(int(280 * 1.3) + 60, 1000, 280, 950), None)
check("usage: a padded answer count is caught", uc(100000, 1000, 5, 950) is not None, True)
check("usage: hidden reasoning is taken off the reported count", uc(500, 1000, 60, 950, reasoning_tokens=440), None)
check("usage: shown reasoning is counted with the answer", uc(500, 1000, 480, 950, reasoning_tokens=440, reasoning_shown=True), None)
check("usage: a padded prompt count is caught", "prompt" in (uc(10, 50000, 10, 1000) or ""), True)
check("usage: images and tools make the prompt looser", uc(10, 2400, 10, 1000), None)
check("usage: prompt not counted -> not checked", uc(10, 50000, 10, 0), None)
check("usage: a tool call gets the wider margin (measured: 108 reported for 8 counted)", uc(108, 400, 8, 390, tool_call=True), None)
check("usage: tool call, 439 reported for 253 counted, adds up", uc(439, 18752, 253, 18000, tool_call=True), None)
check("usage: a padded tool call is still caught", uc(50000, 400, 8, 390, tool_call=True) is not None, True)
check("usage: plain text keeps the tight margin", uc(108, 400, 8, 390) is not None, True)
ans = {"id": "chatcmpl-1", "object": "chat.completion", "model": "qwen", "usage": {"completion_tokens": 9},
       "choices": [{"index": 0, "finish_reason": "stop", "message": {"role": "assistant", "content": "Kalmar",
                    "tool_calls": [{"id": "c1", "type": "function", "function": {"name": "f", "arguments": "{\"x\": 1}"}}]}}]}
t, shown = cb.answer_text(ans)
check("usage: answer text = content + tool arguments, not ids or usage", ("Kalmar" in t, '{"x": 1}' in t, "chatcmpl" in t, "stop" in t, shown), (True, True, False, False, False))
check("usage: a tool call is seen as one", (cb.has_tool_call(ans), cb.has_tool_call({"choices": [{"message": {"content": "hi", "tool_calls": None}}]})), (True, False))
empty = {"choices": [{"message": {"content": "", "tool_calls": [
    {"id": "c1", "type": "function", "function": {"name": "f", "arguments": "{}"}},
    {"id": "c2", "type": "function", "function": {"name": "g", "arguments": ""}},
    {"id": "c3", "type": "function", "function": {"name": "h", "arguments": "{\"path\": \"a\"}"}}]}}]}
check("tool: empty arguments are counted, real ones not", (cb.empty_tool_calls(empty), cb.empty_tool_calls(ans), cb.empty_tool_calls({"choices": [{"message": {"content": "hi"}}]})), (2, 0, 0))
ans["choices"][0]["message"]["reasoning_content"] = "let me think"
check("usage: reasoning text counts as shown", cb.answer_text(ans)[1], True)
cb._suspect.clear()
check("usage: nothing suspect -> list untouched", cb.drop_suspects(ML) is ML, True)
cb._suspect["home-5090"] = time.time()
check("usage: a suspect garage is routed around", ids(cb.drop_suspects(ML)), ids([d for d in ML if d["model_info"]["garage"] != "home-5090"]))
check("usage: and comes back after SUSPECT_S", ids(cb.drop_suspects(ML, now=time.time() + cb.SUSPECT_S + 1)), ids(ML))
cb._suspect.clear()


async def _run_check():
    cb._suspect.clear()
    kw = {"standard_logging_object": {"model_group": "qwen/qwen3.8-27b", "model_id": "pool-5090", "id": "call-1",
                                      "completion_tokens": 50000, "prompt_tokens": 20, "metadata": {"usage_object": {}}},
          "messages": [U("Name one Swedish city.")]}
    why = await cb.check_usage(kw, ans)
    kw["standard_logging_object"]["completion_tokens"] = 12
    ok = await cb.check_usage(kw, ans)
    kw["standard_logging_object"]["model_group"] = "mine/home-5090/qwen/qwen3.8-27b"
    kw["standard_logging_object"]["completion_tokens"] = 50000
    mine = await cb.check_usage(kw, ans)
    return why is not None, ok, mine

import asyncio as _aio
check("usage: the success hook flags a padded count, passes an honest one, ignores mine/", _aio.run(_run_check()), (True, None, None))

# Rule 13 (2026-11-07): thinking off unless asked for.
def _dk(runtime, group="garageai/auto", **extra):
    return {"model": "openai/x", "messages": [U("hi")], "model_info": {"id": "d1", "garage": "g", "runtime": runtime},
            "litellm_params": {"metadata": {"model_group": group}}, **extra}
check("thinking: off on ollama for a market request", cb.quiet_thinking(_dk("ollama")), ({**_dk("ollama"), "reasoning_effort": "none"}, True))
check("thinking: off on vllm", cb.quiet_thinking(_dk("vllm"))[1], True)
check("thinking: a provider is left alone", cb.quiet_thinking(_dk("openai"))[1], False)
check("thinking: Runner is left alone", cb.quiet_thinking(_dk("xyntetik"))[1], False)
check("thinking: the owner's own machine is left alone", cb.quiet_thinking(_dk("ollama", group="mine/g/qwen/qwen3"))[1], False)
check("thinking: a probe is left alone", cb.quiet_thinking(_dk("ollama", group="probe/g"))[1], False)
for k, v in (("reasoning_effort", "high"), ("thinking", {"type": "enabled", "budget_tokens": 1024}), ("reasoning", {"effort": "low"}),
             ("chat_template_kwargs", {"enable_thinking": True}), ("extra_body", {"chat_template_kwargs": {"enable_thinking": True}})):
    kw, quiet = cb.quiet_thinking(_dk("ollama", **{k: v}))
    check(f"thinking: the buyer's own {k} is kept", (quiet, kw.get(k), kw.get("reasoning_effort") if k != "reasoning_effort" else kw[k]), (False, v, None if k != "reasoning_effort" else v))

print("all tests passed" if not fails else f"{fails} test(s) failed")
sys.exit(1 if fails else 0)

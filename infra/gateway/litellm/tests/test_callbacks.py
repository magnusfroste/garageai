"""Unit tests for garageai_callbacks.py. They need LiteLLM, so run them in the LiteLLM container:
  sudo docker exec -i garageai-litellm-litellm-1 python - < infra/gateway/litellm/tests/test_callbacks.py
with the new garageai_callbacks.py copied in first (see infra/gateway/README.md), or after a deploy."""
import importlib.util
import sys

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

print("all tests passed" if not fails else f"{fails} test(s) failed")
sys.exit(1 if fails else 0)

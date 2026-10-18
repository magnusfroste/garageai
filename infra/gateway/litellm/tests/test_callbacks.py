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
check("provider 401: buyer gets 503, not 401", (type(away).__name__, str(away.code)), ("ModelTemporarilyUnavailable", "503"))
check("provider 401: auto suggested", "garageai/auto" in away.message, True)
cool = cb.provider_unavailable(Exception("No deployments available for selected model, Try again in 60 seconds"), {"model": "garageai/auto"})
check("all cooling down: 503", str(cool.code), "503")
check("auto does not suggest itself", "use garageai/auto" in cool.message, False)
check("other errors: untouched", cb.provider_unavailable(ValueError("boom"), {"model": "x"}), None)

print("all tests passed" if not fails else f"{fails} test(s) failed")
sys.exit(1 if fails else 0)

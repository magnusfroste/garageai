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
check("tool messages keep their order", h([S("a"), U("q"), {"role": "assistant", "tool_calls": [1]}, {"role": "tool", "content": "t"}, S("b")]),
      [S("a\n\nb"), U("q"), {"role": "assistant", "tool_calls": [1]}, {"role": "tool", "content": "t"}])
print("all tests passed" if not fails else f"{fails} test(s) failed")
sys.exit(1 if fails else 0)

package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func mcpRun(t *testing.T, lines ...string) []map[string]any {
	t.Helper()
	var out bytes.Buffer
	if code := mcpServe(strings.NewReader(strings.Join(lines, "\n")+"\n"), &out); code != 0 {
		t.Fatalf("exit %d", code)
	}
	var res []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("bad line %q: %v", l, err)
		}
		res = append(res, m)
	}
	return res
}

func text(t *testing.T, m map[string]any) string {
	t.Helper()
	r, _ := m["result"].(map[string]any)
	c, _ := r["content"].([]any)
	if len(c) == 0 {
		t.Fatalf("no content: %v", m)
	}
	return c[0].(map[string]any)["text"].(string)
}

func TestMCP(t *testing.T) {
	res := mcpRun(t,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"explain_problem","arguments":{"code":"mac_sleeps"}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"search_knowledge","arguments":{"query":"ollama only answers on localhost"}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"explain_problem","arguments":{"code":"nope"}}}`,
		`{"jsonrpc":"2.0","id":6,"method":"nope"}`,
		`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"garage_status","arguments":{}}}`,
	)
	if len(res) != 7 {
		t.Fatalf("want 7 replies (the notification gets none), got %d", len(res))
	}
	init := res[0]["result"].(map[string]any)
	if init["protocolVersion"] != "2025-03-26" || init["instructions"] == nil {
		t.Fatalf("initialize: %v", init)
	}
	tools := res[1]["result"].(map[string]any)["tools"].([]any)
	names := []string{}
	for _, x := range tools {
		names = append(names, x.(map[string]any)["name"].(string))
	}
	if strings.Join(names, ",") != "doctor,explain_problem,search_knowledge,garage_status" {
		t.Fatalf("tools: %v", names)
	}
	if s := text(t, res[2]); !strings.Contains(s, "Mac Studio") {
		t.Fatalf("explain: %s", s)
	}
	if s := text(t, res[3]); !strings.Contains(s, "localhost_only") {
		t.Fatalf("search: %s", s)
	}
	if res[4]["result"].(map[string]any)["isError"] != true {
		t.Fatalf("unknown code should be a tool error: %v", res[4])
	}
	if res[5]["error"] == nil {
		t.Fatalf("unknown method: %v", res[5])
	}
	if s := text(t, res[6]); !strings.Contains(s, "bridge_version") || strings.Contains(s, "register_token") {
		t.Fatalf("status: %s", s)
	}
}

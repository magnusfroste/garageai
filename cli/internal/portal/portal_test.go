package portal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type seen struct {
	sync.Mutex
	calls []map[string]any
	paths []string
	auth  []string
}

func fakePortal(t *testing.T, status int, reply string) (*Client, *seen) {
	t.Helper()
	s := &seen{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		s.Lock()
		s.calls = append(s.calls, m)
		s.paths = append(s.paths, r.URL.Path)
		s.auth = append(s.auth, r.Header.Get("Authorization"))
		s.Unlock()
		w.WriteHeader(status)
		w.Write([]byte(reply))
	}))
	t.Cleanup(srv.Close)
	return New(srv.URL+"/functions/v1/register-node", "tok-1"), s
}

func TestRegisterPayloadMatchesTheScript(t *testing.T) {
	c, s := fakePortal(t, 200, `{"acceptance":[{"model":"qwen3:4b","passed":true,"tokens_per_second":29.3,"ttft_ms":565}]}`)
	res, err := c.Register(context.Background(), Registration{Name: "m2", MeshIP: "100.83.1.2", Port: 11434, Runtime: "ollama",
		Models: []string{"qwen3:4b"}, Contexts: map[string]int{"qwen3:4b": 8192}, RuntimeAPIKey: ""})
	if err != nil || len(res.Acceptance) != 1 || !res.Acceptance[0].Passed || *res.Acceptance[0].TokensPerSecond != 29.3 {
		t.Fatalf("register: %v %+v", err, res)
	}
	got := s.calls[0]
	for _, k := range []string{"name", "mesh_ip", "port", "runtime", "models", "contexts"} {
		if _, ok := got[k]; !ok {
			t.Errorf("payload lacks %s: %v", k, got)
		}
	}
	if _, ok := got["runtime_api_key"]; ok {
		t.Error("an empty runtime_api_key was sent; the script omits it")
	}
	if _, ok := got["context_length"]; ok {
		t.Error("context_length 0 was sent; the script omits it")
	}
	if s.auth[0] != "Bearer tok-1" || !strings.HasSuffix(s.paths[0], "/register-node") {
		t.Errorf("auth %q path %q", s.auth[0], s.paths[0])
	}
}

func TestHeartbeatRejected(t *testing.T) {
	c, s := fakePortal(t, 401, `{"error":"Unauthorized"}`)
	_, err := c.Heartbeat(context.Background(), Registration{Name: "m2", Port: 11434, Runtime: "ollama", Models: []string{}})
	if !errors.Is(err, ErrRejected) {
		t.Fatalf("a 401 is not reported as rejected: %v", err)
	}
	if !strings.HasSuffix(s.paths[0], "/node-heartbeat") {
		t.Errorf("heartbeat went to %s", s.paths[0])
	}
	if m, ok := s.calls[0]["models"].([]any); !ok || len(m) != 0 {
		t.Errorf("an empty model list must be sent as [] so the portal stops routing: %v", s.calls[0])
	}
}

func TestReportIsBestEffort(t *testing.T) {
	c, s := fakePortal(t, 204, "")
	c.Report(Report{Step: "3/6  Inference runtime", Status: "failed", Message: strings.Repeat("x", 600), NodeName: "m2"})
	if len(s.calls) != 1 || !strings.HasSuffix(s.paths[0], "/onboarding-report") {
		t.Fatalf("report not sent: %v", s.paths)
	}
	if msg := s.calls[0]["message"].(string); len(msg) != 500 {
		t.Errorf("message not truncated to 500: %d", len(msg))
	}
	dead := New("http://127.0.0.1:9/functions/v1/register-node", "tok")
	dead.Report(Report{Step: "x", Status: "started"}) // must not panic or block
	t.Setenv("GARAGEAI_REPORT", "0")
	c.Report(Report{Step: "x", Status: "started"})
	if len(s.calls) != 1 {
		t.Error("GARAGEAI_REPORT=0 still sent a report")
	}
}

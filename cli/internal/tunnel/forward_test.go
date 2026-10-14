package tunnel

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestForwardRewritesHostAndFiltersPaths(t *testing.T) {
	var seenHost, seenPath string
	runtime := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenHost, seenPath = r.Host, r.URL.Path
		w.Write([]byte(`{"data":[{"id":"qwen3:4b"}]}`))
	}))
	defer runtime.Close()
	local := strings.TrimPrefix(runtime.URL, "http://")
	fw := httptest.NewServer(ForwardHandler(local, nil))
	defer fw.Close()

	req, _ := http.NewRequest("GET", fw.URL+"/v1/models", nil)
	req.Host = "10.66.0.4:11434" // what the gateway sends through the tunnel
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("models: %v %v", err, resp)
	}
	if seenHost != local || seenPath != "/v1/models" {
		t.Fatalf("runtime saw Host %q path %q, want %q /v1/models", seenHost, seenPath, local)
	}
	for _, path := range []string{"/api/pull", "/api/delete", "/api/create", "/unload", "/v1/runner/contexts", "/", "/v1/modelsx"} {
		seenPath = ""
		resp, err := http.Post(fw.URL+path, "application/json", strings.NewReader(`{"model":"x"}`))
		if err != nil || resp.StatusCode != http.StatusForbidden {
			t.Fatalf("%s: want 403, got %v %v", path, err, resp)
		}
		if seenPath != "" {
			t.Fatalf("%s reached the runtime", path)
		}
	}
	for _, path := range []string{"/v1/chat/completions", "/v1/messages", "/api/ps", "/metrics", "/v1/models/qwen3:4b"} {
		if !Allowed(path) {
			t.Errorf("%s should be allowed", path)
		}
	}
}

func TestForwardStreamsWithoutBuffering(t *testing.T) {
	runtime := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		f := w.(http.Flusher)
		for i := 0; i < 3; i++ {
			fmt.Fprintf(w, "data: {\"n\":%d}\n\n", i)
			f.Flush()
			time.Sleep(150 * time.Millisecond)
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer runtime.Close()
	fw := httptest.NewServer(ForwardHandler(strings.TrimPrefix(runtime.URL, "http://"), nil))
	defer fw.Close()
	t0 := time.Now()
	resp, err := http.Post(fw.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{"stream":true}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	br := bufio.NewReader(resp.Body)
	first, _ := br.ReadString('\n')
	firstAt := time.Since(t0)
	rest, _ := io.ReadAll(br)
	if !strings.HasPrefix(first, "data: {\"n\":0}") || !strings.Contains(string(rest), "[DONE]") {
		t.Fatalf("stream: first %q rest %q", first, rest)
	}
	if firstAt > 120*time.Millisecond {
		t.Fatalf("first chunk arrived after %v: the forward buffers the stream", firstAt)
	}
}

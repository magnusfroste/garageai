package runtime

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func fake(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	port, _ := strconv.Atoi(srv.URL[strings.LastIndex(srv.URL, ":")+1:])
	return New("127.0.0.1", port, "")
}

func TestModelsAndContexts(t *testing.T) {
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			w.Write([]byte(`{"data":[{"id":"a","max_model_len":262144},{"id":"b","meta":{"n_ctx_train":8192}},{"id":"c"}]}`))
		case "/api/ps":
			w.Write([]byte(`{"models":[{"name":"c","context_length":4096}]}`))
		}
	})
	ids, ctx, err := c.Models()
	if err != nil || len(ids) != 3 || ctx["a"] != 262144 || ctx["b"] != 8192 {
		t.Fatalf("models %v ctx %v err %v", ids, ctx, err)
	}
	if _, ok := ctx["c"]; ok {
		t.Error("a model without a window got one")
	}
	if c.OllamaLoadedContext("c") != 4096 || c.OllamaLoadedContext("a") != 0 {
		t.Error("Ollama loaded windows wrong")
	}
}

func TestAuthAndUsage(t *testing.T) {
	c := fake(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer k" {
			w.WriteHeader(401)
			w.Write([]byte(`{"error":"key"}`))
			return
		}
		if r.Method == "POST" {
			w.Write([]byte("data: {\"choices\":[]}\n\ndata: {\"usage\":{\"prompt_tokens\":1}}\n\ndata: [DONE]\n"))
			return
		}
		w.Write([]byte(`{"data":[{"id":"m"}]}`))
	})
	if _, _, err := c.Models(); !errors.Is(err, ErrAuth) {
		t.Fatalf("missing key not reported as auth: %v", err)
	}
	c.Key = "k"
	if ids, _, err := c.Models(); err != nil || len(ids) != 1 {
		t.Fatalf("with key: %v %v", ids, err)
	}
	if !c.UsageReported(context.Background(), "m") {
		t.Error("usage in the stream not detected")
	}
}

func TestOffer(t *testing.T) {
	served := []string{"qwen3:4b", "nomic-embed-text:latest", "bge-m3"}
	off, skip := Offer(served, "")
	if len(off) != 1 || off[0] != "qwen3:4b" || len(skip) != 2 {
		t.Errorf("default offer %v skip %v", off, skip)
	}
	off, skip = Offer(served, "bge-m3, qwen3:4b")
	if len(off) != 2 || len(skip) != 1 {
		t.Errorf("explicit offer %v skip %v", off, skip)
	}
	if DefaultPort("ollama") != 11434 || DefaultPort("other") != 0 {
		t.Error("default ports")
	}
}

// Idle Ollama models get a window: num_ctx from the Modelfile, else the server default, capped by
// the model's maximum; a loaded model reports what it runs with.
func TestOllamaContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/ps":
			w.Write([]byte(`{"models":[{"name":"loaded:7b","model":"loaded:7b","context_length":16384}]}`))
		case "/api/show":
			b, _ := io.ReadAll(r.Body)
			switch {
			case strings.Contains(string(b), "custom"):
				w.Write([]byte(`{"parameters":"num_ctx                        8192\ntemperature 1","model_info":{"llama.context_length":131072}}`))
			case strings.Contains(string(b), "small"):
				w.Write([]byte(`{"parameters":"","model_info":{"phi.context_length":2048}}`))
			default:
				w.Write([]byte(`{"parameters":"","model_info":{"qwen3.context_length":262144}}`))
			}
		}
	}))
	defer srv.Close()
	port, _ := strconv.Atoi(srv.URL[strings.LastIndex(srv.URL, ":")+1:])
	c := New("127.0.0.1", port, "")
	for model, tc := range map[string]struct{ def, want int }{
		"loaded:7b": {0, 16384}, "custom:1b": {0, 8192}, "plain:4b": {32768, 32768}, "plain2:4b": {0, 4096}, "small:1b": {65536, 2048},
	} {
		if got := c.OllamaContext(model, tc.def); got != tc.want {
			t.Errorf("%s: got %d, want %d", model, got, tc.want)
		}
	}
}

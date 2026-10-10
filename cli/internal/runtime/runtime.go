// Package runtime asks an OpenAI-compatible inference runtime what it serves, the way the connect
// script does: models, whether streamed replies carry usage (needed for per-token billing), and
// each model's context window.
package runtime

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// NonChat models are never offered: embedding and reranker models cannot chat.
var NonChat = regexp.MustCompile(`(?i)embed|rerank|bge-|e5-|minilm|clip|whisper|tts`)

func DefaultPort(runtime string) int {
	return map[string]int{"ollama": 11434, "lmstudio": 1234, "llamacpp": 8080, "vllm": 8000, "sglang": 30000,
		"paddock": 11540, "unsloth": 8888, "mlx": 8080, "lemonade": 13305}[runtime]
}

type Client struct {
	Host, Key string
	Port      int
	HTTP      *http.Client
}

func New(host string, port int, key string) *Client {
	return &Client{Host: host, Port: port, Key: key, HTTP: &http.Client{Timeout: 5 * time.Second}}
}

func (c *Client) base() string {
	h := c.Host
	if strings.Contains(h, ":") {
		h = "[" + h + "]"
	}
	return fmt.Sprintf("http://%s:%d", h, c.Port)
}

func (c *Client) get(path string) (int, []byte, error) {
	req, _ := http.NewRequest("GET", c.base()+path, nil)
	if c.Key != "" {
		req.Header.Set("Authorization", "Bearer "+c.Key)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	return resp.StatusCode, b, nil
}

// Models lists the model ids /v1/models serves, and the context window of each where reported.
// ErrAuth when the runtime wants a (different) API key.
func (c *Client) Models() (ids []string, contexts map[string]int, err error) {
	code, b, err := c.get("/v1/models")
	if err != nil {
		return nil, nil, err
	}
	if code == 401 || code == 403 {
		return nil, nil, ErrAuth
	}
	var list struct {
		Data []map[string]any `json:"data"`
	}
	if code != 200 || json.Unmarshal(b, &list) != nil || list.Data == nil {
		return nil, nil, fmt.Errorf("no OpenAI-compatible API (HTTP %d)", code)
	}
	contexts = map[string]int{}
	for _, m := range list.Data {
		id, _ := m["id"].(string)
		if id == "" {
			continue
		}
		ids = append(ids, id)
		for _, k := range []string{"max_model_len", "context_length", "context_window"} {
			if f, ok := m[k].(float64); ok && f > 0 {
				contexts[id] = int(f)
				break
			}
		}
		if _, ok := contexts[id]; !ok {
			if meta, ok := m["meta"].(map[string]any); ok {
				if f, ok := meta["n_ctx_train"].(float64); ok && f > 0 {
					contexts[id] = int(f)
				}
			}
		}
	}
	return ids, contexts, nil
}

var ErrAuth = fmt.Errorf("the runtime requires an API key")

// OllamaLoadedContext is the window Ollama runs a model with, from /api/ps (0 when not loaded).
func (c *Client) OllamaLoadedContext(model string) int {
	code, b, err := c.get("/api/ps")
	if err != nil || code != 200 {
		return 0
	}
	var ps struct {
		Models []struct {
			Name, Model   string
			ContextLength int `json:"context_length"`
		} `json:"models"`
	}
	if json.Unmarshal(b, &ps) != nil {
		return 0
	}
	for _, m := range ps.Models {
		if m.Name == model || m.Model == model {
			return m.ContextLength
		}
	}
	return 0
}

// UsageReported sends a 1-token streamed request and looks for a usage object in the stream.
// Buyers pay per token, so the runtime must report usage in streamed replies.
func (c *Client) UsageReported(ctx context.Context, model string) bool {
	body, _ := json.Marshal(map[string]any{"model": model, "max_tokens": 1, "stream": true,
		"stream_options": map[string]bool{"include_usage": true},
		"messages":       []map[string]string{{"role": "user", "content": "hi"}}})
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "POST", c.base()+"/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if c.Key != "" {
		req.Header.Set("Authorization", "Bearer "+c.Key)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	usage := regexp.MustCompile(`"usage"\s*:\s*\{`)
	for sc.Scan() {
		if usage.Match(sc.Bytes()) {
			return true
		}
	}
	return false
}

// Offer picks the models to sell: the --models list when given, else every chat model.
func Offer(served []string, wanted string) (offered, skipped []string) {
	if wanted != "" {
		want := map[string]bool{}
		for _, w := range strings.Split(wanted, ",") {
			if w = strings.TrimSpace(w); w != "" {
				want[w] = true
			}
		}
		for _, m := range served {
			if want[m] {
				offered = append(offered, m)
			} else {
				skipped = append(skipped, m)
			}
		}
		return
	}
	for _, m := range served {
		if NonChat.MatchString(m) {
			skipped = append(skipped, m)
		} else {
			offered = append(offered, m)
		}
	}
	return
}

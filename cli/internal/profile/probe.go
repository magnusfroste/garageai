package profile

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var client = &http.Client{Timeout: 3 * time.Second}

// probe asks one listening port whether it is an OpenAI-compatible runtime. nil when it is not.
func probe(port int, binds []string, owner Listener) *Runtime {
	host := "127.0.0.1"
	if !anyOf(binds, "127.0.0.1", "0.0.0.0", "*", "::", "::1") && len(binds) > 0 {
		host = binds[0]
	}
	base := "http://" + hostport(host, port)
	code, body := get(base + "/v1/models")
	if code != 200 && code != 401 && code != 403 {
		return nil
	}
	var list struct {
		Data []map[string]any `json:"data"`
	}
	if code == 200 {
		if json.Unmarshal(body, &list) != nil || list.Data == nil {
			return nil
		}
	} else {
		// An OpenAI-compatible server rejects a missing key with a JSON error. A 401/403 without
		// one is something else: macOS's AirPlay receiver answers 403 on port 5000 to everything.
		var e map[string]any
		if json.Unmarshal(body, &e) != nil {
			return nil
		}
	}
	rt := &Runtime{Port: port, Binds: binds, Models: []Model{}, Process: strp(owner.Process)}
	rt.API = "openai"
	if code != 200 {
		rt.API = "openai (needs API key)"
	}
	for _, m := range list.Data {
		id, _ := m["id"].(string)
		md := Model{ID: id, Context: firstInt(m, "max_model_len", "context_length", "context_window")}
		if md.Context == nil {
			if meta, ok := m["meta"].(map[string]any); ok {
				md.Context = firstInt(meta, "n_ctx_train")
			}
		}
		if o, ok := m["owned_by"].(string); ok {
			md.OwnedBy = &o
		}
		rt.Models = append(rt.Models, md)
	}
	rt.Network = anyOf(binds, "0.0.0.0", "*", "::") || anyPrefix(binds, "100.")

	if c, b := get(base + "/api/version"); c == 200 {
		var v struct{ Version string }
		if json.Unmarshal(b, &v) == nil && v.Version != "" {
			rt.Kind, rt.Version = "ollama", &v.Version
			rt.OllamaLoaded = ollamaLoaded(base)
		}
	}
	if rt.Kind == "" {
		owned := ""
		if len(rt.Models) > 0 && rt.Models[0].OwnedBy != nil {
			owned = *rt.Models[0].OwnedBy
		}
		rt.Kind = kindOf(owned + " " + strings.ToLower(owner.Process))
	}
	if owner.PID > 0 {
		args := cmdline(owner.PID)
		if rt.Kind == "vllm" || rt.Kind == "sglang" || strings.Contains(strings.Join(args, " "), "vllm") {
			rt.Flags = readFlags(args)
		}
		if rt.Kind == "other" && owner.Process != "" {
			rt.Kind = sanitize(strings.ToLower(owner.Process))
		}
	}
	return rt
}

func kindOf(s string) string {
	switch {
	case strings.Contains(s, "vllm"):
		return "vllm"
	case strings.Contains(s, "sglang"):
		return "sglang"
	case strings.Contains(s, "llama"):
		return "llamacpp"
	case strings.Contains(s, "lms") || strings.Contains(s, "lm studio"):
		return "lmstudio"
	case strings.Contains(s, "lemonade"):
		return "lemonade"
	case strings.Contains(s, "mlx"):
		return "mlx"
	case strings.Contains(s, "unsloth"):
		return "unsloth"
	case strings.Contains(s, "paddock"):
		return "paddock"
	}
	return "other"
}

// readFlags takes only these flags; a command line can hold an API key, which is never reported.
func readFlags(args []string) Flags {
	val := func(flag string) string {
		for i, a := range args {
			if a == flag && i+1 < len(args) {
				return args[i+1]
			}
			if strings.HasPrefix(a, flag+"=") {
				return a[len(flag)+1:]
			}
		}
		return ""
	}
	num := func(s string) *int {
		if n, err := strconv.Atoi(s); err == nil {
			return &n
		}
		return nil
	}
	key, details := false, false
	for _, a := range args {
		key = key || a == "--api-key" || strings.HasPrefix(a, "--api-key=")
		details = details || a == "--enable-prompt-tokens-details"
	}
	return Flags{Host: strp(val("--host")), Port: num(val("--port")), MaxModelLen: num(val("--max-model-len")),
		ContextLength: num(val("--context-length")), APIKeySet: &key, PromptTokensDetails: &details}
}

func ollamaLoaded(base string) []LoadedModel {
	c, b := get(base + "/api/ps")
	if c != 200 {
		return nil
	}
	var ps struct {
		Models []map[string]any `json:"models"`
	}
	if json.Unmarshal(b, &ps) != nil {
		return nil
	}
	out := []LoadedModel{}
	for _, m := range ps.Models {
		name, _ := m["name"].(string)
		out = append(out, LoadedModel{Name: name, Context: firstInt(m, "context_length")})
	}
	return out
}

func get(url string) (int, []byte) {
	resp, err := client.Get(url)
	if err != nil {
		return 0, nil
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	return resp.StatusCode, b
}

func firstInt(m map[string]any, keys ...string) *int {
	for _, k := range keys {
		if f, ok := m[k].(float64); ok && f > 0 {
			return intp(int(f))
		}
	}
	return nil
}

func hostport(host string, port int) string {
	if strings.Contains(host, ":") {
		return "[" + host + "]:" + itoa(port)
	}
	return host + ":" + itoa(port)
}

func anyOf(list []string, vals ...string) bool {
	for _, a := range list {
		for _, v := range vals {
			if a == v {
				return true
			}
		}
	}
	return false
}

func anyPrefix(list []string, prefix string) bool {
	for _, a := range list {
		if strings.HasPrefix(a, prefix) {
			return true
		}
	}
	return false
}

func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, s)
}

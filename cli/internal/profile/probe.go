package profile

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RuntimeKey is the runtime's API key (from connect or the heartbeat's config), used only when
// probing RuntimeKeyPort (0: any port, during connect before the port is known).
var (
	RuntimeKey     string
	RuntimeKeyPort int
	// KnownRuntime is the runtime kind the garage was connected with (bridge.json), for the port
	// RuntimeKeyPort. The heartbeat probes that port every 5 minutes; once the kind is known there
	// is no reason to ask it the other runtimes' questions (a 404 for /api/version in vLLM's log
	// every 5 minutes on dgx-city, 2026-10-18) or to try without the key first (a 401 line).
	KnownRuntime string
)

func knownPort(port int) bool { return RuntimeKeyPort != 0 && port == RuntimeKeyPort }

var client = &http.Client{Timeout: 3 * time.Second}

// probe asks one listening port whether it is an OpenAI-compatible runtime. nil when it is not.
func probe(port int, binds []string, owner Listener) *Runtime {
	host := "127.0.0.1"
	if !anyOf(binds, "127.0.0.1", "0.0.0.0", "*", "::", "::1") && len(binds) > 0 {
		host = binds[0]
	}
	base := "http://" + hostport(host, port)
	var code int
	var body []byte
	if RuntimeKey != "" && knownPort(port) {
		code, body = getAuth(base+"/v1/models", RuntimeKey) // the configured port, with its key: no unauthenticated try first
	} else {
		code, body = get(base + "/v1/models")
	}
	if code != 200 && code != 401 && code != 403 {
		return nil
	}
	if (code == 401 || code == 403) && RuntimeKey != "" && (RuntimeKeyPort == 0 || port == RuntimeKeyPort) {
		// The operator gave the runtime's key (connect, or the heartbeat's config): with it the
		// model list, windows and owner are visible, not just "needs a key". Only to the configured
		// port: any local process could answer 401 on another port and collect the key.
		if c2, b2 := getAuth(base+"/v1/models", RuntimeKey); c2 == 200 {
			code, body = c2, b2
		}
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
	rt := &Runtime{Port: port, Binds: binds, Models: []Model{}, Process: strp(owner.Process), PID: owner.PID}
	rt.API = "openai"
	if code != 200 {
		rt.API = "openai (needs API key)"
	}
	for _, m := range list.Data {
		id, _ := m["id"].(string)
		// llama.cpp and LM Studio report file paths as ids; the path (with the user's name) stays
		// here, the file name is what the profile carries.
		if strings.HasPrefix(id, "/") || strings.HasPrefix(id, "~") || regexp.MustCompile(`^[A-Za-z]:\\`).MatchString(id) {
			id = filepath.Base(id)
		}
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

	if knownPort(port) && KnownRuntime != "" && KnownRuntime != "ollama" {
		rt.Kind = KnownRuntime // the garage's configured runtime on its configured port: no need to ask
	} else if c, b := get(base + "/api/version"); c == 200 {
		var v struct{ Version string }
		if json.Unmarshal(b, &v) == nil && v.Version != "" {
			rt.Kind, rt.Version = "ollama", &v.Version
			rt.OllamaLoaded = ollamaLoaded(base)
		}
	}
	if rt.Kind == "" && ollamaRunner(owner.Process, rt.Models) {
		// Ollama's internal runner: a localhost port per loaded model, not the API buyers reach.
		return nil
	}
	if rt.Kind == "" {
		owned := ""
		if len(rt.Models) > 0 && rt.Models[0].OwnedBy != nil {
			owned = *rt.Models[0].OwnedBy
		}
		rt.Kind = kindOf(owned + " " + strings.ToLower(owner.Process))
		if rt.Kind == "other" {
			// Behind Docker the process is docker-proxy, and with an API key the model list (and its
			// owned_by) is hidden. vLLM and SGLang leave /metrics open: their metric names say who it is.
			if k := metricsKind(base); k != "" {
				rt.Kind = k
			}
		}
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
	if rt.Kind == "xyntetik" {
		// Runner's model list has no window; /v1/capabilities has the one it runs with.
		if c, b := get(base + "/v1/capabilities"); c == 200 {
			var caps struct {
				Context float64 `json:"context"`
			}
			if json.Unmarshal(b, &caps) == nil && caps.Context > 0 {
				w := int(caps.Context)
				rt.RunnerContext = &w
				for i := range rt.Models {
					if rt.Models[i].Context == nil {
						rt.Models[i].Context = &w
					}
				}
			}
		}
	}
	return rt
}

// ollamaRunner: the process Ollama starts for a loaded model. Its name is "ollama" (Ollama's own
// engine) or "llama-server" (llama.cpp), and its model is a file in Ollama's blob store.
func ollamaRunner(process string, models []Model) bool {
	if strings.Contains(strings.ToLower(process), "ollama") {
		return true
	}
	for _, m := range models {
		id := strings.ReplaceAll(m.ID, `\`, "/")
		// On Windows the model is named by its blob alone ("sha256-7f40…") and the process
		// behind the port is not always known (w11, 2026-10-17).
		if strings.Contains(id, "/.ollama/models/blobs/") || strings.Contains(id, "/ollama/models/blobs/") || strings.HasPrefix(id, "sha256-") {
			return true
		}
	}
	return false
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
	case strings.Contains(s, "xyntetik") || s == "runner" || strings.HasSuffix(s, "/runner") || strings.HasPrefix(s, "runner "):
		return "xyntetik" // Xyntetik Runner: a single binary called `runner`
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

// metricsKind recognises vLLM and SGLang from their Prometheus metrics ("vllm:", "sglang:").
func metricsKind(base string) string {
	code, b := get(base + "/metrics")
	if code != 200 {
		return ""
	}
	s := string(b)
	switch {
	case strings.Contains(s, "vllm:"):
		return "vllm"
	case strings.Contains(s, "sglang:"):
		return "sglang"
	case strings.Contains(s, "runner_"):
		return "xyntetik" // Xyntetik Runner exports its counters under the runner_ prefix
	}
	return ""
}

func getAuth(url, key string) (int, []byte) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	return resp.StatusCode, b
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

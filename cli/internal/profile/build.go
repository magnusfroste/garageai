package profile

import (
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// KnownPorts are the default ports of the runtimes GarageAI supports; ports owned by a process
// that looks like a runtime are probed too, so an unexpected port is still found.
var KnownPorts = []int{11434, 1234, 8080, 8000, 30000, 11540, 8888, 13305, 5000, 5001, 8001}

var runtimeProc = regexp.MustCompile(`(?i)ollama|vllm|sglang|llama|lms|lm studio|lmstudio|lemonade|mlx|unsloth|paddock|python|uvicorn|docker-proxy|koboldcpp|tabby|text-generation|tgi|aphrodite|exllama`)

// Build discovers this machine and returns its garage profile.
func Build(version string) Profile {
	p := Profile{Schema: 1, ScriptVersion: version, Tool: "garageai-bridge", GeneratedAt: time.Now().UTC().Format(time.RFC3339)}
	var wg sync.WaitGroup
	var hb Heartbeat
	var cfg GarageAI
	wg.Add(4)
	go func() { defer wg.Done(); p.Machine = machine() }()
	go func() { defer wg.Done(); p.GPUs = gpus() }()
	go func() { defer wg.Done(); p.NetBird = netbird(); p.Firewall = firewall() }()
	go func() { defer wg.Done(); hb, cfg = heartbeatAndConfig() }()
	p.Runtimes = discover(listeners())
	wg.Wait()
	p.Heartbeat, p.GarageAI = hb, cfg
	p.darwin = runtime.GOOS == "darwin"
	p.heartbeatLast, p.ollamaHostPersistent, p.sleepMinutes, p.displayMinutes = localFacts(hb)
	p.Problems = problems(p)
	p.OK = true
	for _, pr := range p.Problems {
		if pr.Severity == "error" {
			p.OK = false
		}
	}
	return p
}

// discover probes candidate ports in parallel: known runtime ports and ports owned by a runtime-like process.
func discover(ls []Listener) []Runtime {
	ports, addrs, owner := byPort(ls)
	candidate := map[int]bool{}
	for _, kp := range KnownPorts {
		candidate[kp] = true
	}
	for _, port := range ports {
		if runtimeProc.MatchString(owner[port].Process) {
			candidate[port] = true
		}
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	out := []Runtime{}
	for _, port := range ports {
		if !candidate[port] {
			continue
		}
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			if rt := probe(port, addrs[port], owner[port]); rt != nil {
				mu.Lock()
				out = append(out, *rt)
				mu.Unlock()
			}
		}(port)
	}
	wg.Wait()
	sort.Slice(out, func(i, j int) bool { return out[i].Port < out[j].Port })
	return out
}

func problems(p Profile) []Problem {
	var out []Problem
	add := func(sev, code, msg, fix string) { out = append(out, Problem{sev, code, msg, fix}) }
	if len(p.Runtimes) == 0 {
		add("error", "no_runtime", "No OpenAI-compatible runtime answers on this machine",
			"Start your runtime (Ollama, LM Studio, vLLM, llama.cpp, ...) and run this again")
	}
	target := ollamaContextTarget(p.Machine.MemoryGB)
	for _, rt := range p.Runtimes {
		if !rt.Network {
			fix := "Restart it bound to 0.0.0.0"
			switch rt.Kind {
			case "ollama":
				fix = fmt.Sprintf("Set OLLAMA_HOST=0.0.0.0:%d and restart Ollama", rt.Port)
			case "lmstudio":
				fix = "LM Studio: Developer → Settings → Serve on Local Network"
			case "vllm", "sglang":
				fix = "Restart it with --host 0.0.0.0"
			case "llamacpp":
				fix = "Restart llama-server with --host 0.0.0.0"
			}
			add("error", "localhost_only", fmt.Sprintf("%s on port %d only listens on %s, so the gateway cannot reach it",
				rt.Kind, rt.Port, strings.Join(rt.Binds, ", ")), fix)
		}
		if rt.API == "openai (needs API key)" {
			add("info", "needs_api_key", fmt.Sprintf("%s on port %d requires an API key", rt.Kind, rt.Port),
				"Pass the same key with --runtime-api-key when you connect")
		}
		if rt.Kind == "ollama" {
			for _, m := range rt.OllamaLoaded {
				if m.Context != nil && *m.Context < target {
					add("warning", "small_context", fmt.Sprintf("Ollama runs %s with a %d-token window; longer prompts are cut silently", m.Name, *m.Context),
						fmt.Sprintf("Set OLLAMA_CONTEXT_LENGTH=%d and restart Ollama", target))
				}
			}
		}
		if (rt.Kind == "vllm" || rt.Kind == "sglang") && rt.Flags.PromptTokensDetails != nil && !*rt.Flags.PromptTokensDetails {
			add("info", "no_cached_token_report", fmt.Sprintf("%s on port %d does not report cached prompt tokens, so buyers pay full input price for cache hits", rt.Kind, rt.Port),
				"Add --enable-prompt-tokens-details to the vLLM command")
		}
	}
	switch {
	case !p.NetBird.Installed:
		add("error", "netbird_missing", "NetBird is not installed", "Run the connect command from the GarageAI portal")
	case !p.NetBird.Connected:
		add("error", "netbird_disconnected", "NetBird is installed but not connected", "sudo netbird up, or get a new command from the portal")
	}
	if cp := p.GarageAI.ConfiguredPort; cp != nil {
		found, kinds := false, []string{}
		for _, rt := range p.Runtimes {
			found = found || rt.Port == *cp
			kinds = append(kinds, fmt.Sprintf("%s:%d", rt.Kind, rt.Port))
		}
		if !found {
			msg := fmt.Sprintf("GarageAI is set up for port %d, but no runtime answers there", *cp)
			if len(kinds) > 0 {
				msg += " (found: " + strings.Join(kinds, ", ") + ")"
			}
			add("error", "port_mismatch", msg, fmt.Sprintf("Start the runtime on port %d, or run the connect command again with --port", *cp))
		}
	}
	switch {
	case p.GarageAI.ConfiguredRuntime != nil && !p.Heartbeat.Installed:
		add("warning", "heartbeat_missing", "The heartbeat is not installed", "Run the connect command from the portal again")
	case p.Heartbeat.Installed && !p.Heartbeat.Active:
		add("warning", "heartbeat_stopped", "The heartbeat is installed but not running", "Run the connect command from the portal again")
	}
	if strings.Contains(p.heartbeatLast, "401") || strings.Contains(strings.ToLower(p.heartbeatLast), "unauthorized") {
		add("error", "heartbeat_rejected", "GarageAI rejects this garage's heartbeat (its token was replaced or revoked)",
			"My garages → New command, and run that command here")
	}
	if p.darwin {
		hasOllama := false
		for _, rt := range p.Runtimes {
			hasOllama = hasOllama || rt.Kind == "ollama"
		}
		if hasOllama && !p.ollamaHostPersistent {
			add("warning", "ollama_host_not_persistent", "OLLAMA_HOST is not set permanently: after a restart Ollama listens on localhost again",
				"Run the connect command again and accept the offer to make it permanent")
		}
		if p.sleepMinutes > 0 {
			msg := fmt.Sprintf("This Mac sleeps %d min after its display turns off", p.sleepMinutes)
			if p.displayMinutes > 0 {
				msg += fmt.Sprintf(" (display off after %d min)", p.displayMinutes)
			}
			add("info", "mac_sleeps", msg+"; a sleeping Mac is offline for buyers",
				"System Settings → Battery/Energy → prevent automatic sleeping when the display is off")
		}
	}
	if len(p.GPUs) == 0 {
		add("info", "no_gpu", "No GPU found (nvidia-smi, rocm-smi, Apple silicon)", "Inference on CPU only is slow; buyers will see it")
	}
	if out == nil {
		out = []Problem{}
	}
	return out
}

// ListenAddrs are the local addresses a port is listening on ("*" for all), for connect's step 4.
func ListenAddrs(port int) []string {
	_, addrs, _ := byPort(listeners())
	return addrs[port]
}

// MeshIPFromInterfaces is the NetBird address on this machine's tunnel interface, or "".
func MeshIPFromInterfaces() string { return meshIPFromInterfaces() }

// Network reports whether a runtime bound to these addresses is reachable from the mesh.
func Network(binds []string) bool {
	return anyOf(binds, "0.0.0.0", "*", "::") || anyPrefix(binds, "100.")
}

// OllamaContextTarget is the window the connect script recommends for this machine's memory.
func OllamaContextTarget(mem *int) int { return ollamaContextTarget(mem) }

package profile

import (
	"fmt"
	gmachine "github.com/magnusfroste/garageai/cli/internal/machine"
	"net"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/service"
)

// KnownPorts are the default ports of the runtimes GarageAI supports; ports owned by a process
// that looks like a runtime are probed too, so an unexpected port is still found.
var KnownPorts = []int{11434, 1234, 8080, 8000, 30000, 11540, 8888, 13305, 5000, 5001, 8001}

var runtimeProc = regexp.MustCompile(`(?i)ollama|vllm|sglang|llama|lms|lm studio|lmstudio|lemonade|mlx|unsloth|paddock|xyntetik|^runner$|python|uvicorn|docker-proxy|koboldcpp|tabby|text-generation|tgi|aphrodite|exllama`)

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
	if cfg.Transport == "tunnel" {
		inst, active := service.TunnelInstalled()
		p.Tunnel = &Tunnel{IP: cfg.TunnelIP, Installed: inst, Active: active}
	}
	p.darwin = runtime.GOOS == "darwin"
	p.heartbeatLast, p.ollamaHostPersistent, p.sleepMinutes, p.displayMinutes = localFacts(hb)
	p.laptop = p.darwin && strings.Contains(run("pmset", "-g", "batt"), "InternalBattery")
	if runtime.GOOS == "windows" {
		p.winSleepACMin, p.winLidSleepsAC = windowsSleep()
	}
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
		if _, seen := addrs[kp]; !seen && loopbackOpen(kp) {
			// A listener the listing cannot show: on macOS a normal user's lsof does not see a
			// root daemon's sockets (Runner installed by garageai runtime, CI 2026-10-17). It
			// answers on loopback, so that is the bind we know of.
			ports = append(ports, kp)
			addrs[kp] = []string{"127.0.0.1"}
		}
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
			case "xyntetik":
				fix = "Xyntetik Runner serves localhost only by design: connect with the Bridge tunnel (the default), which reaches it on 127.0.0.1"
			}
			if p.GarageAI.Transport == "tunnel" {
				// Bridge forwards the tunnel to 127.0.0.1: localhost is exactly right for the runtime,
				// and anything else on localhost (Ollama's own runner, a second runtime) is out of the
				// gateway's reach anyway. Before v0.6.2 doctor flagged Ollama's runner on garage-m1/m4.
			} else {
				add("error", "localhost_only", fmt.Sprintf("%s on port %d only listens on %s, so the gateway cannot reach it",
					rt.Kind, rt.Port, strings.Join(rt.Binds, ", ")), fix)
			}
		}
		if rt.API == "openai (needs API key)" {
			fix := "Pass the same key with --runtime-api-key when you connect"
			if os.Geteuid() != 0 && p.GarageAI.ConfiguredPort != nil && rt.Port == *p.GarageAI.ConfiguredPort {
				// The key is in the root-only config; as a user, doctor cannot use it (dgx-city, 2026-10-13).
				fix = "The key is in Bridge's root-only configuration: run `sudo garageai doctor` to probe with it. The heartbeat (root) uses it already"
			}
			add("info", "needs_api_key", fmt.Sprintf("%s on port %d requires an API key", rt.Kind, rt.Port), fix)
		}
		if rt.Kind == "xyntetik" && rt.RunnerContext != nil && *rt.RunnerContext < target {
			add("warning", "small_context", fmt.Sprintf("Xyntetik Runner runs with a %d-token window; buyers' tools send longer prompts", *rt.RunnerContext),
				fmt.Sprintf("Reinstall it with a larger window: garageai runtime install --force --context %d", target))
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
	if vpns := otherVPNs(); len(vpns) > 0 && p.GarageAI.Transport != "tunnel" {
		add("info", "other_vpn", strings.Join(vpns, " and ")+" is installed: NetBird runs next to it and only routes GarageAI's "+MeshNet().String(),
			"Nothing to do. Only a device of yours with an address in "+MeshNet().String()+" would be out of reach while NetBird is on")
	}
	tunnel := p.GarageAI.Transport == "tunnel"
	switch {
	case tunnel:
		// The Bridge tunnel replaces NetBird: whether NetBird is there does not matter.
	case !p.NetBird.Installed:
		add("error", "netbird_missing", "NetBird is not installed", "Run: garageai connect (it installs NetBird)")
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
		add("warning", "heartbeat_missing", "The heartbeat is not installed", "Run: garageai connect")
	case p.Heartbeat.Installed && !p.Heartbeat.Active:
		add("warning", "heartbeat_stopped", "The heartbeat is installed but not running", "Run: garageai connect")
	}
	// How the runtime is kept alive: a terminal window or a login session is not a garage that
	// stays up. Only for the runtime the garage is configured with.
	if cp := p.GarageAI.ConfiguredPort; cp != nil {
		for _, rt := range p.Runtimes {
			if rt.Port != *cp || rt.PID == 0 {
				continue
			}
			h := howRun(rt.PID)
			switch {
			case h.Terminal != "":
				add("warning", "runtime_in_terminal", fmt.Sprintf("%s runs in a terminal window (%s): it stops when the window is closed or the machine restarts, and the garage goes offline", rt.Kind, h.Terminal),
					runAsServiceFix(rt.Kind))
			case h.UserSess && p.darwin && humanUsers() > 1:
				add("info", "runtime_user_session", fmt.Sprintf("%s runs in your login session on a Mac with several user accounts: it stops when you log out (another user logging in with fast user switching is fine)", rt.Kind),
					"Keep your account logged in, or run "+rt.Kind+" as a system service (a LaunchDaemon). Bridge will be able to set that up")
			}
		}
	}
	if p.GarageAI.Transport == "tunnel" {
		if inst, active := service.TunnelInstalled(); !inst {
			add("error", "tunnel_missing", "This garage uses the Bridge tunnel, but the tunnel service is not installed", "Run: garageai connect --tunnel bridge")
		} else if !active {
			add("error", "tunnel_stopped", "The Bridge tunnel service is installed but not running, so the gateway cannot reach this garage",
				"Linux: sudo systemctl restart garageai-tunnel. macOS: sudo launchctl kickstart -k system/eu.garageai.tunnel. Then: garageai doctor")
		}
	}
	if st := service.LoadUpdateState(); st.LastError != "" {
		add("warning", "update_failed", "Bridge could not update itself: "+st.LastError, "Run: garageai update. If it fails the same way, check the machine's internet access to github.com")
	}
	if ps, _ := gmachine.ReadPanelState(time.Now()); ps != nil && ps.State.Kind == "failed" {
		add("error", "runtime_manager_failed", ps.Panel+" reports a failed runtime: "+ps.State.Reason, "Open the runtime manager (vLLM Panel) for the cause; the garage is down for buyers until the model serves again")
	}
	for _, m := range service.ModelStates(service.LoadModelsState()) {
		if m.Status == "failed" {
			add("warning", "model_pull_failed", "GarageAI asked this garage to fetch "+m.Model+", but it failed: "+m.Error,
				"Check the disk and the internet access to ollama.com; Bridge retries after an hour. See: garageai models")
		}
	}
	if prog := service.Program(); prog != "" {
		if why := service.UnsafeProgram(prog); why != "" {
			// Not heartbeat_*: this one must reach the Operations Center with the heartbeat.
			add("error", "service_binary_unsafe", fmt.Sprintf("The heartbeat runs as root from %s, but %s: anything running as that user could get root", prog, why),
				"Run the install command again (curl -fsSL https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.sh | sh), or: sudo garageai install-service")
		}
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
		if hasOllama && !p.ollamaHostPersistent && !tunnel {
			add("warning", "ollama_host_not_persistent", "OLLAMA_HOST is not set permanently: after a restart Ollama listens on localhost again",
				"Ollama app: Settings → 'Expose Ollama to the network'. Homebrew or terminal: start it with OLLAMA_HOST=0.0.0.0:11434")
		}
		if p.sleepMinutes > 0 {
			msg := fmt.Sprintf("This Mac sleeps %d min after its display turns off", p.sleepMinutes)
			if p.displayMinutes > 0 {
				msg += fmt.Sprintf(" (display off after %d min)", p.displayMinutes)
			}
			fix := "System Settings → Energy → turn on 'Prevent automatic sleeping when the display is off'"
			if p.laptop {
				fix = "System Settings → Battery → Options… → turn on 'Prevent automatic sleeping on power adapter when the display is off', and keep it plugged in (on battery it sleeps anyway)"
			}
			add("info", "mac_sleeps", msg+"; a sleeping Mac is offline for buyers", fix)
		}
	}
	if p.winSleepACMin > 0 || p.winLidSleepsAC {
		var why []string
		if p.winSleepACMin > 0 {
			why = append(why, fmt.Sprintf("sleeps after %d min on the power adapter", p.winSleepACMin))
		}
		if p.winLidSleepsAC {
			why = append(why, "sleeps when the lid is closed on the power adapter")
		}
		add("info", "windows_sleeps", "This PC "+strings.Join(why, " and ")+"; a sleeping PC is offline for buyers",
			"In PowerShell as Administrator: powercfg /change standby-timeout-ac 0; powercfg /setacvalueindex SCHEME_CURRENT SUB_BUTTONS LIDACTION 0; powercfg /setactive SCHEME_CURRENT (on battery your own settings still apply)")
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

func runAsServiceFix(kind string) string {
	switch kind {
	case "ollama":
		return "Use the Ollama app (it keeps Ollama running), or on Linux the ollama systemd service; then close the terminal"
	case "vllm", "sglang":
		return "Run it as a service: a systemd unit, or Docker with --restart unless-stopped"
	}
	return "Run it as a service (systemd, launchd or Docker with --restart) instead of in a terminal"
}

// MeshIPFromInterfaces is the NetBird address on this machine's tunnel interface, or "".
func MeshIPFromInterfaces() string { return meshIPFromInterfaces() }

// Network reports whether a runtime bound to these addresses is reachable from the mesh.
func Network(binds []string) bool {
	return anyOf(binds, "0.0.0.0", "*", "::") || anyPrefix(binds, "100.")
}

// OllamaContextTarget is the window the connect script recommends for this machine's memory.
func OllamaContextTarget(mem *int) int { return ollamaContextTarget(mem) }

// loopbackOpen: something accepts TCP connections on 127.0.0.1:port.
func loopbackOpen(port int) bool {
	c, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), 300*time.Millisecond)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

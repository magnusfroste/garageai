package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/machine"
	"github.com/magnusfroste/garageai/cli/internal/portal"
	"github.com/magnusfroste/garageai/cli/internal/profile"
)

// link connects this machine to the operator's GarageAI account in the browser: Bridge shows a
// code, the operator approves the machine at app.garageai.eu/bridge/link, and Bridge picks up the
// garage's credentials itself. No command with a token is copied, pasted or shown. It runs as the
// operator (their browser, their session) before connect re-runs itself as root.
//
// It fills o (and the GARAGEAI_* environment the root half reads) with the garage's name, the
// register URL and token, the setup key and the management URL, and the runtime it found when
// none was given.
func link(o *Options, runtimeGiven bool) int {
	facts := profile.Build(version)
	if !runtimeGiven {
		if kind, port, ok := pickRuntime(facts); ok {
			o.Runtime, o.Port = kind, port
		} else {
			noRuntime("")
			return 1
		}
	} else if !runtimeCheck(facts, o.Port) {
		noRuntime(fmt.Sprintf(" on port %d", o.Port))
		return 1
	}
	host := computerName(o.Name)
	req := portal.LinkRequest{MachineID: machine.ID(), Hostname: host, OS: facts.Machine.OS, Arch: facts.Machine.Arch,
		BridgeVersion: version, Runtime: o.Runtime, SuggestedName: suggestName(host)}
	if facts.Machine.MemoryGB != nil {
		req.MemoryGB = *facts.Machine.MemoryGB
	}
	if len(facts.GPUs) > 0 {
		req.GPU = facts.GPUs[0].Name
	}
	for _, r := range facts.Runtimes {
		if r.Kind == o.Runtime && r.Port == o.Port {
			for _, m := range r.Models {
				req.Models = append(req.Models, m.ID)
			}
		}
	}

	fmt.Println(sgr("1", "GarageAI Bridge "+version+" — link this machine to your GarageAI account"))
	fmt.Println()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	start, err := portal.StartLink(ctx, req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "  Could not reach GarageAI:", err)
		fmt.Fprintln(os.Stderr, "  Check the internet connection and run garageai connect again.")
		return 1
	}
	fmt.Printf("  1. Open   %s\n", start.VerificationURLComplete)
	fmt.Printf("     (or open %s and enter the code %s)\n", start.VerificationURL, sgr("1", start.UserCode))
	fmt.Println("  2. Check that it is this machine, choose the garage, and approve.")
	fmt.Println("     On a server without a browser, open the link on your laptop or phone.")
	fmt.Println()
	if openBrowser(start.VerificationURLComplete) {
		fmt.Println("  (Opened in your browser. Signed out? Sign in, and the code is still filled in.)")
	}
	fmt.Println()
	if !colorOn {
		fmt.Println("  Waiting for your approval in the browser (10 minutes)...")
	}
	res, err := portal.WaitForLink(ctx, start, func(left time.Duration) {
		if colorOn {
			fmt.Printf("\r  Waiting for your approval in the browser... %d:%02d left ", int(left.Minutes()), int(left.Seconds())%60)
		}
	})
	if colorOn {
		fmt.Print("\r\033[K")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "  Stopped:", err)
		return 130
	}
	switch res.Status {
	case "approved":
	case "denied":
		fmt.Println("  Denied in the browser. Nothing was changed on this machine.")
		return 1
	case "expired":
		fmt.Println("  The code expired (10 minutes). Run garageai connect again for a new one.")
		return 1
	default:
		fmt.Printf("  The link ended (%s). Run garageai connect again.\n", res.Status)
		return 1
	}
	fmt.Printf("  %s Approved: this machine is the garage %s\n\n", sgr("32", "✓"), sgr("1", res.Garage.Name))

	o.Name, o.RegisterURL, o.RegisterToken = res.Garage.Name, res.RegisterURL, res.RegisterToken
	o.SetupKey, o.ManagementURL = res.SetupKey, res.ManagementURL
	for k, v := range map[string]string{
		"GARAGEAI_NODE_NAME": o.Name, "GARAGEAI_REGISTER_URL": o.RegisterURL, "GARAGEAI_REGISTER_TOKEN": o.RegisterToken,
		"GARAGEAI_SETUP_KEY": o.SetupKey, "GARAGEAI_MANAGEMENT_URL": o.ManagementURL,
		"GARAGEAI_RUNTIME": o.Runtime, "GARAGEAI_PORT": strconv.Itoa(o.Port),
	} {
		os.Setenv(k, v)
	}
	return 0
}

// noRuntime stops before anything is sent to GarageAI: approving a machine that has nothing to sell
// only costs the operator a round trip.
func noRuntime(where string) {
	fmt.Println(sgr("1", "GarageAI Bridge "+version))
	fmt.Println()
	fmt.Printf("  %s No inference runtime with a model answers on this machine%s.\n", sgr("31", "✗"), where)
	fmt.Println("    Start one with a model loaded, then run garageai connect again. For example Ollama:")
	fmt.Println("      install it from https://ollama.com/download, then: ollama pull qwen3:4b")
	fmt.Println("    LM Studio, vLLM, llama.cpp and others work too; garageai doctor shows what it finds.")
}

var runtimeCheck = hasRuntimeOn // replaced in tests, which have no runtime

func hasRuntimeOn(p profile.Profile, port int) bool {
	for _, r := range p.Runtimes {
		if r.Port == port {
			return true
		}
	}
	return false
}

var connectKinds = map[string]bool{"ollama": true, "lmstudio": true, "llamacpp": true, "vllm": true, "sglang": true,
	"paddock": true, "unsloth": true, "mlx": true, "lemonade": true}

// pickRuntime chooses the runtime to connect when the operator named none: one with models, and
// of those the one the mesh can already reach.
func pickRuntime(p profile.Profile) (string, int, bool) {
	best, score := -1, -1
	for i, r := range p.Runtimes {
		s := 0
		if len(r.Models) > 0 {
			s += 2
		}
		if r.Network {
			s++
		}
		if s > score {
			best, score = i, s
		}
	}
	if best < 0 {
		return "", 0, false
	}
	r := p.Runtimes[best]
	kind := r.Kind
	if !connectKinds[kind] {
		kind = "other"
	}
	return kind, r.Port, true
}

var notName = regexp.MustCompile(`[^a-z0-9-]+`)

// suggestName turns a hostname ("Magnus's MacBook Air") into a garage name ("magnuss-macbook-air").
func suggestName(host string) string {
	n := strings.NewReplacer(" ", "-", "’", "", "'", "").Replace(strings.ToLower(host))
	n = strings.Trim(regexp.MustCompile(`-+`).ReplaceAllString(notName.ReplaceAllString(n, ""), "-"), "-")
	if len(n) > 40 {
		n = strings.Trim(n[:40], "-")
	}
	if len(n) < 2 {
		return ""
	}
	return n
}

// computerName is the name the operator knows the machine by: on a Mac "Magnus's MacBook Air"
// rather than the hostname "Mac".
func computerName(host string) string {
	if runtime.GOOS == "darwin" {
		if out, err := exec.Command("scutil", "--get", "ComputerName").Output(); err == nil && strings.TrimSpace(string(out)) != "" {
			return strings.TrimSpace(string(out))
		}
	}
	return host
}

func openBrowser(url string) bool {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
			return false // a server over ssh: the operator opens the link on another device
		}
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start() == nil
}

// runtimeGiven: the operator chose the runtime or port (flag or environment); link then keeps it.
func runtimeGiven(args []string) bool {
	if os.Getenv("GARAGEAI_RUNTIME") != "" || os.Getenv("GARAGEAI_PORT") != "" {
		return true
	}
	for _, a := range args {
		if a == "--runtime" || a == "--port" {
			return true
		}
	}
	return false
}

package main

import (
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/profile"
	"github.com/magnusfroste/garageai/cli/internal/runnerfit"
	"github.com/magnusfroste/garageai/cli/internal/service"
)

// garageai runtime: for an operator with no runtime. It looks at the machine, says whether
// Xyntetik Runner is a good fit and with which model, and installs it as a service of Bridge's.
// The decision is runnerfit.Assess; the service is service.InstallRunner.
func runtimeCmd(args []string) int {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	switch sub {
	case "", "check":
		return runtimeStatus()
	case "install":
		return runtimeInstall(args[1:])
	case "remove", "uninstall":
		if !isAdmin() {
			return asRootSimple("runtime", args)
		}
		service.RemoveRunner()
		fmt.Println("✓ Runner's service is removed. Its binary and downloaded models stay in", filepath.Dir(service.RunnerPath()), "and", service.RunnerHome())
		return 0
	}
	fmt.Fprintln(os.Stderr, "usage: garageai runtime [check] | install [--model owner/repo[:TAG]] [--port N] [--context TOKENS] [--force] [--yes] | remove")
	return 2
}

func assessThisMachine() (runnerfit.Assessment, profile.Profile) {
	p := profile.Build(version)
	m := runnerfit.Machine{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
	if p.Machine.MemoryGB != nil {
		m.MemoryGB = *p.Machine.MemoryGB
	}
	for _, g := range p.GPUs {
		mb := 0
		if g.MemoryMB != nil {
			mb = *g.MemoryMB
		}
		m.GPUs = append(m.GPUs, runnerfit.GPU{Vendor: g.Vendor, Name: g.Name, MemoryMB: mb, Driver: g.Driver, UnifiedMemory: g.UnifiedMemory})
	}
	ours, _ := service.RunnerInstalled()
	for _, r := range p.Runtimes {
		if len(r.Models) == 0 || (ours && r.Kind == "xyntetik") {
			continue
		}
		m.Runtimes = append(m.Runtimes, r.Kind)
	}
	return runnerfit.Assess(m), p
}

func printAssessment(a runnerfit.Assessment) {
	switch a.Verdict {
	case runnerfit.Fit:
		fmt.Printf("  ✓ Xyntetik Runner fits this machine (%s).\n", a.Backend)
	case runnerfit.CPUOnly:
		fmt.Println("  · Xyntetik Runner works here, on the processor: fine for a small model, slow for a large one.")
	case runnerfit.KeepCurrent:
		fmt.Println("  · This machine serves models already.")
	case runnerfit.Unsupported:
		fmt.Println("  ✗ Xyntetik Runner does not run on this machine yet.")
	}
	for _, r := range a.Reasons {
		fmt.Println("      " + r)
	}
	if a.Model.HF != "" && a.Verdict != runnerfit.Unsupported {
		fmt.Printf("    Model: %s, %.1f GB download: %s.\n", a.Model.Name, a.Model.SizeGB, a.Model.Why)
	}
}

func runtimeStatus() int {
	fmt.Println("GarageAI Bridge " + version + " — a runtime for this machine")
	fmt.Println()
	a, _ := assessThisMachine()
	if inst, model := service.RunnerInstalled(); inst {
		state := "starting or downloading the model"
		if runnerAnswers(runnerfit.Port) {
			state = "serving on 127.0.0.1:" + strconv.Itoa(runnerfit.Port)
		}
		fmt.Printf("  Runner   installed by Bridge, %s (%s)\n\n", state, model)
	}
	printAssessment(a)
	fmt.Println()
	switch a.Verdict {
	case runnerfit.Fit, runnerfit.CPUOnly:
		fmt.Println("  Install it:  garageai runtime install")
	case runnerfit.KeepCurrent:
		fmt.Println("  Nothing to install. Connect with: garageai connect")
	}
	return 0
}

func runtimeInstall(args []string) int {
	model, port, force, yes, context := "", runnerfit.Port, false, false, 0
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--model":
			if i+1 < len(args) {
				i++
				model = args[i]
			}
		case "--port":
			if i+1 < len(args) {
				i++
				port, _ = strconv.Atoi(args[i])
			}
		case "--context":
			if i+1 < len(args) {
				i++
				context, _ = strconv.Atoi(args[i])
			}
		case "--force":
			force = true
		case "--yes", "-y":
			yes = true
		default:
			fmt.Fprintln(os.Stderr, "garageai runtime install: unknown option", args[i])
			return 2
		}
	}
	if port <= 0 || port > 65535 {
		fmt.Fprintln(os.Stderr, "garageai runtime install: --port must be 1-65535")
		return 2
	}
	if model != "" && !validHF(model) {
		fmt.Fprintln(os.Stderr, "garageai runtime install: --model must be owner/repo or owner/repo:TAG (a Hugging Face GGUF repository)")
		return 2
	}

	fmt.Println("GarageAI Bridge " + version + " — install Xyntetik Runner " + runnerfit.Version)
	fmt.Println()
	a, p := assessThisMachine()
	printAssessment(a)
	fmt.Println()
	switch {
	case a.Verdict == runnerfit.Unsupported:
		return 1
	case a.Verdict == runnerfit.KeepCurrent && !force:
		fmt.Println("  Not installing: keep the runtime you have. (--force installs Runner next to it.)")
		return 0
	}
	if model == "" {
		model = a.Model.HF
	}
	if context <= 0 {
		context = profile.ContextTarget(p.Machine.MemoryGB)
	}
	if inst, _ := service.RunnerInstalled(); !inst && portBusy(port) {
		fmt.Fprintf(os.Stderr, "  ✗ Something else listens on port %d already. Choose another: garageai runtime install --port 8090\n", port)
		return 1
	}
	if !yes {
		fmt.Printf("  Install Runner %s as a service and serve %s on 127.0.0.1:%d with a %d-token window? [Y/n] ", runnerfit.Version, model, port, context)
		var ans string
		fmt.Scanln(&ans)
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(ans)), "n") {
			fmt.Println("  Nothing installed.")
			return 0
		}
		args = append(args, "--yes")
	}
	if !isAdmin() {
		return asRootSimple("runtime", append([]string{"install"}, args...))
	}

	bin := service.RunnerPath()
	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "  ✗", err)
		return 1
	}
	fmt.Printf("  Downloading Runner %s (%s), checked against its SHA256SUMS...\n", runnerfit.Version, a.Asset)
	base := "https://github.com/Joakimpalm-Zen/xyntetik-runner/releases/download/" + runnerfit.Version
	tmp := bin + ".new"
	if err := downloadAsset(base, a.Asset, tmp, 512<<20); err != nil {
		os.Remove(tmp)
		fmt.Fprintln(os.Stderr, "  ✗ Could not download Runner:", err)
		return 1
	}
	service.RemoveRunner() // a running Runner holds its file on Windows
	if err := os.Rename(tmp, bin); err != nil {
		os.Remove(tmp)
		fmt.Fprintln(os.Stderr, "  ✗ Could not install Runner:", err)
		return 1
	}
	service.RecordRunnerVersion(runnerfit.Version)
	fmt.Println("  ✓ Runner installed:", bin)
	if err := service.InstallRunner(bin, model, port, context); err != nil {
		fmt.Fprintln(os.Stderr, "  ✗ Could not start Runner as a service:", err)
		return 1
	}
	fmt.Printf("  ✓ Runner runs as a service and fetches %s (checked against the Hub's SHA-256).\n", model)
	fmt.Println("    The first start downloads the model; this can take a while. Waiting for it to answer...")

	deadline := time.Now().Add(60 * time.Minute)
	last := time.Time{}
	for !runnerAnswers(port) {
		if time.Now().After(deadline) {
			fmt.Println("  · Still not answering after an hour. It keeps going in the background: check with garageai runtime")
			return 0
		}
		if time.Since(last) > 30*time.Second {
			fmt.Printf("    ... %.1f GB downloaded\n", float64(dirSize(service.RunnerHome()))/(1<<30))
			last = time.Now()
		}
		time.Sleep(5 * time.Second)
	}
	fmt.Printf("  ✓ Runner answers on 127.0.0.1:%d.\n\n", port)
	if _, err := service.Load(); err == nil {
		fmt.Printf("  Next: offer it on GarageAI (keeps your link):  garageai connect --runtime xyntetik --port %d\n", port)
	} else {
		fmt.Println("  Next: connect this machine to GarageAI:  garageai connect")
	}
	return 0
}

func validHF(s string) bool {
	repo, _, _ := strings.Cut(s, ":")
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	for _, r := range s {
		if !(r == '-' || r == '_' || r == '.' || r == '/' || r == ':' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return false
		}
	}
	return true
}

func portBusy(port int) bool {
	c, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

func runnerAnswers(port int) bool {
	cl := &http.Client{Timeout: 3 * time.Second}
	resp, err := cl.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/v1/models")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func dirSize(dir string) int64 {
	var n int64
	_ = filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if fi, err := d.Info(); err == nil {
				n += fi.Size()
			}
		}
		return nil
	})
	return n
}

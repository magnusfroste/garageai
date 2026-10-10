package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/netbird"
	"github.com/magnusfroste/garageai/cli/internal/portal"
	"github.com/magnusfroste/garageai/cli/internal/profile"
	rt "github.com/magnusfroste/garageai/cli/internal/runtime"
	"github.com/magnusfroste/garageai/cli/internal/service"
)

// Options mirror garageai-connect.sh: every flag can also be given as GARAGEAI_* in the
// environment, so the portal's command only swaps the script for `garageai connect`.
type Options struct {
	SetupKey      string
	ManagementURL string
	Runtime       string
	Port          int
	Name          string
	RuntimeAPIKey string
	Models        string
	RegisterURL   string
	RegisterToken string
	SkipInstall   bool
	NoHeartbeat   bool
	Yes           bool
	MeshWait      time.Duration
}

func optionsFromEnv() Options {
	host, _ := os.Hostname()
	if i := strings.Index(host, "."); i > 0 {
		host = host[:i]
	}
	o := Options{
		SetupKey: os.Getenv("GARAGEAI_SETUP_KEY"), ManagementURL: os.Getenv("GARAGEAI_MANAGEMENT_URL"),
		Runtime: env("GARAGEAI_RUNTIME", "ollama"), Name: env("GARAGEAI_NODE_NAME", host),
		RuntimeAPIKey: os.Getenv("GARAGEAI_RUNTIME_API_KEY"), Models: os.Getenv("GARAGEAI_MODELS"),
		RegisterURL: os.Getenv("GARAGEAI_REGISTER_URL"), RegisterToken: os.Getenv("GARAGEAI_REGISTER_TOKEN"),
		MeshWait: 90 * time.Second,
	}
	o.Port, _ = strconv.Atoi(os.Getenv("GARAGEAI_PORT"))
	if s, err := strconv.Atoi(os.Getenv("GARAGEAI_MESH_WAIT_SECONDS")); err == nil && s > 0 {
		o.MeshWait = time.Duration(s) * time.Second
	}
	return o
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func parseConnectArgs(args []string, o *Options) error {
	for i := 0; i < len(args); i++ {
		a := args[i]
		next := func() (string, error) {
			if i+1 >= len(args) {
				return "", fmt.Errorf("%s needs a value", a)
			}
			i++
			return args[i], nil
		}
		var v string
		var err error
		switch a {
		case "--setup-key":
			o.SetupKey, err = next()
		case "--management-url":
			o.ManagementURL, err = next()
		case "--runtime":
			o.Runtime, err = next()
		case "--port":
			v, err = next()
			if err == nil {
				o.Port, err = strconv.Atoi(v)
			}
		case "--name":
			o.Name, err = next()
		case "--runtime-api-key":
			o.RuntimeAPIKey, err = next()
		case "--models":
			o.Models, err = next()
		case "--register-url":
			o.RegisterURL, err = next()
		case "--register-token":
			o.RegisterToken, err = next()
		case "--skip-install":
			o.SkipInstall = true
		case "--no-heartbeat":
			o.NoHeartbeat = true
		case "--yes", "-y":
			o.Yes = true
		default:
			return fmt.Errorf("unknown option %s", a)
		}
		if err != nil {
			return err
		}
	}
	if o.Port == 0 {
		o.Port = rt.DefaultPort(o.Runtime)
	}
	if o.Port == 0 {
		return fmt.Errorf("--port is required for runtime %q", o.Runtime)
	}
	if strings.ContainsAny(o.RuntimeAPIKey, " \t\n") || strings.Contains(o.RuntimeAPIKey, "GARAGEAI_") {
		return errors.New("the runtime API key contains spaces or 'GARAGEAI_': something else was pasted into it")
	}
	if strings.HasPrefix(o.RuntimeAPIKey, "<") || strings.HasPrefix(o.RuntimeAPIKey, "YOUR_") {
		return errors.New("replace the placeholder in --runtime-api-key with your runtime's real API key, or leave it out")
	}
	if o.RegisterURL != "" && o.RegisterToken == "" {
		return errors.New("--register-url given without --register-token (or GARAGEAI_REGISTER_TOKEN)")
	}
	return nil
}

// ---------------------------------------------------------------- output and reporting

type connectRun struct {
	o       Options
	portal  *portal.Client
	step    string
	facts   profile.Profile
	profile json.RawMessage
	done    bool
}

func (c *connectRun) bold(s string) { fmt.Printf("\033[1m%s\033[0m\n", s) }
func (c *connectRun) ok(s string)   { fmt.Printf("  \033[32m✓\033[0m %s\n", s) }
func (c *connectRun) info(s string) { fmt.Printf("  %s\n", s) }
func (c *connectRun) warn(s string) {
	fmt.Fprintf(os.Stderr, "  \033[33m!\033[0m %s\n", s)
	c.report("warning", s, nil)
}

func (c *connectRun) report(status, msg string, prof json.RawMessage) {
	if c.portal == nil {
		return
	}
	var gpu string
	for i, g := range c.facts.GPUs {
		if i > 0 {
			gpu += ";"
		}
		gpu += g.Name
		if g.MemoryMB != nil {
			gpu += fmt.Sprintf(" (%d MB)", *g.MemoryMB)
		}
	}
	mem := ""
	if c.facts.Machine.MemoryGB != nil {
		mem = strconv.Itoa(*c.facts.Machine.MemoryGB)
	}
	c.portal.Report(portal.Report{Step: c.step, Status: status, Message: msg, ScriptVersion: version + " (bridge)",
		NodeName: c.o.Name, Runtime: c.o.Runtime, Port: strconv.Itoa(c.o.Port), OS: c.facts.Machine.OS,
		Arch: c.facts.Machine.Arch, GPU: gpu, MemoryGB: mem, Profile: prof})
}

func (c *connectRun) begin(step string) {
	c.step = step
	c.bold(step)
	if c.profile == nil {
		c.profile, _ = json.Marshal(c.facts)
		c.report("started", "", c.profile)
		return
	}
	c.report("started", "", nil)
}

// fail prints the error, reports it with the profile, and exits 1.
func (c *connectRun) fail(msg string) {
	fmt.Fprintf(os.Stderr, "  \033[31m✗\033[0m %s\n", msg)
	c.report("failed", msg, c.profile)
	os.Exit(1)
}

func (c *connectRun) confirm(q string) bool {
	if c.o.Yes {
		return true
	}
	fmt.Printf("  %s [y/N] ", q)
	var a string
	fmt.Scanln(&a)
	return strings.EqualFold(a, "y") || strings.EqualFold(a, "yes")
}

// ---------------------------------------------------------------- root

// asRoot re-runs this binary through sudo when we are not root. The secrets travel on stdin, never
// on the command line, so they do not show up in the process list or sudo's log.
func asRoot(args []string) int {
	self, _ := os.Executable()
	cmd := exec.Command("sudo", "-p", "  [sudo] password for %p (joining the mesh and installing the heartbeat need it): ",
		self, "__connect-as-root")
	secrets, _ := json.Marshal(map[string]string{
		"GARAGEAI_SETUP_KEY": os.Getenv("GARAGEAI_SETUP_KEY"), "GARAGEAI_REGISTER_TOKEN": os.Getenv("GARAGEAI_REGISTER_TOKEN"),
		"GARAGEAI_RUNTIME_API_KEY": os.Getenv("GARAGEAI_RUNTIME_API_KEY"), "GARAGEAI_MODELS": os.Getenv("GARAGEAI_MODELS"),
		"GARAGEAI_NODE_NAME": os.Getenv("GARAGEAI_NODE_NAME"), "GARAGEAI_RUNTIME": os.Getenv("GARAGEAI_RUNTIME"),
		"GARAGEAI_PORT": os.Getenv("GARAGEAI_PORT"), "GARAGEAI_REGISTER_URL": os.Getenv("GARAGEAI_REGISTER_URL"),
		"GARAGEAI_MANAGEMENT_URL": os.Getenv("GARAGEAI_MANAGEMENT_URL"), "GARAGEAI_REPORT": os.Getenv("GARAGEAI_REPORT"),
		"GARAGEAI_MESH_WAIT_SECONDS": os.Getenv("GARAGEAI_MESH_WAIT_SECONDS"), "GARAGEAI_USER_HOME": os.Getenv("HOME"),
		"ARGS": strings.Join(args, "\x00"),
	})
	cmd.Stdin = strings.NewReader(string(secrets) + "\n")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "could not run sudo:", err)
		return 1
	}
	return 0
}

// connectAsRoot is the root half: it reads the secrets from stdin and runs connect.
func connectAsRoot() int {
	var secrets map[string]string
	line, _ := io.ReadAll(io.LimitReader(os.Stdin, 1<<20))
	if json.Unmarshal(line, &secrets) != nil {
		fmt.Fprintln(os.Stderr, "no options on stdin")
		return 1
	}
	args := strings.Split(secrets["ARGS"], "\x00")
	if secrets["ARGS"] == "" {
		args = nil
	}
	for k, v := range secrets {
		if strings.HasPrefix(k, "GARAGEAI_") && v != "" {
			os.Setenv(k, v)
		}
	}
	return connect(args)
}

// ---------------------------------------------------------------- the six steps

func connect(args []string) int {
	o := optionsFromEnv()
	if err := parseConnectArgs(args, &o); err != nil {
		fmt.Fprintln(os.Stderr, "garageai connect:", err)
		return 2
	}
	if os.Geteuid() != 0 && runtime.GOOS != "windows" {
		return asRoot(args)
	}
	c := &connectRun{o: o}
	if o.RegisterURL != "" && o.RegisterToken != "" {
		c.portal = portal.New(o.RegisterURL, o.RegisterToken)
	}
	c.facts = profile.Build(version)
	c.step = "0/6  Start"
	c.profile, _ = json.Marshal(c.facts)
	c.report("started", "Bridge started", c.profile)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() { <-stop; c.report("stopped", "Interrupted (Ctrl-C)", nil); os.Exit(130) }()
	defer func() {
		if !c.done {
			c.report("stopped", "Waiting for the operator: fix what Bridge said, then run the command again", nil)
		}
	}()

	c.bold("GarageAI Bridge " + version + " — connect " + o.Name)
	fmt.Println()

	// 1. NetBird client
	c.begin("1/6  NetBird client")
	if netbird.Installed() {
		c.ok("netbird is installed (" + netbird.Version() + ")")
	} else if o.SkipInstall {
		c.fail("netbird is not installed and --skip-install was given.")
	} else {
		c.info("NetBird is not installed. It will be installed from https://pkgs.netbird.io/install.sh")
		if !c.confirm("Install NetBird now?") {
			c.fail("Aborted. Install NetBird yourself (https://netbird.io) and run the command again.")
		}
		if err := netbird.Install(); err != nil {
			c.fail("NetBird installation failed: " + err.Error())
		}
		c.ok("netbird installed")
	}
	fmt.Println()

	// 2. Join the mesh
	c.begin("2/6  Join the GarageAI mesh")
	meshIP := netbird.MeshIP()
	switch {
	case meshIP != "" && o.SetupKey == "":
		c.ok("Already on the mesh (no setup key given, keeping the current connection)")
	case o.SetupKey == "":
		c.fail("This machine is not on the mesh and no setup key was given. Get a new command from the portal: My garages → New command (it includes a setup key when the garage is not connected).")
	case o.ManagementURL == "":
		c.fail("No management URL. Pass --management-url (you get it from GarageAI).")
	default:
		if err := netbird.Up(o.SetupKey, o.ManagementURL, o.Name, exec.Command); err != nil {
			c.fail("netbird up failed: " + err.Error() + ". A setup key can be used once: get a new command from the portal if this one was used.")
		}
		meshIP = netbird.WaitForMeshIP(o.MeshWait)
	}
	if meshIP == "" {
		c.fail("Could not get a mesh IP. Check 'netbird status' and that the setup key is valid.")
	}
	c.ok("Mesh IP: " + meshIP)
	fmt.Println()

	// 3. Local runtime
	c.begin(fmt.Sprintf("3/6  Inference runtime (%s, port %d)", o.Runtime, o.Port))
	var served []string
	var contexts map[string]int
	probeHost := ""
	for _, host := range []string{"127.0.0.1", meshIP} {
		ids, ctxs, err := rt.New(host, o.Port, o.RuntimeAPIKey).Models()
		if err == nil {
			served, contexts, probeHost = ids, ctxs, host
			break
		}
		if errors.Is(err, rt.ErrAuth) {
			if o.RuntimeAPIKey != "" {
				c.fail(fmt.Sprintf("%s answers on port %d but rejects the API key. Check GARAGEAI_RUNTIME_API_KEY: it must be exactly the key the runtime was started with.", o.Runtime, o.Port))
			}
			c.fail(fmt.Sprintf("%s answers on port %d but requires an API key. Set it: export GARAGEAI_RUNTIME_API_KEY='<key>' and run again.", o.Runtime, o.Port))
		}
	}
	if probeHost == "" {
		c.hintRuntime()
		c.fail(fmt.Sprintf("No OpenAI-compatible API answers on port %d. Start %s and run the command again.", o.Port, o.Runtime))
	}
	c.ok(fmt.Sprintf("OpenAI-compatible API answers on %s:%d", probeHost, o.Port))
	if len(served) == 0 {
		c.fail("The runtime answered but lists no models. Load or pull a model first.")
	}
	fmt.Println()

	// 4. Reachable over the mesh
	c.begin("4/6  Reachable over the mesh")
	binds := profile.ListenAddrs(o.Port)
	switch {
	case probeHost == meshIP:
		c.ok(fmt.Sprintf("Reachable on %s:%d", meshIP, o.Port))
	case len(binds) == 0:
		c.info("(could not read which address the runtime listens on; the gateway verifies the mesh path next)")
	case profile.Network(binds) || contains(binds, meshIP):
		c.ok(fmt.Sprintf("Listening on %s (port %d); the gateway verifies the mesh path next", strings.Join(binds, " "), o.Port))
	default:
		c.warn(fmt.Sprintf("The runtime only listens on %s, so the gateway cannot reach it.", strings.Join(binds, " ")))
		c.hintBind()
		c.fail("Restart the runtime bound to 0.0.0.0 and run the command again.")
	}
	if hint := ufwBlocks(o.Port); hint != "" {
		c.warn(fmt.Sprintf("ufw is active and has no rule for port %d, so the gateway may be blocked.", o.Port))
		c.info("  → " + hint)
	}
	offered, skipped := rt.Offer(served, o.Models)
	for _, m := range offered {
		c.info("model: " + m)
	}
	for _, m := range skipped {
		if o.Models == "" {
			c.info("model: " + m + " (skipped: embedding/reranker model)")
		} else {
			c.info("model: " + m + " (not offered)")
		}
	}
	if len(offered) == 0 {
		c.fail(fmt.Sprintf("No chat model to offer. Load one in %s (or check --models) and run the command again.", o.Runtime))
	}
	client := rt.New(probeHost, o.Port, o.RuntimeAPIKey)
	if client.UsageReported(context.Background(), offered[0]) {
		c.ok("Token usage is reported (needed for per-token billing)")
	} else {
		c.warn(fmt.Sprintf("No token usage in the streamed reply from %s. Billing may be incomplete; check that the runtime supports stream_options.include_usage.", offered[0]))
	}
	if o.Runtime == "ollama" {
		// The request above loaded the model, so /api/ps now shows the window it runs with.
		target := profile.OllamaContextTarget(c.facts.Machine.MemoryGB)
		if ctx := client.OllamaLoadedContext(offered[0]); ctx > 0 {
			contexts[offered[0]] = ctx
			if ctx >= target {
				c.ok(fmt.Sprintf("Context window: %d tokens", ctx))
			} else {
				c.warn(fmt.Sprintf("Ollama runs %s with a %d-token context window and silently cuts longer prompts; buyers' coding agents send far longer prompts. Recommended for this machine: %d tokens.", offered[0], ctx, target))
				c.hintOllamaContext(target)
				c.info("  Continuing: the garage is registered with the window it actually runs with, so the gateway rejects longer prompts instead.")
			}
		} else {
			c.info(fmt.Sprintf("Could not read Ollama's context window; make sure OLLAMA_CONTEXT_LENGTH is at least %d.", target))
		}
	}
	fmt.Println()

	// 5. Register
	c.begin("5/6  Register with GarageAI")
	reg := portal.Registration{Name: o.Name, MeshIP: meshIP, Port: o.Port, Runtime: o.Runtime, Models: offered,
		Contexts: onlyOffered(contexts, offered), RuntimeAPIKey: o.RuntimeAPIKey}
	if c.portal == nil {
		c.done = true
		c.info("No --register-url given. Send these details to GarageAI to activate the node:")
		shown := reg
		shown.RuntimeAPIKey = ""
		b, _ := json.MarshalIndent(shown, "  ", "  ")
		fmt.Println("  " + string(b))
		return 0
	}
	c.info("Registering and running the acceptance test (a real request through the gateway)...")
	res, err := c.portal.Register(context.Background(), reg)
	if err != nil {
		c.fail("Registration failed: " + err.Error())
	}
	passed := 0
	for _, a := range res.Acceptance {
		if a.Passed {
			passed++
			c.ok(fmt.Sprintf("%s: passed (%s tok/s, first token after %s ms)", a.Model, num(a.TokensPerSecond), num(a.TTFTms)))
		} else {
			c.warn(fmt.Sprintf("%s: failed (%s)", a.Model, a.Error))
		}
	}
	if passed > 0 {
		c.ok("Node registered — your garage is live on GarageAI.")
		c.report("done", fmt.Sprintf("Registered: %d model(s) passed the acceptance test", passed), c.profile)
	} else {
		c.warn("Registered, but no model passed the acceptance test, so nothing is for sale yet. Check that the runtime answers on the mesh IP and that the model loads, then run this again.")
	}
	c.done = true

	// 6. Heartbeat
	if o.NoHeartbeat {
		return 0
	}
	c.begin("6/6  Heartbeat")
	self := installedBinary()
	cfg := &service.Config{HeartbeatURL: strings.TrimSuffix(o.RegisterURL, "/register-node") + "/node-heartbeat",
		RegisterURL: o.RegisterURL, RegisterToken: o.RegisterToken, Name: o.Name, Runtime: o.Runtime, Port: o.Port,
		RuntimeAPIKey: o.RuntimeAPIKey, MeshIP: meshIP}
	if err := service.Save(cfg); err != nil {
		c.fail("Could not write " + service.ConfigPath() + ": " + err.Error())
	}
	if removed := service.RemoveLegacyScriptHeartbeat(); removed {
		c.info("Removed the connect script's old heartbeat; Bridge reports from now on.")
	}
	if err := service.Install(self); err != nil {
		c.fail("Could not install the heartbeat service: " + err.Error())
	}
	c.ok("Installed: Bridge reports your models every 5 minutes. Load a new model and it shows up")
	c.info("  under My garages in the portal, where you choose to offer it. Remove with: garageai uninstall")
	c.report("done", "Heartbeat installed", nil)
	c.notes()
	return 0
}

// notes prints what doctor would flag that is not an error: things to know for a garage that
// should stay up (a sleeping Mac, OLLAMA_HOST not persistent, a small window).
func (c *connectRun) notes() {
	p := profile.Build(version)
	first := true
	for _, pr := range p.Problems {
		if pr.Severity == "error" {
			continue
		}
		if first {
			fmt.Println()
			c.bold("Good to know")
			first = false
		}
		mark := "·"
		if pr.Severity == "warning" {
			mark = "!"
		}
		fmt.Printf("  %s %s\n      → %s\n", mark, pr.Message, pr.Fix)
	}
}

// installedBinary makes sure the service runs a root-owned copy in /usr/local/bin: the installer
// may have put us in the operator's ~/.local/bin, and a root service should not execute a file
// the user can rewrite. Falls back to our own path when the copy is not possible.
func installedBinary() string {
	self, err := os.Executable()
	if err != nil {
		return "garageai"
	}
	if runtime.GOOS == "windows" || strings.HasPrefix(self, "/usr/local/bin/") {
		return self
	}
	const dest = "/usr/local/bin/garageai"
	b, err := os.ReadFile(self)
	if err != nil {
		return self
	}
	if err := os.MkdirAll("/usr/local/bin", 0o755); err != nil {
		return self
	}
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, b, 0o755); err != nil {
		return self
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return self
	}
	return dest
}

// ufwBlocks: on Linux with ufw active and no allow rule for the port, the command to allow the
// mesh in; "" otherwise. Mesh traffic arrives on NetBird's wt0 interface.
func ufwBlocks(port int) string {
	if runtime.GOOS != "linux" {
		return ""
	}
	out, err := exec.Command("ufw", "status").Output()
	if err != nil || !strings.Contains(string(out), "Status: active") {
		return ""
	}
	if strings.Contains(string(out), strconv.Itoa(port)) {
		return ""
	}
	return fmt.Sprintf("sudo ufw allow in on wt0 to any port %d proto tcp comment garageai", port)
}

func (c *connectRun) hintRuntime() {
	switch c.o.Runtime {
	case "ollama":
		c.info("  Ollama: OLLAMA_HOST=0.0.0.0:11434 ollama serve   (or the Ollama app, with the network setting on)")
	case "lmstudio":
		c.info("  LM Studio: Developer → Start server, and enable 'Serve on local network'")
	case "vllm":
		c.info(fmt.Sprintf("  vLLM: vllm serve <model> --host 0.0.0.0 --port %d --enable-prompt-tokens-details", c.o.Port))
	case "llamacpp":
		c.info(fmt.Sprintf("  llama.cpp: llama-server -m <model> --host 0.0.0.0 --port %d", c.o.Port))
	}
}

func (c *connectRun) hintBind() {
	switch c.o.Runtime {
	case "ollama":
		c.info("  → set OLLAMA_HOST=0.0.0.0:11434 for the way you run Ollama:")
		c.info("      terminal:    OLLAMA_HOST=0.0.0.0:11434 ollama serve")
		c.info("      Homebrew:    brew services stop ollama, then run it from a terminal as above, or set it in the app")
		c.info("      Ollama app:  Settings → 'Expose Ollama to the network'")
	case "lmstudio":
		c.info("  → LM Studio: Developer → Settings → Serve on Local Network")
	default:
		c.info("  → restart it with --host 0.0.0.0")
	}
}

func (c *connectRun) hintOllamaContext(target int) {
	c.info(fmt.Sprintf("  → set OLLAMA_CONTEXT_LENGTH=%d the same way you set OLLAMA_HOST, restart Ollama, and the heartbeat", target))
	c.info("    picks the new window up within 5 minutes; no need to run this command again.")
}

func onlyOffered(contexts map[string]int, offered []string) map[string]int {
	out := map[string]int{}
	for _, m := range offered {
		if v, ok := contexts[m]; ok && v > 0 {
			out[m] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func num(f *float64) string {
	if f == nil {
		return "?"
	}
	return strconv.FormatFloat(*f, 'f', -1, 64)
}

func sortedKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

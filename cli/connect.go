package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/machine"
	"github.com/magnusfroste/garageai/cli/internal/netbird"
	"github.com/magnusfroste/garageai/cli/internal/portal"
	"github.com/magnusfroste/garageai/cli/internal/profile"
	rt "github.com/magnusfroste/garageai/cli/internal/runtime"
	"github.com/magnusfroste/garageai/cli/internal/service"
	"github.com/magnusfroste/garageai/cli/internal/tunnel"
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
	Tunnel        string // "netbird" (default) or "bridge" (docs/bridge-tunnel.md)
	NoAutoUpdate  bool   // keep this Bridge version until the operator updates by hand
	Relink        bool   // link in the browser again although this machine is linked already
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
		MeshWait: 90 * time.Second, Tunnel: os.Getenv("GARAGEAI_TUNNEL"),
	}
	o.Port, _ = strconv.Atoi(os.Getenv("GARAGEAI_PORT"))
	profile.RuntimeKey = o.RuntimeAPIKey
	if isAdmin() {
		// The root half has what it needs in o: the secrets leave the environment so that no
		// child (netbird's installer, lsof, docker, ...) inherits them.
		for _, k := range []string{"GARAGEAI_REGISTER_TOKEN", "GARAGEAI_SETUP_KEY", "GARAGEAI_RUNTIME_API_KEY"} {
			os.Unsetenv(k)
		}
	}
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
		case "--relink":
			o.Relink = true
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
		case "--tunnel":
			o.Tunnel, err = next()
		case "--no-auto-update":
			o.NoAutoUpdate = true
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
	// The Bridge tunnel is the default since v0.4.2. A command that carries a NetBird setup key (the
	// portal's legacy command) keeps meaning NetBird, and --tunnel netbird asks for it.
	if o.Tunnel == "" {
		o.Tunnel = "bridge"
		if o.SetupKey != "" {
			o.Tunnel = "netbird"
		}
	}
	if o.Tunnel != "netbird" && o.Tunnel != "bridge" {
		return fmt.Errorf("--tunnel must be bridge (the default) or netbird, not %q", o.Tunnel)
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

func (c *connectRun) bold(s string) { fmt.Println(sgr("1", s)) }
func (c *connectRun) ok(s string)   { fmt.Printf("  %s %s\n", sgr("32", "✓"), s) }
func (c *connectRun) info(s string) { fmt.Printf("  %s\n", s) }
func (c *connectRun) warn(s string) {
	fmt.Fprintf(os.Stderr, "  %s %s\n", sgr("33", "!"), s)
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
		Arch: c.facts.Machine.Arch, GPU: gpu, MemoryGB: mem, Profile: prof, MachineID: machine.ID()})
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
	fmt.Fprintf(os.Stderr, "  %s %s\n", sgr("31", "✗"), msg)
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
		"GARAGEAI_MESH_WAIT_SECONDS": os.Getenv("GARAGEAI_MESH_WAIT_SECONDS"),
		"GARAGEAI_TUNNEL":            os.Getenv("GARAGEAI_TUNNEL"), "GARAGEAI_REUSE_LINK": os.Getenv("GARAGEAI_REUSE_LINK"),
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
	// No token: link this machine in the browser (the default since the portal stopped showing
	// commands with tokens). It runs as the operator, before the root half.
	// Ask for the password now, before the browser opens: after approving, the terminal would
	// otherwise sit behind the browser waiting for it, and nobody sees it.
	if runtime.GOOS == "windows" && !isAdmin() {
		code := elevate("connect", args)
		if code == 0 {
			offerTray(o.Yes) // back in the operator's own, unelevated window
		}
		return code
	}
	if !isAdmin() && runtime.GOOS != "windows" {
		v := exec.Command("sudo", "-v", "-p", "  [sudo] password for %p (Bridge installs its service; asked now so the browser step is not interrupted): ")
		v.Stdin, v.Stdout, v.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := v.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "garageai connect: sudo is needed to install Bridge's service:", err)
			return 1
		}
	}
	reuseLink := func() {
		if cfg, err := service.Load(); err == nil && cfg.RegisterToken != "" {
			o.RegisterURL, o.RegisterToken, o.Name = cfg.RegisterURL, cfg.RegisterToken, cfg.Name
			if o.Tunnel == "" && cfg.Transport != "" {
				o.Tunnel = cfg.Transport
			}
		}
	}
	if isAdmin() && os.Getenv("GARAGEAI_REUSE_LINK") == "1" && o.RegisterToken == "" {
		reuseLink() // the user half found this machine linked already: the root half takes the saved link
	}
	if o.RegisterToken == "" && os.Getenv("GARAGEAI_NO_LINK") == "" {
		if name := linkedAs(o); name != "" {
			// Changing the runtime or the port of a linked garage needs no new link (and no browser):
			// garage-m1 was asked to link again twice on 2026-10-16 for a runtime test.
			fmt.Printf("  Already linked as %s: keeping the link (garageai connect --relink to link again).\n\n", sgr("1", name))
			os.Setenv("GARAGEAI_REUSE_LINK", "1")
			if runtime.GOOS == "windows" {
				reuseLink() // one process on Windows: no root half to hand over to
			}
		} else if code := link(&o, runtimeGiven(args)); code != 0 {
			return code
		}
	}
	if !isAdmin() && runtime.GOOS != "windows" {
		code := asRoot(args)
		if code == 0 {
			offerTray(o.Yes) // the root half is done; the icon belongs to the operator
		}
		return code
	}
	c := &connectRun{o: o}
	if o.RegisterToken != "" && o.RegisterURL == "" {
		o.RegisterURL = portal.API() + "/register-node" // a token alone (e.g. left in the shell): GarageAI's own portal
		c.o.RegisterURL = o.RegisterURL
	}
	if o.RegisterURL != "" && !portal.SecureURL(o.RegisterURL) {
		fmt.Fprintln(os.Stderr, "garageai connect: the portal address must be https:", o.RegisterURL)
		return 2
	}
	if o.RegisterURL != "" && o.RegisterToken != "" {
		c.portal = portal.New(o.RegisterURL, o.RegisterToken)
	}
	c.facts = profile.Build(version)
	if o.Tunnel == "bridge" {
		c.facts.Problems = withoutTunnelMoot(c.facts.Problems)
	}
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
	if o.Tunnel == "bridge" {
		c.ok("Not needed: Bridge carries the tunnel itself (no NetBird, no network interface)")
	} else if netbird.Installed() {
		c.ok("netbird is installed (" + netbird.Version() + ")")
	} else if o.SkipInstall {
		c.fail("netbird is not installed and --skip-install was given.")
	} else {
		c.info("NetBird is not installed. It will be installed from https://pkgs.netbird.io/install.sh")
		for _, pr := range c.facts.Problems {
			if pr.Code == "other_vpn" {
				c.info(pr.Message + ".")
				c.info("Your other VPN keeps working: " + pr.Fix + ".")
			}
		}
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
	var meshIP string
	var live *tunnel.Tunnel
	var grant *portal.TunnelGrant
	if o.Tunnel == "bridge" {
		live, grant = c.bridgeTunnel()
		meshIP = grant.TunnelIP
		defer live.Close()
	} else {
		meshIP = netbird.MeshIP()
		if c.portal != nil {
			// Back on NetBird (or never on the tunnel): the portal must route to the mesh address.
			_ = c.portal.TunnelRelease(context.Background())
		}
	}
	switch {
	case o.Tunnel == "bridge":
		c.ok("Bridge tunnel up: " + meshIP + " (this machine dials out to " + grant.HubEndpoint + "; nothing listens on your network)")
	case meshIP != "" && o.SetupKey == "":
		c.ok("Already on the mesh (no setup key given, keeping the current connection)")
	case o.SetupKey == "":
		c.fail("This machine is not on the mesh and no setup key was given. Run garageai connect without options: you approve this machine in the browser and Bridge gets a setup key.")
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
				c.fail(fmt.Sprintf("%s answers on port %d but rejects the API key: it must be exactly the key the runtime was started with.", o.Runtime, o.Port))
			}
			// The runtime wants a key: find it where the runtime was started (command line,
			// environment, Docker), and only ask the operator when that fails. Never on the
			// command line: pasted keys end up in shell histories and chats.
			if key, source := profile.FindRuntimeKey(o.Port); key != "" {
				if ids, ctxs, err := rt.New(host, o.Port, key).Models(); err == nil {
					o.RuntimeAPIKey, c.o.RuntimeAPIKey = key, key
					profile.RuntimeKey = key
					profile.RuntimeKey = key
					c.info(fmt.Sprintf("%s on port %d requires an API key: found it in the %s.", o.Runtime, o.Port, source))
					served, contexts, probeHost = ids, ctxs, host
					break
				}
			}
			for tries := 0; tries < 3 && probeHost == ""; tries++ {
				key := askSecret(fmt.Sprintf("  %s on port %d requires an API key. Paste the key it was started with (not shown): ", o.Runtime, o.Port))
				if key == "" {
					break
				}
				if ids, ctxs, err := rt.New(host, o.Port, key).Models(); err == nil {
					o.RuntimeAPIKey, c.o.RuntimeAPIKey = key, key
					served, contexts, probeHost = ids, ctxs, host
				} else {
					c.warn("The runtime rejects that key.")
				}
			}
			if probeHost == "" {
				c.fail(fmt.Sprintf("%s answers on port %d but requires an API key, and Bridge could not find or get it. Run the command again and paste the key when asked.", o.Runtime, o.Port))
			}
			break
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
	case o.Tunnel == "bridge":
		c.ok(fmt.Sprintf("Bridge forwards %s:%d in the tunnel to 127.0.0.1:%d; the runtime can stay on localhost", meshIP, o.Port, o.Port))
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
	if hint := ufwBlocks(o.Port); hint != "" && o.Tunnel != "bridge" {
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
		Contexts: onlyOffered(contexts, offered), RuntimeAPIKey: o.RuntimeAPIKey, MachineID: machine.ID()}
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
	switch {
	case res.Message != "" || len(res.Offered) > 0 || len(res.NotOffered) > 0:
		// The portal's own account of the models (2026-10-16): counts and reasons, no test talk
		// unless a test ran and failed, and never "run this again".
		if len(res.Offered) > 0 {
			c.ok(fmt.Sprintf("Offered on GarageAI: %s", strings.Join(res.Offered, ", ")))
		}
		for _, n := range res.NotOffered {
			c.info(fmt.Sprintf("  Not offered: %s (%s)", n.Model, n.Reason))
		}
		if res.Message != "" {
			// The portal's message may run over several lines: indent every one.
			c.info("  " + strings.ReplaceAll(strings.TrimSpace(res.Message), "\n", "\n    "))
		}
		if len(res.Offered) > 0 || passed > 0 {
			c.ok("Node registered — your garage is live on GarageAI.")
			c.report("done", fmt.Sprintf("Registered: %d model(s) offered", len(res.Offered)), c.profile)
		} else {
			c.report("done", "Registered: no model offered yet", c.profile)
		}
	case passed > 0:
		c.ok("Node registered — your garage is live on GarageAI.")
		c.report("done", fmt.Sprintf("Registered: %d model(s) passed the acceptance test", passed), c.profile)
	default:
		c.warn("Registered, but no model is offered yet. GarageAI tests the models again automatically within an hour; nothing to do here. The portal (My garage) shows which models are recognised.")
	}
	c.done = true

	// 6. Heartbeat
	if o.NoHeartbeat {
		return 0
	}
	c.begin("6/6  Heartbeat")
	self, err := serviceBinary()
	if err != nil {
		c.fail("Could not install Bridge for the heartbeat service: " + err.Error())
	}
	cfg := &service.Config{HeartbeatURL: strings.TrimSuffix(o.RegisterURL, "/register-node") + "/node-heartbeat",
		RegisterURL: o.RegisterURL, RegisterToken: o.RegisterToken, Name: o.Name, Runtime: o.Runtime, Port: o.Port,
		RuntimeAPIKey: o.RuntimeAPIKey, MeshIP: meshIP}
	if o.Tunnel == "bridge" {
		cfg.Transport, cfg.HubPublicKey, cfg.HubEndpoint = "tunnel", grant.HubPublicKey, grant.HubEndpoint
	}
	if o.NoAutoUpdate {
		f := false
		cfg.AutoUpdate = &f
	}
	if err := service.Save(cfg); err != nil {
		c.fail("Could not write " + service.ConfigPath() + ": " + err.Error())
	}
	if removed := service.RemoveLegacyScriptHeartbeat(); removed {
		c.info("Removed the connect script's old heartbeat; Bridge reports from now on.")
	}
	if err := service.Install(self); err != nil {
		c.fail("Could not install the heartbeat service: " + err.Error())
	}
	if o.Tunnel == "bridge" {
		live.Close() // the service takes over the tunnel from here
		if err := service.InstallTunnel(self); err != nil {
			c.fail("Could not install the tunnel service: " + err.Error())
		}
		c.ok("Installed: the Bridge tunnel runs as a service and reconnects by itself after sleep or a network change.")
	} else {
		service.RemoveTunnel()
	}
	c.ok("Installed: Bridge reports your models every 5 minutes. Load a new model and it shows up")
	c.info("  under My garages in the portal within 5 minutes. Remove with: garageai uninstall")
	fmt.Println()
	c.bold("Next: the portal")
	c.info("Which models you offer, and at what price, you decide under My garages at https://app.garageai.eu/dashboard/garages.")
	c.info("A new model is not offered until you say so there.")
	c.report("done", "Heartbeat installed", nil)
	c.notes()
	if runtime.GOOS == "windows" && os.Getenv("GARAGEAI_VIA_HOLD") == "" {
		offerTray(c.o.Yes) // started in an Administrator window directly: no unelevated half to ask in
	}
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

// bridgeTunnel is step 2 with the Bridge tunnel: a key pair on this machine (kept across runs),
// the public key to the portal, the tunnel address and the hub back, and the tunnel up so the
// portal's acceptance test can reach the runtime through it.
func (c *connectRun) bridgeTunnel() (*tunnel.Tunnel, *portal.TunnelGrant) {
	if c.portal == nil {
		c.fail("The Bridge tunnel needs the garage's registration (run garageai connect without options, or pass --register-url and --register-token).")
	}
	var priv, pub string
	if b, err := os.ReadFile(service.TunnelKeyPath()); err == nil {
		priv = strings.TrimSpace(string(b))
		if pub, err = tunnel.PublicKey(priv); err != nil {
			priv = ""
		}
	}
	if priv == "" {
		var err error
		if priv, pub, err = tunnel.GenerateKey(); err != nil {
			c.fail("Could not make a tunnel key: " + err.Error())
		}
		if err := service.SaveTunnelKey(priv); err != nil {
			c.fail("Could not save the tunnel key: " + err.Error())
		}
		c.info("Made a tunnel key pair for this machine; only the public key goes to GarageAI.")
	}
	grant, err := c.portal.TunnelAssign(context.Background(), pub)
	switch {
	case errors.Is(err, portal.ErrRejected):
		c.fail("GarageAI rejects this garage's token. Run garageai connect without options to link this machine again.")
	case err != nil:
		c.fail("Could not get a tunnel address from GarageAI: " + err.Error())
	}
	addr, err := netip.ParseAddr(grant.TunnelIP)
	if err != nil {
		c.fail("GarageAI sent an unusable tunnel address: " + grant.TunnelIP)
	}
	t, err := tunnel.Start(tunnel.Config{PrivateKey: priv, HubPublicKey: grant.HubPublicKey, HubEndpoint: grant.HubEndpoint, Address: addr,
		Forwards: map[uint16]string{uint16(c.o.Port): fmt.Sprintf("127.0.0.1:%d", c.o.Port)}})
	if err != nil {
		c.fail("Could not start the tunnel: " + err.Error())
	}
	// The hub learns about a new garage from the portal on its next round (every 10 s), so the
	// first handshakes of a new garage can go unanswered for a little while.
	c.info("Waiting for the gateway to recognise this machine (usually a few seconds, at most a minute)...")
	deadline := time.Now().Add(75 * time.Second)
	for t.LastHandshake().IsZero() {
		if time.Now().After(deadline) {
			t.Close()
			c.fail("No answer from the gateway's tunnel hub at " + grant.HubEndpoint + " within 75 s. Does this network allow UDP out? Run the command again; if it fails the same way, this network blocks UDP (a TCP fallback is planned): connect with NetBird instead, garageai connect --tunnel netbird.")
		}
		time.Sleep(200 * time.Millisecond)
	}
	return t, grant
}

// withoutTunnelMoot drops what the Bridge tunnel makes irrelevant: NetBird, and a runtime that only
// listens on localhost (Bridge forwards to it there).
func withoutTunnelMoot(ps []profile.Problem) []profile.Problem {
	out := ps[:0:0]
	for _, p := range ps {
		switch p.Code {
		case "netbird_missing", "netbird_disconnected", "localhost_only", "ollama_host_not_persistent", "other_vpn":
			continue
		}
		out = append(out, p)
	}
	return out
}

// serviceBinary installs this binary as the heartbeat service's own root-only copy.
func serviceBinary() (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", err
	}
	if self, err = filepath.EvalSymlinks(self); err != nil {
		return "", err
	}
	if self == service.BinaryPath() {
		return self, nil
	}
	return service.InstallBinary(self)
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
	case "xyntetik":
		c.info(fmt.Sprintf("  Xyntetik Runner: runner -m <model.gguf> --serve --port %d --parallel 2   (localhost only; the Bridge tunnel reaches it)", c.o.Port))
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
	switch runtime.GOOS {
	case "darwin":
		c.info(fmt.Sprintf("  → Ollama app: Settings → Context length → %dk or more. Homebrew or a terminal: start Ollama with OLLAMA_CONTEXT_LENGTH=%d.", target/1024, target))
	case "windows":
		c.info(fmt.Sprintf("  → Ollama app: Settings → Context length → %dk or more, or in PowerShell: setx OLLAMA_CONTEXT_LENGTH %d, then quit Ollama in the system tray and start it again.", target/1024, target))
	default:
		c.info(fmt.Sprintf("  → start Ollama with OLLAMA_CONTEXT_LENGTH=%d (for the systemd service: sudo systemctl edit ollama, Environment=\"OLLAMA_CONTEXT_LENGTH=%d\").", target, target))
	}
	c.info("    The heartbeat picks the new window up within 5 minutes; no need to run this command again.")
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

// linkedAs is the garage this machine is linked as, when it is and the operator did not ask to
// link again; "" otherwise. Reads the secret-free config, which any user may.
func linkedAs(o Options) string {
	if o.Relink || o.RegisterToken != "" || o.SetupKey != "" {
		return ""
	}
	p, err := service.LoadPublic()
	if err != nil || p.Name == "" {
		return ""
	}
	return p.Name
}

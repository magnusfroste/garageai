package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/energy"
	"github.com/magnusfroste/garageai/cli/internal/machine"
	"github.com/magnusfroste/garageai/cli/internal/portal"
	"github.com/magnusfroste/garageai/cli/internal/profile"
	rt "github.com/magnusfroste/garageai/cli/internal/runtime"
	"github.com/magnusfroste/garageai/cli/internal/service"
	"github.com/magnusfroste/garageai/cli/internal/tray"
	"github.com/magnusfroste/garageai/cli/internal/tunnel"
)

// run is the heartbeat: every 5 minutes it reports what the runtime serves, and each model's
// context window, so loading or removing a model updates what GarageAI sells without re-running
// connect. --once does one report (what the service runs).
func run(args []string) int {
	once := len(args) > 0 && args[0] == "--once"
	cfg, err := service.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "garageai run: no configuration yet, run `garageai connect` first:", err)
		return 1
	}
	client := portal.New(cfg.RegisterURL, cfg.RegisterToken)
	profile.RuntimeKey, profile.RuntimeKeyPort = cfg.RuntimeAPIKey, cfg.Port // so doctor's probe sees past the key, on that port only
	profile.KnownRuntime = cfg.Runtime                                       // and asks that port only its own runtime's questions
	_ = service.WritePublic(cfg)                                             // garages set up before v0.4.5 get the secret-free config for doctor
	os.Remove(service.BinaryPath() + ".old")                                 // Windows: the previous version, stepped aside by the last self-update
	for {
		if cfg.Transport == "tunnel" && service.EnsureTunnelRunning() {
			fmt.Println(time.Now().UTC().Format(time.RFC3339), "the tunnel service stood still: started it again")
		}
		code := heartbeat(cfg, client)
		if once {
			return code
		}
		// A report that failed on the network (DNS, a reset connection: typical right after a Mac
		// wakes or Wi-Fi reconnects, garage-m1 2026-10-23) is tried again soon, twice, instead of
		// leaving the garage unreported for 5 minutes. A rejected token (2) is not retried.
		for i := 0; code == 1 && i < 2; i++ {
			time.Sleep(30 * time.Second)
			code = heartbeat(cfg, client)
		}
		time.Sleep(5 * time.Minute)
	}
}

func heartbeat(cfg *service.Config, client *portal.Client) int {
	models, contexts := inventory(cfg)
	ag := agent(cfg) // first: it may correct cfg.Runtime
	res, err := client.Heartbeat(context.Background(), portal.Registration{Name: cfg.Name, Port: cfg.Port, Runtime: cfg.Runtime,
		Models: models, Contexts: nonEmpty(contexts), RuntimeAPIKey: cfg.RuntimeAPIKey, Agent: ag, MachineID: machine.ID()})
	writeStatus(cfg, models, contexts, ag, res, err)
	stamp := time.Now().UTC().Format(time.RFC3339)
	switch {
	case errors.Is(err, portal.ErrRejected):
		fmt.Printf("%s heartbeat rejected: %v (My garages → New command, and run it here)\n", stamp, err)
		return 2
	case err != nil:
		fmt.Printf("%s heartbeat failed: %v\n", stamp, err)
		return 1
	}
	line := fmt.Sprintf("%s heartbeat ok: %d model(s)", stamp, len(models))
	if res.Changed {
		line += fmt.Sprintf(", changed (added %s, removed %s, contexts %s)", strings.Join(res.Added, ","), strings.Join(res.Removed, ","), strings.Join(res.ContextsChanged, ","))
	}
	fmt.Println(line)
	maybeSelfUpdate(cfg, res)
	maybeUpgradeRunner(cfg)
	// Managed models: fetch or remove what the portal wants. beat re-reads the inventory and
	// re-sends the heartbeat: during a long pull so the garage does not look dead, and after a
	// reconcile that changed anything so the portal sees the outcome and the new model together
	// (it offers a fetched model when the report says ready and the inventory lists it).
	beat := func() {
		m, c := inventory(cfg)
		a := agent(cfg)
		r, e := client.Heartbeat(context.Background(), portal.Registration{Name: cfg.Name, Port: cfg.Port, Runtime: cfg.Runtime,
			Models: m, Contexts: nonEmpty(c), RuntimeAPIKey: cfg.RuntimeAPIKey, Agent: a, MachineID: machine.ID()})
		writeStatus(cfg, m, c, a, r, e) // download progress for the menu-bar icon
	}
	if reconcileModels(cfg, res, models, rt.New("127.0.0.1", cfg.Port, cfg.RuntimeAPIKey), beat) {
		beat()
	}
	return 0
}

// inventory is everything the runtime serves right now, and each model's context window; which
// models are offered is decided in the portal. If the runtime does not answer it is empty, so
// buyers are not routed here.
func inventory(cfg *service.Config) (models []string, contexts map[string]int) {
	contexts = map[string]int{}
	for _, host := range []string{"127.0.0.1", cfg.MeshIP} {
		if host == "" {
			continue
		}
		c := rt.New(host, cfg.Port, cfg.RuntimeAPIKey)
		ids, ctxs, err := c.Models()
		if err != nil {
			continue
		}
		models, contexts = ids, ctxs
		if cfg.Runtime == "ollama" {
			// Every model gets a window, loaded or not: Ollama only tells /api/ps for loaded ones,
			// and a garage with ten idle models had no caps at the gateway (mac-mini-1cab).
			def := profile.OllamaServerContext(cfg.Port)
			for _, m := range ids {
				if ctx := c.OllamaContext(m, def); ctx > 0 {
					contexts[m] = ctx
				}
			}
		}
		break
	}
	if models == nil {
		models = []string{}
	}
	return models, contexts
}

// maybeSelfUpdate: the portal's bridge_stable decides; Bridge moves there when it is newer, unless
// the operator turned auto-update off. The outcome is recorded for the next heartbeat and doctor.
func maybeSelfUpdate(cfg *service.Config, res *portal.HeartbeatResult) {
	if cfg.AutoUpdate != nil && !*cfg.AutoUpdate {
		return
	}
	st := service.LoadUpdateState()
	st.LastCheck = service.Now()
	if res.BridgeStable == "" || !newerVersion(version, res.BridgeStable) {
		st.LastError = ""
		service.SaveUpdateState(st)
		return
	}
	fmt.Printf("%s updating Bridge %s → %s (the portal's stable version)\n", service.Now(), version, res.BridgeStable)
	if err := selfUpdate(res.BridgeStable, res.BridgeDownloadBase); err != nil {
		st.LastError = err.Error()
		fmt.Printf("%s update failed: %v\n", service.Now(), err)
	} else {
		st.LastError, st.LastUpdate, st.Version = "", service.Now(), res.BridgeStable
		fmt.Printf("%s Bridge %s installed; the next heartbeat runs it\n", service.Now(), res.BridgeStable)
	}
	service.SaveUpdateState(st)
}

// agent is Bridge's own report for the heartbeat: its version and what doctor finds right now.
func agent(cfg *service.Config) *portal.Agent {
	p := profile.Build(version)
	correctRuntime(cfg, p)
	a := &portal.Agent{Tool: p.Tool, Version: version, OS: p.Machine.OS, Arch: p.Machine.Arch, Problems: []portal.AgentProblem{}}
	a.ResumedAt = service.LastResumed(time.Now())
	ust := service.LoadUpdateState()
	a.Update = &portal.AgentUpdate{Auto: cfg.AutoUpdate == nil || *cfg.AutoUpdate, LastCheck: ust.LastCheck, LastUpdate: ust.LastUpdate, LastError: ust.LastError}
	tel := machine.Collect()
	a.Machine = &tel
	a.State, a.Nodes, _, _ = effectiveState(time.Now())
	if cfg.Transport == "tunnel" {
		a.Requests = tunnel.ReadSummary(5 * time.Minute)
		now := time.Now()
		a.Energy = energy.Summarize(energy.ReadMinutes(), tunnel.ReadSamples(), now.Add(-5*time.Minute), now)
	}
	for _, m := range service.ModelStates(service.LoadModelsState()) {
		a.Models = append(a.Models, portal.AgentModel{Model: m.Model, Status: m.Status, By: m.By, CompletedMB: m.CompletedMB, TotalMB: m.TotalMB, Error: m.Error, At: m.At})
	}
	if e := ust.RunnerError; e != "" {
		a.Problems = append(a.Problems, portal.AgentProblem{Severity: "warning", Code: "runner_update_failed", Message: e})
	}
	for _, pr := range p.Problems {
		// The heartbeat's own state is known where it arrives, and the log doctor reads it from is
		// the previous run's.
		if strings.HasPrefix(pr.Code, "heartbeat_") {
			continue
		}
		msg := pr.Message
		if len(msg) > 200 {
			msg = msg[:200]
		}
		a.Problems = append(a.Problems, portal.AgentProblem{Severity: pr.Severity, Code: pr.Code, Message: msg})
	}
	return a
}

// correctRuntime fixes a garage registered as runtime "other" once doctor recognises the runtime on
// its port (dgx-city's vLLM behind Docker and an API key was "other" before v0.4.3). The next
// heartbeat carries the right runtime, and the portal updates the garage.
func correctRuntime(cfg *service.Config, p profile.Profile) {
	if cfg.Runtime != "other" {
		return
	}
	for _, r := range p.Runtimes {
		if r.Port == cfg.Port && connectKinds[r.Kind] {
			cfg.Runtime = r.Kind
			if err := service.Save(cfg); err == nil {
				fmt.Printf("%s runtime recognised as %s (was other)\n", time.Now().UTC().Format(time.RFC3339), r.Kind)
			}
			return
		}
	}
}

func nonEmpty(m map[string]int) map[string]int {
	if len(m) == 0 {
		return nil
	}
	return m
}

// uninstall removes the heartbeat service and Bridge's configuration. The runtime and NetBird
// stay; the garage is removed under My garages in the portal.
func uninstall(args []string) int {
	yes := len(args) > 0 && (args[0] == "--yes" || args[0] == "-y")
	if !isAdmin() {
		_ = tray.RemoveLoginItem(true) // the icon is the operator's: removed in their half
		return asRootSimple("uninstall", args)
	}
	if runtime.GOOS == "windows" {
		_ = tray.RemoveLoginItem(true)
	}
	fmt.Println("This removes Bridge's heartbeat and configuration. Your runtime, models and NetBird are not touched.")
	if !yes {
		fmt.Print("Continue? [y/N] ")
		var a string
		fmt.Scanln(&a)
		if !strings.EqualFold(a, "y") {
			fmt.Println("Aborted.")
			return 1
		}
	}
	service.Remove()
	service.RemoveLegacyScriptHeartbeat()
	fmt.Println("  ✓ Heartbeat removed. The garage shows as offline on GarageAI after 15 minutes.")
	fmt.Println("  To leave the mesh too: sudo netbird down. Finally, remove the garage under My garages in the portal.")
	return 0
}

// installService makes the heartbeat service run this binary: a root-only copy of it, and the
// service pointed at that copy. The installer runs it after an update, so `curl ... | sh` updates
// the heartbeat too. It also moves a service that ran a copy in the operator's home (Bridge before
// v0.2.4) to the safe place.
func installService() int {
	if service.Program() == "" {
		fmt.Println("No heartbeat service on this machine: nothing to update. (Connect with the command from the portal.)")
		return 0
	}
	if !isAdmin() {
		return asRootSimple("install-service", nil)
	}
	dest, err := serviceBinary()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not install Bridge for the heartbeat service:", err)
		return 1
	}
	if err := service.Install(dest); err != nil {
		fmt.Fprintln(os.Stderr, "Could not update the heartbeat service:", err)
		return 1
	}
	fmt.Printf("✓ The heartbeat service runs Bridge %s from %s\n", version, dest)
	// The tunnel service is a long-running process of the same binary: without a restart it keeps
	// the old version (dgx-city ran a v0.6.1 tunnel under a v0.6.5 heartbeat, 2026-10-14, so the
	// 10 s state poll never ran). A restart reconnects within seconds; requests in flight break.
	if inst, _ := service.TunnelInstalled(); inst {
		service.RestartTunnel()
		fmt.Println("✓ The tunnel service was restarted on the new version (reconnects within seconds)")
	}
	return 0
}

// writeStatus leaves what this heartbeat saw and heard in status.json for the menu-bar icon
// (docs/bridge-menubar.md). Nothing secret goes in: no token, no keys, no addresses.
func writeStatus(cfg *service.Config, models []string, contexts map[string]int, ag *portal.Agent, res *portal.HeartbeatResult, err error) {
	st := service.Status{UpdatedAt: time.Now().UTC(), Version: version, Name: cfg.Name, Runtime: cfg.Runtime, Port: cfg.Port,
		Transport: cfg.Transport, Models: models, Contexts: nonEmpty(contexts)}
	var lastOK *time.Time
	if prev := service.LoadStatus(); prev != nil {
		lastOK = prev.Heartbeat.LastOK
		if lastOK == nil && prev.Heartbeat.OK {
			t := prev.UpdatedAt
			lastOK = &t
		}
	}
	switch {
	case errors.Is(err, portal.ErrRejected):
		st.Heartbeat = service.StatusHeartbeat{Rejected: true, Error: "the portal does not accept this garage's link any more", LastOK: lastOK}
	case err != nil:
		msg := "GarageAI could not be reached"
		if strings.HasPrefix(err.Error(), "HTTP ") {
			msg = "GarageAI did not accept the last report (" + strings.SplitN(err.Error(), ":", 2)[0] + ")"
		}
		st.Heartbeat = service.StatusHeartbeat{Error: msg, LastOK: lastOK}
	default:
		now := st.UpdatedAt
		st.Heartbeat = service.StatusHeartbeat{OK: true, LastOK: &now}
	}
	if ag != nil {
		if ag.State != nil {
			st.State = &service.StatusState{Kind: ag.State.Kind, Reason: ag.State.Reason, Until: ag.State.Until, Source: ag.State.Source}
		}
		for _, m := range ag.Models {
			st.Managed = append(st.Managed, service.StatusModel{Model: m.Model, Status: m.Status, CompletedMB: m.CompletedMB, TotalMB: m.TotalMB, Error: m.Error})
		}
		for _, p := range ag.Problems {
			st.Problems = append(st.Problems, service.StatusProblem{Severity: p.Severity, Code: p.Code, Message: p.Message})
		}
		if ag.Machine != nil {
			st.Power, st.Battery = ag.Machine.Power, ag.Machine.BatteryPct
		}
		if ag.Update != nil {
			st.Update.Auto, st.Update.LastError = ag.Update.Auto, ag.Update.LastError
		}
	}
	if res != nil {
		st.Update.Stable = res.BridgeStable
		if g := res.Garage; g != nil {
			st.Portal = &service.StatusPortal{Live: g.Live, Offered: g.Offered, RequestsToday: g.RequestsToday,
				TokensToday: g.TokensToday, EarnedToday: g.EarnedToday, Currency: g.Currency}
		}
	}
	if cfg.Transport == "tunnel" {
		// Today's energy for the menu-bar icon: since local midnight. The request samples cover the
		// last hour only, so today's tokens come from the energy minutes' own day and the samples
		// that remain; the portal's daily figures are the complete ones.
		now := time.Now()
		y, m, d := now.Date()
		if e := energy.Summarize(energy.ReadMinutes(), tunnel.ReadSamples(), time.Date(y, m, d, 0, 0, 0, 0, now.Location()), now); e != nil {
			st.Energy = &service.StatusEnergy{Wh: e.Wh, BusyWh: e.BusyWh, AvgW: e.AvgW, Source: e.Source}
		}
		if e := energy.Summarize(energy.ReadMinutes(), tunnel.ReadSamples(), now.Add(-time.Hour), now); e != nil && st.Energy != nil {
			st.Energy.HourTokensPerKWh = e.TokensPerKWh
		}
	}
	if e := service.WriteStatus(st); e != nil {
		fmt.Printf("%s could not write %s: %v\n", time.Now().UTC().Format(time.RFC3339), service.StatusPath(), e)
	}
}

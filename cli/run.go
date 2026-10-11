package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/portal"
	"github.com/magnusfroste/garageai/cli/internal/profile"
	rt "github.com/magnusfroste/garageai/cli/internal/runtime"
	"github.com/magnusfroste/garageai/cli/internal/service"
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
	for {
		if code := heartbeat(cfg, client); code != 0 && once {
			return code
		}
		if once {
			return 0
		}
		time.Sleep(5 * time.Minute)
	}
}

func heartbeat(cfg *service.Config, client *portal.Client) int {
	// The heartbeat reports everything the runtime serves (the inventory); which models are
	// offered is decided in the portal. If the runtime does not answer, it reports no models so
	// buyers are not routed here.
	var models []string
	contexts := map[string]int{}
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
			for _, m := range ids {
				if ctx := c.OllamaLoadedContext(m); ctx > 0 {
					contexts[m] = ctx
				}
			}
		}
		break
	}
	if models == nil {
		models = []string{}
	}
	res, err := client.Heartbeat(context.Background(), portal.Registration{Name: cfg.Name, Port: cfg.Port, Runtime: cfg.Runtime,
		Models: models, Contexts: nonEmpty(contexts), RuntimeAPIKey: cfg.RuntimeAPIKey, Agent: agent()})
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
	return 0
}

// agent is Bridge's own report for the heartbeat: its version and what doctor finds right now.
func agent() *portal.Agent {
	p := profile.Build(version)
	a := &portal.Agent{Tool: p.Tool, Version: version, OS: p.Machine.OS, Arch: p.Machine.Arch, Problems: []portal.AgentProblem{}}
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
	if os.Geteuid() != 0 {
		return asRootSimple("uninstall", args)
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
	if os.Geteuid() != 0 {
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
	return 0
}

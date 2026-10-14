package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/machine"
	"github.com/magnusfroste/garageai/cli/internal/portal"
	"github.com/magnusfroste/garageai/cli/internal/service"
)

// effectiveState is the garage's lifecycle state right now: Bridge's own pause first, else a
// runtime manager's state file (vLLM Panel), else nil (serving). Also returns the manager's
// nodes and, when a state file exists but is not used, why.
func effectiveState(now time.Time) (st *portal.AgentState, nodes []machine.Node, ignored string, panel *machine.PanelState) {
	if p := service.LoadPauseState(now); p != nil {
		since := p.Since
		st = &portal.AgentState{Kind: p.Kind, Reason: p.Reason, Since: &since, Until: p.Until, Source: "garageai pause"}
	}
	panel, ignored = machine.ReadPanelState(now)
	if panel != nil {
		nodes = panel.Nodes
		if st == nil {
			if kind := panel.Effect(); kind != "" {
				st = &portal.AgentState{Kind: kind, Reason: panel.State.Reason, Since: panel.State.Since, Until: panel.State.Until, Source: panel.Panel}
			}
		}
	}
	return st, nodes, ignored, panel
}

// stateKey is what the tunnel service watches for changes: a heartbeat goes out at once when it changes.
func stateKey(now time.Time) string {
	st, _, _, _ := effectiveState(now)
	if st == nil {
		return "serving"
	}
	return st.Kind + "|" + st.Reason
}

// pauseCmd: `garageai pause [--reason "..."] [--for 15m]` and `garageai resume`.
func pauseCmd(args []string) int {
	if os.Geteuid() != 0 {
		return asRootSimple("pause", args)
	}
	p := service.PauseState{Kind: "maintenance", Since: time.Now().UTC()}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--reason":
			if i+1 < len(args) {
				i++
				p.Reason = args[i]
			}
		case "--for":
			if i+1 < len(args) {
				i++
				d, err := time.ParseDuration(args[i])
				if err != nil || d <= 0 {
					fmt.Fprintln(os.Stderr, "garageai pause: --for wants a duration such as 15m or 2h")
					return 2
				}
				u := p.Since.Add(d)
				p.Until = &u
			}
		default:
			fmt.Fprintln(os.Stderr, "usage: garageai pause [--reason \"redeploy\"] [--for 15m]")
			return 2
		}
	}
	if err := service.SavePauseState(p); err != nil {
		fmt.Fprintln(os.Stderr, "garageai pause:", err)
		return 1
	}
	until := "until garageai resume"
	if p.Until != nil {
		until = "until " + p.Until.Local().Format("15:04") + " or garageai resume"
	}
	fmt.Printf("✓ Paused (%s): the portal drains this garage within a minute, buyers go elsewhere, no alerts. %s.\n", strings.TrimSpace(p.Reason+" "), until)
	return beatNow()
}

func resumeCmd() int {
	if os.Geteuid() != 0 {
		return asRootSimple("resume", nil)
	}
	service.ClearPauseState()
	fmt.Println("✓ Resumed: the garage is offered again after the next heartbeat (now).")
	return beatNow()
}

// beatNow runs one heartbeat so the portal hears the change at once.
func beatNow() int {
	cfg, err := service.Load()
	if err != nil {
		return 0 // not connected yet: nothing to tell
	}
	return heartbeat(cfg, portal.New(cfg.RegisterURL, cfg.RegisterToken))
}

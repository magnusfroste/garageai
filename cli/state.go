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

// effectiveState is the garage's lifecycle state right now: Door's own pause first, else a
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
	if st == nil && powerSource() == "battery" && !service.ServeOnBattery() {
		st = &portal.AgentState{Kind: "away", Reason: batteryReason, Source: "garageai battery"}
	}
	return st, nodes, ignored, panel
}

// batteryReason is the away state a machine on battery reports (see service.ServeOnBatteryPath).
const batteryReason = "on battery; live again when plugged in"

// powerSource is a variable so tests do not depend on the machine they run on.
var powerSource = machine.PowerSource

// stateKey is what the tunnel service watches for changes: a heartbeat goes out at once when it
// changes. The power source is part of it: a laptop unplugged and closed sleeps before the next
// 5-minute heartbeat could say "battery", and Ops then paged "tunnel down" instead of "away"
// (garage-m1, 2026-10-16).
func stateKey(now time.Time) string {
	key := "serving"
	if st, _, _, _ := effectiveState(now); st != nil {
		key = st.Kind + "|" + st.Reason
	}
	return key + "|power=" + powerSource()
}

// pauseCmd: `garageai pause [--reason "..."] [--for 15m]` and `garageai resume`.
func pauseCmd(args []string) int {
	if !isAdmin() {
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
	if !isAdmin() {
		return asRootSimple("resume", nil)
	}
	service.ClearPauseState()
	service.MarkResumed(time.Now())
	fmt.Println("✓ Resumed: the garage is offered again after the next heartbeat (now), also when it was paused in the portal.")
	code := beatNow()
	if why := adminPause(); why != "" {
		fmt.Println("  ! " + why)
	}
	return code
}

// adminPause says, from the last heartbeat's answer, that GarageAI's admin has paused this garage:
// a pause that neither resume nor the portal's own Resume button lifts (garage-m1 and garage-m4
// were paused this way for weeks and nothing on the machine said so, 2026-11-11).
func adminPause() string {
	st := service.LoadStatus()
	if st == nil || st.Portal == nil || st.Portal.PausedBy != "admin" {
		return ""
	}
	why := "This garage is paused by GarageAI, not by you: garageai resume does not lift that. Ask support@garageai.eu"
	if r := st.Portal.PausedReason; r != "" {
		why += " (reason given: " + r + ")"
	}
	return why + "."
}

// batteryCmd: `garageai battery [serve|away]`: whether this garage stays on the market on battery.
func batteryCmd(args []string) int {
	if len(args) == 0 {
		if service.ServeOnBattery() {
			fmt.Println("On battery this garage keeps serving buyers (garageai battery away changes that).")
		} else {
			fmt.Println("On battery this garage is away for buyers and live again when plugged in (garageai battery serve changes that).")
		}
		return 0
	}
	if len(args) != 1 || (args[0] != "serve" && args[0] != "away") {
		fmt.Fprintln(os.Stderr, "usage: garageai battery [serve|away]")
		return 2
	}
	if !isAdmin() {
		return asRootSimple("battery", args)
	}
	if err := service.SetServeOnBattery(args[0] == "serve"); err != nil {
		fmt.Fprintln(os.Stderr, "garageai battery:", err)
		return 1
	}
	if args[0] == "serve" {
		fmt.Println("✓ This garage keeps serving buyers on battery. A laptop that sleeps or leaves the network still drops their requests.")
	} else {
		fmt.Println("✓ On battery this garage is away for buyers, and live again when plugged in.")
	}
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

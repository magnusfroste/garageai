// Package tray is GarageAI in the menu bar (macOS) and the notification area (Windows):
// docs/bridge-menubar.md. This file is the portable part: what the menu says, built from the
// files the services write. The platform files only draw it and pass clicks back.
//
// Security (the same rules as the design): the icon runs as the logged-in user, reads only
// secret-free, world-readable files, never reads bridge.json or a key, opens no port and makes no
// network request. Its only outward actions open a fixed portal address in the browser and run
// `garageai doctor` in a terminal. Nothing a file says is ever opened, run or followed.
package tray

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/service"
	"github.com/magnusfroste/garageai/cli/internal/tunnel"
)

// MyGarageURL is the only address the icon opens.
const MyGarageURL = "https://app.garageai.eu/dashboard/garages"

// ConnectURL is shown when this machine is not connected yet.
const ConnectURL = "https://app.garageai.eu/connect"

// Level is the icon's state.
type Level int

const (
	NotConnected Level = iota
	Live
	Serving
	Paused
	Problem
	Offline
)

func (l Level) String() string {
	return [...]string{"Not connected", "Live", "Serving", "Paused", "Needs you", "Offline"}[l]
}

// Snapshot is everything the icon knows, read from files.
type Snapshot struct {
	Now       time.Time
	Public    *service.Public
	Status    *service.Status
	Pause     *service.PauseState
	Samples   []tunnel.Sample // the tunnel's request timings, last hour
	LoginItem bool            // the icon starts at login
	Version   string          // this binary
}

// Load reads the files. Every read is size-limited and regular-files-only; anything missing or
// invalid is simply absent.
func Load(now time.Time, version string) Snapshot {
	s := Snapshot{Now: now, Version: version, LoginItem: LoginItemInstalled()}
	if b, err := service.ReadSmallFile(service.PublicPath(), 64<<10); err == nil {
		var p service.Public
		if json.Unmarshal(b, &p) == nil && p.Name != "" {
			s.Public = &p
		}
	}
	s.Status = service.LoadStatus()
	if b, err := service.ReadSmallFile(service.PauseStatePath(), 64<<10); err == nil {
		var p service.PauseState
		if json.Unmarshal(b, &p) == nil && p.Kind != "" && (p.Until == nil || now.Before(*p.Until)) {
			s.Pause = &p
		}
	}
	if b, err := service.ReadSmallFile(tunnel.StatsPath(), 8<<20); err == nil {
		_ = json.Unmarshal(b, &s.Samples)
	}
	return s
}

// Item is one row of the menu.
type Item struct {
	Label     string
	Action    string // "" for an information row
	Checked   bool
	Separator bool
}

// Menu is what the platform code draws.
type Menu struct {
	Level   Level
	Tooltip string
	Items   []Item
}

const heartbeatStale = 15 * time.Minute

// Build turns a snapshot into the menu.
func Build(s Snapshot) Menu {
	var items []Item
	info := func(format string, a ...any) { items = append(items, Item{Label: fmt.Sprintf(format, a...)}) }
	sep := func() { items = append(items, Item{Separator: true}) }
	act := func(label, action string) { items = append(items, Item{Label: label, Action: action}) }

	if s.Public == nil {
		info("GarageAI")
		info("This machine is not connected yet")
		sep()
		act("Connect this machine…", "connect")
		act("Check this machine…", "doctor")
		sep()
		tail(&items, s)
		return Menu{Level: NotConnected, Tooltip: "GarageAI · not connected", Items: items}
	}

	level, word, why := state(s)
	name := s.Public.Name
	info("GarageAI · %s", name)
	if why != "" {
		info("%s: %s", word, why)
	} else {
		info("%s", word)
	}
	if st := s.Status; st != nil && len(st.Models) > 0 {
		rt := runtimeName(s.Public.Runtime)
		for i, m := range st.Models {
			if i == 3 {
				info("   and %d more", len(st.Models)-3)
				break
			}
			if i == 0 {
				info("%s · %s", rt, m)
			} else {
				info("   %s", m)
			}
		}
	}

	sep()
	n, last := lastHour(s.Samples, s.Now)
	if s.Status != nil && s.Status.Portal != nil && s.Status.Portal.RequestsToday != nil {
		line := fmt.Sprintf("Today: %d requests", *s.Status.Portal.RequestsToday)
		if e := s.Status.Portal.EarnedToday; e != nil {
			line += " · " + money(*e, s.Status.Portal.Currency)
		}
		info("%s", line)
	}
	switch {
	case n > 0:
		info("Last hour: %d %s, the latest %s", n, plural(n, "request", "requests"), ago(s.Now.Sub(last)))
	case s.Public.Transport == "tunnel":
		info("No requests in the last hour")
	}
	if st := s.Status; st != nil {
		for _, m := range st.Managed {
			switch m.Status {
			case "downloading":
				if m.TotalMB > 0 {
					info("Fetching %s: %.1f of %.1f GB", m.Model, float64(m.CompletedMB)/1024, float64(m.TotalMB)/1024)
				} else {
					info("Fetching %s", m.Model)
				}
			case "failed", "refused":
				info("Could not fetch %s", m.Model)
			}
		}
	}

	sep()
	act("Open My garage…", "open-garage")
	act("Check this machine…", "doctor")
	sep()
	tail(&items, s)

	tip := "GarageAI · " + name + " · " + word
	if len(tip) > 120 {
		tip = tip[:120]
	}
	return Menu{Level: level, Tooltip: tip, Items: items}
}

// tail: the version and the icon's own settings.
func tail(items *[]Item, s Snapshot) {
	v := s.Version
	if s.Status != nil && s.Status.Version != "" {
		v = s.Status.Version // what the services run
	}
	line := "Bridge " + v
	if s.Status != nil && s.Status.Update.Stable != "" && s.Status.Update.Stable != v && strings.TrimPrefix(v, "v") != "dev" {
		if s.Status.Update.Auto {
			line += " · updating to " + s.Status.Update.Stable
		} else {
			line += " · " + s.Status.Update.Stable + " available"
		}
	}
	*items = append(*items, Item{Label: line})
	*items = append(*items, Item{Label: "Open at login", Action: "login", Checked: s.LoginItem})
	*items = append(*items, Item{Label: "Hide icon (the garage keeps running)", Action: "hide"})
}

// state decides the icon: the most important thing first.
func state(s Snapshot) (Level, string, string) {
	st := s.Status
	if s.Pause != nil {
		w := "Paused"
		if s.Pause.Until != nil {
			w += " until " + s.Pause.Until.Local().Format("15:04")
		}
		return Paused, w, s.Pause.Reason
	}
	if st == nil {
		return Offline, "Waiting", "the first report from Bridge's service has not come yet"
	}
	if age := s.Now.Sub(st.UpdatedAt); age > heartbeatStale {
		return Offline, "Offline", "Bridge's service has not reported for " + strings.TrimSuffix(ago(age), " ago")
	}
	if st.Heartbeat.Rejected {
		return Problem, "Needs you", st.Heartbeat.Error + " (connect again)"
	}
	if !st.Heartbeat.OK {
		return Problem, "Offline", st.Heartbeat.Error
	}
	if st.State != nil {
		switch st.State.Kind {
		case "away":
			why := st.State.Reason
			if why == "" && st.Power == "battery" {
				why = "on battery; live again when plugged in"
			}
			return Paused, "Away", why
		case "starting":
			return Paused, "Starting", st.State.Reason
		default:
			return Paused, "Paused", st.State.Reason
		}
	}
	for _, p := range st.Problems {
		if p.Severity == "error" {
			return Problem, "Needs you", firstSentence(p.Message)
		}
	}
	if len(st.Models) == 0 {
		return Problem, "Needs you", runtimeName(st.Runtime) + " does not answer"
	}
	if p := st.Portal; p != nil && p.Live != nil && !*p.Live {
		return Paused, "Not on the market", "nothing offered, or paused in the portal"
	}
	if n, last := lastHour(s.Samples, s.Now); n > 0 && s.Now.Sub(last) < 2*time.Minute {
		return Serving, "Serving", ""
	}
	return Live, "Live", ""
}

func lastHour(samples []tunnel.Sample, now time.Time) (n int, last time.Time) {
	for _, sm := range samples {
		if sm.Path != "chat" && sm.Path != "completions" && sm.Path != "embeddings" {
			continue
		}
		if now.Sub(sm.At) > time.Hour || sm.At.After(now.Add(time.Minute)) {
			continue
		}
		n++
		if sm.At.After(last) {
			last = sm.At
		}
	}
	return n, last
}

func ago(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d / time.Minute)
		return fmt.Sprintf("%d %s ago", m, plural(m, "minute", "minutes"))
	case d < 48*time.Hour:
		h := int(d / time.Hour)
		return fmt.Sprintf("%d %s ago", h, plural(h, "hour", "hours"))
	}
	days := int(d / (24 * time.Hour))
	return fmt.Sprintf("%d days ago", days)
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func money(v float64, cur string) string {
	switch strings.ToUpper(cur) {
	case "", "EUR":
		return fmt.Sprintf("€%.2f", v)
	case "USD":
		return fmt.Sprintf("$%.2f", v)
	}
	return fmt.Sprintf("%.2f %s", v, strings.ToUpper(cur))
}

func firstSentence(s string) string {
	if i := strings.IndexAny(s, ".:;"); i > 0 && i < 90 {
		return s[:i]
	}
	if len(s) > 90 {
		return s[:90] + "…"
	}
	return s
}

var runtimeNames = map[string]string{"ollama": "Ollama", "lmstudio": "LM Studio", "llamacpp": "llama.cpp", "vllm": "vLLM",
	"sglang": "SGLang", "xyntetik": "Xyntetik Runner", "paddock": "Paddock", "mlx": "MLX", "lemonade": "Lemonade", "unsloth": "Unsloth"}

func runtimeName(r string) string {
	if n, ok := runtimeNames[r]; ok {
		return n
	}
	if r == "" {
		return "The runtime"
	}
	return r
}

// Text renders the menu for `garageai tray --print` (and the tests).
func (m Menu) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] %s\n", m.Level, m.Tooltip)
	for _, it := range m.Items {
		switch {
		case it.Separator:
			b.WriteString("  ────\n")
		case it.Action != "" && it.Checked:
			fmt.Fprintf(&b, "  ✓ %s\n", it.Label)
		case it.Action != "":
			fmt.Fprintf(&b, "  > %s\n", it.Label)
		default:
			fmt.Fprintf(&b, "    %s\n", it.Label)
		}
	}
	return b.String()
}

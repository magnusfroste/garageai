package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// PauseState is what `garageai pause` leaves behind: the garage is in maintenance until it is
// resumed or the time passes. The heartbeat carries it as agent.state, the portal drains the
// garage, Ops does not page. World-readable, no secrets.
type PauseState struct {
	Kind   string     `json:"kind"` // maintenance
	Reason string     `json:"reason,omitempty"`
	Since  time.Time  `json:"since"`
	Until  *time.Time `json:"until,omitempty"`
}

// PauseStatePath is a variable so tests can point it at a temp dir.
var PauseStatePath = func() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "pause.json")
	}
	return "/etc/garageai/pause.json"
}

// LoadPauseState returns the pause when it is set and not over, else nil.
func LoadPauseState(now time.Time) *PauseState {
	b, err := os.ReadFile(PauseStatePath())
	if err != nil {
		return nil
	}
	var p PauseState
	if json.Unmarshal(b, &p) != nil || p.Kind == "" {
		return nil
	}
	if p.Until != nil && now.After(*p.Until) {
		return nil
	}
	return &p
}

func SavePauseState(p PauseState) error {
	b, _ := json.MarshalIndent(p, "", "  ")
	tmp := PauseStatePath() + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, PauseStatePath()); err != nil {
		return err
	}
	MakeUserReadable(PauseStatePath()) // the menu-bar icon shows the pause
	return nil
}

func ClearPauseState() { os.Remove(PauseStatePath()) }

// The time `garageai resume` last ran, so the portal can lift a pause set in the portal too: a
// portal pause takes the garage off the market (and, before 2026-10-19, its tunnel), and resume
// on the machine only cleared Bridge's own pause (garage-m4, 2026-10-18). Sent with the heartbeat
// for a day after the resume.
func resumedPath() string { return filepath.Join(filepath.Dir(PauseStatePath()), "resumed-at") }

// MarkResumed records now as the time of the last `garageai resume`.
func MarkResumed(now time.Time) {
	_ = os.WriteFile(resumedPath(), []byte(now.UTC().Format(time.RFC3339)), 0o644)
}

// LastResumed is the time of the last `garageai resume` within the last day, nil otherwise.
func LastResumed(now time.Time) *time.Time {
	b, err := os.ReadFile(resumedPath())
	if err != nil {
		return nil
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(b)))
	if err != nil || now.Sub(t) > 24*time.Hour || t.After(now.Add(time.Minute)) {
		return nil
	}
	return &t
}

package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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
	return os.Rename(tmp, PauseStatePath())
}

func ClearPauseState() { os.Remove(PauseStatePath()) }

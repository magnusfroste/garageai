package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// UpdateState is what the last self-update attempt left behind, for the heartbeat's report and
// doctor. World-readable, no secrets.
type UpdateState struct {
	LastCheck  string `json:"last_check,omitempty"`
	LastUpdate string `json:"last_update,omitempty"`
	LastError  string `json:"last_error,omitempty"`
	Version    string `json:"version,omitempty"` // the version installed by the last update
}

func UpdateStatePath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "update.json")
	}
	return "/etc/garageai/update.json"
}

func LoadUpdateState() UpdateState {
	var st UpdateState
	if b, err := os.ReadFile(UpdateStatePath()); err == nil {
		_ = json.Unmarshal(b, &st)
	}
	return st
}

func SaveUpdateState(st UpdateState) {
	b, _ := json.MarshalIndent(st, "", "  ")
	tmp := UpdateStatePath() + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		_ = os.Rename(tmp, UpdateStatePath())
	}
}

func Now() string { return time.Now().UTC().Format(time.RFC3339) }

// RestartTunnel makes the tunnel service pick up a new binary. The heartbeat needs nothing: it is
// a timer that runs the binary afresh every 5 minutes.
func RestartTunnel() {
	if inst, _ := TunnelInstalled(); !inst {
		return
	}
	switch runtime.GOOS {
	case "linux":
		_ = run("systemctl", "restart", "garageai-tunnel.service")
	case "darwin":
		_ = run("launchctl", "kickstart", "-k", "system/eu.garageai.tunnel")
	case "windows":
		windowsRestartTunnel()
	}
}

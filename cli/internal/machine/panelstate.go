package machine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// A runtime manager (vLLM Panel 1.28.0 first) tells Bridge what it is doing through a small state
// file in the operator's home, written without sudo (docs/bridge.md, lifecycle states). Bridge
// runs as root and reads it with care: a regular file only, small, fresh, and only states that
// take traffic away are honoured (the file can take a garage out of the pool, never put one in).

const (
	PanelStateMaxBytes = 64 << 10
	PanelStateFresh    = 180 * time.Second
)

type PanelState struct {
	Schema  int       `json:"schema"`
	Panel   string    `json:"panel"`
	Updated time.Time `json:"updated"`
	State   struct {
		Kind   string     `json:"kind"` // serving, starting, maintenance, stopped, failed
		Reason string     `json:"reason"`
		Since  *time.Time `json:"since"`
		Until  *time.Time `json:"until"`
	} `json:"state"`
	Model struct {
		ServedName  string `json:"served_name"`
		MaxModelLen int    `json:"max_model_len"`
	} `json:"model"`
	Nodes []Node `json:"nodes"`
	Path  string `json:"-"`
}

// Node is one machine of a multi-node runtime, as the manager sees it.
type Node struct {
	Name           string   `json:"name"`
	Role           string   `json:"role"`
	OK             bool     `json:"ok"`
	Container      string   `json:"container,omitempty"`
	GPUTempC       *float64 `json:"gpu_temp_c,omitempty"`
	GPUPowerW      *float64 `json:"gpu_power_w,omitempty"`
	GPUUtilPct     *float64 `json:"gpu_util_pct,omitempty"`
	MemAvailableMB *int     `json:"mem_available_mb,omitempty"`
}

// panelStatePaths: every user's state file, and root's.
func panelStatePaths() []string {
	var out []string
	for _, pattern := range []string{"/home/*/.local/state/vllm-panel/state.json", "/Users/*/.local/state/vllm-panel/state.json", "/root/.local/state/vllm-panel/state.json"} {
		m, _ := filepath.Glob(pattern)
		out = append(out, m...)
	}
	return out
}

// ReadPanelState returns the freshest valid state file, or nil and why the file(s) found were not
// used ("" when there is none).
func ReadPanelState(now time.Time) (*PanelState, string) {
	return readPanelState(panelStatePaths(), now)
}

func readPanelState(paths []string, now time.Time) (*PanelState, string) {
	var best *PanelState
	why := ""
	for _, p := range paths {
		st, err := os.Lstat(p)
		if err != nil || !st.Mode().IsRegular() || st.Size() > PanelStateMaxBytes {
			why = p + ": not a regular file under 64 KB"
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var ps PanelState
		if json.Unmarshal(b, &ps) != nil || ps.Schema != 1 {
			why = p + ": not a schema 1 state file"
			continue
		}
		if now.Sub(ps.Updated) > PanelStateFresh || ps.Updated.After(now.Add(time.Minute)) {
			why = p + ": older than 3 minutes (the panel's agent is gone or stuck)"
			continue
		}
		ps.Path = p
		if best == nil || ps.Updated.After(best.Updated) {
			best = &ps
		}
	}
	if best != nil {
		return best, ""
	}
	return nil, why
}

// Effect maps the panel's kind to Bridge's state kind: only states that reduce traffic.
// "" means no override (serving, failed, unknown).
func (ps *PanelState) Effect() string {
	switch ps.State.Kind {
	case "maintenance":
		return "maintenance"
	case "starting":
		return "starting"
	case "stopped":
		return "away"
	}
	return ""
}

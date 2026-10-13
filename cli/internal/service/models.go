package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

// ModelState is what Bridge knows about a model the portal asked it to fetch or remove
// (docs/bridge.md, "Models chosen in the portal, fetched by Bridge"). World-readable, no secrets.
// A model Bridge pulled is marked so: the portal may ask to remove those, never one the operator
// added themselves (the guard rail lives here, not only in the portal).
type ModelState struct {
	Model       string `json:"model"`
	Status      string `json:"status"`       // downloading, ready, failed, removed, refused
	By          string `json:"by,omitempty"` // who wanted it: operator, garageai
	CompletedMB int64  `json:"completed_mb,omitempty"`
	TotalMB     int64  `json:"total_mb,omitempty"`
	Error       string `json:"error,omitempty"`
	At          string `json:"at,omitempty"`
	Pulled      bool   `json:"pulled_by_bridge,omitempty"`
}

// ModelsStatePath is a variable so tests can point it at a temp dir.
var ModelsStatePath = func() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "models.json")
	}
	return "/etc/garageai/models.json"
}

// LoadModelsState returns the states by model name (empty when there is no file).
func LoadModelsState() map[string]ModelState {
	out := map[string]ModelState{}
	var list []ModelState
	if b, err := os.ReadFile(ModelsStatePath()); err == nil && json.Unmarshal(b, &list) == nil {
		for _, m := range list {
			out[m.Model] = m
		}
	}
	return out
}

func SaveModelsState(states map[string]ModelState) {
	list := ModelStates(states)
	b, _ := json.MarshalIndent(list, "", "  ")
	tmp := ModelsStatePath() + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		_ = os.Rename(tmp, ModelsStatePath())
	}
}

// ModelStates is the map as a list, sorted by model name.
func ModelStates(states map[string]ModelState) []ModelState {
	list := make([]ModelState, 0, len(states))
	for _, m := range states {
		list = append(list, m)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Model < list[j].Model })
	return list
}

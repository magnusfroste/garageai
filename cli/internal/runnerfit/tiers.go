package runnerfit

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Runner's roster of the model files it validates (docs/model-tiers.json), with the backends each
// one passed on. Since 1.7.0 Tier 1 holds only the files above Runner's download bar and Tier 2
// carries the same kind of recorded pass, so both count. tiers.json is a compact snapshot taken at
// the pinned Version: regenerate it whenever Version moves (the PR that bumps Runner shows both),
// so Bridge recommends what that exact release was validated with, and never fetches a list at run
// time.
//
//go:embed tiers.json
var tiersJSON []byte

type tierRow struct {
	ID           string              `json:"id"`
	Model        string              `json:"model"`
	Quantization string              `json:"quantization"`
	Features     []string            `json:"features"`
	Repo         string              `json:"repo"`
	File         string              `json:"file"`
	Bytes        int64               `json:"bytes"`
	Pass         map[string]tierPass `json:"pass"` // backend (cpu, cuda, metal) -> its passing gate
}

type tierPass struct {
	Date          string `json:"date"`
	DeviceClass   string `json:"device_class"`
	RunnerVersion string `json:"runner_version"`
}

var tierRows = func() []tierRow {
	var t struct {
		Rows []tierRow `json:"rows"`
	}
	if err := json.Unmarshal(tiersJSON, &t); err != nil {
		panic("runnerfit: tiers.json: " + err.Error())
	}
	return t.Rows
}()

// chatModel: a plain text model a buyer can chat with. Not a decision head, an MTP draft head, a
// LoRA adapter, or an experimental surgery of a model.
func (r tierRow) chatModel() bool {
	text := false
	for _, f := range r.Features {
		switch {
		case f == "text":
			text = true
		case f == "decisions", f == "mtp", f == "adapters":
			return false
		}
	}
	return text && !strings.Contains(r.ID, "surgical") && !strings.Contains(r.ID, "removed")
}

func (r tierRow) tools() bool {
	for _, f := range r.Features {
		if strings.HasPrefix(f, "tools") {
			return true
		}
	}
	return false
}

// hf is the -hf spec: the repository and the file's whole stem as the tag. Runner matches a tag as
// a whole token of the file's stem (src/hfhub.c, tag_matches), so the full stem picks exactly this
// file even in a repository with a ladder of quants.
func (r tierRow) hf() string {
	return r.Repo + ":" + strings.TrimSuffix(r.File, ".gguf")
}

// validatedModel picks the best Tier 1 model for a backend ("CUDA", "Metal", "CPU") within budgetGB
// of memory for the model: one that passed on that backend, with room for its KV cache, preferring
// models that call tools (agents) and then the largest. On the CPU it stays at 8 GB or less, where
// decode is still usable. ok is false when none fits; the caller falls back to its own table.
func validatedModel(backend string, budgetGB float64) (Model, bool) {
	key := strings.ToLower(backend)
	var fit []tierRow
	for _, r := range tierRows {
		if _, ok := r.Pass[key]; !ok || !r.chatModel() {
			continue
		}
		gb := float64(r.Bytes) / 1e9
		if gb*1.2+1 > budgetGB || (key == "cpu" && gb > 8) {
			continue
		}
		fit = append(fit, r)
	}
	if len(fit) == 0 {
		return Model{}, false
	}
	sort.SliceStable(fit, func(i, j int) bool {
		if fit[i].tools() != fit[j].tools() {
			return fit[i].tools()
		}
		return fit[i].Bytes > fit[j].Bytes
	})
	r := fit[0]
	p := r.Pass[key]
	why := fmt.Sprintf("validated by Runner on %s (%s) on %s with Runner %s", p.DeviceClass, backend, p.Date, p.RunnerVersion)
	if r.tools() {
		why += "; it calls tools, which agents need"
	}
	return Model{HF: r.hf(), Name: fmt.Sprintf("%s (%s)", r.Model, r.Quantization), SizeGB: roundGB(float64(r.Bytes) / 1e9), Why: why}, true
}

func roundGB(v float64) float64 { return float64(int(v*10+0.5)) / 10 }

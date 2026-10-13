package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/portal"
	rt "github.com/magnusfroste/garageai/cli/internal/runtime"
	"github.com/magnusfroste/garageai/cli/internal/service"
)

// Managed models: the portal's wish list (wanted_models in the heartbeat answer) says which models
// this garage should have; Bridge makes it so and reports. Only Ollama for now. One pull per
// heartbeat run, at most: a pull blocks this run, and the pace is the portal's (macro, not micro).
// A model Bridge pulled may be removed on request; one the operator added is never touched.

// puller is what reconcileModels needs from the runtime; a fake in tests.
type puller interface {
	OllamaPull(ctx context.Context, model string, progress rt.PullProgress) error
	OllamaDelete(model string) error
}

// pullTimeout bounds one pull: a 20 GB model on slow broadband, with margin.
var pullTimeout = 3 * time.Hour

// minFreeGB is what a pull must leave on the disk the models live on.
const minFreeGB = 10

func reconcileModels(cfg *service.Config, res *portal.HeartbeatResult, served []string, client puller, keepalive func()) {
	if len(res.WantedModels) == 0 {
		return
	}
	states := service.LoadModelsState()
	have := map[string]bool{}
	for _, m := range served {
		have[m] = true
	}
	if cfg.Runtime != "ollama" {
		for _, w := range res.WantedModels {
			states[w.Model] = service.ModelState{Model: w.Model, Status: "refused", By: w.By, At: service.Now(),
				Error: "Bridge can fetch models for Ollama only (this garage runs " + cfg.Runtime + ")"}
		}
		service.SaveModelsState(states)
		return
	}
	pulled := false
	for _, w := range res.WantedModels {
		st := states[w.Model]
		st.Model, st.By = w.Model, w.By
		switch w.Action {
		case "pull":
			if have[w.Model] {
				if st.Status != "ready" {
					st.Status, st.Error, st.At = "ready", "", service.Now()
					states[w.Model] = st
				}
				continue
			}
			if pulled || (st.Status == "failed" && recent(st.At, time.Hour)) {
				continue // one pull per run; a failed one is retried after an hour, not every 5 minutes
			}
			pulled = true
			fmt.Printf("%s fetching %s (wanted by %s)\n", service.Now(), w.Model, orUnknown(w.By))
			st.Status, st.Error, st.At, st.Pulled = "downloading", "", service.Now(), true
			states[w.Model] = st
			service.SaveModelsState(states)
			err := pullWithProgress(client, w.Model, func(completed, total int64) {
				st.CompletedMB, st.TotalMB = completed>>20, total>>20
				states[w.Model] = st
				service.SaveModelsState(states)
			}, keepalive)
			if err != nil {
				st.Status, st.Error, st.At = "failed", err.Error(), service.Now()
				fmt.Printf("%s fetching %s failed: %v\n", service.Now(), w.Model, err)
			} else {
				st.Status, st.Error, st.At = "ready", "", service.Now()
				st.CompletedMB = st.TotalMB
				fmt.Printf("%s %s is ready; the next heartbeat reports it\n", service.Now(), w.Model)
			}
			states[w.Model] = st
		case "remove":
			if !have[w.Model] {
				if st.Status != "removed" {
					st.Status, st.At = "removed", service.Now()
					states[w.Model] = st
				}
				continue
			}
			if !st.Pulled {
				st.Status, st.Error, st.At = "refused", "added by the operator, not by GarageAI: remove it yourself (ollama rm "+w.Model+")", service.Now()
				states[w.Model] = st
				continue
			}
			if err := client.OllamaDelete(w.Model); err != nil {
				st.Status, st.Error, st.At = "failed", err.Error(), service.Now()
			} else {
				st.Status, st.Error, st.At = "removed", "", service.Now()
				fmt.Printf("%s removed %s (wanted by %s)\n", service.Now(), w.Model, orUnknown(w.By))
			}
			states[w.Model] = st
		}
	}
	service.SaveModelsState(states)
}

// pullWithProgress runs one pull with the timeout, checks the disk once the size is known, and
// calls keepalive every 2 minutes so the portal keeps hearing from the garage during a long pull.
func pullWithProgress(client puller, model string, progress rt.PullProgress, keepalive func()) error {
	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()
	var diskErr error
	last, checked := time.Now(), false
	err := client.OllamaPull(ctx, model, func(completed, total int64) {
		if !checked {
			checked = true
			if free := freeGB(); free > 0 && float64(total-completed)/1e9 > free-minFreeGB {
				diskErr = fmt.Errorf("not enough disk: %.0f GB needed, %.0f GB free (Bridge keeps %d GB free)", float64(total-completed)/1e9, free, minFreeGB)
				cancel()
				return
			}
		}
		progress(completed, total)
		if keepalive != nil && time.Since(last) > 2*time.Minute {
			last = time.Now()
			keepalive()
		}
	})
	if diskErr != nil {
		return diskErr
	}
	return err
}

func recent(stamp string, within time.Duration) bool {
	t, err := time.Parse(time.RFC3339, stamp)
	return err == nil && time.Since(t) < within
}

func orUnknown(s string) string {
	if s == "" {
		return "the portal"
	}
	return s
}

// modelsCmd: `garageai models` shows what the portal asked for and how it went.
func modelsCmd() int {
	states := service.ModelStates(service.LoadModelsState())
	if len(states) == 0 {
		fmt.Println("No models managed by GarageAI on this garage yet. Add one under My garages in the portal, or let GarageAI choose.")
		return 0
	}
	fmt.Println("Models managed by GarageAI (" + service.ModelsStatePath() + "):")
	for _, m := range states {
		line := fmt.Sprintf("  %-32s %s", m.Model, m.Status)
		if m.Status == "downloading" && m.TotalMB > 0 {
			line += fmt.Sprintf(" %d of %d MB", m.CompletedMB, m.TotalMB)
		}
		if m.By != "" {
			line += " (wanted by " + m.By + ")"
		}
		if m.Error != "" {
			line += ": " + m.Error
		}
		fmt.Println(line)
	}
	if _, err := os.Stat(service.ModelsStatePath()); err != nil {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(err.Error()))
	}
	return 0
}

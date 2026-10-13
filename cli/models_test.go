package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/magnusfroste/garageai/cli/internal/portal"
	rt "github.com/magnusfroste/garageai/cli/internal/runtime"
	"github.com/magnusfroste/garageai/cli/internal/service"
)

type fakeRuntime struct {
	pulled, deleted []string
	pullErr         error
	total           int64
}

func (f *fakeRuntime) OllamaPull(ctx context.Context, model string, p rt.PullProgress) error {
	f.pulled = append(f.pulled, model)
	if f.pullErr != nil {
		return f.pullErr
	}
	if p != nil {
		p(f.total/2, f.total)
		p(f.total, f.total)
	}
	return nil
}

func (f *fakeRuntime) OllamaDelete(model string) error {
	f.deleted = append(f.deleted, model)
	return nil
}

func useTempState(t *testing.T) {
	t.Helper()
	// Point the state file at a temp dir: the tests must not touch /etc.
	dir := t.TempDir()
	old := service.ModelsStatePath
	service.ModelsStatePath = func() string { return filepath.Join(dir, "models.json") }
	t.Cleanup(func() { service.ModelsStatePath = old })
}

func TestReconcilePullsWantedAndReportsReady(t *testing.T) {
	useTempState(t)
	f := &fakeRuntime{total: 4 << 30}
	cfg := &service.Config{Runtime: "ollama", Port: 11434}
	res := &portal.HeartbeatResult{WantedModels: []portal.WantedModel{{Model: "qwen3:4b", Action: "pull", By: "garageai"}, {Model: "gemma4:12b", Action: "pull"}}}
	reconcileModels(cfg, res, []string{"gemma4:12b"}, f, nil)
	if len(f.pulled) != 1 || f.pulled[0] != "qwen3:4b" {
		t.Fatalf("pulled %v, want only qwen3:4b (gemma4 is already there)", f.pulled)
	}
	st := service.LoadModelsState()
	if st["qwen3:4b"].Status != "ready" || !st["qwen3:4b"].Pulled || st["qwen3:4b"].TotalMB != 4096 {
		t.Fatalf("qwen3:4b state %+v", st["qwen3:4b"])
	}
	if st["gemma4:12b"].Status != "ready" || st["gemma4:12b"].Pulled {
		t.Fatalf("gemma4:12b state %+v: ready, and not marked as pulled by Bridge", st["gemma4:12b"])
	}
}

func TestReconcileOnePullPerRunAndRetryAfterAnHour(t *testing.T) {
	useTempState(t)
	f := &fakeRuntime{pullErr: errors.New("pull qwen3:4b: no route to ollama.com")}
	cfg := &service.Config{Runtime: "ollama"}
	res := &portal.HeartbeatResult{WantedModels: []portal.WantedModel{{Model: "a:1", Action: "pull"}, {Model: "b:1", Action: "pull"}}}
	reconcileModels(cfg, res, nil, f, nil)
	if len(f.pulled) != 1 {
		t.Fatalf("pulled %v, want one pull per run", f.pulled)
	}
	if st := service.LoadModelsState()["a:1"]; st.Status != "failed" || st.Error == "" {
		t.Fatalf("a:1 state %+v", st)
	}
	reconcileModels(cfg, res, nil, f, nil) // a:1 failed just now: skip it, b:1 gets its turn
	if len(f.pulled) != 2 || f.pulled[1] != "b:1" {
		t.Fatalf("pulled %v, want b:1 second", f.pulled)
	}
}

func TestReconcileRemovesOnlyWhatBridgePulled(t *testing.T) {
	useTempState(t)
	f := &fakeRuntime{total: 1 << 30}
	cfg := &service.Config{Runtime: "ollama"}
	reconcileModels(cfg, &portal.HeartbeatResult{WantedModels: []portal.WantedModel{{Model: "ours:1", Action: "pull"}}}, nil, f, nil)
	res := &portal.HeartbeatResult{WantedModels: []portal.WantedModel{{Model: "ours:1", Action: "remove"}, {Model: "theirs:1", Action: "remove"}}}
	reconcileModels(cfg, res, []string{"ours:1", "theirs:1"}, f, nil)
	if len(f.deleted) != 1 || f.deleted[0] != "ours:1" {
		t.Fatalf("deleted %v, want only ours:1", f.deleted)
	}
	st := service.LoadModelsState()
	if st["ours:1"].Status != "removed" || st["theirs:1"].Status != "refused" {
		t.Fatalf("states ours=%s theirs=%s", st["ours:1"].Status, st["theirs:1"].Status)
	}
}

func TestReconcileRefusesOtherRegistries(t *testing.T) {
	useTempState(t)
	f := &fakeRuntime{total: 1 << 30}
	res := &portal.HeartbeatResult{WantedModels: []portal.WantedModel{{Model: "evil.example/x/y:1", Action: "pull"}, {Model: "hf.co/org/model:Q4", Action: "pull"}}}
	reconcileModels(&service.Config{Runtime: "ollama"}, res, nil, f, nil)
	st := service.LoadModelsState()
	if len(f.pulled) != 1 || f.pulled[0] != "hf.co/org/model:Q4" || st["evil.example/x/y:1"].Status != "refused" {
		t.Fatalf("pulled %v, states %+v", f.pulled, st)
	}
}

func TestReconcileRefusesOtherRuntimes(t *testing.T) {
	useTempState(t)
	f := &fakeRuntime{}
	reconcileModels(&service.Config{Runtime: "vllm"}, &portal.HeartbeatResult{WantedModels: []portal.WantedModel{{Model: "x", Action: "pull"}}}, nil, f, nil)
	if len(f.pulled) != 0 || service.LoadModelsState()["x"].Status != "refused" {
		t.Fatalf("vLLM garage: pulled %v, state %+v", f.pulled, service.LoadModelsState()["x"])
	}
}

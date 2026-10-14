package service

import (
	"path/filepath"
	"testing"
	"time"
)

func TestPauseStateRoundTripAndExpiry(t *testing.T) {
	dir := t.TempDir()
	old := PauseStatePath
	PauseStatePath = func() string { return filepath.Join(dir, "pause.json") }
	t.Cleanup(func() { PauseStatePath = old })
	now := time.Now().UTC()
	until := now.Add(15 * time.Minute)
	if err := SavePauseState(PauseState{Kind: "maintenance", Reason: "redeploy", Since: now, Until: &until}); err != nil {
		t.Fatal(err)
	}
	if p := LoadPauseState(now.Add(5 * time.Minute)); p == nil || p.Reason != "redeploy" {
		t.Fatalf("within --for: %+v", p)
	}
	if p := LoadPauseState(now.Add(20 * time.Minute)); p != nil {
		t.Fatalf("after --for: %+v", p)
	}
	ClearPauseState()
	if p := LoadPauseState(now); p != nil {
		t.Fatal("after resume")
	}
}

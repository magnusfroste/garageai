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

func TestResumedAt(t *testing.T) {
	dir := t.TempDir()
	old := PauseStatePath
	PauseStatePath = func() string { return filepath.Join(dir, "pause.json") }
	defer func() { PauseStatePath = old }()
	now := time.Date(2026, 10, 19, 1, 0, 0, 0, time.UTC)
	if LastResumed(now) != nil {
		t.Fatal("nothing recorded yet")
	}
	MarkResumed(now)
	if got := LastResumed(now.Add(time.Hour)); got == nil || !got.Equal(now) {
		t.Fatalf("an hour later: %v", got)
	}
	if LastResumed(now.Add(25*time.Hour)) != nil {
		t.Fatal("older than a day must not be sent")
	}
}

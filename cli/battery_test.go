package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/service"
)

func TestAwayOnBattery(t *testing.T) {
	dir := t.TempDir()
	oldPause, oldPower := service.PauseStatePath, powerSource
	service.PauseStatePath = func() string { return filepath.Join(dir, "pause.json") }
	defer func() { service.PauseStatePath, powerSource = oldPause, oldPower }()
	t.Setenv("HOME", dir) // no runtime manager's state file

	powerSource = func() string { return "ac" }
	if st, _, _, _ := effectiveState(time.Now()); st != nil {
		t.Fatalf("on the power adapter the garage serves, got %+v", st)
	}
	powerSource = func() string { return "battery" }
	st, _, _, _ := effectiveState(time.Now())
	if st == nil || st.Kind != "away" || st.Reason != batteryReason {
		t.Fatalf("on battery the garage is away, got %+v", st)
	}
	if err := service.SetServeOnBattery(true); err != nil {
		t.Fatal(err)
	}
	if st, _, _, _ := effectiveState(time.Now()); st != nil {
		t.Fatalf("garageai battery serve keeps it on the market, got %+v", st)
	}
	if err := service.SetServeOnBattery(false); err != nil {
		t.Fatal(err)
	}
	// The operator's own pause wins over the battery.
	service.SavePauseState(service.PauseState{Kind: "maintenance", Reason: "redeploy", Since: time.Now()})
	if st, _, _, _ := effectiveState(time.Now()); st == nil || st.Kind != "maintenance" {
		t.Fatalf("a pause wins over the battery state, got %+v", st)
	}
}

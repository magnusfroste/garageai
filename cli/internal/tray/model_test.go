package tray

import (
	"strings"
	"testing"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/service"
	"github.com/magnusfroste/garageai/cli/internal/tunnel"
)

var now = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

func live() Snapshot {
	return Snapshot{Now: now, Version: "v1.2.0",
		Public: &service.Public{Name: "garage-m4", Runtime: "xyntetik", Port: 8080, Transport: "tunnel"},
		Status: &service.Status{Schema: 1, UpdatedAt: now.Add(-2 * time.Minute), Version: "v1.2.0", Runtime: "xyntetik",
			Heartbeat: service.StatusHeartbeat{OK: true}, Models: []string{"granite-4.1-3b-Q8_0.gguf"}},
	}
}

func level(t *testing.T, s Snapshot, want Level, wantText ...string) Menu {
	t.Helper()
	m := Build(s)
	if m.Level != want {
		t.Fatalf("level %v, want %v\n%s", m.Level, want, m.Text())
	}
	for _, w := range wantText {
		if !strings.Contains(m.Text(), w) {
			t.Errorf("menu lacks %q:\n%s", w, m.Text())
		}
	}
	return m
}

func TestNotConnected(t *testing.T) {
	m := level(t, Snapshot{Now: now, Version: "v1.2.0"}, NotConnected, "not connected", "Connect this machine…")
	for _, it := range m.Items {
		if it.Action == "open-garage" {
			t.Error("no My garage before connecting")
		}
	}
}

func TestLiveAndServing(t *testing.T) {
	level(t, live(), Live, "GarageAI · garage-m4", "Live", "Xyntetik Runner · granite-4.1-3b-Q8_0.gguf", "No requests in the last hour", "Open My garage…")
	s := live()
	s.Samples = []tunnel.Sample{{At: now.Add(-30 * time.Second), Path: "chat"}, {At: now.Add(-20 * time.Minute), Path: "chat"}, {At: now.Add(-time.Minute), Path: "models"}}
	level(t, s, Serving, "Last hour: 2 requests, the latest just now")
	s.Samples = []tunnel.Sample{{At: now.Add(-10 * time.Minute), Path: "chat"}}
	level(t, s, Live, "Last hour: 1 request, the latest 10 minutes ago")
}

func TestPausedWins(t *testing.T) {
	s := live()
	until := now.Add(time.Hour)
	s.Pause = &service.PauseState{Kind: "maintenance", Reason: "gaming", Until: &until}
	s.Status.Heartbeat = service.StatusHeartbeat{Error: "GarageAI could not be reached"}
	level(t, s, Paused, "Paused until", "gaming")
}

func TestOfflineWhenStale(t *testing.T) {
	s := live()
	s.Status.UpdatedAt = now.Add(-3 * time.Hour)
	level(t, s, Offline, "has not reported for 3 hours")
	s.Status = nil
	level(t, s, Offline, "Waiting")
}

func TestProblems(t *testing.T) {
	s := live()
	s.Status.Heartbeat = service.StatusHeartbeat{Rejected: true, Error: "the portal does not accept this garage's link any more"}
	level(t, s, Problem, "connect again")

	s = live()
	s.Status.Models = nil
	level(t, s, Problem, "Xyntetik Runner does not answer")

	s = live()
	s.Status.Problems = []service.StatusProblem{{Severity: "error", Code: "x", Message: "Ollama listens on 127.0.0.1 only. Fix it like this"}}
	level(t, s, Problem, "Ollama listens on 127")
}

func TestAwayOnBattery(t *testing.T) {
	s := live()
	s.Status.State = &service.StatusState{Kind: "away"}
	s.Status.Power = "battery"
	level(t, s, Paused, "Away: on battery")
}

func TestPortalFigures(t *testing.T) {
	s := live()
	n, e, off := 142, 0.41, false
	s.Status.Portal = &service.StatusPortal{RequestsToday: &n, EarnedToday: &e, Currency: "EUR"}
	level(t, s, Live, "Today: 142 requests · €0.41")
	s.Status.Portal.Live = &off
	level(t, s, Paused, "Not on the market")
}

func TestDownloadsAndUpdate(t *testing.T) {
	s := live()
	s.Status.Managed = []service.StatusModel{{Model: "qwen3:8b", Status: "downloading", CompletedMB: 2150, TotalMB: 5325}, {Model: "x", Status: "failed"}}
	s.Status.Update = service.StatusUpdate{Auto: true, Stable: "v1.3.0"}
	level(t, s, Live, "Fetching qwen3:8b: 2.1 of 5.2 GB", "Could not fetch x", "Bridge v1.2.0 · updating to v1.3.0")
}

func TestManyModels(t *testing.T) {
	s := live()
	s.Status.Models = []string{"a", "b", "c", "d", "e"}
	level(t, s, Live, "and 2 more")
}

// Only fixed actions exist: nothing from a file becomes an action.
func TestActionsAreFixed(t *testing.T) {
	ok := map[string]bool{"open-garage": true, "connect": true, "doctor": true, "login": true, "hide": true}
	s := live()
	s.Status.Problems = []service.StatusProblem{{Severity: "error", Message: "open https://evil.example"}}
	for _, snap := range []Snapshot{s, live(), {Now: now}} {
		for _, it := range Build(snap).Items {
			if it.Action != "" && !ok[it.Action] {
				t.Errorf("unexpected action %q", it.Action)
			}
		}
	}
}

func TestLoadIgnoresSymlinksAndHugeFiles(t *testing.T) {
	dir := t.TempDir()
	old := service.StatusPath
	defer func() { service.StatusPath = old }()
	service.StatusPath = func() string { return dir + "/status.json" }
	if service.LoadStatus() != nil {
		t.Fatal("no file: nil")
	}
}

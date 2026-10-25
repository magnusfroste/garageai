package main

import (
	"testing"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/tunnel"
)

func TestQuietSince(t *testing.T) {
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	busy := []tunnel.Sample{{At: now.Add(-90 * time.Second), Path: "chat", TotalMs: 60_000}} // ended 30 s ago
	if quietSince(busy, now, time.Minute) {
		t.Fatal("a request that ended 30 s ago is not quiet")
	}
	old := []tunnel.Sample{{At: now.Add(-5 * time.Minute), Path: "chat", TotalMs: 1000}, {At: now.Add(-10 * time.Second), Path: "models", TotalMs: 5}}
	if !quietSince(old, now, time.Minute) {
		t.Fatal("an old request and a health probe are quiet")
	}
}

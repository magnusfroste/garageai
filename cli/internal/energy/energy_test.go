package energy

import (
	"testing"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/tunnel"
)

func TestParseNvidiaPower(t *testing.T) {
	if w, ok := parseNvidiaPower("142.37\n98.10\n"); !ok || w < 240.4 || w > 240.5 {
		t.Fatalf("two GPUs: %v %v", w, ok)
	}
	if w, ok := parseNvidiaPower("[N/A]\n23.5\n"); !ok || w != 23.5 {
		t.Fatalf("N/A skipped: %v %v", w, ok)
	}
	if _, ok := parseNvidiaPower(""); ok {
		t.Fatal("no nvidia-smi output must be !ok")
	}
}

func TestParsePowermetrics(t *testing.T) {
	out := "CPU Power: 1200 mW\nGPU Power: 8300 mW\nANE Power: 0 mW\nCombined Power (CPU + GPU + ANE): 9500 mW\n"
	if w, ok := parsePowermetrics(out); !ok || w != 9.5 {
		t.Fatalf("combined: %v %v", w, ok)
	}
	if w, ok := parsePowermetrics("CPU Power: 1000 mW\nGPU Power: 4000 mW\n"); !ok || w != 5 {
		t.Fatalf("parts: %v %v", w, ok)
	}
	if _, ok := parsePowermetrics("powermetrics must be invoked as the superuser"); ok {
		t.Fatal("no reading must be !ok")
	}
}

func TestMeterAndSummary(t *testing.T) {
	t0 := time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)
	m := &Meter{}
	// 10 minutes at 100 W every 5 s; busy during the first 5 minutes only.
	for i := 0; i <= 120; i++ {
		now := t0.Add(time.Duration(i) * 5 * time.Second)
		m.Add(now, 100, i <= 60, "nvidia-smi")
	}
	samples := []tunnel.Sample{
		{At: t0.Add(time.Minute), Path: "chat", Status: 200, Model: "qwen3:8b", InTokens: 1000, OutTokens: 3000},
		{At: t0.Add(2 * time.Minute), Path: "chat", Status: 200, Model: "gemma4:12b", InTokens: 500, OutTokens: 1000},
		{At: t0.Add(3 * time.Minute), Path: "chat", Status: 500, Model: "qwen3:8b", OutTokens: 99999}, // failed: not counted
	}
	s := Summarize(m.Snapshot(), samples, t0, t0.Add(10*time.Minute))
	if s == nil {
		t.Fatal("no summary")
	}
	// 600 s at 100 W = 60,000 J = 16.667 Wh; busy 300 s = 8.333 Wh.
	if s.Wh < 16.6 || s.Wh > 16.7 || s.BusyWh < 8.3 || s.BusyWh > 8.4 || s.AvgW != 100 {
		t.Fatalf("energy: %+v", s)
	}
	if s.OutTokens != 4000 || s.InTokens != 1500 {
		t.Fatalf("tokens: %+v", s)
	}
	// 4000 tokens / 0.008333 kWh = 480,000 tokens per kWh.
	if s.TokensPerKWh == nil || *s.TokensPerKWh < 479000 || *s.TokensPerKWh > 481000 {
		t.Fatalf("tokens per kWh: %v", s.TokensPerKWh)
	}
	if len(s.Models) != 2 || s.Models[1].Model != "qwen3:8b" || s.Models[1].OutTokens != 3000 {
		t.Fatalf("models: %+v", s.Models)
	}
	// A gap (sleep) is not filled in.
	m2 := &Meter{}
	m2.Add(t0, 100, true, "x")
	m2.Add(t0.Add(10*time.Minute), 100, true, "x")
	if len(m2.Snapshot()) != 0 {
		t.Fatal("a 10-minute gap was counted")
	}
}

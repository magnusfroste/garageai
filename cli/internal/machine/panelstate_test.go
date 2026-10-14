package machine

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const sample = `{"schema": 1, "panel": "vLLM Panel 1.28.0", "updated": "2026-10-14T18:21:31Z",
 "state": {"kind": "%s", "reason": "redeploy", "since": "2026-10-14T18:21:31Z", "until": "2026-10-14T18:36:31Z"},
 "model": {"served_name": "mimo-v2.6-flash", "max_model_len": 524288},
 "nodes": [{"name": "gx10-cb0f", "role": "head", "ok": true, "container": "running", "gpu_temp_c": 47.0, "gpu_power_w": 13.58, "gpu_util_pct": 0.0, "mem_available_mb": 17533},
           {"name": "gx10-34ee", "role": "worker", "ok": true, "container": "running", "gpu_temp_c": 44.0, "gpu_power_w": 11.5, "gpu_util_pct": 0.0, "mem_available_mb": 20268}]}`

func write(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestReadPanelState(t *testing.T) {
	dir := t.TempDir()
	now, _ := time.Parse(time.RFC3339, "2026-10-14T18:22:00Z")
	p := write(t, dir, "state.json", sprintf(sample, "maintenance"))
	ps, why := readPanelState([]string{p}, now)
	if ps == nil {
		t.Fatalf("not read: %s", why)
	}
	if ps.Effect() != "maintenance" || ps.State.Reason != "redeploy" || len(ps.Nodes) != 2 || ps.Nodes[1].Role != "worker" || *ps.Nodes[0].MemAvailableMB != 17533 || ps.Model.MaxModelLen != 524288 {
		t.Fatalf("parsed %+v", ps)
	}
	for kind, want := range map[string]string{"serving": "", "failed": "", "starting": "starting", "stopped": "away", "weird": ""} {
		ps, _ := readPanelState([]string{write(t, dir, kind+".json", sprintf(sample, kind))}, now)
		if ps == nil || ps.Effect() != want {
			t.Errorf("kind %s: effect %v, want %q", kind, ps, want)
		}
	}
	// Stale: older than 3 minutes is ignored, with a reason.
	if ps, why := readPanelState([]string{p}, now.Add(10*time.Minute)); ps != nil || why == "" {
		t.Fatalf("stale file used: %v %q", ps, why)
	}
	// A symlink is not followed, and a wrong schema is ignored.
	link := filepath.Join(dir, "link.json")
	_ = os.Symlink(p, link)
	if ps, _ := readPanelState([]string{link}, now); ps != nil {
		t.Fatal("symlink followed")
	}
	if ps, why := readPanelState([]string{write(t, dir, "v2.json", `{"schema": 2}`)}, now); ps != nil || why == "" {
		t.Fatalf("schema 2 used: %v", ps)
	}
	if ps, why := readPanelState(nil, now); ps != nil || why != "" {
		t.Fatalf("no file: %v %q", ps, why)
	}
}

func sprintf(format, kind string) string {
	out := ""
	for i := 0; i < len(format); i++ {
		if i+1 < len(format) && format[i] == '%' && format[i+1] == 's' {
			out += kind
			i++
			continue
		}
		out += string(format[i])
	}
	return out
}

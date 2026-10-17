package runnerfit

import "testing"

func TestAssess(t *testing.T) {
	for _, c := range []struct {
		name    string
		m       Machine
		verdict string
		backend string
		model   string
	}{
		{"Torbjörn: Ryzen laptop, Windows, no usable GPU", Machine{GOOS: "windows", GOARCH: "amd64", MemoryGB: 16}, CPUOnly, "CPU", granite.HF},
		{"w11: 4 GB VM", Machine{GOOS: "windows", GOARCH: "amd64", MemoryGB: 4}, CPUOnly, "CPU", gemmaSmoke.HF},
		{"RTX 3070 8 GB, new driver", Machine{GOOS: "windows", GOARCH: "amd64", MemoryGB: 32, GPUs: []GPU{{Vendor: "nvidia", Name: "RTX 3070", MemoryMB: 8192, Driver: "581.15"}}}, Fit, "CUDA", granite.HF},
		{"RTX 4090 24 GB", Machine{GOOS: "linux", GOARCH: "amd64", MemoryGB: 64, GPUs: []GPU{{Vendor: "nvidia", Name: "RTX 4090", MemoryMB: 24564, Driver: "580.65"}}}, Fit, "CUDA", qwen30.HF},
		{"old NVIDIA driver falls back to CPU", Machine{GOOS: "linux", GOARCH: "amd64", MemoryGB: 16, GPUs: []GPU{{Vendor: "nvidia", Name: "GTX 1080", MemoryMB: 8192, Driver: "535.104"}}}, CPUOnly, "CPU", granite.HF},
		{"M1 8 GB", Machine{GOOS: "darwin", GOARCH: "arm64", MemoryGB: 8, GPUs: []GPU{{Vendor: "apple", Name: "Apple M1", UnifiedMemory: true}}}, Fit, "Metal", granite.HF},
		{"M4 Max 64 GB", Machine{GOOS: "darwin", GOARCH: "arm64", MemoryGB: 64, GPUs: []GPU{{Vendor: "apple", Name: "Apple M4 Max", UnifiedMemory: true}}}, Fit, "Metal", qwen30.HF},
		{"Ollama already serves", Machine{GOOS: "darwin", GOARCH: "arm64", MemoryGB: 16, GPUs: []GPU{{Vendor: "apple"}}, Runtimes: []string{"ollama"}}, KeepCurrent, "Metal", granite.HF},
		{"DGX Spark: linux/arm64", Machine{GOOS: "linux", GOARCH: "arm64", MemoryGB: 128}, Unsupported, "CPU", ""},
		{"Intel Mac", Machine{GOOS: "darwin", GOARCH: "amd64", MemoryGB: 16}, Unsupported, "CPU", ""},
	} {
		a := Assess(c.m)
		if a.Verdict != c.verdict || a.Backend != c.backend || a.Model.HF != c.model {
			t.Errorf("%s: got %s/%s/%s, want %s/%s/%s (%v)", c.name, a.Verdict, a.Backend, a.Model.HF, c.verdict, c.backend, c.model, a.Reasons)
		}
		if len(a.Reasons) == 0 {
			t.Errorf("%s: no reason given", c.name)
		}
	}
}

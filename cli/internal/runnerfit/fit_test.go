package runnerfit

import (
	"strings"
	"testing"
)

func TestAssess(t *testing.T) {
	for _, c := range []struct {
		name    string
		m       Machine
		verdict string
		backend string
		model   string
	}{
		{"Torbjörn: Ryzen laptop, Windows, no usable GPU (validated on CPU, calls tools)", Machine{GOOS: "windows", GOARCH: "amd64", MemoryGB: 16}, CPUOnly, "CPU", "ibm-granite/granite-4.2-8b-GGUF:granite-4.2-8b-Q4_K_M"},
		{"w11: 4 GB VM", Machine{GOOS: "windows", GOARCH: "amd64", MemoryGB: 4}, CPUOnly, "CPU", gemmaSmoke.HF},
		{"RTX 3070 8 GB, new driver", Machine{GOOS: "windows", GOARCH: "amd64", MemoryGB: 32, GPUs: []GPU{{Vendor: "nvidia", Name: "RTX 3070", MemoryMB: 8192, Driver: "581.15"}}}, Fit, "CUDA", "ibm-granite/granite-4.2-8b-GGUF:granite-4.2-8b-Q4_K_M"}, // validated on CUDA, calls tools
		{"RTX 4090 24 GB", Machine{GOOS: "linux", GOARCH: "amd64", MemoryGB: 64, GPUs: []GPU{{Vendor: "nvidia", Name: "RTX 4090", MemoryMB: 24564, Driver: "580.65"}}}, Fit, "CUDA", qwen30.HF},
		{"old NVIDIA driver falls back to CPU", Machine{GOOS: "linux", GOARCH: "amd64", MemoryGB: 16, GPUs: []GPU{{Vendor: "nvidia", Name: "GTX 1080", MemoryMB: 8192, Driver: "535.104"}}}, CPUOnly, "CPU", "ibm-granite/granite-4.2-8b-GGUF:granite-4.2-8b-Q4_K_M"},
		{"M1 8 GB", Machine{GOOS: "darwin", GOARCH: "arm64", MemoryGB: 8, GPUs: []GPU{{Vendor: "apple", Name: "Apple M1", UnifiedMemory: true}}}, Fit, "Metal", granite.HF},
		// Nemotron Bolt E4B is validated on Metal and larger, but cannot call tools: Granite 4.1 3B stays.
		{"M4 16 GB", Machine{GOOS: "darwin", GOARCH: "arm64", MemoryGB: 16, GPUs: []GPU{{Vendor: "apple", Name: "Apple M4", UnifiedMemory: true}}}, Fit, "Metal", granite.HF},
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

// Runner's validated list drives the choice, but a model much smaller than what the machine can run
// does not replace a larger one (a 64 GB Mac keeps Qwen3 30B-A3B), and one without tool calling does
// not replace a pick that has it.
func TestValidatedModel(t *testing.T) {
	m, ok := validatedModel("Metal", 9.6)
	if !ok || m.HF != "nvidia/NVIDIA-Nemotron-3.5-Bolt-E4B-4GB:Nemotron-3.5-Bolt-E4B-Q4_K_M" || m.Tools {
		t.Fatalf("Metal: %+v %v", m, ok)
	}
	if m, ok := validatedModel("CPU", 9.6); !ok || !m.Tools {
		t.Fatalf("CPU 16 GB should get a model that calls tools: %+v %v", m, ok)
	}
	if m, ok := validatedModel("CUDA", 12); !ok || m.HF != "ibm-granite/granite-4.2-8b-GGUF:granite-4.2-8b-Q4_K_M" && m.HF != "unsloth/gemma-4-12b-it-qat-GGUF:gemma-4-12b-it-qat-Q4_0" {
		t.Logf("CUDA 12 GB picks %s", m.HF)
	}
	for _, r := range tierRows {
		if (strings.Contains(r.ID, "clef") || strings.Contains(r.ID, "lora") || strings.HasSuffix(r.ID, "mtp-q8_0")) && r.chatModel() {
			t.Errorf("%s counted as a chat model", r.ID)
		}
	}
	if _, ok := validatedModel("CPU", 2); ok {
		t.Fatal("nothing fits 2 GB")
	}
}

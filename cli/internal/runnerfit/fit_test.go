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
	if m, ok := validatedModel("CUDA", 12); !ok || m.HF != "ibm-granite/granite-4.2-8b-GGUF:granite-4.2-8b-Q4_K_M" && m.HF != "google/gemma-4-12B-it-qat-q4_0-gguf:gemma-4-12b-it-qat-q4_0" {
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

// A file Runner says fails its fidelity bar, or one whose licence limits commercial use, runs but
// is never recommended to a garage that sells its answers (Runner 1.15.0 admits both kinds at Tier 2).
func TestValidatedModelSellableOnly(t *testing.T) {
	var failing, restricted int
	for _, r := range tierRows {
		if r.FailsFidelity {
			failing++
		}
		if restrictedLicences[r.Licence] {
			restricted++
		}
	}
	if failing == 0 || restricted == 0 {
		t.Fatalf("snapshot should hold both kinds to test against (failing %d, restricted %d)", failing, restricted)
	}
	for _, backend := range []string{"CPU", "CUDA", "Metal"} {
		for budget := 1.0; budget <= 128; budget += 0.5 {
			m, ok := validatedModel(backend, budget)
			if !ok {
				continue
			}
			for _, r := range tierRows {
				if r.hf() == m.HF && !r.sellable() {
					t.Fatalf("%s %.1f GB recommends %s (fails fidelity %v, licence %s)", backend, budget, r.ID, r.FailsFidelity, r.Licence)
				}
			}
		}
	}
}

// Runner words a failing file's claim in more than one way; each of these says the file misses the
// bar (Runner 1.15.0), and Hemmingway's CC-BY-NC-4.0 forbids selling its answers.
func TestNotSellableRows(t *testing.T) {
	want := map[string]bool{
		"qwen3.8-27b-gsq-rco-iq3_s-recovered": true, // "0.41 mean KLD ..., 8x the bar"; was the CUDA 16 GB pick at 1.10.0
		"granite-4.2-30b-q4_k_m":              true, // "runs below the fidelity bar"; calls tools, would take a 24 GB card
		"qwen3.6-27b-q4_k_m":                  true, // "runs below the fidelity bar"
		"mellum2.1-12b-a2.5b-thinking-q4_k_m": true, // "fails the 0.05 bar"
		"hemmingway-1.5-q8_0":                 true, // passes, but non-commercial
		"granite-4.2-30b-q8_0":                false,
		"deepseek-v2-lite-chat-q6_k":          false, // "both below the bar" is about its siblings
	}
	seen := 0
	for _, r := range tierRows {
		if notSellable, ok := want[r.ID]; ok {
			seen++
			if r.sellable() == notSellable {
				t.Errorf("%s: sellable %v (fails fidelity %v, licence %s)", r.ID, r.sellable(), r.FailsFidelity, r.Licence)
			}
		}
	}
	if seen != len(want) {
		t.Errorf("found %d of %d rows in tiers.json", seen, len(want))
	}
	// The 16 GB card that got the 8x-the-bar Qwen3.8-27B at 1.10.0 now gets Gemma 4 12B.
	if m, _ := validatedModel("CUDA", 16); m.HF != "google/gemma-4-12B-it-qat-q4_0-gguf:gemma-4-12b-it-qat-q4_0" {
		t.Errorf("CUDA 16 GB: %s", m.HF)
	}
}

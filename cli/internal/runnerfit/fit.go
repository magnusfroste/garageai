// Package runnerfit decides whether Xyntetik Runner is a good runtime for this machine, and with
// which model. It is a pure function of what doctor already knows (OS, architecture, GPUs, memory,
// runtimes that answer), so the recommendation is testable and the same in doctor, connect and
// `garageai runtime`. Facts about Runner come from its README and MANUAL (v1.4.2, 2026-10-17):
// release binaries for linux/amd64, darwin/arm64 and windows/amd64; NVIDIA through CUDA with a
// driver from the R580 series or newer (no toolkit); Apple silicon through Metal; no Vulkan, so
// AMD and Intel graphics use the CPU path.
package runnerfit

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is the Runner release Bridge installs. Pinned, like Bridge's own stable version: Runner
// releases often (three in two days), and a new one is tried on our garages before it moves here.
const Version = "v1.4.2"

// Port Runner serves on (its default). Loopback only by design; the Bridge tunnel reaches it there.
const Port = 8080

// GPU is the part of doctor's GPU record the decision needs.
type GPU struct {
	Vendor        string // "nvidia", "amd", "apple", "intel"
	Name          string
	MemoryMB      int // 0 when unknown or unified
	Driver        string
	UnifiedMemory bool
}

// Machine is what the decision is made from.
type Machine struct {
	GOOS, GOARCH string
	MemoryGB     int
	GPUs         []GPU
	Runtimes     []string // kinds of runtimes that answer with at least one model, e.g. "ollama"
}

// Model is a model Runner serves well, with the reason it was chosen.
type Model struct {
	HF     string  // owner/repo[:TAG] for runner -hf
	Name   string  // what the operator reads
	SizeGB float64 // download size
	Why    string
}

// Verdict values.
const (
	Fit         = "fit"         // Runner is a good choice: install it
	CPUOnly     = "cpu"         // works, on the processor: fine for a small model
	KeepCurrent = "keep"        // a runtime already serves models here: keep it
	Unsupported = "unsupported" // no Runner build for this platform
)

// Assessment is the answer.
type Assessment struct {
	Verdict string
	Asset   string // the release file for this platform, "" when unsupported
	Backend string // "CUDA", "Metal" or "CPU"
	Model   Model
	Reasons []string // one line each, in the order they matter
}

var (
	granite = Model{HF: "ibm-granite/granite-4.1-3b-GGUF:Q8_0", Name: "Granite 4.1 3B (Q8_0)", SizeGB: 3.6,
		Why: "the smallest model that passes Runner's fidelity gate; Runner's own recommendation from 8 GB"}
	qwen30 = Model{HF: "Joakimpalm-Zen/Qwen3-30B-A3B-selective-attnQ8_0-expQ4_0-GGUF", Name: "Qwen3 30B-A3B (selective precision)", SizeGB: 18.0,
		Why: "Runner's lead artifact: passes the fidelity gate where the official Q4_K_M fails, and only 3B parameters are active per token"}
	gemmaSmoke = Model{HF: "Joakimpalm-Zen/gemma-4-E2B-it-Q4_0-GGUF", Name: "Gemma 4 E2B (Q4_0, smoke test)", SizeGB: 2.6,
		Why: "the only Runner-published model that fits; it fails Runner's fidelity gate, so treat it as a trial"}
)

// Asset is Runner's release file for a platform, "" when there is none.
func Asset(goos, goarch string) string {
	switch goos + "/" + goarch {
	case "linux/amd64":
		return "runner-linux-x86_64"
	case "darwin/arm64":
		return "runner-macos-arm64"
	case "windows/amd64":
		return "runner-windows-x86_64.exe"
	}
	return ""
}

// Assess decides. It never refuses to answer: an unsupported platform or a machine that serves
// already gets a verdict with the reason, and `garageai runtime install --force` can still go on
// where a build exists.
func Assess(m Machine) Assessment {
	a := Assessment{Asset: Asset(m.GOOS, m.GOARCH), Backend: "CPU"}
	if a.Asset == "" {
		a.Verdict = Unsupported
		a.Reasons = append(a.Reasons, fmt.Sprintf("Runner has no release build for %s/%s yet (there are builds for Linux x86_64, macOS on Apple silicon and Windows x86_64).", m.GOOS, m.GOARCH))
		if m.GOOS == "linux" && m.GOARCH == "arm64" {
			a.Reasons = append(a.Reasons, "On an ARM Linux machine such as DGX Spark, use vLLM, SGLang, llama.cpp or Ollama.")
		}
		return a
	}

	budgetGB := m.MemoryGB // what the model may use: RAM on CPU and unified memory, VRAM on a discrete card
	for _, g := range m.GPUs {
		switch strings.ToLower(g.Vendor) {
		case "nvidia":
			if major := driverMajor(g.Driver); major > 0 && major < 580 {
				a.Reasons = append(a.Reasons, fmt.Sprintf("%s: driver %s is older than the R580 series Runner's CUDA build needs; update the NVIDIA driver to use the GPU (until then Runner runs on the processor).", g.Name, g.Driver))
				continue
			}
			a.Backend = "CUDA"
			if !g.UnifiedMemory && g.MemoryMB > 0 {
				budgetGB = g.MemoryMB / 1024
			}
			a.Reasons = append(a.Reasons, fmt.Sprintf("%s: Runner uses it through CUDA (driver %s).", g.Name, orUnknown(g.Driver)))
		case "apple":
			a.Backend = "Metal"
			a.Reasons = append(a.Reasons, "Apple silicon: Runner uses the GPU through Metal, with the machine's unified memory.")
		case "amd", "intel":
			a.Reasons = append(a.Reasons, fmt.Sprintf("%s: Runner has no Vulkan backend, so AMD and Intel graphics run on the processor.", orUnknown(g.Name)))
		}
		if a.Backend != "CPU" {
			break
		}
	}
	if a.Backend == "CPU" && len(a.Reasons) == 0 {
		a.Reasons = append(a.Reasons, "No GPU Runner can use: it runs on the processor.")
	}

	switch {
	// 22 GB: an 18 GB file plus its KV cache; a "24 GB" card reports 23.9.
	case a.Backend != "CPU" && budgetGB >= 22, a.Backend == "Metal" && m.MemoryGB >= 32, a.Backend == "CPU" && m.MemoryGB >= 48:
		a.Model = qwen30
	case m.MemoryGB >= 6 || budgetGB >= 6:
		a.Model = granite
	default:
		a.Model = gemmaSmoke
	}

	if a.Backend == "CPU" {
		a.Verdict = CPUOnly
	} else {
		a.Verdict = Fit
	}
	if len(m.Runtimes) > 0 {
		a.Verdict = KeepCurrent
		a.Reasons = append([]string{fmt.Sprintf("%s already serves models here, and Bridge works with it: keep it. Runner is for a machine with no runtime (garageai runtime install --force installs it next to it anyway).", strings.Join(m.Runtimes, " and "))}, a.Reasons...)
	}
	return a
}

func driverMajor(v string) int {
	n, _ := strconv.Atoi(strings.SplitN(strings.TrimSpace(v), ".", 2)[0])
	return n
}

func orUnknown(s string) string {
	if strings.TrimSpace(s) == "" {
		return "unknown"
	}
	return s
}

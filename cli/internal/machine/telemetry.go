package machine

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Telemetry is what the heartbeat tells the portal about the machine right now, so that the
// gateway can prefer an idle garage, Smart knows what fits, and a laptop on battery is "away"
// rather than "down" (docs/bridge.md, phase 4). Cheap to collect: one sample every 5 minutes.
// Fields are pointers where "unknown" is a real answer.
type Telemetry struct {
	LoadPct    *int   `json:"load_pct,omitempty"`     // 1-minute load average as a share of the CPUs, 0-100+
	MemTotalMB *int   `json:"mem_total_mb,omitempty"` //
	MemFreeMB  *int   `json:"mem_free_mb,omitempty"`  // free + reclaimable
	DiskFreeGB *int   `json:"disk_free_gb,omitempty"` // on the filesystem that holds the models (/ for now)
	Power      string `json:"power,omitempty"`        // "ac", "battery" or "" (no battery / unknown)
	BatteryPct *int   `json:"battery_pct,omitempty"`  //
	UserActive *bool  `json:"user_active,omitempty"`  // a person used the keyboard or mouse in the last 5 minutes
	GPUUtilPct *int   `json:"gpu_util_pct,omitempty"` // NVIDIA only (nvidia-smi); Apple silicon has no cheap counter
	VRAMUsedMB *int   `json:"vram_used_mb,omitempty"` // empty on unified-memory GPUs (DGX Spark): mem_free_mb is the headroom
	VRAMTotMB  *int   `json:"vram_total_mb,omitempty"`
	GPUTempC   *int   `json:"gpu_temp_c,omitempty"`
	GPUPowerW  *int   `json:"gpu_power_w,omitempty"`
	GPUUnified bool   `json:"gpu_unified,omitempty"` // the GPU shares the machine's memory (nvidia-smi reports N/A)
}

// Collect gathers the telemetry; it never fails, it leaves out what it cannot read.
func Collect() Telemetry {
	t := collectOS()
	if g, ok := nvidiaSMI(); ok {
		t.GPUUtilPct = intp(g.util)
		if g.unified {
			t.GPUUnified = true
		} else {
			t.VRAMUsedMB, t.VRAMTotMB = intp(g.usedMB), intp(g.totalMB)
		}
		if g.tempC >= 0 {
			t.GPUTempC = intp(g.tempC)
		}
		if g.powerW >= 0 {
			t.GPUPowerW = intp(g.powerW)
		}
	}
	return t
}

func run(name string, args ...string) string {
	if _, err := exec.LookPath(name); err != nil {
		return ""
	}
	cmd := exec.Command(name, args...)
	done := make(chan []byte, 1)
	go func() { out, _ := cmd.Output(); done <- out }()
	select {
	case out := <-done:
		return string(out)
	case <-time.After(4 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return ""
	}
}

type nvidiaGPU struct {
	util, usedMB, totalMB, tempC, powerW int
	unified                              bool
}

// nvidiaSMI: the first GPU's utilisation, memory, temperature and power, when nvidia-smi exists.
// A unified-memory GPU (DGX Spark / GB10) answers [N/A] for memory.
func nvidiaSMI() (nvidiaGPU, bool) {
	out := run("nvidia-smi", "--query-gpu=utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw", "--format=csv,noheader,nounits")
	return parseNvidia(out)
}

func parseNvidia(out string) (nvidiaGPU, bool) {
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, ",")
		if len(f) < 3 {
			continue
		}
		num := func(i int) (int, bool) {
			if i >= len(f) {
				return -1, false
			}
			v, err := strconv.ParseFloat(strings.TrimSpace(f[i]), 64)
			if err != nil {
				return -1, false
			}
			return int(v), true
		}
		g := nvidiaGPU{tempC: -1, powerW: -1}
		u, ok := num(0)
		if !ok {
			continue
		}
		g.util = u
		used, okU := num(1)
		total, okT := num(2)
		if okU && okT {
			g.usedMB, g.totalMB = used, total
		} else {
			g.unified = true
		}
		if t, ok := num(3); ok {
			g.tempC = t
		}
		if p, ok := num(4); ok {
			g.powerW = p
		}
		return g, true
	}
	return nvidiaGPU{}, false
}

func intp(v int) *int    { return &v }
func boolp(v bool) *bool { return &v }

// statfsFreeGB is implemented per OS (0 when unknown).
func diskFreeGB(path string) *int {
	if gb := statfsFreeGB(path); gb > 0 {
		return intp(gb)
	}
	return nil
}

func readFile(p string) string {
	b, _ := os.ReadFile(p)
	return string(b)
}

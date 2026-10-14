package machine

import (
	"os"
	"os/exec"
	"regexp"
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
	VRAMUsedMB *int   `json:"vram_used_mb,omitempty"` //
	VRAMTotMB  *int   `json:"vram_total_mb,omitempty"`
}

// Collect gathers the telemetry; it never fails, it leaves out what it cannot read.
func Collect() Telemetry {
	t := collectOS()
	if u, used, total, ok := nvidiaSMI(); ok {
		t.GPUUtilPct, t.VRAMUsedMB, t.VRAMTotMB = &u, &used, &total
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

var nvidiaLine = regexp.MustCompile(`^\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)`)

// nvidiaSMI: utilisation and memory of the first GPU, when nvidia-smi exists.
func nvidiaSMI() (util, usedMB, totalMB int, ok bool) {
	out := run("nvidia-smi", "--query-gpu=utilization.gpu,memory.used,memory.total", "--format=csv,noheader,nounits")
	return parseNvidia(out)
}

func parseNvidia(out string) (util, usedMB, totalMB int, ok bool) {
	for _, line := range strings.Split(out, "\n") {
		if m := nvidiaLine.FindStringSubmatch(line); m != nil {
			util, _ = strconv.Atoi(m[1])
			usedMB, _ = strconv.Atoi(m[2])
			totalMB, _ = strconv.Atoi(m[3])
			return util, usedMB, totalMB, true
		}
	}
	return 0, 0, 0, false
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

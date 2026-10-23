package energy

import (
	"context"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Watts is the compute power now: all NVIDIA GPUs (nvidia-smi power.draw), or Apple silicon's CPU,
// GPU and Neural Engine together (powermetrics, which needs root: the tunnel service has it).
// ok is false when this machine has no reading.
func Watts() (float64, string, bool) {
	if runtime.GOOS == "darwin" {
		out := runFor(3*time.Second, "powermetrics", "-n", "1", "-i", "500", "--samplers", "cpu_power,gpu_power")
		if w, ok := parsePowermetrics(out); ok {
			return w, "powermetrics", true
		}
		return 0, "", false
	}
	out := runFor(3*time.Second, "nvidia-smi", "--query-gpu=power.draw", "--format=csv,noheader,nounits")
	if w, ok := parseNvidiaPower(out); ok {
		return w, "nvidia-smi", true
	}
	return 0, "", false
}

func runFor(d time.Duration, name string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		return ""
	}
	return string(out)
}

// parseNvidiaPower sums power.draw over the GPUs ("142.37" per line; "[N/A]" lines are skipped).
func parseNvidiaPower(out string) (float64, bool) {
	total, any := 0.0, false
	for _, line := range strings.Split(out, "\n") {
		v, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
		if err != nil || v < 0 || v > 5000 {
			continue
		}
		total += v
		any = true
	}
	return total, any
}

var (
	combinedRe = regexp.MustCompile(`Combined Power \(CPU \+ GPU \+ ANE\):\s*([\d.]+)\s*mW`)
	partRe     = regexp.MustCompile(`(?m)^(CPU|GPU|ANE) Power:\s*([\d.]+)\s*mW`)
)

// parsePowermetrics reads Apple silicon's combined power (mW) from powermetrics, or adds the CPU,
// GPU and ANE lines when the combined line is missing (older macOS).
func parsePowermetrics(out string) (float64, bool) {
	if m := combinedRe.FindStringSubmatch(out); m != nil {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			return v / 1000, true
		}
	}
	total, any := 0.0, false
	for _, m := range partRe.FindAllStringSubmatch(out, -1) {
		if v, err := strconv.ParseFloat(m[2], 64); err == nil {
			total += v
			any = true
		}
	}
	return total / 1000, any
}

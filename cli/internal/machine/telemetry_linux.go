package machine

import (
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

func collectOS() Telemetry {
	var t Telemetry
	t.LoadPct = loadPct(readFile("/proc/loadavg"), runtime.NumCPU())
	if total, free, ok := linuxMemory(readFile("/proc/meminfo")); ok {
		t.MemTotalMB, t.MemFreeMB = intp(total), intp(free)
	}
	t.DiskFreeGB = diskFreeGB("/")
	t.Power, t.BatteryPct = powerSource()
	// user_active: no cheap, display-server-independent source on Linux; left unknown.
	return t
}

// powerSupplies reads type/status/online/capacity of every supply under the sysfs directory.
func powerSupplies(dir string) map[string]map[string]string {
	out := map[string]map[string]string{}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		s := map[string]string{}
		for _, k := range []string{"type", "status", "online", "capacity"} {
			if b, err := os.ReadFile(filepath.Join(dir, e.Name(), k)); err == nil {
				s[k] = string(b)
			}
		}
		if len(s) > 0 {
			out[e.Name()] = s
		}
	}
	return out
}

func statfsFreeGB(path string) int {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0
	}
	return int(uint64(st.Bavail) * uint64(st.Bsize) / 1e9)
}

func powerSource() (string, *int) { return linuxPower(powerSupplies("/sys/class/power_supply")) }

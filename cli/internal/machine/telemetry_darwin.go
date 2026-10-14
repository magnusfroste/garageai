package machine

import (
	"runtime"
	"syscall"
)

func collectOS() Telemetry {
	var t Telemetry
	t.LoadPct = loadPct(run("sysctl", "-n", "vm.loadavg"), runtime.NumCPU())
	if total, free, ok := darwinMemory(run("sysctl", "-n", "hw.memsize"), run("vm_stat")); ok {
		t.MemTotalMB, t.MemFreeMB = intp(total), intp(free)
	}
	t.DiskFreeGB = diskFreeGB("/")
	t.Power, t.BatteryPct = parsePmsetBatt(run("pmset", "-g", "batt"))
	if idle, ok := parseHIDIdle(run("ioreg", "-c", "IOHIDSystem", "-d", "4", "-r")); ok {
		t.UserActive = boolp(idle < 300)
	}
	return t
}

func statfsFreeGB(path string) int {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0
	}
	return int(uint64(st.Bavail) * uint64(st.Bsize) / 1e9)
}

//go:build !windows

package profile

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

func totalMemoryGB() *int {
	if runtime.GOOS == "darwin" {
		if n, err := strconv.ParseInt(strings.TrimSpace(run("sysctl", "-n", "hw.memsize")), 10, 64); err == nil {
			return intp(int(n / (1 << 30)))
		}
		return nil
	}
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil
	}
	for _, line := range strings.Split(string(b), "\n") {
		if f := strings.Fields(line); len(f) >= 2 && f[0] == "MemTotal:" {
			if kb, err := strconv.ParseInt(f[1], 10, 64); err == nil {
				return intp(int(kb / 1048576))
			}
		}
	}
	return nil
}

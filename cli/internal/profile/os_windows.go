//go:build windows

package profile

import (
	"encoding/csv"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func listeners() []Listener {
	ls := parseNetstat(run("netstat", "-ano", "-p", "TCP") + "\n" + run("netstat", "-ano", "-p", "TCPv6"))
	names := map[int]string{}
	r := csv.NewReader(strings.NewReader(run("tasklist", "/fo", "csv", "/nh")))
	if rows, err := r.ReadAll(); err == nil {
		for _, row := range rows {
			if len(row) >= 2 {
				if pid, err := strconv.Atoi(row[1]); err == nil {
					names[pid] = strings.TrimSuffix(row[0], ".exe")
				}
			}
		}
	}
	for i := range ls {
		ls[i].Process = names[ls[i].PID]
	}
	return ls
}

// cmdline is not read on Windows yet (it needs WMI); flags stay unknown there.
func cmdline(int) []string { return nil }

// totalMemoryGB uses GlobalMemoryStatusEx from kernel32 (stdlib syscall, no dependency).
func totalMemoryGB() *int {
	type memStatusEx struct {
		Length               uint32
		MemoryLoad           uint32
		TotalPhys            uint64
		AvailPhys            uint64
		TotalPageFile        uint64
		AvailPageFile        uint64
		TotalVirtual         uint64
		AvailVirtual         uint64
		AvailExtendedVirtual uint64
	}
	var m memStatusEx
	m.Length = uint32(unsafe.Sizeof(m))
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
	if ok, _, _ := proc.Call(uintptr(unsafe.Pointer(&m))); ok == 0 {
		return nil
	}
	return intp(int(m.TotalPhys / (1 << 30)))
}

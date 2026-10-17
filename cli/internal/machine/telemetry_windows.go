package machine

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows: memory, disk and power from kernel32 (no commands, no WMI). Not here: load (Windows
// has no load average) and user_active (the heartbeat runs as SYSTEM in session 0, which never
// sees the operator's keyboard). Power is what makes a laptop garage "away" rather than "down"
// (Torbjörn's Lenovo, 2026-10-17).
var (
	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetSystemPowerStatus = kernel32.NewProc("GetSystemPowerStatus")
)

type memoryStatusEx struct {
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

type systemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

func collectOS() Telemetry {
	var t Telemetry
	m := memoryStatusEx{}
	m.Length = uint32(unsafe.Sizeof(m))
	if r, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&m))); r != 0 {
		t.MemTotalMB, t.MemFreeMB = intp(int(m.TotalPhys>>20)), intp(int(m.AvailPhys>>20))
	}
	drive := os.Getenv("SystemDrive")
	if drive == "" {
		drive = "C:"
	}
	t.DiskFreeGB = diskFreeGB(drive + `\`)
	t.Power, t.BatteryPct = powerSource()
	return t
}

func statfsFreeGB(path string) int {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0
	}
	var free, total, totalFree uint64
	if windows.GetDiskFreeSpaceEx(p, &free, &total, &totalFree) != nil {
		return 0
	}
	return int(free >> 30)
}

// powerSource: "ac" or "battery" on a machine with a battery, "" on a desktop (no battery: the
// question does not apply), as on macOS and Linux.
func powerSource() (string, *int) {
	var s systemPowerStatus
	if r, _, _ := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&s))); r == 0 {
		return "", nil
	}
	return windowsPower(s.ACLineStatus, s.BatteryFlag, s.BatteryLifePercent)
}

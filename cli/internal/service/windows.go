package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// Windows: the heartbeat is a scheduled task every 5 minutes as SYSTEM, the tunnel is a Windows
// service (sc.exe) running `garageai tunnel` under the service control manager (tunnel_service_
// windows.go in the main package), and the files under %ProgramData%\GarageAI are readable by
// Administrators and SYSTEM only (icacls), with the secret-free bridge-public.json readable by
// Users. All of it goes through the stock tools, so it compiles everywhere and is tested in CI
// on windows-latest.

const (
	WindowsTask    = "GarageAI Bridge heartbeat"
	WindowsService = "GarageAIBridgeTunnel"
)

// Well-known SIDs, locale-independent: Administrators, SYSTEM, Users.
const (
	sidAdmins = "*S-1-5-32-544"
	sidSystem = "*S-1-5-18"
	sidUsers  = "*S-1-5-32-545"
)

func quoteWin(s string) string { return `"` + s + `"` }

func windowsInstallHeartbeat(self string) error {
	// /RU SYSTEM: runs without anyone logged in, and as the account that may read bridge.json.
	if err := run("schtasks", "/Create", "/TN", WindowsTask, "/SC", "MINUTE", "/MO", "5", "/RU", "SYSTEM", "/RL", "HIGHEST",
		"/TR", quoteWin(self)+" run --once", "/F"); err != nil {
		return err
	}
	_ = run("schtasks", "/Run", "/TN", WindowsTask) // the first heartbeat now, not in 5 minutes
	return nil
}

func windowsHeartbeatInstalled() (installed, active bool) {
	out, err := exec.Command("schtasks", "/Query", "/TN", WindowsTask, "/FO", "LIST").Output()
	if err != nil {
		return false, false
	}
	return true, !strings.Contains(strings.ToLower(string(out)), "disabled")
}

var taskCommand = regexp.MustCompile(`<Command>([^<]+)</Command>`)

// windowsHeartbeatProgram is the exe the scheduled task runs ("" when there is no task). Before
// 1.0.3 Program() knew no Windows case, so install-service after an update said there was no
// heartbeat service and left the old version running (w11, 2026-10-17).
func windowsHeartbeatProgram() string {
	out, err := exec.Command("schtasks", "/Query", "/TN", WindowsTask, "/XML").Output()
	if err != nil {
		return ""
	}
	s := strings.ReplaceAll(string(out), "\x00", "") // UTF-16 when redirected, on some Windows
	if m := taskCommand.FindStringSubmatch(s); m != nil {
		return strings.Trim(strings.TrimSpace(m[1]), `"`)
	}
	return BinaryPath()
}

func windowsRemoveHeartbeat() { _ = run("schtasks", "/Delete", "/TN", WindowsTask, "/F") }

func windowsInstallTunnel(self string) error {
	bin := quoteWin(self) + " tunnel"
	if inst, _ := windowsTunnelInstalled(); inst {
		_ = run("sc", "stop", WindowsService)
		if err := run("sc", "config", WindowsService, "binPath=", bin, "start=", "auto"); err != nil {
			return err
		}
	} else if err := run("sc", "create", WindowsService, "binPath=", bin, "start=", "auto", "DisplayName=", "GarageAI Bridge tunnel"); err != nil {
		return err
	}
	_ = run("sc", "description", WindowsService, "This garage's tunnel to the GarageAI gateway (dials out; nothing listens).")
	_ = run("sc", "failure", WindowsService, "reset=", "60", "actions=", "restart/5000/restart/5000/restart/5000")
	// Also restart after an orderly stop with an error code, not only after a crash (w11's tunnel
	// service stood still from 18:40, 2026-10-17, and Windows did nothing).
	_ = run("sc", "failureflag", WindowsService, "1")
	return windowsStartTunnel()
}

// windowsStartTunnel starts the tunnel service; "already running" (1056) is success: the
// heartbeat's self-heal may have started it a moment earlier.
func windowsStartTunnel() error {
	if err := run("sc", "start", WindowsService); err != nil && !strings.Contains(err.Error(), "1056") {
		return err
	}
	return nil
}

func windowsTunnelInstalled() (installed, active bool) {
	out, err := exec.Command("sc", "query", WindowsService).Output()
	if err != nil {
		return false, false
	}
	s := strings.ToUpper(string(out))
	return !strings.Contains(s, "1060"), strings.Contains(s, "RUNNING")
}

func windowsRemoveTunnel() {
	_ = run("sc", "stop", WindowsService)
	_ = run("sc", "delete", WindowsService)
}

func windowsRestartTunnel() {
	_ = run("sc", "stop", WindowsService)
	// sc stop returns at once; starting while it is still stopping fails as "already running".
	for i := 0; i < 20; i++ {
		if out, _ := exec.Command("sc", "query", WindowsService).Output(); strings.Contains(strings.ToUpper(string(out)), "STOPPED") {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	_ = windowsStartTunnel()
}

// windowsRestrict makes path readable and writable by Administrators and SYSTEM only (and, with
// usersRead, readable by Users). Directories pass the rule on to what is created inside them.
func windowsRestrict(path string, isDir, usersRead bool) error {
	inh := ""
	if isDir {
		inh = "(OI)(CI)"
	}
	args := []string{path, "/inheritance:r", "/grant:r", sidAdmins + ":" + inh + "F", "/grant:r", sidSystem + ":" + inh + "F"}
	if usersRead {
		args = append(args, "/grant:r", sidUsers+":"+inh+"R")
	}
	return run("icacls", args...)
}

// WindowsLogPath is where the tunnel service writes (a service has no console).
func WindowsLogPath() string {
	return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "tunnel.log")
}

// WindowsUserBinDir is where the operator's own copy of garageai.exe goes, on their PATH.
func WindowsUserBinDir() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "GarageAI")
}

func windowsErr(what string, err error) error { return fmt.Errorf("%s: %w", what, err) }

// EnsureTunnelRunning starts the tunnel service when it is installed but stands still. Windows
// only: systemd and launchd keep their services alive themselves. The heartbeat calls it every
// 5 minutes, so a stopped tunnel is back within one beat instead of waiting for someone.
func EnsureTunnelRunning() (started bool) {
	if runtime.GOOS != "windows" {
		return false
	}
	if inst, active := windowsTunnelInstalled(); inst && !active {
		return windowsStartTunnel() == nil
	}
	return false
}

package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Xyntetik Runner as a service of Bridge's (`garageai runtime install`): for an operator with no
// runtime. Runner's own tray app runs in the user's session; a garage must survive a restart with
// nobody logged in, so Bridge runs `runner -hf MODEL --serve` the way it runs the tunnel: a systemd
// unit, a launchd daemon, and on Windows a scheduled task at startup (Runner is not a Windows
// service program, so the service control manager cannot run it). Runner binds loopback only;
// the Bridge tunnel reaches it there. Its model cache lives under a home of its own.
const (
	runnerUnitPath  = "/etc/systemd/system/garageai-runner.service"
	runnerPlistPath = "/Library/LaunchDaemons/eu.garageai.runner.plist"
	runnerLogPath   = "/var/log/garageai-runner.log"
	WindowsRunner   = "GarageAI Runner"
)

// RunnerPath is where Bridge keeps the Runner binary (next to its own service copy, root-only).
func RunnerPath() string {
	name := "runner"
	if runtime.GOOS == "windows" {
		name = "runner.exe"
	}
	return filepath.Join(filepath.Dir(BinaryPath()), name)
}

// RunnerHome is Runner's HOME: its Hugging Face cache (~/.cache/xyntetik-runner/hf) lands here.
func RunnerHome() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "runner-home")
	case "darwin":
		return "/Library/Application Support/GarageAI/runner-home"
	}
	return "/var/lib/garageai-runner"
}

// RunnerArgs is the command line the service runs. --ttl 0 keeps the model loaded: a garage that
// unloads after 5 idle minutes (Runner's default) makes the next buyer wait for the load.
func RunnerArgs(model string, port, context int) []string {
	args := []string{"-hf", model, "--serve", "--port", strconv.Itoa(port), "--ttl", "0"}
	if context > 0 {
		// Runner's default window is 4096 tokens: too small for buyers' coding agents, so Bridge
		// sets one from the machine's memory (garage-m4 served Granite with 4096, 2026-10-18).
		args = append(args, "-c", strconv.Itoa(context))
	}
	return args
}

// InstallRunner registers and starts Runner serving model on port. Must run as root/Administrator.
func InstallRunner(bin, model string, port, context int) error {
	home := RunnerHome()
	if err := os.MkdirAll(home, 0o755); err != nil {
		return err
	}
	args := RunnerArgs(model, port, context)
	switch runtime.GOOS {
	case "linux":
		unit := fmt.Sprintf("[Unit]\nDescription=Xyntetik Runner for GarageAI (installed by garageai runtime install)\nAfter=network-online.target\nWants=network-online.target\n\n[Service]\nType=simple\nEnvironment=HOME=%s\nExecStart=%s %s\nRestart=always\nRestartSec=10\n\n[Install]\nWantedBy=multi-user.target\n",
			home, bin, strings.Join(args, " "))
		if err := os.WriteFile(runnerUnitPath, []byte(unit), 0o644); err != nil {
			return err
		}
		for _, a := range [][]string{{"daemon-reload"}, {"enable", "garageai-runner.service"}, {"restart", "garageai-runner.service"}} {
			if err := run("systemctl", a...); err != nil {
				return err
			}
		}
		return nil
	case "darwin":
		var b strings.Builder
		for _, a := range append([]string{bin}, args...) {
			b.WriteString("<string>" + xmlEscape(a) + "</string>")
		}
		plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>eu.garageai.runner</string>
  <key>ProgramArguments</key><array>%s</array>
  <key>EnvironmentVariables</key><dict><key>HOME</key><string>%s</string></dict>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardOutPath</key><string>%s</string>
  <key>StandardErrorPath</key><string>%s</string>
</dict>
</plist>
`, b.String(), xmlEscape(home), runnerLogPath, runnerLogPath)
		if err := os.WriteFile(runnerPlistPath, []byte(plist), 0o644); err != nil {
			return err
		}
		return launchdReplace("eu.garageai.runner", runnerPlistPath)
	case "windows":
		// A task's command line is limited to 261 characters and a task is not restarted when
		// its program exits: so the task runs a small script that sets Runner's home and starts
		// it again 10 s after it stops, as Restart=always does on Linux.
		dir := filepath.Dir(bin)
		script := filepath.Join(dir, "runner-service.cmd")
		body := fmt.Sprintf("@echo off\r\nrem Written by garageai runtime install. Runs Xyntetik Runner for GarageAI, restarting it if it stops.\r\nset USERPROFILE=%s\r\nset HOME=%s\r\nset LOCALAPPDATA=%s\r\nset XDG_CACHE_HOME=%s\\.cache\r\n:loop\r\n\"%s\" %s >> \"%s\" 2>&1\r\ntimeout /t 10 /nobreak > nul\r\ngoto loop\r\n",
			home, home, home, home, bin, strings.Join(args, " "), filepath.Join(dir, "runner.log"))
		if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
			return err
		}
		_ = rootOnly(script, 0o755)
		_ = run("schtasks", "/End", "/TN", WindowsRunner)
		_ = run("taskkill", "/F", "/IM", "runner.exe")
		if err := run("schtasks", "/Create", "/TN", WindowsRunner, "/SC", "ONSTART", "/RU", "SYSTEM", "/RL", "HIGHEST", "/TR", `"`+script+`"`, "/F"); err != nil {
			return err
		}
		return run("schtasks", "/Run", "/TN", WindowsRunner)
	}
	return fmt.Errorf("Runner as a service is not implemented on %s", runtime.GOOS)
}

// RunnerSettings is how Bridge's Runner service runs: what an upgrade must keep.
type RunnerSettings struct {
	Model   string
	Port    int
	Context int
}

// RunnerConfig reads the model, port and context window from the installed service's command line.
func RunnerConfig() (RunnerSettings, bool) {
	cmdline, ok := runnerCmdline()
	if !ok {
		return RunnerSettings{}, false
	}
	return ParseRunnerCmdline(cmdline)
}

// ParseRunnerCmdline reads -hf, --port and -c from a service definition (systemd unit, launchd
// plist with its <string> tags removed, or the Windows script).
func ParseRunnerCmdline(cmdline string) (RunnerSettings, bool) {
	st := RunnerSettings{Port: 8080}
	f := strings.Fields(strings.NewReplacer(`"`, " ").Replace(cmdline))
	for i := 0; i+1 < len(f); i++ {
		switch f[i] {
		case "-hf":
			st.Model = f[i+1]
		case "--port":
			if n, err := strconv.Atoi(f[i+1]); err == nil {
				st.Port = n
			}
		case "-c":
			if n, err := strconv.Atoi(f[i+1]); err == nil {
				st.Context = n
			}
		}
	}
	return st, st.Model != ""
}

// RunnerVersionPath records which Runner release Bridge installed (written at install and upgrade).
func RunnerVersionPath() string { return RunnerPath() + ".version" }

// InstalledRunnerVersion is the recorded release, else what the binary says (`runner --version`
// prints "runner vX.Y.Z ..."), else "".
func InstalledRunnerVersion() string {
	if b, err := os.ReadFile(RunnerVersionPath()); err == nil {
		if v := strings.TrimSpace(string(b)); v != "" {
			return v
		}
	}
	out, err := exec.Command(RunnerPath(), "--version").Output()
	if err != nil {
		return ""
	}
	return parseRunnerVersion(string(out))
}

// parseRunnerVersion finds the release in `runner --version` output. Runner prints "runner 1.4.2"
// (optionally followed by a build flavour); releases are tagged "v1.4.2", so the result always
// carries the "v".
func parseRunnerVersion(out string) string {
	for _, w := range strings.Fields(out) {
		w = strings.TrimPrefix(strings.TrimRight(w, ",;"), "v")
		parts := strings.Split(w, ".")
		if len(parts) != 3 {
			continue
		}
		ok := true
		for _, p := range parts {
			if p == "" || strings.Trim(p, "0123456789") != "" {
				ok = false
			}
		}
		if ok {
			return "v" + w
		}
	}
	return ""
}

// RecordRunnerVersion writes the installed release.
func RecordRunnerVersion(v string) { _ = os.WriteFile(RunnerVersionPath(), []byte(v+"\n"), 0o644) }

func runnerCmdline() (string, bool) {
	switch runtime.GOOS {
	case "linux":
		b, err := os.ReadFile(runnerUnitPath)
		return string(b), err == nil
	case "darwin":
		b, err := os.ReadFile(runnerPlistPath)
		return strings.NewReplacer("<string>", " ", "</string>", " ").Replace(string(b)), err == nil
	case "windows":
		if err := exec.Command("schtasks", "/Query", "/TN", WindowsRunner).Run(); err != nil {
			return "", false
		}
		b, _ := os.ReadFile(filepath.Join(filepath.Dir(RunnerPath()), "runner-service.cmd"))
		return string(b), true
	}
	return "", false
}

// RunnerInstalled reports whether Bridge's Runner service exists, and the model it serves.
func RunnerInstalled() (installed bool, model string) {
	var cmdline string
	switch runtime.GOOS {
	case "linux":
		b, err := os.ReadFile(runnerUnitPath)
		if err != nil {
			return false, ""
		}
		cmdline = string(b)
	case "darwin":
		b, err := os.ReadFile(runnerPlistPath)
		if err != nil {
			return false, ""
		}
		cmdline = strings.NewReplacer("<string>", " ", "</string>", " ").Replace(string(b))
	case "windows":
		if err := exec.Command("schtasks", "/Query", "/TN", WindowsRunner).Run(); err != nil {
			return false, ""
		}
		b, _ := os.ReadFile(filepath.Join(filepath.Dir(RunnerPath()), "runner-service.cmd"))
		cmdline = string(b)
	default:
		return false, ""
	}
	f := strings.Fields(cmdline)
	for i, w := range f {
		if w == "-hf" && i+1 < len(f) {
			return true, f[i+1]
		}
	}
	return true, ""
}

// RemoveRunner stops and removes the Runner service (the binary and its model cache stay, so a
// reinstall downloads nothing).
func RemoveRunner() {
	switch runtime.GOOS {
	case "linux":
		_ = run("systemctl", "disable", "--now", "garageai-runner.service")
		os.Remove(runnerUnitPath)
		_ = run("systemctl", "daemon-reload")
	case "darwin":
		_ = run("launchctl", "bootout", "system/eu.garageai.runner")
		os.Remove(runnerPlistPath)
	case "windows":
		_ = run("schtasks", "/End", "/TN", WindowsRunner)
		_ = run("schtasks", "/Delete", "/TN", WindowsRunner, "/F")
		_ = run("taskkill", "/F", "/IM", "runner.exe")
	}
}

func xmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

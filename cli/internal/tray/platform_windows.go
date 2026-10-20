//go:build windows

package tray

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/magnusfroste/garageai/cli/internal/service"
)

// Supported: the icon exists on this OS.
const Supported = true

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const runValue = "GarageAI"

// binary is what the login item starts: the user's own garageai.exe (the services' copy in
// ProgramData is for Administrators and SYSTEM only), else this binary.
func binary() string {
	p := filepath.Join(service.WindowsUserBinDir(), "garageai.exe")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return self()
}

func conhost() string {
	return filepath.Join(os.Getenv("SystemRoot"), "System32", "conhost.exe")
}

func LoginItemInstalled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(runValue)
	return err == nil
}

// InstallLoginItem adds the icon to the user's Run key (no Administrator needed); conhost
// --headless starts it without a console window. With start it is started now too, unless this
// process is elevated: an icon started from an Administrator window would run as Administrator,
// so then it appears at the next sign-in.
func InstallLoginItem(start bool) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	cmd := `"` + conhost() + `" --headless "` + binary() + `" tray`
	if err := k.SetStringValue(runValue, cmd); err != nil {
		return err
	}
	if start && !windows.GetCurrentProcessToken().IsElevated() {
		c := exec.Command(conhost(), "--headless", binary(), "tray")
		c.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP}
		return c.Start()
	}
	return nil
}

// RemoveLoginItem deletes the Run value; with stop it also ends a running icon.
func RemoveLoginItem(stop bool) error {
	if k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE); err == nil {
		_ = k.DeleteValue(runValue)
		k.Close()
	}
	if stop {
		_ = exec.Command("taskkill", "/F", "/FI", "WINDOWTITLE eq GarageAI", "/IM", "garageai.exe").Run()
	}
	return nil
}

func openURL(u string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
}

// openDoctor runs `garageai doctor` in a new PowerShell window that stays open.
func openDoctor() error {
	exe := strings.ReplaceAll(self(), "'", "''")
	c := exec.Command("cmd.exe")
	c.SysProcAttr = &syscall.SysProcAttr{CmdLine: `cmd.exe /c start "GarageAI doctor" powershell.exe -NoExit -NoProfile -Command "& '` + exe + `' doctor"`}
	return c.Start()
}

// lock keeps one icon per user session: a named mutex.
func lock() (func(), error) {
	name, _ := windows.UTF16PtrFromString(`Local\GarageAITray`)
	h, err := windows.CreateMutex(nil, false, name)
	if err == windows.ERROR_ALREADY_EXISTS {
		windows.CloseHandle(h)
		return nil, ErrRunning
	}
	if err != nil {
		return nil, err
	}
	return func() { windows.CloseHandle(h) }, nil
}

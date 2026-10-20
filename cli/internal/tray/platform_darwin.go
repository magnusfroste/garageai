//go:build darwin

package tray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/magnusfroste/garageai/cli/internal/service"
)

// Supported: the icon exists on this OS.
const Supported = true

const loginLabel = "eu.garageai.tray"

func loginPlist() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", loginLabel+".plist")
}

// binary is what the login item starts: the root-owned copy the services run when it exists (the
// user cannot replace it), else this binary.
func binary() string {
	if _, err := os.Stat(service.BinaryPath()); err == nil {
		return service.BinaryPath()
	}
	return self()
}

func LoginItemInstalled() bool {
	_, err := os.Stat(loginPlist())
	return err == nil
}

// InstallLoginItem writes a LaunchAgent in the user's own Library (no root), and with start loads
// it now, which starts the icon.
func InstallLoginItem(start bool) error {
	if os.Geteuid() == 0 {
		return fmt.Errorf("the menu-bar icon belongs to your user, not root: run it without sudo")
	}
	p := loginPlist()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	body := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>` + loginLabel + `</string>
  <key>ProgramArguments</key>
  <array><string>` + xmlEscape(binary()) + `</string><string>tray</string></array>
  <key>RunAtLoad</key><true/>
  <key>LimitLoadToSessionType</key><string>Aqua</string>
  <key>ProcessType</key><string>Interactive</string>
</dict>
</plist>
`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		return err
	}
	if start {
		dom := fmt.Sprintf("gui/%d", os.Getuid())
		_ = exec.Command("launchctl", "bootout", dom+"/"+loginLabel).Run()
		if out, err := exec.Command("launchctl", "bootstrap", dom, p).CombinedOutput(); err != nil {
			return fmt.Errorf("launchctl bootstrap: %v %s", err, strings.TrimSpace(string(out)))
		}
	}
	return nil
}

// RemoveLoginItem deletes the LaunchAgent; with stop it also ends a running icon.
func RemoveLoginItem(stop bool) error {
	err := os.Remove(loginPlist())
	if stop {
		_ = exec.Command("launchctl", "bootout", fmt.Sprintf("gui/%d/%s", os.Getuid(), loginLabel)).Run()
	}
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func openURL(u string) error { return exec.Command("/usr/bin/open", u).Start() }

// openDoctor runs `garageai doctor` in a new Terminal window.
func openDoctor() error {
	cmd := "'" + strings.ReplaceAll(self(), "'", `'\''`) + "' doctor"
	as := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(cmd)
	return exec.Command("/usr/bin/osascript", "-e", `tell application "Terminal" to do script "`+as+`"`,
		"-e", `tell application "Terminal" to activate`).Start()
}

// lock keeps one icon per user: a second `garageai tray` exits.
func lock() (func(), error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	dir = filepath.Join(dir, "GarageAI")
	_ = os.MkdirAll(dir, 0o700)
	f, err := os.OpenFile(filepath.Join(dir, "tray.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, ErrRunning
	}
	return func() { f.Close() }, nil
}

func xmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

package service

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// The Bridge tunnel runs as its own long-running service next to the heartbeat (which is a
// 5-minute one-shot): garageai-tunnel.service on Linux, eu.garageai.tunnel on macOS.
const (
	tunnelUnitPath  = "/etc/systemd/system/garageai-tunnel.service"
	tunnelPlistPath = "/Library/LaunchDaemons/eu.garageai.tunnel.plist"
	tunnelLogPath   = "/var/log/garageai-tunnel.log"
)

// TunnelKeyPath holds the garage's WireGuard private key (root, 0600).
func TunnelKeyPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "tunnel.key")
	}
	return "/etc/garageai/tunnel.key"
}

// SaveTunnelKey writes the private key root-only.
func SaveTunnelKey(private string) error {
	p := TunnelKeyPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, []byte(private+"\n"), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// InstallTunnel registers `self tunnel` as a service that is always running. Must run as root.
func InstallTunnel(self string) error {
	switch runtime.GOOS {
	case "linux":
		unit := fmt.Sprintf("[Unit]\nDescription=GarageAI Bridge tunnel (this garage's tunnel to the gateway)\nAfter=network-online.target\nWants=network-online.target\n\n[Service]\nType=simple\nExecStart=%s tunnel\nRestart=always\nRestartSec=5\n\n[Install]\nWantedBy=multi-user.target\n", self)
		if err := os.WriteFile(tunnelUnitPath, []byte(unit), 0o644); err != nil {
			return err
		}
		for _, args := range [][]string{{"daemon-reload"}, {"enable", "garageai-tunnel.service"}, {"restart", "garageai-tunnel.service"}} {
			if err := run("systemctl", args...); err != nil {
				return err
			}
		}
		return nil
	case "darwin":
		plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>eu.garageai.tunnel</string>
  <key>ProgramArguments</key><array><string>%s</string><string>tunnel</string></array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardOutPath</key><string>%s</string>
  <key>StandardErrorPath</key><string>%s</string>
</dict>
</plist>
`, self, tunnelLogPath, tunnelLogPath)
		if err := os.WriteFile(tunnelPlistPath, []byte(plist), 0o644); err != nil {
			return err
		}
		return launchdReplace("eu.garageai.tunnel", tunnelPlistPath)
	case "windows":
		return windowsInstallTunnel(self)
	}
	return fmt.Errorf("the tunnel service is not implemented on %s yet", runtime.GOOS)
}

// TunnelInstalled reports whether the tunnel service exists and runs.
func TunnelInstalled() (installed, active bool) {
	switch runtime.GOOS {
	case "linux":
		_, err := os.Stat(tunnelUnitPath)
		return err == nil, run("systemctl", "is-active", "--quiet", "garageai-tunnel.service") == nil
	case "darwin":
		_, err := os.Stat(tunnelPlistPath)
		return err == nil, run("launchctl", "print", "system/eu.garageai.tunnel") == nil
	case "windows":
		return windowsTunnelInstalled()
	}
	return false, false
}

// RemoveTunnel stops and removes the tunnel service and the key. Must run as root.
func RemoveTunnel() {
	switch runtime.GOOS {
	case "linux":
		_ = run("systemctl", "disable", "--now", "garageai-tunnel.service")
		os.Remove(tunnelUnitPath)
		_ = run("systemctl", "daemon-reload")
	case "darwin":
		_ = run("launchctl", "bootout", "system/eu.garageai.tunnel")
		os.Remove(tunnelPlistPath)
	case "windows":
		windowsRemoveTunnel()
	}
	os.Remove(TunnelKeyPath())
}

// Package service installs `garageai run` as the heartbeat service and keeps its configuration:
// the register token and the runtime key are secrets, so the config file is root-only.
package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config is what `garageai run` needs. Written by connect, read by the service.
type Config struct {
	HeartbeatURL  string `json:"heartbeat_url"`
	RegisterURL   string `json:"register_url"`
	RegisterToken string `json:"register_token"`
	Name          string `json:"name"`
	Runtime       string `json:"runtime"`
	Port          int    `json:"port"`
	RuntimeAPIKey string `json:"runtime_api_key,omitempty"`
	MeshIP        string `json:"mesh_ip,omitempty"`
}

func ConfigPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "bridge.json")
	}
	return "/etc/garageai/bridge.json"
}

func Load() (*Config, error) {
	b, err := os.ReadFile(ConfigPath())
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("%s: %w", ConfigPath(), err)
	}
	return &c, nil
}

// Save writes the config root-only (0600). Must run as root.
func Save(c *Config) error {
	p := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(c, "", "  ")
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

const (
	unitPath   = "/etc/systemd/system/garageai-bridge.service"
	timerPath  = "/etc/systemd/system/garageai-bridge.timer"
	plistPath  = "/Library/LaunchDaemons/eu.garageai.bridge.plist"
	logPath    = "/var/log/garageai-bridge.log"
	LogPathMac = logPath
)

// Install registers the heartbeat service for this OS. Must run as root. self is the path of
// this binary, which the service runs as `self run --once` every 5 minutes.
func Install(self string) error {
	switch runtime.GOOS {
	case "linux":
		unit := fmt.Sprintf("[Unit]\nDescription=GarageAI Bridge heartbeat (reports this garage's models)\nAfter=network-online.target\n\n[Service]\nType=oneshot\nExecStart=%s run --once\n", self)
		timer := "[Unit]\nDescription=GarageAI Bridge heartbeat every 5 minutes\n\n[Timer]\nOnBootSec=1min\nOnUnitActiveSec=5min\nAccuracySec=30s\n\n[Install]\nWantedBy=timers.target\n"
		if err := os.WriteFile(unitPath, []byte(unit), 0o644); err != nil {
			return err
		}
		if err := os.WriteFile(timerPath, []byte(timer), 0o644); err != nil {
			return err
		}
		for _, args := range [][]string{{"daemon-reload"}, {"enable", "--now", "garageai-bridge.timer"}} {
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
  <key>Label</key><string>eu.garageai.bridge</string>
  <key>ProgramArguments</key><array><string>%s</string><string>run</string><string>--once</string></array>
  <key>StartInterval</key><integer>300</integer>
  <key>RunAtLoad</key><true/>
  <key>StandardOutPath</key><string>%s</string>
  <key>StandardErrorPath</key><string>%s</string>
</dict>
</plist>
`, self, logPath, logPath)
		_ = run("launchctl", "bootout", "system/eu.garageai.bridge")
		if err := os.WriteFile(plistPath, []byte(plist), 0o644); err != nil {
			return err
		}
		return run("launchctl", "bootstrap", "system", plistPath)
	}
	return fmt.Errorf("the heartbeat service is not implemented on %s yet", runtime.GOOS)
}

// Installed reports whether the service is registered and running.
func Installed() (installed, active bool) {
	switch runtime.GOOS {
	case "linux":
		_, err := os.Stat(timerPath)
		return err == nil, run("systemctl", "is-active", "--quiet", "garageai-bridge.timer") == nil
	case "darwin":
		_, err := os.Stat(plistPath)
		return err == nil, run("launchctl", "print", "system/eu.garageai.bridge") == nil
	}
	return false, false
}

// Remove stops and removes the service and its config. Must run as root.
func Remove() {
	switch runtime.GOOS {
	case "linux":
		_ = run("systemctl", "disable", "--now", "garageai-bridge.timer")
		os.Remove(unitPath)
		os.Remove(timerPath)
		_ = run("systemctl", "daemon-reload")
	case "darwin":
		_ = run("launchctl", "bootout", "system/eu.garageai.bridge")
		os.Remove(plistPath)
	}
	os.Remove(ConfigPath())
}

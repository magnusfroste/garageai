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
	// The Bridge tunnel (docs/bridge-tunnel.md) instead of NetBird: transport "tunnel", the
	// garage's address in it, the hub's key and endpoint. The private key is in TunnelKeyPath().
	// Self-update: nil means on. garageai connect --no-auto-update sets it to false.
	AutoUpdate   *bool  `json:"auto_update,omitempty"`
	Transport    string `json:"transport,omitempty"`
	HubPublicKey string `json:"hub_public_key,omitempty"`
	HubEndpoint  string `json:"hub_endpoint,omitempty"`
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
	if err := os.Rename(tmp, p); err != nil {
		return err
	}
	return WritePublic(c)
}

// Public is the configuration without secrets: what doctor, run as a normal user, may know.
type Public struct {
	Name        string `json:"name"`
	Runtime     string `json:"runtime"`
	Port        int    `json:"port"`
	Transport   string `json:"transport,omitempty"`
	MeshIP      string `json:"mesh_ip,omitempty"`
	HubEndpoint string `json:"hub_endpoint,omitempty"`
}

func PublicPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "bridge-public.json")
	}
	return "/etc/garageai/bridge-public.json"
}

func (c *Config) Public() Public {
	return Public{Name: c.Name, Runtime: c.Runtime, Port: c.Port, Transport: c.Transport, MeshIP: c.MeshIP, HubEndpoint: c.HubEndpoint}
}

// LoadPublic reads the secret-free configuration (what any user may know about this garage).
func LoadPublic() (*Public, error) {
	b, err := os.ReadFile(PublicPath())
	if err != nil {
		return nil, err
	}
	var p Public
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// WritePublic writes the secret-free configuration world-readable (0644), next to the root-only
// one. bridge.json is 0600, so doctor run by the operator could not see the transport or the port
// and reported NetBird problems on a tunnel garage (dgx-city, 2026-10-12). The heartbeat writes it
// too, so garages set up before this file existed get it on their next run.
func WritePublic(c *Config) error {
	p := PublicPath()
	b, _ := json.MarshalIndent(c.Public(), "", "  ")
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, p); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		_ = windowsRestrict(p, false, true) // the directory is Administrators/SYSTEM only; this file is for everyone
	}
	return nil
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
		if err := os.WriteFile(plistPath, []byte(plist), 0o644); err != nil {
			return err
		}
		return launchdReplace("eu.garageai.bridge", plistPath)
	case "windows":
		return windowsInstallHeartbeat(self)
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
	case "windows":
		return windowsHeartbeatInstalled()
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
	case "windows":
		windowsRemoveHeartbeat()
	}
	RemoveTunnel()
	os.Remove(ConfigPath())
	os.Remove(PublicPath())
	os.Remove(StatusPath())
	if runtime.GOOS != "windows" {
		os.Remove(BinaryPath())
		os.Remove(filepath.Dir(BinaryPath())) // only when empty
	}
}

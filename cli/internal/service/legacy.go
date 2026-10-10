package service

import (
	"os"
	"runtime"
)

// RemoveLegacyScriptHeartbeat removes the heartbeat garageai-connect.sh installed (a shell script
// run by systemd or launchd with its own token file). Bridge reports instead; two heartbeats with
// two tokens would make the old one fail with 401 as soon as a new token is issued. Returns
// whether anything was there. Must run as root.
func RemoveLegacyScriptHeartbeat() bool {
	found := false
	rm := func(p string) {
		if _, err := os.Stat(p); err == nil {
			found = true
			os.Remove(p)
		}
	}
	switch runtime.GOOS {
	case "linux":
		if _, err := os.Stat("/etc/systemd/system/garageai-heartbeat.timer"); err == nil {
			_ = run("systemctl", "disable", "--now", "garageai-heartbeat.timer")
		}
		rm("/etc/systemd/system/garageai-heartbeat.timer")
		rm("/etc/systemd/system/garageai-heartbeat.service")
		if found {
			_ = run("systemctl", "daemon-reload")
		}
	case "darwin":
		if _, err := os.Stat("/Library/LaunchDaemons/eu.garageai.heartbeat.plist"); err == nil {
			_ = run("launchctl", "bootout", "system/eu.garageai.heartbeat")
		}
		rm("/Library/LaunchDaemons/eu.garageai.heartbeat.plist")
	}
	rm("/usr/local/bin/garageai-heartbeat")
	rm("/etc/garageai/heartbeat.env")
	return found
}

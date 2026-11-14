package service

import (
	"runtime"
	"strings"
	"testing"
)

// The product was renamed from Bridge to Door in v1.6.0. The names installed garages, the
// installers and the portal find things by did not change, and must not: a new name here strands
// every garage that is already connected.
func TestInstalledNamesStay(t *testing.T) {
	for got, want := range map[string]string{
		WindowsTask:    "GarageAI Bridge heartbeat",
		WindowsService: "GarageAIBridgeTunnel",
		unitPath:       "/etc/systemd/system/garageai-bridge.service",
		timerPath:      "/etc/systemd/system/garageai-bridge.timer",
		plistPath:      "/Library/LaunchDaemons/eu.garageai.bridge.plist",
		logPath:        "/var/log/garageai-bridge.log",
	} {
		if got != want {
			t.Errorf("installed name changed: %q, want %q", got, want)
		}
	}
	if runtime.GOOS != "windows" && !strings.HasSuffix(ConfigPath(), "/garageai/bridge.json") {
		t.Errorf("config path changed: %s", ConfigPath())
	}
}

package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/tunnel"
)

// TestWindowsServicesRoundTrip registers the heartbeat task and the tunnel service with the real
// schtasks/sc, checks they exist and run, and removes them. Runs in CI on windows-latest only
// (an Administrator, a throwaway machine): GARAGEAI_WINDOWS_SERVICE_TEST=1 and GARAGEAI_TEST_BINARY
// pointing at a built garageai.exe.
func TestWindowsServicesRoundTrip(t *testing.T) {
	if runtime.GOOS != "windows" || os.Getenv("GARAGEAI_WINDOWS_SERVICE_TEST") != "1" {
		t.Skip("Windows service round trip runs in CI on windows-latest only")
	}
	exe := os.Getenv("GARAGEAI_TEST_BINARY")
	if _, err := os.Stat(exe); err != nil {
		t.Fatalf("GARAGEAI_TEST_BINARY: %v", err)
	}
	// A config the tunnel service can start with (the hub is unreachable: the loop retries, the
	// service stays running).
	priv, _, _ := tunnel.GenerateKey()
	dir := filepath.Dir(ConfigPath())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := Config{Name: "ci-windows", Runtime: "ollama", Port: 11434, Transport: "tunnel", MeshIP: "10.66.0.250",
		HubPublicKey: "y2tVj41DZ7Db+C6djttop0A4Z3I0NbpnZ70Y1AJ1/R0=", HubEndpoint: "127.0.0.1:51830", RegisterURL: "https://127.0.0.1:1/register-node", RegisterToken: "ci"}
	if err := Save(&cfg); err != nil {
		t.Fatal(err)
	}
	if err := SaveTunnelKey(priv); err != nil {
		t.Fatal(err)
	}
	defer func() { Remove(); os.RemoveAll(dir) }()

	if err := Install(exe); err != nil {
		t.Fatalf("install heartbeat task: %v", err)
	}
	if inst, _ := Installed(); !inst {
		t.Fatal("heartbeat task not registered")
	}
	if err := InstallTunnel(exe); err != nil {
		t.Fatalf("install tunnel service: %v", err)
	}
	deadline := time.Now().Add(20 * time.Second)
	for {
		inst, active := TunnelInstalled()
		if inst && active {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("tunnel service installed=%v running=%v after 20 s", inst, active)
		}
		time.Sleep(time.Second)
	}
	RestartTunnel()
	if inst, active := TunnelInstalled(); !inst || !active {
		t.Fatalf("after restart: installed=%v running=%v", inst, active)
	}
	// The public config exists and the service log is being written.
	if _, err := os.Stat(PublicPath()); err != nil {
		t.Errorf("public config: %v", err)
	}
	if b, err := os.ReadFile(WindowsLogPath()); err != nil || len(b) == 0 {
		t.Errorf("tunnel log: %v (%d bytes)", err, len(b))
	}
	Remove()
	if inst, _ := TunnelInstalled(); inst {
		t.Fatal("tunnel service still registered after Remove")
	}
	if inst, _ := Installed(); inst {
		t.Fatal("heartbeat task still registered after Remove")
	}
	var p Public
	_ = json.Unmarshal(nil, &p)
}

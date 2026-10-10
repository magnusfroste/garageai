// Package netbird drives the NetBird client the way the connect script does.
package netbird

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/profile"
)

func Installed() bool { _, err := exec.LookPath("netbird"); return err == nil }

func Version() string {
	out, _ := exec.Command("netbird", "version").Output()
	return strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
}

// Install runs NetBird's installer (CLI only, no desktop app), like the script.
func Install() error {
	cmd := exec.Command("sh", "-c", "curl -fsSL https://pkgs.netbird.io/install.sh | SKIP_UI_APP=true sh")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if !Installed() {
		return fmt.Errorf("the installer did not put 'netbird' on PATH")
	}
	return nil
}

// MeshIP is this machine's NetBird address: from the service when it answers, otherwise from the
// tunnel interface (a normal user may not be allowed to ask the service on macOS).
func MeshIP() string {
	var st struct {
		NetbirdIP string `json:"netbirdIp"`
	}
	if out, err := exec.Command("netbird", "status", "--json").Output(); err == nil && json.Unmarshal(out, &st) == nil && st.NetbirdIP != "" {
		return strings.SplitN(st.NetbirdIP, "/", 2)[0]
	}
	if out, err := exec.Command("netbird", "status").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if k, v, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(k) == "NetBird IP" {
				return strings.SplitN(strings.TrimSpace(v), "/", 2)[0]
			}
		}
	}
	return profile.MeshIPFromInterfaces()
}

// Up joins the mesh with a setup key. The key goes in the environment, not on the command line,
// so it does not show up in the process list. Needs root: run through sudo when we are not.
func Up(setupKey, managementURL, hostname string, sudo func(name string, args ...string) *exec.Cmd) error {
	cmd := sudo("env", "NB_SETUP_KEY="+setupKey, "netbird", "up", "--management-url", managementURL, "--hostname", hostname)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// WaitForMeshIP polls until the client has an address, at most wait.
func WaitForMeshIP(wait time.Duration) string {
	deadline := time.Now().Add(wait)
	for {
		if ip := MeshIP(); ip != "" {
			return ip
		}
		if time.Now().After(deadline) {
			return ""
		}
		time.Sleep(time.Second)
	}
}

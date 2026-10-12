//go:build !windows

package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnsafeProgram(t *testing.T) {
	// A binary in a normal user's directory (the operator's ~/.local/bin) is unsafe for root.
	if os.Geteuid() != 0 {
		p := filepath.Join(t.TempDir(), "garageai")
		os.WriteFile(p, []byte("x"), 0o755)
		if why := UnsafeProgram(p); !strings.Contains(why, "normal user") {
			t.Fatalf("user-owned binary: %q", why)
		}
	}
	// A root-owned system binary is safe.
	if why := UnsafeProgram("/bin/sh"); why != "" {
		t.Fatalf("/bin/sh: %q", why)
	}
}

func TestProgramFromPlist(t *testing.T) {
	plist := "<dict>\n  <key>ProgramArguments</key><array><string>/Users/mafr/.local/bin/garageai</string><string>run</string>"
	if m := plistProgram.FindStringSubmatch(plist); m == nil || m[1] != "/Users/mafr/.local/bin/garageai" {
		t.Fatalf("plist: %v", m)
	}
	unit := "[Service]\nType=oneshot\nExecStart=/usr/local/lib/garageai/garageai run --once\n"
	if m := unitProgram.FindStringSubmatch(unit); m == nil || m[1] != "/usr/local/lib/garageai/garageai" {
		t.Fatalf("unit: %v", m)
	}
}

func TestPublicHasNoSecrets(t *testing.T) {
	c := &Config{Name: "g", Runtime: "vllm", Port: 8000, RegisterToken: "grg_secret", RuntimeAPIKey: "sk-secret", Transport: "tunnel", MeshIP: "10.66.0.6", HubPublicKey: "hubpub", HubEndpoint: "api.garageai.eu:51830"}
	b, _ := json.Marshal(c.Public())
	s := string(b)
	for _, secret := range []string{"grg_secret", "sk-secret", "hubpub"} {
		if strings.Contains(s, secret) {
			t.Fatalf("public config leaks %q: %s", secret, s)
		}
	}
	for _, want := range []string{`"transport":"tunnel"`, `"mesh_ip":"10.66.0.6"`, `"port":8000`} {
		if !strings.Contains(s, want) {
			t.Fatalf("public config lacks %s: %s", want, s)
		}
	}
}

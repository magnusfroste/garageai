//go:build !windows

package service

import (
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

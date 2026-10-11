// Package machine gives this computer a stable id, so the portal can tell when a garage suddenly
// reports from another machine (one garage name reused on a second laptop). The OS's own id is
// hashed: the raw value never leaves the machine.
package machine

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var (
	ioregUUID = regexp.MustCompile(`"IOPlatformUUID"\s*=\s*"([^"]+)"`)
	regGUID   = regexp.MustCompile(`MachineGuid\s+REG_SZ\s+(\S+)`)
)

// ID is "m_" and 32 hex characters, the same for every run on this machine; "" when the OS gives
// no id.
func ID() string {
	raw := rawID()
	if raw == "" {
		return ""
	}
	sum := sha256.Sum256([]byte("garageai-machine:" + raw))
	return "m_" + hex.EncodeToString(sum[:16])
}

func rawID() string {
	switch runtime.GOOS {
	case "darwin":
		out, _ := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
		if m := ioregUUID.FindSubmatch(out); m != nil {
			return string(m[1])
		}
	case "windows":
		out, _ := exec.Command("reg", "query", `HKLM\SOFTWARE\Microsoft\Cryptography`, "/v", "MachineGuid").Output()
		if m := regGUID.FindSubmatch(out); m != nil {
			return string(m[1])
		}
	default:
		for _, p := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
			if b, err := os.ReadFile(p); err == nil && strings.TrimSpace(string(b)) != "" {
				return strings.TrimSpace(string(b))
			}
		}
	}
	return ""
}

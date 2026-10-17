package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"
)

// BinaryPath is where the heartbeat service's own copy of Bridge lives. The service runs as root,
// so the file and its directory must be root's alone: a copy in the operator's home (where the
// installer puts the command they type) could be replaced by anything running as that user.
// Not /usr/local/bin: with Homebrew on an Intel Mac that directory belongs to the user.
func BinaryPath() string {
	if p := os.Getenv("GARAGEAI_SERVICE_BINARY"); p != "" && os.Geteuid() != 0 {
		return p // tests and dry runs; never as root, where it would become the service's ExecStart
	}
	switch runtime.GOOS {
	case "darwin":
		return "/Library/Application Support/GarageAI/garageai"
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "garageai.exe")
	}
	return "/usr/local/lib/garageai/garageai"
}

// InstallBinary copies src to BinaryPath, root-owned and not writable by others. Must run as root.
func InstallBinary(src string) (string, error) {
	dest := BinaryPath()
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := rootOnly(dir, 0o755); err != nil {
		return "", err
	}
	b, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, b, 0o755); err != nil {
		return "", err
	}
	if err := rootOnly(tmp, 0o755); err != nil {
		os.Remove(tmp)
		return "", err
	}
	if err := replaceFile(tmp, dest); err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("replace %s: %w", dest, err)
	}
	return dest, nil
}

// replaceFile moves tmp onto dest. On Windows a running exe cannot be overwritten (the tunnel
// service runs this very file) but it can be renamed, and an antivirus scan of a freshly written
// exe holds it for a moment: so the current file steps aside as .old (removed on the next run)
// and the move is retried for a few seconds. Elsewhere a rename replaces atomically.
func replaceFile(tmp, dest string) error {
	if runtime.GOOS != "windows" {
		return os.Rename(tmp, dest)
	}
	old := dest + ".old"
	var err error
	for i := 0; i < 20; i++ {
		if err = os.Rename(tmp, dest); err == nil {
			return nil
		}
		if _, statErr := os.Stat(dest); statErr == nil {
			_ = os.Remove(old)
			_ = os.Rename(dest, old)
		}
		time.Sleep(500 * time.Millisecond)
	}
	if _, statErr := os.Stat(dest); statErr != nil {
		_ = os.Rename(old, dest) // put the previous copy back rather than leave nothing
	}
	return err
}

var (
	plistProgram = regexp.MustCompile(`<key>ProgramArguments</key>\s*<array>\s*<string>([^<]+)</string>`)
	unitProgram  = regexp.MustCompile(`(?m)^ExecStart=(\S+)`)
)

// Program is the binary the installed heartbeat service runs, "" when there is none.
func Program() string {
	switch runtime.GOOS {
	case "darwin":
		if b, err := os.ReadFile(plistPath); err == nil {
			if m := plistProgram.FindSubmatch(b); m != nil {
				return string(m[1])
			}
		}
	case "linux":
		if b, err := os.ReadFile(unitPath); err == nil {
			if m := unitProgram.FindSubmatch(b); m != nil {
				return string(m[1])
			}
		}
	}
	return ""
}

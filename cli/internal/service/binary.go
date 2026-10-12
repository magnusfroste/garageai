package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

// BinaryPath is where the heartbeat service's own copy of Bridge lives. The service runs as root,
// so the file and its directory must be root's alone: a copy in the operator's home (where the
// installer puts the command they type) could be replaced by anything running as that user.
// Not /usr/local/bin: with Homebrew on an Intel Mac that directory belongs to the user.
func BinaryPath() string {
	if p := os.Getenv("GARAGEAI_SERVICE_BINARY"); p != "" {
		return p // tests and dry runs
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
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("replace %s: %w", dest, err)
	}
	return dest, nil
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

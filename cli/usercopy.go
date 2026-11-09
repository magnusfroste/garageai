package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magnusfroste/garageai/cli/internal/service"
)

// The operator's own garageai (what they type: %LOCALAPPDATA%\Programs\GarageAI on Windows,
// ~/.local/bin or Homebrew's /usr/local/bin elsewhere) is a second copy next to the service's.
// Self-update replaces the service's copy only (root never writes in a user's directory), so the
// operator's copy could stay versions behind: on garage-w11 `garageai runtime` was v1.1.0's
// "unknown command" while its service ran v1.5.0 (2026-11-09). An own copy older than the
// service's now brings itself up to it, and runs the service's copy for this command.

// delegatedEnv marks a process started by an older copy, so it never hands over again.
const delegatedEnv = "GARAGEAI_DELEGATED"

// ownCopyCommands are the commands an operator types; the services' own (run, tunnel, the
// elevated halves) and the ones about this binary itself (version, update, install-service) are not.
var ownCopyCommands = map[string]bool{
	"connect": true, "login": true, "doctor": true, "uninstall": true, "mcp": true, "models": true,
	"runtime": true, "pause": true, "resume": true, "tray": true, "help": true, "--help": true, "-h": true,
}

// useServiceCopy runs the command with the service's copy when this copy is older, after trying to
// update this copy to it. ok is false when this copy should go on itself.
func useServiceCopy(args []string) (code int, ok bool) {
	if len(args) == 0 || !ownCopyCommands[args[0]] || os.Getenv(delegatedEnv) != "" {
		return 0, false
	}
	self, err := os.Executable()
	if err != nil {
		return 0, false
	}
	if s, err := filepath.EvalSymlinks(self); err == nil {
		self = s
	}
	os.Remove(self + ".old") // Windows: the copy that stepped aside the last time this one was updated
	svc := service.BinaryPath()
	if sameFile(self, svc) || service.UnsafeProgram(svc) != "" {
		return 0, false
	}
	have := versionOf(svc)
	if !newerVersion(version, have) {
		return 0, false
	}
	if err := refreshCopy(self, svc); err == nil {
		fmt.Fprintf(os.Stderr, "garageai: updated %s from %s to %s, the version the service runs.\n", self, version, have)
	} else {
		fmt.Fprintf(os.Stderr, "garageai: %s is %s and the service runs %s: using the service's copy (%v).\n", self, version, have, err)
	}
	cmd := exec.Command(svc, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = append(os.Environ(), delegatedEnv+"=1")
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode(), true
		}
		return 0, false // the service's copy did not start: this copy does the command after all
	}
	return 0, true
}

// versionOf asks a Bridge binary its version ("GarageAI Bridge v1.5.0" → "v1.5.0"), "" when it
// does not answer.
func versionOf(bin string) string {
	out, err := exec.Command(bin, "version").Output()
	if err != nil {
		return ""
	}
	f := strings.Fields(string(out))
	if len(f) == 0 {
		return ""
	}
	return f[len(f)-1]
}

func sameFile(a, b string) bool {
	fa, err1 := os.Stat(a)
	fb, err2 := os.Stat(b)
	return err1 == nil && err2 == nil && os.SameFile(fa, fb)
}

// refreshCopy replaces the operator's copy dest with src. Only as the operator (writeOwnCopyRefusal):
// as root, writing into a user's directory would let that user redirect the write (a symlink)
// anywhere on the machine. On Windows the running exe steps aside as .old, which the next run removes.
func refreshCopy(dest, src string) error {
	if why := writeOwnCopyRefusal(dest); why != "" {
		return errors.New(why)
	}
	fi, err := os.Lstat(dest)
	if err != nil || !fi.Mode().IsRegular() {
		return errors.New("not a regular file")
	}
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	tmp := dest + ".new"
	os.Remove(tmp)
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL|noFollow, 0o755)
	if err != nil {
		return err
	}
	_, werr := f.Write(b)
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		os.Remove(tmp)
		return werr
	}
	if runtime.GOOS == "windows" {
		old := dest + ".old"
		os.Remove(old)
		if err := os.Rename(dest, old); err != nil {
			os.Remove(tmp)
			return err
		}
		if err := os.Rename(tmp, dest); err != nil {
			_ = os.Rename(old, dest)
			os.Remove(tmp)
			return err
		}
		return nil
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// refreshUserCopyCmd is `garageai __refresh-user-copy`: the service's copy, started by the
// heartbeat as a signed-in Windows user, brings that user's own garageai up to itself.
func refreshUserCopyCmd() int {
	dest := filepath.Join(service.WindowsUserBinDir(), "garageai.exe")
	if _, err := os.Stat(dest); err != nil {
		return 0 // this user has no copy of their own
	}
	if !newerVersion(versionOf(dest), version) {
		return 0
	}
	self, err := os.Executable()
	if err != nil {
		return 1
	}
	if err := refreshCopy(dest, self); err != nil {
		fmt.Fprintln(os.Stderr, "refresh", dest+":", err)
		return 1
	}
	fmt.Println("updated", dest, "to", version)
	return 0
}

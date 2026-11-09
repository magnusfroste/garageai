//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/magnusfroste/garageai/cli/internal/service"
)

// refreshUserCopies: the heartbeat (SYSTEM) has each signed-in user bring their own garageai up to
// the service's version, by running the service's copy as that user (their token and
// environment), so SYSTEM itself never writes in a user's profile. Copies from before this
// version cannot do it themselves. Once per version, once some user was signed in.
func refreshUserCopies() {
	mark := filepath.Join(filepath.Dir(service.BinaryPath()), "user-copies")
	if b, err := os.ReadFile(mark); err == nil && strings.TrimSpace(string(b)) == version {
		return
	}
	if err := service.AllowUsersRun(service.BinaryPath()); err != nil {
		return
	}
	var list *windows.WTS_SESSION_INFO
	var n uint32
	if err := windows.WTSEnumerateSessions(0, 0, 1, &list, &n); err != nil {
		return
	}
	defer windows.WTSFreeMemory(uintptr(unsafe.Pointer(list)))
	reached := false
	for _, s := range unsafe.Slice(list, n) {
		if s.State != windows.WTSActive && s.State != windows.WTSDisconnected {
			continue
		}
		var tok windows.Token
		if windows.WTSQueryUserToken(s.SessionID, &tok) != nil {
			continue
		}
		if env, err := tok.Environ(false); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			cmd := exec.CommandContext(ctx, service.BinaryPath(), "__refresh-user-copy")
			cmd.Env = env
			cmd.SysProcAttr = &syscall.SysProcAttr{Token: syscall.Token(tok), HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
			if out, err := cmd.CombinedOutput(); err == nil {
				reached = true
				if s := strings.TrimSpace(string(out)); s != "" {
					fmt.Println(time.Now().UTC().Format(time.RFC3339), s)
				}
			}
			cancel()
		}
		tok.Close()
	}
	if reached {
		_ = os.WriteFile(mark, []byte(version), 0o644)
	}
}

// writeOwnCopyRefusal on Windows: the copy must be under this account's own LOCALAPPDATA, with no
// junction or symlink on the way, and the writer must not be SYSTEM. Elevated is fine: on a
// machine with UAC off or the built-in Administrator every window is elevated (garage-w11,
// 2026-11-09), and an account cannot gain anything by redirecting its own writes.
func writeOwnCopyRefusal(dest string) string {
	if u, err := windows.GetCurrentProcessToken().GetTokenUser(); err == nil && u.User.Sid.IsWellKnown(windows.WinLocalSystemSid) {
		return "not as SYSTEM"
	}
	base := os.Getenv("LOCALAPPDATA")
	rel, err := filepath.Rel(base, dest)
	if base == "" || err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return "not in this account's own LOCALAPPDATA"
	}
	for p := filepath.Dir(dest); ; p = filepath.Dir(p) {
		fi, err := os.Lstat(p)
		if err != nil || !fi.IsDir() || fi.Mode()&(os.ModeSymlink|os.ModeIrregular) != 0 {
			return p + " is a link or junction"
		}
		if len(p) <= len(base) {
			return ""
		}
	}
}

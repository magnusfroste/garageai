//go:build !windows

package service

import (
	"os"
	"path/filepath"
	"syscall"
)

func rootOnly(path string, mode os.FileMode) error {
	if err := os.Chown(path, 0, 0); err != nil {
		return err
	}
	return os.Chmod(path, mode)
}

// UnsafeProgram says why a root service running path could be made to run someone else's code:
// the file or a directory above it is not root's, or others can write to it. "" when it is safe.
func UnsafeProgram(path string) string {
	for p := path; ; p = filepath.Dir(p) {
		fi, err := os.Stat(p)
		if err != nil {
			return ""
		}
		st, ok := fi.Sys().(*syscall.Stat_t)
		if ok && st.Uid != 0 {
			return p + " belongs to a normal user account"
		}
		// Sticky directories (/tmp) are writable by all but not replaceable; anything else is.
		// Group write is fine for wheel (0) and macOS's admin (80): their members can sudo anyway.
		perm := fi.Mode().Perm()
		groupOK := ok && (st.Gid == 0 || st.Gid == 80)
		if fi.Mode()&os.ModeSticky == 0 && (perm&0o002 != 0 || (perm&0o020 != 0 && !groupOK)) {
			return p + " can be changed by other users"
		}
		if p == "/" || p == "." {
			return ""
		}
	}
}

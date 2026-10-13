//go:build !windows

package main

import (
	"os"
	"syscall"
)

const noFollow = syscall.O_NOFOLLOW

// rootOwnedDir: the directory belongs to root and nobody else can write in it. What root may
// trust on PATH, and where it may write.
func rootOwnedDir(d string) bool {
	fi, err := os.Lstat(d)
	if err != nil || !fi.IsDir() {
		return false
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	return ok && st.Uid == 0 && fi.Mode().Perm()&0o022 == 0
}

// rootOwnedRegular: a regular file (not a symlink) owned by root.
func rootOwnedRegular(p string) bool {
	fi, err := os.Lstat(p)
	if err != nil || !fi.Mode().IsRegular() {
		return false
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	return ok && st.Uid == 0
}

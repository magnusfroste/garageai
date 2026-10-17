package service

import "os"

// rootOnly on Windows: Administrators and SYSTEM only. A directory passes the rule on to what is
// created inside it; a file gets plain grants (inheritance flags on a file leave it with no usable
// entries at all, not even the right to rename it: garage-w11, 2026-10-17).
func rootOnly(path string, _ os.FileMode) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	return windowsRestrict(path, fi.IsDir(), false)
}

// UnsafeProgram: not checked on Windows yet (the directory is restricted at install).
func UnsafeProgram(string) string { return "" }

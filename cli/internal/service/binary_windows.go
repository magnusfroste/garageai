package service

import "os"

// rootOnly on Windows: Administrators and SYSTEM only, inherited by what is created inside.
func rootOnly(dir string, _ os.FileMode) error { return windowsRestrict(dir, true, false) }

// UnsafeProgram: not checked on Windows yet (the directory is restricted at install).
func UnsafeProgram(string) string { return "" }

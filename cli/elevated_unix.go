//go:build !windows

package main

import "os"

func isAdmin() bool { return os.Geteuid() == 0 }

const adminHint = "run it with sudo"

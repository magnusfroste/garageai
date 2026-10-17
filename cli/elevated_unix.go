//go:build !windows

package main

import "os"

func isAdmin() bool { return os.Geteuid() == 0 }

const adminHint = "run it with sudo"

// elevate and holdWindow exist for Windows (UAC); elsewhere sudo does this.
func elevate(cmd string, args []string) int { return 1 }

func holdWindow(args []string) int { return 1 }

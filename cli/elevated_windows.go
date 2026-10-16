//go:build windows

package main

import "golang.org/x/sys/windows"

// isAdmin: an elevated process (PowerShell "Run as administrator"). Installing the services and
// writing under %ProgramData%\GarageAI need it; there is no sudo to ask for it.
func isAdmin() bool { return windows.GetCurrentProcessToken().IsElevated() }

const adminHint = "Open PowerShell as Administrator (right-click → Run as administrator) and run this again"

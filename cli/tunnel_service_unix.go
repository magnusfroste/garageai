//go:build !windows

package main

import "context"

// On macOS and Linux the tunnel is run by launchd/systemd as a plain process.
func runAsWindowsService(func(ctx context.Context) int) (handled bool, code int) { return false, 0 }

//go:build windows

package main

import (
	"context"
	"os"

	"github.com/magnusfroste/garageai/cli/internal/service"
	"golang.org/x/sys/windows/svc"
)

// runAsWindowsService runs the tunnel under the service control manager when the process was
// started by it (sc start): the SCM wants a service handler that answers Stop/Shutdown, or it
// kills the process as unresponsive. Started from a terminal it reports false and the caller
// runs the plain loop. Output goes to the log file: a service has no console.
func runAsWindowsService(loop func(ctx context.Context) int) (handled bool, code int) {
	isSvc, err := svc.IsWindowsService()
	if err != nil || !isSvc {
		return false, 0
	}
	if f, err := os.OpenFile(service.WindowsLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
		os.Stdout, os.Stderr = f, f
	}
	h := &tunnelHandler{loop: loop}
	if err := svc.Run(service.WindowsService, h); err != nil {
		return true, 1
	}
	return true, h.code
}

type tunnelHandler struct {
	loop func(ctx context.Context) int
	code int
}

func (h *tunnelHandler) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	const accepted = svc.AcceptStop | svc.AcceptShutdown
	s <- svc.Status{State: svc.StartPending}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	go func() { done <- h.loop(ctx) }()
	s <- svc.Status{State: svc.Running, Accepts: accepted}
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				s <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s <- svc.Status{State: svc.StopPending}
				cancel()
				h.code = <-done
				return false, 0
			}
		case h.code = <-done:
			cancel()
			return false, uint32(h.code)
		}
	}
}

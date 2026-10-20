package main

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/service"
	"github.com/magnusfroste/garageai/cli/internal/tunnel"
)

// tunnelCmd is `garageai tunnel`: this garage's tunnel to the gateway, kept up until stopped. The
// installed service runs it. It needs the config connect wrote (transport "tunnel") and the key.
func tunnelCmd() int {
	cfg, err := service.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "garageai tunnel: no configuration:", err, "(run garageai connect --tunnel bridge)")
		return 2
	}
	if cfg.Transport != "tunnel" || cfg.MeshIP == "" || cfg.HubPublicKey == "" || cfg.HubEndpoint == "" {
		fmt.Fprintln(os.Stderr, "garageai tunnel: this garage is not set up for the Bridge tunnel (run garageai connect --tunnel bridge)")
		return 2
	}
	key, err := os.ReadFile(service.TunnelKeyPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "garageai tunnel: no key:", err)
		return 2
	}
	addr, err := netip.ParseAddr(cfg.MeshIP)
	if err != nil {
		fmt.Fprintln(os.Stderr, "garageai tunnel: bad tunnel address:", cfg.MeshIP)
		return 2
	}
	loop := func(ctx context.Context) int {
		return tunnelLoop(ctx, cfg, strings.TrimSpace(string(key)), addr)
	}
	if handled, code := runAsWindowsService(loop); handled {
		return code
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return loop(ctx)
}

// tunnelLoop keeps the tunnel up until ctx ends: starts it, watches the network, reconnects, and
// restarts it when the hub has been silent too long.
func tunnelLoop(ctx context.Context, cfg *service.Config, key string, addr netip.Addr) int {
	logf := func(format string, args ...any) {
		fmt.Printf("%s %s\n", time.Now().UTC().Format(time.RFC3339), fmt.Sprintf(format, args...))
	}
	tunnel.MakeReadable = service.MakeUserReadable
	go tunnel.FlushStatsEvery(30*time.Second, ctx.Done()) // the garage's own request timings, for the heartbeat
	go watchState(ctx, logf)                              // a pause or a manager's state change: heartbeat at once
	for ctx.Err() == nil {
		t, err := tunnel.Start(tunnel.Config{PrivateKey: key, HubPublicKey: cfg.HubPublicKey, HubEndpoint: cfg.HubEndpoint,
			Address: addr, Forwards: map[uint16]string{uint16(cfg.Port): fmt.Sprintf("127.0.0.1:%d", cfg.Port)}, Logf: logf})
		if err != nil {
			logf("could not start the tunnel: %v; retrying in 10 s", err)
			select {
			case <-ctx.Done():
				return 0
			case <-time.After(10 * time.Second):
			}
			continue
		}
		logf("tunnel up: %s via %s, Bridge %s", addr, cfg.HubEndpoint, version)
		go t.Watch(ctx)
		// Report the state now and then, and restart the tunnel if the hub has been silent too long
		// (the hub's key or endpoint changed, or something is wedged that a rebind did not fix).
		up := false
		tick := time.NewTicker(30 * time.Second)
		for ctx.Err() == nil {
			select {
			case <-ctx.Done():
			case <-tick.C:
			}
			if ctx.Err() != nil {
				break
			}
			hs := t.LastHandshake()
			if !hs.IsZero() && time.Since(hs) < 3*time.Minute {
				if !up {
					logf("connected to the hub (handshake %s ago)", time.Since(hs).Round(time.Second))
					up = true
				}
			} else if up || time.Since(hs) > 5*time.Minute {
				logf("no handshake with the hub for %s; reconnecting", time.Since(hs).Round(time.Second))
				up = false
				_ = t.Rebind()
				if time.Since(hs) > 10*time.Minute {
					break // a full restart of the tunnel
				}
			}
		}
		tick.Stop()
		t.Close()
	}
	logf("tunnel stopped")
	return 0
}

// watchState polls the lifecycle state every 10 s and runs one heartbeat when it changes, so the
// portal drains the garage before a redeploy stops the runtime (vLLM Panel waits ≥ 15 s).
func watchState(ctx context.Context, logf func(string, ...any)) {
	last := stateKey(time.Now())
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if k := stateKey(time.Now()); k != last {
				logf("state %s → %s: heartbeat now", last, k)
				last = k
				cmd := exec.CommandContext(ctx, service.BinaryPath(), "run", "--once")
				cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
				_ = cmd.Run()
			}
		}
	}
}

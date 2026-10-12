// Package tunnel is Bridge's own tunnel to the GarageAI gateway (docs/bridge-tunnel.md): WireGuard
// in user space (wireguard-go and the gVisor netstack). No network interface, no routes, no root.
// The garage dials out to the hub, so any NAT works, and the runtime stays on 127.0.0.1: Bridge
// listens on its tunnel address inside its own stack and forwards to the runtime.
package tunnel

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

// HubAddress is the gateway's address in the tunnel network (10.66.0.0/16).
var HubAddress = netip.MustParseAddr("10.66.0.1")

type Config struct {
	PrivateKey   string            // the garage's key, base64
	HubPublicKey string            // base64
	HubEndpoint  string            // host:port of the hub's UDP port (resolved at start)
	Address      netip.Addr        // this garage's tunnel address
	Forwards     map[uint16]string // tunnel port -> local address (e.g. 11434 -> 127.0.0.1:11434)
	Keepalive    int               // seconds; 0 means 25
	WatchEvery   time.Duration     // how often to look for network changes; 0 means 2 s
	Logf         func(format string, args ...any)
}

// GenerateKey returns a new key pair, base64 like WireGuard's own tools.
func GenerateKey() (private, public string, err error) {
	k, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(k.Bytes()), base64.StdEncoding.EncodeToString(k.PublicKey().Bytes()), nil
}

// PublicKey derives the public key from a private key.
func PublicKey(private string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(private))
	if err != nil || len(b) != 32 {
		return "", errors.New("not a WireGuard private key")
	}
	k, err := ecdh.X25519().NewPrivateKey(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(k.PublicKey().Bytes()), nil
}

func hexKey(b64 string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
	if err != nil || len(b) != 32 {
		return "", errors.New("not a WireGuard key")
	}
	return hex.EncodeToString(b), nil
}

// Tunnel is a running tunnel.
type Tunnel struct {
	cfg  Config
	dev  *device.Device
	tnet *netstack.Net
	hub  device.NoisePublicKey
	mu   sync.Mutex
	lns  []net.Listener
}

// Start brings the tunnel up and starts forwarding. Close it to stop.
func Start(cfg Config) (*Tunnel, error) {
	if cfg.Keepalive == 0 {
		cfg.Keepalive = 25
	}
	if cfg.WatchEvery == 0 {
		cfg.WatchEvery = 2 * time.Second
	}
	if cfg.Logf == nil {
		cfg.Logf = func(string, ...any) {}
	}
	priv, err := hexKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("private key: %w", err)
	}
	hubHex, err := hexKey(cfg.HubPublicKey)
	if err != nil {
		return nil, fmt.Errorf("hub key: %w", err)
	}
	ep, err := net.ResolveUDPAddr("udp", cfg.HubEndpoint)
	if err != nil {
		return nil, fmt.Errorf("hub endpoint %s: %w", cfg.HubEndpoint, err)
	}
	tun, tnet, err := netstack.CreateNetTUN([]netip.Addr{cfg.Address}, nil, 1420)
	if err != nil {
		return nil, err
	}
	dev := device.NewDevice(tun, conn.NewDefaultBind(), device.NewLogger(device.LogLevelSilent, ""))
	uapi := fmt.Sprintf("private_key=%s\npublic_key=%s\nendpoint=%s\nallowed_ip=%s/32\npersistent_keepalive_interval=%d\n",
		priv, hubHex, ep.String(), HubAddress, cfg.Keepalive)
	if err := dev.IpcSet(uapi); err != nil {
		dev.Close()
		return nil, err
	}
	if err := dev.Up(); err != nil {
		dev.Close()
		return nil, err
	}
	t := &Tunnel{cfg: cfg, dev: dev, tnet: tnet}
	_ = t.hub.FromHex(hubHex)
	ports := make([]int, 0, len(cfg.Forwards))
	for p := range cfg.Forwards {
		ports = append(ports, int(p))
	}
	sort.Ints(ports)
	for _, p := range ports {
		ln, err := tnet.ListenTCPAddrPort(netip.AddrPortFrom(cfg.Address, uint16(p)))
		if err != nil {
			t.Close()
			return nil, fmt.Errorf("listen on %s:%d in the tunnel: %w", cfg.Address, p, err)
		}
		t.lns = append(t.lns, ln)
		go t.forward(ln, cfg.Forwards[uint16(p)])
		cfg.Logf("forwarding %s:%d in the tunnel to %s", cfg.Address, p, cfg.Forwards[uint16(p)])
	}
	// Handshake at once, and again shortly after: the first one can go out before the receive side
	// is listening, and WireGuard's own retry waits 5 s.
	t.Handshake()
	go func() {
		for _, d := range []time.Duration{150 * time.Millisecond, 600 * time.Millisecond, 1500 * time.Millisecond} {
			time.Sleep(d)
			if !t.LastHandshake().IsZero() && time.Since(t.LastHandshake()) < 3*time.Second {
				return
			}
			t.Handshake()
		}
	}()
	return t, nil
}

func (t *Tunnel) forward(ln net.Listener, local string) {
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		go func() {
			defer c.Close()
			up, err := net.DialTimeout("tcp", local, 10*time.Second)
			if err != nil {
				t.cfg.Logf("forward to %s: %v", local, err)
				return
			}
			defer up.Close()
			done := make(chan struct{}, 2)
			go func() { io.Copy(up, c); done <- struct{}{} }()
			go func() { io.Copy(c, up); done <- struct{}{} }()
			<-done
		}()
	}
}

// Handshake sends a handshake to the hub now, instead of waiting for the next keepalive. It is sent
// as a retry: a normal initiation is skipped when one went out in the last 5 s (REKEY_TIMEOUT),
// which is exactly when a restart or a network change needs one.
func (t *Tunnel) Handshake() {
	if p := t.dev.LookupPeer(t.hub); p != nil {
		_ = p.SendHandshakeInitiation(true)
	}
}

// Rebind reopens the UDP socket and handshakes: what a network change needs.
func (t *Tunnel) Rebind() error {
	if err := t.dev.BindUpdate(); err != nil {
		return err
	}
	t.Handshake()
	return nil
}

// LastHandshake is the time of the latest completed handshake with the hub (zero if none).
func (t *Tunnel) LastHandshake() time.Time {
	s, err := t.dev.IpcGet()
	if err != nil {
		return time.Time{}
	}
	var sec, nsec int64
	for _, line := range strings.Split(s, "\n") {
		if v, ok := strings.CutPrefix(line, "last_handshake_time_sec="); ok {
			fmt.Sscan(v, &sec)
		}
		if v, ok := strings.CutPrefix(line, "last_handshake_time_nsec="); ok {
			fmt.Sscan(v, &nsec)
		}
	}
	if sec == 0 {
		return time.Time{}
	}
	return time.Unix(sec, nsec)
}

// Watch rebinds and handshakes whenever this machine's addresses change (a new Wi-Fi, a hotspot,
// a wake from sleep), until ctx ends. Without it, recovery waits for the next keepalive.
func (t *Tunnel) Watch(ctx context.Context) {
	last := addrFingerprint()
	tick := time.NewTicker(t.cfg.WatchEvery)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
		if now := addrFingerprint(); now != last {
			last = now
			if err := t.Rebind(); err != nil {
				t.cfg.Logf("network changed; rebind failed: %v", err)
			} else {
				t.cfg.Logf("network changed; reconnected to the hub")
			}
		}
	}
}

// addrFingerprint: this machine's non-loopback addresses, sorted.
func addrFingerprint() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	var s []string
	for _, a := range addrs {
		if ipn, ok := a.(*net.IPNet); ok && !ipn.IP.IsLoopback() && !ipn.IP.IsLinkLocalUnicast() {
			s = append(s, ipn.IP.String())
		}
	}
	sort.Strings(s)
	return strings.Join(s, ",")
}

// Close stops forwarding and the tunnel.
func (t *Tunnel) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, ln := range t.lns {
		ln.Close()
	}
	t.lns = nil
	t.dev.Close()
}

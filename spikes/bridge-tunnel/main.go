// Spike: can GarageAI Bridge terminate the tunnel itself, the way cloudflared does?
//
// Both ends run WireGuard in user space (wireguard-go with its gVisor netstack), so neither side
// needs a tun device, root, or a NetBird install.
//
//	garage: connects outward to the hub, accepts TCP on its tunnel address inside its own stack,
//	        and forwards each connection to a runtime on 127.0.0.1. The runtime never listens on
//	        the network.
//	hub:    the gateway's end. It listens on UDP, learns the garage's endpoint from the handshake
//	        (so the garage can sit behind any NAT), and calls the garage's runtime through the
//	        tunnel: /v1/models, a streamed chat request, and a bulk download to measure throughput.
//
// Run: spike keys; spike garage ...; spike hub ...  (see usage). This is a measurement, not product
// code: if it holds, the garage half moves into cli/ and the hub half onto the gateway.
package main

import (
	"bufio"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "keys":
		priv, pub := newKey()
		fmt.Printf("private=%s\npublic=%s\n", priv, pub)
	case "garage":
		garage(os.Args[2:])
	case "hub":
		hub(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  spike keys
  spike garage -key PRIV -hub-key PUB -hub HOST:PORT -addr 10.66.0.2 -forward 8001=127.0.0.1:8001
  spike hub -key PRIV -garage-key PUB -listen 51830 -addr 10.66.0.1 -garage-addr 10.66.0.2 -port 8001 [-bulk-mb 200]`)
	os.Exit(2)
}

// newKey makes a WireGuard (X25519) key pair, base64 like `wg genkey`, with the stdlib.
func newKey() (priv, pub string) {
	k, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(k.Bytes()), base64.StdEncoding.EncodeToString(k.PublicKey().Bytes())
}

// hexKey converts a base64 key to the hex form wireguard-go's IPC configuration wants.
func hexKey(b64 string) string {
	b, err := base64.StdEncoding.DecodeString(b64)
	if err != nil || len(b) != 32 {
		log.Fatalf("bad key %q", b64)
	}
	return hex.EncodeToString(b)
}

func stack(addr string) (*netstack.Net, *device.Device) {
	tunDev, tnet, err := netstack.CreateNetTUN([]netip.Addr{netip.MustParseAddr(addr)}, nil, 1420)
	if err != nil {
		log.Fatal(err)
	}
	dev := device.NewDevice(tunDev, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, "wg "))
	return tnet, dev
}

// ---------------------------------------------------------------- the garage end

func garage(args []string) {
	fs := flag.NewFlagSet("garage", flag.ExitOnError)
	key := fs.String("key", "", "garage private key (base64)")
	hubKey := fs.String("hub-key", "", "hub public key (base64)")
	hubAddr := fs.String("hub", "127.0.0.1:51830", "hub UDP endpoint")
	addr := fs.String("addr", "10.66.0.2", "this garage's tunnel address")
	forward := fs.String("forward", "8001=127.0.0.1:8001", "tunnel port = local address")
	fs.Parse(args)

	tnet, dev := stack(*addr)
	// The garage connects outward and keeps the NAT mapping alive; it needs no listening port.
	cfg := fmt.Sprintf("private_key=%s\npublic_key=%s\nendpoint=%s\nallowed_ip=10.66.0.1/32\npersistent_keepalive_interval=25\n",
		hexKey(*key), hexKey(*hubKey), *hubAddr)
	if err := dev.IpcSet(cfg); err != nil {
		log.Fatal(err)
	}
	if err := dev.Up(); err != nil {
		log.Fatal(err)
	}
	port, local, _ := strings.Cut(*forward, "=")
	ln, err := tnet.ListenTCPAddrPort(netip.MustParseAddrPort(*addr + ":" + port))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("garage: tunnel %s, forwarding %s:%s -> %s (runtime stays on localhost)", *addr, *addr, port, local)
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			defer c.Close()
			up, err := net.Dial("tcp", local)
			if err != nil {
				log.Printf("forward: %v", err)
				return
			}
			defer up.Close()
			done := make(chan struct{}, 2)
			go func() { io.Copy(up, c); done <- struct{}{} }()
			go func() { io.Copy(c, up); done <- struct{}{} }()
			<-done
		}(c)
	}
}

// ---------------------------------------------------------------- the hub (gateway) end

func hub(args []string) {
	fs := flag.NewFlagSet("hub", flag.ExitOnError)
	key := fs.String("key", "", "hub private key (base64)")
	garageKey := fs.String("garage-key", "", "garage public key (base64)")
	listen := fs.Int("listen", 51830, "UDP port")
	addr := fs.String("addr", "10.66.0.1", "hub tunnel address")
	garageAddr := fs.String("garage-addr", "10.66.0.2", "garage tunnel address")
	port := fs.Int("port", 8001, "runtime port on the garage's tunnel address")
	bulkMB := fs.Int("bulk-mb", 200, "MB to download for the throughput test")
	rounds := fs.Int("rounds", 1, "repeat the checks this many times, 5 s apart (for reconnect tests)")
	fs.Parse(args)

	tnet, dev := stack(*addr)
	// No endpoint for the garage: the hub learns it from the garage's handshake, whatever NAT it is behind.
	cfg := fmt.Sprintf("private_key=%s\nlisten_port=%d\npublic_key=%s\nallowed_ip=%s/32\n",
		hexKey(*key), *listen, hexKey(*garageKey), *garageAddr)
	if err := dev.IpcSet(cfg); err != nil {
		log.Fatal(err)
	}
	if err := dev.Up(); err != nil {
		log.Fatal(err)
	}
	client := &http.Client{Transport: &http.Transport{DialContext: tnet.DialContext}, Timeout: 120 * time.Second}
	base := fmt.Sprintf("http://%s:%d", *garageAddr, *port)

	for r := 1; r <= *rounds; r++ {
		if r > 1 {
			time.Sleep(5 * time.Second)
		}
		log.Printf("round %d", r)
		// 1. Wait for the garage's first handshake, then list models.
		t0 := time.Now()
		var models string
		for {
			resp, err := client.Get(base + "/v1/models")
			if err == nil {
				b, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				models = strings.TrimSpace(string(b))
				break
			}
			if time.Since(t0) > 60*time.Second {
				log.Printf("  /v1/models: no answer after 60 s: %v", err)
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		if models != "" {
			if len(models) > 120 {
				models = models[:120] + "…"
			}
			log.Printf("  /v1/models in %v: %s", time.Since(t0).Round(time.Millisecond), models)
		}

		// 2. A streamed chat request: time to first byte and the whole stream.
		t1 := time.Now()
		resp, err := client.Post(base+"/v1/chat/completions", "application/json",
			strings.NewReader(`{"model":"x","stream":true,"messages":[{"role":"user","content":"hi"}]}`))
		if err != nil {
			log.Printf("  stream: %v", err)
		} else {
			br := bufio.NewReader(resp.Body)
			_, _ = br.ReadByte()
			ttfb := time.Since(t1)
			rest, _ := io.ReadAll(br)
			resp.Body.Close()
			log.Printf("  stream: first byte after %v, %d bytes in %v", ttfb.Round(time.Millisecond), len(rest)+1, time.Since(t1).Round(time.Millisecond))
		}

		// 3. Throughput: download bulk-mb through the tunnel.
		t2 := time.Now()
		resp, err = client.Get(fmt.Sprintf("%s/bulk?mb=%d", base, *bulkMB))
		if err != nil {
			log.Printf("  bulk: %v", err)
			continue
		}
		n, _ := io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		secs := time.Since(t2).Seconds()
		log.Printf("  bulk: %d MB in %.2f s = %.0f Mbit/s", n>>20, secs, float64(n)*8/secs/1e6)
	}
	_ = context.Background()
}

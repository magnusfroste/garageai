package tunnel

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

// hub is a user-space stand-in for the gateway's wg-hub: no endpoint for the garage, it learns it.
func hub(t *testing.T, garagePub string) (port int, client *http.Client, pub string) {
	t.Helper()
	priv, pub, _ := GenerateKey()
	l, _ := net.ListenPacket("udp", "127.0.0.1:0")
	port = l.LocalAddr().(*net.UDPAddr).Port
	l.Close()
	tun, tnet, err := netstack.CreateNetTUN([]netip.Addr{HubAddress}, nil, 1420)
	if err != nil {
		t.Fatal(err)
	}
	dev := device.NewDevice(tun, conn.NewDefaultBind(), device.NewLogger(device.LogLevelSilent, ""))
	ph, _ := hexKey(priv)
	gh, _ := hexKey(garagePub)
	if err := dev.IpcSet(fmt.Sprintf("private_key=%s\nlisten_port=%d\npublic_key=%s\nallowed_ip=10.66.0.2/32\n", ph, port, gh)); err != nil {
		t.Fatal(err)
	}
	dev.Up()
	t.Cleanup(dev.Close)
	return port, &http.Client{Transport: &http.Transport{DialContext: tnet.DialContext, DisableKeepAlives: true}, Timeout: 700 * time.Millisecond}, pub
}

func get(c *http.Client, url string) (string, error) {
	r, err := c.Get(url)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	return string(b), err
}

func waitFor(t *testing.T, c *http.Client, url, want string) time.Duration {
	t.Helper()
	t0 := time.Now()
	for time.Since(t0) < 15*time.Second {
		if body, err := get(c, url); err == nil && strings.Contains(body, want) {
			return time.Since(t0)
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("no answer through the tunnel from %s", url)
	return 0
}

func TestTunnelForwardsRebindsAndReconnects(t *testing.T) {
	runtime := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[{"id":"qwen3:4b"}]}`))
	}))
	defer runtime.Close()
	local := strings.TrimPrefix(runtime.URL, "http://")
	priv, pub, _ := GenerateKey()
	if p2, _ := PublicKey(priv); p2 != pub {
		t.Fatal("PublicKey does not match GenerateKey")
	}
	port, client, hubPub := hub(t, pub)
	cfg := Config{PrivateKey: priv, HubPublicKey: hubPub, HubEndpoint: "127.0.0.1:" + strconv.Itoa(port),
		Address: netip.MustParseAddr("10.66.0.2"), Forwards: map[uint16]string{11434: local}}
	url := "http://10.66.0.2:11434/v1/models"

	tn, err := Start(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("first answer after %v", waitFor(t, client, url, "qwen3:4b"))
	if tn.LastHandshake().IsZero() {
		t.Fatal("no handshake recorded")
	}

	// A network change: a new UDP socket (new source port), and a handshake at once.
	if err := tn.Rebind(); err != nil {
		t.Fatal(err)
	}
	t.Logf("after rebind: answer in %v", waitFor(t, client, url, "qwen3:4b"))

	// The garage restarts: a new process, a new source port; the hub learns it from the handshake.
	tn.Close()
	// A real restart takes longer than WireGuard's replay window: an initiation within ~20 ms of the
	// last one, or with the same coarse timestamp, is dropped as a replay and retried 5 s later.
	time.Sleep(1500 * time.Millisecond)
	tn2, err := Start(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer tn2.Close()
	t.Logf("after restart: answer in %v", waitFor(t, client, url, "qwen3:4b"))
}

func TestKeys(t *testing.T) {
	if _, err := PublicKey("not a key"); err == nil {
		t.Fatal("accepted a bad key")
	}
}

package profile

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/service"
)

// run executes a command with a timeout and returns stdout; "" when it is missing or fails.
func run(name string, args ...string) string {
	if _, err := exec.LookPath(name); err != nil {
		return ""
	}
	cmd := exec.Command(name, args...)
	done := make(chan struct{})
	var out []byte
	go func() { out, _ = cmd.Output(); close(done) }()
	select {
	case <-done:
		return string(out)
	case <-time.After(5 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return ""
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

func machine() Machine {
	m := Machine{Arch: unameArch(), MemoryGB: totalMemoryGB()}
	switch runtime.GOOS {
	case "linux":
		m.OS = "Linux"
		if f, err := os.Open("/etc/os-release"); err == nil {
			sc := bufio.NewScanner(f)
			for sc.Scan() {
				if v, ok := strings.CutPrefix(sc.Text(), "PRETTY_NAME="); ok {
					m.OS = strings.Trim(v, `"'`)
				}
			}
			f.Close()
		}
	case "darwin":
		m.OS = strings.TrimSpace("macOS " + strings.TrimSpace(run("sw_vers", "-productVersion")))
	case "windows":
		m.OS = strings.TrimSpace(run("cmd", "/c", "ver"))
		if m.OS == "" {
			m.OS = "Windows"
		}
	default:
		m.OS = runtime.GOOS
	}
	return m
}

// unameArch matches what `uname -m` prints, as the bash script reports.
func unameArch() string {
	switch {
	case runtime.GOARCH == "amd64":
		return "x86_64"
	case runtime.GOARCH == "arm64" && runtime.GOOS == "linux":
		return "aarch64"
	}
	return runtime.GOARCH
}

func gpus() []GPU {
	out := []GPU{}
	if s := run("nvidia-smi", "--query-gpu=name,memory.total,driver_version", "--format=csv,noheader,nounits"); s != "" {
		r := csv.NewReader(strings.NewReader(s))
		r.TrimLeadingSpace = true
		rows, _ := r.ReadAll()
		for _, row := range rows {
			if len(row) >= 3 {
				g := GPU{Vendor: "nvidia", Name: row[0], Driver: row[2]}
				if n, err := strconv.Atoi(strings.TrimSpace(row[1])); err == nil {
					g.MemoryMB = &n
				}
				out = append(out, g)
			}
		}
		return out
	}
	if runtime.GOOS == "darwin" {
		g := GPU{Vendor: "apple", Name: strings.TrimSpace(run("sysctl", "-n", "machdep.cpu.brand_string")), UnifiedMemory: true}
		if n, err := strconv.ParseInt(strings.TrimSpace(run("sysctl", "-n", "hw.memsize")), 10, 64); err == nil {
			g.MemoryMB = intp(int(n / (1 << 20)))
		}
		return append(out, g)
	}
	return out
}

func netbird() NetBird {
	nb := NetBird{}
	if _, err := exec.LookPath("netbird"); err != nil {
		return nb
	}
	nb.Installed = true
	nb.Version = strp(strings.TrimSpace(strings.SplitN(run("netbird", "version"), "\n", 2)[0]))
	var st struct {
		NetbirdIP  string `json:"netbirdIp"`
		Management struct {
			Connected bool `json:"connected"`
		} `json:"management"`
	}
	if json.Unmarshal([]byte(run("netbird", "status", "--json")), &st) == nil {
		nb.Connected = st.Management.Connected
		nb.MeshIP = strp(strings.SplitN(st.NetbirdIP, "/", 2)[0])
	}
	if !nb.Connected {
		// A normal user may not be allowed to ask the NetBird service (macOS). The mesh address on
		// the tunnel interface (100.64.0.0/10) shows the client is up all the same.
		if ip := meshIPFromInterfaces(); ip != "" {
			nb.Connected, nb.MeshIP = true, &ip
		}
	}
	return nb
}

// MeshNet is GarageAI's NetBird network. Only an address inside it is the garage's mesh address:
// Tailscale and other WireGuard VPNs also hand out 100.64.0.0/10 addresses on utun interfaces, and
// one of those must never be taken for the mesh (GARAGEAI_MESH_CIDR overrides, for other setups).
func MeshNet() *net.IPNet {
	cidr := os.Getenv("GARAGEAI_MESH_CIDR")
	if cidr == "" {
		cidr = "100.83.0.0/16"
	}
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		_, n, _ = net.ParseCIDR("100.83.0.0/16")
	}
	return n
}

var cgnat = &net.IPNet{IP: net.IPv4(100, 64, 0, 0).To4(), Mask: net.CIDRMask(10, 32)}

// tunnelAddrs: the IPv4 addresses on tunnel interfaces (wt*, nb*, utun*, tailscale*, zt*) in 100.64.0.0/10.
func tunnelAddrs() []net.IP {
	var out []net.IP
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, it := range ifaces {
		n := it.Name
		if !(strings.HasPrefix(n, "wt") || strings.HasPrefix(n, "nb") || strings.HasPrefix(n, "utun") || strings.HasPrefix(n, "tailscale") || strings.HasPrefix(n, "zt")) {
			continue
		}
		addrs, _ := it.Addrs()
		for _, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok {
				if v4 := ipn.IP.To4(); v4 != nil && cgnat.Contains(v4) {
					out = append(out, v4)
				}
			}
		}
	}
	return out
}

func meshIPFromInterfaces() string {
	mesh := MeshNet()
	for _, ip := range tunnelAddrs() {
		if mesh.Contains(ip) {
			return ip.String()
		}
	}
	return ""
}

// otherVPNs names other mesh VPNs on this machine (Tailscale, ZeroTier): they run next to NetBird,
// and operators worry that they do not.
func otherVPNs() []string {
	var out []string
	_, tsApp := os.Stat("/Applications/Tailscale.app")
	_, tsCLI := exec.LookPath("tailscale")
	foreign := false
	mesh := MeshNet()
	for _, ip := range tunnelAddrs() {
		if !mesh.Contains(ip) {
			foreign = true
		}
	}
	if tsApp == nil || tsCLI == nil || foreign {
		out = append(out, "Tailscale")
	}
	if _, err := exec.LookPath("zerotier-cli"); err == nil {
		out = append(out, "ZeroTier")
	}
	return out
}

func firewall() Firewall {
	fw := Firewall{Tool: "none"}
	switch runtime.GOOS {
	case "darwin":
		fw.Tool = "macos-application-firewall"
		fw.Active = strings.Contains(strings.ToLower(run("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate")), "enabled")
	case "windows":
		fw.Tool = "windows-firewall"
		fw.Active = strings.Contains(run("netsh", "advfirewall", "show", "allprofiles", "state"), "ON")
	default:
		if strings.Contains(run("ufw", "status"), "Status: active") {
			fw.Tool, fw.Active = "ufw", true
		} else if _, err := exec.LookPath("nft"); err == nil {
			fw.Tool = "iptables"
		} else if _, err := exec.LookPath("iptables"); err == nil {
			fw.Tool = "iptables"
		}
		_, err := os.Stat("/etc/systemd/system/garageai-firewall.service")
		fw.GarageAIRuleInstalled = err == nil
	}
	return fw
}

// heartbeat and the configuration the connect script left behind.
func heartbeatAndConfig() (Heartbeat, GarageAI) {
	hb, cfg := Heartbeat{}, GarageAI{}
	switch runtime.GOOS {
	case "windows":
		// Bridge's scheduled task, or the connect script's old one (heartbeat.json, before 1.0).
		hb.Installed, hb.Active = service.Installed()
		if !hb.Installed {
			if b, err := os.ReadFile(filepath.Join(os.Getenv("ProgramData"), "GarageAI", "heartbeat.json")); err == nil {
				var c struct {
					Runtime string `json:"runtime"`
					Port    any    `json:"port"`
				}
				if json.Unmarshal(b, &c) == nil {
					cfg.ConfiguredRuntime = strp(c.Runtime)
					if p, err := strconv.Atoi(strings.TrimSpace(strings.Trim(jsonString(c.Port), `"`))); err == nil {
						cfg.ConfiguredPort = &p
					}
				}
				hb.Installed = true
				hb.Active = strings.Contains(run("schtasks", "/query", "/tn", "GarageAI heartbeat"), "Ready") ||
					strings.Contains(run("schtasks", "/query", "/tn", "GarageAI heartbeat"), "Running")
			}
		}
		if inst, _ := service.TunnelInstalled(); inst {
			cfg.Transport = "tunnel"
		}
		readBridgeConfig(&cfg)
	default:
		// Bridge's own service (garageai run) or the connect script's heartbeat: either counts.
		_, errBridge := os.Stat("/etc/garageai/bridge.json")
		_, errScript := os.Stat("/usr/local/bin/garageai-heartbeat")
		hb.Installed = errBridge == nil || errScript == nil
		if inst, _ := service.TunnelInstalled(); inst && cfg.Transport == "" {
			cfg.Transport = "tunnel" // the service is there even when no config is readable
		}
		if runtime.GOOS == "darwin" {
			// launchctl may refuse a normal user the system domain: a log written in the last 15
			// minutes (the heartbeat runs every 5) also proves it is running.
			hb.Active = run("launchctl", "print", "system/eu.garageai.bridge") != "" || run("launchctl", "print", "system/eu.garageai.heartbeat") != ""
			for _, log := range []string{"/var/log/garageai-bridge.log", "/var/log/garageai-heartbeat.log"} {
				if fi, err := os.Stat(log); err == nil && time.Since(fi.ModTime()) < 15*time.Minute {
					hb.Active = true
				}
			}
		} else {
			hb.Active = exec.Command("systemctl", "is-active", "--quiet", "garageai-bridge.timer").Run() == nil ||
				exec.Command("systemctl", "is-active", "--quiet", "garageai-heartbeat.timer").Run() == nil
		}
		if readBridgeConfig(&cfg) {
			// Bridge's own configuration
		} else if b, err := os.ReadFile("/etc/garageai/heartbeat.env"); err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				k, v, ok := strings.Cut(line, "=")
				v = strings.Trim(v, `'"`)
				if ok && k == "GARAGEAI_RUNTIME" {
					cfg.ConfiguredRuntime = strp(v)
				} else if ok && k == "GARAGEAI_PORT" {
					if p, err := strconv.Atoi(v); err == nil {
						cfg.ConfiguredPort = &p
					}
				}
			}
		}
	}
	return hb, cfg
}

// readBridgeConfig fills cfg from Bridge's configuration (the secret-free public copy when the
// full one is not readable, as for doctor run by a normal user). Same file layout on every OS:
// before 1.0.1 the Windows branch read only the old script's heartbeat.json, so a tunnel garage
// on Windows reported localhost_only and netbird_missing (w11, 2026-10-17).
func readBridgeConfig(cfg *GarageAI) bool {
	b, err := os.ReadFile(service.ConfigPath())
	if err != nil {
		b, err = os.ReadFile(service.PublicPath())
	}
	if err != nil {
		return false
	}
	var c struct {
		Transport string `json:"transport"`
		MeshIP    string `json:"mesh_ip"`
		Runtime   string `json:"runtime"`
		Port      int    `json:"port"`
	}
	if json.Unmarshal(b, &c) != nil {
		return false
	}
	cfg.ConfiguredRuntime = strp(c.Runtime)
	if c.Port > 0 {
		cfg.ConfiguredPort = &c.Port
	}
	if c.Transport == "tunnel" {
		cfg.Transport, cfg.TunnelIP = "tunnel", c.MeshIP
	}
	return true
}

func jsonString(v any) string { b, _ := json.Marshal(v); return string(b) }

// ollamaContextTarget mirrors the connect script: more memory, a bigger window.
func ollamaContextTarget(mem *int) int {
	if v, err := strconv.Atoi(os.Getenv("GARAGEAI_OLLAMA_CONTEXT")); err == nil && v > 0 {
		return v
	}
	gb := 0
	if mem != nil {
		gb = *mem
	}
	switch {
	case gb >= 64:
		return 65536
	case gb >= 32:
		return 32768
	case gb >= 16:
		return 16384
	}
	return 8192
}

// localFacts: the heartbeat's last output, whether OLLAMA_HOST is set for good on macOS (the
// LaunchAgent the connect script offers), and after how many minutes a Mac goes to sleep.
// userHomes is this user's home and, when running as root (the heartbeat under launchd, or sudo),
// also the home of the user at the console or behind sudo: their settings are the ones that count.
func userHomes() []string {
	var homes []string
	if home, err := os.UserHomeDir(); err == nil {
		homes = append(homes, home)
	}
	if os.Geteuid() == 0 {
		for _, u := range []string{os.Getenv("SUDO_USER"), strings.TrimSpace(run("stat", "-f", "%Su", "/dev/console"))} {
			if u != "" && u != "root" && !strings.ContainsAny(u, "/.") {
				homes = append(homes, filepath.Join("/Users", u))
			}
		}
	}
	return homes
}

func localFacts(hb Heartbeat) (last string, ollamaPersistent bool, sleepMin, displayMin int) {
	switch runtime.GOOS {
	case "darwin":
		for _, home := range userHomes() {
			if _, err := os.Stat(filepath.Join(home, "Library/LaunchAgents/eu.garageai.ollama-host.plist")); err == nil {
				ollamaPersistent = true
			}
		}
		// pmset: "sleep N" is minutes after the display turns off ("displaysleep M"); 0 = never.
		for _, line := range strings.Split(run("pmset", "-g"), "\n") {
			if f := strings.Fields(line); len(f) >= 2 {
				switch f[0] {
				case "sleep":
					sleepMin, _ = strconv.Atoi(f[1])
				case "displaysleep":
					displayMin, _ = strconv.Atoi(f[1])
				}
			}
		}
		for _, log := range []string{"/var/log/garageai-bridge.log", "/var/log/garageai-heartbeat.log"} {
			if b, err := os.ReadFile(log); err == nil && strings.TrimSpace(string(b)) != "" {
				lines := strings.Split(strings.TrimSpace(string(b)), "\n")
				last = lines[len(lines)-1]
				break
			}
		}
	case "linux":
		if hb.Installed {
			last = strings.TrimSpace(run("journalctl", "-u", "garageai-bridge.service", "-n", "1", "-o", "cat"))
			if last == "" {
				last = strings.TrimSpace(run("journalctl", "-u", "garageai-heartbeat.service", "-n", "1", "-o", "cat"))
			}
		}
	}
	return
}

package profile

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// unameArch matches what `uname -m` prints, as profiles always reported it.
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
	// The Bridge tunnel is the only transport: a config without one (written before transport
	// existed) is a tunnel garage too.
	cfg.Transport, cfg.TunnelIP = "tunnel", c.MeshIP
	return true
}

func jsonString(v any) string { b, _ := json.Marshal(v); return string(b) }

// ollamaContextTarget: more memory, a bigger window.
// ContextTarget is the window Bridge recommends for this machine's memory (Ollama and Runner).
func ContextTarget(mem *int) int { return ollamaContextTarget(mem) }

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

// localFacts: the heartbeat's last output, and after how many minutes a Mac goes to sleep.
func localFacts(hb Heartbeat) (last string, sleepMin, displayMin int) {
	switch runtime.GOOS {
	case "darwin":
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

// windowsSleep reads the active power plan with powercfg: minutes until sleep on the power
// adapter and whether closing the lid sleeps (any action but "do nothing") on the adapter. The
// output is localised, so it reads only the hex values: the last two are the AC and the DC
// setting. A desktop has no lid setting, which reads as "no". Unknown is (0, false): no advice.
func windowsSleep() (acMin int, lidSleeps bool) {
	if v, ok := powercfgAC(run("powercfg", "/query", "SCHEME_CURRENT", "SUB_SLEEP", "STANDBYIDLE")); ok {
		acMin = v / 60
	}
	if v, ok := powercfgAC(run("powercfg", "/query", "SCHEME_CURRENT", "SUB_BUTTONS", "LIDACTION")); ok {
		lidSleeps = v != 0
	}
	return
}

var hexValue = regexp.MustCompile(`0x([0-9a-fA-F]{8})`)

func powercfgAC(out string) (int, bool) {
	m := hexValue.FindAllStringSubmatch(out, -1)
	if len(m) < 2 {
		return 0, false
	}
	v, err := strconv.ParseUint(m[len(m)-2][1], 16, 32)
	return int(v), err == nil
}

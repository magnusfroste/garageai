package profile

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
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
	return nb
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
		conf := filepath.Join(os.Getenv("ProgramData"), "GarageAI", "heartbeat.json")
		if b, err := os.ReadFile(conf); err == nil {
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
		}
		hb.Active = strings.Contains(run("schtasks", "/query", "/tn", "GarageAI heartbeat"), "Ready") ||
			strings.Contains(run("schtasks", "/query", "/tn", "GarageAI heartbeat"), "Running")
	default:
		_, err := os.Stat("/usr/local/bin/garageai-heartbeat")
		hb.Installed = err == nil
		if runtime.GOOS == "darwin" {
			// launchctl may refuse a normal user the system domain: a log written in the last 15
			// minutes (the heartbeat runs every 5) also proves it is running.
			hb.Active = run("launchctl", "print", "system/eu.garageai.heartbeat") != ""
			if fi, err := os.Stat("/var/log/garageai-heartbeat.log"); err == nil && time.Since(fi.ModTime()) < 15*time.Minute {
				hb.Active = true
			}
		} else {
			hb.Active = exec.Command("systemctl", "is-active", "--quiet", "garageai-heartbeat.timer").Run() == nil
		}
		if b, err := os.ReadFile("/etc/garageai/heartbeat.env"); err == nil {
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
func localFacts(hb Heartbeat) (last string, ollamaPersistent bool, sleepMin int) {
	switch runtime.GOOS {
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			_, err := os.Stat(filepath.Join(home, "Library/LaunchAgents/eu.garageai.ollama-host.plist"))
			ollamaPersistent = err == nil
		}
		for _, line := range strings.Split(run("pmset", "-g"), "\n") {
			if f := strings.Fields(line); len(f) >= 2 && f[0] == "sleep" {
				sleepMin, _ = strconv.Atoi(f[1])
				break
			}
		}
		if b, err := os.ReadFile("/var/log/garageai-heartbeat.log"); err == nil {
			lines := strings.Split(strings.TrimSpace(string(b)), "\n")
			last = lines[len(lines)-1]
		}
	case "linux":
		if hb.Installed {
			last = strings.TrimSpace(run("journalctl", "-u", "garageai-heartbeat.service", "-n", "1", "-o", "cat"))
		}
	}
	return
}

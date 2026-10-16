// GarageAI Bridge: the bridge between an operator's GPU and the GarageAI network, as one command,
// `garageai`. The roadmap is docs/bridge.md.
//
//	garageai connect [options]   join the mesh, register the runtime, install the heartbeat
//	garageai doctor [--json]     what runs on this machine, what is wrong, and how to fix it
//	garageai run [--once]        the heartbeat (what the installed service runs)
//	garageai uninstall           remove the heartbeat and Bridge's configuration
//	garageai install-service     make the heartbeat service run this version
//	garageai mcp                 Bridge's tools for the operator's own agent (MCP over stdio)
//	garageai version
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/machine"
	"github.com/magnusfroste/garageai/cli/internal/profile"
	"github.com/magnusfroste/garageai/cli/internal/service"
)

// version is set at build time: -ldflags "-X main.version=..."
var version = "dev"

func main() {
	completePath()
	args := os.Args[1:]
	if len(args) == 0 {
		usage(2)
	}
	switch args[0] {
	case "version", "--version", "-v":
		fmt.Println("GarageAI Bridge", version)
	case "doctor":
		os.Exit(doctor(len(args) > 1 && args[1] == "--json"))
	case "connect", "login":
		os.Exit(connect(args[1:]))
	case "__connect-as-root":
		os.Exit(connectAsRoot())
	case "run":
		os.Exit(run(args[1:]))
	case "uninstall":
		os.Exit(uninstall(args[1:]))
	case "install-service":
		os.Exit(installService())
	case "mcp":
		os.Exit(mcpServe(os.Stdin, os.Stdout))
	case "tunnel":
		os.Exit(tunnelCmd())
	case "models":
		os.Exit(modelsCmd())
	case "pause":
		os.Exit(pauseCmd(args[1:]))
	case "resume":
		os.Exit(resumeCmd())
	case "update":
		os.Exit(updateCmd(args[1:]))
	case "help", "--help", "-h":
		usage(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		usage(2)
	}
}

// completePath adds the usual program directories to PATH. The heartbeat runs under launchd or
// systemd with a minimal PATH (/usr/bin:/bin:/usr/sbin:/sbin), where netbird in /usr/local/bin or a
// Homebrew tool is not found and doctor would report NetBird as missing.
func completePath() {
	have := map[string]bool{}
	for _, d := range filepath.SplitList(os.Getenv("PATH")) {
		have[d] = true
	}
	path := os.Getenv("PATH")
	for _, d := range []string{"/usr/local/bin", "/usr/local/sbin", "/opt/homebrew/bin", "/usr/bin", "/bin", "/usr/sbin", "/sbin"} {
		// As root (the heartbeat, connect's root half) only directories nobody else can write in:
		// Homebrew's are the user's, and a planted nvidia-smi there would run as root every 5 minutes.
		if os.Geteuid() == 0 && !rootOwnedDir(d) {
			continue
		}
		if !have[d] {
			if path != "" {
				path += string(os.PathListSeparator)
			}
			path += d
		}
	}
	os.Setenv("PATH", path)
}

// colorOn: colours only on a terminal, and never with NO_COLOR (https://no-color.org).
var colorOn = func() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}()

// sgr wraps s in an ANSI style (1 bold, 31 red, 32 green, 33 yellow) when colours are on.
func sgr(code, s string) string {
	if !colorOn {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}

func usage(code int) {
	fmt.Fprintln(os.Stderr, `GarageAI Bridge — connects your GPU to the GarageAI network

usage:
  garageai connect [options]   join the mesh, check the runtime, register, install the heartbeat
      Without options you approve this machine in the browser; nothing to copy.
      --setup-key KEY --management-url URL     (from the portal's command; the key is used once)
      --register-url URL --register-token TOK  (from the portal's command)
      --runtime ollama|lmstudio|llamacpp|vllm|sglang|paddock|unsloth|mlx|lemonade|other
      --port N  --name NAME  --models a,b  --runtime-api-key KEY  --skip-install  --no-heartbeat  --yes
      --no-auto-update             keep this version until you run garageai update yourself
      --tunnel bridge|netbird      bridge (default): Bridge carries the tunnel itself, no NetBird, and the
                                   runtime can stay on localhost. netbird: the NetBird mesh, the fallback
                                   for networks that block UDP
      Every option can also be given as GARAGEAI_SETUP_KEY, GARAGEAI_REGISTER_TOKEN, ... in the environment.
  garageai doctor [--json]     what runs on this machine, what is wrong, and how to fix it
  garageai run [--once]        the heartbeat: report the runtime's models (the service runs this)
  garageai uninstall           remove the heartbeat and Bridge's configuration
  garageai install-service     make the heartbeat service run this version (the installer does this)
  garageai mcp                 Bridge's tools for your own AI agent (MCP over stdio)
  garageai tunnel              this garage's own tunnel to the gateway (the installed service runs this)
  garageai connect --runtime <kind> --port <n>
                               change the runtime or port of a linked garage (no new link; --relink to link again)
  garageai models              models the portal asked this garage to fetch or remove, and how that went
  garageai pause [--reason "redeploy"] [--for 15m]
                               planned absence: the portal drains this garage, buyers go elsewhere, no alerts
  garageai resume              offer the garage again
  garageai update [vX.Y.Z]     update Bridge by hand (it updates itself to the portal's stable version otherwise)
  garageai version`)
	os.Exit(code)
}

// asRootSimple re-runs a command through sudo (no secrets involved).
func asRootSimple(cmd string, args []string) int {
	self, _ := os.Executable()
	c := exec.Command("sudo", append([]string{self, cmd}, args...)...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := c.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode()
		}
		return 1
	}
	return 0
}

func doctor(asJSON bool) int {
	p := profile.Build(version)
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(p)
	} else {
		printHuman(p)
	}
	if p.OK {
		return 0
	}
	return 1
}

func printHuman(p profile.Profile) {
	fmt.Printf("GarageAI doctor — %s, %s", p.Machine.OS, p.Machine.Arch)
	if p.Machine.MemoryGB != nil {
		fmt.Printf(", %d GB", *p.Machine.MemoryGB)
	}
	fmt.Println()
	for _, g := range p.GPUs {
		fmt.Printf("  GPU      %s", g.Name)
		if g.MemoryMB != nil {
			fmt.Printf(" (%d MB)", *g.MemoryMB)
		}
		fmt.Println()
	}
	nb := "not installed"
	if p.NetBird.Installed {
		nb = "installed, not connected"
		if p.NetBird.Connected && p.NetBird.MeshIP != nil {
			nb = "connected as " + *p.NetBird.MeshIP
		}
	}
	if p.Tunnel != nil {
		st := "service not installed"
		if p.Tunnel.Installed {
			st = "service installed, not running"
			if p.Tunnel.Active {
				st = "running"
			}
		}
		fmt.Printf("  Tunnel   Bridge tunnel %s, %s\n", p.Tunnel.IP, st)
		nb = "not needed (Bridge carries the tunnel)"
	}
	fmt.Println("  NetBird  " + nb)
	hb := "not installed"
	if p.Heartbeat.Installed {
		hb = "installed, not running"
		if p.Heartbeat.Active {
			hb = "running"
		}
	}
	fmt.Println("  Heartbeat " + hb)
	if st, nodes, ignored, panel := effectiveState(time.Now()); st != nil || ignored != "" || len(nodes) > 0 {
		if st != nil {
			line := fmt.Sprintf("  State    %s", st.Kind)
			if st.Reason != "" {
				line += " (" + st.Reason + ")"
			}
			if st.Since != nil {
				line += " since " + st.Since.Local().Format("15:04")
			}
			if st.Until != nil {
				line += ", until ~" + st.Until.Local().Format("15:04")
			}
			fmt.Println(line + ", by " + st.Source + ": the portal sends buyers elsewhere meanwhile")
		} else if panel != nil {
			fmt.Printf("  Manager  %s says %s\n", panel.Panel, panel.State.Kind)
		}
		for _, n := range nodes {
			fmt.Printf("  Node     %s (%s): %s%s\n", n.Name, n.Role, map[bool]string{true: "ok", false: "not ok"}[n.OK], nodeDetail(n))
		}
		if ignored != "" {
			fmt.Println("  Manager  state file ignored: " + ignored)
		}
	}
	if tel := machine.Collect(); tel.LoadPct != nil || tel.Power != "" {
		line := "  Now     "
		if tel.LoadPct != nil {
			line += fmt.Sprintf(" load %d %%", *tel.LoadPct)
		}
		if tel.MemFreeMB != nil && tel.MemTotalMB != nil {
			line += fmt.Sprintf(", %d of %d GB memory free", *tel.MemFreeMB/1024, *tel.MemTotalMB/1024)
		}
		if tel.DiskFreeGB != nil {
			line += fmt.Sprintf(", %d GB disk free", *tel.DiskFreeGB)
		}
		switch tel.Power {
		case "battery":
			line += fmt.Sprintf(", on battery (%s): away for buyers until plugged in", pctText(tel.BatteryPct))
		case "ac":
			line += ", on power"
		}
		if tel.UserActive != nil && *tel.UserActive {
			line += ", someone is using this machine"
		}
		if tel.GPUUtilPct != nil {
			line += fmt.Sprintf(", GPU %d %%", *tel.GPUUtilPct)
			if tel.GPUTempC != nil {
				line += fmt.Sprintf(" %d °C", *tel.GPUTempC)
			}
			if tel.GPUPowerW != nil {
				line += fmt.Sprintf(" %d W", *tel.GPUPowerW)
			}
			if tel.GPUUnified {
				line += " (unified memory: the free memory above is the GPU's headroom)"
			}
		}
		fmt.Println(line)
	}
	for _, m := range service.ModelStates(service.LoadModelsState()) {
		if m.Status == "downloading" {
			fmt.Printf("  Fetching %s: %d of %d MB (asked for by %s)\n", m.Model, m.CompletedMB, m.TotalMB, orUnknown(m.By))
		}
	}
	if len(p.Runtimes) == 0 {
		fmt.Println("  Runtime  none found")
	}
	for _, rt := range p.Runtimes {
		reach := "only this machine"
		if p.Tunnel != nil {
			reach = "reached through the Bridge tunnel" // the tunnel forwards to localhost: listening wider is not needed
		} else if rt.Network {
			reach = "reachable over the mesh"
		}
		fmt.Printf("  Runtime  %s on port %d (%s), %d model(s), listens on %s: %s\n", rt.Kind, rt.Port, rt.API, len(rt.Models), strings.Join(rt.Binds, " "), reach)
		loaded := map[string]int{}
		for _, l := range rt.OllamaLoaded {
			if l.Context != nil {
				loaded[l.Name] = *l.Context
			}
		}
		for _, m := range rt.Models {
			ctx := "context ?"
			switch {
			case m.Context != nil:
				ctx = fmt.Sprintf("context %d", *m.Context)
			case loaded[m.ID] > 0:
				ctx = fmt.Sprintf("loaded, context %d", loaded[m.ID])
			case rt.Kind == "ollama":
				ctx = "not loaded; Ollama loads it on the first request"
			}
			fmt.Printf("             %s (%s)\n", m.ID, ctx)
		}
	}
	fmt.Println()
	if len(p.Problems) == 0 {
		fmt.Println("  ✓ No problems found.")
		return
	}
	mark := map[string]string{"error": "✗", "warning": "!", "info": "·"}
	for _, pr := range p.Problems {
		fmt.Printf("  %s %s\n      → %s\n", mark[pr.Severity], pr.Message, pr.Fix)
	}
}

func pctText(p *int) string {
	if p == nil {
		return "?"
	}
	return fmt.Sprintf("%d %%", *p)
}

func nodeDetail(n machine.Node) string {
	out := ""
	if n.GPUTempC != nil {
		out += fmt.Sprintf(", GPU %.0f °C", *n.GPUTempC)
	}
	if n.GPUPowerW != nil {
		out += fmt.Sprintf(", %.0f W", *n.GPUPowerW)
	}
	if n.MemAvailableMB != nil {
		out += fmt.Sprintf(", %d GB memory available", *n.MemAvailableMB/1024)
	}
	return out
}

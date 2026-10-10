// GarageAI Bridge: the bridge between an operator's GPU and the GarageAI network, as one command,
// `garageai`. The roadmap is docs/bridge.md.
//
//	garageai connect [options]   join the mesh, register the runtime, install the heartbeat
//	garageai doctor [--json]     what runs on this machine, what is wrong, and how to fix it
//	garageai run [--once]        the heartbeat (what the installed service runs)
//	garageai uninstall           remove the heartbeat and Bridge's configuration
//	garageai version
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/magnusfroste/garageai/cli/internal/profile"
)

// version is set at build time: -ldflags "-X main.version=..."
var version = "dev"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage(2)
	}
	switch args[0] {
	case "version", "--version", "-v":
		fmt.Println("GarageAI Bridge", version)
	case "doctor":
		os.Exit(doctor(len(args) > 1 && args[1] == "--json"))
	case "connect":
		os.Exit(connect(args[1:]))
	case "__connect-as-root":
		os.Exit(connectAsRoot())
	case "run":
		os.Exit(run(args[1:]))
	case "uninstall":
		os.Exit(uninstall(args[1:]))
	case "help", "--help", "-h":
		usage(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		usage(2)
	}
}

func usage(code int) {
	fmt.Fprintln(os.Stderr, `GarageAI Bridge — connects your GPU to the GarageAI network

usage:
  garageai connect [options]   join the mesh, check the runtime, register, install the heartbeat
      --setup-key KEY --management-url URL     (from the portal's command; the key is used once)
      --register-url URL --register-token TOK  (from the portal's command)
      --runtime ollama|lmstudio|llamacpp|vllm|sglang|paddock|unsloth|mlx|lemonade|other
      --port N  --name NAME  --models a,b  --runtime-api-key KEY  --skip-install  --no-heartbeat  --yes
      Every option can also be given as GARAGEAI_SETUP_KEY, GARAGEAI_REGISTER_TOKEN, ... in the environment.
  garageai doctor [--json]     what runs on this machine, what is wrong, and how to fix it
  garageai run [--once]        the heartbeat: report the runtime's models (the service runs this)
  garageai uninstall           remove the heartbeat and Bridge's configuration
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
	fmt.Println("  NetBird  " + nb)
	hb := "not installed"
	if p.Heartbeat.Installed {
		hb = "installed, not running"
		if p.Heartbeat.Active {
			hb = "running"
		}
	}
	fmt.Println("  Heartbeat " + hb)
	if len(p.Runtimes) == 0 {
		fmt.Println("  Runtime  none found")
	}
	for _, rt := range p.Runtimes {
		reach := "only this machine"
		if rt.Network {
			reach = "reachable over the mesh"
		}
		fmt.Printf("  Runtime  %s on port %d (%s), %d model(s), listens on %s: %s\n", rt.Kind, rt.Port, rt.API, len(rt.Models), strings.Join(rt.Binds, " "), reach)
		for _, m := range rt.Models {
			ctx := "?"
			if m.Context != nil {
				ctx = fmt.Sprint(*m.Context)
			}
			fmt.Printf("             %s (context %s)\n", m.ID, ctx)
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

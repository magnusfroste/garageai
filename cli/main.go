// garageai: GarageAI's command for garage operators. This first version does one thing, the
// garage profile, so it can be compared with scripts/garageai-connect.sh --doctor --json.
//
//	garageai doctor           what runs on this machine, what is wrong, how to fix it
//	garageai doctor --json    the same as the garage profile (schema 1)
//	garageai version
package main

import (
	"encoding/json"
	"fmt"
	"os"

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
		fmt.Println("garageai", version)
	case "doctor":
		os.Exit(doctor(len(args) > 1 && args[1] == "--json"))
	case "help", "--help", "-h":
		usage(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		usage(2)
	}
}

func usage(code int) {
	fmt.Fprintln(os.Stderr, "usage: garageai doctor [--json] | garageai version")
	os.Exit(code)
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
	if len(p.Runtimes) == 0 {
		fmt.Println("  Runtime  none found")
	}
	for _, rt := range p.Runtimes {
		reach := "only this machine"
		if rt.Network {
			reach = "reachable over the mesh"
		}
		fmt.Printf("  Runtime  %s on port %d (%s), %d model(s), listens on %v: %s\n", rt.Kind, rt.Port, rt.API, len(rt.Models), rt.Binds, reach)
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

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/tray"
)

// trayCmd: `garageai tray` shows GarageAI in the menu bar (macOS) or the notification area
// (Windows); docs/bridge-menubar.md.
//
//	garageai tray              run the icon (what the login item starts)
//	garageai tray --install    start it now and at every login
//	garageai tray --uninstall  stop it and remove it from login
//	garageai tray --print      print the menu as text (any OS; for support and tests)
func trayCmd(args []string) int {
	mode := ""
	if len(args) > 0 {
		mode = args[0]
	}
	switch mode {
	case "--print":
		fmt.Print(tray.Build(tray.Load(time.Now(), version)).Text())
		return 0
	case "--install":
		if !tray.Supported {
			fmt.Println("The GarageAI icon is for macOS and Windows.")
			return 0
		}
		if err := tray.InstallLoginItem(true); err != nil {
			fmt.Fprintln(os.Stderr, "Could not add the GarageAI icon:", err)
			return 1
		}
		fmt.Println("✓ GarageAI is in the " + trayPlace() + " and opens at login. Hide it from its own menu; the garage keeps running.")
		return 0
	case "--uninstall":
		if err := tray.RemoveLoginItem(true); err != nil {
			fmt.Fprintln(os.Stderr, "Could not remove the GarageAI icon:", err)
			return 1
		}
		fmt.Println("✓ The GarageAI icon is removed (the garage keeps running).")
		return 0
	case "":
	default:
		fmt.Fprintln(os.Stderr, "usage: garageai tray [--install | --uninstall | --print]")
		return 2
	}
	if runtime.GOOS != "windows" && isAdmin() {
		fmt.Fprintln(os.Stderr, "The GarageAI icon runs as your user, not as root: run garageai tray without sudo.")
		return 2
	}
	if err := tray.Run(version); err != nil {
		if errors.Is(err, tray.ErrRunning) {
			fmt.Println("GarageAI is already in the " + trayPlace() + ".")
			return 0
		}
		fmt.Fprintln(os.Stderr, "garageai tray:", err)
		return 1
	}
	return 0
}

func trayPlace() string {
	if runtime.GOOS == "windows" {
		return "notification area"
	}
	return "menu bar"
}

// offerTray ends a successful connect on macOS and Windows: show GarageAI in the menu bar? It runs
// in the operator's own half of connect (not as root), so the icon is theirs. --yes says yes;
// without a terminal to ask in, it only says how.
func offerTray(yes bool) {
	if !tray.Supported || tray.LoginItemInstalled() {
		return
	}
	fmt.Println()
	if !yes {
		if fi, err := os.Stdin.Stat(); err != nil || fi.Mode()&os.ModeCharDevice == 0 {
			fmt.Println("  Show GarageAI in the " + trayPlace() + " (status, today's requests, models): garageai tray --install")
			return
		}
		fmt.Printf("  Show GarageAI in the %s? It shows the garage's status, today's requests and models, and only reads; it can't change anything. [Y/n] ", trayPlace())
		a, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if a = strings.TrimSpace(strings.ToLower(a)); a != "" && a != "y" && a != "yes" && a != "j" && a != "ja" {
			fmt.Println("  Later: garageai tray --install")
			return
		}
	}
	if err := tray.InstallLoginItem(true); err != nil {
		fmt.Println("  Could not add the icon:", err, "(later: garageai tray --install)")
		return
	}
	if runtime.GOOS == "windows" && isAdmin() {
		fmt.Println("  ✓ GarageAI appears in the notification area at your next sign-in (or run garageai tray in a normal window now).")
		return
	}
	fmt.Println("  ✓ GarageAI is in the " + trayPlace() + ". Hide it from its own menu; the garage keeps running.")
}

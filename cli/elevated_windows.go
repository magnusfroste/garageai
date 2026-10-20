//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows"
)

// isAdmin: an elevated process (PowerShell "Run as administrator"). Installing the services and
// writing under %ProgramData%\GarageAI need it.
func isAdmin() bool { return windows.GetCurrentProcessToken().IsElevated() }

const adminHint = "Open PowerShell as Administrator (right-click Start → Terminal (Admin)) and run this again"

// elevate runs `garageai cmd args...` elevated, the way sudo asks for the password on a Mac:
// Windows shows its own Yes/No dialog (UAC), the command runs in a new window that stays open at
// the end so its output can be read, and this process waits for it. Torbjörn, the first operator
// on his own Windows PC (2026-10-17), ran the line in a normal PowerShell and read the old
// "open PowerShell as Administrator" stop as a Defender block.
func elevate(cmd string, args []string) int {
	self, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "garageai "+cmd+": "+adminHint)
		return 1
	}
	list := []string{psQuote("__hold"), psQuote(cmd)}
	for _, a := range args {
		if strings.ContainsAny(a, " \t") {
			a = `"` + a + `"`
		}
		list = append(list, psQuote(a))
	}
	fmt.Println("Bridge installs Windows services, which needs Administrator: Windows asks you to allow it.")
	fmt.Println("Answer Yes; Bridge continues in a new window.")
	// 87 when Windows refused (No in the dialog, or no administrator account): a code Bridge itself
	// never exits with, so a real failure of the command is not mistaken for it.
	script := fmt.Sprintf("try { $p = Start-Process -FilePath %s -ArgumentList %s -Verb RunAs -Wait -PassThru -ErrorAction Stop } catch { exit 87 }; exit $p.ExitCode",
		psQuote(self), strings.Join(list, ","))
	c := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	if err := c.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() != 87 {
			fmt.Println("It stopped: see the other window for why.")
			return ee.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "Windows did not run it as Administrator (No in the dialog?). "+adminHint+".")
		return 1
	}
	fmt.Println("Done: see the other window for what happened.")
	return 0
}

func psQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// holdWindow runs `garageai args...` in this (elevated, new) console and keeps the window open
// at the end, so the operator can read the outcome before it closes.
func holdWindow(args []string) int {
	self, _ := os.Executable()
	c := exec.Command(self, args...)
	c.Env = append(os.Environ(), "GARAGEAI_VIA_HOLD=1") // connect: the unelevated window offers the icon
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	code := 0
	if err := c.Run(); err != nil {
		code = 1
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		}
	}
	fmt.Print("\nPress Enter to close this window ")
	fmt.Scanln()
	return code
}

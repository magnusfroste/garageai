package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// askSecret reads a secret from the terminal with echo off, straight from /dev/tty: connect's
// root half has the secrets on stdin, not the terminal, and a secret must never show on screen or
// end up in the shell history.
func askSecret(prompt string) string {
	if runtime.GOOS == "windows" {
		fmt.Print(prompt)
		s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		return strings.TrimSpace(s)
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return ""
	}
	defer tty.Close()
	fmt.Fprint(tty, prompt)
	stty := func(arg string) {
		c := exec.Command("stty", arg)
		c.Stdin = tty
		_ = c.Run()
	}
	stty("-echo")
	s, _ := bufio.NewReader(tty).ReadString('\n')
	stty("echo")
	fmt.Fprintln(tty)
	return strings.TrimSpace(s)
}

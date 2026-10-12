//go:build !windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func chownLike(path string, fi os.FileInfo) error {
	if st, ok := fi.Sys().(*syscall.Stat_t); ok {
		return os.Chown(path, int(st.Uid), int(st.Gid))
	}
	return nil
}

// operatorHomes: the home directories of the human users on this machine (where ~/.local/bin lives).
func operatorHomes() []string {
	var homes []string
	if h := os.Getenv("GARAGEAI_USER_HOME"); h != "" {
		homes = append(homes, h)
	}
	for _, base := range []string{"/Users", "/home"} {
		entries, _ := os.ReadDir(base)
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") && e.Name() != "Shared" {
				homes = append(homes, filepath.Join(base, e.Name()))
			}
		}
	}
	if u := strings.TrimSpace(func() string { o, _ := exec.Command("stat", "-f", "%Su", "/dev/console").Output(); return string(o) }()); u != "" && u != "root" {
		homes = append(homes, filepath.Join("/Users", u))
	}
	return homes
}

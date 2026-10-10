//go:build linux

package profile

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// listeners reads /proc/net/tcp{,6} and finds each socket's process through /proc/<pid>/fd. No
// ss or lsof needed. Sockets of other users' processes are listed without a process (not root).
func listeners() []Listener {
	byInode := map[string]Listener{}
	for _, f := range []struct {
		path string
		v6   bool
	}{{"/proc/net/tcp", false}, {"/proc/net/tcp6", true}} {
		if b, err := os.ReadFile(f.path); err == nil {
			for k, v := range parseProcNet(string(b), f.v6) {
				byInode[k] = v
			}
		}
	}
	procs, _ := filepath.Glob("/proc/[0-9]*/fd/*")
	for _, fd := range procs {
		link, err := os.Readlink(fd)
		if err != nil || !strings.HasPrefix(link, "socket:[") {
			continue
		}
		inode := strings.TrimSuffix(strings.TrimPrefix(link, "socket:["), "]")
		l, ok := byInode[inode]
		if !ok || l.PID != 0 {
			continue
		}
		pid, _ := strconv.Atoi(strings.Split(fd, "/")[2])
		comm, _ := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/comm")
		l.PID, l.Process = pid, strings.TrimSpace(string(comm))
		byInode[inode] = l
	}
	out := make([]Listener, 0, len(byInode))
	for _, l := range byInode {
		out = append(out, l)
	}
	return out
}

// cmdline is the process's arguments, one per element.
func cmdline(pid int) []string {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimRight(string(b), "\x00"), "\x00")
}

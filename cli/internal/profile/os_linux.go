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

func parentOf(pid int) (int, string) {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return 0, ""
	}
	// pid (comm) state ppid ...: comm may contain spaces, so split after the last ")".
	s := string(b)
	i := strings.LastIndex(s, ")")
	if i < 0 {
		return 0, ""
	}
	f := strings.Fields(s[i+1:])
	if len(f) < 2 {
		return 0, ""
	}
	ppid, _ := strconv.Atoi(f[1])
	args := cmdline(ppid)
	if len(args) == 0 {
		return ppid, ""
	}
	return ppid, args[0]
}

func processUID(pid int) int {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/status")
	if err != nil {
		return -1
	}
	for _, line := range strings.Split(string(b), "\n") {
		if v, ok := strings.CutPrefix(line, "Uid:"); ok {
			f := strings.Fields(v)
			if len(f) > 0 {
				n, _ := strconv.Atoi(f[0])
				return n
			}
		}
	}
	return -1
}

func humanUsers() int { return 0 } // only matters on macOS

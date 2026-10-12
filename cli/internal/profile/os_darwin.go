//go:build darwin

package profile

import "strings"

func listeners() []Listener {
	return parseLsof(run("lsof", "-nP", "-iTCP", "-sTCP:LISTEN", "-F", "pcn"))
}

func cmdline(pid int) []string {
	return strings.Fields(run("ps", "-o", "command=", "-p", itoa(pid)))
}

func parentOf(pid int) (int, string) {
	f := strings.Fields(run("ps", "-o", "ppid=,command=", "-p", itoa(pid)))
	if len(f) == 0 {
		return 0, ""
	}
	ppid := 0
	for _, c := range f[0] {
		if c < '0' || c > '9' {
			return 0, ""
		}
		ppid = ppid*10 + int(c-'0')
	}
	// the command of the parent, not of pid
	pc := strings.Fields(run("ps", "-o", "command=", "-p", itoa(ppid)))
	if len(pc) == 0 {
		return ppid, ""
	}
	return ppid, pc[0]
}

func processUID(pid int) int {
	s := strings.TrimSpace(run("ps", "-o", "uid=", "-p", itoa(pid)))
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return -1
		}
		n = n*10 + int(c-'0')
	}
	if s == "" {
		return -1
	}
	return n
}

// humanUsers counts macOS accounts with UID 501 and up (the ones a person logs in as).
func humanUsers() int {
	n := 0
	for _, line := range strings.Split(run("dscl", ".", "-list", "/Users", "UniqueID"), "\n") {
		f := strings.Fields(line)
		if len(f) == 2 {
			uid := 0
			for _, c := range f[1] {
				uid = uid*10 + int(c-'0')
			}
			if uid >= 501 && uid < 1000 {
				n++
			}
		}
	}
	return n
}

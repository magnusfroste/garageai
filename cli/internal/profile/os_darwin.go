//go:build darwin

package profile

import "strings"

func listeners() []Listener {
	return parseLsof(run("lsof", "-nP", "-iTCP", "-sTCP:LISTEN", "-F", "pcn"))
}

func cmdline(pid int) []string {
	return strings.Fields(run("ps", "-o", "command=", "-p", itoa(pid)))
}

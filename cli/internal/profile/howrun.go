package profile

import (
	"path/filepath"
	"strings"
)

// HowRun says how a runtime process is kept alive: by the system (a service), by a user's login
// session (a LaunchAgent, which stops at logout), or by a terminal window (which stops when the
// window closes). Walks up the parents a few steps.
type HowRun struct {
	Service  bool   // started by launchd (pid 1) or systemd as a system service
	UserSess bool   // macOS: a launchd user agent (parent pid 1, process not root)
	Terminal string // the shell or terminal app it runs under, "" if none
}

var shells = map[string]bool{"zsh": true, "bash": true, "sh": true, "fish": true, "login": true, "tmux": true, "screen": true,
	"Terminal": true, "iTerm2": true, "WezTerm": true, "alacritty": true, "kitty": true, "gnome-terminal-server": true, "konsole": true, "sshd": true, "sshd-session": true}

func howRun(pid int) HowRun {
	h := HowRun{}
	ppid, pcmd := parentOf(pid)
	if ppid <= 1 {
		h.Service = ppid == 1 && processUID(pid) == 0
		h.UserSess = ppid == 1 && processUID(pid) != 0
		return h
	}
	cur := ppid
	cmd := pcmd
	for i := 0; i < 5 && cur > 1; i++ {
		base := filepath.Base(strings.Fields(cmd + " ")[0])
		if shells[base] {
			h.Terminal = base
			return h
		}
		cur, cmd = parentOf(cur)
	}
	return h
}

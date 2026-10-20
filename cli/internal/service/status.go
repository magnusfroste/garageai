package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Status is what the heartbeat last saw and heard, for the menu-bar icon (docs/bridge-menubar.md):
// rewritten by every heartbeat, world-readable and free of secrets. The icon runs as the logged-in
// user and reads only this, bridge-public.json, pause.json and the tunnel's request timings; it
// never sees bridge.json, the tunnel key or the runtime's API key, and never talks to the network.
type Status struct {
	Schema    int       `json:"schema"` // 1
	UpdatedAt time.Time `json:"updated_at"`
	Version   string    `json:"version"` // the heartbeat's Bridge
	Name      string    `json:"name"`
	Runtime   string    `json:"runtime"`
	Port      int       `json:"port"`
	Transport string    `json:"transport,omitempty"`

	Heartbeat StatusHeartbeat `json:"heartbeat"`
	// Models the runtime serves right now (what the heartbeat reported), with context windows.
	Models   []string       `json:"models"`
	Contexts map[string]int `json:"contexts,omitempty"`
	// Why the garage takes no traffic right now (garageai pause, a runtime manager); nil = serving.
	State    *StatusState    `json:"state,omitempty"`
	Managed  []StatusModel   `json:"managed,omitempty"`  // models the portal asked Bridge to fetch
	Problems []StatusProblem `json:"problems,omitempty"` // doctor's findings
	Power    string          `json:"power,omitempty"`    // ac, battery
	Battery  *int            `json:"battery_pct,omitempty"`
	Update   StatusUpdate    `json:"update"`
	// What the portal says about the garage (live, today's figures), when the portal sends it.
	Portal *StatusPortal `json:"portal,omitempty"`
}

type StatusHeartbeat struct {
	OK       bool   `json:"ok"`
	Rejected bool   `json:"rejected,omitempty"` // the portal refused the garage's token
	Error    string `json:"error,omitempty"`
}

type StatusState struct {
	Kind   string     `json:"kind"`
	Reason string     `json:"reason,omitempty"`
	Until  *time.Time `json:"until,omitempty"`
	Source string     `json:"source,omitempty"`
}

type StatusModel struct {
	Model       string `json:"model"`
	Status      string `json:"status"`
	CompletedMB int64  `json:"completed_mb,omitempty"`
	TotalMB     int64  `json:"total_mb,omitempty"`
	Error       string `json:"error,omitempty"`
}

type StatusProblem struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type StatusUpdate struct {
	Auto      bool   `json:"auto"`
	Stable    string `json:"stable,omitempty"` // the portal's stable version
	LastError string `json:"last_error,omitempty"`
}

type StatusPortal struct {
	Live          *bool    `json:"live,omitempty"`
	Offered       []string `json:"offered,omitempty"`
	RequestsToday *int     `json:"requests_today,omitempty"`
	TokensToday   *int64   `json:"tokens_today,omitempty"`
	EarnedToday   *float64 `json:"earned_today,omitempty"`
	Currency      string   `json:"currency,omitempty"`
}

// StatusPath is a variable so tests can point it at a temp dir.
var StatusPath = func() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "status.json")
	}
	return "/etc/garageai/status.json"
}

// WriteStatus replaces the status file (0644; on Windows readable by Users, inside a directory
// that is otherwise Administrators and SYSTEM only).
func WriteStatus(s Status) error {
	s.Schema = 1
	b, _ := json.MarshalIndent(s, "", "  ")
	p := StatusPath()
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, p); err != nil {
		return err
	}
	MakeUserReadable(p)
	return nil
}

// MakeUserReadable lets every local user read a file Bridge wrote as root/SYSTEM. On Unix the mode
// does it (0644 in a 0755 directory); on Windows the file needs its own grant, because the
// directory passes on Administrators and SYSTEM only.
func MakeUserReadable(path string) {
	if runtime.GOOS == "windows" {
		_ = windowsRestrict(path, false, true)
	}
}

// ReadSmallFile reads a file for an unprivileged reader (the menu-bar icon): a regular file only,
// no symlink, at most max bytes.
func ReadSmallFile(path string, max int64) ([]byte, error) {
	st, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() || st.Size() > max {
		return nil, os.ErrInvalid
	}
	return os.ReadFile(path)
}

// LoadStatus reads the status file; nil when there is none or it is not valid.
func LoadStatus() *Status {
	b, err := ReadSmallFile(StatusPath(), 1<<20)
	if err != nil {
		return nil
	}
	var s Status
	if json.Unmarshal(b, &s) != nil || s.Schema != 1 {
		return nil
	}
	return &s
}

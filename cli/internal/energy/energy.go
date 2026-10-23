// Package energy measures what a garage's tokens cost in electricity: tokens per kWh, the number
// home owners with solar panels and batteries care about (docs/bridge.md, energy).
//
// The tunnel service (always running, as root) samples the compute power every few seconds:
// nvidia-smi's power.draw summed over the GPUs on Linux and Windows, powermetrics' combined
// CPU+GPU+ANE power on Apple silicon. Each sample counts toward the minute's energy, and toward the
// busy energy when a request is being served at that moment. The heartbeat sets that against the
// tokens the runtime reported for the same minutes. Only power figures and counts are kept.
package energy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Minute is one minute of energy (joules), all of it and the part spent while serving.
type Minute struct {
	At       time.Time `json:"at"` // the minute's start (UTC)
	Joules   float64   `json:"j"`
	BusyJ    float64   `json:"busy_j"`
	Seconds  float64   `json:"s"` // how much of the minute was sampled
	PeakW    float64   `json:"peak_w,omitempty"`
	Measured string    `json:"src,omitempty"`
}

const keep = 26 * time.Hour

// Path is where the tunnel service leaves the minutes for the heartbeat (root only, no secrets).
func Path() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "energy.json")
	}
	return "/etc/garageai/energy.json"
}

// Meter accumulates samples into minutes.
type Meter struct {
	mu      sync.Mutex
	minutes []Minute
	last    time.Time
}

// Add counts power w (watts), measured now by src, since the previous sample; busy says a request
// was being served. A gap longer than maxGap (sleep, a stalled reader) is not filled in.
func (m *Meter) Add(now time.Time, w float64, busy bool, src string) {
	const maxGap = 30 * time.Second
	m.mu.Lock()
	defer m.mu.Unlock()
	dt := now.Sub(m.last)
	first := m.last.IsZero()
	m.last = now
	if first || dt <= 0 || dt > maxGap || w < 0 {
		return
	}
	at := now.UTC().Truncate(time.Minute)
	if n := len(m.minutes); n == 0 || !m.minutes[n-1].At.Equal(at) {
		m.minutes = append(m.minutes, Minute{At: at, Measured: src})
	}
	cur := &m.minutes[len(m.minutes)-1]
	j := w * dt.Seconds()
	cur.Joules += j
	cur.Seconds += dt.Seconds()
	if busy {
		cur.BusyJ += j
	}
	if w > cur.PeakW {
		cur.PeakW = w
	}
	cut := now.Add(-keep)
	i := 0
	for i < len(m.minutes) && m.minutes[i].At.Before(cut) {
		i++
	}
	m.minutes = m.minutes[i:]
}

// Snapshot returns a copy of the minutes.
func (m *Meter) Snapshot() []Minute {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Minute(nil), m.minutes...)
}

// Load reads minutes written earlier (after a restart of the tunnel service).
func (m *Meter) Load(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var mins []Minute
	if json.Unmarshal(b, &mins) == nil {
		m.mu.Lock()
		m.minutes = mins
		m.mu.Unlock()
	}
}

// Write saves the minutes for the heartbeat.
func (m *Meter) Write(path string) error {
	b, _ := json.Marshal(m.Snapshot())
	if runtime.GOOS == "windows" {
		return os.WriteFile(path, b, 0o600) // in place: keeps the directory's Administrators/SYSTEM rights
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ReadMinutes reads the minutes the tunnel service wrote; nil when there are none.
func ReadMinutes() []Minute {
	b, err := os.ReadFile(Path())
	if err != nil {
		return nil
	}
	var mins []Minute
	if json.Unmarshal(b, &mins) != nil {
		return nil
	}
	return mins
}

// Run samples power every interval until stop is closed, counting a sample as busy while busy()
// says so, and writes the minutes to path every 30 s. It returns at once when this machine has no
// power reading (no NVIDIA GPU, not Apple silicon).
func Run(path string, interval time.Duration, busy func() bool, stop <-chan struct{}) {
	if _, _, ok := Watts(); !ok {
		return
	}
	m := &Meter{}
	m.Load(path)
	tick := time.NewTicker(interval)
	defer tick.Stop()
	lastWrite := time.Now()
	for {
		select {
		case <-stop:
			_ = m.Write(path)
			return
		case now := <-tick.C:
			w, src, ok := Watts()
			if ok {
				m.Add(now, w, busy(), src)
			}
			if now.Sub(lastWrite) >= 30*time.Second {
				_ = m.Write(path)
				lastWrite = now
			}
		}
	}
}

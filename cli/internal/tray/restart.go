package tray

import (
	"os"
	"time"
)

// binWatch notices that the binary the icon was started from has been replaced: Bridge's
// self-update (macOS, the root-owned copy) or the heartbeat's refresh of the operator's copy
// (Windows). The icon then starts the new binary and exits, so the operator sees the new icon
// without restarting it by hand (garage-m4 kept the old circle after v1.5.4, 2026-11-11).
type binWatch struct {
	path string
	fi   os.FileInfo
	next time.Time
}

// restartEnv marks an icon started by the one it replaces: it waits for the old one's lock.
const restartEnv = "GARAGEAI_TRAY_RESTART"

func newBinWatch(path string) *binWatch {
	fi, _ := os.Stat(path)
	return &binWatch{path: path, fi: fi}
}

// replaced is checked on every refresh and looks at the file once a minute. A file written less
// than 10 seconds ago may still be in the middle of its swap: it waits for the next look.
func (w *binWatch) replaced(now time.Time) bool {
	if w == nil || w.fi == nil || now.Before(w.next) {
		return false
	}
	w.next = now.Add(time.Minute)
	fi, err := os.Stat(w.path)
	if err != nil || now.Sub(fi.ModTime()) < 10*time.Second {
		return false
	}
	return !os.SameFile(fi, w.fi) || !fi.ModTime().Equal(w.fi.ModTime()) || fi.Size() != w.fi.Size()
}

// lockWait takes the one-icon lock; an icon started by restart waits up to 15 seconds for the old
// icon to let go of it.
func lockWait() (func(), error) {
	unlock, err := lock()
	for i := 0; err == ErrRunning && os.Getenv(restartEnv) == "1" && i < 30; i++ {
		time.Sleep(500 * time.Millisecond)
		unlock, err = lock()
	}
	return unlock, err
}

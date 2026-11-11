package tray

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBinWatch(t *testing.T) {
	p := filepath.Join(t.TempDir(), "garageai")
	os.WriteFile(p, []byte("v1"), 0o755)
	old := time.Now().Add(-time.Hour)
	os.Chtimes(p, old, old)
	w := newBinWatch(p)
	now := time.Now()
	if w.replaced(now) {
		t.Fatal("unchanged binary reported as replaced")
	}
	// Replaced the way self-update does it: a new file renamed over the old.
	os.WriteFile(p+".new", []byte("v2, longer"), 0o755)
	os.Rename(p+".new", p)
	if w.replaced(now.Add(30 * time.Second)) {
		t.Fatal("looked again before a minute had passed")
	}
	if w.replaced(time.Now()) {
		t.Fatal("a file written just now must settle for 10 seconds first")
	}
	if !w.replaced(time.Now().Add(3 * time.Minute)) {
		t.Fatal("replaced binary not noticed")
	}
	if (&binWatch{path: filepath.Join(t.TempDir(), "missing")}).replaced(now) {
		t.Fatal("a binary that was missing at start is never replaced")
	}
}

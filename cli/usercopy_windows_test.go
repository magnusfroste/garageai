//go:build windows

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteOwnCopyRefusalWindows(t *testing.T) {
	base := t.TempDir()
	t.Setenv("LOCALAPPDATA", base)
	dir := filepath.Join(base, "Programs", "GarageAI")
	os.MkdirAll(dir, 0o755)
	if why := writeOwnCopyRefusal(filepath.Join(dir, "garageai.exe")); why != "" {
		t.Fatalf("own copy refused: %s", why)
	}
	if why := writeOwnCopyRefusal(filepath.Join(t.TempDir(), "garageai.exe")); why == "" {
		t.Fatal("a copy outside LOCALAPPDATA must be refused")
	}
}

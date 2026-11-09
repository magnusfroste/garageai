//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRefreshCopyReplacesRegularFile(t *testing.T) {
	if isAdmin() {
		t.Skip("refreshCopy refuses root")
	}
	dir := t.TempDir()
	dest, src := filepath.Join(dir, "garageai"), filepath.Join(dir, "service")
	os.WriteFile(dest, []byte("old"), 0o755)
	os.WriteFile(src, []byte("new"), 0o755)
	if err := refreshCopy(dest, src); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(dest); string(b) != "new" {
		t.Fatalf("dest = %q, want new", b)
	}
	if _, err := os.Stat(dest + ".new"); !os.IsNotExist(err) {
		t.Fatal("the temporary file is left behind")
	}
}

func TestRefreshCopyRefusesSymlink(t *testing.T) {
	if isAdmin() {
		t.Skip("refreshCopy refuses root")
	}
	dir := t.TempDir()
	target, dest, src := filepath.Join(dir, "elsewhere"), filepath.Join(dir, "garageai"), filepath.Join(dir, "service")
	os.WriteFile(target, []byte("keep"), 0o644)
	os.Symlink(target, dest)
	os.WriteFile(src, []byte("new"), 0o755)
	if err := refreshCopy(dest, src); err == nil {
		t.Fatal("a symlinked copy must not be replaced")
	}
	if b, _ := os.ReadFile(target); string(b) != "keep" {
		t.Fatalf("the symlink's target changed to %q", b)
	}
}

func TestVersionOf(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "bridge")
	os.WriteFile(bin, []byte("#!/bin/sh\necho 'GarageAI Bridge v1.5.0'\n"), 0o755)
	if got := versionOf(bin); got != "v1.5.0" {
		t.Fatalf("versionOf = %q, want v1.5.0", got)
	}
	if got := versionOf(filepath.Join(t.TempDir(), "missing")); got != "" {
		t.Fatalf("versionOf(missing) = %q, want empty", got)
	}
}

func TestUseServiceCopyOnlyForOperatorCommands(t *testing.T) {
	for _, args := range [][]string{nil, {"run"}, {"tunnel"}, {"version"}, {"update"}, {"install-service"}, {"__refresh-user-copy"}} {
		if _, ok := useServiceCopy(args); ok {
			t.Errorf("useServiceCopy(%v) handed over", args)
		}
	}
	t.Setenv(delegatedEnv, "1")
	if _, ok := useServiceCopy([]string{"doctor"}); ok {
		t.Error("a delegated process must not hand over again")
	}
}

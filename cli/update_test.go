package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewerVersion(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"v0.4.6", "v0.5.0", true}, {"v0.4.6", "v0.4.7", true}, {"v0.4.6", "v0.4.6", false}, {"v0.5.0", "v0.4.9", false},
		{"dev", "v0.5.0", false}, {"v0.4.0-rc1", "v0.5.0", false}, {"v0.4.6", "latest", false}, {"v1.0.0", "v0.99.99", false},
	} {
		if got := newerVersion(c.a, c.b); got != c.want {
			t.Errorf("newerVersion(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

// A release server: the asset and SHA256SUMS. The download is accepted only when the sums match.
func TestDownloadVerified(t *testing.T) {
	asset := []byte("#!/bin/sh\necho GarageAI Bridge v9.9.9\n")
	sum := sha256.Sum256(asset)
	good := hex.EncodeToString(sum[:]) + "  " + assetName() + "\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/good/SHA256SUMS":
			w.Write([]byte(good))
		case r.URL.Path == "/bad/SHA256SUMS":
			w.Write([]byte("0000000000000000000000000000000000000000000000000000000000000000  " + assetName() + "\n"))
		case filepath.Base(r.URL.Path) == assetName():
			w.Write(asset)
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	dir := t.TempDir()
	p, err := downloadVerified(srv.URL+"/good", dir)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(p); string(b) != string(asset) {
		t.Fatal("downloaded file differs")
	}
	if _, err := downloadVerified(srv.URL+"/bad", dir); err == nil {
		t.Fatal("a wrong checksum was accepted")
	}
	if _, err := downloadVerified(srv.URL+"/none", dir); err == nil {
		t.Fatal("a missing SHA256SUMS was accepted")
	}
}

func TestTrustedBaseOnlyGitHubReleases(t *testing.T) {
	own := "https://github.com/magnusfroste/garageai/releases/download/bridge-v0.6.1"
	for _, c := range []struct{ base, want string }{
		{"", own},
		{own, own},
		{own + "/", own},
		{"https://github.com/magnusfroste/garageai/releases/download/bridge-v0.6.2", "https://github.com/magnusfroste/garageai/releases/download/bridge-v0.6.2"},
		{"https://evil.example/releases", own},
		{"http://github.com/magnusfroste/garageai/releases/download/bridge-v0.6.1", own},
		{"https://github.com/magnusfroste/garageai/releases/download/bridge-v0.6.1/../../../other/releases/download/x", own},
		{"https://github.com/magnusfroste/garageai/releases/download/bridge-v0.6.1?x=1", own},
		{"https://github.com/someone-else/garageai/releases/download/bridge-v0.6.1", own},
	} {
		if got := trustedBase("v0.6.1", c.base); got != c.want {
			t.Errorf("trustedBase(%q) = %q, want %q", c.base, got, c.want)
		}
	}
}

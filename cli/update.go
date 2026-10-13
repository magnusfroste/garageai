package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/service"
)

// Self-update (docs/bridge.md): the portal says which version garages should run (bridge_stable in
// the heartbeat's answer); Bridge fetches that release from GitHub, checks the SHA256SUMS, runs the
// new binary once, and only then swaps the service's copy and restarts the tunnel. Never straight
// from "latest": a bad release must not reach every garage at once.

const releasesBase = "https://github.com/magnusfroste/garageai/releases/download/bridge-"

// trustedBase is where the release is fetched from. The portal may say so (bridge_download_base),
// but only within this repository's GitHub releases: SHA256SUMS comes from the same place as the
// binary, so whoever picks the place picks what runs as root on every garage. A portal that is
// wrong or compromised must not be able to point Bridge anywhere else (found in the 2026-10-13
// security review). An unacceptable base is ignored, and the release's own directory is used.
func trustedBase(stable, base string) string {
	own := releasesBase + stable
	if base == "" {
		return own
	}
	b := strings.TrimSuffix(base, "/")
	if b == own || strings.HasPrefix(b, releasesBase) && !strings.ContainsAny(strings.TrimPrefix(b, releasesBase), "/?#") {
		return b
	}
	fmt.Printf("%s ignoring download base %q from the portal: not a GitHub release of this repository\n", service.Now(), base)
	return own
}

// newerVersion reports whether b (vX.Y.Z) is newer than a. Non-release versions ("dev", "v0.4.0-rc1")
// never update automatically.
func newerVersion(a, b string) bool {
	pa, oka := semver(a)
	pb, okb := semver(b)
	if !oka || !okb {
		return false
	}
	for i := 0; i < 3; i++ {
		if pb[i] != pa[i] {
			return pb[i] > pa[i]
		}
	}
	return false
}

func semver(v string) ([3]int, bool) {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

func assetName() string {
	name := "garageai-" + runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

var updateHTTP = &http.Client{Timeout: 5 * time.Minute}

func fetch(url string, max int64) ([]byte, error) {
	resp, err := updateHTTP.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, max))
}

// downloadVerified fetches the release asset for this OS and checks it against SHA256SUMS.
// Returns the path of the verified file (in dir).
func downloadVerified(base, dir string) (string, error) {
	base = strings.TrimSuffix(base, "/")
	sums, err := fetch(base+"/SHA256SUMS", 64<<10)
	if err != nil {
		return "", err
	}
	want := ""
	for _, line := range strings.Split(string(sums), "\n") {
		f := strings.Fields(line)
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == assetName() {
			want = f[0]
		}
	}
	if want == "" {
		return "", fmt.Errorf("no checksum for %s in SHA256SUMS", assetName())
	}
	b, err := fetch(base+"/"+assetName(), 200<<20)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != want {
		return "", errors.New("checksum mismatch: the download does not match SHA256SUMS")
	}
	p := filepath.Join(dir, "garageai.new")
	if err := os.WriteFile(p, b, 0o755); err != nil {
		return "", err
	}
	return p, nil
}

// selfUpdate brings the service's Bridge to version stable from base (the release's download
// directory). Runs in the heartbeat as root. It is a no-op when already there or on Windows.
func selfUpdate(stable, base string) error {
	if runtime.GOOS == "windows" {
		return errors.New("self-update is not implemented on windows yet")
	}
	if !newerVersion(version, stable) {
		return nil
	}
	base = trustedBase(stable, base)
	dir := filepath.Dir(service.BinaryPath())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := downloadVerified(base, dir)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)
	// The new binary must run and be what it claims before it replaces anything.
	out, err := exec.Command(tmp, "version").Output()
	if err != nil || !strings.Contains(string(out), stable) {
		return fmt.Errorf("the downloaded Bridge does not run as %s (%q, %v)", stable, strings.TrimSpace(string(out)), err)
	}
	if _, err := service.InstallBinary(tmp); err != nil {
		return err
	}
	service.RestartTunnel()
	updateOperatorCopies(tmp)
	return nil
}

// updateOperatorCopies refreshes the operator's own garageai (what they type), when it is the
// usual install location, so `garageai doctor` and the service agree. Best effort, keeps owners.
func updateOperatorCopies(src string) {
	b, err := os.ReadFile(src)
	if err != nil {
		return
	}
	var paths []string
	if runtime.GOOS != "windows" {
		paths = append(paths, "/usr/local/bin/garageai")
		for _, h := range operatorHomes() {
			paths = append(paths, filepath.Join(h, ".local", "bin", "garageai"))
		}
	}
	for _, p := range paths {
		fi, err := os.Stat(p)
		if err != nil || fi.IsDir() {
			continue
		}
		tmp := p + ".new"
		if os.WriteFile(tmp, b, fi.Mode().Perm()) != nil {
			continue
		}
		_ = chownLike(tmp, fi)
		if os.Rename(tmp, p) != nil {
			os.Remove(tmp)
		}
	}
}

// updateCmd is `garageai update [vX.Y.Z]`: the same as the automatic update, by hand. Without a
// version it takes the newest release, which is what the installer does too.
func updateCmd(args []string) int {
	if os.Geteuid() != 0 && runtime.GOOS != "windows" {
		return asRootSimple("update", args)
	}
	target := ""
	if len(args) > 0 {
		target = args[0]
	}
	if target == "" {
		v, err := latestRelease()
		if err != nil {
			fmt.Fprintln(os.Stderr, "garageai update: could not read the releases:", err)
			return 1
		}
		target = v
	}
	if !newerVersion(version, target) {
		fmt.Printf("Bridge %s is already %s or newer.\n", version, target)
		return 0
	}
	fmt.Printf("Updating Bridge %s → %s (checksum verified before anything is replaced)...\n", version, target)
	if err := selfUpdate(target, ""); err != nil {
		fmt.Fprintln(os.Stderr, "garageai update:", err)
		return 1
	}
	service.SaveUpdateState(service.UpdateState{LastCheck: service.Now(), LastUpdate: service.Now(), Version: target})
	fmt.Printf("✓ Bridge %s installed for the service; the heartbeat uses it from its next run, the tunnel was restarted.\n", target)
	return 0
}

// latestRelease: the newest stable bridge-vX.Y.Z release on GitHub (what install.sh picks).
func latestRelease() (string, error) {
	b, err := fetch("https://api.github.com/repos/magnusfroste/garageai/releases?per_page=30", 4<<20)
	if err != nil {
		return "", err
	}
	best := ""
	for _, line := range strings.Split(string(b), ",") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, `"tag_name":"bridge-`); ok {
			v = strings.Trim(v, `"`)
			if _, ok := semver(v); ok && (best == "" || newerVersion(best, v)) {
				best = v
			}
		}
	}
	if best == "" {
		return "", errors.New("no bridge-vX.Y.Z release found")
	}
	return best, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/runnerfit"
	"github.com/magnusfroste/garageai/cli/internal/service"
	"github.com/magnusfroste/garageai/cli/internal/tunnel"
)

// Runner follows Bridge. A Runner that `garageai runtime install` installed stays on the release it
// was installed with unless something moves it; Bridge pins a Runner release (runnerfit.Version), and
// a Bridge self-update brings a newer pin. Then the heartbeat upgrades Runner too, the same staged
// way Bridge itself moves (canary garages first, everyone 24 h later):
//
//   - only Runner installed by Bridge, only to a newer release, not with auto-update off;
//   - only when it is quiet: Runner reports no active request, and none came in the last minute;
//   - same model, port and context window as before;
//   - the previous binary is kept, and restored when the new one does not answer within 3 minutes;
//     a failure is reported through doctor (runner_update_failed), which pages Ops.

const runnerStartLimit = 3 * time.Minute

// runnerVersionUnknownLogged keeps the "version unknown" line to once per run.
var runnerVersionUnknownLogged bool

func maybeUpgradeRunner(cfg *service.Config) {
	if cfg.AutoUpdate != nil && !*cfg.AutoUpdate {
		return
	}
	settings, ok := service.RunnerConfig()
	if !ok {
		return
	}
	have := service.InstalledRunnerVersion()
	if have == "" {
		if !runnerVersionUnknownLogged {
			runnerVersionUnknownLogged = true
			fmt.Printf("%s Runner version unknown (no runner.version, no answer to --version); not upgrading it\n", service.Now())
		}
		return
	}
	if !newerVersion(have, runnerfit.Version) {
		return
	}
	if !runnerQuiet(settings.Port) {
		fmt.Printf("%s Runner %s → %s waits: requests are being served\n", service.Now(), have, runnerfit.Version)
		return
	}
	st := service.LoadUpdateState()
	err := upgradeRunner(settings, have)
	if err != nil {
		st.RunnerError = fmt.Sprintf("Runner %s did not start; back on %s: %v", runnerfit.Version, have, err)
		fmt.Printf("%s %s\n", service.Now(), st.RunnerError)
	} else {
		st.RunnerError, st.RunnerUpdate = "", service.Now()
		fmt.Printf("%s Runner %s → %s: serving %s again\n", service.Now(), have, runnerfit.Version, settings.Model)
	}
	service.SaveUpdateState(st)
}

func upgradeRunner(s service.RunnerSettings, have string) error {
	asset := runnerfit.Asset(runtime.GOOS, runtime.GOARCH)
	if asset == "" {
		return fmt.Errorf("no Runner build for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	bin := service.RunnerPath()
	tmp, old := bin+".new", bin+".old"
	base := "https://github.com/Joakimpalm-Zen/xyntetik-runner/releases/download/" + runnerfit.Version
	if err := downloadAsset(base, asset, tmp, 512<<20); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("download: %w", err)
	}
	service.RemoveRunner() // stops it; a running Runner holds its file on Windows
	os.Remove(old)
	if err := os.Rename(bin, old); err != nil {
		os.Remove(tmp)
		_ = service.InstallRunner(bin, s.Model, s.Port, s.Context)
		return fmt.Errorf("keep the old binary: %w", err)
	}
	if err := os.Rename(tmp, bin); err != nil {
		_ = os.Rename(old, bin)
		_ = service.InstallRunner(bin, s.Model, s.Port, s.Context)
		return fmt.Errorf("install: %w", err)
	}
	if err := service.InstallRunner(bin, s.Model, s.Port, s.Context); err == nil && waitRunner(s.Port, runnerStartLimit) {
		service.RecordRunnerVersion(runnerfit.Version)
		os.Remove(old)
		return nil
	}
	// The new release does not serve: back to the old one, as it was.
	service.RemoveRunner()
	_ = os.Rename(old, bin)
	_ = service.InstallRunner(bin, s.Model, s.Port, s.Context)
	service.RecordRunnerVersion(have)
	return fmt.Errorf("it did not answer on 127.0.0.1:%d within %s", s.Port, runnerStartLimit)
}

func waitRunner(port int, limit time.Duration) bool {
	deadline := time.Now().Add(limit)
	for time.Now().Before(deadline) {
		if runnerAnswers(port) {
			return true
		}
		time.Sleep(5 * time.Second)
	}
	return false
}

// runnerQuiet: Runner serves nothing now (its /health active_requests) and the tunnel saw no
// request in the last minute.
func runnerQuiet(port int) bool {
	cl := &http.Client{Timeout: 3 * time.Second}
	resp, err := cl.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/health")
	if err == nil {
		var h struct {
			Active *int `json:"active_requests"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&h)
		resp.Body.Close()
		if h.Active != nil && *h.Active > 0 {
			return false
		}
	}
	return quietSince(tunnel.ReadSamples(), time.Now(), time.Minute)
}

func quietSince(samples []tunnel.Sample, now time.Time, d time.Duration) bool {
	for _, s := range samples {
		end := s.At.Add(time.Duration(s.TotalMs) * time.Millisecond)
		if now.Sub(end) < d && (s.Path == "chat" || s.Path == "completions" || s.Path == "embeddings") {
			return false
		}
	}
	return true
}

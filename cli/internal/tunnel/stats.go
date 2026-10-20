package tunnel

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

// The garage's own view of every request (docs/bridge.md, observability): the forward sees each
// one, so it can time it where it happens. The gateway times the same requests from its end;
// the difference is the tunnel and the network, and that separates a slow network from a slow
// model, which nothing could before. The tunnel process keeps the last hour of samples and
// writes them to a file every 30 s; the heartbeat (another process) reads the file and reports
// a summary with each beat.

// Sample is one request through the forward.
type Sample struct {
	At       time.Time `json:"at"`
	Path     string    `json:"path"`     // a class: chat, completions, embeddings, models, other
	Status   int       `json:"status"`   //
	TTFBms   int       `json:"ttfb_ms"`  // runtime's first byte (headers) after the request reached Bridge
	TotalMs  int       `json:"total_ms"` //
	Streamed bool      `json:"streamed"`
	Cold     bool      `json:"cold"` // first request after ≥5 min of silence that took ≥3 s to the first byte: a model load
}

// Summary is what the heartbeat reports: the last window of samples, aggregated.
type Summary struct {
	WindowMin  int `json:"window_min"`
	Requests   int `json:"requests"`
	Errors     int `json:"errors"` // status ≥ 500 (the runtime's or Bridge's), not the buyer's 4xx
	Streamed   int `json:"streamed"`
	ColdStarts int `json:"cold_starts"`
	TTFBp50ms  int `json:"ttfb_p50_ms"`
	TTFBp95ms  int `json:"ttfb_p95_ms"`
	TotalP50ms int `json:"total_p50_ms"`
	Inference  int `json:"inference"` // chat + completions + embeddings (the rest are probes)
}

const keepSamples = time.Hour

// StatsPath is where the tunnel process leaves its samples for the heartbeat.
func StatsPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramData"), "GarageAI", "forward-stats.json")
	}
	return "/etc/garageai/forward-stats.json"
}

type stats struct {
	mu       sync.Mutex
	samples  []Sample
	last     time.Time // the previous request's time, for the cold-start heuristic
	dirty    bool
	readable bool // Windows: the read grant for users is set (once per process)
}

var global = &stats{}

func pathClass(p string) string {
	switch {
	case p == "/v1/chat/completions" || p == "/api/chat" || p == "/v1/messages" || p == "/v1/responses":
		return "chat"
	case p == "/v1/completions" || p == "/api/generate":
		return "completions"
	case p == "/v1/embeddings" || p == "/api/embed" || p == "/api/embeddings":
		return "embeddings"
	case p == "/v1/models" || p == "/api/tags":
		return "models"
	}
	return "other"
}

func (s *stats) record(sm Sample) {
	s.mu.Lock()
	defer s.mu.Unlock()
	class := sm.Path
	if class == "chat" || class == "completions" {
		sm.Cold = sm.TTFBms >= 3000 && (s.last.IsZero() || sm.At.Sub(s.last) >= 5*time.Minute)
		s.last = sm.At
	}
	s.samples = append(s.samples, sm)
	cut := sm.At.Add(-keepSamples)
	i := 0
	for i < len(s.samples) && s.samples[i].At.Before(cut) {
		i++
	}
	s.samples = s.samples[i:]
	s.dirty = true
}

// flush writes the samples when something changed. Called every 30 s by the tunnel process.
func (s *stats) flush(path string) {
	s.mu.Lock()
	if !s.dirty {
		s.mu.Unlock()
		return
	}
	b, _ := json.Marshal(s.samples)
	s.dirty = false
	s.mu.Unlock()
	if runtime.GOOS == "windows" {
		// In place: a renamed file would take the directory's Administrators-and-SYSTEM-only
		// rights and lose the read grant the menu-bar icon needs (a half-written file is skipped
		// by readers as invalid JSON and read again 30 s later).
		if os.WriteFile(path, b, 0o644) == nil && !s.readable && MakeReadable != nil {
			MakeReadable(path)
			s.readable = true
		}
		return
	}
	tmp := path + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		_ = os.Rename(tmp, path)
	}
}

// MakeReadable is set by the tunnel service to service.MakeUserReadable (the tunnel package does
// not import service): the request timings are for the menu-bar icon too, which runs as the user.
var MakeReadable func(path string)

// FlushStatsEvery keeps the stats file current; returns when stop is closed.
func FlushStatsEvery(d time.Duration, stop <-chan struct{}) {
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			global.flush(StatsPath())
		case <-stop:
			global.flush(StatsPath())
			return
		}
	}
}

// Summarize aggregates the samples of the last window (the heartbeat's report).
func Summarize(samples []Sample, window time.Duration, now time.Time) Summary {
	out := Summary{WindowMin: int(window / time.Minute)}
	var ttfb, total []int
	cut := now.Add(-window)
	for _, s := range samples {
		if s.At.Before(cut) {
			continue
		}
		out.Requests++
		if s.Status >= 500 {
			out.Errors++
		}
		if s.Streamed {
			out.Streamed++
		}
		if s.Cold {
			out.ColdStarts++
		}
		if s.Path == "chat" || s.Path == "completions" || s.Path == "embeddings" {
			out.Inference++
			ttfb = append(ttfb, s.TTFBms)
			total = append(total, s.TotalMs)
		}
	}
	sort.Ints(ttfb)
	sort.Ints(total)
	pct := func(v []int, p float64) int { // nearest rank
		if len(v) == 0 {
			return 0
		}
		i := int(math.Ceil(p*float64(len(v)))) - 1
		if i < 0 {
			i = 0
		}
		return v[i]
	}
	out.TTFBp50ms, out.TTFBp95ms, out.TotalP50ms = pct(ttfb, 0.5), pct(ttfb, 0.95), pct(total, 0.5)
	return out
}

// ReadSummary reads the tunnel process's samples and summarises the last window; nil without a file.
func ReadSummary(window time.Duration) *Summary {
	b, err := os.ReadFile(StatsPath())
	if err != nil {
		return nil
	}
	var samples []Sample
	if json.Unmarshal(b, &samples) != nil {
		return nil
	}
	s := Summarize(samples, window, time.Now())
	return &s
}

// recorder wraps a ResponseWriter to time the first byte and count the status.
type recorder struct {
	http.ResponseWriter
	start    time.Time
	status   int
	firstAt  time.Time
	streamed bool
}

func (r *recorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
		r.firstAt = time.Now()
		r.streamed = r.ResponseWriter.Header().Get("Content-Type") == "text/event-stream" || r.ResponseWriter.Header().Get("Content-Type") == "application/x-ndjson"
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}

func (r *recorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// recordRequests is the middleware the forward wraps around the proxy.
func recordRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &recorder{ResponseWriter: w, start: time.Now()}
		next.ServeHTTP(rec, r)
		end := time.Now()
		if rec.firstAt.IsZero() {
			rec.firstAt = end
		}
		global.record(Sample{At: rec.start, Path: pathClass(r.URL.Path), Status: rec.status, Streamed: rec.streamed,
			TTFBms: int(rec.firstAt.Sub(rec.start) / time.Millisecond), TotalMs: int(end.Sub(rec.start) / time.Millisecond)})
	})
}

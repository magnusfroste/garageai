package tunnel

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSummarizeAndColdStart(t *testing.T) {
	s := &stats{}
	now := time.Now()
	s.record(Sample{At: now.Add(-20 * time.Minute), Path: "chat", Status: 200, TTFBms: 9000, TotalMs: 9500, Streamed: true}) // first after silence: cold
	s.record(Sample{At: now.Add(-19 * time.Minute), Path: "chat", Status: 200, TTFBms: 400, TotalMs: 2000})
	s.record(Sample{At: now.Add(-3 * time.Minute), Path: "models", Status: 200, TTFBms: 5, TotalMs: 6})
	s.record(Sample{At: now.Add(-2 * time.Minute), Path: "chat", Status: 200, TTFBms: 500, TotalMs: 2500, Streamed: true})
	s.record(Sample{At: now.Add(-1 * time.Minute), Path: "chat", Status: 502, TTFBms: 100, TotalMs: 100})
	if !s.samples[0].Cold || s.samples[1].Cold || s.samples[3].Cold {
		t.Fatalf("cold flags: %+v", s.samples)
	}
	sum := Summarize(s.samples, 5*time.Minute, now)
	if sum.Requests != 3 || sum.Inference != 2 || sum.Errors != 1 || sum.Streamed != 1 || sum.ColdStarts != 0 {
		t.Fatalf("5 min: %+v", sum)
	}
	if sum.TTFBp50ms != 100 && sum.TTFBp50ms != 500 {
		t.Fatalf("ttfb p50 %d", sum.TTFBp50ms)
	}
	hour := Summarize(s.samples, time.Hour, now)
	if hour.Requests != 5 || hour.ColdStarts != 1 || hour.TTFBp95ms != 9000 {
		t.Fatalf("1 h: %+v", hour)
	}
}

func TestPathClass(t *testing.T) {
	for p, want := range map[string]string{"/v1/chat/completions": "chat", "/v1/messages": "chat", "/v1/embeddings": "embeddings", "/v1/models": "models", "/metrics": "other"} {
		if got := pathClass(p); got != want {
			t.Errorf("%s: %s, want %s", p, got, want)
		}
	}
}

// A handler aborted with http.ErrAbortHandler (the reverse proxy does this when the buyer drops a
// stream) must not leave the in-flight count up, and the request is still recorded.
func TestAbortedRequestIsCounted(t *testing.T) {
	before := InFlight()
	h := recordRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body) // as the reverse proxy does
		w.WriteHeader(200)
		w.Write([]byte(`data: {"usage":{"completion_tokens":7}}`))
		panic(http.ErrAbortHandler)
	}))
	func() {
		defer func() { _ = recover() }()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(`{"model":"m"}`)))
	}()
	if InFlight() != before {
		t.Fatalf("in-flight stayed at %d after an aborted request", InFlight())
	}
	global.mu.Lock()
	last := global.samples[len(global.samples)-1]
	global.mu.Unlock()
	if last.Model != "m" || last.OutTokens != 7 {
		t.Fatalf("the aborted request was not recorded: %+v", last)
	}
}

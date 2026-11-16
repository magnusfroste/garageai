package portal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOwnAddress(t *testing.T) {
	for in, want := range map[string]string{
		LegacyAPI + "/register-node":    DefaultAPI + "/register-node",
		LegacyAPI:                       DefaultAPI,
		DefaultAPI + "/node-heartbeat":  DefaultAPI + "/node-heartbeat",
		LegacyAPI + "x/register-node":   LegacyAPI + "x/register-node", // another host that only starts the same
		"https://example.test/register": "https://example.test/register",
	} {
		if got := ownAddress(in); got != want {
			t.Errorf("ownAddress(%q) = %q, want %q", in, got, want)
		}
	}
	if c := New(LegacyAPI+"/register-node", "t"); c.sibling("node-heartbeat") != DefaultAPI+"/node-heartbeat" {
		t.Errorf("a garage connected before v1.7.0 must report to the own address, got %s", c.sibling("node-heartbeat"))
	}
}

// The own address in front, the legacy one behind: Door reaches the portal when the own name
// answers, when its proxy answers 502, and when nothing listens there at all.
func TestSendFallsBackToLegacy(t *testing.T) {
	var ownHits, legacyHits int
	ownStatus := http.StatusOK
	own := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ownHits++
		w.WriteHeader(ownStatus)
		w.Write([]byte(`{"ok":true,"via":"own"}`))
	}))
	legacy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legacyHits++
		if r.Header.Get("Authorization") != "Bearer tok" || r.URL.Path != "/functions/v1/node-heartbeat" {
			t.Errorf("fallback request lost its token or path: %q %s", r.Header.Get("Authorization"), r.URL.Path)
		}
		w.Write([]byte(`{"ok":true,"via":"legacy"}`))
	}))
	defer legacy.Close()
	oldOwn, oldLegacy := ownAPI, legacyAPI
	ownAPI, legacyAPI = own.URL+"/functions/v1", legacy.URL+"/functions/v1"
	defer func() { ownAPI, legacyAPI = oldOwn, oldLegacy }()

	c := New(legacyAPI+"/register-node", "tok") // as saved by an older Door
	post := func() string {
		var out struct{ Via string }
		if _, err := c.post(context.Background(), c.sibling("node-heartbeat"), map[string]int{"a": 1}, &out, 5*time.Second); err != nil {
			t.Fatalf("post: %v", err)
		}
		return out.Via
	}
	if via := post(); via != "own" || ownHits != 1 || legacyHits != 0 {
		t.Fatalf("own address answers: via %q, own %d, legacy %d", via, ownHits, legacyHits)
	}
	ownStatus = http.StatusBadGateway
	if via := post(); via != "legacy" || legacyHits != 1 {
		t.Fatalf("own address answers 502: via %q, legacy %d", via, legacyHits)
	}
	ownStatus = http.StatusUnauthorized // the portal's own answer: never retried elsewhere
	if _, err := c.post(context.Background(), c.sibling("node-heartbeat"), nil, nil, 5*time.Second); err == nil || legacyHits != 1 {
		t.Fatalf("a 401 from the portal must not fall back: err %v, legacy %d", err, legacyHits)
	}
	own.Close()
	if via := post(); via != "legacy" || legacyHits != 2 {
		t.Fatalf("nothing listens on the own address: via %q, legacy %d", via, legacyHits)
	}
}

func TestSendNoFallbackForOtherAddresses(t *testing.T) {
	var hits int
	legacy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits++ }))
	defer legacy.Close()
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadGateway) }))
	defer other.Close()
	oldLegacy := legacyAPI
	legacyAPI = legacy.URL + "/functions/v1"
	defer func() { legacyAPI = oldLegacy }()
	c := New(other.URL+"/functions/v1/register-node", "tok")
	if _, err := c.post(context.Background(), c.RegisterURL, nil, nil, 5*time.Second); err == nil || hits != 0 {
		t.Fatalf("an operator's own portal address must not fall back to GarageAI's: err %v, hits %d", err, hits)
	}
}

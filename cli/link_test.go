package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/magnusfroste/garageai/cli/internal/profile"
)

func TestSuggestName(t *testing.T) {
	for in, want := range map[string]string{
		"Magnus’s MacBook Air": "macbook-air-ab12", "Mac mini som tillhör Erik Hedin": "mac-mini-ab12",
		"dgx-city": "dgx-city", "Kontorets Mac Studio": "kontorets-mac-studio", "Ärlig Örn": "arlig-orn",
		"A": "", "  --x y--  ": "x-y",
	} {
		if got := suggestName(in, "m_ab12cd"); got != want {
			t.Errorf("suggestName(%q) = %q, want %q", in, got, want)
		}
	}
}

// No runtime: link stops before contacting GarageAI.
func TestLinkStopsWithoutRuntime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Fatal("contacted the portal") }))
	defer srv.Close()
	t.Setenv("GARAGEAI_PORTAL_API", srv.URL)
	runtimeCheck = func(profile.Profile, int) bool { return false }
	defer func() { runtimeCheck = hasRuntimeOn }()
	o := Options{Name: "h", Runtime: "vllm", Port: 8000}
	if code := link(&o, true); code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
}

func TestPickRuntime(t *testing.T) {
	p := profile.Profile{Runtimes: []profile.Runtime{
		{Kind: "llamacpp", Port: 8080},
		{Kind: "ollama", Port: 11434, Network: true, Models: []profile.Model{{ID: "qwen3:4b"}}},
		{Kind: "myserver", Port: 9000, Models: []profile.Model{{ID: "x"}}},
	}}
	if k, port, ok := pickRuntime(p); !ok || k != "ollama" || port != 11434 {
		t.Fatalf("got %s %d %v", k, port, ok)
	}
	p.Runtimes = p.Runtimes[2:]
	if k, _, _ := pickRuntime(p); k != "other" {
		t.Fatalf("unknown kind should be other, got %s", k)
	}
}

// The whole link: start, pending, approved; then the options and the environment for the root
// half carry the garage's credentials.
func TestLinkFlow(t *testing.T) {
	var polls int32
	var started map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bridge-link-start":
			json.NewDecoder(r.Body).Decode(&started)
			w.Write([]byte(`{"device_code":"gbd_x","user_code":"ABCD-EFGH","verification_url":"https://app/bridge/link","verification_url_complete":"https://app/bridge/link?code=ABCD-EFGH","interval":1,"expires_in":30}`))
		case "/bridge-link-poll":
			if atomic.AddInt32(&polls, 1) < 2 {
				w.Write([]byte(`{"status":"pending"}`))
				return
			}
			w.Write([]byte(`{"status":"approved","garage":{"id":"g1","name":"garage-m2"},"register_url":"https://p/functions/v1/register-node","register_token":"grg_t","setup_key":"SK","management_url":"https://nb"}`))
		}
	}))
	defer srv.Close()
	t.Setenv("GARAGEAI_PORTAL_API", srv.URL)
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	for _, k := range []string{"GARAGEAI_REGISTER_TOKEN", "GARAGEAI_SETUP_KEY", "GARAGEAI_NODE_NAME", "GARAGEAI_REGISTER_URL",
		"GARAGEAI_MANAGEMENT_URL", "GARAGEAI_RUNTIME", "GARAGEAI_PORT"} {
		t.Setenv(k, "")
	}
	runtimeCheck = func(profile.Profile, int) bool { return true }
	defer func() { runtimeCheck = hasRuntimeOn }()
	o := Options{Name: "testhost", Runtime: "vllm", Port: 8000}
	if code := link(&o, true); code != 0 {
		t.Fatalf("link exit %d", code)
	}
	if o.Name != "garage-m2" || o.RegisterToken != "grg_t" || o.SetupKey != "SK" || o.Runtime != "vllm" {
		t.Fatalf("options: %+v", o)
	}
	if os.Getenv("GARAGEAI_REGISTER_TOKEN") != "grg_t" || os.Getenv("GARAGEAI_PORT") != "8000" {
		t.Fatal("environment for the root half not set")
	}
	// On a Mac the computer's name (scutil) is used instead of the hostname.
	mid, _ := started["machine_id"].(string)
	if started["suggested_name"] != suggestName(computerName("testhost"), mid) || started["bridge_version"] == nil {
		t.Fatalf("start request: %v", started)
	}
}

package profile

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestParseProcNet(t *testing.T) {
	v4 := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:1F40 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 111 1 0000000000000000 100 0 0 10 0
   1: 00000000:2CAA 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 222 1 0000000000000000 100 0 0 10 0
   2: 0100007F:1F41 0100007F:9C40 01 00000000:00000000 00:00000000 00000000  1000        0 333 1 0000000000000000 100 0 0 10 0`
	got := parseProcNet(v4, false)
	if l := got["111"]; l.Port != 8000 || l.Addr != "127.0.0.1" {
		t.Errorf("127.0.0.1:8000 parsed as %+v", l)
	}
	if l := got["222"]; l.Port != 11434 || l.Addr != "0.0.0.0" {
		t.Errorf("0.0.0.0:11434 parsed as %+v", l)
	}
	if _, ok := got["333"]; ok {
		t.Error("an established connection was listed as listening")
	}
	v6 := `  sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000000000000000000000000000:04D2 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 444 1
   1: 00000000000000000000000001000000:1F90 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 555 1`
	got6 := parseProcNet(v6, true)
	if l := got6["444"]; l.Port != 1234 || l.Addr != "::" {
		t.Errorf("[::]:1234 parsed as %+v", l)
	}
	if l := got6["555"]; l.Port != 8080 || l.Addr != "::1" {
		t.Errorf("[::1]:8080 parsed as %+v", l)
	}
}

func TestParseLsof(t *testing.T) {
	ls := parseLsof("p501\ncollama\nn127.0.0.1:11434\np777\ncLM Studio\nn*:1234\nn[::]:1234\n")
	if len(ls) != 3 || ls[0].Process != "ollama" || ls[0].Port != 11434 || ls[1].Addr != "*" || ls[2].Addr != "::" || ls[2].PID != 777 {
		t.Errorf("lsof parsed as %+v", ls)
	}
}

func TestParseNetstat(t *testing.T) {
	out := `
Active Connections

  Proto  Local Address          Foreign Address        State           PID
  TCP    0.0.0.0:11434          0.0.0.0:0              LISTENING       4242
  TCP    127.0.0.1:1234         0.0.0.0:0              LISTENING       99
  TCP    192.168.1.5:50000      1.2.3.4:443            ESTABLISHED     12
  TCP    [::]:8080              [::]:0                 LISTENING       7`
	ls := parseNetstat(out)
	if len(ls) != 3 || ls[0].Port != 11434 || ls[0].PID != 4242 || ls[1].Addr != "127.0.0.1" || ls[2].Addr != "::" {
		t.Errorf("netstat parsed as %+v", ls)
	}
}

func fakeRuntime(t *testing.T, ollama bool) (port int) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/version" && ollama:
			w.Write([]byte(`{"version":"0.12.0"}`))
		case r.URL.Path == "/api/ps" && ollama:
			w.Write([]byte(`{"models":[{"name":"qwen3:4b","context_length":4096}]}`))
		case r.URL.Path == "/v1/models":
			w.Write([]byte(`{"data":[{"id":"qwen3:4b","max_model_len":32768,"owned_by":"vllm"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	p, _ := strconv.Atoi(srv.URL[strings.LastIndex(srv.URL, ":")+1:])
	return p
}

func TestProbeAndProblems(t *testing.T) {
	port := fakeRuntime(t, false)
	rt := probe(port, []string{"127.0.0.1"}, Listener{})
	if rt == nil || rt.Kind != "vllm" || rt.Network || len(rt.Models) != 1 || *rt.Models[0].Context != 32768 {
		t.Fatalf("vLLM-like runtime probed as %+v", rt)
	}
	p := Profile{Runtimes: []Runtime{*rt}, NetBird: NetBird{Installed: true, Connected: true}, GPUs: []GPU{{Name: "x"}}}
	codes := map[string]bool{}
	for _, pr := range problems(p) {
		codes[pr.Code] = true
	}
	if !codes["localhost_only"] || codes["no_runtime"] {
		t.Errorf("problems for a localhost-only runtime: %v", codes)
	}
	rt.Binds, rt.Network = []string{"0.0.0.0"}, true
	p.Runtimes = []Runtime{*rt}
	for _, pr := range problems(p) {
		if pr.Code == "localhost_only" {
			t.Error("a network bind is reported as localhost-only")
		}
	}
}

func TestOllamaSmallContext(t *testing.T) {
	port := fakeRuntime(t, true)
	rt := probe(port, []string{"0.0.0.0"}, Listener{})
	if rt == nil || rt.Kind != "ollama" || len(rt.OllamaLoaded) != 1 {
		t.Fatalf("Ollama probed as %+v", rt)
	}
	mem := 16
	p := Profile{Machine: Machine{MemoryGB: &mem}, Runtimes: []Runtime{*rt}, NetBird: NetBird{Installed: true, Connected: true}}
	found := false
	for _, pr := range problems(p) {
		found = found || (pr.Code == "small_context" && strings.Contains(pr.Fix, "16384"))
	}
	if !found {
		t.Error("a 4096-token Ollama window on a 16 GB machine is not flagged")
	}
}

func TestFlagsNeverHoldTheKey(t *testing.T) {
	f := readFlags([]string{"vllm", "serve", "m", "--host", "0.0.0.0", "--api-key", "sk-secret", "--max-model-len=262144"})
	if f.APIKeySet == nil || !*f.APIKeySet || *f.MaxModelLen != 262144 || *f.Host != "0.0.0.0" || *f.PromptTokensDetails {
		t.Errorf("flags read as %+v", f)
	}
}

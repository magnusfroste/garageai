package profile

import (
	"net"
	"os/exec"
	goruntime "runtime"
	"strconv"
	"testing"
	"time"
)

func TestKeyFromArgsAndEnv(t *testing.T) {
	if k := keyFromArgs([]string{"python3", "-m", "vllm.entrypoints.openai.api_server", "--api-key", "sk-x", "--port", "8000"}); k != "sk-x" {
		t.Fatalf("args: %q", k)
	}
	if k := keyFromArgs([]string{"vllm", "serve", "m", "--api-key=sk-y"}); k != "sk-y" {
		t.Fatalf("args=: %q", k)
	}
	if k := keyFromArgs([]string{"ollama", "serve"}); k != "" {
		t.Fatalf("no key: %q", k)
	}
	if k := keyFromEnv([]string{"PATH=/bin", "VLLM_API_KEY='sk-z'", "HOME=/x"}); k != "sk-z" {
		t.Fatalf("env: %q", k)
	}
	if k := keyFromEnv([]string{"PATH=/bin"}); k != "" {
		t.Fatalf("no env key: %q", k)
	}
}

func TestParentOf(t *testing.T) {
	if ppid, _ := parentOf(1); ppid != 0 {
		t.Fatalf("pid 1 has parent %d", ppid)
	}
	h := howRun(1)
	if h.Terminal != "" {
		t.Fatalf("pid 1 in a terminal: %+v", h)
	}
}

// A real listening process started with --api-key: FindRuntimeKey reads it off its command line.
func TestFindRuntimeKeyFromProcess(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("no cmdline on windows")
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("no python3")
	}
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	cmd := exec.Command("python3", "-c", "import socket,sys,time\ns=socket.socket()\ns.bind(('127.0.0.1',int(sys.argv[1])))\ns.listen(1)\ntime.sleep(60)", strconv.Itoa(port), "--api-key", "sk-from-args")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if c, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), 200*time.Millisecond); err == nil {
			c.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the test process never listened")
		}
		time.Sleep(100 * time.Millisecond)
	}
	key, source := FindRuntimeKey(port)
	if key != "sk-from-args" {
		t.Fatalf("key %q from %q", key, source)
	}
}

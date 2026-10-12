package profile

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// FindRuntimeKey looks for the API key a runtime on port was started with, so the operator does not
// have to paste it: on its command line (--api-key), in its environment (VLLM_API_KEY, API_KEY,
// LLAMA_API_KEY, SGLANG_API_KEY) and, when the listener is Docker's proxy, in the container's
// command and environment. Needs root for another user's process; connect has it. Returns the key
// and where it was found, or "" and "".
func FindRuntimeKey(port int) (key, source string) {
	_, _, owner := byPort(listeners())
	l, ok := owner[port]
	if !ok || l.PID == 0 {
		return "", ""
	}
	if k := keyFromArgs(cmdline(l.PID)); k != "" {
		return k, l.Process + " command line"
	}
	if k := keyFromEnv(processEnv(l.PID)); k != "" {
		return k, l.Process + " environment"
	}
	p := strings.ToLower(l.Process)
	if strings.Contains(p, "docker") || strings.Contains(p, "containerd") || strings.Contains(p, "podman") {
		if id := containerOnPort(port); id != "" {
			out := run("docker", "inspect", "-f", `{{join .Config.Cmd "\n"}}{{"\n"}}{{join .Config.Entrypoint "\n"}}{{"\n"}}{{join .Config.Env "\n"}}`, id)
			lines := strings.Split(out, "\n")
			if k := keyFromArgs(lines); k != "" {
				return k, "container " + id[:12] + " command"
			}
			if k := keyFromEnv(lines); k != "" {
				return k, "container " + id[:12] + " environment"
			}
		}
	}
	return "", ""
}

func keyFromArgs(args []string) string {
	for i, a := range args {
		a = strings.TrimSpace(a)
		if a == "--api-key" && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if v, ok := strings.CutPrefix(a, "--api-key="); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

var keyEnvNames = []string{"VLLM_API_KEY", "SGLANG_API_KEY", "LLAMA_API_KEY", "API_KEY", "OPENAI_API_KEY"}

func keyFromEnv(env []string) string {
	for _, name := range keyEnvNames {
		for _, e := range env {
			if v, ok := strings.CutPrefix(strings.TrimSpace(e), name+"="); ok && v != "" {
				return strings.Trim(v, `"'`)
			}
		}
	}
	return ""
}

var portMap = regexp.MustCompile(`:(\d+)->`)

// containerOnPort: the id of the Docker container that publishes port, or "".
func containerOnPort(port int) string {
	for _, line := range strings.Split(run("docker", "ps", "--no-trunc", "--format", "{{.ID}}\t{{.Ports}}"), "\n") {
		id, ports, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		for _, m := range portMap.FindAllStringSubmatch(ports, -1) {
			if n, _ := strconv.Atoi(m[1]); n == port {
				return strings.TrimSpace(id)
			}
		}
	}
	return ""
}

// processEnv is the environment of a running process (root for another user's), or nil.
func processEnv(pid int) []string {
	if b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/environ"); err == nil {
		return strings.Split(strings.TrimRight(string(b), "\x00"), "\x00")
	}
	// macOS: ps prints the environment after the command with the e flag.
	out := run("ps", "eww", "-o", "command=", "-p", strconv.Itoa(pid))
	return strings.Fields(out)
}

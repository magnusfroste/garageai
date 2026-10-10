package profile

import (
	"bufio"
	"encoding/hex"
	"net"
	"sort"
	"strconv"
	"strings"
)

// Listener is one listening TCP socket: port, local address, and the owning process if known.
type Listener struct {
	Port    int
	Addr    string
	PID     int
	Process string
}

// parseProcNet reads /proc/net/tcp or /proc/net/tcp6 and returns listening sockets with their
// inode (to be matched to a process). Addresses are little-endian hex, per 32-bit word.
func parseProcNet(data string, v6 bool) map[string]Listener {
	out := map[string]Listener{}
	sc := bufio.NewScanner(strings.NewReader(data))
	sc.Scan() // header
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 10 || f[3] != "0A" { // 0A = LISTEN
			continue
		}
		hostport := strings.SplitN(f[1], ":", 2)
		if len(hostport) != 2 {
			continue
		}
		port, err := strconv.ParseUint(hostport[1], 16, 16)
		if err != nil {
			continue
		}
		raw, err := hex.DecodeString(hostport[0])
		if err != nil || (len(raw) != 4 && len(raw) != 16) {
			continue
		}
		for i := 0; i+4 <= len(raw); i += 4 { // each 32-bit word is little-endian
			raw[i], raw[i+1], raw[i+2], raw[i+3] = raw[i+3], raw[i+2], raw[i+1], raw[i]
		}
		ip := net.IP(raw)
		addr := ip.String()
		if v6 && ip.Equal(net.IPv6zero) {
			addr = "::"
		}
		out[f[9]] = Listener{Port: int(port), Addr: addr}
	}
	return out
}

// parseLsof reads `lsof -nP -iTCP -sTCP:LISTEN -F pcn` (macOS): p<pid>, c<command>, n<addr:port>.
func parseLsof(data string) []Listener {
	var out []Listener
	pid, cmd := 0, ""
	for _, line := range strings.Split(data, "\n") {
		if line == "" {
			continue
		}
		switch line[0] {
		case 'p':
			pid, _ = strconv.Atoi(line[1:])
		case 'c':
			cmd = line[1:]
		case 'n':
			addr, port := splitHostPort(line[1:])
			if port > 0 {
				out = append(out, Listener{Port: port, Addr: addr, PID: pid, Process: cmd})
			}
		}
	}
	return out
}

// parseNetstat reads Windows `netstat -ano -p TCP` / `-p TCPv6` lines in the LISTENING state.
func parseNetstat(data string) []Listener {
	var out []Listener
	for _, line := range strings.Split(data, "\n") {
		f := strings.Fields(line)
		if len(f) < 5 || !strings.EqualFold(f[0], "TCP") || f[3] != "LISTENING" {
			continue
		}
		addr, port := splitHostPort(f[1])
		pid, _ := strconv.Atoi(f[4])
		if port > 0 {
			out = append(out, Listener{Port: port, Addr: addr, PID: pid})
		}
	}
	return out
}

// splitHostPort handles "127.0.0.1:8000", "*:11434", "[::]:8080" and "[::1]:1234".
func splitHostPort(s string) (string, int) {
	i := strings.LastIndex(s, ":")
	if i < 0 {
		return "", 0
	}
	port, err := strconv.Atoi(s[i+1:])
	if err != nil {
		return "", 0
	}
	host := strings.Trim(s[:i], "[]")
	if j := strings.Index(host, "%"); j >= 0 {
		host = host[:j]
	}
	return host, port
}

// byPort groups listeners: port -> sorted unique addresses, and the first known process.
func byPort(ls []Listener) (ports []int, addrs map[int][]string, owner map[int]Listener) {
	addrs, owner = map[int][]string{}, map[int]Listener{}
	seen := map[int]map[string]bool{}
	for _, l := range ls {
		if seen[l.Port] == nil {
			seen[l.Port] = map[string]bool{}
			ports = append(ports, l.Port)
		}
		if !seen[l.Port][l.Addr] {
			seen[l.Port][l.Addr] = true
			addrs[l.Port] = append(addrs[l.Port], l.Addr)
		}
		if o, ok := owner[l.Port]; !ok || (o.Process == "" && l.Process != "") {
			owner[l.Port] = l
		}
	}
	sort.Ints(ports)
	for p := range addrs {
		sort.Strings(addrs[p])
	}
	return ports, addrs, owner
}

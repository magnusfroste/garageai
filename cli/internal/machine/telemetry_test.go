package machine

import "testing"

func TestLoadPct(t *testing.T) {
	if got := loadPct("{ 2.50 1.10 0.98 }", 10); got == nil || *got != 25 {
		t.Fatalf("macOS loadavg: %v", got)
	}
	if got := loadPct("0.40 0.30 0.20 1/300 12345", 4); got == nil || *got != 10 {
		t.Fatalf("linux loadavg: %v", got)
	}
	if loadPct("", 4) != nil || loadPct("1.0", 0) != nil {
		t.Fatal("unknown load must be nil")
	}
}

func TestDarwinMemory(t *testing.T) {
	vmstat := "Mach Virtual Memory Statistics: (page size of 16384 bytes)\nPages free:                               10000.\nPages active:                             100000.\nPages inactive:                           20000.\nPages speculative:                         5000.\nPages purgeable:                           5000.\n"
	total, free, ok := darwinMemory("8589934592\n", vmstat)
	if !ok || total != 8192 || free != 40000*16384>>20 {
		t.Fatalf("got %d %d %v", total, free, ok)
	}
}

func TestLinuxMemory(t *testing.T) {
	total, free, ok := linuxMemory("MemTotal:       16000000 kB\nMemFree:         1000000 kB\nMemAvailable:    9000000 kB\n")
	if !ok || total != 16000000>>10 || free != 9000000>>10 {
		t.Fatalf("got %d %d %v", total, free, ok)
	}
}

func TestParsePmsetBatt(t *testing.T) {
	p, pct := parsePmsetBatt("Now drawing from 'Battery Power'\n -InternalBattery-0 (id=1234)\t87%; discharging; 3:10 remaining present: true\n")
	if p != "battery" || pct == nil || *pct != 87 {
		t.Fatalf("laptop on battery: %q %v", p, pct)
	}
	p, pct = parsePmsetBatt("Now drawing from 'AC Power'\n -InternalBattery-0 (id=1234)\t100%; charged; 0:00 remaining present: true\n")
	if p != "ac" || pct == nil || *pct != 100 {
		t.Fatalf("laptop on AC: %q %v", p, pct)
	}
	if p, pct = parsePmsetBatt("Now drawing from 'AC Power'\n"); p != "" || pct != nil {
		t.Fatalf("desktop: %q %v", p, pct)
	}
}

func TestParseHIDIdle(t *testing.T) {
	if s, ok := parseHIDIdle(`    |   "HIDIdleTime" = 125000000000`); !ok || s != 125 {
		t.Fatalf("got %d %v", s, ok)
	}
	if _, ok := parseHIDIdle("nothing"); ok {
		t.Fatal("no field must be !ok")
	}
}

func TestLinuxPower(t *testing.T) {
	p, pct := linuxPower(map[string]map[string]string{"BAT0": {"type": "Battery\n", "status": "Discharging\n", "capacity": "55\n"}, "AC": {"type": "Mains\n", "online": "0\n"}})
	if p != "battery" || pct == nil || *pct != 55 {
		t.Fatalf("laptop on battery: %q %v", p, pct)
	}
	p, _ = linuxPower(map[string]map[string]string{"BAT0": {"type": "Battery\n", "status": "Charging\n", "capacity": "55\n"}, "AC": {"type": "Mains\n", "online": "1\n"}})
	if p != "ac" {
		t.Fatalf("laptop charging: %q", p)
	}
	if p, _ = linuxPower(map[string]map[string]string{"AC": {"type": "Mains\n", "online": "1\n"}}); p != "" {
		t.Fatalf("server without battery: %q", p)
	}
}

func TestParseNvidia(t *testing.T) {
	g, ok := parseNvidia("37, 12034, 24576, 61, 142.50\n")
	if !ok || g.util != 37 || g.usedMB != 12034 || g.totalMB != 24576 || g.tempC != 61 || g.powerW != 142 || g.unified {
		t.Fatalf("got %+v %v", g, ok)
	}
	// DGX Spark / GB10: unified memory, nvidia-smi answers [N/A] for memory.
	g, ok = parseNvidia("0, [N/A], [N/A], 47, 13.58\n")
	if !ok || !g.unified || g.util != 0 || g.tempC != 47 || g.powerW != 13 || g.usedMB != 0 {
		t.Fatalf("spark: %+v %v", g, ok)
	}
	if _, ok := parseNvidia(""); ok {
		t.Fatal("no nvidia-smi must be !ok")
	}
}

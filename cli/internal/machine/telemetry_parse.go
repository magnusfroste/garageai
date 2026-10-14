package machine

import (
	"regexp"
	"strconv"
	"strings"
)

// Pure parsers for the telemetry sources, so they are tested on every platform.

var loadRe = regexp.MustCompile(`([0-9]+\.[0-9]+)`)

// loadPct: "{ 1.23 1.10 0.98 }" (macOS sysctl) or "1.23 1.10 0.98 ..." (/proc/loadavg) → the
// 1-minute load as a share of the CPUs, in percent.
func loadPct(out string, cpus int) *int {
	m := loadRe.FindStringSubmatch(out)
	if m == nil || cpus <= 0 {
		return nil
	}
	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return nil
	}
	return intp(int(f / float64(cpus) * 100))
}

var vmStatRe = regexp.MustCompile(`Pages (free|inactive|speculative|purgeable):\s+(\d+)`)
var pageSizeRe = regexp.MustCompile(`page size of (\d+) bytes`)

// darwinMemory: total from hw.memsize; "free" as free + inactive + speculative + purgeable pages
// (what a model could take without swapping).
func darwinMemory(memsize, vmstat string) (totalMB, freeMB int, ok bool) {
	total, err := strconv.ParseInt(strings.TrimSpace(memsize), 10, 64)
	if err != nil || total <= 0 {
		return 0, 0, false
	}
	page := int64(4096)
	if m := pageSizeRe.FindStringSubmatch(vmstat); m != nil {
		page, _ = strconv.ParseInt(m[1], 10, 64)
	}
	var pages int64
	for _, m := range vmStatRe.FindAllStringSubmatch(vmstat, -1) {
		n, _ := strconv.ParseInt(m[2], 10, 64)
		pages += n
	}
	return int(total >> 20), int(pages * page >> 20), true
}

var meminfoRe = regexp.MustCompile(`(?m)^(MemTotal|MemAvailable):\s+(\d+) kB`)

// linuxMemory from /proc/meminfo: MemTotal and MemAvailable.
func linuxMemory(meminfo string) (totalMB, freeMB int, ok bool) {
	vals := map[string]int64{}
	for _, m := range meminfoRe.FindAllStringSubmatch(meminfo, -1) {
		n, _ := strconv.ParseInt(m[2], 10, 64)
		vals[m[1]] = n
	}
	if vals["MemTotal"] == 0 {
		return 0, 0, false
	}
	return int(vals["MemTotal"] >> 10), int(vals["MemAvailable"] >> 10), true
}

var battRe = regexp.MustCompile(`(\d+)%`)

// parsePmsetBatt (macOS): "Now drawing from 'AC Power'" / "'Battery Power'", and
// "-InternalBattery-0 (id=...)	87%; charging; ...". A desktop has no InternalBattery line:
// then the distinction does not apply and power is "".
func parsePmsetBatt(out string) (power string, pct *int) {
	switch {
	case strings.Contains(out, "'Battery Power'"):
		power = "battery"
	case strings.Contains(out, "'AC Power'"):
		power = "ac"
	}
	if strings.Contains(out, "InternalBattery") {
		if m := battRe.FindStringSubmatch(out); m != nil {
			n, _ := strconv.Atoi(m[1])
			pct = intp(n)
		}
	} else if power == "ac" {
		power = ""
	}
	return power, pct
}

var hidIdleRe = regexp.MustCompile(`"HIDIdleTime"\s*=\s*(\d+)`)

// parseHIDIdle (macOS ioreg): seconds since the last keyboard or mouse event (reported in ns).
func parseHIDIdle(out string) (seconds int, ok bool) {
	m := hidIdleRe.FindStringSubmatch(out)
	if m == nil {
		return 0, false
	}
	ns, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return int(ns / 1e9), true
}

// linuxPower from /sys/class/power_supply: "ac" when a Mains/USB supply is online or no battery
// exists on a machine with a supply, "battery" when a battery discharges, "" when nothing is known.
func linuxPower(supplies map[string]map[string]string) (power string, pct *int) {
	hasBattery := false
	for _, s := range supplies {
		switch strings.TrimSpace(s["type"]) {
		case "Battery":
			hasBattery = true
			if strings.TrimSpace(s["status"]) == "Discharging" {
				power = "battery"
			}
			if n, err := strconv.Atoi(strings.TrimSpace(s["capacity"])); err == nil {
				pct = intp(n)
			}
		case "Mains", "USB", "USB_PD":
			if strings.TrimSpace(s["online"]) == "1" && power == "" {
				power = "ac"
			}
		}
	}
	if !hasBattery {
		return "", nil
	}
	if power == "" {
		power = "ac" // a battery that is not discharging: charging or full
	}
	return power, pct
}

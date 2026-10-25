package service

import "testing"

func TestParseRunnerCmdline(t *testing.T) {
	unit := "[Service]\nExecStart=/usr/local/lib/garageai/runner -hf ibm-granite/granite-4.1-3b-GGUF:Q8_0 --serve --port 8080 --ttl 0 -c 16384\n"
	s, ok := ParseRunnerCmdline(unit)
	if !ok || s.Model != "ibm-granite/granite-4.1-3b-GGUF:Q8_0" || s.Port != 8080 || s.Context != 16384 {
		t.Fatalf("unit: %+v %v", s, ok)
	}
	win := "\"C:\\ProgramData\\GarageAI\\runner.exe\" -hf a/b:Q4_K_M --serve --port 8090 --ttl 0 >> \"x.log\""
	if s, ok := ParseRunnerCmdline(win); !ok || s.Model != "a/b:Q4_K_M" || s.Port != 8090 || s.Context != 0 {
		t.Fatalf("windows: %+v %v", s, ok)
	}
	if _, ok := ParseRunnerCmdline("nothing here"); ok {
		t.Fatal("no -hf must be !ok")
	}
}

func TestParseRunnerVersion(t *testing.T) {
	for out, want := range map[string]string{
		"runner 1.4.2\n":                "v1.4.2",
		"runner 1.6.0 (cuda)\n":         "v1.6.0",
		"runner v1.5.1\n":               "v1.5.1",
		"runner 1.6.0, built today\n":   "v1.6.0",
		"runner\n":                      "",
		"usage: runner [options] 1.2\n": "",
		"":                              "",
	} {
		if got := parseRunnerVersion(out); got != want {
			t.Errorf("parseRunnerVersion(%q) = %q, want %q", out, got, want)
		}
	}
}

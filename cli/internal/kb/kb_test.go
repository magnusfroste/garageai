package kb

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The guide lives in docs/; the binary carries a copy. They must not drift apart.
func TestCopyMatchesDocs(t *testing.T) {
	b, err := os.ReadFile("../../../docs/garage-troubleshooting.md")
	if err != nil {
		t.Skip("docs/ not available:", err)
	}
	if string(b) != Text {
		t.Fatal("cli/internal/kb/troubleshooting.md differs from docs/garage-troubleshooting.md: cp docs/garage-troubleshooting.md cli/internal/kb/troubleshooting.md")
	}
}

// Every problem doctor can report has a section in the guide.
func TestEveryProblemCodeIsExplained(t *testing.T) {
	b, err := os.ReadFile("../profile/build.go")
	if err != nil {
		t.Fatal(err)
	}
	codes := regexp.MustCompile(`add\("(?:error|warning|info)", "([a-z_]+)"`).FindAllStringSubmatch(string(b), -1)
	if len(codes) < 10 {
		t.Fatalf("found only %d codes in build.go", len(codes))
	}
	for _, m := range codes {
		if len(Explain(m[1])) == 0 {
			t.Errorf("no section for %s in docs/garage-troubleshooting.md", m[1])
		}
	}
}

func TestSearch(t *testing.T) {
	got := Search("my mac studio goes to sleep", 3)
	if len(got) == 0 || got[0].Code != "mac_sleeps" {
		t.Fatalf("search: %+v", got)
	}
	if s := Explain("heartbeat_stopped"); len(s) != 1 || !strings.Contains(s[0].Body, "connect command") {
		t.Fatalf("explain heartbeat_stopped: %+v", s)
	}
}

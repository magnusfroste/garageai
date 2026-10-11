package machine

import (
	"regexp"
	"testing"
)

func TestIDIsStableAndHashed(t *testing.T) {
	a, b := ID(), ID()
	if a != b {
		t.Fatalf("not stable: %s vs %s", a, b)
	}
	if a == "" {
		t.Skip("no machine id on this system")
	}
	if !regexp.MustCompile(`^m_[0-9a-f]{32}$`).MatchString(a) {
		t.Fatalf("format: %s", a)
	}
	if raw := rawID(); raw != "" && regexp.MustCompile(regexp.QuoteMeta(raw)).MatchString(a) {
		t.Fatal("raw id leaked")
	}
}

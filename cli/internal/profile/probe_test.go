package profile

import "testing"


func TestKindOfXyntetikRunner(t *testing.T) {
	for s, want := range map[string]string{"runner": "xyntetik", "/usr/local/bin/runner": "xyntetik", "runner -m x.gguf --serve": "xyntetik", "xyntetik-runner": "xyntetik", "gitlab-runner": "", "vllm": "vllm"} {
		if got := kindOf(s); got != want && !(want == "" && got != "xyntetik") {
			t.Errorf("kindOf(%q) = %q, want %q", s, got, want)
		}
	}
}

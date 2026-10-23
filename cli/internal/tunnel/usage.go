package tunnel

import (
	"io"
	"regexp"
	"strconv"
	"sync/atomic"
)

// inFlight counts requests the forward is serving right now: the energy sampler counts the GPU's
// power as busy only while it is above zero.
var inFlight atomic.Int64

// InFlight reports how many requests are being served now.
func InFlight() int64 { return inFlight.Load() }

// headCapture passes a request body through unchanged and keeps its first KB, where the JSON's
// "model" is (OpenAI, Anthropic and Ollama bodies all put it near the start).
type headCapture struct {
	r   io.ReadCloser
	buf []byte
}

func (h *headCapture) Read(p []byte) (int, error) {
	n, err := h.r.Read(p)
	if room := 1024 - len(h.buf); room > 0 && n > 0 {
		k := n
		if k > room {
			k = room
		}
		h.buf = append(h.buf, p[:k]...)
	}
	return n, err
}

func (h *headCapture) Close() error { return h.r.Close() }

var modelRe = regexp.MustCompile(`"model"\s*:\s*"([^"]{1,200})"`)

// requestModel is the "model" of a JSON request body's first KB, or "".
func requestModel(head []byte) string {
	if m := modelRe.FindSubmatch(head); m != nil {
		return string(m[1])
	}
	return ""
}

var (
	inRes  = []*regexp.Regexp{regexp.MustCompile(`"prompt_tokens"\s*:\s*(\d+)`), regexp.MustCompile(`"input_tokens"\s*:\s*(\d+)`), regexp.MustCompile(`"prompt_eval_count"\s*:\s*(\d+)`)}
	outRes = []*regexp.Regexp{regexp.MustCompile(`"completion_tokens"\s*:\s*(\d+)`), regexp.MustCompile(`"output_tokens"\s*:\s*(\d+)`), regexp.MustCompile(`"eval_count"\s*:\s*(\d+)`)}
)

// usageTokens reads the prompt and generated token counts the runtime reported, from the
// response's first and last KB: the last mention wins (a stream's final usage), the start is the
// fallback (Anthropic's input_tokens arrive in message_start). 0 when the runtime reported none.
func usageTokens(head, tail []byte) (in, out int) {
	return lastNumber(inRes, tail, head), lastNumber(outRes, tail, head)
}

func lastNumber(res []*regexp.Regexp, bufs ...[]byte) int {
	for _, b := range bufs {
		for _, re := range res {
			if all := re.FindAllSubmatch(b, -1); len(all) > 0 {
				if n, err := strconv.Atoi(string(all[len(all)-1][1])); err == nil {
					return n
				}
			}
		}
	}
	return 0
}

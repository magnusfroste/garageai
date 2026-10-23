package energy

import (
	"math"
	"sort"
	"time"

	"github.com/magnusfroste/garageai/cli/internal/tunnel"
)

// Summary is energy set against tokens for a window: what the heartbeat reports and the menu-bar
// icon shows. Wh are watt-hours. TokensPerKWh uses the busy energy (while requests were served),
// so an idle machine does not make a model look worse; IdleWh says what idling cost.
type Summary struct {
	WindowMin    int           `json:"window_min"`
	Source       string        `json:"source,omitempty"` // nvidia-smi, powermetrics
	Wh           float64       `json:"wh"`
	BusyWh       float64       `json:"busy_wh"`
	IdleWh       float64       `json:"idle_wh"`
	AvgW         float64       `json:"avg_w"`
	PeakW        float64       `json:"peak_w,omitempty"`
	InTokens     int           `json:"in_tokens"`
	OutTokens    int           `json:"out_tokens"`
	TokensPerKWh *int          `json:"tokens_per_kwh,omitempty"` // generated tokens per busy kWh
	Models       []ModelEnergy `json:"models,omitempty"`
}

// ModelEnergy splits the busy energy over models by their share of generated tokens.
type ModelEnergy struct {
	Model        string  `json:"model"`
	OutTokens    int     `json:"out_tokens"`
	BusyWh       float64 `json:"busy_wh"`
	TokensPerKWh *int    `json:"tokens_per_kwh,omitempty"`
}

// Summarize sets the minutes from since to now against the request samples of the same period.
// nil when nothing was measured.
func Summarize(mins []Minute, samples []tunnel.Sample, since, now time.Time) *Summary {
	s := Summary{WindowMin: int(math.Round(now.Sub(since).Minutes()))}
	var seconds float64
	for _, m := range mins {
		if m.At.Before(since.UTC().Truncate(time.Minute)) || m.At.After(now) {
			continue
		}
		s.Wh += m.Joules / 3600
		s.BusyWh += m.BusyJ / 3600
		seconds += m.Seconds
		if m.PeakW > s.PeakW {
			s.PeakW = m.PeakW
		}
		if m.Measured != "" {
			s.Source = m.Measured
		}
	}
	if seconds == 0 {
		return nil
	}
	s.AvgW = round(s.Wh*3600/seconds, 1)
	per := map[string]int{}
	for _, sm := range samples {
		if sm.At.Before(since) || sm.At.After(now) || sm.Status >= 400 {
			continue
		}
		s.InTokens += sm.InTokens
		s.OutTokens += sm.OutTokens
		if sm.OutTokens > 0 {
			per[sm.Model] += sm.OutTokens
		}
	}
	s.IdleWh = round(s.Wh-s.BusyWh, 3)
	s.TokensPerKWh = perKWh(s.OutTokens, s.BusyWh)
	names := make([]string, 0, len(per))
	for m := range per {
		names = append(names, m)
	}
	sort.Strings(names)
	for _, m := range names {
		share := s.BusyWh * float64(per[m]) / float64(s.OutTokens)
		s.Models = append(s.Models, ModelEnergy{Model: m, OutTokens: per[m], BusyWh: round(share, 3), TokensPerKWh: perKWh(per[m], share)})
	}
	s.Wh, s.BusyWh = round(s.Wh, 3), round(s.BusyWh, 3)
	return &s
}

func perKWh(tokens int, wh float64) *int {
	if tokens <= 0 || wh < 0.001 {
		return nil
	}
	v := int(math.Round(float64(tokens) / (wh / 1000)))
	return &v
}

func round(v float64, places int) float64 {
	p := math.Pow(10, float64(places))
	return math.Round(v*p) / p
}

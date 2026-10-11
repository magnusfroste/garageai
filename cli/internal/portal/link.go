package portal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// The link flow: Bridge asks the portal for a code, the operator approves this machine in the
// browser, and Bridge picks up the garage's credentials. No token is ever copied or shown.
// It is the device-authorization pattern (gh auth login).

// DefaultAPI is the portal's functions URL; GARAGEAI_PORTAL_API overrides it (tests, staging).
const DefaultAPI = "https://sjopaiquzqotymqmrcke.supabase.co/functions/v1"

func API() string {
	if v := os.Getenv("GARAGEAI_PORTAL_API"); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return DefaultAPI
}

// LinkRequest describes this machine to the operator who approves it.
type LinkRequest struct {
	MachineID     string   `json:"machine_id,omitempty"`
	Hostname      string   `json:"hostname,omitempty"`
	OS            string   `json:"os,omitempty"`
	Arch          string   `json:"arch,omitempty"`
	BridgeVersion string   `json:"bridge_version,omitempty"`
	GPU           string   `json:"gpu,omitempty"`
	MemoryGB      int      `json:"memory_gb,omitempty"`
	Runtime       string   `json:"runtime,omitempty"`
	Models        []string `json:"models,omitempty"`
	SuggestedName string   `json:"suggested_name,omitempty"`
}

type LinkStart struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURL         string `json:"verification_url"`
	VerificationURLComplete string `json:"verification_url_complete"`
	Interval                int    `json:"interval"`
	ExpiresIn               int    `json:"expires_in"`
}

type LinkResult struct {
	Status string `json:"status"` // pending, slow_down, approved, denied, expired, picked_up, unknown
	Garage struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"garage"`
	RegisterURL   string `json:"register_url"`
	RegisterToken string `json:"register_token"`
	SetupKey      string `json:"setup_key"`
	ManagementURL string `json:"management_url"`
}

func linkPost(ctx context.Context, name string, body, out any) (int, error) {
	b, _ := json.Marshal(body)
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "POST", API()+"/"+name, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err := json.Unmarshal(data, out); err != nil && resp.StatusCode/100 == 2 {
		return resp.StatusCode, fmt.Errorf("%s: unexpected answer", name)
	}
	if resp.StatusCode/100 != 2 && resp.StatusCode != 404 {
		msg := strings.TrimSpace(string(data))
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return resp.StatusCode, fmt.Errorf("%s: HTTP %d: %s", name, resp.StatusCode, msg)
	}
	return resp.StatusCode, nil
}

func StartLink(ctx context.Context, r LinkRequest) (*LinkStart, error) {
	var out LinkStart
	if _, err := linkPost(ctx, "bridge-link-start", r, &out); err != nil {
		return nil, err
	}
	if out.DeviceCode == "" || out.UserCode == "" {
		return nil, fmt.Errorf("bridge-link-start: no code in the answer")
	}
	if out.Interval < 1 {
		out.Interval = 3
	}
	if out.ExpiresIn < 1 {
		out.ExpiresIn = 600
	}
	return &out, nil
}

// WaitForLink polls until the operator approves or denies, the code expires, or ctx ends.
// tick is called before every poll (for a spinner or a countdown).
func WaitForLink(ctx context.Context, s *LinkStart, tick func(left time.Duration)) (*LinkResult, error) {
	interval := time.Duration(s.Interval) * time.Second
	deadline := time.Now().Add(time.Duration(s.ExpiresIn) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
		if tick != nil {
			tick(time.Until(deadline))
		}
		var r LinkResult
		_, err := linkPost(ctx, "bridge-link-poll", map[string]string{"device_code": s.DeviceCode}, &r)
		switch {
		case err != nil:
			// A network hiccup is not the end: keep polling until the code expires.
		case r.Status == "approved":
			if r.RegisterToken == "" || r.RegisterURL == "" {
				return nil, fmt.Errorf("approved, but the portal sent no credentials")
			}
			return &r, nil
		case r.Status == "slow_down":
			interval += 2 * time.Second
		case r.Status == "pending":
		default:
			return &r, nil // denied, expired, picked_up, unknown
		}
		if time.Now().After(deadline) {
			return &LinkResult{Status: "expired"}, nil
		}
	}
}

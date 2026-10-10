// Package portal talks to the GarageAI portal the way garageai-connect.sh does: register-node,
// node-heartbeat and onboarding-report, with the garage's register token.
package portal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	RegisterURL string // .../functions/v1/register-node
	Token       string
	HTTP        *http.Client
}

func New(registerURL, token string) *Client {
	return &Client{RegisterURL: registerURL, Token: token, HTTP: &http.Client{Timeout: 190 * time.Second}}
}

func (c *Client) sibling(name string) string {
	return strings.TrimSuffix(c.RegisterURL, "/register-node") + "/" + name
}

// Registration is the body of register-node and node-heartbeat (mesh_ip only at registration).
type Registration struct {
	Name          string         `json:"name"`
	MeshIP        string         `json:"mesh_ip,omitempty"`
	Port          int            `json:"port"`
	Runtime       string         `json:"runtime"`
	Models        []string       `json:"models"`
	ContextLength int            `json:"context_length,omitempty"`
	Contexts      map[string]int `json:"contexts,omitempty"`
	RuntimeAPIKey string         `json:"runtime_api_key,omitempty"`
}

type Acceptance struct {
	Model           string   `json:"model"`
	Passed          bool     `json:"passed"`
	TokensPerSecond *float64 `json:"tokens_per_second"`
	TTFTms          *float64 `json:"ttft_ms"`
	Error           string   `json:"error"`
}

type RegisterResult struct {
	Acceptance []Acceptance `json:"acceptance"`
}

type HeartbeatResult struct {
	OK              bool         `json:"ok"`
	Changed         bool         `json:"changed"`
	Added           []string     `json:"added"`
	Removed         []string     `json:"removed"`
	ContextsChanged []string     `json:"contexts_changed"`
	Acceptance      []Acceptance `json:"acceptance"`
	Error           string       `json:"error"`
}

func (c *Client) post(ctx context.Context, url string, body any, out any, timeout time.Duration) (int, error) {
	b, _ := json.Marshal(body)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if out != nil && len(data) > 0 {
		_ = json.Unmarshal(data, out)
	}
	if resp.StatusCode/100 != 2 {
		msg := strings.TrimSpace(string(data))
		var e struct{ Error string }
		if json.Unmarshal(data, &e) == nil && e.Error != "" {
			msg = e.Error
		}
		if len(msg) > 300 {
			msg = msg[:300]
		}
		return resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}
	return resp.StatusCode, nil
}

// Register registers the node; the portal runs the acceptance test, which can take a minute.
func (c *Client) Register(ctx context.Context, r Registration) (*RegisterResult, error) {
	var out RegisterResult
	_, err := c.post(ctx, c.RegisterURL, r, &out, 190*time.Second)
	return &out, err
}

// Heartbeat reports the current inventory; the portal rebuilds routing when something changed.
func (c *Client) Heartbeat(ctx context.Context, r Registration) (*HeartbeatResult, error) {
	var out HeartbeatResult
	code, err := c.post(ctx, c.sibling("node-heartbeat"), r, &out, 190*time.Second)
	if code == 401 || code == 403 {
		return &out, fmt.Errorf("%w: %v", ErrRejected, err)
	}
	return &out, err
}

// ErrRejected: the token was replaced or revoked, or the garage was disabled.
var ErrRejected = errors.New("GarageAI rejects this garage's token")

// Report is one onboarding step, best effort: it never fails the caller and waits at most 5 s.
type Report struct {
	Step          string          `json:"step"`
	Status        string          `json:"status"` // started, warning, failed, stopped, done
	Message       string          `json:"message"`
	ScriptVersion string          `json:"script_version"`
	NodeName      string          `json:"node_name"`
	Runtime       string          `json:"runtime"`
	Port          string          `json:"port"`
	OS            string          `json:"os"`
	Arch          string          `json:"arch"`
	GPU           string          `json:"gpu"`
	MemoryGB      string          `json:"memory_gb"`
	Profile       json.RawMessage `json:"profile,omitempty"`
}

func (c *Client) Report(r Report) {
	if c == nil || c.RegisterURL == "" || c.Token == "" || os.Getenv("GARAGEAI_REPORT") == "0" {
		return
	}
	if len(r.Message) > 500 {
		r.Message = r.Message[:500]
	}
	_, _ = c.post(context.Background(), c.sibling("onboarding-report"), r, nil, 5*time.Second)
}

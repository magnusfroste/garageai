package tunnel

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// The forward from the tunnel to the runtime is an HTTP reverse proxy, not a raw TCP copy, for
// two reasons:
//
//   - Only inference reaches the runtime. Ollama's port also carries its management API (pull,
//     delete, create); a raw forward handed all of it to whoever is on the other end of the tunnel.
//     docs/garage-security.md promised "GarageAI does not call them"; this makes it a property of
//     the garage instead of a promise (2026-10-13 security review).
//   - The Host header is the runtime's own. The gateway dials 10.66.0.x:<port> and sends that as
//     Host; a runtime that only accepts loopback authorities (Xyntetik Runner) refused it. The
//     proxy sets Host to 127.0.0.1:<port>, which is where the request is going anyway.
//
// Streaming is passed through as it comes (no buffering), and the wait for the first byte is long:
// a 100k-token prompt takes minutes to prefill on a home GPU.

// allowedPrefixes are the paths the gateway may reach: the OpenAI-style inference API (and the
// Anthropic and Responses variants some runtimes serve), the read-only Ollama endpoints the
// gateway's health check and Ops use, and /metrics and /health for Ops.
var allowedPrefixes = []string{
	"/v1/chat/completions", "/v1/completions", "/v1/embeddings", "/v1/models", "/v1/messages",
	"/v1/responses", "/v1/rerank", "/v1/capabilities",
	"/api/tags", "/api/ps", "/api/show", "/api/version", "/api/chat", "/api/generate", "/api/embed", "/api/embeddings",
	"/metrics", "/health", "/healthz", "/v1/health",
}

// Allowed reports whether a request path may pass through the tunnel to the runtime.
func Allowed(path string) bool {
	for _, p := range allowedPrefixes {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

// ForwardHandler proxies requests to the runtime at local (host:port on this machine), with the
// path allowlist and the Host rewrite. logf may be nil.
func ForwardHandler(local string, logf func(format string, args ...any)) http.Handler {
	target := &url.URL{Scheme: "http", Host: local}
	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(r *http.Request) {
		director(r)
		r.Host = local // the runtime sees its own address, not the tunnel's
		r.Header.Del("X-Forwarded-For")
	}
	proxy.Transport = &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
		ResponseHeaderTimeout: 15 * time.Minute, // prefill of a long prompt on a home GPU
		DisableCompression:    true,
		MaxIdleConns:          16,
		IdleConnTimeout:       90 * time.Second,
	}
	proxy.FlushInterval = -1 // tokens as they come
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if logf != nil {
			logf("forward to %s: %v", local, err)
		}
		writeJSON(w, http.StatusBadGateway, "the runtime did not answer: "+err.Error())
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Allowed(r.URL.Path) {
			if logf != nil {
				logf("refused %s %s through the tunnel (not an inference path)", r.Method, r.URL.Path)
			}
			writeJSON(w, http.StatusForbidden, "GarageAI Bridge forwards inference paths only")
			return
		}
		proxy.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": msg, "type": "bridge_forward", "code": code}})
}

// serveForward runs the HTTP forward on a tunnel listener until the listener closes.
func (t *Tunnel) serveForward(ln net.Listener, local string) {
	srv := &http.Server{
		Handler:           ForwardHandler(local, t.cfg.Logf),
		ReadHeaderTimeout: 30 * time.Second,
		IdleTimeout:       5 * time.Minute,
	}
	t.servers = append(t.servers, srv)
	_ = srv.Serve(ln)
}

func closeServers(servers []*http.Server) {
	for _, s := range servers {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = s.Shutdown(ctx)
		cancel()
	}
}

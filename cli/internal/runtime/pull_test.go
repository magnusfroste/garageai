package runtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestOllamaPullStreamsProgress(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pull" || r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`{"status":"pulling manifest"}` + "\n" +
			`{"status":"pulling abc","digest":"sha256:abc","total":1000,"completed":400}` + "\n" +
			`{"status":"pulling abc","digest":"sha256:abc","total":1000,"completed":1000}` + "\n" +
			`{"status":"verifying sha256 digest"}` + "\n" + `{"status":"success"}` + "\n"))
	}))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)
	port, _ := strconv.Atoi(u.Port())
	c := New(u.Hostname(), port, "")
	var seen []int64
	if err := c.OllamaPull(context.Background(), "qwen3:4b", func(done, total int64) { seen = append(seen, done) }); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 || seen[0] != 400 || seen[1] != 1000 {
		t.Fatalf("progress %v", seen)
	}
}

func TestOllamaPullReportsOllamaError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"pulling manifest"}` + "\n" + `{"error":"pull model manifest: file does not exist"}` + "\n"))
	}))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)
	port, _ := strconv.Atoi(u.Port())
	err := New(u.Hostname(), port, "").OllamaPull(context.Background(), "nope:1", nil)
	if err == nil || err.Error() != "pull nope:1: pull model manifest: file does not exist" {
		t.Fatalf("err %v", err)
	}
}

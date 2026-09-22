package jev_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name) //nolint:gosec // test fixture path
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

type cannedResponse struct {
	status int
	body   []byte
}

type capture struct {
	mu     sync.Mutex
	calls  int
	method string
	path   string
	auth   string
	body   []byte
}

func (c *capture) record(r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	c.method = r.Method
	c.path = r.URL.Path
	c.auth = r.Header.Get("Authorization")
	c.body = body
}

func (c *capture) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

// newServer returns a server that replays the canned responses in order,
// repeating the last one once the queue is exhausted.
func newServer(t *testing.T, responses ...cannedResponse) (*httptest.Server, *capture) {
	t.Helper()
	rec := &capture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.record(r)
		i := rec.count() - 1
		if i >= len(responses) {
			i = len(responses) - 1
		}
		resp := responses[i]
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.status)
		_, _ = w.Write(resp.body)
	}))
	t.Cleanup(srv.Close)
	return srv, rec
}

package jev_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/shanehull/go-jev"
)

func TestMapEvaluatesAllInOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			State string `json:"state"`
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &req)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"model":%q,"answers":{}}`, req.State)
	}))
	t.Cleanup(srv.Close)

	client := newClient(t, srv.URL)
	states := []jev.Input{jev.State("a"), jev.State("b"), jev.State("c")}
	results := client.Map(context.Background(), states, testQuestions(), jev.WithConcurrency(3))

	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	for i, want := range []string{"a", "b", "c"} {
		if results[i].Err != nil {
			t.Fatalf("result %d: %v", i, results[i].Err)
		}
		if results[i].Response.Model != want {
			t.Errorf("result %d model = %q, want %q", i, results[i].Response.Model, want)
		}
	}
}

func TestMapRetriesRetryableFailure(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"boom"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"ok","answers":{}}`))
	}))
	t.Cleanup(srv.Close)

	client := newClient(t, srv.URL)
	results := client.Map(context.Background(), []jev.Input{jev.State("a")}, testQuestions(),
		jev.WithConcurrency(1), jev.WithMaxRetries(2))

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("result error: %v", results[0].Err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}
}

func TestMapDoesNotRetryClientError(t *testing.T) {
	srv, rec := newServer(t, cannedResponse{status: http.StatusUnauthorized, body: []byte(`{"error":"nope"}`)})
	client := newClient(t, srv.URL)

	results := client.Map(context.Background(), []jev.Input{jev.State("a")}, testQuestions(),
		jev.WithConcurrency(1), jev.WithMaxRetries(3))

	if results[0].Err == nil {
		t.Fatal("expected an error")
	}
	if got := rec.count(); got != 1 {
		t.Errorf("calls = %d, want 1 (no retry on 401)", got)
	}
}

package jev_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/shanehull/go-jev"
)

func testQuestions() jev.Questions {
	return jev.Questions{
		"department": jev.Choice("Which team should handle this", map[string]string{
			"billing":   "Payment or subscription issues",
			"technical": "Bugs or integration problems",
		}),
		"frustration": jev.Score("How frustrated the customer appears", []string{
			"Calm, just stating facts",
			"Frustrated but civil",
			"Very angry, strong language",
		}),
		"is_urgent": jev.Noul("The message conveys urgency or time sensitivity"),
	}
}

func newClient(t *testing.T, baseURL string) *jev.Client {
	t.Helper()
	client, err := jev.New(jev.WithAPIKey("test-key"), jev.WithBaseURL(baseURL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client
}

func TestSystemOneRequest(t *testing.T) {
	srv, rec := newServer(t, cannedResponse{status: http.StatusOK, body: readFixture(t, "systemone_response.json")})
	client := newClient(t, srv.URL)

	if _, err := client.SystemOne(context.Background(), jev.State("hello"), testQuestions()); err != nil {
		t.Fatalf("SystemOne: %v", err)
	}

	if rec.method != http.MethodPost {
		t.Errorf("method = %s, want POST", rec.method)
	}
	if rec.path != "/v1/systemone" {
		t.Errorf("path = %s, want /v1/systemone", rec.path)
	}
	if rec.auth != "Bearer test-key" {
		t.Errorf("auth = %q", rec.auth)
	}

	var req struct {
		State     string `json:"state"`
		Model     string `json:"model"`
		Questions map[string]struct {
			Type         string          `json:"type"`
			Instructions string          `json:"instructions"`
			Criteria     json.RawMessage `json:"criteria"`
		} `json:"questions"`
	}
	if err := json.Unmarshal(rec.body, &req); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if req.State != "hello" {
		t.Errorf("state = %q, want hello", req.State)
	}
	if req.Model != "jev-latest" {
		t.Errorf("model = %q, want jev-latest", req.Model)
	}
	if req.Questions["is_urgent"].Type != "noul" || len(req.Questions["is_urgent"].Criteria) != 0 {
		t.Errorf("is_urgent = %+v", req.Questions["is_urgent"])
	}
	if req.Questions["department"].Type != "choice" {
		t.Errorf("department type = %q", req.Questions["department"].Type)
	}
	var choiceCriteria map[string]string
	if err := json.Unmarshal(req.Questions["department"].Criteria, &choiceCriteria); err != nil {
		t.Fatalf("choice criteria: %v", err)
	}
	if choiceCriteria["technical"] != "Bugs or integration problems" {
		t.Errorf("choice criteria = %v", choiceCriteria)
	}
	var scoreCriteria []string
	if err := json.Unmarshal(req.Questions["frustration"].Criteria, &scoreCriteria); err != nil {
		t.Fatalf("score criteria: %v", err)
	}
	if len(scoreCriteria) != 3 {
		t.Errorf("score criteria = %v", scoreCriteria)
	}
}

func TestSystemOneResponse(t *testing.T) {
	srv, _ := newServer(t, cannedResponse{status: http.StatusOK, body: readFixture(t, "systemone_response.json")})
	client := newClient(t, srv.URL)

	res, err := client.SystemOne(context.Background(), jev.State("hello"), testQuestions())
	if err != nil {
		t.Fatalf("SystemOne: %v", err)
	}
	if res.Model != "jev-1.13.0" {
		t.Errorf("model = %q", res.Model)
	}
	if res.Usage.InputTokens != 392 || res.Usage.OutputTokens != 65 {
		t.Errorf("usage = %+v", res.Usage)
	}

	noul, ok := res.Noul("is_urgent")
	if !ok || noul.Noul != 1.0 {
		t.Errorf("noul = %+v ok=%v", noul, ok)
	}
	choice, ok := res.Choice("department")
	if !ok || choice.Choice != "technical" || choice.Probabilities["technical"] != 0.85 {
		t.Errorf("choice = %+v ok=%v", choice, ok)
	}
	score, ok := res.Score("frustration")
	if !ok || score.Score != 1.0 || score.Legend["1"] != "Frustrated but civil" {
		t.Errorf("score = %+v ok=%v", score, ok)
	}
	if len(res.Answers) != 3 {
		t.Errorf("answers = %d, want 3", len(res.Answers))
	}
}

func TestSystemOneAnswerKinds(t *testing.T) {
	srv, _ := newServer(t, cannedResponse{status: http.StatusOK, body: readFixture(t, "systemone_response.json")})
	client := newClient(t, srv.URL)

	res, err := client.SystemOne(context.Background(), jev.State("hello"), testQuestions())
	if err != nil {
		t.Fatalf("SystemOne: %v", err)
	}
	kinds := map[string]string{"department": "choice", "frustration": "score", "is_urgent": "noul"}
	for name, want := range kinds {
		if got := res.Answers[name].Kind(); got != want {
			t.Errorf("answer %s kind = %s, want %s", name, got, want)
		}
	}
}

func TestSystemOneAPIError(t *testing.T) {
	srv, _ := newServer(t, cannedResponse{status: http.StatusUnauthorized, body: []byte(`{"error":"invalid API key"}`)})
	client := newClient(t, srv.URL)

	_, err := client.SystemOne(context.Background(), jev.State("hello"), testQuestions())
	if err == nil {
		t.Fatal("expected an error")
	}
	var apiErr *jev.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *jev.APIError: %v", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d", apiErr.StatusCode)
	}
	if apiErr.Message != "invalid API key" {
		t.Errorf("message = %q", apiErr.Message)
	}
}

func TestSystemOneNoQuestions(t *testing.T) {
	srv, _ := newServer(t, cannedResponse{status: http.StatusOK, body: []byte(`{}`)})
	client := newClient(t, srv.URL)

	if _, err := client.SystemOne(context.Background(), jev.State("hello"), nil); !errors.Is(err, jev.ErrNoQuestions) {
		t.Fatalf("error = %v, want ErrNoQuestions", err)
	}
}

func TestSystemOneRequestModelOverride(t *testing.T) {
	srv, rec := newServer(t, cannedResponse{status: http.StatusOK, body: []byte(`{"model":"x","answers":{}}`)})
	client := newClient(t, srv.URL)

	if _, err := client.SystemOne(context.Background(), jev.State("hello"), testQuestions(), jev.WithRequestModel("custom")); err != nil {
		t.Fatalf("SystemOne: %v", err)
	}
	var req struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(rec.body, &req); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if req.Model != "custom" {
		t.Errorf("model = %q, want custom", req.Model)
	}
}

func TestNewRequiresKey(t *testing.T) {
	t.Setenv("TYPESAFE_API_KEY", "")
	if _, err := jev.New(); err == nil {
		t.Fatal("expected an error without an API key")
	}
}

func TestSystemOneRetriesTransient(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(529)
			_, _ = w.Write([]byte(`{"error":"overloaded"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(readFixture(t, "systemone_response.json"))
	}))
	t.Cleanup(srv.Close)

	client := newClient(t, srv.URL)
	if _, err := client.SystemOne(context.Background(), jev.State("hello"), testQuestions()); err != nil {
		t.Fatalf("SystemOne: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}
}

func TestSystemOneDoesNotRetryClientError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	t.Cleanup(srv.Close)

	client := newClient(t, srv.URL)
	if _, err := client.SystemOne(context.Background(), jev.State("hello"), testQuestions()); err == nil {
		t.Fatal("expected an error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

func TestSystemOneMaxAttemptsExhausted(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"unavailable"}`))
	}))
	t.Cleanup(srv.Close)

	client, err := jev.New(jev.WithAPIKey("test-key"), jev.WithBaseURL(srv.URL), jev.WithMaxAttempts(2))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := client.SystemOne(context.Background(), jev.State("hello"), testQuestions()); err == nil {
		t.Fatal("expected an error")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}
}

func TestEvaluate(t *testing.T) {
	srv, _ := newServer(t, cannedResponse{status: http.StatusOK, body: readFixture(t, "systemone_response.json")})
	client := newClient(t, srv.URL)

	res, err := client.Evaluate(context.Background(), jev.Request{State: "hello", Questions: testQuestions()})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if _, ok := res.Noul("is_urgent"); !ok {
		t.Fatalf("missing answer: %+v", res.Answers)
	}
}

func TestAPIErrorRetryable(t *testing.T) {
	tests := []struct {
		status int
		want   bool
	}{
		{http.StatusTooManyRequests, true},
		{529, true},
		{http.StatusInternalServerError, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusNotFound, false},
	}
	for _, tt := range tests {
		err := jev.APIError{StatusCode: tt.status}
		if got := err.Retryable(); got != tt.want {
			t.Errorf("status %d retryable = %v, want %v", tt.status, got, tt.want)
		}
	}
}

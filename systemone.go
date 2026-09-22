package jev

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shanehull/go-jev/internal"
)

// Request is a full System One request. Send it directly with Evaluate, or use
// SystemOne for the common case.
type Request struct {
	State     any
	Model     string
	Questions Questions
}

type systemOneRequest struct {
	State     any       `json:"state"`
	Model     string    `json:"model"`
	Questions Questions `json:"questions"`
}

// Evaluate sends a request and returns the decoded response. Transient
// failures (HTTP 429, 529, and 5xx) are retried with exponential backoff.
func (c *Client) Evaluate(ctx context.Context, req Request) (*Response, error) {
	return c.evaluate(ctx, req, c.maxAttempts)
}

// SystemOne evaluates a set of questions against a state in a single request.
// This is the fan-out primitive: every question is answered in one round trip.
func (c *Client) SystemOne(ctx context.Context, state Input, questions Questions, opts ...RequestOption) (*Response, error) {
	params := &requestParams{model: c.model}
	for _, opt := range opts {
		opt(params)
	}
	return c.evaluate(ctx, Request{State: state.value, Model: params.model, Questions: questions}, c.maxAttempts)
}

func (c *Client) evaluate(ctx context.Context, req Request, attempts int) (*Response, error) {
	if len(req.Questions) == 0 {
		return nil, ErrNoQuestions
	}

	model := req.Model
	if model == "" {
		model = c.model
	}

	payload, err := json.Marshal(systemOneRequest{
		State:     req.State,
		Model:     model,
		Questions: req.Questions,
	})
	if err != nil {
		return nil, fmt.Errorf("jev: encode request: %w", err)
	}

	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if err := sleepBackoff(ctx, attempt); err != nil {
				return nil, err
			}
		}

		body, err := internal.Post(ctx, c.httpClient, c.baseURL+"/v1/systemone", c.apiKey, payload)
		if err == nil {
			var resp Response
			if err := json.Unmarshal(body, &resp); err != nil {
				return nil, fmt.Errorf("jev: decode response: %w", err)
			}
			return &resp, nil
		}

		lastErr = err
		if !retryable(err) {
			return nil, err
		}
	}
	return nil, lastErr
}

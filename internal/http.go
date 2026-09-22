// Package internal provides the HTTP transport shared by the client.
package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError represents a non-success response from the TypeSafe API.
type APIError struct {
	StatusCode int
	URL        string
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("jev: API error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("jev: HTTP %d for %s", e.StatusCode, e.URL)
}

// Retryable reports whether the status code is one worth retrying: rate
// limiting, the Cloudflare overload code, or a server error.
func (e *APIError) Retryable() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode == 529 || e.StatusCode >= 500
}

// Post sends a JSON request with bearer auth and returns the response body.
func Post(ctx context.Context, client *http.Client, url, apiKey string, payload []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("jev: request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jev: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("jev: reading response failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			URL:        url,
			Message:    errorMessage(body),
			Body:       string(body),
		}
	}
	return body, nil
}

func errorMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var generic map[string]json.RawMessage
	if err := json.Unmarshal(body, &generic); err != nil {
		return ""
	}
	for _, key := range []string{"error", "message", "detail"} {
		raw, ok := generic[key]
		if !ok {
			continue
		}
		var s string
		if err := json.Unmarshal(raw, &s); err == nil && s != "" {
			return s
		}
	}
	return ""
}

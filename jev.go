// Package jev provides a client for the TypeSafe System One API.
//
// It sends a state and a set of typed questions, and returns calibrated
// answers: a choice from a list, a score on an ordered rubric, or the
// probability that a statement is true.
//
// Usage:
//
//	client, err := jev.New()
//	if err != nil {
//		log.Fatal(err)
//	}
//	res, err := client.SystemOne(ctx, jev.State(text),
//		jev.Questions{
//			"material": jev.Noul("The text describes a material event"),
//		})
package jev

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	defaultBaseURL     = "https://api.typesafe.ai"
	defaultModel       = "jev-latest"
	defaultTimeout     = 30 * time.Second
	defaultMaxAttempts = 3
)

// Client is a TypeSafe System One client. Must be constructed via New.
// Safe for concurrent use by multiple goroutines.
//
// API key resolution order:
//  1. WithAPIKey("...")
//  2. $TYPESAFE_API_KEY environment variable
type Client struct {
	apiKey      string
	model       string
	httpClient  *http.Client
	baseURL     string
	maxAttempts int
}

// New creates a System One client.
func New(opts ...ClientOption) (*Client, error) {
	c := &Client{
		model:       defaultModel,
		httpClient:  &http.Client{Timeout: defaultTimeout},
		baseURL:     defaultBaseURL,
		maxAttempts: defaultMaxAttempts,
	}

	if key := os.Getenv("TYPESAFE_API_KEY"); key != "" {
		c.apiKey = key
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.apiKey == "" {
		return nil, fmt.Errorf("jev: API key required; set $TYPESAFE_API_KEY or use WithAPIKey()")
	}
	return c, nil
}

// ClientOption configures a Client.
type ClientOption func(*Client) error

// WithAPIKey sets the API key.
func WithAPIKey(key string) ClientOption {
	return func(c *Client) error {
		c.apiKey = key
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) error {
		if hc == nil {
			return fmt.Errorf("jev: http client cannot be nil")
		}
		c.httpClient = hc
		return nil
	}
}

// WithBaseURL sets a custom base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		if baseURL == "" {
			return fmt.Errorf("jev: base URL cannot be empty")
		}
		c.baseURL = baseURL
		return nil
	}
}

// WithModel sets the default model. It defaults to jev-latest.
func WithModel(model string) ClientOption {
	return func(c *Client) error {
		if model == "" {
			return fmt.Errorf("jev: model cannot be empty")
		}
		c.model = model
		return nil
	}
}

// WithMaxAttempts sets the number of attempts made for a single request,
// including the first. It defaults to 3. Transient failures (HTTP 429, 529, and
// 5xx) are retried with exponential backoff.
func WithMaxAttempts(n int) ClientOption {
	return func(c *Client) error {
		if n < 1 {
			return fmt.Errorf("jev: max attempts must be at least 1")
		}
		c.maxAttempts = n
		return nil
	}
}

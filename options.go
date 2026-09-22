package jev

import "time"

type requestParams struct {
	model string
}

// RequestOption configures a single SystemOne call.
type RequestOption func(*requestParams)

// WithRequestModel overrides the model for one call.
func WithRequestModel(model string) RequestOption {
	return func(p *requestParams) { p.model = model }
}

type mapParams struct {
	concurrency int
	maxRetries  int
	minInterval time.Duration
	model       string
}

// MapOption configures a Map call.
type MapOption func(*mapParams)

// WithConcurrency sets the number of in-flight requests. It defaults to 4.
func WithConcurrency(n int) MapOption {
	return func(p *mapParams) { p.concurrency = n }
}

// WithMaxRetries sets the number of retries after the first attempt for
// retryable failures. It defaults to 2.
func WithMaxRetries(n int) MapOption {
	return func(p *mapParams) { p.maxRetries = n }
}

// WithMinInterval sets the minimum spacing between requests, a simple rate limit.
func WithMinInterval(d time.Duration) MapOption {
	return func(p *mapParams) { p.minInterval = d }
}

// WithMapModel overrides the model for a Map call.
func WithMapModel(model string) MapOption {
	return func(p *mapParams) { p.model = model }
}

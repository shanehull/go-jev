package jev

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/shanehull/go-jev/internal"
)

// MapResult is the outcome of evaluating one state in a Map call.
type MapResult struct {
	Response *Response
	Err      error
}

// Map evaluates the same questions against many states with bounded
// concurrency, retries with exponential backoff on retryable failures, and
// optional rate limiting. Results are returned in input order.
func (c *Client) Map(ctx context.Context, states []Input, questions Questions, opts ...MapOption) []MapResult {
	params := &mapParams{concurrency: 4, maxRetries: 2, model: c.model}
	for _, opt := range opts {
		opt(params)
	}
	if params.concurrency < 1 {
		params.concurrency = 1
	}

	results := make([]MapResult, len(states))
	sem := make(chan struct{}, params.concurrency)
	limiter := &paceLimiter{interval: params.minInterval}

	var wg sync.WaitGroup
	for i, state := range states {
		wg.Add(1)
		go func(i int, state Input) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[i] = MapResult{Err: ctx.Err()}
				return
			}

			if err := limiter.wait(ctx); err != nil {
				results[i] = MapResult{Err: err}
				return
			}
			results[i] = c.mapOne(ctx, state, questions, params)
		}(i, state)
	}
	wg.Wait()
	return results
}

func (c *Client) mapOne(ctx context.Context, state Input, questions Questions, params *mapParams) MapResult {
	var lastErr error
	for attempt := 0; attempt <= params.maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepBackoff(ctx, attempt); err != nil {
				return MapResult{Err: err}
			}
		}

		resp, err := c.evaluate(ctx, Request{State: state.value, Model: params.model, Questions: questions}, 1)
		if err == nil {
			return MapResult{Response: resp}
		}
		lastErr = err
		if !retryable(err) {
			break
		}
	}
	return MapResult{Err: lastErr}
}

func retryable(err error) bool {
	var apiErr *internal.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Retryable()
	}
	return true
}

func sleepBackoff(ctx context.Context, attempt int) error {
	delay := time.Duration(1<<uint(attempt-1)) * 200 * time.Millisecond
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type paceLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	last     time.Time
}

func (p *paceLimiter) wait(ctx context.Context) error {
	if p.interval <= 0 {
		return nil
	}

	p.mu.Lock()
	now := time.Now()
	wait := time.Duration(0)
	if !p.last.IsZero() {
		wait = time.Until(p.last.Add(p.interval))
	}
	if wait < 0 {
		wait = 0
	}
	p.last = now.Add(wait)
	p.mu.Unlock()

	if wait == 0 {
		return nil
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

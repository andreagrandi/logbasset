package client

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultMaxRetries = 3
	defaultBaseDelay  = 200 * time.Millisecond
	defaultMaxDelay   = 10 * time.Second
)

// RetryPolicy controls how transient failures are retried.
type RetryPolicy struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

func defaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries: defaultMaxRetries,
		BaseDelay:  defaultBaseDelay,
		MaxDelay:   defaultMaxDelay,
	}
}

// isRetryableStatus returns true for HTTP status codes that indicate a
// transient failure worth retrying (5xx server errors and 429 throttling).
func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || (code >= 500 && code <= 599)
}

// parseRetryAfter parses the Retry-After header value, which may be either a
// delta in seconds or an HTTP-date. Returns 0 when absent or unparseable.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(value); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0
		}
		return d
	}
	return 0
}

// backoffDelay computes the delay before the next retry attempt using
// exponential backoff with full jitter, bounded by MaxDelay.
func (p RetryPolicy) backoffDelay(attempt int) time.Duration {
	if p.BaseDelay <= 0 {
		return 0
	}
	if attempt < 0 {
		attempt = 0
	}
	if attempt > 30 {
		attempt = 30
	}
	delay := p.BaseDelay << attempt
	if delay <= 0 || delay > p.MaxDelay {
		delay = p.MaxDelay
	}
	return time.Duration(rand.Int63n(int64(delay) + 1))
}

// sleepWithContext sleeps for d or returns early when the context is done.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fastRetryPolicy(maxRetries int) RetryPolicy {
	return RetryPolicy{
		MaxRetries: maxRetries,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
	}
}

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusForbidden, false},
		{http.StatusNotFound, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, isRetryableStatus(tt.code), "status %d", tt.code)
	}
}

func TestParseRetryAfter(t *testing.T) {
	assert.Equal(t, time.Duration(0), parseRetryAfter(""))
	assert.Equal(t, 5*time.Second, parseRetryAfter("5"))
	assert.Equal(t, time.Duration(0), parseRetryAfter("not-a-number"))

	future := time.Now().Add(3 * time.Second).UTC().Format(http.TimeFormat)
	d := parseRetryAfter(future)
	assert.Greater(t, d, time.Duration(0))
	assert.LessOrEqual(t, d, 4*time.Second)

	past := time.Now().Add(-10 * time.Second).UTC().Format(http.TimeFormat)
	assert.Equal(t, time.Duration(0), parseRetryAfter(past))
}

func TestBackoffDelay_BoundedByMaxDelay(t *testing.T) {
	policy := RetryPolicy{BaseDelay: 100 * time.Millisecond, MaxDelay: 500 * time.Millisecond}
	for attempt := 0; attempt < 10; attempt++ {
		d := policy.backoffDelay(attempt)
		assert.LessOrEqual(t, d, policy.MaxDelay)
		assert.GreaterOrEqual(t, d, time.Duration(0))
	}
}

func TestBackoffDelay_ZeroBaseDelay(t *testing.T) {
	policy := RetryPolicy{BaseDelay: 0, MaxDelay: 500 * time.Millisecond}
	assert.Equal(t, time.Duration(0), policy.backoffDelay(0))
	assert.Equal(t, time.Duration(0), policy.backoffDelay(5))
}

func TestSleepWithContext_CompletesNormally(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	err := sleepWithContext(ctx, 5*time.Millisecond)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, time.Since(start), 5*time.Millisecond)
}

func TestSleepWithContext_CancelledEarly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := sleepWithContext(ctx, 10*time.Second)
	assert.Equal(t, context.Canceled, err)
}

func TestSleepWithContext_ZeroDuration(t *testing.T) {
	err := sleepWithContext(context.Background(), 0)
	assert.NoError(t, err)
}

func TestClient_makeRequest_RetriesTransientServerError(t *testing.T) {
	var attempts int32
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			n := atomic.AddInt32(&attempts, 1)
			if n < 3 {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("server boom")),
					Header:     http.Header{},
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status":"success"}`)),
				Header:     http.Header{},
			}, nil
		},
	}

	client := NewWithHTTPClient("token", "https://test.com", false, mockClient)
	client.SetRetryPolicy(fastRetryPolicy(3))

	resp, err := client.makeRequest(context.Background(), "query", map[string]interface{}{})
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestClient_makeRequest_RetriesTransientNetworkError(t *testing.T) {
	var attempts int32
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			n := atomic.AddInt32(&attempts, 1)
			if n < 2 {
				return nil, errors.New("connection reset")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status":"success"}`)),
				Header:     http.Header{},
			}, nil
		},
	}

	client := NewWithHTTPClient("token", "https://test.com", false, mockClient)
	client.SetRetryPolicy(fastRetryPolicy(3))

	resp, err := client.makeRequest(context.Background(), "query", map[string]interface{}{})
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestClient_makeRequest_FailsAfterMaxRetries_ServerError(t *testing.T) {
	var attempts int32
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&attempts, 1)
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       io.NopCloser(strings.NewReader("bad gateway")),
				Header:     http.Header{},
			}, nil
		},
	}

	client := NewWithHTTPClient("token", "https://test.com", false, mockClient)
	client.SetRetryPolicy(fastRetryPolicy(2))

	_, err := client.makeRequest(context.Background(), "query", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "502")
	assert.Contains(t, err.Error(), "after 3 attempts")
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestClient_makeRequest_FailsAfterMaxRetries_NetworkError(t *testing.T) {
	var attempts int32
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&attempts, 1)
			return nil, errors.New("dial tcp: connection refused")
		},
	}

	client := NewWithHTTPClient("token", "https://test.com", false, mockClient)
	client.SetRetryPolicy(fastRetryPolicy(2))

	_, err := client.makeRequest(context.Background(), "query", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute request")
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestClient_makeRequest_DoesNotRetry4xx(t *testing.T) {
	var attempts int32
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&attempts, 1)
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader(`{"status":"error","message":"invalid token"}`)),
				Header:     http.Header{},
			}, nil
		},
	}

	client := NewWithHTTPClient("token", "https://test.com", false, mockClient)
	client.SetRetryPolicy(fastRetryPolicy(5))

	resp, err := client.makeRequest(context.Background(), "query", map[string]interface{}{})
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestClient_makeRequest_DoesNotRetry_BadRequest(t *testing.T) {
	var attempts int32
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&attempts, 1)
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"status":"error","message":"bad query"}`)),
				Header:     http.Header{},
			}, nil
		},
	}

	client := NewWithHTTPClient("token", "https://test.com", false, mockClient)
	client.SetRetryPolicy(fastRetryPolicy(5))

	resp, err := client.makeRequest(context.Background(), "query", map[string]interface{}{})
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestClient_makeRequest_RespectsRetryAfter429(t *testing.T) {
	var attempts int32
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			n := atomic.AddInt32(&attempts, 1)
			if n == 1 {
				header := http.Header{}
				header.Set("Retry-After", "1")
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       io.NopCloser(strings.NewReader("throttled")),
					Header:     header,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status":"success"}`)),
				Header:     http.Header{},
			}, nil
		},
	}

	client := NewWithHTTPClient("token", "https://test.com", false, mockClient)
	client.SetRetryPolicy(RetryPolicy{
		MaxRetries: 2,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   5 * time.Second,
	})

	start := time.Now()
	resp, err := client.makeRequest(context.Background(), "query", map[string]interface{}{})
	elapsed := time.Since(start)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
	assert.GreaterOrEqual(t, elapsed, 900*time.Millisecond, "should have waited per Retry-After")
}

func TestClient_makeRequest_RetryAfterCappedByMaxDelay(t *testing.T) {
	var attempts int32
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			n := atomic.AddInt32(&attempts, 1)
			if n == 1 {
				header := http.Header{}
				header.Set("Retry-After", "3600")
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       io.NopCloser(strings.NewReader("throttled")),
					Header:     header,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status":"success"}`)),
				Header:     http.Header{},
			}, nil
		},
	}

	client := NewWithHTTPClient("token", "https://test.com", false, mockClient)
	client.SetRetryPolicy(RetryPolicy{
		MaxRetries: 1,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   50 * time.Millisecond,
	})

	start := time.Now()
	resp, err := client.makeRequest(context.Background(), "query", map[string]interface{}{})
	elapsed := time.Since(start)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Less(t, elapsed, 1*time.Second, "Retry-After should be capped by MaxDelay")
}

func TestClient_makeRequest_ContextCancelledDuringBackoff(t *testing.T) {
	var attempts int32
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&attempts, 1)
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("boom")),
				Header:     http.Header{},
			}, nil
		},
	}

	client := NewWithHTTPClient("token", "https://test.com", false, mockClient)
	client.SetRetryPolicy(RetryPolicy{
		MaxRetries: 5,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   1 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := client.makeRequest(ctx, "query", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
	assert.Less(t, int32(atomic.LoadInt32(&attempts)), int32(6), "should stop retrying once context is cancelled")
}

func TestClient_makeRequest_NoRetriesWhenDisabled(t *testing.T) {
	var attempts int32
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&attempts, 1)
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("boom")),
				Header:     http.Header{},
			}, nil
		},
	}

	client := NewWithHTTPClient("token", "https://test.com", false, mockClient)
	client.SetRetryPolicy(RetryPolicy{MaxRetries: 0})

	_, err := client.makeRequest(context.Background(), "query", map[string]interface{}{})
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestClient_makeRequest_AuthErrorNotRetried(t *testing.T) {
	t.Setenv("scalyr_readlog_token", "")
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			t.Fatal("HTTP client should not be called when token is missing")
			return nil, nil
		},
	}
	client := NewWithHTTPClient("", "https://test.com", false, mockClient)
	client.SetRetryPolicy(fastRetryPolicy(5))

	_, err := client.makeRequest(context.Background(), "query", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API token is required")
}

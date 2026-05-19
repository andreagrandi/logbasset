package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/andreagrandi/logbasset/internal/logging"
)

const (
	DefaultServer = "https://www.scalyr.com"
	APIVersion    = "v1"
	redactedValue = "***REDACTED***"
)

var sensitiveParamKeys = []string{"token", "Token", "TOKEN"}

// redactSensitiveParams returns a shallow copy of params with sensitive values
// (such as the Scalyr API token) replaced by a placeholder. The original map is
// left untouched so the real request payload is unaffected.
func redactSensitiveParams(params map[string]interface{}) map[string]interface{} {
	if params == nil {
		return nil
	}
	redacted := make(map[string]interface{}, len(params))
	for k, v := range params {
		redacted[k] = v
	}
	for _, key := range sensitiveParamKeys {
		if _, ok := redacted[key]; ok {
			redacted[key] = redactedValue
		}
	}
	return redacted
}

type Client struct {
	server      string
	token       string
	httpClient  HTTPClient
	verbose     bool
	retryPolicy RetryPolicy
}

func New(token, server string, verbose bool) *Client {
	return NewWithHTTPClient(token, server, verbose, &http.Client{Timeout: 30 * time.Second})
}

func NewWithHTTPClient(token, server string, verbose bool, httpClient HTTPClient) *Client {
	if server == "" {
		server = os.Getenv("scalyr_server")
		if server == "" {
			server = DefaultServer
		}
	}

	if token == "" {
		token = os.Getenv("scalyr_readlog_token")
	}

	return &Client{
		server:      strings.TrimSuffix(server, "/"),
		token:       token,
		httpClient:  httpClient,
		verbose:     verbose,
		retryPolicy: defaultRetryPolicy(),
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

// SetRetryPolicy overrides the default retry policy. Setting MaxRetries to 0
// disables retries entirely.
func (c *Client) SetRetryPolicy(policy RetryPolicy) {
	c.retryPolicy = policy
}

// drainAndClose discards any remaining bytes in the response body so the
// underlying TCP connection can be reused, then closes it.
func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}

func (c *Client) makeRequest(ctx context.Context, endpoint string, params map[string]interface{}) (*http.Response, error) {
	if c.token == "" {
		return nil, errors.NewAuthError("API token is required", nil)
	}

	if _, err := url.Parse(c.server); err != nil {
		return nil, errors.NewConfigError(fmt.Sprintf("invalid server URL '%s'", c.server), err)
	}

	params["token"] = c.token

	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, errors.NewParseError("failed to marshal request data", err)
	}

	requestURL := fmt.Sprintf("%s/api/%s", c.server, endpoint)
	if c.verbose {
		logging.WithFields(map[string]any{
			"url":      requestURL,
			"endpoint": endpoint,
		}).Debug("Making HTTP request")
		redactedJSON, err := json.Marshal(redactSensitiveParams(params))
		if err != nil {
			logging.WithField("error", err).Debug("Failed to marshal redacted request payload for logging")
		} else {
			logging.WithField("request_data", string(redactedJSON)).Debug("Request payload")
		}
	}

	policy := c.retryPolicy
	if policy.MaxRetries < 0 {
		policy.MaxRetries = 0
	}

	var (
		resp        *http.Response
		execErr     error
		lastStatus  int
		lastBodyStr string
	)

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := policy.backoffDelay(attempt - 1)
			if resp != nil {
				if ra := parseRetryAfter(resp.Header.Get("Retry-After")); ra > 0 {
					delay = ra
					if delay > policy.MaxDelay {
						delay = policy.MaxDelay
					}
				}
				drainAndClose(resp.Body)
				resp = nil
			}
			if c.verbose {
				logging.WithFields(map[string]any{
					"attempt":     attempt,
					"delay":       delay.String(),
					"max_retries": policy.MaxRetries,
				}).Debug("Retrying request after transient failure")
			}
			if err := sleepWithContext(ctx, delay); err != nil {
				return nil, errors.NewContextError("request was cancelled or timed out", err)
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewReader(jsonData))
		if err != nil {
			return nil, errors.NewNetworkError("failed to create request", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, execErr = c.httpClient.Do(req)
		if execErr != nil {
			if ctx.Err() != nil {
				return nil, errors.NewContextError("request was cancelled or timed out", ctx.Err())
			}
			if attempt < policy.MaxRetries {
				if c.verbose {
					logging.WithFields(map[string]any{
						"attempt": attempt + 1,
						"error":   execErr.Error(),
					}).Debug("Transient network error, will retry")
				}
				continue
			}
			return nil, errors.NewNetworkError("failed to execute request", execErr)
		}

		if !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		lastStatus = resp.StatusCode
		if c.verbose {
			snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			lastBodyStr = string(snippet)
			logging.WithFields(map[string]any{
				"status":  resp.StatusCode,
				"attempt": attempt + 1,
				"body":    lastBodyStr,
			}).Debug("Retryable HTTP status received")
		}

		if attempt >= policy.MaxRetries {
			drainAndClose(resp.Body)
			msg := fmt.Sprintf("server returned status %d after %d attempts", lastStatus, attempt+1)
			if lastBodyStr != "" {
				msg = fmt.Sprintf("%s: %s", msg, lastBodyStr)
			}
			return nil, errors.NewNetworkError(msg, nil)
		}
	}

	return resp, execErr
}

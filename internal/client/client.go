package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
)

type Client struct {
	server     string
	token      string
	httpClient HTTPClient
	verbose    bool
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
		server:     strings.TrimSuffix(server, "/"),
		token:      token,
		httpClient: httpClient,
		verbose:    verbose,
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
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

	url := fmt.Sprintf("%s/api/%s", c.server, endpoint)
	if c.verbose {
		logging.WithFields(map[string]any{
			"url":      url,
			"endpoint": endpoint,
		}).Debug("Making HTTP request")
		logging.WithField("request_data", string(jsonData)).Debug("Request payload")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.NewNetworkError("failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if the error is context-related
		if ctx.Err() != nil {
			return nil, errors.NewContextError("request was cancelled or timed out", ctx.Err())
		}
		return nil, errors.NewNetworkError("failed to execute request", err)
	}

	return resp, nil
}

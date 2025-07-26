package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	DefaultServer = "https://www.scalyr.com"
	APIVersion    = "v1"
)

type Client struct {
	server     string
	token      string
	httpClient *http.Client
	verbose    bool
}

func New(token, server string, verbose bool) *Client {
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
		httpClient: &http.Client{Timeout: 30 * time.Second},
		verbose:    verbose,
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) makeRequest(ctx context.Context, endpoint string, params map[string]interface{}) (*http.Response, error) {
	if c.token == "" {
		return nil, fmt.Errorf("API token is required. Set scalyr_readlog_token environment variable or use --token flag")
	}

	params["token"] = c.token

	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	url := fmt.Sprintf("%s/api/%s", c.server, endpoint)
	if c.verbose {
		fmt.Fprintf(os.Stderr, "Making request to: %s\n", url)
		fmt.Fprintf(os.Stderr, "Request data: %s\n", string(jsonData))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

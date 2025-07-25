package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func (c *Client) Query(ctx context.Context, params QueryParams) (*QueryResponse, error) {
	requestParams := map[string]interface{}{
		"queryType": "log",
	}

	if params.Filter != "" {
		requestParams["filter"] = params.Filter
	}
	if params.StartTime != "" {
		requestParams["startTime"] = params.StartTime
	}
	if params.EndTime != "" {
		requestParams["endTime"] = params.EndTime
	}
	if params.Count > 0 {
		requestParams["maxCount"] = params.Count
	}
	if params.Mode != "" {
		requestParams["pageMode"] = params.Mode
	}
	if params.Columns != "" {
		requestParams["columns"] = params.Columns
	}
	if params.Priority != "" {
		requestParams["priority"] = params.Priority
	}

	resp, err := c.makeRequest(ctx, "query", requestParams)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
	}

	var result QueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

func (c *Client) PowerQuery(ctx context.Context, params PowerQueryParams) (*PowerQueryResponse, error) {
	requestParams := map[string]interface{}{
		"queryType": "powerQuery",
		"query":     params.Query,
		"startTime": params.StartTime,
	}

	if params.EndTime != "" {
		requestParams["endTime"] = params.EndTime
	}
	if params.Priority != "" {
		requestParams["priority"] = params.Priority
	}

	resp, err := c.makeRequest(ctx, "powerQuery", requestParams)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
	}

	var result PowerQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

func (c *Client) NumericQuery(ctx context.Context, params NumericQueryParams) (*NumericQueryResponse, error) {
	requestParams := map[string]interface{}{
		"queryType": "numeric",
		"startTime": params.StartTime,
	}

	if params.Filter != "" {
		requestParams["filter"] = params.Filter
	}
	if params.Function != "" {
		requestParams["function"] = params.Function
	}
	if params.EndTime != "" {
		requestParams["endTime"] = params.EndTime
	}
	if params.Buckets > 0 {
		requestParams["buckets"] = params.Buckets
	}
	if params.Priority != "" {
		requestParams["priority"] = params.Priority
	}

	resp, err := c.makeRequest(ctx, "numericQuery", requestParams)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
	}

	var result NumericQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

func (c *Client) FacetQuery(ctx context.Context, params FacetQueryParams) (*FacetQueryResponse, error) {
	requestParams := map[string]interface{}{
		"queryType": "facet",
		"field":     params.Field,
		"startTime": params.StartTime,
	}

	if params.Filter != "" {
		requestParams["filter"] = params.Filter
	}
	if params.EndTime != "" {
		requestParams["endTime"] = params.EndTime
	}
	if params.Count > 0 {
		requestParams["maxCount"] = params.Count
	}
	if params.Priority != "" {
		requestParams["priority"] = params.Priority
	}

	resp, err := c.makeRequest(ctx, "facetQuery", requestParams)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
	}

	var result FacetQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

func (c *Client) TimeseriesQuery(ctx context.Context, params TimeseriesQueryParams) (*NumericQueryResponse, error) {
	requestParams := map[string]interface{}{
		"queryType": "numeric",
		"startTime": params.StartTime,
	}

	if params.Filter != "" {
		requestParams["filter"] = params.Filter
	}
	if params.Function != "" {
		requestParams["function"] = params.Function
	}
	if params.EndTime != "" {
		requestParams["endTime"] = params.EndTime
	}
	if params.Buckets > 0 {
		requestParams["buckets"] = params.Buckets
	}
	if params.Priority != "" {
		requestParams["priority"] = params.Priority
	}
	if params.OnlyUseSummaries {
		requestParams["onlyUseSummaries"] = true
	}
	if params.NoCreateSummaries {
		requestParams["noCreateSummaries"] = true
	}

	resp, err := c.makeRequest(ctx, "timeseriesQuery", requestParams)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
	}

	var result NumericQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

func (c *Client) Tail(ctx context.Context, params TailParams, outputChan chan<- LogEvent) error {
	defer close(outputChan)

	requestParams := map[string]interface{}{
		"queryType": "log",
		"pageMode":  "tail",
		"maxCount":  params.Lines,
	}

	if params.Filter != "" {
		requestParams["filter"] = params.Filter
	}
	if params.Priority != "" {
		requestParams["priority"] = params.Priority
	}

	resp, err := c.makeRequest(ctx, "query", requestParams)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
	}

	var result QueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		return fmt.Errorf("API error: %s", result.Message)
	}

	for _, event := range result.Matches {
		outputChan <- event
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			requestParams["continuationToken"] = result.ContinuationToken
			requestParams["maxCount"] = 1000

			resp, err := c.makeRequest(ctx, "query", requestParams)
			if err != nil {
				return err
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				resp.Body.Close()
				return fmt.Errorf("failed to read response body: %w", err)
			}
			resp.Body.Close()

			if err := json.Unmarshal(body, &result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			if result.Status != "success" {
				return fmt.Errorf("API error: %s", result.Message)
			}

			for _, event := range result.Matches {
				outputChan <- event
			}
		}
	}
}

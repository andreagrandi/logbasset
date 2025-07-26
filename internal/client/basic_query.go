package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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

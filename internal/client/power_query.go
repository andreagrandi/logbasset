package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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

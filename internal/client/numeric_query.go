package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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

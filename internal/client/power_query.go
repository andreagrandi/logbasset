package client

import (
	"context"
	"encoding/json"
	"io"

	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/andreagrandi/logbasset/internal/logging"
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
		return nil, errors.NewNetworkError("failed to read response body", err)
	}

	if c.verbose {
		logging.WithField("response_body", string(body)).Debug("API response received")
	}

	var result PowerQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.NewParseError("failed to parse response", err)
	}

	if result.Status != "success" {
		return nil, errors.NewAPIError(result.Message, nil)
	}

	return &result, nil
}

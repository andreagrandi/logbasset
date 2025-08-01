package client

import (
	"context"
	"encoding/json"
	"io"

	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/andreagrandi/logbasset/internal/logging"
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
		return nil, errors.NewNetworkError("failed to read response body", err)
	}

	if c.verbose {
		logging.WithField("response_body", string(body)).Debug("API response received")
	}

	var result QueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.NewParseError("failed to parse response", err)
	}

	if result.Status != "success" {
		return nil, errors.NewAPIError(result.Message, nil)
	}

	return &result, nil
}

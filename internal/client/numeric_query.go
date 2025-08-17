package client

import (
	"context"
	"encoding/json"
	"io"

	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/andreagrandi/logbasset/internal/logging"
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
		return nil, errors.NewNetworkError("failed to read response body", err)
	}

	if c.verbose {
		logging.WithField("response_body", string(body)).Debug("API response received")
	}

	var result NumericQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.NewParseError("failed to parse response", err)
	}

	if result.Status != "success" {
		return nil, errors.NewAPIError(result.Message, nil)
	}

	return &result, nil
}

package client

import (
	"context"
	"encoding/json"
	"io"

	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/andreagrandi/logbasset/internal/logging"
)

func (c *Client) TimeseriesQuery(ctx context.Context, params TimeseriesQueryParams) (*TimeseriesQueryResponse, error) {
	// Build the inner query object per Scalyr API spec
	query := map[string]interface{}{
		"queryType": "numeric",
		"startTime": params.StartTime,
	}

	if params.Filter != "" {
		query["filter"] = params.Filter
	}
	if params.Function != "" {
		query["function"] = params.Function
	}
	if params.EndTime != "" {
		query["endTime"] = params.EndTime
	}
	if params.Buckets > 0 {
		query["buckets"] = params.Buckets
	}
	if params.Priority != "" {
		query["priority"] = params.Priority
	}
	if params.OnlyUseSummaries {
		query["onlyUseSummaries"] = true
	}
	// createSummaries is the inverse of NoCreateSummaries
	if params.NoCreateSummaries {
		query["createSummaries"] = false
	}

	// Wrap query in queries array per Scalyr API spec
	requestParams := map[string]interface{}{
		"queries": []map[string]interface{}{query},
	}

	resp, err := c.makeRequest(ctx, "timeseriesQuery", requestParams)
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

	var result TimeseriesQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.NewParseError("failed to parse response", err)
	}

	if result.Status != "success" {
		return nil, errors.NewAPIError(result.Message, nil)
	}

	return &result, nil
}

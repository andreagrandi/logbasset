package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/andreagrandi/logbasset/internal/errors"
)

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
		return nil, errors.NewNetworkError("failed to read response body", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
	}

	var result FacetQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.NewParseError("failed to parse response", err)
	}

	if result.Status != "success" {
		return nil, errors.NewAPIError(result.Message, nil)
	}

	return &result, nil
}

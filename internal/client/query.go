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

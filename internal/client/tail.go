package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

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

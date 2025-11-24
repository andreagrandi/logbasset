package client

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/andreagrandi/logbasset/internal/logging"
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
		return errors.NewNetworkError("failed to read response body", err)
	}

	if c.verbose {
		logging.WithField("response_body", string(body)).Debug("API response received")
	}

	var result QueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return errors.NewParseError("failed to parse response", err)
	}

	if result.Status != "success" {
		return errors.NewAPIError(result.Message, nil)
	}

	for _, event := range result.Matches {
		outputChan <- event
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			// Create new params map for continuation - don't reuse original
			// to avoid sending conflicting parameters like pageMode with continuationToken
			continuationParams := map[string]interface{}{
				"queryType":         "log",
				"continuationToken": result.ContinuationToken,
				"maxCount":          1000,
			}

			// Only include priority if it was specified
			if params.Priority != "" {
				continuationParams["priority"] = params.Priority
			}

			resp, err := c.makeRequest(ctx, "query", continuationParams)
			if err != nil {
				return err
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				resp.Body.Close()
				return errors.NewNetworkError("failed to read response body", err)
			}
			resp.Body.Close()

			if err := json.Unmarshal(body, &result); err != nil {
				return errors.NewParseError("failed to parse response", err)
			}

			if result.Status != "success" {
				return errors.NewAPIError(result.Message, nil)
			}

			for _, event := range result.Matches {
				outputChan <- event
			}
		}
	}
}

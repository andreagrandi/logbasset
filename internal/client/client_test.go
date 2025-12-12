package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Clear environment variables for clean test
	os.Unsetenv("scalyr_server")
	os.Unsetenv("scalyr_readlog_token")

	tests := []struct {
		name     string
		token    string
		server   string
		verbose  bool
		expected *Client
	}{
		{
			name:    "default server",
			token:   "test-token",
			server:  "",
			verbose: false,
			expected: &Client{
				server:  DefaultServer,
				token:   "test-token",
				verbose: false,
			},
		},
		{
			name:    "custom server",
			token:   "test-token",
			server:  "https://custom.scalyr.com",
			verbose: true,
			expected: &Client{
				server:  "https://custom.scalyr.com",
				token:   "test-token",
				verbose: true,
			},
		},
		{
			name:    "server with trailing slash",
			token:   "test-token",
			server:  "https://custom.scalyr.com/",
			verbose: false,
			expected: &Client{
				server:  "https://custom.scalyr.com",
				token:   "test-token",
				verbose: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(tt.token, tt.server, tt.verbose)
			assert.Equal(t, tt.expected.server, client.server)
			assert.Equal(t, tt.expected.token, client.token)
			assert.Equal(t, tt.expected.verbose, client.verbose)
			assert.NotNil(t, client.httpClient)
		})
	}
}

func TestNewWithEnvironmentVariables(t *testing.T) {
	// Test with environment variables
	os.Setenv("scalyr_server", "https://env.scalyr.com")
	os.Setenv("scalyr_readlog_token", "env-token")
	defer func() {
		os.Unsetenv("scalyr_server")
		os.Unsetenv("scalyr_readlog_token")
	}()

	client := New("", "", false)
	assert.Equal(t, "https://env.scalyr.com", client.server)
	assert.Equal(t, "env-token", client.token)
}

func TestNewWithHTTPClient(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewWithHTTPClient("token", "https://test.com", true, mockClient)

	assert.Equal(t, "token", client.token)
	assert.Equal(t, "https://test.com", client.server)
	assert.True(t, client.verbose)
	assert.Equal(t, mockClient, client.httpClient)
}

func TestClient_SetToken(t *testing.T) {
	client := New("old-token", "", false)
	client.SetToken("new-token")
	assert.Equal(t, "new-token", client.token)
}

func TestClient_makeRequest_MissingToken(t *testing.T) {
	client := New("", "", false)
	ctx := context.Background()

	_, err := client.makeRequest(ctx, "query", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API token is required")
}

func TestClient_makeRequest_InvalidServerURL(t *testing.T) {
	client := NewWithHTTPClient("token", "://invalid-url", false, &MockHTTPClient{})
	ctx := context.Background()

	_, err := client.makeRequest(ctx, "query", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid server URL")
}

func TestClient_makeRequest_NetworkError(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	client := NewWithHTTPClient("token", "https://test.com", false, mockClient)
	ctx := context.Background()

	_, err := client.makeRequest(ctx, "query", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute request")
}

func TestClient_makeRequest_Success(t *testing.T) {
	responseBody := `{"status": "success"}`
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request structure
			assert.Equal(t, "POST", req.Method)
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			assert.Contains(t, req.URL.String(), "/api/query")

			// Verify request body contains token
			body, _ := io.ReadAll(req.Body)
			var reqData map[string]interface{}
			json.Unmarshal(body, &reqData)
			assert.Equal(t, "test-token", reqData["token"])

			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(responseBody)),
			}, nil
		},
	}

	client := NewWithHTTPClient("test-token", "https://test.com", false, mockClient)
	ctx := context.Background()

	resp, err := client.makeRequest(ctx, "query", map[string]interface{}{"filter": "test"})
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestClient_Query(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/query", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"matches": [
				{
					"timestamp": "2023-01-01T00:00:00Z",
					"severity": 3,
					"message": "Test log message",
					"thread": "main"
				}
			]
		}`))
	}))
	defer server.Close()

	client := New("test-token", server.URL, false)

	params := QueryParams{
		Filter:    "test filter",
		StartTime: "1h",
		Count:     10,
	}

	ctx := context.Background()
	result, err := client.Query(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)
	assert.Len(t, result.Matches, 1)
	assert.Equal(t, "Test log message", result.Matches[0].Message)
	assert.Equal(t, 3, result.Matches[0].Severity)
}

func TestClient_Query_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "error",
			"message": "Invalid query"
		}`))
	}))
	defer server.Close()

	client := New("test-token", server.URL, false)

	params := QueryParams{
		Filter: "invalid filter",
	}

	ctx := context.Background()
	_, err := client.Query(ctx, params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid query")
}

func TestClient_Query_ParseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := New("test-token", server.URL, false)
	ctx := context.Background()

	_, err := client.Query(ctx, QueryParams{Filter: "test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestClient_NumericQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/numericQuery", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"values": [1.5, 2.0, 3.5]
		}`))
	}))
	defer server.Close()

	client := New("test-token", server.URL, false)

	params := NumericQueryParams{
		Filter:    "test filter",
		Function:  "mean(responseTime)",
		StartTime: "1h",
		Buckets:   3,
	}

	ctx := context.Background()
	result, err := client.NumericQuery(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)
	assert.Equal(t, []float64{1.5, 2.0, 3.5}, result.Values)
}

func TestClient_FacetQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/facetQuery", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"values": [
				{"value": "/index.html", "count": 100},
				{"value": "/about.html", "count": 50}
			]
		}`))
	}))
	defer server.Close()

	client := New("test-token", server.URL, false)

	params := FacetQueryParams{
		Filter:    "test filter",
		Field:     "uriPath",
		StartTime: "1h",
		Count:     10,
	}

	ctx := context.Background()
	result, err := client.FacetQuery(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)
	assert.Len(t, result.Values, 2)
	assert.Equal(t, "/index.html", result.Values[0].Value)
	assert.Equal(t, 100, result.Values[0].Count)
}

func TestClient_PowerQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/powerQuery", r.URL.Path)

		// Verify request body contains correct queryType
		body, _ := io.ReadAll(r.Body)
		var reqData map[string]interface{}
		json.Unmarshal(body, &reqData)
		assert.Equal(t, "complex", reqData["queryType"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"matchingEvents": 150,
			"omittedEvents": 0,
			"columns": [{"name": "path"}, {"name": "count"}],
			"values": [["/index.html", 100], ["/about.html", 50]],
			"warnings": []
		}`))
	}))
	defer server.Close()

	client := New("test-token", server.URL, false)

	params := PowerQueryParams{
		Query:     "dataset='accesslog' | group count() by uriPath",
		StartTime: "1h",
	}

	ctx := context.Background()
	result, err := client.PowerQuery(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)
	assert.Len(t, result.Columns, 2)
	assert.Equal(t, "path", result.Columns[0].Name)
	assert.Equal(t, "count", result.Columns[1].Name)
	assert.Len(t, result.Values, 2)
}

func TestClient_TimeseriesQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/timeseriesQuery", r.URL.Path)

		// Verify request body contains queries array per Scalyr API spec
		body, _ := io.ReadAll(r.Body)
		var reqData map[string]interface{}
		json.Unmarshal(body, &reqData)

		// Verify queries array exists and contains the query
		queries, ok := reqData["queries"].([]interface{})
		assert.True(t, ok, "queries should be an array")
		assert.Len(t, queries, 1, "queries should contain one query")

		query := queries[0].(map[string]interface{})
		assert.Equal(t, "numeric", query["queryType"])
		assert.Equal(t, true, query["onlyUseSummaries"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"results": [{"values": [1.0, 2.0, 3.0]}]
		}`))
	}))
	defer server.Close()

	client := New("test-token", server.URL, false)

	params := TimeseriesQueryParams{
		Filter:           "test filter",
		Function:         "mean(responseTime)",
		StartTime:        "1h",
		Buckets:          3,
		OnlyUseSummaries: true,
	}

	ctx := context.Background()
	result, err := client.TimeseriesQuery(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, []float64{1.0, 2.0, 3.0}, result.Results[0].Values)
}

func TestClient_Tail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/query", r.URL.Path)

		// Verify request body contains tail-specific parameters
		body, _ := io.ReadAll(r.Body)
		var reqData map[string]interface{}
		json.Unmarshal(body, &reqData)
		assert.Equal(t, "log", reqData["queryType"])
		assert.Equal(t, "tail", reqData["pageMode"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"matches": [
				{
					"timestamp": "2023-01-01T00:00:00Z",
					"severity": 3,
					"message": "Tail log message",
					"thread": "main"
				}
			],
			"continuationToken": "test-token"
		}`))
	}))
	defer server.Close()

	client := New("test-token", server.URL, false)

	params := TailParams{
		Filter: "test filter",
		Lines:  100,
	}

	ctx, cancel := context.WithCancel(context.Background())
	outputChan := make(chan LogEvent, 10)

	// Cancel after receiving first event to prevent infinite loop
	go func() {
		<-outputChan
		cancel()
	}()

	err := client.Tail(ctx, params, outputChan)
	assert.Error(t, err) // Should error due to context cancellation
	assert.Equal(t, context.Canceled, err)
}

func TestTail_PaginationParameters(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/query", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var reqData map[string]interface{}
		json.Unmarshal(body, &reqData)

		requestCount++
		if requestCount == 1 {
			// First request should have pageMode and NO continuationToken
			assert.Equal(t, "log", reqData["queryType"])
			assert.Equal(t, "tail", reqData["pageMode"])
			assert.Equal(t, "test filter", reqData["filter"])
			assert.Equal(t, "high", reqData["priority"])
			assert.NotContains(t, reqData, "continuationToken")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"status": "success",
				"matches": [{"timestamp": "2023-01-01T00:00:00Z", "severity": 3, "message": "First"}],
				"continuationToken": "token-123"
			}`))
		} else {
			// Second request should repeat pageMode and filter along with continuationToken
			// per Scalyr API docs: "Make sure you repeat the same filter, startTime, endTime, and pageMode"
			assert.Equal(t, "log", reqData["queryType"])
			assert.Equal(t, "token-123", reqData["continuationToken"])
			assert.Equal(t, "tail", reqData["pageMode"], "pageMode should be repeated with continuationToken")
			assert.Equal(t, "test filter", reqData["filter"], "filter should be repeated with continuationToken")
			assert.Equal(t, "high", reqData["priority"])

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"status": "success",
				"matches": [{"timestamp": "2023-01-01T00:00:01Z", "severity": 3, "message": "Second"}],
				"continuationToken": "token-456"
			}`))
		}
	}))
	defer server.Close()

	client := New("test-token", server.URL, false)

	params := TailParams{
		Filter:   "test filter",
		Lines:    100,
		Priority: "high",
	}

	ctx, cancel := context.WithCancel(context.Background())
	outputChan := make(chan LogEvent, 10)

	eventsReceived := 0
	go func() {
		for range outputChan {
			eventsReceived++
			if eventsReceived == 2 {
				cancel()
			}
		}
	}()

	err := client.Tail(ctx, params, outputChan)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, 2, requestCount, "Should have made exactly 2 requests")
}

func TestClientInterface_Implementation(t *testing.T) {
	// Verify that Client implements ClientInterface
	var _ ClientInterface = (*Client)(nil)
}

func TestMockHTTPClient(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("test"))),
			}, nil
		},
	}

	req, _ := http.NewRequest("GET", "http://test.com", nil)
	resp, err := mockClient.Do(req)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestMockHTTPClient_DefaultBehavior(t *testing.T) {
	mockClient := &MockHTTPClient{}

	req, _ := http.NewRequest("GET", "http://test.com", nil)
	resp, err := mockClient.Do(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestClient_makeRequest_ContextCancellation(t *testing.T) {
	// Mock client that simulates a slow response
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Check if context was cancelled before processing
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			default:
				time.Sleep(100 * time.Millisecond) // Simulate slow request
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"status":"success"}`)),
				}, nil
			}
		},
	}

	client := NewWithHTTPClient("test-token", "https://test.com", false, mockClient)

	// Create a context that gets cancelled quickly
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.makeRequest(ctx, "query", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request was cancelled or timed out")
}

func TestClient_makeRequest_ContextTimeout(t *testing.T) {
	// Mock client that simulates a slow response
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Simulate a slow request
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(200 * time.Millisecond):
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"status":"success"}`)),
				}, nil
			}
		},
	}

	client := NewWithHTTPClient("test-token", "https://test.com", false, mockClient)

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.makeRequest(ctx, "query", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request was cancelled or timed out")
}

func TestClient_Query_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success", "matches": []}`))
	}))
	defer server.Close()

	client := New("test-token", server.URL, false)

	// Create a context that gets cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	params := QueryParams{Filter: "test"}
	_, err := client.Query(ctx, params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "request was cancelled or timed out")
}

func TestClient_Tail_ContextCancellation(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success", "matches": [], "continuationToken": "token123"}`))
	}))
	defer server.Close()

	client := New("test-token", server.URL, false)

	// Create a context that gets cancelled after a short time
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	eventChan := make(chan LogEvent, 10)
	params := TailParams{Lines: 10}

	err := client.Tail(ctx, params, eventChan)

	// Should return context.DeadlineExceeded
	assert.Equal(t, context.DeadlineExceeded, err)

	// Should have made at least one request
	assert.GreaterOrEqual(t, callCount, 1)
}

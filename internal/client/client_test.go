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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"columns": ["path", "count"],
			"results": [
				{"path": "/index.html", "count": 100},
				{"path": "/about.html", "count": 50}
			]
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
	assert.Equal(t, []string{"path", "count"}, result.Columns)
	assert.Len(t, result.Results, 2)
}

func TestClient_TimeseriesQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/timeseriesQuery", r.URL.Path)

		// Verify request body contains expected parameters
		body, _ := io.ReadAll(r.Body)
		var reqData map[string]interface{}
		json.Unmarshal(body, &reqData)
		assert.Equal(t, "numeric", reqData["queryType"])
		assert.Equal(t, true, reqData["onlyUseSummaries"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"values": [1.0, 2.0, 3.0]
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
	assert.Equal(t, []float64{1.0, 2.0, 3.0}, result.Values)
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

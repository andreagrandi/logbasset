package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
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

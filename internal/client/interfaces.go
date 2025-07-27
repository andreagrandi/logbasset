package client

import (
	"context"
	"net/http"
)

// HTTPClient interface for HTTP operations to enable testing with mocks
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientInterface defines the contract for the API client
type ClientInterface interface {
	Query(ctx context.Context, params QueryParams) (*QueryResponse, error)
	PowerQuery(ctx context.Context, params PowerQueryParams) (*PowerQueryResponse, error)
	NumericQuery(ctx context.Context, params NumericQueryParams) (*NumericQueryResponse, error)
	FacetQuery(ctx context.Context, params FacetQueryParams) (*FacetQueryResponse, error)
	TimeseriesQuery(ctx context.Context, params TimeseriesQueryParams) (*NumericQueryResponse, error)
	Tail(ctx context.Context, params TailParams, outputChan chan<- LogEvent) error
	SetToken(token string)
}

// MockHTTPClient is a mock implementation of HTTPClient for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return &http.Response{}, nil
}

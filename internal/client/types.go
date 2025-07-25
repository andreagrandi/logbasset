package client

type QueryParams struct {
	Filter    string
	StartTime string
	EndTime   string
	Count     int
	Mode      string
	Columns   string
	Priority  string
}

type PowerQueryParams struct {
	Query     string
	StartTime string
	EndTime   string
	Priority  string
}

type NumericQueryParams struct {
	Filter    string
	Function  string
	StartTime string
	EndTime   string
	Buckets   int
	Priority  string
}

type FacetQueryParams struct {
	Filter    string
	Field     string
	StartTime string
	EndTime   string
	Count     int
	Priority  string
}

type TimeseriesQueryParams struct {
	Filter            string
	Function          string
	StartTime         string
	EndTime           string
	Buckets           int
	Priority          string
	OnlyUseSummaries  bool
	NoCreateSummaries bool
}

type TailParams struct {
	Filter   string
	Lines    int
	Priority string
}

type LogEvent struct {
	Timestamp  string                 `json:"timestamp"`
	Severity   int                    `json:"severity"`
	Message    string                 `json:"message"`
	Thread     string                 `json:"thread,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type QueryResponse struct {
	Status            string     `json:"status"`
	Message           string     `json:"message,omitempty"`
	Matches           []LogEvent `json:"matches"`
	ContinuationToken string     `json:"continuationToken,omitempty"`
}

type PowerQueryResponse struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message,omitempty"`
	Results []map[string]interface{} `json:"results"`
	Columns []string                 `json:"columns"`
}

type NumericQueryResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message,omitempty"`
	Values  []float64 `json:"values"`
}

type FacetValue struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

type FacetQueryResponse struct {
	Status  string       `json:"status"`
	Message string       `json:"message,omitempty"`
	Values  []FacetValue `json:"values"`
}

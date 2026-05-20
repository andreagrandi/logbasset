package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Scalyr API responses for the end-to-end tests. The field names match the
// JSON tags the client unmarshals into, so a tag regression surfaces here.
const (
	mockQueryResponse = `{"status":"success","matches":[` +
		`{"timestamp":"1700000000000000000","severity":3,"message":"user logged in","thread":"main","attributes":{"host":"web-01"}},` +
		`{"timestamp":"1700000001000000000","severity":5,"message":"db connection failed"}` +
		`]}`

	mockPowerQueryResponse = `{"status":"success",` +
		`"columns":[{"name":"uriPath"},{"name":"requests"}],` +
		`"values":[["/login",100],["/home",250]]}`

	mockNumericQueryResponse = `{"status":"success","values":[1.5,2,3.14]}`

	mockFacetQueryResponse = `{"status":"success","values":[` +
		`{"value":"/index.html","count":42},{"value":"/about","count":17}]}`

	mockTimeseriesQueryResponse = `{"status":"success","results":[{"values":[10,20,30]}]}`
)

// resetCLIFlags restores every persistent and command flag to its default so
// values cannot leak between end-to-end runs of the shared rootCmd.
func resetCLIFlags() {
	reset := func(fs *pflag.FlagSet) {
		fs.VisitAll(func(f *pflag.Flag) {
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		})
	}
	reset(rootCmd.PersistentFlags())
	for _, sub := range rootCmd.Commands() {
		reset(sub.Flags())
	}
}

type cliRun struct {
	stdout  string
	request map[string]any
}

// runCLI executes rootCmd end-to-end against a mock Scalyr server that replies
// with mockResponse, and returns captured stdout plus the parsed request body.
func runCLI(t *testing.T, mockResponse string, args ...string) cliRun {
	t.Helper()

	var (
		mu      sync.Mutex
		request map[string]any
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Scalyr API requests must be POST")
		assert.True(t, strings.HasPrefix(r.URL.Path, "/api/"), "unexpected API path %q", r.URL.Path)

		raw, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		var body map[string]any
		assert.NoError(t, json.Unmarshal(raw, &body))

		mu.Lock()
		request = body
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, mockResponse)
	}))
	defer server.Close()

	resetCLIFlags()
	errors.OutputJSON = false
	defer func() { errors.OutputJSON = false }()

	fullArgs := append(append([]string{}, args...), "--token", "test-token", "--server", server.URL)

	out := captureStdout(t, func() {
		rootCmd.SetArgs(fullArgs)
		require.NoError(t, rootCmd.Execute())
	})
	rootCmd.SetArgs(nil)

	mu.Lock()
	defer mu.Unlock()
	return cliRun{stdout: out, request: request}
}

func TestE2EQueryOutputFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "multiline",
			format: "multiline",
			check: func(t *testing.T, out string) {
				expected := "Timestamp: 1700000000000000000\n" +
					"Severity: 3\n" +
					"Message: user logged in\n" +
					"Thread: main\n" +
					"Attributes:\n" +
					"  host: web-01\n" +
					"\n" +
					"Timestamp: 1700000001000000000\n" +
					"Severity: 5\n" +
					"Message: db connection failed\n"
				assert.Equal(t, expected, out)
			},
		},
		{
			name:   "singleline",
			format: "singleline",
			check: func(t *testing.T, out string) {
				expected := "1700000000000000000 [3] user logged in (thread: main) [host=web-01]\n" +
					"1700000001000000000 [5] db connection failed\n"
				assert.Equal(t, expected, out)
			},
		},
		{
			name:   "compact",
			format: "compact",
			check: func(t *testing.T, out string) {
				expected := "22:13:20 I user logged in\n" +
					"22:13:21 E db connection failed\n"
				assert.Equal(t, expected, out)
			},
		},
		{
			name:   "csv",
			format: "csv",
			check: func(t *testing.T, out string) {
				expected := "timestamp,severity,message\n" +
					"1700000000000000000,3,user logged in\n" +
					"1700000001000000000,5,db connection failed\n"
				assert.Equal(t, expected, out)
			},
		},
		{
			name:   "json",
			format: "json",
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockQueryResponse, out)
			},
		},
		{
			name:   "json-pretty",
			format: "json-pretty",
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockQueryResponse, out)
				assert.Contains(t, out, "\n  ", "json-pretty output should be indented")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := runCLI(t, mockQueryResponse, "query", "severity >= 3", "--output", tt.format)
			tt.check(t, run.stdout)
		})
	}
}

func TestE2EQueryColumnsCSV(t *testing.T) {
	run := runCLI(t, mockQueryResponse,
		"query", `$source="accessLog"`,
		"--output", "csv", "--columns", "severity,message,host")

	expected := "severity,message,host\n" +
		"3,user logged in,web-01\n" +
		"5,db connection failed,\n"
	assert.Equal(t, expected, run.stdout)
	assert.Equal(t, "severity,message,host", run.request["columns"])
}

func TestE2EQueryFieldsJSON(t *testing.T) {
	run := runCLI(t, mockQueryResponse,
		"query", "severity >= 3",
		"--output", "json", "--fields", "timestamp,message")

	expected := `[{"timestamp":"1700000000000000000","message":"user logged in"},` +
		`{"timestamp":"1700000001000000000","message":"db connection failed"}]`
	assert.JSONEq(t, expected, run.stdout)
}

// TestE2EQueryNonTTYDefaultsToJSON guards the documented behavior that piped
// output (stdout is a pipe during tests) falls back to compact JSON.
func TestE2EQueryNonTTYDefaultsToJSON(t *testing.T) {
	run := runCLI(t, mockQueryResponse, "query", "severity >= 3")
	assert.JSONEq(t, mockQueryResponse, run.stdout)
}

func TestE2EQueryRequestPayload(t *testing.T) {
	run := runCLI(t, mockQueryResponse,
		"query", `$serverHost="host100"`,
		"--count=100", "--start=24h", "--output", "json")

	assert.Equal(t, "log", run.request["queryType"])
	assert.Equal(t, `$serverHost="host100"`, run.request["filter"])
	assert.Equal(t, float64(100), run.request["maxCount"])
	assert.Equal(t, "24h", run.request["startTime"])
	assert.Equal(t, "test-token", run.request["token"])
}

func TestE2EPowerQueryOutputFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "csv",
			format: "csv",
			check: func(t *testing.T, out string) {
				assert.Equal(t, "uriPath,requests\n/login,100\n/home,250\n", out)
			},
		},
		{
			name:   "json",
			format: "json",
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockPowerQueryResponse, out)
			},
		},
		{
			name:   "json-pretty",
			format: "json-pretty",
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockPowerQueryResponse, out)
				assert.Contains(t, out, "\n  ", "json-pretty output should be indented")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := runCLI(t, mockPowerQueryResponse,
				"power-query", "dataset='accesslog' | group requests = count() by uriPath",
				"--start", "24h", "--output", tt.format)
			tt.check(t, run.stdout)
		})
	}
}

func TestE2ENumericQueryOutputFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "csv",
			format: "csv",
			check: func(t *testing.T, out string) {
				assert.Equal(t, "1.5,2,3.14\n", out)
			},
		},
		{
			name:   "json",
			format: "json",
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockNumericQueryResponse, out)
			},
		},
		{
			name:   "json-pretty",
			format: "json-pretty",
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockNumericQueryResponse, out)
				assert.Contains(t, out, "\n  ", "json-pretty output should be indented")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := runCLI(t, mockNumericQueryResponse,
				"numeric-query", `$dataset="accesslog"`,
				"--start", "24h", "--buckets", "3", "--output", tt.format)
			tt.check(t, run.stdout)
		})
	}
}

func TestE2EFacetQueryOutputFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "csv",
			format: "csv",
			check: func(t *testing.T, out string) {
				assert.Equal(t, "count,value\n42,/index.html\n17,/about\n", out)
			},
		},
		{
			name:   "json",
			format: "json",
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockFacetQueryResponse, out)
			},
		},
		{
			name:   "json-pretty",
			format: "json-pretty",
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockFacetQueryResponse, out)
				assert.Contains(t, out, "\n  ", "json-pretty output should be indented")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := runCLI(t, mockFacetQueryResponse,
				"facet-query", `$dataset="accesslog"`, "uriPath",
				"--start", "24h", "--output", tt.format)
			tt.check(t, run.stdout)
		})
	}
}

func TestE2ETimeseriesQueryOutputFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, out string)
	}{
		{
			name:   "csv",
			format: "csv",
			check: func(t *testing.T, out string) {
				assert.Equal(t, "10,20,30\n", out)
			},
		},
		{
			name:   "json",
			format: "json",
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockTimeseriesQueryResponse, out)
			},
		},
		{
			name:   "json-pretty",
			format: "json-pretty",
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockTimeseriesQueryResponse, out)
				assert.Contains(t, out, "\n  ", "json-pretty output should be indented")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := runCLI(t, mockTimeseriesQueryResponse,
				"timeseries-query", `$dataset="accesslog"`,
				"--function", "bytes", "--start", "24h", "--buckets", "3", "--output", tt.format)
			tt.check(t, run.stdout)
		})
	}
}

// TestE2EDocsExamples runs the representative command invocations from README.md
// end-to-end against mocked responses, exercising one example per command.
func TestE2EDocsExamples(t *testing.T) {
	tests := []struct {
		name     string
		response string
		args     []string
		check    func(t *testing.T, out string)
	}{
		{
			name:     "query csv with columns and count",
			response: mockQueryResponse,
			args:     []string{"query", `$source="accessLog"`, "--output=csv", "--columns=status,uriPath", "--count=1000"},
			check: func(t *testing.T, out string) {
				assert.Equal(t, "status,uriPath\n,\n,\n", out)
			},
		},
		{
			name:     "power-query json-pretty",
			response: mockPowerQueryResponse,
			args: []string{"power-query",
				"dataset = 'accesslog' | group requests = count() by uriPath | sort -requests",
				"--start", "7d", "--end", "0d", "--output=json-pretty"},
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockPowerQueryResponse, out)
				assert.Contains(t, out, "\n  ", "json-pretty output should be indented")
			},
		},
		{
			name:     "numeric-query with buckets",
			response: mockNumericQueryResponse,
			args:     []string{"numeric-query", `"/login"`, "--start", "24h", "--buckets", "24"},
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockNumericQueryResponse, out)
			},
		},
		{
			name:     "facet-query for a field",
			response: mockFacetQueryResponse,
			args:     []string{"facet-query", `$dataset="accesslog"`, "uriPath", "--start", "24h"},
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockFacetQueryResponse, out)
			},
		},
		{
			name:     "timeseries-query with function",
			response: mockTimeseriesQueryResponse,
			args:     []string{"timeseries-query", `$dataset="accesslog"`, "--function", "bytes", "--start", "24h", "--buckets", "24"},
			check: func(t *testing.T, out string) {
				assert.JSONEq(t, mockTimeseriesQueryResponse, out)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := runCLI(t, tt.response, tt.args...)
			tt.check(t, run.stdout)
		})
	}
}

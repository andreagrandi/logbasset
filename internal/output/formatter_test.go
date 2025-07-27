package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andreagrandi/logbasset/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name     string
		format   Format
		expected Formatter
	}{
		{
			name:     "JSON formatter",
			format:   JSON,
			expected: &JSONFormatter{},
		},
		{
			name:     "CSV formatter",
			format:   CSV,
			expected: &CSVFormatter{},
		},
		{
			name:     "Table formatter",
			format:   Table,
			expected: &TableFormatter{},
		},
		{
			name:     "Raw formatter",
			format:   Raw,
			expected: &RawFormatter{},
		},
		{
			name:     "Unknown format defaults to JSON",
			format:   Format("unknown"),
			expected: &JSONFormatter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFormatter(tt.format)
			assert.IsType(t, tt.expected, result)
		})
	}
}

func TestJSONFormatter_FormatLogEvents(t *testing.T) {
	formatter := &JSONFormatter{}
	events := []client.LogEvent{
		{
			Timestamp: "2023-01-01T00:00:00Z",
			Severity:  3,
			Message:   "Test message 1",
			Thread:    "main",
		},
		{
			Timestamp: "2023-01-01T00:01:00Z",
			Severity:  2,
			Message:   "Test message 2",
			Thread:    "worker",
		},
	}

	var buf bytes.Buffer
	err := formatter.FormatLogEvents(events, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Test message 1")
	assert.Contains(t, output, "Test message 2")
	assert.Contains(t, output, "2023-01-01T00:00:00Z")
}

func TestJSONFormatter_FormatPowerQueryResults(t *testing.T) {
	formatter := &JSONFormatter{}
	results := []map[string]interface{}{
		{"path": "/index.html", "count": 100},
		{"path": "/about.html", "count": 50},
	}
	columns := []string{"path", "count"}

	var buf bytes.Buffer
	err := formatter.FormatPowerQueryResults(results, columns, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "/index.html")
	assert.Contains(t, output, "100")
}

func TestJSONFormatter_FormatNumericResults(t *testing.T) {
	formatter := &JSONFormatter{}
	values := []float64{1.5, 2.0, 3.5}

	var buf bytes.Buffer
	err := formatter.FormatNumericResults(values, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "1.5")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "3.5")
}

func TestJSONFormatter_FormatFacetResults(t *testing.T) {
	formatter := &JSONFormatter{}
	values := []client.FacetValue{
		{Value: "/index.html", Count: 100},
		{Value: "/about.html", Count: 50},
	}

	var buf bytes.Buffer
	err := formatter.FormatFacetResults(values, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "/index.html")
	assert.Contains(t, output, "100")
}

func TestCSVFormatter_FormatLogEvents(t *testing.T) {
	formatter := &CSVFormatter{}
	events := []client.LogEvent{
		{
			Timestamp: "2023-01-01T00:00:00Z",
			Severity:  3,
			Message:   "Test message",
			Thread:    "main",
		},
	}

	var buf bytes.Buffer
	err := formatter.FormatLogEvents(events, &buf)

	require.NoError(t, err)
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check header
	assert.Equal(t, "timestamp,severity,message,thread", lines[0])
	// Check data
	assert.Contains(t, lines[1], "2023-01-01T00:00:00Z")
	assert.Contains(t, lines[1], "3")
	assert.Contains(t, lines[1], "Test message")
	assert.Contains(t, lines[1], "main")
}

func TestCSVFormatter_FormatPowerQueryResults(t *testing.T) {
	formatter := &CSVFormatter{}
	results := []map[string]interface{}{
		{"path": "/index.html", "count": 100},
		{"path": "/about.html", "count": 50},
	}
	columns := []string{"path", "count"}

	var buf bytes.Buffer
	err := formatter.FormatPowerQueryResults(results, columns, &buf)

	require.NoError(t, err)
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check header
	assert.Equal(t, "path,count", lines[0])
	// Check data
	assert.Contains(t, lines[1], "/index.html")
	assert.Contains(t, lines[1], "100")
}

func TestCSVFormatter_FormatNumericResults(t *testing.T) {
	formatter := &CSVFormatter{}
	values := []float64{1.5, 2.0}

	var buf bytes.Buffer
	err := formatter.FormatNumericResults(values, &buf)

	require.NoError(t, err)
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	assert.Equal(t, "value", lines[0])
	assert.Contains(t, lines[1], "1.5")
	assert.Contains(t, lines[2], "2.0")
}

func TestCSVFormatter_FormatFacetResults(t *testing.T) {
	formatter := &CSVFormatter{}
	values := []client.FacetValue{
		{Value: "/index.html", Count: 100},
	}

	var buf bytes.Buffer
	err := formatter.FormatFacetResults(values, &buf)

	require.NoError(t, err)
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	assert.Equal(t, "value,count", lines[0])
	assert.Contains(t, lines[1], "/index.html")
	assert.Contains(t, lines[1], "100")
}

func TestTableFormatter_FormatLogEvents(t *testing.T) {
	formatter := &TableFormatter{}
	events := []client.LogEvent{
		{
			Timestamp: "2023-01-01T00:00:00Z",
			Severity:  3,
			Message:   "Test message",
			Thread:    "main",
		},
	}

	var buf bytes.Buffer
	err := formatter.FormatLogEvents(events, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "TIMESTAMP")
	assert.Contains(t, output, "SEVERITY")
	assert.Contains(t, output, "Test message")
	assert.Contains(t, output, "2023-01-01 00:00:00")
}

func TestTableFormatter_FormatLogEvents_Empty(t *testing.T) {
	formatter := &TableFormatter{}
	var events []client.LogEvent

	var buf bytes.Buffer
	err := formatter.FormatLogEvents(events, &buf)

	require.NoError(t, err)
	assert.Empty(t, buf.String())
}

func TestTableFormatter_FormatPowerQueryResults(t *testing.T) {
	formatter := &TableFormatter{}
	results := []map[string]interface{}{
		{"path": "/index.html", "count": 100},
	}
	columns := []string{"path", "count"}

	var buf bytes.Buffer
	err := formatter.FormatPowerQueryResults(results, columns, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "path")
	assert.Contains(t, output, "count")
	assert.Contains(t, output, "/index.html")
	assert.Contains(t, output, "100")
}

func TestTableFormatter_FormatPowerQueryResults_Empty(t *testing.T) {
	formatter := &TableFormatter{}
	var results []map[string]interface{}
	columns := []string{"path", "count"}

	var buf bytes.Buffer
	err := formatter.FormatPowerQueryResults(results, columns, &buf)

	require.NoError(t, err)
	assert.Empty(t, buf.String())
}

func TestTableFormatter_FormatNumericResults(t *testing.T) {
	formatter := &TableFormatter{}
	values := []float64{1.5, 2.0}

	var buf bytes.Buffer
	err := formatter.FormatNumericResults(values, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "VALUE")
	assert.Contains(t, output, "1.50")
	assert.Contains(t, output, "2.00")
}

func TestTableFormatter_FormatFacetResults(t *testing.T) {
	formatter := &TableFormatter{}
	values := []client.FacetValue{
		{Value: "/index.html", Count: 100},
	}

	var buf bytes.Buffer
	err := formatter.FormatFacetResults(values, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "VALUE")
	assert.Contains(t, output, "COUNT")
	assert.Contains(t, output, "/index.html")
	assert.Contains(t, output, "100")
}

func TestRawFormatter_FormatLogEvents(t *testing.T) {
	formatter := &RawFormatter{}
	events := []client.LogEvent{
		{Message: "First message"},
		{Message: "Second message"},
	}

	var buf bytes.Buffer
	err := formatter.FormatLogEvents(events, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "First message")
	assert.Contains(t, output, "Second message")
}

func TestRawFormatter_FormatPowerQueryResults(t *testing.T) {
	formatter := &RawFormatter{}
	results := []map[string]interface{}{
		{"path": "/index.html", "count": 100},
	}
	columns := []string{"path", "count"}

	var buf bytes.Buffer
	err := formatter.FormatPowerQueryResults(results, columns, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "/index.html")
	assert.Contains(t, output, "100")
}

func TestRawFormatter_FormatPowerQueryResults_MissingColumns(t *testing.T) {
	formatter := &RawFormatter{}
	results := []map[string]interface{}{
		{"path": "/index.html"}, // missing "count" column
	}
	columns := []string{"path", "count"}

	var buf bytes.Buffer
	err := formatter.FormatPowerQueryResults(results, columns, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "/index.html")
}

func TestRawFormatter_FormatNumericResults(t *testing.T) {
	formatter := &RawFormatter{}
	values := []float64{1.5, 2.0}

	var buf bytes.Buffer
	err := formatter.FormatNumericResults(values, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "1.5")
	assert.Contains(t, output, "2.0")
}

func TestRawFormatter_FormatFacetResults(t *testing.T) {
	formatter := &RawFormatter{}
	values := []client.FacetValue{
		{Value: "/index.html", Count: 100},
	}

	var buf bytes.Buffer
	err := formatter.FormatFacetResults(values, &buf)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "/index.html: 100")
}

func TestTableFormatter_FormatLogEvents_TimestampParsing(t *testing.T) {
	formatter := &TableFormatter{}
	events := []client.LogEvent{
		{
			Timestamp: "invalid-timestamp",
			Severity:  3,
			Message:   "Test message",
			Thread:    "main",
		},
	}

	var buf bytes.Buffer
	err := formatter.FormatLogEvents(events, &buf)

	require.NoError(t, err)
	output := buf.String()
	// Should handle invalid timestamp gracefully
	assert.Contains(t, output, "Test message")
}

func TestFormatConstants(t *testing.T) {
	assert.Equal(t, Format("json"), JSON)
	assert.Equal(t, Format("csv"), CSV)
	assert.Equal(t, Format("table"), Table)
	assert.Equal(t, Format("raw"), Raw)
}

func TestFormatterInterface(t *testing.T) {
	// Verify all formatters implement the Formatter interface
	var _ Formatter = (*JSONFormatter)(nil)
	var _ Formatter = (*CSVFormatter)(nil)
	var _ Formatter = (*TableFormatter)(nil)
	var _ Formatter = (*RawFormatter)(nil)
}

func TestCSVFormatter_EmptyData(t *testing.T) {
	formatter := &CSVFormatter{}

	t.Run("empty log events", func(t *testing.T) {
		var buf bytes.Buffer
		err := formatter.FormatLogEvents([]client.LogEvent{}, &buf)
		require.NoError(t, err)
		// Should still have header
		assert.Contains(t, buf.String(), "timestamp,severity,message,thread")
	})

	t.Run("empty power query results", func(t *testing.T) {
		var buf bytes.Buffer
		err := formatter.FormatPowerQueryResults([]map[string]interface{}{}, []string{"col1"}, &buf)
		require.NoError(t, err)
		// Should still have header
		assert.Contains(t, buf.String(), "col1")
	})

	t.Run("empty numeric results", func(t *testing.T) {
		var buf bytes.Buffer
		err := formatter.FormatNumericResults([]float64{}, &buf)
		require.NoError(t, err)
		// Should still have header
		assert.Contains(t, buf.String(), "value")
	})

	t.Run("empty facet results", func(t *testing.T) {
		var buf bytes.Buffer
		err := formatter.FormatFacetResults([]client.FacetValue{}, &buf)
		require.NoError(t, err)
		// Should still have header
		assert.Contains(t, buf.String(), "value,count")
	})
}

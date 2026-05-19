package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w
	defer func() {
		os.Stdout = originalStdout
	}()

	fn()
	require.NoError(t, w.Close())

	out, err := io.ReadAll(r)
	require.NoError(t, err)

	return string(out)
}

func TestOutputJSON_Compact(t *testing.T) {
	data := map[string]any{"key": "value", "num": 42}

	out := captureStdout(t, func() {
		outputJSON(data, false)
	})

	assert.NotContains(t, out, "\n  ", "compact output should not contain indentation")
	assert.Contains(t, out, `"key":"value"`)
	assert.Contains(t, out, `"num":42`)
	assert.True(t, strings.HasSuffix(out, "\n"), "Println should append a newline")
}

func TestOutputJSON_Pretty(t *testing.T) {
	data := map[string]any{"key": "value"}

	out := captureStdout(t, func() {
		outputJSON(data, true)
	})

	assert.Contains(t, out, "  \"key\": \"value\"", "pretty output should be indented with two spaces")
	assert.Contains(t, out, "\n", "pretty output should be multi-line")
}

func TestOutputNumericCSV(t *testing.T) {
	out := captureStdout(t, func() {
		outputNumericCSV([]float64{1.5, 2, 3.14})
	})

	assert.Equal(t, "1.5,2,3.14\n", out)
}

func TestOutputNumericCSV_Empty(t *testing.T) {
	out := captureStdout(t, func() {
		outputNumericCSV([]float64{})
	})

	assert.Equal(t, "\n", out)
}

func TestFormatCompactTimestamp(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty input", in: "", want: ""},
		{name: "nanoseconds since epoch", in: "1700000000000000000", want: "22:13:20"},
		{name: "rfc3339 utc", in: "2024-05-19T07:08:09Z", want: "07:08:09"},
		{name: "rfc3339 with offset is normalised to utc", in: "2024-05-19T09:08:09+02:00", want: "07:08:09"},
		{name: "unparseable falls back to input", in: "not-a-timestamp", want: "not-a-timestamp"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatCompactTimestamp(tt.in))
		})
	}
}

func TestSeverityChar(t *testing.T) {
	tests := []struct {
		sev  int
		want string
	}{
		{sev: 0, want: "D"},
		{sev: 1, want: "D"},
		{sev: 2, want: "D"},
		{sev: 3, want: "I"},
		{sev: 4, want: "W"},
		{sev: 5, want: "E"},
		{sev: 6, want: "F"},
		{sev: 9, want: "F"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, severityChar(tt.sev), "severity %d", tt.sev)
	}
}

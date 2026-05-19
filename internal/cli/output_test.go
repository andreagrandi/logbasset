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

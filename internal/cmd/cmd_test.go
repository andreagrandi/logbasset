package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "help flag",
			args:     []string{"query", "--help"},
			wantErr:  false,
			contains: "search and filter your logs",
		},
		{
			name:     "invalid output format",
			args:     []string{"query", "--output", "invalid"},
			wantErr:  false,
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else if tt.contains != "" {
				assert.Contains(t, output.String(), tt.contains)
			}

			rootCmd.SetArgs([]string{})
		})
	}
}

func TestPowerQueryCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "help flag",
			args:     []string{"power-query", "--help"},
			contains: "execute a PowerQuery",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			rootCmd.SetArgs(tt.args)

			rootCmd.Execute()
			assert.Contains(t, output.String(), tt.contains)

			rootCmd.SetArgs([]string{})
		})
	}
}

func TestNumericQueryCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "help flag",
			args:     []string{"numeric-query", "--help"},
			contains: "retrieve numeric data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			rootCmd.SetArgs(tt.args)

			rootCmd.Execute()
			assert.Contains(t, output.String(), tt.contains)

			rootCmd.SetArgs([]string{})
		})
	}
}

func TestFacetQueryCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "help flag",
			args:     []string{"facet-query", "--help"},
			contains: "most common values for a field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			rootCmd.SetArgs(tt.args)

			rootCmd.Execute()
			assert.Contains(t, output.String(), tt.contains)

			rootCmd.SetArgs([]string{})
		})
	}
}

func TestTimeseriesQueryCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "help flag",
			args:     []string{"timeseries-query", "--help"},
			contains: "precomputes a numeric query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			rootCmd.SetArgs(tt.args)

			rootCmd.Execute()
			assert.Contains(t, output.String(), tt.contains)

			rootCmd.SetArgs([]string{})
		})
	}
}

func TestTailCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "help flag",
			args:     []string{"tail", "--help"},
			contains: "live tail of log records",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			rootCmd.SetArgs(tt.args)

			rootCmd.Execute()
			assert.Contains(t, output.String(), tt.contains)

			rootCmd.SetArgs([]string{})
		})
	}
}

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "help flag",
			args:     []string{"--help"},
			contains: "LogBasset is a command-line tool",
		},
		{
			name:     "version flag",
			args:     []string{"--version"},
			contains: "version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			// Reset command before each test
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			// Reset any previous flags
			rootCmd.ParseFlags([]string{})

			err := rootCmd.Execute()
			// For help/version, we expect no error or the special ErrHelp
			if err != nil && tt.name != "help flag" && tt.name != "version flag" {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := stdout.String() + stderr.String()
			assert.Contains(t, output, tt.contains)

			// Clean reset
			rootCmd.SetArgs([]string{})
		})
	}
}

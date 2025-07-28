package cli

import (
	"testing"

	"github.com/andreagrandi/logbasset/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestValidationIntegration(t *testing.T) {
	config := validation.DefaultConfig()

	tests := []struct {
		name      string
		params    validation.QueryValidationParams
		wantError bool
	}{
		{
			name: "valid query command parameters",
			params: validation.QueryValidationParams{
				StartTime:     "24h",
				Count:         100,
				Output:        "json",
				Priority:      "high",
				ValidateCount: true,
			},
			wantError: false,
		},
		{
			name: "invalid count for query command",
			params: validation.QueryValidationParams{
				Count:         0,
				ValidateCount: true,
			},
			wantError: true,
		},
		{
			name: "valid numeric query parameters",
			params: validation.QueryValidationParams{
				StartTime:       "1h",
				Buckets:         24,
				Output:          "csv",
				ValidateBuckets: true,
			},
			wantError: false,
		},
		{
			name: "invalid buckets for numeric query",
			params: validation.QueryValidationParams{
				Buckets:         0,
				ValidateBuckets: true,
			},
			wantError: true,
		},
		{
			name: "valid tail parameters",
			params: validation.QueryValidationParams{
				Lines:         100,
				Output:        "messageonly",
				ValidateLines: true,
			},
			wantError: false,
		},
		{
			name: "invalid lines for tail command",
			params: validation.QueryValidationParams{
				Lines:         0,
				ValidateLines: true,
			},
			wantError: true,
		},
		{
			name: "invalid time format",
			params: validation.QueryValidationParams{
				StartTime: "invalid-time-format",
			},
			wantError: true,
		},
		{
			name: "invalid output format",
			params: validation.QueryValidationParams{
				Output: "xml",
			},
			wantError: true,
		},
		{
			name: "invalid priority",
			params: validation.QueryValidationParams{
				Priority: "medium",
			},
			wantError: true,
		},
		{
			name: "invalid columns format",
			params: validation.QueryValidationParams{
				Columns: "timestamp,,message",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateQueryParams(tt.params, config)

			if tt.wantError {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for test case: %s", tt.name)
			}
		})
	}
}

func TestTimeFormatValidation(t *testing.T) {
	validTimes := []string{
		"24h", "1d", "30m", "60s",
		"2024-01-15", "2024-01-15 14:30", "2024-01-15 14:30:45",
		"14:30", "14:30:45", "2:30 PM", "2:30:45 PM",
	}

	invalidTimes := []string{
		"24x", "invalid", "-1h", "not-a-time", "24",
	}

	for _, timeStr := range validTimes {
		t.Run("valid_"+timeStr, func(t *testing.T) {
			err := validation.ValidateTimeFormat(timeStr)
			assert.NoError(t, err, "Expected %s to be valid", timeStr)
		})
	}

	for _, timeStr := range invalidTimes {
		t.Run("invalid_"+timeStr, func(t *testing.T) {
			err := validation.ValidateTimeFormat(timeStr)
			assert.Error(t, err, "Expected %s to be invalid", timeStr)
		})
	}
}

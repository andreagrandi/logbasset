package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTimeFormat(t *testing.T) {
	tests := []struct {
		name      string
		timeStr   string
		wantError bool
	}{
		{"empty string", "", false},
		{"relative hours", "24h", false},
		{"relative minutes", "30m", false},
		{"relative days", "7d", false},
		{"relative seconds", "60s", false},
		{"absolute date", "2024-01-15", false},
		{"absolute datetime", "2024-01-15 14:30", false},
		{"absolute datetime with seconds", "2024-01-15 14:30:45", false},
		{"time only 24h", "14:30", false},
		{"time only 24h with seconds", "14:30:45", false},
		{"time only 12h", "2:30 PM", false},
		{"time only 12h with seconds", "2:30:45 PM", false},
		{"invalid relative time", "24x", true},
		{"invalid format", "not-a-time", true},
		{"negative relative time", "-1h", true},
		{"missing unit", "24", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeFormat(tt.timeStr)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCount(t *testing.T) {
	tests := []struct {
		name      string
		count     int
		maxCount  int
		wantError bool
	}{
		{"valid count", 100, 5000, false},
		{"minimum count", 1, 5000, false},
		{"maximum count", 5000, 5000, false},
		{"zero count", 0, 5000, true},
		{"negative count", -1, 5000, true},
		{"exceeds maximum", 5001, 5000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCount(tt.count, tt.maxCount)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBuckets(t *testing.T) {
	tests := []struct {
		name       string
		buckets    int
		maxBuckets int
		wantError  bool
	}{
		{"valid buckets", 24, 5000, false},
		{"minimum buckets", 1, 5000, false},
		{"maximum buckets", 5000, 5000, false},
		{"zero buckets", 0, 5000, true},
		{"negative buckets", -1, 5000, true},
		{"exceeds maximum", 5001, 5000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBuckets(tt.buckets, tt.maxBuckets)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOutput(t *testing.T) {
	validOutputs := []string{"json", "csv", "multiline"}

	tests := []struct {
		name      string
		output    string
		wantError bool
	}{
		{"empty output", "", false},
		{"valid json", "json", false},
		{"valid csv", "csv", false},
		{"valid multiline", "multiline", false},
		{"invalid output", "xml", true},
		{"case sensitive", "JSON", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutput(tt.output, validOutputs)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePriority(t *testing.T) {
	validPriorities := []string{"high", "low"}

	tests := []struct {
		name      string
		priority  string
		wantError bool
	}{
		{"empty priority", "", false},
		{"valid high", "high", false},
		{"valid low", "low", false},
		{"invalid priority", "medium", true},
		{"case sensitive", "HIGH", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePriority(tt.priority, validPriorities)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMode(t *testing.T) {
	validModes := []string{"head", "tail"}

	tests := []struct {
		name      string
		mode      string
		wantError bool
	}{
		{"empty mode", "", false},
		{"valid head", "head", false},
		{"valid tail", "tail", false},
		{"invalid mode", "middle", true},
		{"case sensitive", "HEAD", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMode(tt.mode, validModes)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateColumns(t *testing.T) {
	tests := []struct {
		name      string
		columns   string
		wantError bool
	}{
		{"empty columns", "", false},
		{"single column", "timestamp", false},
		{"multiple columns", "timestamp,severity,message", false},
		{"columns with spaces", "timestamp, severity, message", false},
		{"empty column in list", "timestamp,,message", true},
		{"only spaces in column", "timestamp,   ,message", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColumns(tt.columns)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateQuerySyntax(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantError bool
	}{
		{"empty query", "", false},
		{"simple query", `severity >= 3`, false},
		{"complex query", `$serverHost="web01" AND status=404`, false},
		{"query with quotes", `message contains "error"`, false},
		{"very long query", string(make([]byte, 10001)), true},
		{"maximum length query", string(make([]byte, 10000)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuerySyntax(tt.query)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRequiredField(t *testing.T) {
	tests := []struct {
		name       string
		fieldName  string
		fieldValue string
		wantError  bool
	}{
		{"valid field", "start", "24h", false},
		{"empty field", "start", "", true},
		{"whitespace only field", "start", "   ", true},
		{"tab field", "start", "\t", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredField(tt.fieldName, tt.fieldValue)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateQueryParams(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name      string
		params    QueryValidationParams
		wantError bool
	}{
		{
			name: "valid query params",
			params: QueryValidationParams{
				StartTime:     "24h",
				EndTime:       "1h",
				Count:         100,
				Mode:          "tail",
				Columns:       "timestamp,message",
				Output:        "json",
				Priority:      "high",
				Query:         `severity >= 3`,
				ValidateCount: true,
			},
			wantError: false,
		},
		{
			name: "invalid start time",
			params: QueryValidationParams{
				StartTime: "invalid-time",
			},
			wantError: true,
		},
		{
			name: "invalid count",
			params: QueryValidationParams{
				Count:         0,
				ValidateCount: true,
			},
			wantError: true,
		},
		{
			name: "invalid output",
			params: QueryValidationParams{
				Output: "xml",
			},
			wantError: true,
		},
		{
			name: "invalid priority",
			params: QueryValidationParams{
				Priority: "medium",
			},
			wantError: true,
		},
		{
			name: "query too long",
			params: QueryValidationParams{
				Query: string(make([]byte, 10001)),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQueryParams(tt.params, config)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	require.NotNil(t, config)
	assert.Equal(t, 5000, config.MaxCount)
	assert.Equal(t, 5000, config.MaxBuckets)
	assert.Equal(t, 1000, config.MaxFacetCount)
	assert.Equal(t, 10000, config.MaxTailLines)
	assert.Contains(t, config.ValidOutputs, "json")
	assert.Contains(t, config.ValidPriorities, "high")
	assert.Contains(t, config.ValidModes, "head")
}

func TestParseRelativeTime(t *testing.T) {
	tests := []struct {
		name      string
		timeStr   string
		wantError bool
	}{
		{"hours", "24h", false},
		{"minutes", "30m", false},
		{"days", "7d", false},
		{"seconds", "60s", false},
		{"invalid unit", "24x", true},
		{"no unit", "24", true},
		{"non-numeric", "abch", true},
		{"negative", "-1h", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseRelativeTime(tt.timeStr)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

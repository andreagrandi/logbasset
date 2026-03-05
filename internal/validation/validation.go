package validation

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/andreagrandi/logbasset/internal/errors"
)

var (
	timeRegex = regexp.MustCompile(`^(\d+[dhm]|(\d{4}-\d{2}-\d{2}(\s+\d{1,2}:\d{2}(:\d{2})?(\s*(AM|PM))?)?)|(\d{1,2}:\d{2}(:\d{2})?(\s*(AM|PM))?))$`)
)

type ValidationConfig struct {
	MaxCount        int
	MaxBuckets      int
	MaxFacetCount   int
	MaxTailLines    int
	ValidOutputs    []string
	ValidPriorities []string
	ValidModes      []string
}

func DefaultConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxCount:        5000,
		MaxBuckets:      5000,
		MaxFacetCount:   1000,
		MaxTailLines:    10000,
		ValidOutputs:    []string{"multiline", "singleline", "csv", "json", "json-pretty", "messageonly"},
		ValidPriorities: []string{"high", "low"},
		ValidModes:      []string{"head", "tail"},
	}
}

func ValidateTimeFormat(timeStr string) error {
	if timeStr == "" {
		return nil
	}

	timeStr = strings.TrimSpace(timeStr)

	if timeRegex.MatchString(timeStr) {
		return nil
	}

	if _, err := parseRelativeTime(timeStr); err == nil {
		return nil
	}

	if _, err := time.Parse("2006-01-02", timeStr); err == nil {
		return nil
	}

	if _, err := time.Parse("2006-01-02 15:04", timeStr); err == nil {
		return nil
	}

	if _, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
		return nil
	}

	if _, err := time.Parse("15:04", timeStr); err == nil {
		return nil
	}

	if _, err := time.Parse("15:04:05", timeStr); err == nil {
		return nil
	}

	if _, err := time.Parse("3:04 PM", timeStr); err == nil {
		return nil
	}

	if _, err := time.Parse("3:04:05 PM", timeStr); err == nil {
		return nil
	}

	return errors.NewValidationError(
		fmt.Sprintf("invalid time format: %s", timeStr),
		fmt.Errorf("time format must be one of: relative (24h, 1d, 30m), absolute (2006-01-02 15:04:05), or time only (15:04, 3:04 PM)"),
	)
}

func parseRelativeTime(timeStr string) (time.Duration, error) {
	if len(timeStr) < 2 {
		return 0, fmt.Errorf("invalid relative time format")
	}

	unit := timeStr[len(timeStr)-1:]
	valueStr := timeStr[:len(timeStr)-1]

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value in relative time")
	}

	if value < 0 {
		return 0, fmt.Errorf("negative values not allowed in relative time")
	}

	switch unit {
	case "s":
		return time.Duration(value) * time.Second, nil
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid time unit: %s", unit)
	}
}
func ValidateCount(count int, maxCount int) error {
	if count < 1 {
		return errors.NewValidationError(
			"count must be at least 1",
			fmt.Errorf("provided count: %d", count),
		)
	}
	if count > maxCount {
		return errors.NewValidationError(
			fmt.Sprintf("count cannot exceed %d", maxCount),
			fmt.Errorf("provided count: %d", count),
		)
	}
	return nil
}

func ValidateBuckets(buckets int, maxBuckets int) error {
	if buckets < 1 {
		return errors.NewValidationError(
			"buckets must be at least 1",
			fmt.Errorf("provided buckets: %d", buckets),
		)
	}
	if buckets > maxBuckets {
		return errors.NewValidationError(
			fmt.Sprintf("buckets cannot exceed %d", maxBuckets),
			fmt.Errorf("provided buckets: %d", buckets),
		)
	}
	return nil
}

func ValidateOutput(output string, validOutputs []string) error {
	if output == "" {
		return nil
	}

	if slices.Contains(validOutputs, output) {
		return nil
	}

	return errors.NewValidationError(
		fmt.Sprintf("invalid output format: %s", output),
		fmt.Errorf("valid formats: %s", strings.Join(validOutputs, ", ")),
	)
}
func ValidatePriority(priority string, validPriorities []string) error {
	if priority == "" {
		return nil
	}

	if slices.Contains(validPriorities, priority) {
		return nil
	}

	return errors.NewValidationError(
		fmt.Sprintf("invalid priority: %s", priority),
		fmt.Errorf("valid priorities: %s", strings.Join(validPriorities, ", ")),
	)
}
func ValidateMode(mode string, validModes []string) error {
	if mode == "" {
		return nil
	}

	if slices.Contains(validModes, mode) {
		return nil
	}

	return errors.NewValidationError(
		fmt.Sprintf("invalid mode: %s", mode),
		fmt.Errorf("valid modes: %s", strings.Join(validModes, ", ")),
	)
}

// ValidateNoControlChars rejects ASCII control characters (0x00-0x1F except tab/newline/CR, and 0x7F).
func ValidateNoControlChars(input string, fieldName string) error {
	for i, r := range input {
		if r == '\t' || r == '\n' || r == '\r' {
			continue
		}
		if (r >= 0x00 && r <= 0x1F) || r == 0x7F {
			return errors.NewValidationError(
				fmt.Sprintf("%s contains invalid control character at position %d", fieldName, i),
				fmt.Errorf("control characters (except tab, newline, CR) are not allowed"),
			)
		}
	}
	return nil
}

func ValidateColumns(columns string) error {
	if columns == "" {
		return nil
	}

	columnList := strings.Split(columns, ",")
	if len(columnList) == 0 {
		return errors.NewValidationError(
			"columns cannot be empty when specified",
			nil,
		)
	}

	for _, col := range columnList {
		col = strings.TrimSpace(col)
		if col == "" {
			return errors.NewValidationError(
				"column names cannot be empty",
				nil,
			)
		}

		if err := ValidateNoControlChars(col, "column name"); err != nil {
			return err
		}

		if strings.Contains(col, "../") || strings.Contains(col, "..\\") || strings.HasPrefix(col, "/") || strings.HasPrefix(col, "\\") {
			return errors.NewValidationError(
				fmt.Sprintf("column name contains invalid path characters: %s", col),
				fmt.Errorf("column names must not contain path traversal sequences"),
			)
		}
	}

	return nil
}

// ValidateFields validates the --fields flag value.
func ValidateFields(fields string) error {
	if fields == "" {
		return nil
	}

	fieldList := strings.Split(fields, ",")
	for _, f := range fieldList {
		f = strings.TrimSpace(f)
		if f == "" {
			return errors.NewValidationError(
				"field names cannot be empty",
				nil,
			)
		}

		if err := ValidateNoControlChars(f, "field name"); err != nil {
			return err
		}

		if strings.Contains(f, "../") || strings.Contains(f, "..\\") || strings.HasPrefix(f, "/") || strings.HasPrefix(f, "\\") {
			return errors.NewValidationError(
				fmt.Sprintf("field name contains invalid path characters: %s", f),
				fmt.Errorf("field names must not contain path traversal sequences"),
			)
		}
	}

	return nil
}

func ValidateQuerySyntax(query string) error {
	if query == "" {
		return nil
	}

	query = strings.TrimSpace(query)
	if len(query) > 10000 {
		return errors.NewValidationError(
			"query is too long (maximum 10000 characters)",
			fmt.Errorf("query length: %d", len(query)),
		)
	}

	return nil
}

func ValidateRequiredField(fieldName, fieldValue string) error {
	if strings.TrimSpace(fieldValue) == "" {
		return errors.NewValidationError(
			fmt.Sprintf("%s is required", fieldName),
			nil,
		)
	}
	return nil
}

type QueryValidationParams struct {
	StartTime       string
	EndTime         string
	Count           int
	Buckets         int
	Mode            string
	Columns         string
	Output          string
	Priority        string
	Query           string
	Lines           int
	ValidateCount   bool // Whether to validate count (some commands don't use count)
	ValidateBuckets bool // Whether to validate buckets
	ValidateLines   bool // Whether to validate lines
}

func ValidateQueryParams(params QueryValidationParams, config *ValidationConfig) error {
	if err := ValidateTimeFormat(params.StartTime); err != nil {
		return err
	}

	if err := ValidateTimeFormat(params.EndTime); err != nil {
		return err
	}

	if params.ValidateCount {
		if err := ValidateCount(params.Count, config.MaxCount); err != nil {
			return err
		}
	}

	if params.ValidateBuckets {
		if err := ValidateBuckets(params.Buckets, config.MaxBuckets); err != nil {
			return err
		}
	}

	if err := ValidateMode(params.Mode, config.ValidModes); err != nil {
		return err
	}

	if err := ValidateColumns(params.Columns); err != nil {
		return err
	}

	if err := ValidateOutput(params.Output, config.ValidOutputs); err != nil {
		return err
	}

	if err := ValidatePriority(params.Priority, config.ValidPriorities); err != nil {
		return err
	}

	if err := ValidateNoControlChars(params.Query, "filter"); err != nil {
		return err
	}

	if err := ValidateQuerySyntax(params.Query); err != nil {
		return err
	}

	if params.ValidateLines {
		if err := ValidateCount(params.Lines, config.MaxTailLines); err != nil {
			return err
		}
	}

	return nil
}

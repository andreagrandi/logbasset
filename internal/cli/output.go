package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/andreagrandi/logbasset/internal/errors"
)

// formatCompactTimestamp converts a Scalyr nanosecond timestamp string into HH:MM:SS.
// Falls back to the original value if parsing fails.
func formatCompactTimestamp(ts string) string {
	if ts == "" {
		return ""
	}
	if nanos, err := strconv.ParseInt(ts, 10, 64); err == nil {
		return time.Unix(0, nanos).UTC().Format("15:04:05")
	}
	if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		return t.UTC().Format("15:04:05")
	}
	return ts
}

// severityChar maps a Scalyr severity level to a single character for compact output.
func severityChar(sev int) string {
	switch {
	case sev <= 2:
		return "D"
	case sev == 3:
		return "I"
	case sev == 4:
		return "W"
	case sev == 5:
		return "E"
	default:
		return "F"
	}
}

func outputJSON(data any, pretty bool) {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		errors.HandleErrorAndExit(errors.NewParseError("failed to marshal JSON", err))
	}

	fmt.Println(string(output))
}

func outputNumericCSV(values []float64) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	record := make([]string, len(values))
	for i, val := range values {
		record[i] = strconv.FormatFloat(val, 'f', -1, 64)
	}
	writer.Write(record)
}

package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/andreagrandi/logbasset/internal/client"
)

type Format string

const (
	JSON  Format = "json"
	CSV   Format = "csv"
	Table Format = "table"
	Raw   Format = "raw"
)

type Formatter interface {
	FormatLogEvents(events []client.LogEvent, writer io.Writer) error
	FormatPowerQueryResults(results []map[string]interface{}, columns []string, writer io.Writer) error
	FormatNumericResults(values []float64, writer io.Writer) error
	FormatFacetResults(values []client.FacetValue, writer io.Writer) error
}

type JSONFormatter struct{}
type CSVFormatter struct{}
type TableFormatter struct{}
type RawFormatter struct{}

func NewFormatter(format Format) Formatter {
	switch format {
	case JSON:
		return &JSONFormatter{}
	case CSV:
		return &CSVFormatter{}
	case Table:
		return &TableFormatter{}
	case Raw:
		return &RawFormatter{}
	default:
		return &JSONFormatter{}
	}
}

func (f *JSONFormatter) FormatLogEvents(events []client.LogEvent, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(events)
}

func (f *JSONFormatter) FormatPowerQueryResults(results []map[string]interface{}, columns []string, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func (f *JSONFormatter) FormatNumericResults(values []float64, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(values)
}

func (f *JSONFormatter) FormatFacetResults(values []client.FacetValue, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(values)
}

func (f *CSVFormatter) FormatLogEvents(events []client.LogEvent, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	csvWriter.Write([]string{"timestamp", "severity", "message", "thread"})

	for _, event := range events {
		csvWriter.Write([]string{
			event.Timestamp,
			fmt.Sprintf("%d", event.Severity),
			event.Message,
			event.Thread,
		})
	}

	return csvWriter.Error()
}

func (f *CSVFormatter) FormatPowerQueryResults(results []map[string]interface{}, columns []string, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	csvWriter.Write(columns)

	for _, result := range results {
		row := make([]string, len(columns))
		for i, col := range columns {
			if val, ok := result[col]; ok {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		csvWriter.Write(row)
	}

	return csvWriter.Error()
}

func (f *CSVFormatter) FormatNumericResults(values []float64, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	csvWriter.Write([]string{"value"})

	for _, value := range values {
		csvWriter.Write([]string{fmt.Sprintf("%f", value)})
	}

	return csvWriter.Error()
}

func (f *CSVFormatter) FormatFacetResults(values []client.FacetValue, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	csvWriter.Write([]string{"value", "count"})

	for _, facet := range values {
		csvWriter.Write([]string{facet.Value, fmt.Sprintf("%d", facet.Count)})
	}

	return csvWriter.Error()
}

func (f *TableFormatter) FormatLogEvents(events []client.LogEvent, writer io.Writer) error {
	if len(events) == 0 {
		return nil
	}

	fmt.Fprintf(writer, "%-20s %-8s %-10s %s\n", "TIMESTAMP", "SEVERITY", "THREAD", "MESSAGE")
	fmt.Fprintf(writer, "%s\n", strings.Repeat("-", 80))

	for _, event := range events {
		timestamp, _ := time.Parse(time.RFC3339, event.Timestamp)
		fmt.Fprintf(writer, "%-20s %-8d %-10s %s\n",
			timestamp.Format("2006-01-02 15:04:05"),
			event.Severity,
			event.Thread,
			event.Message)
	}

	return nil
}

func (f *TableFormatter) FormatPowerQueryResults(results []map[string]interface{}, columns []string, writer io.Writer) error {
	if len(results) == 0 {
		return nil
	}

	for _, col := range columns {
		fmt.Fprintf(writer, "%-20s ", col)
	}
	fmt.Fprintf(writer, "\n%s\n", strings.Repeat("-", len(columns)*21))

	for _, result := range results {
		for _, col := range columns {
			if val, ok := result[col]; ok {
				fmt.Fprintf(writer, "%-20v ", val)
			} else {
				fmt.Fprintf(writer, "%-20s ", "")
			}
		}
		fmt.Fprintf(writer, "\n")
	}

	return nil
}

func (f *TableFormatter) FormatNumericResults(values []float64, writer io.Writer) error {
	fmt.Fprintf(writer, "%-10s\n", "VALUE")
	fmt.Fprintf(writer, "%s\n", strings.Repeat("-", 10))

	for _, value := range values {
		fmt.Fprintf(writer, "%-10.2f\n", value)
	}

	return nil
}

func (f *TableFormatter) FormatFacetResults(values []client.FacetValue, writer io.Writer) error {
	fmt.Fprintf(writer, "%-30s %-10s\n", "VALUE", "COUNT")
	fmt.Fprintf(writer, "%s\n", strings.Repeat("-", 42))

	for _, facet := range values {
		fmt.Fprintf(writer, "%-30s %-10d\n", facet.Value, facet.Count)
	}

	return nil
}

func (f *RawFormatter) FormatLogEvents(events []client.LogEvent, writer io.Writer) error {
	for _, event := range events {
		fmt.Fprintf(writer, "%s\n", event.Message)
	}
	return nil
}

func (f *RawFormatter) FormatPowerQueryResults(results []map[string]interface{}, columns []string, writer io.Writer) error {
	for _, result := range results {
		for _, col := range columns {
			if val, ok := result[col]; ok {
				fmt.Fprintf(writer, "%v ", val)
			}
		}
		fmt.Fprintf(writer, "\n")
	}
	return nil
}

func (f *RawFormatter) FormatNumericResults(values []float64, writer io.Writer) error {
	for _, value := range values {
		fmt.Fprintf(writer, "%f\n", value)
	}
	return nil
}

func (f *RawFormatter) FormatFacetResults(values []client.FacetValue, writer io.Writer) error {
	for _, facet := range values {
		fmt.Fprintf(writer, "%s: %d\n", facet.Value, facet.Count)
	}
	return nil
}

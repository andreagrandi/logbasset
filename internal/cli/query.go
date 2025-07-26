package cli

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/andreagrandi/logbasset/internal/client"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query [filter]",
	Short: "Retrieve log data",
	Long: `Query allows you to search and filter your logs, or simply retrieve raw log data.
The capabilities are similar to the regular log view, though you can retrieve more data at once
and have several output format options.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runQuery,
}

var (
	queryStartTime string
	queryEndTime   string
	queryCount     int
	queryMode      string
	queryColumns   string
	queryOutput    string
)

func init() {
	queryCmd.Flags().StringVar(&queryStartTime, "start", "", "Start time for the query")
	queryCmd.Flags().StringVar(&queryEndTime, "end", "", "End time for the query")
	queryCmd.Flags().IntVar(&queryCount, "count", 10, "Number of log records to retrieve (1-5000)")
	queryCmd.Flags().StringVar(&queryMode, "mode", "", "Display mode: head or tail")
	queryCmd.Flags().StringVar(&queryColumns, "columns", "", "Comma-separated list of columns to display")
	queryCmd.Flags().StringVar(&queryOutput, "output", "multiline", "Output format: multiline|singleline|csv|json|json-pretty")
}

func runQuery(cmd *cobra.Command, args []string) {
	var filter string
	if len(args) > 0 {
		filter = args[0]
	}

	c := getConfig().GetClient()

	params := client.QueryParams{
		Filter:    filter,
		StartTime: queryStartTime,
		EndTime:   queryEndTime,
		Count:     queryCount,
		Mode:      queryMode,
		Columns:   queryColumns,
		Priority:  getConfig().Priority,
	}

	ctx := context.Background()
	result, err := c.Query(ctx, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch queryOutput {
	case "json":
		outputJSON(result, false)
	case "json-pretty":
		outputJSON(result, true)
	case "csv":
		outputCSV(result.Matches, queryColumns)
	case "singleline":
		outputSingleLine(result.Matches)
	default:
		outputMultiLine(result.Matches)
	}
}

func outputJSON(data interface{}, pretty bool) {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

func outputCSV(events []client.LogEvent, columns string) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	if columns == "" {
		columns = "timestamp,severity,message"
	}

	columnList := strings.Split(columns, ",")
	for i, col := range columnList {
		columnList[i] = strings.TrimSpace(col)
	}

	writer.Write(columnList)

	for _, event := range events {
		record := make([]string, len(columnList))
		for i, col := range columnList {
			switch col {
			case "timestamp":
				record[i] = event.Timestamp
			case "severity":
				record[i] = fmt.Sprintf("%d", event.Severity)
			case "message":
				record[i] = event.Message
			case "thread":
				record[i] = event.Thread
			default:
				if val, ok := event.Attributes[col]; ok {
					record[i] = fmt.Sprintf("%v", val)
				}
			}
		}
		writer.Write(record)
	}
}

func outputSingleLine(events []client.LogEvent) {
	for _, event := range events {
		fmt.Printf("%s [%d] %s", event.Timestamp, event.Severity, event.Message)
		if event.Thread != "" {
			fmt.Printf(" (thread: %s)", event.Thread)
		}
		if len(event.Attributes) > 0 {
			attrs := make([]string, 0, len(event.Attributes))
			for k, v := range event.Attributes {
				attrs = append(attrs, fmt.Sprintf("%s=%v", k, v))
			}
			fmt.Printf(" [%s]", strings.Join(attrs, ", "))
		}
		fmt.Println()
	}
}

func outputMultiLine(events []client.LogEvent) {
	for i, event := range events {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("Timestamp: %s\n", event.Timestamp)
		fmt.Printf("Severity: %d\n", event.Severity)
		fmt.Printf("Message: %s\n", event.Message)
		if event.Thread != "" {
			fmt.Printf("Thread: %s\n", event.Thread)
		}
		if len(event.Attributes) > 0 {
			fmt.Println("Attributes:")
			for k, v := range event.Attributes {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
	}
}

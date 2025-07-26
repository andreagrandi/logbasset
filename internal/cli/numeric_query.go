package cli

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/andreagrandi/logbasset/internal/client"
	"github.com/spf13/cobra"
)

var numericQueryCmd = &cobra.Command{
	Use:   "numeric-query [filter]",
	Short: "Retrieve numeric / graph data",
	Long: `Numeric-query allows you to retrieve numeric data, e.g. for graphing. You can count the rate of events
matching some criterion (e.g. error rate), or retrieve a numeric field (e.g. response size).`,
	Args: cobra.MaximumNArgs(1),
	Run:  runNumericQuery,
}

var (
	numericQueryFunction  string
	numericQueryStartTime string
	numericQueryEndTime   string
	numericQueryBuckets   int
	numericQueryOutput    string
)

func init() {
	numericQueryCmd.Flags().StringVar(&numericQueryFunction, "function", "", "Function to compute from matching events")
	numericQueryCmd.Flags().StringVar(&numericQueryStartTime, "start", "", "Start time for the query (required)")
	numericQueryCmd.Flags().StringVar(&numericQueryEndTime, "end", "", "End time for the query")
	numericQueryCmd.Flags().IntVar(&numericQueryBuckets, "buckets", 1, "Number of time buckets (1-5000)")
	numericQueryCmd.Flags().StringVar(&numericQueryOutput, "output", "csv", "Output format: csv|json|json-pretty")
	numericQueryCmd.MarkFlagRequired("start")
}

func runNumericQuery(cmd *cobra.Command, args []string) {
	checkTokenAndExit()

	var filter string
	if len(args) > 0 {
		filter = args[0]
	}

	c := client.New(token, server, verbose)

	params := client.NumericQueryParams{
		Filter:    filter,
		Function:  numericQueryFunction,
		StartTime: numericQueryStartTime,
		EndTime:   numericQueryEndTime,
		Buckets:   numericQueryBuckets,
		Priority:  priority,
	}

	ctx := context.Background()
	result, err := c.NumericQuery(ctx, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch numericQueryOutput {
	case "json":
		outputJSON(result, false)
	case "json-pretty":
		outputJSON(result, true)
	default:
		outputNumericCSV(result.Values)
	}
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

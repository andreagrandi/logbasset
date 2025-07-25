package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/andreagrandi/logbasset/internal/client"
	"github.com/spf13/cobra"
)

var timeseriesQueryCmd = &cobra.Command{
	Use:   "timeseries-query [filter]",
	Short: "Retrieve numeric / graph data from a timeseries",
	Long: `Timeseries-query precomputes a numeric query, allowing you to execute queries almost instantaneously,
and without consuming your account's query budget. This is especially useful if you are using the Scalyr API
to feed a home-built dashboard, alerting system, or other automated tool.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runTimeseriesQuery,
}

var (
	timeseriesQueryFunction          string
	timeseriesQueryStartTime         string
	timeseriesQueryEndTime           string
	timeseriesQueryBuckets           int
	timeseriesQueryOutput            string
	timeseriesQueryOnlyUseSummaries  bool
	timeseriesQueryNoCreateSummaries bool
)

func init() {
	timeseriesQueryCmd.Flags().StringVar(&timeseriesQueryFunction, "function", "", "Function to compute from matching events")
	timeseriesQueryCmd.Flags().StringVar(&timeseriesQueryStartTime, "start", "", "Start time for the query (required)")
	timeseriesQueryCmd.Flags().StringVar(&timeseriesQueryEndTime, "end", "", "End time for the query")
	timeseriesQueryCmd.Flags().IntVar(&timeseriesQueryBuckets, "buckets", 1, "Number of time buckets (1-5000)")
	timeseriesQueryCmd.Flags().StringVar(&timeseriesQueryOutput, "output", "csv", "Output format: csv|json|json-pretty")
	timeseriesQueryCmd.Flags().BoolVar(&timeseriesQueryOnlyUseSummaries, "only-use-summaries", false, "Only query summaries, not the column store")
	timeseriesQueryCmd.Flags().BoolVar(&timeseriesQueryNoCreateSummaries, "no-create-summaries", false, "Don't create summaries for this query")
	timeseriesQueryCmd.MarkFlagRequired("start")
}

func runTimeseriesQuery(cmd *cobra.Command, args []string) {
	checkTokenAndExit()

	var filter string
	if len(args) > 0 {
		filter = args[0]
	}

	c := client.New(token, server, verbose)

	params := client.TimeseriesQueryParams{
		Filter:            filter,
		Function:          timeseriesQueryFunction,
		StartTime:         timeseriesQueryStartTime,
		EndTime:           timeseriesQueryEndTime,
		Buckets:           timeseriesQueryBuckets,
		Priority:          priority,
		OnlyUseSummaries:  timeseriesQueryOnlyUseSummaries,
		NoCreateSummaries: timeseriesQueryNoCreateSummaries,
	}

	ctx := context.Background()
	result, err := c.TimeseriesQuery(ctx, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch timeseriesQueryOutput {
	case "json":
		outputJSON(result, false)
	case "json-pretty":
		outputJSON(result, true)
	default:
		outputNumericCSV(result.Values)
	}
}

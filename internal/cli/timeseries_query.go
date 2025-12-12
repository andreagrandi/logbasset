package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/andreagrandi/logbasset/internal/client"
	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/andreagrandi/logbasset/internal/validation"
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
	var filter string
	if len(args) > 0 {
		filter = args[0]
	}

	// Validate inputs
	validationConfig := validation.DefaultConfig()
	params := validation.QueryValidationParams{
		StartTime:       timeseriesQueryStartTime,
		EndTime:         timeseriesQueryEndTime,
		Buckets:         timeseriesQueryBuckets,
		Output:          timeseriesQueryOutput,
		Priority:        getConfig().Priority,
		Query:           filter,
		ValidateBuckets: true,
	}

	if err := validation.ValidateQueryParams(params, validationConfig); err != nil {
		errors.HandleErrorAndExit(err)
	}

	// Validate required field
	if err := validation.ValidateRequiredField("start", timeseriesQueryStartTime); err != nil {
		errors.HandleErrorAndExit(err)
	}

	c := getConfig().GetClient()

	clientParams := client.TimeseriesQueryParams{
		Filter:            filter,
		Function:          timeseriesQueryFunction,
		StartTime:         timeseriesQueryStartTime,
		EndTime:           timeseriesQueryEndTime,
		Buckets:           timeseriesQueryBuckets,
		Priority:          getConfig().Priority,
		OnlyUseSummaries:  timeseriesQueryOnlyUseSummaries,
		NoCreateSummaries: timeseriesQueryNoCreateSummaries,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), getTimeout())
	defer cancel()

	// Set up signal handling for graceful cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	result, err := c.TimeseriesQuery(ctx, clientParams)
	if err != nil {
		errors.HandleErrorAndExit(err)
	}

	// Extract values from first result in the results array
	var values []float64
	if len(result.Results) > 0 {
		values = result.Results[0].Values
	}

	switch timeseriesQueryOutput {
	case "json":
		outputJSON(result, false)
	case "json-pretty":
		outputJSON(result, true)
	default:
		outputNumericCSV(values)
	}
}

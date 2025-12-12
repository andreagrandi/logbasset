package cli

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/andreagrandi/logbasset/internal/client"
	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/andreagrandi/logbasset/internal/validation"
	"github.com/spf13/cobra"
)

var powerQueryCmd = &cobra.Command{
	Use:   "power-query [query]",
	Short: "Execute PowerQuery",
	Long: `Power-query allows you to execute a PowerQuery. The capabilities are similar to the 
regular PowerQuery page, though you can retrieve more data at once and have several output format options.`,
	Args: cobra.ExactArgs(1),
	Run:  runPowerQuery,
}

var (
	powerQueryStartTime string
	powerQueryEndTime   string
	powerQueryOutput    string
)

func init() {
	powerQueryCmd.Flags().StringVar(&powerQueryStartTime, "start", "", "Start time for the query (required)")
	powerQueryCmd.Flags().StringVar(&powerQueryEndTime, "end", "", "End time for the query")
	powerQueryCmd.Flags().StringVar(&powerQueryOutput, "output", "csv", "Output format: csv|json|json-pretty")
	powerQueryCmd.MarkFlagRequired("start")
}

func runPowerQuery(cmd *cobra.Command, args []string) {
	query := args[0]

	// Validate inputs
	validationConfig := validation.DefaultConfig()
	params := validation.QueryValidationParams{
		StartTime: powerQueryStartTime,
		EndTime:   powerQueryEndTime,
		Output:    powerQueryOutput,
		Priority:  getConfig().Priority,
		Query:     query,
	}

	if err := validation.ValidateQueryParams(params, validationConfig); err != nil {
		errors.HandleErrorAndExit(err)
	}

	// Validate required field
	if err := validation.ValidateRequiredField("start", powerQueryStartTime); err != nil {
		errors.HandleErrorAndExit(err)
	}

	c := getConfig().GetClient()

	clientParams := client.PowerQueryParams{
		Query:     query,
		StartTime: powerQueryStartTime,
		EndTime:   powerQueryEndTime,
		Priority:  getConfig().Priority,
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

	result, err := c.PowerQuery(ctx, clientParams)
	if err != nil {
		errors.HandleErrorAndExit(err)
	}

	switch powerQueryOutput {
	case "json":
		outputJSON(result, false)
	case "json-pretty":
		outputJSON(result, true)
	default:
		outputPowerQueryCSV(result)
	}
}

func outputPowerQueryCSV(result *client.PowerQueryResponse) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Extract column names from column objects
	if len(result.Columns) > 0 {
		columnNames := make([]string, len(result.Columns))
		for i, col := range result.Columns {
			columnNames[i] = col.Name
		}
		writer.Write(columnNames)
	}

	// Output values (array of arrays)
	for _, row := range result.Values {
		record := make([]string, len(row))
		for i, val := range row {
			record[i] = fmt.Sprintf("%v", val)
		}
		writer.Write(record)
	}
}

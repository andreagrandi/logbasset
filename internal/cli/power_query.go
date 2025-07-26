package cli

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/andreagrandi/logbasset/internal/client"
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
	c := getConfig().GetClient()

	params := client.PowerQueryParams{
		Query:     query,
		StartTime: powerQueryStartTime,
		EndTime:   powerQueryEndTime,
		Priority:  getConfig().Priority,
	}

	ctx := context.Background()
	result, err := c.PowerQuery(ctx, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
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

	if len(result.Columns) > 0 {
		writer.Write(result.Columns)
	}

	for _, row := range result.Results {
		record := make([]string, len(result.Columns))
		for i, col := range result.Columns {
			if val, ok := row[col]; ok {
				record[i] = fmt.Sprintf("%v", val)
			}
		}
		writer.Write(record)
	}
}

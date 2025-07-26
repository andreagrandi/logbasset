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

var facetQueryCmd = &cobra.Command{
	Use:   "facet-query [filter] [field]",
	Short: "Retrieve common values for a field",
	Long: `Facet-query allows you to retrieve the most common values for a field. For instance, you can find
the most common URLs accessed on your site, the most common user-agent strings, or the most common response codes returned.`,
	Args: cobra.ExactArgs(2),
	Run:  runFacetQuery,
}

var (
	facetQueryStartTime string
	facetQueryEndTime   string
	facetQueryCount     int
	facetQueryOutput    string
)

func init() {
	facetQueryCmd.Flags().StringVar(&facetQueryStartTime, "start", "", "Start time for the query (required)")
	facetQueryCmd.Flags().StringVar(&facetQueryEndTime, "end", "", "End time for the query")
	facetQueryCmd.Flags().IntVar(&facetQueryCount, "count", 100, "Number of distinct values to return (1-1000)")
	facetQueryCmd.Flags().StringVar(&facetQueryOutput, "output", "csv", "Output format: csv|json|json-pretty")
	facetQueryCmd.MarkFlagRequired("start")
}

func runFacetQuery(cmd *cobra.Command, args []string) {
	filter := args[0]
	field := args[1]

	c := getConfig().GetClient()

	params := client.FacetQueryParams{
		Filter:    filter,
		Field:     field,
		StartTime: facetQueryStartTime,
		EndTime:   facetQueryEndTime,
		Count:     facetQueryCount,
		Priority:  getConfig().Priority,
	}

	ctx := context.Background()
	result, err := c.FacetQuery(ctx, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch facetQueryOutput {
	case "json":
		outputJSON(result, false)
	case "json-pretty":
		outputJSON(result, true)
	default:
		outputFacetCSV(result.Values)
	}
}

func outputFacetCSV(values []client.FacetValue) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	writer.Write([]string{"count", "value"})

	for _, val := range values {
		record := []string{
			strconv.Itoa(val.Count),
			val.Value,
		}
		writer.Write(record)
	}
}

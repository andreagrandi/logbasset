package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "0.1.5"

var rootCmd = &cobra.Command{
	Use:   "logbasset",
	Short: "Command-line tool for accessing Scalyr services",
	Long: `LogBasset is a command-line tool for accessing Scalyr services.
The following commands are currently supported:

- query: Retrieve log data
- power-query: Execute PowerQuery
- numeric-query: Retrieve numeric / graph data
- facet-query: Retrieve common values for a field
- timeseries-query: Retrieve numeric / graph data from a timeseries
- tail: Provide a live 'tail' of a log`,
	Version: Version,
}

var (
	token    string
	server   string
	verbose  bool
	priority string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "API token (can also use scalyr_readlog_token env var)")
	rootCmd.PersistentFlags().StringVar(&server, "server", "", "Scalyr server URL (can also use scalyr_server env var)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&priority, "priority", "high", "Query priority (high|low)")

	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(powerQueryCmd)
	rootCmd.AddCommand(numericQueryCmd)
	rootCmd.AddCommand(facetQueryCmd)
	rootCmd.AddCommand(timeseriesQueryCmd)
	rootCmd.AddCommand(tailCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func checkTokenAndExit() {
	if token == "" {
		token = os.Getenv("scalyr_readlog_token")
		if token == "" {
			fmt.Fprintf(os.Stderr, "Error: API token is required. Set scalyr_readlog_token environment variable or use --token flag.\n")
			fmt.Fprintf(os.Stderr, "You can find API tokens at https://www.scalyr.com/keys\n")
			os.Exit(1)
		}
	}
}

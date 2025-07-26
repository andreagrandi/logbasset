package cli

import (
	"github.com/andreagrandi/logbasset/internal/app"
	"github.com/andreagrandi/logbasset/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfg          *config.Config
	flagToken    string
	flagServer   string
	flagVerbose  bool
	flagPriority string
)

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
	Version: app.Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.New()
		if err != nil {
			return err
		}

		cfg.SetFromFlags(flagToken, flagServer, flagVerbose, flagPriority)

		return cfg.Validate()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "API token (can also use scalyr_readlog_token env var)")
	rootCmd.PersistentFlags().StringVar(&flagServer, "server", "", "Scalyr server URL (can also use scalyr_server env var)")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&flagPriority, "priority", "high", "Query priority (high|low)")

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

func getConfig() *config.Config {
	return cfg
}

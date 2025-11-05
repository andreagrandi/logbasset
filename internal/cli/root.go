package cli

import (
	"time"

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
	flagLogLevel string
	flagTimeout  time.Duration
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
		// Skip authentication for commands that don't need API access
		// Check both the command itself and its parent (for completion subcommands like "bash", "zsh", etc.)
		if cmd.Name() == "completion" || cmd.Name() == "help" ||
			(cmd.Parent() != nil && cmd.Parent().Name() == "completion") {
			return nil
		}

		var err error
		cfg, err = config.NewWithoutValidation()
		if err != nil {
			return err
		}

		cfg.SetFromFlags(flagToken, flagServer, flagVerbose, flagPriority, flagLogLevel)

		if err := cfg.ApplyLogging(); err != nil {
			return err
		}

		return cfg.Validate()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "API token (can also use scalyr_readlog_token env var)")
	rootCmd.PersistentFlags().StringVar(&flagServer, "server", "", "Scalyr server URL (can also use scalyr_server env var)")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&flagPriority, "priority", "high", "Query priority (high|low)")
	rootCmd.PersistentFlags().StringVar(&flagLogLevel, "log-level", "info", "Log level (debug|info|warn|error)")
	rootCmd.PersistentFlags().DurationVar(&flagTimeout, "timeout", 30*time.Second, "Request timeout (e.g., 30s, 2m, 1h)")

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

func getTimeout() time.Duration {
	return flagTimeout
}

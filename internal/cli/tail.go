package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/andreagrandi/logbasset/internal/client"
	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/andreagrandi/logbasset/internal/validation"
	"github.com/spf13/cobra"
)

var tailCmd = &cobra.Command{
	Use:   "tail [filter]",
	Short: "Provide a live 'tail' of a log",
	Long: `Tail is similar to the query command, except it runs continually, printing query results to stdout.
It provides a live tail of log records matching the specified filter.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runTail,
}

var (
	tailLines  int
	tailOutput string
)

func init() {
	tailCmd.Flags().IntVarP(&tailLines, "lines", "n", 10, "Output the previous K lines when starting the tail")
	tailCmd.Flags().StringVar(&tailOutput, "output", "messageonly", "Output format: multiline|singleline|messageonly")
}

func runTail(cmd *cobra.Command, args []string) {
	var filter string
	if len(args) > 0 {
		filter = args[0]
	}

	// Validate inputs
	validationConfig := validation.DefaultConfig()
	params := validation.QueryValidationParams{
		Output:        tailOutput,
		Priority:      getConfig().Priority,
		Query:         filter,
		Lines:         tailLines,
		ValidateLines: true,
	}

	if err := validation.ValidateQueryParams(params, validationConfig); err != nil {
		errors.HandleErrorAndExit(err)
	}

	c := getConfig().GetClient()

	clientParams := client.TailParams{
		Filter:   filter,
		Lines:    tailLines,
		Priority: getConfig().Priority,
	}

	// Create context that can be cancelled by user signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	eventChan := make(chan client.LogEvent)

	go func() {
		if err := c.Tail(ctx, clientParams, eventChan); err != nil {
			// Only report error if it's not due to cancellation
			if ctx.Err() == nil {
				errors.HandleErrorAndExit(err)
			}
		}
	}()

	for event := range eventChan {
		switch tailOutput {
		case "multiline":
			outputTailMultiLine(event)
		case "singleline":
			outputTailSingleLine(event)
		default:
			outputTailMessageOnly(event)
		}
	}
}

func outputTailMessageOnly(event client.LogEvent) {
	fmt.Println(event.Message)
}

func outputTailSingleLine(event client.LogEvent) {
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

func outputTailMultiLine(event client.LogEvent) {
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
	fmt.Println("---")
}

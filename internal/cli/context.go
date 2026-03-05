package cli

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed context_embed.md
var contextContent string

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Print agent context document for AI/LLM consumption",
	Long:  "Prints a structured reference document describing LogBasset's commands, flags, output formats, and usage patterns. Useful for AI agents and automation tools.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(contextContent)
	},
}

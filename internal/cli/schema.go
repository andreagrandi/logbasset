package cli

import (
	"encoding/json"
	"fmt"

	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/spf13/cobra"
)

var schemaPretty bool

var schemaCmd = &cobra.Command{
	Use:   "schema [command]",
	Short: "Print JSON schema for command inputs and outputs",
	Long:  "Prints JSON schema describing a command's parameters, types, defaults, and valid values. With no arguments, lists all commands.",
	Args:  cobra.MaximumNArgs(1),
	Run:   runSchema,
}

func init() {
	schemaCmd.Flags().BoolVar(&schemaPretty, "pretty", false, "Pretty-print JSON output")
}

type commandSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type paramSchema struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Description string      `json:"description"`
}

type commandSchema struct {
	Command    string        `json:"command"`
	Args       []paramSchema `json:"args,omitempty"`
	Flags      []paramSchema `json:"flags"`
	OutputKeys []string      `json:"output_keys,omitempty"`
}

func runSchema(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		printSchemaJSON(getCommandList())
		return
	}

	schema, ok := schemas[args[0]]
	if !ok {
		errors.HandleErrorAndExit(errors.NewUsageError(
			fmt.Sprintf("unknown command: %s", args[0]),
			fmt.Errorf("valid commands: query, power-query, numeric-query, facet-query, timeseries-query, tail"),
		))
	}

	printSchemaJSON(schema)
}

func printSchemaJSON(data interface{}) {
	var output []byte
	var err error

	if schemaPretty {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		errors.HandleErrorAndExit(errors.NewParseError("failed to marshal schema JSON", err))
	}

	fmt.Println(string(output))
}

func getCommandList() []commandSummary {
	return []commandSummary{
		{Name: "query", Description: "Retrieve log data"},
		{Name: "power-query", Description: "Execute PowerQuery"},
		{Name: "numeric-query", Description: "Retrieve numeric/graph data"},
		{Name: "facet-query", Description: "Retrieve common values for a field"},
		{Name: "timeseries-query", Description: "Retrieve timeseries data"},
		{Name: "tail", Description: "Provide a live tail of a log"},
	}
}

var schemas = map[string]commandSchema{
	"query": {
		Command: "query",
		Args: []paramSchema{
			{Name: "filter", Type: "string", Required: false, Description: "Log filter expression"},
		},
		Flags: []paramSchema{
			{Name: "start", Type: "string", Required: false, Description: "Start time (e.g., 1h, 24h, 7d, 2024-01-15)"},
			{Name: "end", Type: "string", Required: false, Description: "End time (e.g., NOW, 1h, 2024-01-15 14:30)"},
			{Name: "count", Type: "integer", Required: false, Default: 10, Description: "Number of log records (1-5000)"},
			{Name: "mode", Type: "string", Required: false, Enum: []string{"head", "tail"}, Description: "Display mode"},
			{Name: "columns", Type: "string", Required: false, Description: "Comma-separated list of columns"},
			{Name: "output", Type: "string", Required: false, Default: "multiline", Enum: []string{"multiline", "singleline", "csv", "json", "json-pretty", "messageonly"}, Description: "Output format"},
			{Name: "fields", Type: "string", Required: false, Description: "Comma-separated fields to include in JSON output (e.g., timestamp,message,severity)"},
		},
		OutputKeys: []string{"timestamp", "severity", "message", "thread", "attributes"},
	},
	"power-query": {
		Command: "power-query",
		Args: []paramSchema{
			{Name: "query", Type: "string", Required: true, Description: "PowerQuery expression"},
		},
		Flags: []paramSchema{
			{Name: "start", Type: "string", Required: true, Description: "Start time (required)"},
			{Name: "end", Type: "string", Required: false, Description: "End time"},
			{Name: "output", Type: "string", Required: false, Default: "csv", Enum: []string{"csv", "json", "json-pretty"}, Description: "Output format"},
		},
	},
	"numeric-query": {
		Command: "numeric-query",
		Args: []paramSchema{
			{Name: "filter", Type: "string", Required: false, Description: "Log filter expression"},
		},
		Flags: []paramSchema{
			{Name: "function", Type: "string", Required: false, Description: "Function to compute (e.g., count, mean, min, max)"},
			{Name: "start", Type: "string", Required: true, Description: "Start time (required)"},
			{Name: "end", Type: "string", Required: false, Description: "End time"},
			{Name: "buckets", Type: "integer", Required: false, Default: 1, Description: "Number of time buckets (1-5000)"},
			{Name: "output", Type: "string", Required: false, Default: "csv", Enum: []string{"csv", "json", "json-pretty"}, Description: "Output format"},
		},
		OutputKeys: []string{"values"},
	},
	"facet-query": {
		Command: "facet-query",
		Args: []paramSchema{
			{Name: "filter", Type: "string", Required: true, Description: "Log filter expression"},
			{Name: "field", Type: "string", Required: true, Description: "Field name to facet on"},
		},
		Flags: []paramSchema{
			{Name: "start", Type: "string", Required: true, Description: "Start time (required)"},
			{Name: "end", Type: "string", Required: false, Description: "End time"},
			{Name: "count", Type: "integer", Required: false, Default: 100, Description: "Number of distinct values (1-1000)"},
			{Name: "output", Type: "string", Required: false, Default: "csv", Enum: []string{"csv", "json", "json-pretty"}, Description: "Output format"},
		},
		OutputKeys: []string{"value", "count"},
	},
	"timeseries-query": {
		Command: "timeseries-query",
		Args: []paramSchema{
			{Name: "filter", Type: "string", Required: false, Description: "Log filter expression"},
		},
		Flags: []paramSchema{
			{Name: "function", Type: "string", Required: false, Description: "Function to compute"},
			{Name: "start", Type: "string", Required: true, Description: "Start time (required)"},
			{Name: "end", Type: "string", Required: false, Description: "End time"},
			{Name: "buckets", Type: "integer", Required: false, Default: 1, Description: "Number of time buckets (1-5000)"},
			{Name: "output", Type: "string", Required: false, Default: "csv", Enum: []string{"csv", "json", "json-pretty"}, Description: "Output format"},
			{Name: "only-use-summaries", Type: "boolean", Required: false, Default: false, Description: "Only query summaries"},
			{Name: "no-create-summaries", Type: "boolean", Required: false, Default: false, Description: "Don't create summaries"},
		},
		OutputKeys: []string{"values"},
	},
	"tail": {
		Command: "tail",
		Args: []paramSchema{
			{Name: "filter", Type: "string", Required: false, Description: "Log filter expression"},
		},
		Flags: []paramSchema{
			{Name: "lines", Type: "integer", Required: false, Default: 10, Description: "Number of previous lines to show (-n)"},
			{Name: "output", Type: "string", Required: false, Default: "messageonly", Enum: []string{"messageonly", "multiline", "singleline", "json"}, Description: "Output format"},
		},
		OutputKeys: []string{"timestamp", "severity", "message", "thread", "attributes"},
	},
}

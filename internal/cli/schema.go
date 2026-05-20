package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/spf13/cobra"
)

var schemaPretty bool

var schemaCmd = &cobra.Command{
	Use:   "schema [command]",
	Short: "Print JSON schema for command inputs and outputs",
	Long:  "Prints JSON schema describing a command's parameters, types, defaults, valid values, and examples. With no arguments, lists all commands. Pass 'global' for the flags shared by every command.",
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
	ReadOnly   bool          `json:"read_only"`
	Args       []paramSchema `json:"args,omitempty"`
	Flags      []paramSchema `json:"flags"`
	OutputKeys []string      `json:"output_keys,omitempty"`
	Examples   []string      `json:"examples,omitempty"`
}

func runSchema(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		printSchemaJSON(getCommandList())
		return
	}

	schema, ok := schemas[args[0]]
	if !ok {
		var names []string
		for _, c := range getCommandList() {
			names = append(names, c.Name)
		}
		errors.HandleErrorAndExit(errors.NewUsageError(
			fmt.Sprintf("unknown command: %s", args[0]),
			fmt.Errorf("valid commands: %s", strings.Join(names, ", ")),
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
		{Name: "global", Description: "Flags shared by every command (schema target only, not a runnable command)"},
	}
}

func globalFlags() []paramSchema {
	return []paramSchema{
		{Name: "token", Type: "string", Required: false, Description: "API token (or scalyr_readlog_token env var)"},
		{Name: "server", Type: "string", Required: false, Default: "https://www.scalyr.com", Description: "Scalyr server URL (or scalyr_server env var)"},
		{Name: "verbose", Type: "boolean", Required: false, Default: false, Description: "Enable verbose output"},
		{Name: "priority", Type: "string", Required: false, Default: "high", Enum: []string{"high", "low"}, Description: "Query priority; use 'low' for heavy or background queries"},
		{Name: "log-level", Type: "string", Required: false, Default: "info", Enum: []string{"debug", "info", "warn", "error"}, Description: "Log level"},
		{Name: "timeout", Type: "string", Required: false, Default: "30s", Description: "Request timeout (e.g., 30s, 2m)"},
		{Name: "error-format", Type: "string", Required: false, Default: "text", Enum: []string{"text", "json"}, Description: "Error output format"},
		{Name: "pager", Type: "boolean", Required: false, Default: false, Description: "Pipe output through $PAGER (default 'less -RF') when stdout is a terminal"},
	}
}

var schemas = map[string]commandSchema{
	"query": {
		Command:  "query",
		ReadOnly: true,
		Args: []paramSchema{
			{Name: "filter", Type: "string", Required: false, Description: "Log filter expression"},
		},
		Flags: []paramSchema{
			{Name: "start", Type: "string", Required: false, Description: "Start time (e.g., 1h, 24h, 7d, 2024-01-15)"},
			{Name: "end", Type: "string", Required: false, Description: "End time (e.g., NOW, 1h, 2024-01-15 14:30)"},
			{Name: "count", Type: "integer", Required: false, Default: 10, Description: "Number of log records (1-5000)"},
			{Name: "mode", Type: "string", Required: false, Enum: []string{"head", "tail"}, Description: "Display mode"},
			{Name: "columns", Type: "string", Required: false, Description: "Comma-separated list of columns"},
			{Name: "output", Type: "string", Required: false, Default: "multiline", Enum: []string{"multiline", "singleline", "compact", "csv", "json", "json-pretty", "messageonly"}, Description: "Output format"},
			{Name: "fields", Type: "string", Required: false, Description: "Comma-separated fields to include in JSON output (e.g., timestamp,message,severity)"},
		},
		OutputKeys: []string{"timestamp", "severity", "message", "thread", "attributes"},
		Examples: []string{
			"logbasset query 'severity=\"error\"' --start 1h --count 100 --output json",
			"logbasset query '\"req-abc123\"' --start 24h --end NOW --output json --fields timestamp,message",
		},
	},
	"power-query": {
		Command:  "power-query",
		ReadOnly: true,
		Args: []paramSchema{
			{Name: "query", Type: "string", Required: true, Description: "PowerQuery expression"},
		},
		Flags: []paramSchema{
			{Name: "start", Type: "string", Required: true, Description: "Start time (required)"},
			{Name: "end", Type: "string", Required: false, Description: "End time"},
			{Name: "output", Type: "string", Required: false, Default: "csv", Enum: []string{"csv", "json", "json-pretty"}, Description: "Output format"},
		},
		Examples: []string{
			"logbasset power-query 'severity=\"error\" | group count by serverHost' --start 1h --output json",
		},
	},
	"numeric-query": {
		Command:  "numeric-query",
		ReadOnly: true,
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
		Examples: []string{
			"logbasset numeric-query 'severity=\"error\"' --function count --start 24h --buckets 24 --output json",
		},
	},
	"facet-query": {
		Command:  "facet-query",
		ReadOnly: true,
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
		Examples: []string{
			"logbasset facet-query '*' uriPath --start 24h --count 20 --output json",
		},
	},
	"timeseries-query": {
		Command:  "timeseries-query",
		ReadOnly: true,
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
		Examples: []string{
			"logbasset timeseries-query 'severity=\"error\"' --function count --start 24h --buckets 24 --output json",
		},
	},
	"tail": {
		Command:  "tail",
		ReadOnly: true,
		Args: []paramSchema{
			{Name: "filter", Type: "string", Required: false, Description: "Log filter expression"},
		},
		Flags: []paramSchema{
			{Name: "lines", Type: "integer", Required: false, Default: 10, Description: "Number of previous lines to show (-n)"},
			{Name: "output", Type: "string", Required: false, Default: "messageonly", Enum: []string{"messageonly", "multiline", "singleline", "compact", "json"}, Description: "Output format"},
		},
		OutputKeys: []string{"timestamp", "severity", "message", "thread", "attributes"},
		Examples: []string{
			"logbasset tail 'severity=\"error\"' --lines 50 --output json",
		},
	},
	"global": {
		Command:  "global",
		ReadOnly: true,
		Flags:    globalFlags(),
	},
}

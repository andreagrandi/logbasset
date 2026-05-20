package cli

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextEmbedMatchesContextMd(t *testing.T) {
	onDisk, err := os.ReadFile("../../CONTEXT.md")
	require.NoError(t, err)

	assert.Equal(t, string(onDisk), contextContent,
		"CONTEXT.md and internal/cli/context_embed.md must be byte-identical")
}

func TestContextCommandOutput(t *testing.T) {
	out := captureStdout(t, func() {
		contextCmd.Run(contextCmd, nil)
	})

	assert.Equal(t, contextContent, out)

	requiredSections := []string{
		"## Commands",
		"## Global Flags",
		"## Safety and Cost Guidance",
		"## Common Workflows",
		"## Output Formats",
		"## Exit Codes",
		"## Examples",
	}
	for _, section := range requiredSections {
		assert.Contains(t, out, section, "agent context is missing section %q", section)
	}
}

func TestSchemaCommandListMatchesSchemas(t *testing.T) {
	listed := make(map[string]bool)
	for _, c := range getCommandList() {
		assert.NotEmpty(t, c.Description, "command %q has no description", c.Name)
		_, ok := schemas[c.Name]
		assert.True(t, ok, "command %q is listed but has no schema entry", c.Name)
		listed[c.Name] = true
	}

	for name := range schemas {
		assert.True(t, listed[name], "schema %q exists but is not in the command list", name)
	}
}

func TestSchemaOutputIsValidJSON(t *testing.T) {
	listOut := captureStdout(t, func() {
		runSchema(schemaCmd, nil)
	})
	assert.True(t, json.Valid([]byte(listOut)), "schema command list is not valid JSON")

	var summaries []commandSummary
	require.NoError(t, json.Unmarshal([]byte(listOut), &summaries))
	assert.Len(t, summaries, len(getCommandList()))

	for name := range schemas {
		out := captureStdout(t, func() {
			runSchema(schemaCmd, []string{name})
		})
		assert.True(t, json.Valid([]byte(out)), "schema %q is not valid JSON", name)

		var schema commandSchema
		require.NoError(t, json.Unmarshal([]byte(out), &schema), "schema %q failed to unmarshal", name)
		assert.Equal(t, name, schema.Command)
		assert.True(t, schema.ReadOnly, "schema %q must report read_only", name)
	}
}

func TestRunnableCommandSchemasHaveExamples(t *testing.T) {
	for name, schema := range schemas {
		if name == "global" {
			continue
		}
		assert.NotEmpty(t, schema.Examples, "command %q schema has no examples", name)
	}
}

// TestSchemaFlagsMatchCobraDefinitions guards against schema drift: every flag a
// command actually accepts must be described by `schema`, and vice versa.
func TestSchemaFlagsMatchCobraDefinitions(t *testing.T) {
	runnable := map[string]*cobra.Command{
		"query":            queryCmd,
		"power-query":      powerQueryCmd,
		"numeric-query":    numericQueryCmd,
		"facet-query":      facetQueryCmd,
		"timeseries-query": timeseriesQueryCmd,
		"tail":             tailCmd,
	}

	for name, cmd := range runnable {
		schema, ok := schemas[name]
		require.True(t, ok, "no schema for command %q", name)

		schemaFlags := make(map[string]bool)
		for _, f := range schema.Flags {
			schemaFlags[f.Name] = true
		}

		cobraFlags := make(map[string]bool)
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if f.Name == "help" {
				return
			}
			cobraFlags[f.Name] = true
		})

		for flag := range cobraFlags {
			assert.True(t, schemaFlags[flag],
				"command %q defines flag %q but `schema %s` does not describe it", name, flag, name)
		}
		for flag := range schemaFlags {
			assert.True(t, cobraFlags[flag],
				"`schema %s` describes flag %q that command %q does not define", name, flag, name)
		}
	}
}

func TestGlobalSchemaMatchesPersistentFlags(t *testing.T) {
	schema, ok := schemas["global"]
	require.True(t, ok, "missing 'global' schema entry")

	schemaFlags := make(map[string]bool)
	for _, f := range schema.Flags {
		schemaFlags[f.Name] = true
	}

	cobraFlags := make(map[string]bool)
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		cobraFlags[f.Name] = true
	})

	for flag := range cobraFlags {
		assert.True(t, schemaFlags[flag],
			"persistent flag %q is not described by `schema global`", flag)
	}
	for flag := range schemaFlags {
		assert.True(t, cobraFlags[flag],
			"`schema global` describes flag %q that is not a persistent flag", flag)
	}
}

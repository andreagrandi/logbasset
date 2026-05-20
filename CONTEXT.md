# LogBasset - Agent Context

LogBasset is a **read-only** CLI for querying Scalyr/DataSet logs. All commands are queries; nothing is mutated.

## Authentication

Set the API token via:
- Environment variable: `scalyr_readlog_token`
- Config file key: `token`
- Flag: `--token`

Server URL (default `https://www.scalyr.com`):
- Environment variable: `scalyr_server`
- Config file key: `server`
- Flag: `--server`

## Commands

| Command | Description | Required args | Required flags |
|---------|-------------|---------------|----------------|
| `query [filter]` | Retrieve log data | none (filter optional) | none |
| `power-query <query>` | Execute PowerQuery | query (positional) | `--start` |
| `numeric-query [filter]` | Retrieve numeric/graph data | none (filter optional) | `--start` |
| `facet-query <filter> <field>` | Get common values for a field | filter, field (positional) | `--start` |
| `timeseries-query [filter]` | Retrieve timeseries data | none (filter optional) | `--start` |
| `tail [filter]` | Live tail of logs | none (filter optional) | none |
| `context` | Print this agent context document | none | none |
| `schema [command]` | Print JSON schema for a command (pass `global` for shared flags) | none | none |

## Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--token` | string | (env) | API token |
| `--server` | string | `https://www.scalyr.com` | Server URL |
| `--verbose` | bool | false | Enable verbose output |
| `--priority` | string | `high` | Query priority: `high` or `low` |
| `--log-level` | string | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `--timeout` | duration | `30s` | Request timeout (e.g., `30s`, `2m`) |
| `--error-format` | string | `text` | Error output format: `text` or `json` |
| `--pager` | bool | false | Pipe output through `$PAGER` (default `less -RF`) when stdout is a terminal |

## Safety and Cost Guidance

Every command is read-only — queries never create, modify, or delete data, so
they are safe to run without confirmation. The `read_only` field in `schema`
output reports this for each command.

To keep queries fast and inexpensive:
- Prefer narrow time ranges (`--start 1h` over `--start 7d`); only add
  `--end NOW` when you need data right up to the current time.
- Cap raw-record fetches with `--count`; the `query` default is 10.
- Use `--priority low` for heavy or background queries so interactive queries
  stay responsive. The default `high` is best for small, time-sensitive lookups.
- Aggregations (`power-query`, `numeric-query`, `facet-query`,
  `timeseries-query`) summarize wide ranges far more cheaply than fetching many
  raw records with `query`.
- Raise `--timeout` for queries over wide ranges; the default is 30s.

## Flags That Do NOT Exist

Agents commonly hallucinate these flags. They will cause errors:
- `--env` — does not exist
- `--minutes` — use `--start "30m"` instead
- `--query` — pass query as positional argument
- `--format` — use `--output` instead
- `--limit` — use `--count` instead
- `--from` / `--to` — use `--start` / `--end` instead

## Time Format Reference

- Relative: `30m`, `1h`, `24h`, `7d` (minutes, hours, days)
- Absolute: `2024-01-15`, `2024-01-15 14:30:05`
- Time-only: `14:30`, `2:30 PM`
- Special: `NOW` (for `--end` to get results up to current time)

When using `--start` without `--end`, the API returns only 24h from start. Use `--end NOW` to get results up to the current time.

## Output Formats

### query command
`--output`: `multiline` (default in TTY), `singleline`, `compact`, `csv`, `json` (default in pipe), `json-pretty`, `messageonly`

`compact` prints one event per line as `HH:MM:SS <severity-char> <message>` where the severity char is `D` (≤2), `I` (3), `W` (4), `E` (5), or `F` (≥6). Designed for scanning large result sets.

### power-query, numeric-query, facet-query, timeseries-query
`--output`: `csv` (default in TTY), `json` (default in pipe), `json-pretty`

### tail command
`--output`: `messageonly` (default in TTY), `multiline`, `singleline`, `compact`, `json` (default in pipe)

Use `--fields` with `query --output json` to select specific fields and reduce output size.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General/API/parse error |
| 2 | Usage error (bad command syntax) |
| 3 | Network error |
| 4 | Authentication error (bad/missing token) |
| 5 | Configuration error |
| 6 | Validation error (bad input) |

## Structured Error Output

Use `--error-format json` to get machine-readable errors on stderr:
```json
{"error":{"type":"AUTH_ERROR","message":"API token is required","suggestion":"...","exit_code":4}}
```

## TTY Auto-Detection

When stdout is not a TTY (piped to another program), the default output format automatically switches to `json`. This can be overridden with an explicit `--output` flag.

## Common Workflows

### Investigate an error spike
1. Find when errors occurred, bucketed by hour:
   `logbasset numeric-query 'severity="error"' --function count --start 24h --buckets 24 --output json`
2. Read sample errors from the affected window:
   `logbasset query 'severity="error"' --start 2h --count 50 --output json`
3. Group errors by host to localize the cause:
   `logbasset power-query 'severity="error" | group count by serverHost' --start 2h --output json`

### Trace a request or correlation ID
Pass the identifier as a quoted text filter:
`logbasset query '"req-abc123"' --start 24h --end NOW --output json`

### Rank the most common values of a field
`logbasset facet-query '*' uriPath --start 24h --count 20 --output json`

## Examples

```bash
# Search logs for errors in the last hour
logbasset query 'severity="error"' --start "1h" --count 100 --output json

# Search with text filter, get JSON output with specific fields
logbasset query '"service timeout"' --start "24h" --output json --fields timestamp,message,severity

# PowerQuery: count errors by server
logbasset power-query 'serverHost = * | group count by serverHost' --start "1h" --output json

# Facet query: top URLs
logbasset facet-query '*' 'url' --start "24h" --count 50 --output json

# Numeric query: error rate over 24h in hourly buckets
logbasset numeric-query 'severity="error"' --function 'count' --start "24h" --buckets 24 --output json

# Tail with JSON for agent consumption
logbasset tail 'severity="error"' --output json --lines 50

# Get structured errors for programmatic handling
logbasset query '"test"' --error-format json 2>/tmp/err.json
```

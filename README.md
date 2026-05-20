# Log Basset

<img src="logbasset.png" width="50%" alt="LogBasset">

Command-line tool for accessing Scalyr services written in Go. This tool provides a fast, efficient way to query and analyze your logs from the command line.

This is a Go implementation inspired by the original [scalyr-tool](https://github.com/scalyr/scalyr-tool), offering improved performance and cross-platform compatibility.

## Features

The following commands are currently supported:

- **query**: Retrieve log data
- **power-query**: Execute PowerQuery
- **numeric-query**: Retrieve numeric / graph data
- **facet-query**: Retrieve common values for a field
- **timeseries-query**: Retrieve numeric / graph data from a timeseries
- **tail**: Provide a live 'tail' of a log

LogBasset includes comprehensive input validation that checks parameters before making API calls, ensuring you get immediate feedback for invalid time formats, counts, or other parameters.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap andreagrandi/tap
brew install logbasset
```

### From Source

```bash
git clone https://github.com/andreagrandi/logbasset
cd logbasset
make build
```

The binary will be created in the `bin/` directory.

### Binary Release

Download the latest binary from the [releases page](https://github.com/andreagrandi/logbasset/releases).

## Configuration

You need to make your Scalyr API token available to the tool. LogBasset supports multiple configuration methods:

### Environment Variables

The easiest way is to set the environment variable:

```bash
export scalyr_readlog_token='your-api-token-here'
```

For custom servers:

```bash
export scalyr_server='https://eu.scalyr.com'
```

### Configuration Files

LogBasset also supports configuration files in YAML format. The tool will look for configuration files in these locations (in order):

1. `./logbasset.yaml` (current directory)
2. `~/.logbasset/logbasset.yaml` (user home directory)
3. `~/.config/logbasset/logbasset.yaml` (XDG config directory)

Example configuration file:

```yaml
token: your-api-token-here
server: https://eu.scalyr.com
verbose: false
priority: high
log_level: info
```

### Command Line Flags

You can also specify configuration values using command line flags:

```bash
logbasset --token=your-token --server=https://eu.scalyr.com query
```

### Configuration Priority

Configuration values are applied in the following order (highest to lowest priority):

1. Command line flags
2. Environment variables
3. Configuration file values
4. Default values

You can find your API tokens at [scalyr.com/keys](https://www.scalyr.com/keys) -- look for "Read Logs" token.

## Time Format Support

LogBasset supports flexible time format specifications for start and end times:

### Relative Time Formats
- `24h` - 24 hours ago
- `7d` - 7 days ago
- `30m` - 30 minutes ago  
- `60s` - 60 seconds ago

### Absolute Time Formats
- `2024-01-15` - Specific date
- `2024-01-15 14:30` - Date and time
- `2024-01-15 14:30:45` - Date, time with seconds

### Time-Only Formats
- `14:30` - Time in 24-hour format
- `14:30:45` - Time with seconds
- `2:30 PM` - Time in 12-hour format
- `2:30:45 PM` - Time with seconds in 12-hour format

## Usage
### Query Logs

Retrieve and search your log data:

```bash
# Display the last 10 log records
logbasset query

# Display the last 100 log records, showing only timestamp, severity, and message
logbasset query --count=100 --columns='timestamp,severity,message'

# Display the first 10 log records beginning at 3:00 PM today, from host100
logbasset query '$serverHost="host100"' --start='3:00 PM'

# Display the last 1000 entries in CSV format
logbasset query '$source="accessLog"' --output=csv --columns='status,uriPath' --count=1000
```

**Options:**
- `--start=xxx`: Beginning of the time range to query
- `--end=xxx`: End of the time range to query  
- `--count=nnn`: Number of log records to retrieve (1-5000), defaults to 10
- `--mode=head|tail`: Whether to display from start or end of time range
- `--columns="..."`: Which log attributes to display (comma-separated)
- `--output=multiline|singleline|compact|csv|json|json-pretty`: Output format
- `--priority=high|low`: Query execution priority

### Power Query

Execute PowerQueries for advanced log analysis:

```bash
# Display log volume summary by log file for the last 24 hours
logbasset power-query "tag='logVolume' metric='logBytes' | group sum(value) by forlogfile" --start="24h"

# Display requests, errors and error rate for the last 7 days in JSON
logbasset power-query "dataset = 'accesslog' | group requests = count(), errors = count(status == 404) by uriPath | let rate = errors / requests | filter rate > 0.01 | sort -rate" --start="7d" --end="0d" --output=json-pretty
```

**Options:**
- `--start=xxx`: Beginning of time range (required)
- `--end=xxx`: End of time range
- `--output=csv|json|json-pretty`: Output format (defaults to csv)
- `--priority=high|low`: Query execution priority

### Numeric Query

Retrieve numeric data for graphing and analysis:

```bash
# Count the rate of "/login" occurrences in each of the last 24 hours
logbasset numeric-query '"/login"' --start 24h --buckets 24

# Display average response size for all requests in the last hour
logbasset numeric-query '$dataset="accesslog"' --function 'bytes' --start 1h
```

**Options:**
- `--function=xxx`: Value to compute (mean, median, count, rate, etc.)
- `--start=xxx`: Beginning of time range (required)
- `--end=xxx`: End of time range
- `--buckets=nnn`: Number of time buckets (1-5000), defaults to 1
- `--output=csv|json|json-pretty`: Output format
- `--priority=high|low`: Query execution priority

### Facet Query

Find the most common values for a field:

```bash
# Display the most common HTTP request URLs in the last 24 hours
logbasset facet-query '$dataset="accesslog"' uriPath --start 24h

# Display the most common HTTP response codes for requests to index.html
logbasset facet-query 'uriPath="/index.html"' status --start 24h
```

**Options:**
- `--count=nnn`: Number of distinct values to return (1-1000), defaults to 100
- `--start=xxx`: Beginning of time range (required)
- `--end=xxx`: End of time range
- `--output=csv|json|json-pretty`: Output format
- `--priority=high|low`: Query execution priority

### Timeseries Query

Retrieve precomputed numeric data for fast dashboard updates:

```bash
# Fast retrieval of access log metrics using timeseries
logbasset timeseries-query '$dataset="accesslog"' --function 'bytes' --start 24h --buckets 24

# Only use existing summaries (fast, may return empty if not backfilled)
logbasset timeseries-query '$dataset="accesslog"' --function 'count' --start 7d --only-use-summaries
```

**Options:**
- `--function=xxx`: Value to compute (mean, median, count, rate, etc.)
- `--start=xxx`: Beginning of time range (required)
- `--end=xxx`: End of time range
- `--buckets=nnn`: Number of time buckets (1-5000), defaults to 1
- `--only-use-summaries`: Only query existing summaries
- `--no-create-summaries`: Don't create new summaries for this query
- `--output=csv|json|json-pretty`: Output format
- `--priority=high|low`: Query execution priority

### Tail Logs

Live tail of log records:

```bash
# Display a live tail of all log records
logbasset tail

# Display a live tail from a specific host
logbasset tail '$serverHost="host100"'

# Display live tail with full record details
logbasset tail --output multiline
```

**Options:**
- `--lines=K` or `-n K`: Output the previous K lines when starting (defaults to 10)
- `--output=multiline|singleline|compact|messageonly`: Output format (defaults to messageonly)
- `--priority=high|low`: Query execution priority

## Global Options

These options are available for all commands:

- `--token=xxx`: Specify the API token
- `--server=xxx`: Specify the Scalyr server URL
- `--verbose`: Enable verbose output for debugging
- `--priority=high|low`: Query execution priority (defaults to high)
- `--log-level=debug|info|warn|error`: Set logging level (defaults to info)
- `--pager`: Pipe output through `$PAGER` (defaults to `less -RF`) when stdout is a terminal

## Output Formats

### JSON Output
- `json`: Compact JSON output
- `json-pretty`: Pretty-printed JSON with indentation

### CSV Output
- Uses Excel CSV format with CRLF line separators
- Headers included when applicable
- Values properly escaped and quoted

### Text Output
- `multiline`: Verbose format with each attribute on separate lines
- `singleline`: Compact format with all attributes on one line
- `compact`: One line per event, `HH:MM:SS <severity> <message>` — designed for scanning large result sets
- `messageonly`: Only the log message (useful for tail)

In `compact` mode the severity column is a single letter: `D` (debug, severity ≤ 2), `I` (info, 3), `W` (warning, 4), `E` (error, 5), `F` (fatal, ≥ 6).

### Paging Large Output

Two ways to make large query results easier to scan:

```bash
# Built-in: pipe through $PAGER (defaults to less -RF) while keeping stderr on the terminal
logbasset query 'error' --count=5000 --output=compact --pager

# Manual: just pipe to your pager of choice
logbasset query 'error' --count=5000 --output=compact | less -R
```

`--pager` only activates when stdout is a terminal, so scripts and pipelines (including JSON/CSV output redirected to a file) are unaffected.

## Usage Limits

Your command line and API queries are limited to 30,000 milliseconds of server processing time, replenished at 36,000 milliseconds per hour. If you exceed this limit, your queries will be intermittently refused.

For the tail command, you're limited to a maximum of 1,000 log records per 10 seconds, and tails automatically expire after 10 minutes.

If you need higher limits, contact [support@scalyr.com](mailto:support@scalyr.com).

## Recipes

Copy-paste-friendly recipes for common log investigation workflows. Every
example uses current CLI flags -- adjust the filters, fields, and time ranges
to match your data.

### Filtering and searching

```bash
# All errors in the last hour
logbasset query 'severity >= 3' --start=1h

# Errors right up to the current moment
# (without --end, only the first 24h after --start is searched -- see Troubleshooting)
logbasset query 'severity >= 3' --start=24h --end=NOW

# Trace a request or correlation ID across all logs (quote the literal text)
logbasset query '"req-abc123"' --start=24h --end=NOW

# Restrict to a single host and log source
logbasset query '$serverHost="host100" $source="accessLog"' --start=1h

# Show only the columns you care about, newest entries first
logbasset query 'severity >= 3' --start=1h --count=100 --columns='timestamp,severity,message' --mode=tail

# Scan a large result set, one event per line, through a pager
logbasset query 'severity >= 3' --start=24h --end=NOW --count=5000 --output=compact --pager
```

### Breaking down values with facets

```bash
# Most common HTTP request URLs
logbasset facet-query '$source="accessLog"' uriPath --start=24h --count=20

# Most common response codes for a single endpoint
logbasset facet-query 'uriPath="/index.html"' status --start=24h

# Most common error messages
logbasset facet-query 'severity >= 3' message --start=24h --count=20
```

### Numeric queries for trends

```bash
# Error count per hour over the last day
logbasset numeric-query 'severity >= 3' --function=count --start=24h --buckets=24

# Request volume per minute for the last hour
logbasset numeric-query '$source="accessLog"' --function=count --start=1h --buckets=60

# Average response time in 5-minute buckets
logbasset numeric-query '$source="accessLog"' --function='mean(responseTime)' --start=1h --buckets=12
```

### Timeseries queries for repeated dashboards

`timeseries-query` takes the same shape as `numeric-query` but reads precomputed
summaries, so it is much faster for queries you run again and again.

```bash
# Error count per hour, backed by summaries
logbasset timeseries-query 'severity >= 3' --function=count --start=24h --buckets=24

# Fastest possible read: only existing summaries (may be empty until backfilled)
logbasset timeseries-query '$source="accessLog"' --function=count --start=7d --buckets=7 --only-use-summaries

# Read without creating new summaries (avoids extra background work)
logbasset timeseries-query '$source="accessLog"' --function=count --start=24h --buckets=24 --no-create-summaries
```

### Live tailing

```bash
# Tail every new error as it arrives
logbasset tail 'severity >= 3'

# Tail a single host, starting with the last 50 lines
logbasset tail '$serverHost="host100"' --lines=50

# Tail with one compact line per event
logbasset tail 'severity >= 3' --output=compact
```

Press `Ctrl+C` to stop a tail cleanly. Tails are capped at 1,000 records per
10 seconds and expire after 10 minutes (see [Usage Limits](#usage-limits)).

### Investigate an error spike end to end

```bash
# 1. Find which hour the spike happened in
logbasset numeric-query 'severity >= 3' --function=count --start=24h --buckets=24

# 2. Read a sample of errors from the affected window
logbasset query 'severity >= 3' --start=3h --end=NOW --count=50

# 3. Localize the cause by grouping errors per host
logbasset power-query 'severity >= 3 | group count = count() by serverHost | sort -count' --start=3h

# 4. Rank the error messages driving the spike
logbasset facet-query 'severity >= 3' message --start=3h --count=20
```

## Troubleshooting

### Authentication failures (exit code 4)

`API token is required` or `Invalid API token` means LogBasset could not
authenticate:

- Provide a token via the `scalyr_readlog_token` environment variable, the
  `--token` flag, or the `token:` key in a config file (see
  [Configuration](#configuration)).
- Use a **Read Logs** token from [scalyr.com/keys](https://www.scalyr.com/keys);
  write or admin tokens do not work for queries.
- Check the server region. EU accounts must point at `https://eu.scalyr.com`
  via `scalyr_server` or `--server` -- a token from one region fails against
  the other.

```bash
# Confirm the token and server in use, and see the raw API request
logbasset query --count=1 --verbose

# Get a machine-readable error instead of plain text
logbasset query --count=1 --error-format=json
```

### Empty results

A query that succeeds but returns nothing usually means the time range or
filter did not match -- not that something is broken:

- **Missing `--end`.** With only `--start`, the API searches just the first
  24 hours after that point. To search up to the current moment, add
  `--end=NOW`:

  ```bash
  # Searches 7d-ago .. 6d-ago only
  logbasset query 'severity >= 3' --start=7d

  # Searches 7d-ago .. now
  logbasset query 'severity >= 3' --start=7d --end=NOW
  ```

- **Filter too narrow, or wrong field name.** Drop the filter to confirm data
  exists, then add conditions back one at a time. Field names are
  case-sensitive.
- **Wrong server region.** An EU account queried against the default US server
  returns no data -- set `--server=https://eu.scalyr.com`.
- **`timeseries-query --only-use-summaries`** stays empty until summaries have
  been backfilled for the range; drop the flag to fall back to a live
  computation.

### Time ranges

- Relative times count backwards from now: `30m`, `1h`, `24h`, `7d`.
- Absolute times: `2024-01-15` or `2024-01-15 14:30:05`. Time-only values such
  as `14:30` or `2:30 PM` mean today.
- `NOW` is valid only for `--end`; it pins the end of the range to the current
  time.
- `power-query`, `numeric-query`, `facet-query`, and `timeseries-query` all
  require `--start`; omitting it is a validation error (exit code 6).

### Timeouts and rate limits

- `operation timed out` means the query exceeded `--timeout` (default `30s`).
  Raise it (`--timeout=2m`), narrow the time range, or fetch fewer records
  with `--count`.
- Aggregations (`power-query`, `numeric-query`, `facet-query`,
  `timeseries-query`) summarize wide ranges far more cheaply than pulling many
  raw records with `query`.
- Use `--priority=low` for heavy or background queries so interactive queries
  keep their share of the processing budget (see [Usage Limits](#usage-limits)).

### Inspecting what LogBasset is doing

```bash
# Show API request and response details
logbasset query 'severity >= 3' --start=1h --verbose --log-level=debug
```

Exit codes let scripts branch on the failure type: `0` success, `1` general or
API error, `2` usage error, `3` network error, `4` authentication error,
`5` configuration error, `6` validation error.

## Building

Requirements:
- Go 1.26.1 or later
- Make (optional, but recommended)

### Using Make (recommended)

```bash
# Build for current platform
make build

# Run tests
make test

# Build for multiple platforms
make build-all

# See all available targets
make help
```

### Using Go directly

```bash
# Build for current platform
go build -o bin/logbasset .

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o bin/logbasset-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -o bin/logbasset-darwin-amd64 .
GOOS=windows GOARCH=amd64 go build -o bin/logbasset-windows-amd64.exe .
```

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Disclaimer

A large part of this codebase has been produced with AI tools. If this doesn't match with your tastes, please
use some other tool 🙂

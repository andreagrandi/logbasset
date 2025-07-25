# LogBasset

![LogBasset](logbasset.png)

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

## Installation

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

You need to make your Scalyr API token available to the tool. The easiest way is to set the environment variable:

```bash
export scalyr_readlog_token='your-api-token-here'
```

Alternatively, you can specify the token on the command line using the `--token` argument.

You can find your API tokens at [scalyr.com/keys](https://www.scalyr.com/keys) -- look for "Read Logs" token.

### Custom Server

If you're using a custom Scalyr server, you can set it via environment variable:

```bash
export scalyr_server='https://eu.scalyr.com'
```

Or use the `--server` flag.

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
- `--output=multiline|singleline|csv|json|json-pretty`: Output format
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
- `--output=multiline|singleline|messageonly`: Output format (defaults to messageonly)
- `--priority=high|low`: Query execution priority

## Global Options

These options are available for all commands:

- `--token=xxx`: Specify the API token
- `--server=xxx`: Specify the Scalyr server URL
- `--verbose`: Enable verbose output for debugging
- `--priority=high|low`: Query execution priority (defaults to high)

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
- `messageonly`: Only the log message (useful for tail)

## Usage Limits

Your command line and API queries are limited to 30,000 milliseconds of server processing time, replenished at 36,000 milliseconds per hour. If you exceed this limit, your queries will be intermittently refused.

For the tail command, you're limited to a maximum of 1,000 log records per 10 seconds, and tails automatically expire after 10 minutes.

If you need higher limits, contact [support@scalyr.com](mailto:support@scalyr.com).

## Examples

### Basic Log Analysis

```bash
# Find all errors in the last hour
logbasset query 'severity >= 3' --start=1h

# Count requests per minute for the last hour
logbasset numeric-query '$dataset="accesslog"' --start=1h --buckets=60

# Find most common error messages
logbasset facet-query 'severity >= 3' message --start=24h --count=20
```

### Performance Monitoring

```bash
# Monitor response times
logbasset numeric-query '$dataset="accesslog"' --function='mean(responseTime)' --start=1h --buckets=12

# Find slowest endpoints
logbasset power-query 'dataset="accesslog" | group avg_time=mean(responseTime), count=count() by uriPath | sort -avg_time' --start=1h
```

### Live Monitoring

```bash
# Monitor errors in real-time
logbasset tail 'severity >= 3'

# Watch specific application logs
logbasset tail '$source="myapp"' --output=singleline
```

## Building

Requirements:
- Go 1.21 or later
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
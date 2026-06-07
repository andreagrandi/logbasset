---
name: logbasset
description: Query, search, aggregate, or live-tail Scalyr/DataSet logs from the command line with the logbasset CLI. Use when the user wants to investigate logs, find events by ID or field, count or aggregate log data over a time range, or live-tail a log.
---

# logbasset

`logbasset` is a **read-only** command-line tool for querying Scalyr/DataSet
logs. It can search raw log records, run aggregations (counts, groupings,
facets, time series), and live-tail a log. Nothing it does mutates data.

This skill does not restate the command and flag reference, because the binary
documents itself and stays current with the installed version. Load that
reference at runtime instead of guessing.

## Workflow

Follow these steps in order:

1. **Check the binary is installed.** Confirm `logbasset` is on `PATH` (e.g.
   `logbasset --version`). If it is not found, stop and tell the user to
   install it (`brew install andreagrandi/tap/logbasset`, or download from
   https://github.com/andreagrandi/logbasset/releases) — do not attempt to
   query without it.
2. **Load the authoritative reference.** Run `logbasset context` to print the
   current list of commands, flags, time formats, output formats, exit codes,
   and worked workflows for the installed version. Treat this output as the
   source of truth.
3. **Confirm exact flags before composing a non-trivial query.** Run
   `logbasset schema <command>` for a specific command's flags, enums, and
   defaults, or `logbasset schema global` for the flags shared by every
   command. Do not invent flags — `logbasset context` lists ones that look
   plausible but do not exist.
4. **Construct and run the query** using only flags confirmed from the steps
   above.

## Safety

Every `logbasset` command is read-only — queries never create, modify, or
delete data — so they are safe to run without asking the user for
confirmation.

## Keeping queries fast and inexpensive

- Prefer narrow time ranges (`--start 1h` over `--start 7d`).
- Cap raw-record fetches with `--count`.
- Prefer aggregation commands over fetching many raw records when the user
  wants counts, rates, or top values.
- Use `--priority low` for heavy or background queries so interactive ones stay
  responsive.

For the exact flag names, output formats, and time syntax, always rely on
`logbasset context` and `logbasset schema` rather than memory.

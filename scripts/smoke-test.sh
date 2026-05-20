#!/usr/bin/env bash
# Smoke-test a built logbasset binary or release artifact.
# Exercises only commands that need no API credentials so it can run in CI
# and against published archives. Exits non-zero on the first failed check.
#
# Usage: scripts/smoke-test.sh [path-to-binary]   (default: bin/logbasset)
set -euo pipefail

BINARY="${1:-bin/logbasset}"

if [ ! -x "$BINARY" ]; then
	echo "smoke-test: binary not found or not executable: $BINARY" >&2
	exit 1
fi

CORE_COMMANDS=(query power-query numeric-query facet-query timeseries-query tail)

fail() {
	echo "smoke-test: FAIL - $1" >&2
	exit 1
}

ok() {
	echo "smoke-test: ok - $1"
}

run() {
	local desc="$1"
	shift
	if "$@" >/dev/null 2>&1; then
		ok "$desc"
	else
		fail "$desc"
	fi
}

# --version prints a recognizable, non-empty version line.
version_output="$("$BINARY" --version 2>&1)" || fail "--version exited non-zero"
case "$version_output" in
	"logbasset version "*) ok "--version ($version_output)" ;;
	*) fail "--version output unexpected: $version_output" ;;
esac

# Top-level help lists every core command.
help_output="$("$BINARY" --help 2>&1)" || fail "--help exited non-zero"
for cmd in "${CORE_COMMANDS[@]}"; do
	case "$help_output" in
		*"$cmd"*) ;;
		*) fail "--help missing command: $cmd" ;;
	esac
done
ok "--help lists core commands"

# Per-command help renders without error.
for cmd in "${CORE_COMMANDS[@]}"; do
	run "$cmd --help" "$BINARY" "$cmd" --help
done

# Commands that need no API credentials work end to end.
run "context" "$BINARY" context
run "schema" "$BINARY" schema
run "completion bash" "$BINARY" completion bash

echo "smoke-test: all checks passed for $BINARY"

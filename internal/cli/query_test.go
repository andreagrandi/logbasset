package cli

import (
	"testing"

	"github.com/andreagrandi/logbasset/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestOutputCompact(t *testing.T) {
	events := []client.LogEvent{
		{Timestamp: "1700000000000000000", Severity: 3, Message: "service ready"},
		{Timestamp: "1700000001000000000", Severity: 5, Message: "boom"},
		{Timestamp: "not-a-time", Severity: 0, Message: "fallback"},
	}

	out := captureStdout(t, func() {
		outputCompact(events)
	})

	expected := "22:13:20 I service ready\n" +
		"22:13:21 E boom\n" +
		"not-a-time D fallback\n"
	assert.Equal(t, expected, out)
}

func TestOutputCompact_Empty(t *testing.T) {
	out := captureStdout(t, func() {
		outputCompact(nil)
	})
	assert.Empty(t, out)
}

func TestOutputTailCompact(t *testing.T) {
	event := client.LogEvent{
		Timestamp: "1700000000000000000",
		Severity:  4,
		Message:   "warning ahead",
	}

	out := captureStdout(t, func() {
		outputTailCompact(event)
	})

	assert.Equal(t, "22:13:20 W warning ahead\n", out)
}

package logging

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetLevel(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		expectError bool
		expectedLvl logrus.Level
	}{
		{"debug level", "debug", false, logrus.DebugLevel},
		{"info level", "info", false, logrus.InfoLevel},
		{"warn level", "warn", false, logrus.WarnLevel},
		{"error level", "error", false, logrus.ErrorLevel},
		{"uppercase level", "INFO", false, logrus.InfoLevel},
		{"invalid level", "invalid", true, logrus.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetLevel(tt.level)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLvl, logger.GetLevel())
			}
		})
	}
}

func TestLoggingFunctions(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel("debug")

	tests := []struct {
		name     string
		logFunc  func()
		contains string
	}{
		{
			name:     "debug logging",
			logFunc:  func() { Debug("debug message") },
			contains: "debug message",
		},
		{
			name:     "info logging",
			logFunc:  func() { Info("info message") },
			contains: "info message",
		},
		{
			name:     "warn logging",
			logFunc:  func() { Warn("warn message") },
			contains: "warn message",
		},
		{
			name:     "error logging",
			logFunc:  func() { Error("error message") },
			contains: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			output := buf.String()
			assert.Contains(t, output, tt.contains)
		})
	}
}

func TestFormattedLogging(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel("debug")

	tests := []struct {
		name     string
		logFunc  func()
		contains string
	}{
		{
			name:     "debugf logging",
			logFunc:  func() { Debugf("debug %s", "formatted") },
			contains: "debug formatted",
		},
		{
			name:     "infof logging",
			logFunc:  func() { Infof("info %d", 42) },
			contains: "info 42",
		},
		{
			name:     "warnf logging",
			logFunc:  func() { Warnf("warn %v", true) },
			contains: "warn true",
		},
		{
			name:     "errorf logging",
			logFunc:  func() { Errorf("error %s", "test") },
			contains: "error test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			output := buf.String()
			assert.Contains(t, output, tt.contains)
		})
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel("info")

	WithFields(logrus.Fields{
		"component": "test",
		"action":    "logging",
	}).Info("structured log message")

	output := buf.String()
	assert.Contains(t, output, "structured log message")
	assert.Contains(t, output, "component=test")
	assert.Contains(t, output, "action=logging")
}

func TestWithField(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel("info")

	WithField("user", "testuser").Info("user action")

	output := buf.String()
	assert.Contains(t, output, "user action")
	assert.Contains(t, output, "user=testuser")
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel("warn")

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	output := buf.String()
	assert.NotContains(t, output, "debug message")
	assert.NotContains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestGetLogger(t *testing.T) {
	l := GetLogger()
	require.NotNil(t, l)
	assert.IsType(t, &logrus.Logger{}, l)
}

func TestSetFormatter(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel("info")

	jsonFormatter := &logrus.JSONFormatter{}
	SetFormatter(jsonFormatter)

	Info("json test message")

	output := buf.String()
	assert.Contains(t, output, `"msg":"json test message"`)
	assert.Contains(t, output, `"level":"info"`)

	textFormatter := &logrus.TextFormatter{
		DisableColors: true,
	}
	SetFormatter(textFormatter)
}

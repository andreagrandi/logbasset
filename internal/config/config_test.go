package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		envVars      map[string]string
		expectError  bool
		expectToken  string
		expectServer string
	}{
		{
			name: "valid config with token",
			envVars: map[string]string{
				"scalyr_readlog_token": "test-token",
			},
			expectError:  false,
			expectToken:  "test-token",
			expectServer: "https://www.scalyr.com",
		},
		{
			name: "valid config with custom server",
			envVars: map[string]string{
				"scalyr_readlog_token": "test-token",
				"scalyr_server":        "https://eu.scalyr.com",
			},
			expectError:  false,
			expectToken:  "test-token",
			expectServer: "https://eu.scalyr.com",
		},
		{
			name:        "missing token",
			envVars:     map[string]string{},
			expectError: true,
		},
		{
			name: "invalid server URL",
			envVars: map[string]string{
				"scalyr_readlog_token": "test-token",
				"scalyr_server":        "invalid-url",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer clearEnv()

			config, err := New()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				assert.Equal(t, tt.expectToken, config.Token)
				assert.Equal(t, tt.expectServer, config.Server)
				assert.Equal(t, "high", config.Priority)
				assert.False(t, config.Verbose)
			}
		})
	}
}

func TestNewWithoutValidation(t *testing.T) {
	tests := []struct {
		name         string
		envVars      map[string]string
		expectError  bool
		expectToken  string
		expectServer string
	}{
		{
			name: "creates config with token without validation",
			envVars: map[string]string{
				"scalyr_readlog_token": "test-token",
			},
			expectError:  false,
			expectToken:  "test-token",
			expectServer: "https://www.scalyr.com",
		},
		{
			name:         "creates config without token - no validation error",
			envVars:      map[string]string{},
			expectError:  false,
			expectToken:  "",
			expectServer: "https://www.scalyr.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer clearEnv()

			config, err := NewWithoutValidation()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				assert.Equal(t, tt.expectToken, config.Token)
				assert.Equal(t, tt.expectServer, config.Server)
				assert.Equal(t, "high", config.Priority)
				assert.False(t, config.Verbose)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				Token:    "test-token",
				Server:   "https://www.scalyr.com",
				Priority: "high",
			},
			expectError: false,
		},
		{
			name: "missing token",
			config: &Config{
				Server:   "https://www.scalyr.com",
				Priority: "high",
			},
			expectError: true,
		},
		{
			name: "invalid priority",
			config: &Config{
				Token:    "test-token",
				Server:   "https://www.scalyr.com",
				Priority: "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid server URL",
			config: &Config{
				Token:    "test-token",
				Server:   "not-a-url",
				Priority: "high",
			},
			expectError: true,
		},
		{
			name: "invalid log level",
			config: &Config{
				Token:    "test-token",
				Server:   "https://www.scalyr.com",
				Priority: "high",
				LogLevel: "invalid",
			},
			expectError: true,
		},
		{
			name: "valid log level",
			config: &Config{
				Token:    "test-token",
				Server:   "https://www.scalyr.com",
				Priority: "high",
				LogLevel: "debug",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetFromFlags(t *testing.T) {
	config := &Config{
		Token:    "original-token",
		Server:   "https://app.scalyr.com",
		Verbose:  false,
		Priority: "high",
	}

	config.SetFromFlags("new-token", "https://eu.scalyr.com", true, "low", "debug")

	assert.Equal(t, "new-token", config.Token)
	assert.Equal(t, "https://eu.scalyr.com", config.Server)
	assert.True(t, config.Verbose)
	assert.Equal(t, "low", config.Priority)
	assert.Equal(t, "debug", config.LogLevel)
}

func TestFlagOverrideFlow(t *testing.T) {
	clearEnv()
	defer clearEnv()

	// Simulate CLI flow: create config without token, then set from flag
	config, err := NewWithoutValidation()
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "", config.Token)

	// Validation should fail without token
	err = config.Validate()
	assert.Error(t, err)

	// Apply flag values
	config.SetFromFlags("flag-token", "", false, "", "")

	// Validation should now pass
	err = config.Validate()
	assert.NoError(t, err)
	assert.Equal(t, "flag-token", config.Token)
}

func TestFlagOverrideEnv(t *testing.T) {
	clearEnv()
	os.Setenv("scalyr_readlog_token", "env-token")
	defer clearEnv()

	// Create config with env token
	config, err := NewWithoutValidation()
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "env-token", config.Token)

	// Flag should override env
	config.SetFromFlags("flag-token", "", false, "", "")
	assert.Equal(t, "flag-token", config.Token)

	// Validation should pass
	err = config.Validate()
	assert.NoError(t, err)
}

func clearEnv() {
	os.Unsetenv("scalyr_readlog_token")
	os.Unsetenv("scalyr_server")
	os.Unsetenv("scalyr_verbose")
	os.Unsetenv("scalyr_priority")
}

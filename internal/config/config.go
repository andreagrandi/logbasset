package config

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/andreagrandi/logbasset/internal/client"
	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/andreagrandi/logbasset/internal/logging"
	"github.com/spf13/viper"
)

type Config struct {
	Server   string `mapstructure:"server"`
	Token    string `mapstructure:"token"`
	Verbose  bool   `mapstructure:"verbose"`
	Priority string `mapstructure:"priority"`
	LogLevel string `mapstructure:"log_level"`
}

func NewWithoutValidation() (*Config, error) {
	v := viper.New()

	setDefaults(v)

	if err := setupViper(v); err != nil {
		return nil, errors.NewConfigError("failed to setup configuration", err)
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, errors.NewConfigError("failed to unmarshal configuration", err)
	}

	return config, nil
}

func New() (*Config, error) {
	config, err := NewWithoutValidation()
	if err != nil {
		return nil, err
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server", client.DefaultServer)
	v.SetDefault("verbose", false)
	v.SetDefault("priority", "high")
	v.SetDefault("log_level", "info")
}

func setupViper(v *viper.Viper) error {
	v.SetConfigName("logbasset")
	v.SetConfigType("yaml")

	homeDir, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(homeDir, ".config", "logbasset"))
		v.AddConfigPath(filepath.Join(homeDir, ".logbasset"))
	}

	v.AddConfigPath(".")

	v.SetEnvPrefix("scalyr")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.BindEnv("token", "scalyr_readlog_token")
	v.BindEnv("server", "scalyr_server")
	v.BindEnv("log_level", "scalyr_log_level")

	v.ReadInConfig()

	return nil
}

func validateConfig(config *Config) error {
	if config.Token == "" {
		return errors.NewAuthError("API token is required", nil)
	}

	if config.Server != "" {
		if _, err := url.Parse(config.Server); err != nil {
			return errors.NewConfigError("invalid server URL", err)
		}

		if !strings.HasPrefix(config.Server, "http://") && !strings.HasPrefix(config.Server, "https://") {
			return errors.NewConfigError("server URL must start with http:// or https://", nil)
		}
	}

	if config.Priority != "" && config.Priority != "high" && config.Priority != "low" {
		return errors.NewValidationError("priority must be 'high' or 'low'", nil)
	}

	if config.LogLevel != "" {
		validLevels := []string{"debug", "info", "warn", "error"}
		valid := false
		for _, level := range validLevels {
			if strings.ToLower(config.LogLevel) == level {
				valid = true
				break
			}
		}
		if !valid {
			return errors.NewValidationError("log level must be one of: debug, info, warn, error", nil)
		}
	}

	return nil
}

func (c *Config) GetClient() *client.Client {
	return client.New(c.Token, c.Server, c.Verbose)
}

func (c *Config) ApplyLogging() error {
	if c.LogLevel != "" {
		return logging.SetLevel(c.LogLevel)
	}
	return nil
}

func (c *Config) SetFromFlags(token, server string, verbose bool, priority, logLevel string) {
	if token != "" {
		c.Token = token
	}
	if server != "" {
		c.Server = server
	}
	c.Verbose = verbose
	if priority != "" {
		c.Priority = priority
	}
	if logLevel != "" {
		c.LogLevel = logLevel
	}
}

func (c *Config) Validate() error {
	return validateConfig(c)
}

package config

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/andreagrandi/logbasset/internal/client"
	"github.com/andreagrandi/logbasset/internal/errors"
	"github.com/spf13/viper"
)

type Config struct {
	Server   string `mapstructure:"server"`
	Token    string `mapstructure:"token"`
	Verbose  bool   `mapstructure:"verbose"`
	Priority string `mapstructure:"priority"`
}

func New() (*Config, error) {
	v := viper.New()

	setDefaults(v)

	if err := setupViper(v); err != nil {
		return nil, errors.NewConfigError("failed to setup configuration", err)
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, errors.NewConfigError("failed to unmarshal configuration", err)
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

	return nil
}

func (c *Config) GetClient() *client.Client {
	return client.New(c.Token, c.Server, c.Verbose)
}

func (c *Config) SetFromFlags(token, server string, verbose bool, priority string) {
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
}

func (c *Config) Validate() error {
	return validateConfig(c)
}

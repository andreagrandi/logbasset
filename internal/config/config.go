package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/andreagrandi/logbasset/internal/client"
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
		return nil, fmt.Errorf("failed to setup configuration: %w", err)
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
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
		return fmt.Errorf("API token is required. Set scalyr_readlog_token environment variable, use --token flag, or add token to config file.\nYou can find API tokens at https://www.scalyr.com/keys")
	}

	if config.Server != "" {
		if _, err := url.Parse(config.Server); err != nil {
			return fmt.Errorf("invalid server URL '%s': %w", config.Server, err)
		}

		if !strings.HasPrefix(config.Server, "http://") && !strings.HasPrefix(config.Server, "https://") {
			return fmt.Errorf("server URL must start with http:// or https://")
		}
	}

	if config.Priority != "" && config.Priority != "high" && config.Priority != "low" {
		return fmt.Errorf("priority must be 'high' or 'low', got '%s'", config.Priority)
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

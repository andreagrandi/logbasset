package config

import (
	"os"

	"github.com/andreagrandi/logbasset/internal/client"
)

type Config struct {
	Server  string
	Token   string
	Verbose bool
}

func New() *Config {
	return &Config{
		Server:  getServer(),
		Token:   getToken(),
		Verbose: false,
	}
}

func (c *Config) GetClient() *client.Client {
	return client.New(c.Token, c.Server, c.Verbose)
}

func (c *Config) SetVerbose(verbose bool) {
	c.Verbose = verbose
}

func (c *Config) SetServer(server string) {
	c.Server = server
}

func (c *Config) SetToken(token string) {
	c.Token = token
}

func getServer() string {
	server := os.Getenv("scalyr_server")
	if server == "" {
		return client.DefaultServer
	}
	return server
}

func getToken() string {
	return os.Getenv("scalyr_readlog_token")
}

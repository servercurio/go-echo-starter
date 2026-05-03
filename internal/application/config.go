package application

import (
	"github.com/servercurio/go-echo-starter/internal/database"
	"github.com/servercurio/go-echo-starter/internal/env"
	"github.com/servercurio/go-echo-starter/internal/logging"
)

type Config struct {
	Logging  *logging.Config  `yaml:"logging" json:"logging"`
	Server   *ServerConfig    `yaml:"server" json:"server"`
	Proxy    *ProxyConfig     `yaml:"proxy" json:"proxy"`
	Database *database.Config `yaml:"database" json:"database"`
	OpenAPI  *OpenAPIConfig   `yaml:"openapi" json:"openapi"`
}

func (c *Config) FromEnv(prefix string) {
	c.Logging.FromEnv(prefix)
	c.Server.FromEnv(env.AddPrefix(prefix, "server"))
	c.Proxy.FromEnv(env.AddPrefix(prefix, "proxy"))
	c.Database.FromEnv(env.AddPrefix(prefix, "database"))
	c.OpenAPI.FromEnv(env.AddPrefix(prefix, "openapi"))
}

func DefaultConfig() *Config {
	return &Config{
		Logging:  logging.DefaultLoggingConfig(),
		Server:   DefaultServerConfig(),
		Proxy:    DefaultProxyConfig(),
		Database: database.DefaultConfig(),
		OpenAPI:  DefaultOpenAPIConfig(),
	}
}

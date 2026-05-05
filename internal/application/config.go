package application

import (
	"errors"

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

// Validate fans out to every sub-config's Validate and joins the results so
// Configure can return a single error containing every issue found. cmd/daemon
// calls Fatal/os.Exit on a non-nil Configure error, so the joined message
// is what the operator sees in the log; surfacing all issues at once means
// they don't have to iterate boot-fix-boot per problem.
func (c *Config) Validate() error {
	if c == nil {
		return nil
	}
	return errors.Join(
		c.Logging.Validate(),
		c.Server.Validate(),
		c.Proxy.Validate(),
		c.Database.Validate(),
		c.OpenAPI.Validate(),
	)
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

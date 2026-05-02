package application

import "github.com/servercurio/go-echo-starter/internal/env"

type ServerConfig struct {
	Http  *HttpConfig `yaml:"http" json:"http"`
	Https *TlsConfig  `yaml:"https" json:"https"`
}

func (c *ServerConfig) FromEnv(prefix string) {
	c.Http.FromEnv(env.AddPrefix(prefix, "http"))
	c.Https.FromEnv(env.AddPrefix(prefix, "https"))
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Http:  DefaultHttpConfig(),
		Https: DefaultTlsConfig(),
	}
}

package application

import "github.com/servercurio/go-echo-starter/internal/env"

type ServerConfig struct {
	Http     *HttpConfig     `yaml:"http" json:"http"`
	Https    *TlsConfig      `yaml:"https" json:"https"`
	Cors     *CorsConfig     `yaml:"cors" json:"cors"`
	Security *SecurityConfig `yaml:"security" json:"security"`
	Csrf     *CsrfConfig     `yaml:"csrf" json:"csrf"`
}

func (c *ServerConfig) FromEnv(prefix string) {
	c.Http.FromEnv(env.AddPrefix(prefix, "http"))
	c.Https.FromEnv(env.AddPrefix(prefix, "https"))
	c.Cors.FromEnv(env.AddPrefix(prefix, "cors"))
	c.Security.FromEnv(env.AddPrefix(prefix, "security"))
	c.Csrf.FromEnv(env.AddPrefix(prefix, "csrf"))
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Http:     DefaultHttpConfig(),
		Https:    DefaultTlsConfig(),
		Cors:     DefaultCorsConfig(),
		Security: DefaultSecurityConfig(),
		Csrf:     DefaultCsrfConfig(),
	}
}

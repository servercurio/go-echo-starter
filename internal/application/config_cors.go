package application

import (
	"strings"

	"github.com/rs/zerolog"
	"github.com/servercurio/go-echo-starter/internal/env"
)

// CorsConfig captures the cross-origin policy applied to both the HTTP and
// HTTPS servers.
//
// The starter ships with empty AllowOrigins by default, which disables CORS
// entirely. Downstream consumers must opt in explicitly. This is deliberate:
// once the server starts handling auth/cookies, a permissive CORS default
// becomes a cross-origin leakage vector.
type CorsConfig struct {
	// AllowOrigins is the list of origins allowed to issue cross-origin
	// requests. When empty, the CORS middleware adds no headers and the
	// browser will block cross-origin requests.
	AllowOrigins []string `yaml:"allowOrigins" json:"allowOrigins"`

	// AllowMethods is the list of HTTP methods the browser may use in a
	// cross-origin request. When empty, defaults to GET, HEAD, PUT, PATCH,
	// POST, DELETE.
	AllowMethods []string `yaml:"allowMethods" json:"allowMethods"`

	// AllowHeaders is the list of request headers the browser may include in
	// a cross-origin request.
	AllowHeaders []string `yaml:"allowHeaders" json:"allowHeaders"`

	// AllowCredentials indicates whether the browser should expose response
	// to frontend JavaScript when the request includes credentials. Pairs
	// with explicit AllowOrigins (wildcard cannot be used with credentials).
	AllowCredentials bool `yaml:"allowCredentials" json:"allowCredentials"`

	// MaxAge is the number of seconds a browser may cache the preflight
	// response. Zero leaves it unset.
	MaxAge int `yaml:"maxAge" json:"maxAge"`
}

// Enabled reports whether at least one origin has been configured. When false,
// the CORS middleware should be skipped entirely.
func (c *CorsConfig) Enabled() bool {
	return c != nil && len(c.AllowOrigins) > 0
}

func (c *CorsConfig) FromEnv(prefix string) {
	var allowOrigins string
	env.SetStringValue(prefix, "allow_origins", &allowOrigins)
	if allowOrigins = strings.TrimSpace(allowOrigins); allowOrigins != "" {
		c.AllowOrigins = splitAndTrim(allowOrigins)
	}

	var allowMethods string
	env.SetStringValue(prefix, "allow_methods", &allowMethods)
	if allowMethods = strings.TrimSpace(allowMethods); allowMethods != "" {
		c.AllowMethods = splitAndTrim(allowMethods)
	}

	var allowHeaders string
	env.SetStringValue(prefix, "allow_headers", &allowHeaders)
	if allowHeaders = strings.TrimSpace(allowHeaders); allowHeaders != "" {
		c.AllowHeaders = splitAndTrim(allowHeaders)
	}

	env.SetBoolValue(prefix, "allow_credentials", &c.AllowCredentials)
	env.SetIntValue(prefix, "max_age", &c.MaxAge)
}

func (c *CorsConfig) MarshalZerologObject(e *zerolog.Event) {
	e.Strs("allowOrigins", c.AllowOrigins).
		Strs("allowMethods", c.AllowMethods).
		Strs("allowHeaders", c.AllowHeaders).
		Bool("allowCredentials", c.AllowCredentials).
		Int("maxAge", c.MaxAge).
		Bool("enabled", c.Enabled())
}

func DefaultCorsConfig() *CorsConfig {
	return &CorsConfig{
		AllowOrigins:     []string{},
		AllowMethods:     []string{},
		AllowHeaders:     []string{},
		AllowCredentials: false,
		MaxAge:           0,
	}
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

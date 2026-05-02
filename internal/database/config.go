// Package database manages the application's optional SQL database connection
// pool, schema migrations, and ORM singleton. The starter ships PostgreSQL
// (pgx) bindings by default; replace the driver and dialect to swap engines.
//
// The database is opt-in: an empty DSN at startup means Connect/Migrate are
// skipped and the readiness probe ignores DB state. This keeps the starter
// usable as a pure HTTP server with no database.
package database

import (
	"github.com/rs/zerolog"
	"github.com/servercurio/go-echo-starter/internal/env"
	"github.com/servercurio/go-echo-starter/internal/obfusicate"
)

// Config holds the database connection parameters.
type Config struct {
	// Driver is the database/sql driver name to register (e.g. "pgx").
	Driver string `yaml:"driver" json:"driver"`

	// DSN is the connection string passed to sql.Open. An empty DSN disables
	// the database subsystem entirely: Connect/Migrate become no-ops and
	// readiness checks skip the database probe.
	DSN string `yaml:"dsn" json:"dsn"`
}

// Enabled reports whether the database subsystem should be initialised. Returns
// true when a non-empty DSN has been configured.
func (c *Config) Enabled() bool {
	return c != nil && c.DSN != ""
}

// FromEnv overlays Config fields with values from environment variables under
// the given prefix (e.g. APP_DATABASE_DRIVER, APP_DATABASE_DSN).
func (c *Config) FromEnv(prefix string) {
	env.SetStringValue(prefix, "driver", &c.Driver)
	env.SetStringValue(prefix, "dsn", &c.DSN)
}

// MarshalZerologObject writes the database configuration to a zerolog event.
// The DSN is obfuscated to keep credentials out of logs.
func (c *Config) MarshalZerologObject(e *zerolog.Event) {
	e.Str("driver", c.Driver).
		Str("dsn", obfusicate.ConcealUriCredential(c.DSN)).
		Bool("enabled", c.Enabled())
}

// DefaultConfig returns a disabled-by-default Config. Set DSN (via config file
// or env var) to enable the database subsystem.
func DefaultConfig() *Config {
	return &Config{
		Driver: "pgx",
		DSN:    "",
	}
}

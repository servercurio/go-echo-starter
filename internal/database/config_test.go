package database

import (
	"testing"

	asrt "github.com/stretchr/testify/assert"
)

func TestConfig_EnabledRequiresNonEmptyDSN(t *testing.T) {
	assert := asrt.New(t)

	assert.False((*Config)(nil).Enabled(), "nil config must not be enabled")
	assert.False((&Config{Driver: "pgx", DSN: ""}).Enabled(), "empty DSN means disabled")
	assert.True((&Config{Driver: "pgx", DSN: "postgres://localhost/app"}).Enabled())
}

func TestConfig_FromEnvOverlaysDriverAndDSN(t *testing.T) {
	assert := asrt.New(t)

	t.Setenv("APP_DATABASE_DRIVER", "pgx")
	t.Setenv("APP_DATABASE_DSN", "postgres://app@db.example.com:5432/app")

	cfg := DefaultConfig()
	cfg.FromEnv("APP_DATABASE")

	assert.Equal("pgx", cfg.Driver)
	assert.Equal("postgres://app@db.example.com:5432/app", cfg.DSN)
	assert.True(cfg.Enabled())
}

func TestDefaultConfig_IsDisabledByDefault(t *testing.T) {
	assert := asrt.New(t)
	cfg := DefaultConfig()

	assert.False(cfg.Enabled(), "starter must default to no database so it runs without external dependencies")
	assert.NotEmpty(cfg.Driver, "default driver should be populated so enabling the DB only requires setting a DSN")
}

package logging

import (
	"strings"

	"github.com/rs/zerolog"
	"github.com/servercurio/go-echo-starter/internal/env"
)

// Config represents the logging configuration for the application.
type Config struct {
	Daemon     *LoggerConfig `yaml:"daemon" json:"daemon"`
	HttpAccess *LoggerConfig `yaml:"httpAccess" json:"httpAccess"`
}

func (c *Config) FromEnv(prefix string) {
	c.Daemon.FromEnv(env.AddPrefix(prefix, "daemon_log"))
	c.HttpAccess.FromEnv(env.AddPrefix(prefix, "http_access_log"))
}

// LoggerConfig represents the configuration for a single logger.
type LoggerConfig struct {
	// Enabled indicates whether the logger is enabled.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Level is the verbosity of the logging output.
	Level string `yaml:"level" json:"level"`

	// PrettyPrint enables human-readable output.
	PrettyPrint bool `yaml:"prettyPrint" json:"prettyPrint"`

	// IncludeCaller enables caller information in the log output.
	IncludeCaller bool `yaml:"includeCaller" json:"includeCaller"`
}

func (l *LoggerConfig) MarshalZerologObject(e *zerolog.Event) {
	e.Bool("enabled", l.Enabled)
	e.Str("logLevel", strings.ToLower(l.Level))
	e.Bool("prettyPrint", l.PrettyPrint)
	e.Bool("includeCaller", l.IncludeCaller)
}

func (l *LoggerConfig) FromEnv(prefix string) {
	env.SetBoolValue(prefix, "enabled", &l.Enabled)
	env.SetStringValue(prefix, "level", &l.Level)
	env.SetBoolValue(prefix, "pretty_print", &l.PrettyPrint)
	env.SetBoolValue(prefix, "include_caller", &l.IncludeCaller)
}

func DefaultLoggingConfig() *Config {
	return &Config{
		Daemon:     NewLoggerConfig(zerolog.LevelInfoValue, true, false, true),
		HttpAccess: NewLoggerConfig(zerolog.LevelErrorValue, false, false, false),
	}
}

func NewConfigFromEnv(prefix string) *Config {
	cfg := DefaultLoggingConfig()
	cfg.FromEnv(prefix)

	return cfg
}

func NewLoggerConfig(level string, prettyPrint, includeCaller, enabled bool) *LoggerConfig {
	level = strings.ToLower(strings.TrimSpace(level))
	if level == "" || !isValidLevel(level) {
		level = zerolog.LevelInfoValue
	}

	return &LoggerConfig{
		Enabled:       enabled,
		Level:         level,
		PrettyPrint:   prettyPrint,
		IncludeCaller: includeCaller,
	}
}

func isValidLevel(level string) bool {
	if _, err := zerolog.ParseLevel(level); err == nil {
		return true
	}

	return false
}

func parseLevel(level string) zerolog.Level {
	if l, err := zerolog.ParseLevel(level); err == nil {
		return l
	}

	return zerolog.NoLevel
}

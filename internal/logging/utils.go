package logging

import (
	"log"

	"github.com/rs/zerolog"
	"github.com/servercurio/go-echo-starter/internal/version"
)

// AsStdLogger wraps a zerolog.Logger so it can be passed to APIs that expect
// the standard library *log.Logger (e.g. goose's SetLogger). Each Write call
// from the std logger is emitted as a single zerolog event at the wrapped
// logger's current level.
func AsStdLogger(logger zerolog.Logger) *log.Logger {
	return log.New(logger, "", 0)
}

func NotifyDaemonStartup(name string, cfg *Config) {
	Initialize(cfg)

	Daemon.Info().
		Str("version", version.Number()).
		Str("commit", version.Commit()).
		Msgf("%s daemon started", name)
}

func NotifyDaemonLoggingStartup(cfg *Config) {
	Initialize(cfg)

	Daemon.Info().
		EmbedObject(cfg.Daemon).
		Msg("daemon logging")
}

func NotifyHttpLoggingStartup(cfg *Config) {
	Initialize(cfg)

	Daemon.Info().
		EmbedObject(cfg.HttpAccess).
		Msg("http access logging")
}

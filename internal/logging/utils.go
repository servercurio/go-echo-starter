package logging

import (
	"github.com/servercurio/go-echo-starter/internal/version"
)

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

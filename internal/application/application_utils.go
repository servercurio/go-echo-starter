package application

import (
	"os"

	"github.com/joomcode/errorx"
	"github.com/servercurio/go-echo-starter/internal/logging"
)

func (app *Application) resolveHomeDirectory() {
	var err error

	if app.userHomeDirectory, err = os.UserHomeDir(); err == nil {
		logging.Daemon.
			Trace().
			Str("path", app.userHomeDirectory).
			Msg("resolved user home directory")
	} else {
		wErr := errorx.ExternalError.Wrap(err, "failed to resolve user home directory")
		logging.Daemon.
			Warn().
			Err(wErr).
			Msg("failed to resolve user home directory")
		app.userHomeDirectory = "."
	}
}

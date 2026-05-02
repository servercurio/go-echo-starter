package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/joomcode/errorx"
	"github.com/servercurio/go-echo-starter/internal/logging"
)

func (app *Application) configureHttpServer() error {
	app.httpServer.Use(app.middleware...)

	if app.config.Server.Https != nil && app.config.Server.Https.Enabled {
		app.httpServer.Pre(HTTPSRedirectWithConfig(app.config.Server.Https))
	}

	app.httpServer.HideBanner = true
	app.httpServer.Logger.SetOutput(io.Discard)
	app.httpServer.StdLogger.SetOutput(io.Discard)

	return nil
}

func (app *Application) startHttpServer() {
	address := fmt.Sprintf("%s:%d", app.config.Server.Http.BindAddress, app.config.Server.Http.Port)
	logging.Daemon.Info().
		Str("address", address).
		Bool("httpsRedirect", app.config.Server.Https != nil && app.config.Server.Https.Enabled).
		Msg("http server started")

	if err := app.httpServer.Start(address); !errors.Is(err, http.ErrServerClosed) {
		logging.Daemon.Error().
			Err(errorx.EnsureStackTrace(err)).
			Msg("http server shutting down due to an error")
	}

}

func (app *Application) shutdownHttpServer() {
	var ctx context.Context
	var cancel context.CancelFunc

	if app.config.Server.Http.ShutdownTimeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), app.config.Server.Http.ShutdownTimeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}

	defer cancel()
	if err := app.httpServer.Shutdown(ctx); err != nil {
		logging.Daemon.Error().
			Err(err).
			Msg("http server shutdown failed")
	}
}

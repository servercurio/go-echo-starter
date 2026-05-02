package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/joomcode/errorx"
	"github.com/labstack/echo/v5"
	"github.com/servercurio/go-echo-starter/internal/logging"
)

func (app *Application) configureHttpServer() error {
	app.httpServer.Use(app.middleware...)

	if app.config.Server.Https != nil && app.config.Server.Https.Enabled {
		app.httpServer.Pre(HTTPSRedirectWithConfig(app.config.Server.Https))
	}

	app.httpServer.Logger = slog.New(slog.DiscardHandler)

	return nil
}

func (app *Application) startHttpServer() {
	address := fmt.Sprintf("%s:%d", app.config.Server.Http.BindAddress, app.config.Server.Http.Port)
	logging.Daemon.Info().
		Str("address", address).
		Bool("httpsRedirect", app.config.Server.Https != nil && app.config.Server.Https.Enabled).
		Msg("http server started")

	sc := &echo.StartConfig{
		HideBanner: true,
		Address:    address,
	}

	if err := sc.Start(context.Background(), app.httpServer); !errors.Is(err, http.ErrServerClosed) {
		logging.Daemon.Error().
			Err(errorx.EnsureStackTrace(err)).
			Msg("http server shutting down due to an error")
	}

}

func (app *Application) shutdownHttpServer() {
}

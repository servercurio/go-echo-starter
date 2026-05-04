package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/joomcode/errorx"
	"github.com/labstack/echo/v5"
	mw "github.com/labstack/echo/v5/middleware"
	"github.com/servercurio/go-echo-starter/internal/logging"
)

func (app *Application) configureHttpServer() error {
	app.httpServer.Use(app.middleware...)

	if cors := CorsMiddleware(app.config.Server.Cors); cors != nil {
		app.httpServer.Use(cors)
	}

	bodyLimit, err := parseByteSize(app.config.Server.Http.MaxBodySize)
	if err != nil {
		return errorx.IllegalArgument.Wrap(err, "invalid http max body size %q", app.config.Server.Http.MaxBodySize)
	}
	app.httpServer.Use(mw.BodyLimit(bodyLimit))

	if app.config.Server.Https != nil && app.config.Server.Https.Enabled {
		app.httpServer.Pre(HTTPSRedirectWithConfig(app.config.Server.Https))
	}

	app.httpServer.Logger = slog.New(slog.DiscardHandler)

	return nil
}

func (app *Application) startHttpServer(ctx context.Context) {
	address := fmt.Sprintf("%s:%d", app.config.Server.Http.BindAddress, app.config.Server.Http.Port)
	logging.Daemon.Info().
		Str("address", address).
		Bool("httpsRedirect", app.config.Server.Https != nil && app.config.Server.Https.Enabled).
		Msg("http server started")

	httpCfg := app.config.Server.Http
	sc := &echo.StartConfig{
		HideBanner:      true,
		Address:         address,
		GracefulTimeout: httpCfg.ShutdownTimeout,
		BeforeServeFunc: func(s *http.Server) error {
			s.ReadTimeout = httpCfg.ReadTimeout
			s.ReadHeaderTimeout = httpCfg.ReadHeaderTimeout
			s.WriteTimeout = httpCfg.WriteTimeout
			s.IdleTimeout = httpCfg.IdleTimeout
			return nil
		},
	}

	if err := sc.Start(ctx, app.httpServer); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logging.Daemon.Error().
			Err(errorx.EnsureStackTrace(err)).
			Msg("http server shutting down due to an error")
	}
}

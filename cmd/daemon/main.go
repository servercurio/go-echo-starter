package main

import (
	"os"

	_ "github.com/joomcode/errorx"
	_ "github.com/labstack/echo/v5"
	"github.com/servercurio/go-echo-starter/internal/api"
	"github.com/servercurio/go-echo-starter/internal/application"
	"github.com/servercurio/go-echo-starter/internal/logging"
	"github.com/servercurio/go-echo-starter/internal/router"
)

func main() {
	cfg := application.DefaultConfig()
	app := application.NewApplication(cfg)

	routerCfg := router.NewConfig()
	// Readiness is the conjunction of application-lifecycle readiness and
	// database health: /readyz returns 200 only when the server has finished
	// startup AND the database (if configured) is reachable. When the
	// database subsystem is disabled, IsDatabaseHealthy() always returns
	// true, so the probe collapses back to lifecycle-only readiness.
	routerCfg.ReadinessProbe = func() bool {
		return app.IsReady() && app.IsDatabaseHealthy()
	}

	_ = app.RegisterModule(api.Module(routerCfg))

	if err := app.Configure(); err != nil {
		logging.Daemon.
			Fatal().
			Err(err).
			Msgf("an unhandled error occurred during %s configuration", app.Name)
		os.Exit(1)
	}

	if err := app.Initialize(); err != nil {
		logging.Daemon.
			Fatal().
			Err(err).
			Msgf("an unhandled error occurred during %s initialization", app.Name)
		os.Exit(1)
	}

	ec, err := app.Start()

	if err != nil {
		logging.Daemon.
			Fatal().
			Err(err).
			Msgf("an unhandled error occurred causing %s to terminate unexpectedly", app.Name)
		os.Exit(ec)
	}
}

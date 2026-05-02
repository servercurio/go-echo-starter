package main

import (
	"os"

	_ "github.com/joomcode/errorx"
	_ "github.com/labstack/echo/v4"
	"github.com/servercurio/go-echo-starter/internal/api"
	"github.com/servercurio/go-echo-starter/internal/application"
	"github.com/servercurio/go-echo-starter/internal/logging"
	"github.com/servercurio/go-echo-starter/internal/router"
)

func main() {
	cfg := application.DefaultConfig()
	app := application.NewApplication(cfg)

	_ = app.RegisterModule(api.Module(router.NewConfig()))

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

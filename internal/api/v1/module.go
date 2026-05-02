package v1

import (
	"github.com/servercurio/go-echo-starter/internal/api/std/module"
	"github.com/servercurio/go-echo-starter/internal/router"
)

const (
	moduleIdentifier = "v1"
	moduleName       = "api/v1"
	modulePrefix     = "v1"
)

func Module(cfg *router.Config) router.Module {
	return module.New(
		moduleIdentifier,
		moduleName,
		modulePrefix,
		module.WithRoutes(
			LivenessRoute(),
			ReadinessRoute(cfg),
			HealthRoute(cfg),
		),
	)
}

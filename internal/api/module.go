package api

import (
	"github.com/servercurio/go-echo-starter/internal/api/std/module"
	v1 "github.com/servercurio/go-echo-starter/internal/api/v1"
	"github.com/servercurio/go-echo-starter/internal/router"
)

const (
	moduleIdentifier = "api"
	moduleName       = "api"
	modulePrefix     = "api"
)

func Module(cfg *router.Config) router.Module {
	return module.New(
		moduleIdentifier,
		moduleName,
		modulePrefix,
		module.WithSubModules(v1.Module(cfg)),
	)
}

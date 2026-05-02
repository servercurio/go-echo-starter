package router

import "github.com/labstack/echo/v4"

type Module interface {
	// Id is the unique identifier for this module.
	Id() string

	// Name is the user-friendly name for this module.
	Name() string

	// Prefix is the URI path prepended to all routes registered with this module.
	Prefix() string

	// Routes returns the list of Route instances registered with this module.
	Routes() []Route

	// SubModules returns the list of registered SubModules
	SubModules() []Module

	// HasRoutes returns true if this Module contains direct routes.
	HasRoutes() bool

	// HasSubModules returns true if this Module contains submodules.
	HasSubModules() bool

	// Middleware returns the list of echo.MiddlewareFunc methods which should be registered for
	// this module.
	Middleware() []echo.MiddlewareFunc

	// AttachGroup registers the echo.Group with the Module.
	AttachGroup(group *echo.Group)

	// Group returns the associated echo.Group instance.
	Group() *echo.Group
}

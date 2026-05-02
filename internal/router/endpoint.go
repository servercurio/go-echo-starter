package router

import "github.com/labstack/echo/v4"

type Endpoint interface {
	// Id returns the unique identifier of the endpoint.
	Id() string

	// Name returns the user readable name of the endpoint.
	Name() string

	// Methods returns the list of HTTP methods supported by the endpoint.
	Methods() []string

	// Middleware returns the list of echo.MiddlewareFunc methods which should be registered for
	// this module.
	Middleware() []echo.MiddlewareFunc

	// HandleRequest processes the incoming request and produces a response.
	HandleRequest(c echo.Context) error

	// AttachGroup registers the echo.Group with the Module.
	AttachGroup(group *echo.Group)

	// Group returns the associated echo.Group instance.
	Group() *echo.Group
}

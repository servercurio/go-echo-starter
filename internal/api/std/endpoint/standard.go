package endpoint

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"slices"
)

type Standard struct {
	id         string
	name       string
	methods    []string
	middleware []echo.MiddlewareFunc
	group      *echo.Group
	handler    echo.HandlerFunc
}

func (e *Standard) Id() string {
	return e.id
}

func (e *Standard) Name() string {
	return e.name
}

func (e *Standard) Methods() []string {
	return e.methods
}

func (e *Standard) Middleware() []echo.MiddlewareFunc {
	return e.middleware
}

func (e *Standard) AttachGroup(group *echo.Group) {
	if group == nil || e.group != nil {
		return
	}

	e.group = group
}

func (e *Standard) HandleRequest(c echo.Context) error {
	if e.handler == nil {
		return c.NoContent(http.StatusNotImplemented)
	}

	return e.handler(c)
}

func (e *Standard) Group() *echo.Group {
	return e.group
}

func New(id, name string, options ...Option) *Standard {
	std := &Standard{
		id:         id,
		name:       name,
		methods:    make([]string, 0),
		middleware: make([]echo.MiddlewareFunc, 0),
	}

	for _, opt := range options {
		opt(std)
	}

	// Remove duplicates from methods
	slices.Sort(std.methods)
	std.methods = slices.Compact(std.methods)

	return std
}

// Option is a function that configures the Standard endpoint.
type Option func(*Standard)

// WithHandler sets the handler for the Standard endpoint.
func WithHandler(handler echo.HandlerFunc) Option {
	return func(e *Standard) {
		e.handler = handler
	}
}

// WithMiddleware adds middleware to the Standard endpoint.
func WithMiddleware(middleware ...echo.MiddlewareFunc) Option {
	return func(e *Standard) {
		if e.middleware == nil {
			e.middleware = make([]echo.MiddlewareFunc, 0)
		}
		e.middleware = append(e.middleware, middleware...)
	}
}

// WithMethods adds HTTP methods to the Standard endpoint.
func WithMethods(methods ...string) Option {
	return func(e *Standard) {
		if e.methods == nil {
			e.methods = make([]string, 0)
		}
		e.methods = append(e.methods, methods...)
	}
}

func WithGetMethod() Option {
	return func(e *Standard) {
		e.methods = append(e.methods, http.MethodGet)
	}
}

func WithPostMethod() Option {
	return func(e *Standard) {
		e.methods = append(e.methods, http.MethodPost)
	}
}

func WithPutMethod() Option {
	return func(e *Standard) {
		e.methods = append(e.methods, http.MethodPut)
	}
}

func WithDeleteMethod() Option {
	return func(e *Standard) {
		e.methods = append(e.methods, http.MethodDelete)
	}
}

func WithPatchMethod() Option {
	return func(e *Standard) {
		e.methods = append(e.methods, http.MethodPatch)
	}
}

func WithOptionsMethod() Option {
	return func(e *Standard) {
		e.methods = append(e.methods, http.MethodOptions)
	}
}

func WithHeadMethod() Option {
	return func(e *Standard) {
		e.methods = append(e.methods, http.MethodHead)
	}
}

func WithTraceMethod() Option {
	return func(e *Standard) {
		e.methods = append(e.methods, http.MethodTrace)
	}
}

func WithConnectMethod() Option {
	return func(e *Standard) {
		e.methods = append(e.methods, http.MethodConnect)
	}
}

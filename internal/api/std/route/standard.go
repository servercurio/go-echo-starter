package route

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/servercurio/go-echo-starter/internal/logging"
	"github.com/servercurio/go-echo-starter/internal/router"
)

type Standard struct {
	id         string
	name       string
	path       string
	endpoints  map[string]router.Endpoint
	middleware []echo.MiddlewareFunc
	group      *echo.Group
}

func (r *Standard) Id() string {
	return r.id
}

func (r *Standard) Name() string {
	return r.name
}

func (r *Standard) Path() string {
	return r.path
}

func (r *Standard) Middleware() []echo.MiddlewareFunc {
	return r.middleware
}

func (r *Standard) Endpoints() []router.Endpoint {
	var ret []router.Endpoint
	for _, v := range r.endpoints {
		ret = append(ret, v)
	}
	return ret
}

func (r *Standard) AttachGroup(group *echo.Group) {
	if group == nil || r.group != nil {
		return
	}

	r.group = group
	group.Use(r.middleware...)

	for _, ep := range r.endpoints {
		ep.AttachGroup(group)

		path := r.path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		for _, method := range ep.Methods() {
			group.Add(method, path, ep.HandleRequest, ep.Middleware()...)
		}

		logging.Daemon.
			Info().
			Str("name", ep.Name()).
			Strs("methods", ep.Methods()).
			Msg("http router - registering route endpoint")
	}
}

func (r *Standard) Group() *echo.Group {
	return r.group
}

func New(id, name, path string, options ...Option) *Standard {
	std := &Standard{
		id:         id,
		name:       name,
		path:       path,
		endpoints:  make(map[string]router.Endpoint),
		middleware: make([]echo.MiddlewareFunc, 0),
	}

	for _, opt := range options {
		opt(std)
	}

	return std
}

type Option func(std *Standard)

func WithMiddleware(middleware ...echo.MiddlewareFunc) Option {
	return func(r *Standard) {
		r.middleware = append(r.middleware, middleware...)
	}
}

func WithEndpoints(endpoints ...router.Endpoint) Option {
	return func(r *Standard) {
		if r.endpoints == nil {
			r.endpoints = make(map[string]router.Endpoint)
		}

		for _, ep := range endpoints {
			if _, ok := r.endpoints[ep.Id()]; !ok {
				r.endpoints[ep.Id()] = ep
			}
		}
	}
}

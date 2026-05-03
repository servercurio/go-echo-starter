package module

import (
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/servercurio/go-echo-starter/internal/logging"
	"github.com/servercurio/go-echo-starter/internal/router"
)

type Standard struct {
	id         string
	name       string
	prefix     string
	routes     map[string]router.Route
	modules    map[string]router.Module
	middleware []echo.MiddlewareFunc
	group      *echo.Group
}

func (m *Standard) Id() string {
	return m.id
}

func (m *Standard) Name() string {
	return m.name
}

// Prefix returns the URL prefix this module mounts under, normalised to
// either:
//
//   - the empty string for root-mounted modules (prefix was "" or "/"), so
//     the caller's `server.Group(prefix, ...)` invocation doesn't introduce
//     a stray leading "/" that Echo would naively concatenate with route
//     paths into "//route" and 404 every request, or
//
//   - a leading-slash form ("/api", "/v1") for nested modules.
//
// Without this normalisation, a module constructed with prefix=""
// effectively can't expose top-level routes — a real footgun, since "no
// prefix" is exactly what the openapi / swagger modules want.
func (m *Standard) Prefix() string {
	trimmed := strings.Trim(m.prefix, router.PathSeparator)
	if trimmed == "" {
		return ""
	}
	return router.PathSeparator + trimmed
}

func (m *Standard) Routes() []router.Route {
	var ret []router.Route
	for _, v := range m.routes {
		ret = append(ret, v)
	}
	return ret
}

func (m *Standard) SubModules() []router.Module {
	var ret []router.Module
	for _, v := range m.modules {
		ret = append(ret, v)
	}
	return ret
}

func (m *Standard) HasRoutes() bool {
	return len(m.routes) > 0
}

func (m *Standard) HasSubModules() bool {
	return len(m.modules) > 0
}

func (m *Standard) Middleware() []echo.MiddlewareFunc {
	return m.middleware
}

func (m *Standard) AttachGroup(group *echo.Group) {
	if group == nil || m.group != nil {
		return
	}

	logging.Daemon.Info().
		Str("name", m.Name()).
		Str("prefix", m.Prefix()).
		Msg("http router - registering module prefix")

	m.group = group
	m.group.Use(m.middleware...)

	if m.HasSubModules() {
		for _, module := range m.modules {
			g := m.group.Group(module.Prefix(), module.Middleware()...)
			module.AttachGroup(g)
		}
	}

	if m.HasRoutes() {
		for _, route := range m.routes {
			route.AttachGroup(m.group)
		}
	}
}

func (m *Standard) Group() *echo.Group {
	return m.group
}

type Option func(m *Standard)

func New(id, name, prefix string, opts ...Option) router.Module {
	m := &Standard{
		id:         id,
		name:       name,
		prefix:     prefix,
		routes:     make(map[string]router.Route),
		modules:    make(map[string]router.Module),
		middleware: make([]echo.MiddlewareFunc, 0),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func WithMiddleware(mw ...echo.MiddlewareFunc) Option {
	return func(m *Standard) {
		if m.middleware == nil {
			m.middleware = make([]echo.MiddlewareFunc, 0)
		}

		m.middleware = append(m.middleware, mw...)
	}
}

func WithRoutes(routes ...router.Route) Option {
	return func(m *Standard) {
		if m.routes == nil {
			m.routes = make(map[string]router.Route)
		}

		for _, route := range routes {
			if _, ok := m.routes[route.Id()]; !ok {
				m.routes[route.Id()] = route
			}
		}
	}
}

func WithSubModules(modules ...router.Module) Option {
	return func(m *Standard) {
		if m.modules == nil {
			m.modules = make(map[string]router.Module)
		}

		for _, mod := range modules {
			if _, ok := m.modules[mod.Id()]; !ok {
				m.modules[mod.Id()] = mod
			}
		}
	}
}

package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/servercurio/go-echo-starter/internal/api/std/endpoint"
	"github.com/servercurio/go-echo-starter/internal/api/std/route"
	"github.com/servercurio/go-echo-starter/internal/router"
)

// LivenessRoute mounts the Kubernetes-style liveness probe at /livez.
//
// Liveness reports whether the process can serve HTTP at all; it must not
// depend on application-level readiness, downstream dependencies, or shutdown
// state. A liveness failure tells the kubelet to restart the pod, so always
// returning 200 here is correct as long as the HTTP listener is up — if the
// listener is down, the request never reaches this handler in the first place.
func LivenessRoute() router.Route {
	return route.New("liveness", "liveness", "/livez",
		route.WithEndpoints(
			endpoint.New("liveness-get", "liveness-get",
				endpoint.WithGetMethod(),
				endpoint.WithHandler(func(c *echo.Context) error {
					return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
				}),
			),
		),
	)
}

// ReadinessRoute mounts the Kubernetes-style readiness probe at /readyz.
//
// Readiness reports whether the process is currently fit to receive traffic.
// It returns 503 during startup before the application is ready and after a
// shutdown signal has been received, so the load balancer can route traffic
// away cleanly while in-flight requests drain.
func ReadinessRoute(cfg *router.Config) router.Route {
	return route.New("readiness", "readiness", "/readyz",
		route.WithEndpoints(readinessEndpoint("readiness-get", "readiness-get", cfg)),
	)
}

// HealthRoute mounts the legacy /healthz path. Kept as an alias for /readyz
// so existing consumers (load balancers, uptime checks) that default to
// /healthz keep working without configuration changes. New code should target
// /livez or /readyz explicitly.
func HealthRoute(cfg *router.Config) router.Route {
	return route.New("health", "health", "/healthz",
		route.WithEndpoints(readinessEndpoint("health-get", "health-get", cfg)),
	)
}

// readinessEndpoint is the shared handler used by /readyz and /healthz so the
// readiness contract has exactly one definition. A nil ReadinessProbe is
// treated as not-ready (fail closed) — the test in health_test.go pins this
// behaviour so a future refactor can't silently turn a misconfiguration into
// a misleading 200 OK.
func readinessEndpoint(id, name string, cfg *router.Config) router.Endpoint {
	return endpoint.New(id, name,
		endpoint.WithGetMethod(),
		endpoint.WithHandler(func(c *echo.Context) error {
			if cfg.ReadinessProbe != nil && cfg.ReadinessProbe() {
				return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
			}
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		}),
	)
}

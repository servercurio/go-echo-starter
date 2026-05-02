package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/servercurio/go-echo-starter/internal/api/std/endpoint"
	"github.com/servercurio/go-echo-starter/internal/api/std/route"
	"github.com/servercurio/go-echo-starter/internal/router"
)

func HealthRoute() router.Route {
	return route.New("health", "health", "/healthz",
		route.WithEndpoints(
			endpoint.New("health-get", "health-get",
				endpoint.WithGetMethod(),
				endpoint.WithHandler(func(c *echo.Context) error {
					return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
				}),
			),
		),
	)

}

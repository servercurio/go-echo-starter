package openapi

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/servercurio/go-echo-starter/internal/api/std/endpoint"
	"github.com/servercurio/go-echo-starter/internal/api/std/module"
	"github.com/servercurio/go-echo-starter/internal/api/std/route"
	"github.com/servercurio/go-echo-starter/internal/router"

	echoSwagger "github.com/swaggo/echo-swagger/v2"
)

// SwaggerOptions configures the Swagger UI mount.
type SwaggerOptions struct {
	// Path is the URL prefix Swagger UI is mounted under (e.g. "/swagger").
	// The actual UI is served at <Path>/index.html and the wildcard
	// <Path>/* catches the supporting JS/CSS asset requests.
	Path string

	// SpecURL is the URL the Swagger UI fetches the OpenAPI document from.
	// Typically "/openapi.yaml" or "/openapi.json" served by Module above.
	SpecURL string
}

// SwaggerModule returns a router.Module that mounts Swaggo's echo-swagger v2
// handler (OpenAPI 3.0 flavour) under opts.Path. Callers should only invoke
// this when the swagger UI is enabled — leaving the route unregistered
// entirely is preferable to returning 404 from a registered handler.
//
// Two routes are registered:
//
//   - <opts.Path>      → 301 redirect to <opts.Path>/index.html. Echo's
//     wildcard `/swagger/*` only matches paths starting with `/swagger/`,
//     not the bare `/swagger`, so without this users typing the prefix
//     into a browser would 404. We redirect directly to /index.html (not
//     to /<Path>/) to skip the intermediate redirect that echo-swagger's
//     own handler would otherwise emit, saving a round trip.
//
//   - <opts.Path>/*    → echo-swagger handler. Serves index.html and all
//     supporting UI assets (JS/CSS/JSON).
func SwaggerModule(opts SwaggerOptions) router.Module {
	cleanPath := "/" + strings.Trim(opts.Path, "/")
	wildcard := cleanPath + "/*"
	redirectTarget := cleanPath + "/index.html"

	// echoSwagger.URL is an APPEND operation, not a SET — it tacks our spec
	// URL onto the default ["doc.json", "doc.yaml"] list. Swagger UI then
	// tries "doc.json" first (no such route → 500/404) before falling back
	// to ours, surfacing a "Fetch error" banner. Reset the slice directly
	// so only our spec is offered.
	specURL := opts.SpecURL
	handler := echoSwagger.EchoWrapHandlerV3(func(c *echoSwagger.Config) {
		c.URLs = []string{specURL}
	})

	return module.New("swagger", "swagger", "",
		module.WithRoutes(
			route.New("swagger-redirect", "swagger-redirect", cleanPath,
				route.WithEndpoints(
					endpoint.New("swagger-redirect-get", "swagger-redirect-get",
						endpoint.WithGetMethod(),
						endpoint.WithHandler(func(c *echo.Context) error {
							return c.Redirect(http.StatusMovedPermanently, redirectTarget)
						}),
					),
				),
			),
			route.New("swagger-ui", "swagger-ui", wildcard,
				route.WithEndpoints(
					endpoint.New("swagger-ui-get", "swagger-ui-get",
						endpoint.WithGetMethod(),
						endpoint.WithHandler(handler),
					),
				),
			),
		),
	)
}

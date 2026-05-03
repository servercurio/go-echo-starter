package openapi

import (
	"strings"

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
// Swaggo's wrapper requires a wildcard route to serve UI assets; the route
// path is constructed as "<opts.Path>/*". The base path (e.g. "/swagger")
// without trailing slash redirects to the index page automatically because
// echo-swagger's handler short-circuits unknown sub-paths to its
// index.html.
func SwaggerModule(opts SwaggerOptions) router.Module {
	cleanPath := "/" + strings.Trim(opts.Path, "/")
	wildcard := cleanPath + "/*"

	handler := echoSwagger.EchoWrapHandlerV3(echoSwagger.URL(opts.SpecURL))

	return module.New("swagger", "swagger", "",
		module.WithRoutes(
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

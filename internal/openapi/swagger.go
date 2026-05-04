package openapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/servercurio/go-echo-starter/internal/api/std/endpoint"
	"github.com/servercurio/go-echo-starter/internal/api/std/module"
	"github.com/servercurio/go-echo-starter/internal/api/std/route"
	"github.com/servercurio/go-echo-starter/internal/router"

	echoSwagger "github.com/swaggo/echo-swagger/v2"
)

// swaggerIndexTemplate is a hand-rolled Swagger UI bootstrap. It uses
// StandaloneLayout so the topbar is rendered (we want a branded header),
// then injects CSS to: (a) hide the spec URL/picker controls — those are
// the `.download-url-wrapper` block in the topbar, the only purpose of
// which would be switching between multiple specs; and (b) replace the
// default Swagger logo image with the Server Curio brandmark served from
// the same prefix as a static asset. The other UI assets
// (swagger-ui-bundle.js, swagger-ui.css, etc.) continue to be served by
// echo-swagger's wildcard handler under the same prefix.
const swaggerIndexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css">
  <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32">
  <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .swagger-ui .topbar { background-color: #ffffff; border-bottom: 1px solid #e6e6e6; padding: 8px 0; }
    .swagger-ui .topbar .download-url-wrapper { display: none !important; }
    .swagger-ui .topbar-wrapper a { display: inline-flex; align-items: center; }
    .swagger-ui .topbar-wrapper a img,
    .swagger-ui .topbar-wrapper a svg { display: none !important; }
    .swagger-ui .topbar-wrapper a::before {
      content: "";
      display: inline-block;
      width: 240px;
      height: 60px;
      background: url("./logo.svg") no-repeat left center;
      background-size: contain;
    }
  </style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="./swagger-ui-bundle.js"></script>
<script src="./swagger-ui-standalone-preset.js"></script>
<script>
window.onload = function() {
  SwaggerUIBundle({
    url: %q,
    dom_id: "#swagger-ui",
    deepLinking: true,
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    plugins: [SwaggerUIBundle.plugins.DownloadUrl],
    layout: "StandaloneLayout"
  });
};
</script>
</body>
</html>
`

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
// Four routes are registered:
//
//   - <opts.Path>             → 301 redirect to <opts.Path>/index.html. Echo's
//     wildcard `/swagger/*` only matches paths starting with `/swagger/`,
//     not the bare `/swagger`, so without this users typing the prefix
//     into a browser would 404. We redirect directly to /index.html (not
//     to /<Path>/) to skip the intermediate redirect that echo-swagger's
//     own handler would otherwise emit, saving a round trip.
//
//   - <opts.Path>/index.html  → custom UI bootstrap (swaggerIndexTemplate)
//     that renders Swagger UI with StandaloneLayout but injects CSS to
//     hide the picker controls and replace the default logo with our
//     branded mark. Echo's radix router prefers static segments over
//     wildcards, so this beats the `/<Path>/*` route below for this exact
//     path.
//
//   - <opts.Path>/logo.svg    → embedded Server Curio brandmark, referenced
//     by the index page's CSS. Likewise beats the wildcard.
//
//   - <opts.Path>/*           → echo-swagger handler. Serves the Swagger UI
//     JS/CSS bundles and other supporting assets.
func SwaggerModule(opts SwaggerOptions) router.Module {
	cleanPath := "/" + strings.Trim(opts.Path, "/")
	wildcard := cleanPath + "/*"
	indexPath := cleanPath + "/index.html"
	logoPath := cleanPath + "/logo.svg"
	redirectTarget := indexPath

	// echoSwagger.URL is an APPEND operation, not a SET — it tacks our spec
	// URL onto the default ["doc.json", "doc.yaml"] list. The picker is
	// also driven off that list. We render index.html ourselves (see the
	// route below), so this handler only ever services the supporting JS
	// and CSS assets; the URL config here is belt-and-braces in case a
	// caller hits a non-overridden path that re-renders the bundled HTML.
	specURL := opts.SpecURL
	assetHandler := echoSwagger.EchoWrapHandlerV3(func(c *echoSwagger.Config) {
		c.URLs = []string{specURL}
	})

	indexBody := fmt.Sprintf(swaggerIndexTemplate, specURL)

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
			route.New("swagger-index", "swagger-index", indexPath,
				route.WithEndpoints(
					endpoint.New("swagger-index-get", "swagger-index-get",
						endpoint.WithGetMethod(),
						endpoint.WithHandler(func(c *echo.Context) error {
							return c.HTML(http.StatusOK, indexBody)
						}),
					),
				),
			),
			route.New("swagger-logo", "swagger-logo", logoPath,
				route.WithEndpoints(
					endpoint.New("swagger-logo-get", "swagger-logo-get",
						endpoint.WithGetMethod(),
						endpoint.WithHandler(func(c *echo.Context) error {
							return c.Blob(http.StatusOK, "image/svg+xml", logoSVG)
						}),
					),
				),
			),
			route.New("swagger-ui", "swagger-ui", wildcard,
				route.WithEndpoints(
					endpoint.New("swagger-ui-get", "swagger-ui-get",
						endpoint.WithGetMethod(),
						endpoint.WithHandler(assetHandler),
					),
				),
			),
		),
	)
}

package openapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	asrt "github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/servercurio/go-echo-starter/internal/api/std/endpoint"
	"github.com/servercurio/go-echo-starter/internal/api/std/module"
	"github.com/servercurio/go-echo-starter/internal/api/std/route"
	"github.com/servercurio/go-echo-starter/internal/openapi"
	"github.com/servercurio/go-echo-starter/internal/router"
)

// fixtureModule builds an api → v1 → (livez, readyz, users/:id) tree so
// the spec generator has multiple methods, a path parameter, and a
// sub-module to traverse — i.e. the same shape as the real default module
// tree, plus a path-param case the live tree doesn't currently exercise.
func fixtureModule() router.Module {
	v1 := module.New("v1", "v1", "v1",
		module.WithRoutes(
			route.New("liveness", "liveness", "/livez",
				route.WithEndpoints(
					endpoint.New("liveness-get", "liveness-get",
						endpoint.WithGetMethod(),
						endpoint.WithHandler(func(c *echo.Context) error { return nil }),
					),
				),
			),
			route.New("user", "user", "/users/:id",
				route.WithEndpoints(
					endpoint.New("user-get", "user-get",
						endpoint.WithGetMethod(),
						endpoint.WithHandler(func(c *echo.Context) error { return nil }),
					),
					endpoint.New("user-delete", "user-delete",
						endpoint.WithDeleteMethod(),
						endpoint.WithHandler(func(c *echo.Context) error { return nil }),
					),
				),
			),
		),
	)
	return module.New("api", "api", "api",
		module.WithSubModules(v1),
	)
}

func TestBuild_PathsAndOperations(t *testing.T) {
	assert := asrt.New(t)

	spec := openapi.Build(
		openapi.Info{Title: "fixture", Version: "0.0.0"},
		nil,
		[]router.Module{fixtureModule()},
	)

	assert.Equal("3.0.3", spec.OpenAPI)
	assert.Equal("fixture", spec.Info.Title)

	livez, ok := spec.Paths["/api/v1/livez"]
	assert.True(ok, "livez should exist under /api/v1/")
	if assert.NotNil(livez.Get) {
		assert.Equal("liveness-get", livez.Get.OperationID)
		assert.Contains(livez.Get.Tags, "v1")
		assert.Contains(livez.Get.Responses, "200")
	}
	assert.Nil(livez.Post, "no POST endpoint declared on /livez")

	users, ok := spec.Paths["/api/v1/users/{id}"]
	assert.True(ok, "Echo's :id should be rewritten to {id}")
	if assert.NotNil(users.Get) {
		assert.Len(users.Get.Parameters, 1)
		assert.Equal("id", users.Get.Parameters[0].Name)
		assert.Equal("path", users.Get.Parameters[0].In)
		assert.True(users.Get.Parameters[0].Required)
	}
	assert.NotNil(users.Delete, "DELETE method should also be wired up")
}

func TestBuild_TagsCollectedFromModuleTree(t *testing.T) {
	assert := asrt.New(t)

	spec := openapi.Build(openapi.Info{Title: "t", Version: "v"}, nil, []router.Module{fixtureModule()})

	tagNames := make([]string, 0, len(spec.Tags))
	for _, tag := range spec.Tags {
		tagNames = append(tagNames, tag.Name)
	}
	assert.ElementsMatch([]string{"api", "v1"}, tagNames, "every traversed module should produce one tag")
}

func TestBuild_EmptyModuleListYieldsValidSpec(t *testing.T) {
	// Edge case: a daemon with no user modules registered should still
	// produce a syntactically-valid spec (no nil paths map, no panics).
	assert := asrt.New(t)

	spec := openapi.Build(openapi.Info{Title: "empty", Version: "0"}, nil, nil)
	assert.NotNil(spec.Paths)
	assert.Empty(spec.Paths)
}

func TestBuild_OutputIsBothValidYAMLAndJSON(t *testing.T) {
	// Pin the marshaling contract — both formats must round-trip the spec
	// without error so the openapi.Module handlers can serve precomputed
	// bytes safely.
	assert := asrt.New(t)

	spec := openapi.Build(openapi.Info{Title: "fmt", Version: "1"}, nil, []router.Module{fixtureModule()})

	yamlBytes, err := yaml.Marshal(spec)
	assert.NoError(err)
	assert.NotEmpty(yamlBytes)

	jsonBytes, err := json.Marshal(spec)
	assert.NoError(err)
	assert.NotEmpty(jsonBytes)

	// Sanity-check round-trip: re-parse YAML and confirm a known path is
	// still present.
	var roundTrip map[string]any
	assert.NoError(yaml.Unmarshal(yamlBytes, &roundTrip))
	paths, _ := roundTrip["paths"].(map[string]any)
	assert.Contains(paths, "/api/v1/livez")
}

// TestModule_ServesPrecomputedBytes proves the openapi.Module factory
// actually wires the byte slices into the route — not the spec struct, not
// a per-request marshal — so changes to the slice after registration would
// not leak to clients.
func TestModule_ServesPrecomputedBytes(t *testing.T) {
	assert := asrt.New(t)

	yamlBody := []byte("openapi: 3.0.3\ninfo:\n  title: cached\n")
	jsonBody := []byte(`{"openapi":"3.0.3","info":{"title":"cached"}}`)

	mod := openapi.Module(yamlBody, jsonBody)

	e := echo.New()
	for _, r := range mod.Routes() {
		r.AttachGroup(e.Group(""))
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))
	assert.Equal(http.StatusOK, rec.Code)
	assert.Contains(rec.Header().Get(echo.HeaderContentType), "application/yaml")
	assert.Equal(string(yamlBody), rec.Body.String())

	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))
	assert.Equal(http.StatusOK, rec.Code)
	assert.Contains(rec.Header().Get(echo.HeaderContentType), "application/json")
	assert.Equal(string(jsonBody), rec.Body.String())
}

// Package openapi generates an OpenAPI 3.0.3 document directly from the
// router.Module / router.Route / router.Endpoint hierarchy this starter
// uses, and exposes the document at /openapi.yaml and /openapi.json. It can
// optionally serve a Swagger UI on top of that spec via Swaggo's
// echo-swagger v2 (see swagger.go).
//
// The generator is deliberately metadata-light: it knows the path, HTTP
// method, and module/tag — but doesn't infer request/response schemas
// because the route abstraction doesn't (yet) carry that information. Add
// richer metadata to router.Route / router.Endpoint and extend Build when
// you need response schemas, request bodies, or example payloads.
package openapi

import (
	"strings"

	"github.com/servercurio/go-echo-starter/internal/router"
)

// Spec is the top-level OpenAPI 3.0.3 document.
type Spec struct {
	OpenAPI string              `yaml:"openapi" json:"openapi"`
	Info    Info                `yaml:"info" json:"info"`
	Servers []Server            `yaml:"servers,omitempty" json:"servers,omitempty"`
	Paths   map[string]PathItem `yaml:"paths" json:"paths"`
	Tags    []Tag               `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// Info populates the OpenAPI `info` block.
type Info struct {
	Title       string `yaml:"title" json:"title"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// Server is one entry in the OpenAPI `servers` array. URL is required;
// Description is optional.
type Server struct {
	URL         string `yaml:"url" json:"url"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// Tag groups operations in the spec; Swagger UI shows one section per tag.
type Tag struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// PathItem holds the per-method operations for one URL template. Unset
// methods stay nil and are omitted from the marshaled output.
type PathItem struct {
	Get     *Operation `yaml:"get,omitempty" json:"get,omitempty"`
	Post    *Operation `yaml:"post,omitempty" json:"post,omitempty"`
	Put     *Operation `yaml:"put,omitempty" json:"put,omitempty"`
	Delete  *Operation `yaml:"delete,omitempty" json:"delete,omitempty"`
	Patch   *Operation `yaml:"patch,omitempty" json:"patch,omitempty"`
	Options *Operation `yaml:"options,omitempty" json:"options,omitempty"`
	Head    *Operation `yaml:"head,omitempty" json:"head,omitempty"`
}

// Operation is a single endpoint description: one method on one path.
type Operation struct {
	OperationID string              `yaml:"operationId,omitempty" json:"operationId,omitempty"`
	Summary     string              `yaml:"summary,omitempty" json:"summary,omitempty"`
	Tags        []string            `yaml:"tags,omitempty" json:"tags,omitempty"`
	Parameters  []Parameter         `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Responses   map[string]Response `yaml:"responses" json:"responses"`
}

// Parameter is currently used only for path parameters (in: "path").
// Extend with `in: "query"` / `in: "header"` once the route abstraction
// surfaces query/header metadata.
type Parameter struct {
	Name     string `yaml:"name" json:"name"`
	In       string `yaml:"in" json:"in"`
	Required bool   `yaml:"required" json:"required"`
	Schema   Schema `yaml:"schema" json:"schema"`
}

// Schema is intentionally minimal — only `type` is populated. Path
// parameters are typed as `string` because that's all Echo guarantees;
// richer typing belongs in route metadata.
type Schema struct {
	Type string `yaml:"type" json:"type"`
}

// Response is the per-status-code response descriptor. The starter declares
// only the success response (HTTP 200) per operation; routes that need
// richer error contracts can post-process the spec after Build returns.
type Response struct {
	Description string `yaml:"description" json:"description"`
}

// Build walks every Module → Route → Endpoint reachable from the supplied
// modules and assembles a Spec. The result is deterministic given the same
// input (map iteration during marshal is sorted by yaml.v3 / encoding/json).
func Build(info Info, servers []Server, modules []router.Module) *Spec {
	spec := &Spec{
		OpenAPI: "3.0.3",
		Info:    info,
		Servers: servers,
		Paths:   map[string]PathItem{},
	}

	tagSet := map[string]struct{}{}
	walk(modules, "", spec.Paths, tagSet)

	for name := range tagSet {
		spec.Tags = append(spec.Tags, Tag{Name: name})
	}
	sortTags(spec.Tags)

	return spec
}

// walk is the recursive driver behind Build. It accumulates paths into the
// supplied map and tag names into the supplied set; both are mutated in
// place.
func walk(mods []router.Module, parentPrefix string, paths map[string]PathItem, tags map[string]struct{}) {
	for _, m := range mods {
		modulePrefix := joinPath(parentPrefix, m.Prefix())
		moduleTag := m.Name()

		// Only mark a tag as present if the module actually exposes
		// reachable operations (directly or via a sub-module). We add the
		// tag preemptively here and trim unused tags later if we end up
		// caring; for the current usage every registered module has at
		// least one route or sub-module so the simpler approach is fine.
		tags[moduleTag] = struct{}{}

		for _, r := range m.Routes() {
			fullPath := joinPath(modulePrefix, r.Path())
			specPath, params := convertPath(fullPath)

			for _, ep := range r.Endpoints() {
				for _, method := range ep.Methods() {
					op := &Operation{
						OperationID: ep.Id(),
						Summary:     r.Name(),
						Tags:        []string{moduleTag},
						Parameters:  params,
						Responses: map[string]Response{
							"200": {Description: "Successful response"},
						},
					}

					item := paths[specPath]
					setMethod(&item, method, op)
					paths[specPath] = item
				}
			}
		}

		walk(m.SubModules(), modulePrefix, paths, tags)
	}
}

// joinPath stitches a parent prefix to a child segment, ensuring exactly
// one leading slash and no doubled separators.
func joinPath(parent, child string) string {
	parent = strings.Trim(parent, "/")
	child = strings.Trim(child, "/")
	switch {
	case parent == "" && child == "":
		return "/"
	case parent == "":
		return "/" + child
	case child == "":
		return "/" + parent
	}
	return "/" + parent + "/" + child
}

// convertPath rewrites Echo's `:name` path-parameter syntax to OpenAPI's
// `{name}` template syntax and returns one Parameter per `:name` it finds.
// Wildcard segments (`*`) become `{wildcard}` parameters.
func convertPath(p string) (string, []Parameter) {
	parts := strings.Split(p, "/")
	var params []Parameter
	for i, part := range parts {
		switch {
		case strings.HasPrefix(part, ":"):
			name := strings.TrimPrefix(part, ":")
			parts[i] = "{" + name + "}"
			params = append(params, Parameter{
				Name:     name,
				In:       "path",
				Required: true,
				Schema:   Schema{Type: "string"},
			})
		case part == "*":
			parts[i] = "{wildcard}"
			params = append(params, Parameter{
				Name:     "wildcard",
				In:       "path",
				Required: true,
				Schema:   Schema{Type: "string"},
			})
		}
	}
	return strings.Join(parts, "/"), params
}

// setMethod assigns op to the right slot on item based on the HTTP method
// string. Unknown methods are silently dropped — extend the switch when you
// add support for a non-standard verb.
func setMethod(item *PathItem, method string, op *Operation) {
	switch strings.ToUpper(method) {
	case "GET":
		item.Get = op
	case "POST":
		item.Post = op
	case "PUT":
		item.Put = op
	case "DELETE":
		item.Delete = op
	case "PATCH":
		item.Patch = op
	case "OPTIONS":
		item.Options = op
	case "HEAD":
		item.Head = op
	}
}

// sortTags is a stable, allocation-free in-place sort. We keep it small and
// dependency-free rather than reaching for sort.Slice for a list that's
// almost always tiny (handful of modules).
func sortTags(tags []Tag) {
	for i := 1; i < len(tags); i++ {
		for j := i; j > 0 && tags[j-1].Name > tags[j].Name; j-- {
			tags[j-1], tags[j] = tags[j], tags[j-1]
		}
	}
}

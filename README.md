<p align="center">
  <img src="docs/logo.svg" alt="Server Curio — Project Templates" width="600">
</p>

# go-echo-starter

A production-ready HTTP server starter template built on [Echo v5](https://github.com/labstack/echo). It provides a modular foundation for Go microservices with structured logging, flexible TLS, and reverse-proxy awareness — without imposing choices about persistence, auth, or business logic.

The compiled binary is named `appsvrd` (application server daemon).

## Table of contents

- [Features](#features)
- [Requirements](#requirements)
- [Quick start](#quick-start)
- [Project layout](#project-layout)
- [Configuration](#configuration)
  - [Config files](#config-files)
  - [Environment variables](#environment-variables)
- [Built-in endpoints](#built-in-endpoints)
  - [Content negotiation](#content-negotiation)
  - [Adding a custom component](#adding-a-custom-component)
- [API specification (OpenAPI 3.0)](#api-specification-openapi-30)
  - [Declaring request and response schemas](#declaring-request-and-response-schemas)
  - [Optional: Swagger UI](#optional-swagger-ui)
- [Database (optional)](#database-optional)
- [Adding a route](#adding-a-route)
- [Build tasks](#build-tasks)
- [Rebranding the starter](#rebranding-the-starter)
- [Container](#container)
- [Releases](#releases)
  - [Verifying a release](#verifying-a-release)
  - [Required repository secrets](#required-repository-secrets)
- [License](#license)

## Features

- **Echo v5** HTTP framework with a curated middleware stack (Recover, RequestID, Gzip, structured access log, CORS, security headers)
- **Dual-server topology** — separate HTTP and HTTPS servers, optional HTTP→HTTPS redirect
- **Three TLS modes** — static certificate files, ephemeral self-signed (ECDSA P-384), or Let's Encrypt via ACME `autocert`
- **Reverse-proxy aware** — configurable trust for direct IP, `X-Real-IP`, and `X-Forwarded-For` with CIDR allowlists
- **Modular routing** — composable `Module → Route → Endpoint` hierarchy with per-module middleware and prefix nesting
- **Structured logging** via [zerolog](https://github.com/rs/zerolog) with two independent loggers: `Daemon` (lifecycle) and `Access` (HTTP)
- **Layered configuration** — YAML/JSON config files plus environment variable overrides (prefix `APP_`)
- **Graceful shutdown** on `SIGINT`, `SIGTERM`, and `SIGUSR1` (Unix) / `SIGINT` (Windows), with configurable timeout
- **Cross-platform builds** for `linux`, `darwin`, `windows` × `amd64`, `arm64`

## Requirements

- Go 1.26+
- [Task](https://taskfile.dev) (for the build/dev workflow)
- Docker (optional, for container builds)

## Quick start

```sh
# Vendor, lint, and build all platform binaries
task

# Run the daemon locally (builds first)
task run:daemon

# Run all tests with race detector and coverage
task test
```

`task run:daemon` binds to `127.0.0.1:8888` (HTTP) and `127.0.0.1:4443` (HTTPS) via the dev overrides in `Taskfile.yaml`. The binary's own defaults — when run outside the Taskfile — are `:8080` (HTTP) and `:8443` (HTTPS).

## Project layout

```
cmd/daemon/         # main package — bootstrap and lifecycle wiring
internal/
  application/      # Application, server setup, TLS, proxy, top-level config
  api/              # Route module hierarchy (api → v1) + std implementations
  router/           # Module/Route/Endpoint interfaces
  config/           # YAML/JSON file loader
  env/              # Type-safe environment variable parsers
  logging/          # Zerolog setup, middleware, named loggers
  errors/           # errorx namespaces (FileSystemErrors)
  health/           # Per-component health Registry, Report model, Accept-header renderer
  openapi/          # OpenAPI 3.0 spec generator + /openapi.{yaml,json} module + optional Swagger UI module
  database/         # Optional SQL connection pool (pgx), Goose migrations, Bun ORM
    migrations/sql/ # Embedded *.sql migration files
    orm/            # Bun ORM singleton + (your) domain models
  version/          # Build-time version metadata
.github/
  workflows/        # CI: PR Formatting, PR Checks, CodeQL Scanning, Deploy Release, reusable callees
    docs/           # Workflow naming-standards reference
.releaserc.json     # semantic-release configuration (consumed by Deploy Release)
Taskfile.yaml       # Build, lint, test, run, container tasks
Dockerfile          # Multi-arch container image (consumes bin/)
```

## Configuration

Configuration is loaded in order: built-in defaults → config file → environment variables.

### Config files

Looked up under standard search paths (e.g. `./`, `~/.appsvr/`, `/etc/appsvr/`) using YAML or JSON.

### Environment variables

All keys are prefixed with `APP_`. Examples:

| Variable                              | Default      | Purpose                                     |
| ------------------------------------- | ------------ | ------------------------------------------- |
| `APP_DAEMON_LOG_LEVEL`                | `info`       | Daemon log level (`trace`–`fatal`)          |
| `APP_DAEMON_LOG_PRETTY_PRINT`         | `true`       | Color console output for daemon log         |
| `APP_DAEMON_LOG_INCLUDE_CALLER`       | `false`      | Include caller file:line in log output      |
| `APP_HTTP_ACCESS_LOG_ENABLED`         | `false`      | Toggle the HTTP access log                  |
| `APP_HTTP_ACCESS_LOG_LEVEL`           | `error`      | Access log level                            |
| `APP_HTTP_ACCESS_LOG_PRETTY_PRINT`    | `false`      | Color console output                        |
| `APP_SERVER_HTTP_PORT`                | `8080`       | HTTP listener port                          |
| `APP_SERVER_HTTP_SHUTDOWN_TIMEOUT`    | `10s`        | Graceful shutdown deadline (HTTP)           |
| `APP_SERVER_HTTP_READ_TIMEOUT`        | `30s`        | Max time to read entire request (HTTP)      |
| `APP_SERVER_HTTP_READ_HEADER_TIMEOUT` | `5s`         | Max time to read request headers (HTTP)     |
| `APP_SERVER_HTTP_WRITE_TIMEOUT`       | `30s`        | Max time to write response (HTTP)           |
| `APP_SERVER_HTTP_IDLE_TIMEOUT`        | `120s`       | Keep-alive idle timeout (HTTP)              |
| `APP_SERVER_HTTP_MAX_BODY_SIZE`       | `1MB`        | Max request body size; oversize → 413 (HTTP). Accepts `B`/`KB`/`MB`/`GB` suffixes or bare bytes. |
| `APP_SERVER_HTTPS_ENABLED`            | `false`      | Enable the TLS server                       |
| `APP_SERVER_HTTPS_PORT`               | `8443`       | HTTPS listener port                         |
| `APP_SERVER_HTTPS_HOSTNAME`           | —            | Hostname presented in self-signed/ACME cert |
| `APP_SERVER_HTTPS_USE_ACME_ISSUER`    | `false`      | Use Let's Encrypt instead of static certs   |
| `APP_SERVER_HTTPS_SHUTDOWN_TIMEOUT`   | `10s`        | Graceful shutdown deadline (HTTPS)          |
| `APP_SERVER_HTTPS_READ_TIMEOUT`       | `30s`        | Max time to read entire request (HTTPS)     |
| `APP_SERVER_HTTPS_READ_HEADER_TIMEOUT`| `5s`         | Max time to read request headers (HTTPS)    |
| `APP_SERVER_HTTPS_WRITE_TIMEOUT`      | `30s`        | Max time to write response (HTTPS)          |
| `APP_SERVER_HTTPS_IDLE_TIMEOUT`       | `120s`       | Keep-alive idle timeout (HTTPS)             |
| `APP_SERVER_HTTPS_MAX_BODY_SIZE`      | `1MB`        | Max request body size; oversize → 413 (HTTPS) |
| `APP_SERVER_CORS_ALLOW_ORIGINS`       | _(empty — CORS disabled)_ | Comma-separated list of origins (e.g. `https://app.example.com,https://admin.example.com`). Empty disables CORS entirely. |
| `APP_SERVER_CORS_ALLOW_METHODS`       | —            | Comma-separated method allowlist; defaults to GET/HEAD/PUT/PATCH/POST/DELETE if unset and CORS is enabled. |
| `APP_SERVER_CORS_ALLOW_HEADERS`       | —            | Comma-separated request-header allowlist                                                                  |
| `APP_SERVER_CORS_ALLOW_CREDENTIALS`   | `false`      | Allow cookies/auth headers in cross-origin requests. Cannot be combined with wildcard origins.            |
| `APP_SERVER_CORS_MAX_AGE`             | `0`          | Preflight cache lifetime in seconds                                                                       |
| `APP_SERVER_SECURITY_HSTS_MAX_AGE`            | `31536000`   | `Strict-Transport-Security` max-age (seconds). `0` suppresses the header. Emitted only on TLS requests (or `X-Forwarded-Proto: https`). |
| `APP_SERVER_SECURITY_HSTS_EXCLUDE_SUBDOMAINS` | `false`      | When `true`, drops the `includeSubDomains` directive from HSTS.                                                                         |
| `APP_SERVER_SECURITY_HSTS_PRELOAD_ENABLED`    | `false`      | When `true`, adds the `preload` directive. Preload submission is one-way — opt in deliberately.                                          |
| `APP_SERVER_SECURITY_CONTENT_SECURITY_POLICY` | —            | `Content-Security-Policy` header value (sent verbatim when non-empty). Empty by default to avoid breaking the bundled Swagger UI.        |
| `APP_SERVER_SECURITY_REFERRER_POLICY`         | —            | `Referrer-Policy` header value (sent verbatim when non-empty). Common values: `no-referrer`, `strict-origin-when-cross-origin`.          |
| `APP_DATABASE_DRIVER`                 | `pgx`        | `database/sql` driver name (PostgreSQL via pgx) |
| `APP_DATABASE_DSN`                    | —            | Connection string. **Empty disables the database subsystem entirely** (Connect/Migrate become no-ops, readiness probe ignores DB state). |
| `APP_DATABASE_MAX_OPEN_CONNS`         | `25`         | Max open connections in the pool. Zero/negative = unlimited.                                          |
| `APP_DATABASE_MAX_IDLE_CONNS`         | `5`          | Max idle connections retained in the pool                                                             |
| `APP_DATABASE_CONN_MAX_LIFETIME`      | `1h`         | Max lifetime of a connection before recycling                                                         |
| `APP_DATABASE_CONN_MAX_IDLE_TIME`     | `5m`         | Max time a connection may sit idle in the pool                                                        |
| `APP_OPENAPI_ENABLED`                 | `true`       | Serve generated `/openapi.yaml` and `/openapi.json`             |
| `APP_OPENAPI_TITLE`                   | `appsvrd`    | OpenAPI `info.title` (defaults to daemon name)                  |
| `APP_OPENAPI_VERSION`                 | _embedded SemVer_ | OpenAPI `info.version` (defaults to `internal/version.Number()`) |
| `APP_OPENAPI_DESCRIPTION`             | —            | OpenAPI `info.description`                                       |
| `APP_OPENAPI_SWAGGER_ENABLED`         | `false`      | Mount Swagger UI                                                 |
| `APP_OPENAPI_SWAGGER_PATH`            | `/swagger`   | URL prefix Swagger UI is mounted under                           |
| `APP_OPENAPI_SWAGGER_SPEC_URL`        | `/openapi.yaml` | URL Swagger UI fetches the spec from                          |

See `internal/application/config_*.go` for the complete schema.

## Built-in endpoints

The default `v1` module ships three Kubernetes-style health endpoints under `/api/v1/`. Response bodies follow the per-component model used by Spring Boot Actuator, Quarkus SmallRye Health, and Micronaut: an overall `status` plus a `components` map with per-subsystem state and optional `details`.

| Path                  | Purpose                  | Status code                                                                                | Components                                                                                       |
|-----------------------|--------------------------|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| `/api/v1/livez`       | Liveness probe           | Always `200`. Liveness must NOT depend on application state — kubelet uses it to decide whether to restart the pod. | `self` only.                                                                                     |
| `/api/v1/readyz`      | Readiness probe          | `200` when every registered component reports `UP`; `503` when any component reports `DOWN` (or when no registry is wired — fail closed). | All active subsystems: `lifecycle`, `http`, `https` (if enabled), `database` (if configured).     |
| `/api/v1/healthz`     | Legacy alias for readyz  | Same as `/readyz`.                                                                         | Same as `/readyz`. Kept so consumers that default to `/healthz` (older uptime checks, default cloud-LB health-check paths) keep working. |

Example `/readyz` response when everything is healthy and a database is configured:

```json
{
  "status": "UP",
  "components": {
    "lifecycle": { "status": "UP" },
    "http":      { "status": "UP", "details": { "port": 8080, "bindAddress": "", "hostname": "" } },
    "https":     { "status": "UP", "details": { "port": 8443, "bindAddress": "", "hostname": "", "autoCertIssuance": true, "useAcmeIssuer": false, "ephemeralCertIssuance": true } },
    "database":  { "status": "UP", "details": { "driver": "pgx" } }
  }
}
```

When the database is unreachable, the response code becomes `503` and the body's overall `status` flips to `DOWN` while the per-component breakdown shows which dependency failed.

### Content negotiation

Both response formats are supported via the `Accept` header:

| `Accept` header                                            | Response Content-Type                |
|------------------------------------------------------------|--------------------------------------|
| _(unset)_, `*/*`, `application/json`, anything not yaml    | `application/json; charset=utf-8`    |
| `application/yaml`, `application/x-yaml`, `text/yaml`, `*+yaml` | `application/yaml; charset=utf-8`    |

```sh
curl -s http://localhost:8080/api/v1/readyz                                # JSON
curl -s -H 'Accept: application/yaml' http://localhost:8080/api/v1/readyz   # YAML
```

### Adding a custom component

Components live on a `*health.Registry` owned by the `Application` (`app.HealthRegistry()`); the v1 handlers snapshot it on every request. To register a new check:

```go
app.HealthRegistry().Register("redis", func() health.ComponentResult {
    if err := redisClient.Ping(ctx).Err(); err != nil {
        return health.ComponentResult{
            Status:  health.StatusDown,
            Details: map[string]any{"reason": err.Error()},
        }
    }
    return health.ComponentResult{Status: health.StatusUp}
})
```

The check closure is invoked on every `/readyz` and `/healthz` request, so it should be cheap (sub-second). Stand a cached background probe in front of expensive checks.

## API specification (OpenAPI 3.0)

The starter generates an OpenAPI 3.0.3 document directly from whatever modules are registered with `Application` and serves it without any external tooling. No annotations, no `swag init`, no codegen step — the spec is built in-process during `Initialize()` and served as precomputed bytes.

| Path             | Format | Notes                                        |
|------------------|--------|----------------------------------------------|
| `/openapi.yaml`  | YAML   | `application/yaml; charset=utf-8`            |
| `/openapi.json`  | JSON   | `application/json; charset=utf-8`            |

Both endpoints are mounted unconditionally when `APP_OPENAPI_ENABLED=true` (the default). Set `APP_OPENAPI_ENABLED=false` to suppress them entirely (e.g. for production deployments that don't want to advertise their API surface).

The generator walks the `Module → Route → Endpoint` tree and emits one operation per `(path, method)` pair. Echo's `:name` path parameters are rewritten to OpenAPI's `{name}` template syntax, and a `Parameter` of `in: path, type: string` is emitted for each. Each module's `Name()` becomes a `tag` so Swagger UI groups operations by module.

### Declaring request and response schemas

Endpoints opt into richer documentation via builder options on `endpoint`. The reflection-based schema generator at `internal/openapi/schema.go` converts the Go types you declare into JSON Schemas, registers each named struct under `components.schemas`, and operations reference it via `$ref` so each shape appears once in the spec.

```go
endpoint.New("user-create-post", "user-create-post",
    endpoint.WithPostMethod(),
    endpoint.WithSummary("Create a user"),
    endpoint.WithDescription("Persists a new user record."),
    endpoint.WithRequest(CreateUserRequest{}),                       // application/json body
    endpoint.WithResponse(http.StatusCreated, User{}, "Created"),     // 201 with User schema
    endpoint.WithResponse(http.StatusBadRequest, ErrorResponse{}, "Validation failed"),
    endpoint.WithHandler(createUser),
)
```

What the generator handles:

| Go construct                       | OpenAPI representation                              |
|------------------------------------|-----------------------------------------------------|
| `string`, `bool`, ints, floats     | inline scalar with `type` + `format` set            |
| `time.Time`                        | `type: string, format: date-time`                   |
| `[]byte`                           | `type: string, format: byte` (base64)               |
| Other slices                       | `type: array, items: <inner>`                       |
| `map[string]T`                     | `type: object, additionalProperties: <T>`           |
| `*T`                               | schema for `T` with `nullable: true` (or omitted from `required`) |
| Named struct                       | registered under `components.schemas`, referenced via `$ref` |
| Anonymous struct                   | inlined                                             |
| Embedded struct                    | fields lifted into the parent schema (matches `encoding/json`) |
| `any` / `interface{}`              | free-form schema (no `type` constraint)             |
| Recursive struct (`A` → `*A`)      | `$ref` back to the same schema (no infinite recursion) |

JSON struct tags are honored: `json:"name"` renames the property, `omitempty` excludes the field from `required`, and `json:"-"` skips it entirely. Unexported fields are skipped, but embedded struct fields are lifted regardless of the embedded field's name visibility (matches `encoding/json` flattening).

Endpoints without any `WithResponse` calls fall back to a generic `200: Successful response` so the spec stays valid; once you declare any response, the generator uses your declarations exclusively (no auto-200 is appended).

### Optional: Swagger UI

Set `APP_OPENAPI_SWAGGER_ENABLED=true` to mount [Swaggo's echo-swagger v2](https://github.com/swaggo/echo-swagger) at `/swagger/`, pointed at the generated spec by default:

```sh
APP_OPENAPI_SWAGGER_ENABLED=true task run:daemon
# then open: http://localhost:8080/swagger/index.html
```

The dependency is always compiled in (~1MB binary overhead) but no routes are mounted unless `Swagger.Enabled` is true. Override `APP_OPENAPI_SWAGGER_PATH` to mount under a different prefix, or `APP_OPENAPI_SWAGGER_SPEC_URL` to point the UI at an externally hosted spec.

## Database (optional)

The starter ships an opt-in database layer using:

- [`pgx`](https://github.com/jackc/pgx) as the `database/sql` driver (PostgreSQL).
- [`pressly/goose`](https://github.com/pressly/goose) for SQL schema migrations, embedded into the binary via `//go:embed internal/database/migrations/sql/*.sql`.
- [`uptrace/bun`](https://github.com/uptrace/bun) as the ORM, configured against the same `*sql.DB` connection pool.

To enable: set `APP_DATABASE_DSN` (e.g. `postgres://user:pass@host:5432/db?sslmode=disable`). On startup the daemon will:

1. `database.Connect(cfg)` — open the pool and verify reachability with a ping.
2. `database.Migrate(cfg)` — apply pending Goose migrations from `internal/database/migrations/sql/`.
3. `orm.Configure()` — wrap the connection in a Bun `*bun.DB` singleton accessible via `orm.Database()`.

To add a migration, drop a new `YYYYMMDDHHMMSS_description.sql` file alongside the no-op initial migration and rebuild — the embed pattern picks it up automatically. The default `Driver` is `pgx`; replace the driver, the dialect in `internal/database/orm/connection.go`, and the `goose.SetDialect` argument in `migration.go` to swap engines.

## Adding a route

Routes are attached to modules. The `v1` API module already wires the three health routes above; new endpoints go under `internal/api/v1/` and are added to the module via `WithRoutes(...)`.

The relevant constructors are:

- `module.New(id, name, prefix string, opts ...module.Option)` — options: `WithRoutes`, `WithSubModules`, `WithMiddleware`
- `route.New(id, name, path string, opts ...route.Option)` — options: `WithEndpoints`, `WithMiddleware`
- `endpoint.New(id, name string, opts ...endpoint.Option)` — options: `WithHandler`, `WithMethods` (or convenience `WithGetMethod`, `WithPostMethod`, etc.), `WithMiddleware`

Echo v5 handlers receive `*echo.Context` (pointer), not the v4 interface.

A typical pattern:

```go
// internal/api/v1/ping.go
package v1

import (
    "net/http"

    "github.com/labstack/echo/v5"
    "github.com/servercurio/go-echo-starter/internal/api/std/endpoint"
    "github.com/servercurio/go-echo-starter/internal/api/std/route"
    "github.com/servercurio/go-echo-starter/internal/router"
)

func PingRoute() router.Route {
    return route.New("ping", "ping", "/ping",
        route.WithEndpoints(
            endpoint.New("ping-get", "ping-get",
                endpoint.WithGetMethod(),
                endpoint.WithHandler(func(c *echo.Context) error {
                    return c.JSON(http.StatusOK, map[string]string{"pong": "ok"})
                }),
            ),
        ),
    )
}
```

Then wire it into `internal/api/v1/module.go` by adding `PingRoute()` to the existing `module.WithRoutes(...)` call. Routes that need application state (e.g. readiness, config) take a `*router.Config` argument like `ReadinessRoute(cfg)` does.

## Build tasks

| Task                  | Description                                         |
| --------------------- | --------------------------------------------------- |
| `task` / `task default` | Clean → vendor → lint → build all platform binaries     |
| `task build`          | Cross-compile for all OS/arch combinations (calls `generate`) |
| `task generate`       | `go generate ./...` — refresh `internal/version/commit.txt` |
| `task hash`           | Write a `bin/<binary>.sha256` file per binary           |
| `task sign`           | GPG-sign each binary and each `.sha256` file (writes `<binary>.asc` and `<binary>.sha256.asc`); depends on `hash` |
| `task vendor`         | `go mod tidy` + `go mod vendor`                         |
| `task lint`           | `go fmt` + `go vet` with strict checks                  |
| `task test`           | `go test -race -cover -coverprofile cover.out -parallel 4 -v ./...` |
| `task run:daemon`     | Build and run the local-platform binary                 |
| `task build:container`| Build the Docker image                                  |
| `task run:container`  | Build and run the Docker image                          |
| `task clean`          | Remove `bin/`, `dist/`, and coverage output             |

## Rebranding the starter

The default identifiers (`appsvrd`, `APP_*` env-var prefix, `appsvr` config-path element) are intentional defaults you'll want to replace per project. There are five places to touch — keep them consistent or you'll get half-renamed binaries that look in the wrong config paths.

In the examples below, assume you're renaming to **`myapi`** with binary **`myapid`**, env-var prefix **`MYAPI`**, and config-path element **`myapi`**.

### 1. `Taskfile.yaml` — drives the binary name and Docker tag

```yaml
vars:
  PROJECT_NAME: MyApi          # was: AppSvr → produces "myapid" binary (lowercase + 'd' suffix)
```

`PROJECT_NAME` is templated as `{{.PROJECT_NAME | lower}}d` to produce the binary name (`bin/myapid-${GOOS}-${GOARCH}`) and the Docker image tag (`myapi:latest`). It must align with the Go-side `defaultName` constant — the binary won't be able to identify itself in logs otherwise.

### 2. `Taskfile.yaml` env block — rename every `APP_*` to your prefix

```yaml
env:
  CGO_ENABLED: 0
  MYAPI_DAEMON_LOG_LEVEL: trace
  MYAPI_DAEMON_LOG_INCLUDE_CALLER: false
  MYAPI_HTTP_ACCESS_LOG_ENABLED: true
  MYAPI_HTTP_ACCESS_LOG_LEVEL: info
  MYAPI_HTTP_ACCESS_LOG_PRETTY_PRINT: true
  MYAPI_SERVER_HTTPS_ENABLED: true
  MYAPI_SERVER_HTTP_PORT: 80
  MYAPI_SERVER_HTTPS_PORT: 443
  MYAPI_SERVER_HTTPS_USE_ACME_ISSUER: false
```

These developer-convenience overrides only take effect when `task` runs — but they must use the new prefix or they'll be silently ignored after step 3.

### 3. `internal/application/application.go` — the four authoritative constants

```go
const defaultName = "myapid"               // binary identity in logs
const defaultEnvPrefix = "MYAPI"           // env-var prefix (no trailing underscore)
const defaultConfigName = "myapid"         // base filename for config files (myapid.yaml, myapid.json)
const defaultConfigPathElement = "myapi"   // directory segment in /etc/<x>/, ~/.config/<x>/
```

`defaultEnvPrefix` is the source of truth — it's what `env.AddPrefix` uses to build keys like `MYAPI_DAEMON_LOG_LEVEL`. The Taskfile env values from step 2 must use this exact prefix (no underscore at the end of the constant; the joiner adds it).

`defaultConfigPathElement` controls the config-file search paths in `internal/application/utils.go`:

- `/etc/myapi/`
- `~/.config/myapi/`
- `./` (current directory)

`defaultConfigName` controls the filename — files are looked up as `myapid.yaml`, `myapid.yml`, or `myapid.json` in each search path.

### 4. `Dockerfile` — paths and binary name are hardcoded

```dockerfile
FROM ubuntu:noble-20250127

COPY ./bin/ /tmp/myapi/

RUN mkdir -p /tmp/myapi && \
    ls -lah /tmp/myapi && \
    ARCH="$(dpkg --print-architecture)" && \
    case "$ARCH" in \
        x86_64|amd64) ARCH="amd64" ;; \
        aarch64|arm64) ARCH="arm64" ;; \
        *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    cp -v /tmp/myapi/myapid-linux-${ARCH} /usr/local/bin/myapid && \
    chmod +x /usr/local/bin/myapid && \
    rm -rf /tmp/myapi

CMD ["myapid"]
```

Replace `appsvr` (path element) and `appsvrd` (binary name) on every line. The binary name in the `cp` source must match what `task build:daemon` produces (i.e. derived from `PROJECT_NAME`).

### 5. `cmd/daemon/main_test.go` — the test prefix override

```go
const envPrefix = "MYAPI"
```

This test sets env vars under a custom prefix to exercise the `env.AddPrefix` path. Update it to match your new `defaultEnvPrefix` so the test continues to actually exercise the configured prefix.

### Optional: Go module path

If you're forking under a new owner, also update the module path in `go.mod`:

```
module github.com/your-org/your-repo
```

Then run `task vendor` to refresh imports across the codebase. Imports that reference `github.com/servercurio/go-echo-starter/internal/...` will need a find-and-replace.

### Verify

```sh
task                # full pipeline rebuilds everything
task run:daemon     # confirm the renamed binary boots and logs use the new name
```

The startup log should show your new binary name and config search paths matching your new `defaultConfigPathElement`.

## Container

```sh
task build:container
task run:container
```

The image is built from a date-pinned `ubuntu:noble-*` tag (see `Dockerfile`) and ships a single static binary built with `CGO_ENABLED=0`.

## Releases

Releases are produced by [semantic-release](https://github.com/semantic-release/semantic-release) and triggered manually via the **Deploy Release** workflow (`100-user-deploy-release-artifact.yaml`) under the GitHub Actions tab. The release flow:

1. Analyses commits since the last tag using the [conventional-commits](https://www.conventionalcommits.org) preset and decides the next semver version.
2. Updates `internal/version/version.txt` to the new version and runs `task generate` to refresh `commit.txt`.
3. Commits the version bump back to the release branch with a `chore(release): X.Y.Z [skip ci]` message. The commit is **GPG-signed** (DCO `Signed-off-by` is appended automatically by a per-clone `prepare-commit-msg` hook installed in CI).
4. Runs `task build` to produce all six platform binaries, then `task sign` (which transitively runs `task hash`) to produce per-binary `.sha256` files, GPG signatures of the binaries (`.asc`), and GPG signatures of the `.sha256` files (`.sha256.asc`).
5. Tags the commit and creates a GitHub Release with the binaries, their `.sha256` files, and the corresponding `.asc` and `.sha256.asc` GPG signatures attached as assets.

### Verifying a release

Each platform ships four files: the binary, its `.sha256`, the binary's `.asc`, and the `.sha256.asc`.

```sh
# byte-correctness only
shasum -a 256 -c appsvrd-linux-amd64.sha256

# verify the .sha256 file itself was signed by the release key, then check the binary against it
gpg --verify appsvrd-linux-amd64.sha256.asc appsvrd-linux-amd64.sha256
shasum -a 256 -c appsvrd-linux-amd64.sha256

# OR verify the binary directly with its detached signature
gpg --verify appsvrd-linux-amd64.asc appsvrd-linux-amd64
```

The first GPG path lets you trust just the (small) `.sha256` file and then chain that trust to the binary; the second verifies the binary directly. Either is sufficient.

Branch policy (from `.releaserc.json`):

| Branch pattern   | Channel        | Notes                                                |
|------------------|----------------|------------------------------------------------------|
| `main`           | latest         | Default release branch                               |
| `release/X.Y`    | `X.Y.x`        | Maintenance branches; release range pinned to `X.Y.x` |
| `alpha/*`        | `alpha`        | Prerelease channel                                   |
| `beta/*`         | `beta`         | Prerelease channel                                   |
| `rc/*`           | `rc`           | Release-candidate channel                            |

Release rules (commit type → version bump):

| Type                | Bump  |
|---------------------|-------|
| `feat`              | minor |
| `fix`               | patch |
| `refactor`, `build` | patch |
| `BREAKING CHANGE` (footer or `!` in subject) | minor |
| `chore`, `ci`, `docs`, `style`, `test` | none  |

Run a dry run from the workflow dispatch UI by checking **Perform dry run** — semantic-release will print the version that *would* be released without tagging, committing, or publishing.

> **Note on protected branches**: the workflow uses the default `GITHUB_TOKEN`. If `main` is protected with restrictions that block the GitHub App from pushing back the version-bump commit, configure a PAT (e.g. `GH_ACCESS_TOKEN`) with bypass rights and pass it as `release-token` from `100-user-deploy-release-artifact.yaml`.

### Required repository secrets

The release workflow requires the following secrets configured under **Settings → Secrets and variables → Actions**:

| Secret             | Purpose                                                                  |
|--------------------|--------------------------------------------------------------------------|
| `GPG_PRIVATE_KEY`  | ASCII-armored private GPG key used to sign the release commit and the `SHA256SUMS` file. Generate with `gpg --armor --export-secret-keys <KEY-ID>`. |
| `GPG_PASSPHRASE`   | Passphrase for the private GPG key.                                      |
| `GH_ACCESS_TOKEN`  | *(Optional.)* PAT with bypass on protected branches. Only needed if `GITHUB_TOKEN` cannot push the version-bump commit. |
| `CODECOV_TOKEN`    | *(Optional, used by PR Checks.)* Codecov upload token; recommended for public repos to avoid rate limits, required for private repos. |

The corresponding **public key** must be added to the GitHub account/org used to publish releases (and shared with consumers) so signatures can be verified.

## License

See [LICENSE](LICENSE).

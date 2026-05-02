# go-echo-starter

A production-ready HTTP server starter template built on [Echo v5](https://github.com/labstack/echo). It provides a modular foundation for Go microservices with structured logging, flexible TLS, and reverse-proxy awareness — without imposing choices about persistence, auth, or business logic.

The compiled binary is named `appsvrd` (application server daemon).

## Features

- **Echo v5** HTTP framework with a curated middleware stack (Recover, RequestID, Gzip, structured access log, CORS, security headers)
- **Dual-server topology** — separate HTTP and HTTPS servers, optional HTTP→HTTPS redirect
- **Three TLS modes** — static certificate files, ephemeral self-signed (ECDSA P-384), or Let's Encrypt via ACME `autocert`
- **Reverse-proxy aware** — configurable trust for direct IP, `X-Real-IP`, and `X-Forwarded-For` with CIDR allowlists
- **Modular routing** — composable `Module → Route → Endpoint` hierarchy with per-module middleware and prefix nesting
- **Structured logging** via [zerolog](https://github.com/rs/zerolog) with two independent loggers: `Daemon` (lifecycle) and `Access` (HTTP)
- **Layered configuration** — YAML/JSON config files plus environment variable overrides (prefix `APP_`)
- **Graceful shutdown** on `SIGINT` / `SIGTERM` with configurable timeout
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

The daemon listens on `:8080` (HTTP) and `:8443` (HTTPS) by default.

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
  version/          # Build-time version metadata
pkg/                # (reserved for future public packages)
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
| `APP_DAEMON_LOG_INCLUDE_CALLER`       | `false`      | Include caller file:line in log output      |
| `APP_HTTP_ACCESS_LOG_ENABLED`         | `true`       | Toggle the HTTP access log                  |
| `APP_HTTP_ACCESS_LOG_LEVEL`           | `error`      | Access log level                            |
| `APP_HTTP_ACCESS_LOG_PRETTY_PRINT`    | `false`      | Color console output                        |
| `APP_SERVER_HTTP_PORT`                | `8080`       | HTTP listener port                          |
| `APP_SERVER_HTTPS_ENABLED`            | `false`      | Enable the TLS server                       |
| `APP_SERVER_HTTPS_PORT`               | `8443`       | HTTPS listener port                         |
| `APP_SERVER_HTTPS_HOSTNAME`           | —            | Hostname presented in self-signed/ACME cert |
| `APP_SERVER_HTTPS_USE_ACME_ISSUER`    | `false`      | Use Let's Encrypt instead of static certs   |

See `internal/application/config_*.go` for the complete schema.

## Adding a route

Routes are attached to modules. The `v1` API module is empty by default — add endpoints under `internal/api/v1/`, then register them via the module's `WithRoutes(...)` option.

A typical pattern:

```go
// internal/api/v1/health.go
func HealthRoute() router.Route {
    return route.NewStandard("health",
        route.WithPath("/health"),
        route.WithEndpoints(
            endpoint.NewStandard(
                endpoint.WithMethods(http.MethodGet),
                endpoint.WithHandler(func(c echo.Context) error {
                    return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
                }),
            ),
        ),
    )
}
```

Then wire it into `internal/api/v1/module.go` via `module.WithRoutes(HealthRoute())`.

## Build tasks

| Task                  | Description                                         |
| --------------------- | --------------------------------------------------- |
| `task` / `task default` | Clean → vendor → lint → build all platform binaries |
| `task build`          | Cross-compile for all OS/arch combinations          |
| `task vendor`         | `go mod tidy` + `go mod vendor`                     |
| `task lint`           | `go fmt` + `go vet` with strict checks              |
| `task test`           | `go test -race -cover -parallel 4 ./...`            |
| `task run:daemon`     | Build and run the local-platform binary             |
| `task build:container`| Build the Docker image                              |
| `task run:container`  | Build and run the Docker image                      |
| `task clean`          | Remove `bin/`, `dist/`, and coverage output         |

## Container

```sh
task build:container
task run:container
```

The image is built `FROM ubuntu:noble` and ships a single static binary built with `CGO_ENABLED=0`.

## License

See [LICENSE](LICENSE).

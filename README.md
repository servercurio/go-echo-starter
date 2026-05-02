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

The image is built `FROM ubuntu:noble` and ships a single static binary built with `CGO_ENABLED=0`.

## License

See [LICENSE](LICENSE).

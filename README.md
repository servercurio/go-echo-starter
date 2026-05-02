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
  database/         # Optional SQL connection pool (pgx), Goose migrations, Bun ORM
    migrations/sql/ # Embedded *.sql migration files
    orm/            # Bun ORM singleton + (your) domain models
  version/          # Build-time version metadata
pkg/                # (reserved for future public packages)
.github/
  workflows/        # CI: PR Formatting, PR Checks, Deploy Release, reusable callees
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
| `APP_SERVER_HTTPS_ENABLED`            | `false`      | Enable the TLS server                       |
| `APP_SERVER_HTTPS_PORT`               | `8443`       | HTTPS listener port                         |
| `APP_SERVER_HTTPS_HOSTNAME`           | —            | Hostname presented in self-signed/ACME cert |
| `APP_SERVER_HTTPS_USE_ACME_ISSUER`    | `false`      | Use Let's Encrypt instead of static certs   |
| `APP_DATABASE_DRIVER`                 | `pgx`        | `database/sql` driver name (PostgreSQL via pgx) |
| `APP_DATABASE_DSN`                    | —            | Connection string. **Empty disables the database subsystem entirely** (Connect/Migrate become no-ops, readiness probe ignores DB state). |

See `internal/application/config_*.go` for the complete schema.

## Built-in endpoints

The default `v1` module ships three Kubernetes-style health endpoints under `/api/v1/`:

| Path                  | Purpose                  | Behaviour                                                                       |
|-----------------------|--------------------------|---------------------------------------------------------------------------------|
| `/api/v1/livez`       | Liveness probe           | Always `200 {"status":"ok"}` while the HTTP listener can respond. Does **not** depend on application state, downstream services, or shutdown — kubelet uses this to decide whether to restart the pod. |
| `/api/v1/readyz`      | Readiness probe          | `200 {"status":"ok"}` once the application has finished starting up, the readiness probe reports ready, AND (if a database is configured) a `PingContext` against the connection pool succeeds. `503 {"status":"not_ready"}` during startup, after a shutdown signal is received, when the database is unreachable, or if the probe is misconfigured (fail closed). Load balancers should drain traffic when this returns 503. |
| `/api/v1/healthz`     | Legacy alias for readyz  | Same semantics as `/readyz`. Kept so consumers that default to `/healthz` (older uptime checks, default Cloud LB health-check paths) keep working.       |

Application readiness is wired via `router.Config.ReadinessProbe` — `cmd/daemon/main.go` composes it as `app.IsReady() && app.IsDatabaseHealthy()`. The lifecycle flag flips to `true` after the server goroutines spawn and back to `false` when a shutdown signal arrives. The database probe issues a 1-second `PingContext` against the connection pool on every `/readyz` and `/healthz` request, so the load balancer notices a downed database within one probe interval. When no database is configured, `IsDatabaseHealthy()` returns `true` unconditionally and the readiness check collapses back to lifecycle-only.

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
| `task test`           | `go test -race -cover -parallel 4 ./...`                |
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

The image is built `FROM ubuntu:noble` and ships a single static binary built with `CGO_ENABLED=0`.

## Releases

Releases are produced by [semantic-release](https://github.com/semantic-release/semantic-release) and triggered manually via the **Deploy Release** workflow (`100-flow-deploy-release-artifact.yaml`) under the GitHub Actions tab. The release flow:

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

> **Note on protected branches**: the workflow uses the default `GITHUB_TOKEN`. If `main` is protected with restrictions that block the GitHub App from pushing back the version-bump commit, configure a PAT (e.g. `GH_ACCESS_TOKEN`) with bypass rights and pass it as `release-token` from `100-flow-deploy-release-artifact.yaml`.

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

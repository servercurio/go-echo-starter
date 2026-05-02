# Module Structure

- `cmd/daemon/` — `main` package; bootstraps and wires the application lifecycle (`Configure → Initialize → Start`).
- `internal/application/` — `Application` type, dual HTTP/TLS Echo server setup, middleware stack, top-level config, ACME / static-cert / self-signed TLS, reverse-proxy IP extraction, and OS-signal handling for graceful shutdown (build-tagged via `signals_unix.go` / `signals_windows.go`).
- `internal/api/` — Route module hierarchy (`api → v1`) plus the `std/` concrete `Module` / `Route` / `Endpoint` implementations.
- `internal/router/` — `Module`, `Route`, `Endpoint` interfaces; the abstraction routes are attached through.
- `internal/config/` — YAML/JSON config-file loader with multi-path search.
- `internal/env/` — Type-safe environment-variable parsers (string, bool, int, uint16, float, duration) and the `APP_` prefix convention.
- `internal/logging/` — Two named zerolog loggers (`Daemon` for lifecycle, `Access` for HTTP) plus the Echo request-logger middleware.
- `internal/errors/` — `joomcode/errorx` namespaces (e.g. `FileSystemErrors`); the canonical pattern for new error categories.
- `internal/health/` — Per-component health-check registry plus the `Report` model returned by `/livez`/`/readyz`/`/healthz`. `health.go` defines `Status`, `ComponentResult`, `Report`, and `Registry` (Spring Boot Actuator-style aggregation: overall UP iff every component UP). `render.go` does Accept-header content negotiation between JSON and YAML.
- `internal/database/` — Optional SQL database subsystem. `config.go` carries `Driver` + `DSN` (empty DSN ⇒ disabled and the rest is a no-op), `connection.go` wraps a mutex-guarded `*sql.DB` singleton with `Connect`/`Disconnect`/`Connection`/`IsHealthy`, `migration.go` runs Goose migrations embedded from `migrations/sql/*.sql`. Default driver is `pgx` (PostgreSQL).
- `internal/database/orm/` — Bun ORM singleton (`Configure`/`Database`/`Reset`). Add domain models as sibling files; the starter ships none.
- `internal/version/` — Build-time version metadata (commit, semver, tag) using `Masterminds/semver`.
- `internal/obfusicate/` — Small string-redaction helper exposing `ConcealPrefix` for masking values in log output. Spelling is intentional. Treat as out of scope unless the task explicitly references redaction.
- `pkg/` — Intentionally empty; reserved for future public packages. Internal code goes in `internal/`.
- `.github/workflows/` — CI workflows. PR Formatting and PR Checks (200-series flow workflows) call the reusable code-compiles and unit-test workflows (800-series). Releases run via `100-flow-deploy-release-artifact.yaml` (manual `workflow_dispatch`) which delegates to the reusable `800-call-semantic-release.yaml`; semantic-release configuration lives in `.releaserc.json` at repo root. Naming follows the convention documented in `.github/workflows/docs/naming-standards.md`.

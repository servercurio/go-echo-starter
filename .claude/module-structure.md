# Module Structure

- `cmd/daemon/` — `main` package; bootstraps and wires the application lifecycle (`Configure → Initialize → Start`).
- `internal/application/` — `Application` type, dual HTTP/TLS Echo server setup, middleware stack, top-level config, ACME / static-cert / self-signed TLS, reverse-proxy IP extraction, and OS-signal handling for graceful shutdown (build-tagged via `signals_unix.go` / `signals_windows.go`).
- `internal/api/` — Route module hierarchy (`api → v1`) plus the `std/` concrete `Module` / `Route` / `Endpoint` implementations.
- `internal/router/` — `Module`, `Route`, `Endpoint` interfaces; the abstraction routes are attached through.
- `internal/config/` — YAML/JSON config-file loader with multi-path search.
- `internal/env/` — Type-safe environment-variable parsers (string, bool, int, uint16, float, duration) and the `APP_` prefix convention.
- `internal/logging/` — Two named zerolog loggers (`Daemon` for lifecycle, `Access` for HTTP) plus the Echo request-logger middleware.
- `internal/errors/` — `joomcode/errorx` namespaces (e.g. `FileSystemErrors`); the canonical pattern for new error categories.
- `internal/version/` — Build-time version metadata (commit, semver, tag) using `Masterminds/semver`.
- `internal/obfusicate/` — Small string-redaction helper exposing `ConcealPrefix` for masking values in log output. Spelling is intentional. Treat as out of scope unless the task explicitly references redaction.
- `pkg/` — Intentionally empty; reserved for future public packages. Internal code goes in `internal/`.
- `.github/workflows/` — CI workflows. PR Formatting and PR Checks (200-series flow workflows) call the reusable code-compiles and unit-test workflows (800-series). Releases run via `100-flow-deploy-release-artifact.yaml` (manual `workflow_dispatch`) which delegates to the reusable `800-call-semantic-release.yaml`; semantic-release configuration lives in `.releaserc.json` at repo root. Naming follows the convention documented in `.github/workflows/docs/naming-standards.md`.

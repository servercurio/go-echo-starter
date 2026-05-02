# Instructions for AI agents

## Tech Stack & Go Version

Go 1.26 is required. The agent should flag any code suggestions targeting older Go versions or using deprecated APIs (e.g. `ioutil.*`, pre-generics patterns where generics are clearer).
[Task](https://taskfile.dev) is the canonical build runner — never suggest invoking raw `go build` / `go test` for anything beyond a quick check; route work through `Taskfile.yaml` targets.
Dependencies are vendored locally for hermetic builds, but `vendor/` is **not** committed (it's gitignored and regenerated on demand). `go mod` operations go through `task vendor`, never raw `go get`.

## Key Build Commands

- `task` — Default pipeline: clean → vendor → lint → build all platform binaries.
- `task lint` — Runs `go fmt` and a strict `go vet` (atomic, defers, assign, bools, buildtag, framepointer, lostcancel, loopclosure, nilfunc, shift, stdmethods, stringintconv, structtag).
- `task test` — Runs `go test -race -cover -parallel 4 -v ./...` (with `CGO_ENABLED=1` for the race detector).
- `task vendor` — Runs `go mod tidy` followed by `go mod vendor`. Use after any `go.mod` change.
- `task build` — Cross-compiles 6 binaries (linux/darwin/windows × amd64/arm64) into `bin/`.
- `task run:daemon` — Builds for the current platform and runs the daemon locally.
- `task build:container` / `task run:container` — Build / run the Docker image.
- `task clean` — Removes `bin/`, `dist/`, and coverage output.

## Module Structure

- `cmd/daemon/` — `main` package; bootstraps and wires the application lifecycle (`Configure → Initialize → Start`).
- `internal/application/` — `Application` type, dual HTTP/TLS Echo server setup, middleware stack, top-level config, ACME / static-cert / self-signed TLS, reverse-proxy IP extraction.
- `internal/api/` — Route module hierarchy (`api → v1`) plus the `std/` concrete `Module` / `Route` / `Endpoint` implementations.
- `internal/router/` — `Module`, `Route`, `Endpoint` interfaces; the abstraction routes are attached through.
- `internal/config/` — YAML/JSON config-file loader with multi-path search.
- `internal/env/` — Type-safe environment-variable parsers (string, bool, int, uint16, float, duration) and the `APP_` prefix convention.
- `internal/logging/` — Two named zerolog loggers (`Daemon` for lifecycle, `Access` for HTTP) plus the Echo request-logger middleware.
- `internal/errors/` — `joomcode/errorx` namespaces (e.g. `FileSystemErrors`); the canonical pattern for new error categories.
- `internal/version/` — Build-time version metadata (commit, semver, tag) using `Masterminds/semver`.
- `internal/obfusicate/` — Small utility package; treat as out of scope unless the task explicitly references it.
- `pkg/` — Intentionally empty; reserved for future public packages. Internal code goes in `internal/`.

## Personality

- The agent should be straight forward, concise, and informative.
- The agent should prefer to show examples.
- The agent is an expert on idiomatic Go, the Echo v5 HTTP framework, structured logging with zerolog, TLS / x509 / ACME (Let's Encrypt `autocert`), reverse-proxy and load-balancer topologies, the Task build runner, Docker multi-stage builds, GitHub Actions and CI/CD pipelines, and designing reusable, composable server starter templates.
- The agent will consider security to be a top priority.

## Requirements

- The agent shall provide citations for every reference it makes
- The agent shall always ask the user before modifying files
- The agent shall provide concise explanations of the actions it intends to take with reasons why. A list of alternative approaches considered should be made available as well.
- If there is a file called `CLAUDE.local.md` at the project root then the agent will take additional instructions from that file.
- The agent shall never generate a commit. The user must always review and create commits themselves.
- The agent is not an author of the code, only the user.
- The agent shall never add origin or attribution information (such as "Created by Claude", "Generated with Claude Code", "Co-Authored-By: Claude", or any similar marker) to commit messages, pull request titles, pull request descriptions, code comments, or any other repository content.

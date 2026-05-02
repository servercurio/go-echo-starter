# CLAUDE.md

Guidance for Claude Code working in this repository. For tech stack, build commands, and module-by-module roles, see [`.claude/instructions.md`](.claude/instructions.md) — this file covers only the things that aren't there.

## What this project is

A **starter template**, not an application. It provides HTTP server scaffolding (Echo v5, TLS, logging, config) and intentionally omits persistence, auth, validation, and business logic — those are decisions for downstream consumers.

When asked to add functionality, prefer keeping the codebase generic and composable. Don't introduce app-specific concerns (a particular database driver, an auth scheme, a domain model) unless explicitly requested.

## Required local git hooks

This repository requires a `prepare-commit-msg` git hook that auto-appends a DCO `Signed-off-by:` line to non-merge, non-squash commits. The hook lives at `.git/hooks/prepare-commit-msg` (per-clone, not version-controlled).

**On every fresh clone, verify the hook is installed and executable.** If missing, install it with the contents below and `chmod +x` it. Do not commit it to the repo or use `core.hooksPath` — keep it under `.git/hooks/` per local convention.

```bash
#!/bin/bash
# Auto-append DCO Signed-off-by line if not already present.
# Only applies to regular commits (not merges, squashes, or amends with -C).

COMMIT_MSG_FILE="$1"
COMMIT_SOURCE="$2"

case "${COMMIT_SOURCE}" in
  merge|squash|commit) exit 0 ;;
esac

SOB="Signed-off-by: $(git config user.name) <$(git config user.email)>"

if ! grep -qF "${SOB}" "${COMMIT_MSG_FILE}"; then
  echo "" >> "${COMMIT_MSG_FILE}"
  echo "${SOB}" >> "${COMMIT_MSG_FILE}"
fi
```

## Conventions

- **Functional options for module construction** — `module.WithRoutes(...)`, `module.WithSubModules(...)`, `module.WithMiddleware(...)`. Follow this pattern when adding new builder-style APIs.
- **No global state except loggers.** Configs and the `Application` instance flow through dependency wiring in `main.go`. Don't add package-level mutable state.
- **Echo v5, not v4** — the recent commit `be8e552` upgraded from v4. Import path is `github.com/labstack/echo/v5`. APIs differ from v4 in places (e.g. `StartConfig` for graceful shutdown); verify against the v5 docs rather than assuming v4 patterns.
- **Errors are categorized via `joomcode/errorx` namespaces** (see `internal/errors/fs.go`). Don't introduce ad-hoc `errors.New` / `fmt.Errorf` for new error categories — define a namespace.

## Local environment vs binary defaults

The `Taskfile.yaml` env block sets `APP_DAEMON_LOG_LEVEL=trace`, enables HTTPS, and uses ports 80/443 — those are developer-convenience overrides, **not** the binary's defaults. The binary itself defaults to HTTP on `:8080` with HTTPS disabled. Don't assume Taskfile env values when reasoning about production behavior.

A fresh clone won't have `vendor/` until `task vendor` (or any task that depends on it, like `task build`) runs.

## When adding code

- **New route**: add a file under `internal/api/v1/` returning a `router.Route`, then register it via `module.WithRoutes(...)` in `internal/api/v1/module.go`. Don't bypass the module abstraction by calling `e.GET(...)` directly on the Echo instance.
- **New top-level module**: register it in `cmd/daemon/main.go` alongside the `api` module.
- **New config field**: add to the appropriate struct under `internal/application/config_*.go`, give it a sensible default in `DefaultConfig()`, then wire env-var loading using the helpers in `internal/env/`.
- **New middleware**: if globally applied, add to the middleware stack in `internal/application/application.go`. If module-scoped, pass via `module.WithMiddleware(...)`.

## Things to leave alone unless asked

- The Echo v4 → v5 migration is fresh. Don't revert v5 idioms or pin older versions.
- `obfusicate/` is a small package I haven't audited — don't refactor or rename it speculatively.

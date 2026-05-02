# CLAUDE.md

Guidance for Claude Code working in this repository. This file covers project intent, conventions, and procedural guidance that isn't captured in the agent reference docs:

- [`.claude/instructions.md`](.claude/instructions.md) — tech stack, personality, requirements
- [`.claude/build-commands.md`](.claude/build-commands.md) — `task` targets and what they run
- [`.claude/module-structure.md`](.claude/module-structure.md) — directory-by-directory roles
- [`.claude/conventions.md`](.claude/conventions.md) — coding conventions for this repo
- [`.claude/git-hooks.md`](.claude/git-hooks.md) — required local git hooks (must be installed per clone)

## What this project is

A **starter template**, not an application. It provides HTTP server scaffolding (Echo v5, TLS, logging, config) and intentionally omits persistence, auth, validation, and business logic — those are decisions for downstream consumers.

When asked to add functionality, prefer keeping the codebase generic and composable. Don't introduce app-specific concerns (a particular database driver, an auth scheme, a domain model) unless explicitly requested.

## Local environment vs binary defaults

The `Taskfile.yaml` env block sets `APP_DAEMON_LOG_LEVEL=trace`, enables HTTPS, and uses ports 80/443 — those are developer-convenience overrides, **not** the binary's defaults. The binary itself defaults to HTTP on `:8080` with HTTPS disabled. Don't assume Taskfile env values when reasoning about production behavior.

A fresh clone won't have `vendor/` until `task vendor` (or any task that depends on it, like `task build`) runs.

## When adding code

- **New route**: add a file under `internal/api/v1/` returning a `router.Route`, then register it via `module.WithRoutes(...)` in `internal/api/v1/module.go`. Don't bypass the module abstraction by calling `e.GET(...)` directly on the Echo instance.
- **New top-level module**: register it in `cmd/daemon/main.go` alongside the `api` module.
- **New config field**: add to the appropriate struct under `internal/application/config_*.go`, give it a sensible default in `DefaultConfig()`, then wire env-var loading using the helpers in `internal/env/`.
- **New middleware**: if globally applied, add to the middleware stack in `internal/application/application.go`. If module-scoped, pass via `module.WithMiddleware(...)`.
- **New CI workflow**: file under `.github/workflows/` following the naming convention in `.github/workflows/docs/naming-standards.md` (`ddd-xxxx-name.yaml` file, matching `ddd: [XXXX] Name` workflow `name:`). PR-triggered workflows use the **200** prefix in this repo (the upstream-standard CITR slot, repurposed locally); reusable workflows use **800**. New `go:generate` directives don't need a workflow change — `task generate` is invoked by both reusable workflows and picks them up via `./...`.

## Things to leave alone unless asked

- The Echo v4 → v5 migration is fresh. Don't revert v5 idioms or pin older versions.
- `obfusicate/` exposes one tiny helper (`ConcealPrefix`) used for log redaction — leave it alone unless the task is explicitly about redaction; the misspelled package name is intentional and matches the import path.

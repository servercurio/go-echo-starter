# Conventions

- **Functional options for module construction** — `module.WithRoutes(...)`, `module.WithSubModules(...)`, `module.WithMiddleware(...)`. Follow this pattern when adding new builder-style APIs.
- **No global state except loggers.** Configs and the `Application` instance flow through dependency wiring in `main.go`. Don't add package-level mutable state.
- **Echo v5, not v4** — the recent commit `be8e552` upgraded from v4. Import path is `github.com/labstack/echo/v5`. APIs differ from v4 in places (e.g. `StartConfig` for graceful shutdown); verify against the v5 docs rather than assuming v4 patterns.
- **Errors are categorized via `joomcode/errorx` namespaces** (see `internal/errors/fs.go`). Don't introduce ad-hoc `errors.New` / `fmt.Errorf` for new error categories — define a namespace.

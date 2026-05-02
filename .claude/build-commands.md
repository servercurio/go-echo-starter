# Key Build Commands

- `task` — Default pipeline: clean → vendor → lint → build all platform binaries.
- `task lint` — Runs `go fmt` and a strict `go vet` (atomic, defers, assign, bools, buildtag, framepointer, lostcancel, loopclosure, nilfunc, shift, stdmethods, stringintconv, structtag).
- `task test` — Runs `go test -race -cover -parallel 4 -v ./...` (with `CGO_ENABLED=1` for the race detector).
- `task vendor` — Runs `go mod tidy` followed by `go mod vendor`. Use after any `go.mod` change.
- `task build` — Cross-compiles 6 binaries (linux/darwin/windows × amd64/arm64) into `bin/`.
- `task run:daemon` — Builds for the current platform and runs the daemon locally.
- `task build:container` / `task run:container` — Build / run the Docker image.
- `task clean` — Removes `bin/`, `dist/`, and coverage output.

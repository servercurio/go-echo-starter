# Key Build Commands

- `task` — Default pipeline: clean → vendor → lint → build all platform binaries.
- `task lint` — Runs `go fmt` and a strict `go vet` (atomic, defers, assign, bools, buildtag, framepointer, lostcancel, loopclosure, nilfunc, shift, stdmethods, stringintconv, structtag).
- `task test` — Runs `go test -parallel 4 -cover -coverprofile cover.out -race -v ./...` (with `CGO_ENABLED=1` for the race detector). The `cover.out` profile is written into the repo root and removed by `task clean`.
- `task vendor` — Runs `go mod tidy` followed by `go mod vendor`. Use after any `go.mod` change.
- `task generate` — Runs `go generate ./...`. Required after a fresh clone (and run automatically by `task build`) because `internal/version/commit.txt` is gitignored but is `//go:embed`-ed.
- `task build` — Cross-compiles 6 binaries (linux/darwin/windows × amd64/arm64) into `bin/`. Calls `task generate` first.
- `task hash` — Writes one `bin/<binary>.sha256` file per built binary (`shasum -c` compatible). Run locally to verify a build (`cd bin && shasum -a 256 -c appsvrd-linux-amd64.sha256`).
- `task sign` — Depends on `task hash`. GPG-signs each built binary AND its `.sha256` file, producing `bin/<binary>.asc` and `bin/<binary>.sha256.asc`. Requires `gpg` with a configured signing key. Verify with `gpg --verify bin/appsvrd-linux-amd64.asc bin/appsvrd-linux-amd64` or `gpg --verify bin/appsvrd-linux-amd64.sha256.asc bin/appsvrd-linux-amd64.sha256`.
- `task run:daemon` — Builds for the current platform and runs the daemon locally.
- `task build:container` / `task run:container` — Build / run the Docker image.
- `task clean` — Removes `bin/`, `dist/`, and coverage output.

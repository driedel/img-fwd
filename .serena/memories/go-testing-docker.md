# Go commands via Docker

The local Go toolchain (brew, go1.26) exists **only** so gopls/Serena can analyze the code. Never run `go test`, `go build`, `go vet` or `go mod` directly — the repo convention (AGENTS.md) is Docker-only, pinned to `golang:1.22-alpine` matching `go.mod` (`go 1.22`).

Run tests:

```bash
docker run --rm -v "$PWD/app:/app" -w /app golang:1.22-alpine go test -v ./...
```

Vet:

```bash
docker run --rm -v "$PWD/app:/app" -w /app golang:1.22-alpine go vet ./...
```

Go module root is `app/` (not repo root). There is no linter/typecheck config — `go test` + `go vet` via Docker is the only verification step.

# img-fwd — Agent Notes

> **For full project context** (demo site, deployment, detailed architecture): See `AGENTS.md`.

Compact Go proxy in front of imgproxy. Source lives in `app/`, Docker config at repo root.

## Development

- **Go is NOT installed locally.** All Go commands (test, build, mod) must run via Docker.
- **Go module:** `app/`
- **Run tests:**
  ```bash
  docker run --rm -v "$PWD/app:/app" -w /app golang:1.25-alpine go test -v ./...
  ```
- **No linter / typecheck config present.** The only verification step is `go test` via Docker.

## Running locally

- **Basic:** `docker compose up -d` (reads `.env`)
- **With external Docker network** (e.g., to reach another local project):
  ```bash
  docker compose -f docker-compose.yml -f docker/docker-compose.local.yml up --build -d
  ```
  Requires `.env` with `EXTERNAL_NETWORK=my-project_default` and optionally `SOURCE_BASE_URL`.

## Build & deploy

- **Docker build context:** repo root (Dockerfile copies from `app/`).
- **CI:** Tests run on every push to `main`. Docker image `driedel/img-fwd` is built and pushed **only on version tags** (`v*`). Multi-platform: `linux/amd64,linux/arm64`.

### Fly.io (production)

Two apps on Fly.io, both in region `gru` (São Paulo):

| App | What it serves | When to redeploy |
|---|---|---|
| `img-fwd` | Go proxy | After CI pushes a new Docker image (post version tag) |
| `img-fwd-demo-static` | Demo site (nginx) | When any file in `demo/` changes |

**Always use the deploy script** — it handles both apps, checks auth, and verifies status:

```bash
./scripts/deploy.sh           # deploy both proxy and nginx
./scripts/deploy.sh --proxy   # proxy only (after new Docker image from CI)
./scripts/deploy.sh --nginx   # nginx only (after demo/ file changes)
```

- The proxy pulls `driedel/img-fwd:latest` from Docker Hub — no local build needed.
- `flyctl auth login` required if not authenticated.

## Key runtime behavior

- **Port defaults differ:** `8888` in `docker-compose.yml` vs `8000` in Go fallback (`PORT` env var).
- **Origin reconstruction:** Uses `Host` header. `SOURCE_BASE_URL` overrides this entirely.
- **Allowed origins:** comma-separated `ALLOWED_ORIGINS`. Empty = allow all (logged as a warning).
- **Routing logic:**
  - Non-image paths (`.html`, `.js`, etc.) go directly to origin, even with `?f=avif&rs=600`.
  - Image paths go through imgproxy **only when transformation params are present**; otherwise they pass through to origin.
- **Auto-format:** `.jpg/.png/.webp/.tiff/.bmp` → AVIF; `.gif/.svg/.ico/.avif` passthrough. Explicit `f=` overrides auto-format.
- **Query param forwarding:** non-imgproxy params (`v=2`, `cache=1`, etc.) are forwarded to the origin URL, not imgproxy.
- **Health:** `GET /healthz` → `200 ok`.

## Entrypoints

- App entry: `app/main.go` (`package main`)
- Docker entrypoint: `docker/entrypoint.sh` (starts imgproxy on `:8080`, then `img-fwd`)
- Base image: `darthsim/imgproxy:latest`; Go binary built via multi-stage `Dockerfile`.

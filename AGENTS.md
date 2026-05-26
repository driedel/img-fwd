# img-fwd — Agent Notes

Compact Go proxy in front of imgproxy. Source lives in `app/`, Docker config at repo root.

## Development

- **Go is NOT installed locally.** All Go commands (test, build, mod) must run via Docker.
- **Go module:** `app/`
- **Run tests:**
  ```bash
  docker run --rm -v "$PWD/app:/app" -w /app golang:1.22-alpine go test -v ./...
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

## Key runtime behavior

- **Port defaults differ:** `8888` in `docker-compose.yml` vs `8000` in Go fallback (`PORT` env var).
- **Origin reconstruction:** Uses `Host` header. `SOURCE_BASE_URL` overrides this entirely.
- **Allowed origins:** comma-separated `ALLOWED_ORIGINS`. Empty = allow all (logged as a warning).
- **Routing logic:**
  - Non-image paths (`.html`, `.js`, etc.) go directly to origin, even with `?f=avif&rs=600`.
  - Image paths go through imgproxy **only when transformation params are present**; otherwise they pass through to origin.
- **Auto-format:** `.jpg/.png/.webp/.tiff/.bmp` → AVIF; `.gif` → animated WebP; `.svg/.ico/.avif` passthrough. Explicit `f=` overrides auto-format.
- **Query param forwarding:** non-imgproxy params (`v=2`, `cache=1`, etc.) are forwarded to the origin URL, not imgproxy.
- **Health:** `GET /healthz` → `200 ok`.

## Demo Site

A live demo is hosted at `https://img-fwd.driedel.dev` (GitHub Pages) showcasing img-fwd capabilities.

### Architecture

```
GitHub Pages (img-fwd.driedel.dev)  →  serves demo/ folder (HTML/CSS/JS + original images)
         ▲
         │  SOURCE_BASE_URL
    Fly.io (cdn.img-fwd.driedel.dev)  →  img-fwd proxy (scale-to-zero, free tier)
         │
    imgproxy (internal)  →  transforms images (AVIF, WebP, resize, etc.)
```

### Deploy the demo

1. **Generate demo images** (AI prompts listed in `demo/README.md`)
2. **GitHub Pages:**
   - Settings → Pages → Source: `/demo` folder
   - Custom domain: `img-fwd.driedel.dev`
3. **Fly.io:**
   - `brew install flyctl` (or follow [docs](https://fly.io/docs/hands-on/install-flyctl/))
   - `fly auth login`
   - `fly deploy` (reads `fly.toml`)
4. **Cloudflare DNS:**
   - CNAME `img-fwd.driedel.dev` → `driedel.github.io` (for GitHub Pages)
   - CNAME `cdn.img-fwd.driedel.dev` → `img-fwd-demo.fly.dev` (for Fly.io proxy)

### Demo files

- `demo/index.html` — landing page
- `demo/css/styles.css` — responsive dark theme
- `demo/js/main.js` — before/after sliders, Resource Timing API
- `demo/images/` — example images (originals served by GitHub Pages)
- `fly.toml` — Fly.io configuration (scale-to-zero, free tier)

## Entrypoints

- App entry: `app/main.go` (`package main`)
- Docker entrypoint: `docker/entrypoint.sh` (starts imgproxy on `:8080`, then `img-fwd`)
- Base image: `darthsim/imgproxy:latest`; Go binary built via multi-stage `Dockerfile`.

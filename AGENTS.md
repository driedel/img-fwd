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

A live demo is hosted at `https://img-fwd.driedel.dev` showcasing img-fwd capabilities.

### Architecture

The img-fwd proxy sits **in front** of the origin. All requests go through the proxy — images without params pass through, images with transformation params are processed.

```
                          ┌─────────────────────────────────────┐
                          │         img-fwd.driedel.dev         │
                          │         (Cloudflare DNS)            │
                          └───────────────┬─────────────────────┘
                                          │
                                          ▼
                          ┌─────────────────────────────────────┐
                          │           Fly.io (free)           │
                          │      img-fwd proxy container      │
                          │  ┌─────────────────────────────┐  │
                          │  │  HTML/CSS/JS → pass through │  │
                          │  │  /images/photo.jpg → origin │  │
                          │  │  /photo.jpg?rs=800 → imgproxy│ │
                          │  └─────────────────────────────┘  │
                          └───────────────┬─────────────────────┘
                                          │
                          ┌───────────────┴─────────────────────┐
                          │                                     │
                          ▼                                     ▼
          ┌─────────────────────────────┐      ┌─────────────────────────────┐
          │   nginx internal app        │      │   imgproxy (internal)       │
          │   img-fwd-demo-static       │      │   transforms images           │
          │   serves demo/ folder       │      │   (AVIF, WebP, resize, etc) │
          └─────────────────────────────┘      └─────────────────────────────┘
```

### How the proxy works in this setup

1. User requests `img-fwd.driedel.dev/index.html` → proxy fetches from nginx internal app → returns HTML
2. User requests `img-fwd.driedel.dev/images/photo.jpg` → proxy fetches from nginx internal app → returns original
3. User requests `img-fwd.driedel.dev/images/photo.jpg?rs=800` → proxy processes via imgproxy → returns optimized AVIF

### Deploy the demo

#### Automated deployment (recommended)

Use the deploy script to deploy both apps:

```bash
# Deploy both proxy and nginx
./scripts/deploy.sh

# Deploy only proxy (when fly.toml changes)
./scripts/deploy.sh --proxy

# Deploy only nginx (when demo/ files change)
./scripts/deploy.sh --nginx
```

**Prerequisites:**
- `flyctl` CLI installed (`brew install flyctl` on macOS)
- Authenticated with Fly.io (`flyctl auth login`)

**OpenCode skill:** When asked to deploy, use the `deploy` skill which provides step-by-step instructions and verification commands.

#### Manual deployment

1. **Generate demo images** (AI prompts listed below) and place in `demo/app/images/`
2. **Internal nginx app (serves static files):**
   - `cd demo/`
   - `fly apps create img-fwd-demo-static`
   - `fly deploy`
3. **Proxy app (public entrypoint):**
   - `cd` (repo root)
   - `fly deploy` (reads `fly.toml`)
4. **Cloudflare DNS:**
   - CNAME `img-fwd.driedel.dev` → `img-fwd-demo.fly.dev` (the proxy is the entrypoint)
   - **No separate CDN subdomain** — the proxy IS the main domain

### Demo files

- `demo/app/index.html` — landing page (uses relative URLs like `images/photo.jpg?rs=800`)
- `demo/app/css/styles.css` — responsive dark theme
- `demo/app/js/main.js` — Resource Timing API for live size measurements
- `demo/app/images/` — example images (served by nginx internal app)
- `fly.toml` — Fly.io proxy configuration with `SOURCE_BASE_URL` pointing to internal nginx app
- `demo/Dockerfile` — nginx:alpine serving `app/` folder
- `demo/fly.toml` — Fly.io internal app (no exposed ports)

## Entrypoints

- App entry: `app/main.go` (`package main`)
- Docker entrypoint: `docker/entrypoint.sh` (starts imgproxy on `:8080`, then `img-fwd`)
- Base image: `darthsim/imgproxy:latest`; Go binary built via multi-stage `Dockerfile`.

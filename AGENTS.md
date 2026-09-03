# img-fwd — Agent Notes

Compact Go proxy in front of imgproxy. Source lives in `app/`, Docker config at repo root.

## Development

- **All Go commands (test, build, mod) must run via Docker.** A local Go toolchain exists only for editor tooling (gopls/Serena).
- **Go module:** `app/` (Go 1.27)
- **Run tests:**
  ```bash
  docker run --rm -v "$PWD/app:/app" -w /app golang:1.27-alpine go test -v ./...
  ```
- **Security checks** (same as CI): `gosec ./...` and `govulncheck ./...` inside `golang:1.27-alpine`; fuzz via `go test -run=NONE -fuzz=FuzzVerifySignature -fuzztime=30s .`

## Running locally

- **Basic:** `docker compose up -d` (reads `.env`)
- **With external Docker network** (e.g., to reach another local project):
  ```bash
  docker compose -f docker-compose.yml -f docker/docker-compose.local.yml up --build -d
  ```
  Requires `.env` with `EXTERNAL_NETWORK=my-project_default` and optionally `SOURCE_BASE_URL`.

## Build & deploy

- **Docker build context:** repo root (Dockerfile copies from `app/`).
- **CI:** Tests + fuzz + security gates (gosec, govulncheck, Trivy image scan) run on every push to `main` and on PRs. Docker image `driedel/img-fwd` is built and pushed **only on version tags** (`v*`). Multi-platform: `linux/amd64,linux/arm64`.
- **Dependabot:** weekly PRs keep GitHub Actions, Go modules and Docker base images up to date.

### Fly.io (production)

Two apps on Fly.io, both in region `gru` (São Paulo):

| App | What it serves | When to redeploy |
|---|---|---|
| `img-fwd` | Go proxy | After CI pushes a new Docker image (post version tag) |
| `img-fwd-demo-static` | Demo site (nginx) | When any file in `demo/` changes |

- The proxy pulls `driedel/img-fwd:latest` from Docker Hub — no local build needed.
- **Always use the deploy script** (see "Deploy the demo" below) — it handles both apps, checks auth, and verifies status.

## Key runtime behavior

- **Port defaults differ:** `8888` in `docker-compose.yml` vs `8000` in Go fallback (`PORT` env var).
- **Origin reconstruction:** Uses `Host` header. `SOURCE_BASE_URL` overrides this entirely.
- **Allowed origins:** comma-separated `ALLOWED_ORIGINS`. Empty = allow all (logged as a warning).
- **Routing logic:**
  - Non-image paths (`.html`, `.js`, etc.) go directly to origin, even with `?f=avif&rs=600`.
  - Image paths go through imgproxy **only when transformation params are present**; otherwise they pass through to origin.
- **Auto-format:** `.jpg/.png/.webp/.tiff/.bmp` → AVIF; `.gif/.svg/.ico/.avif` passthrough. Explicit `f=` overrides auto-format.
- **Query param forwarding:** non-imgproxy params (`v=2`, `cache=1`, etc.) are forwarded to the origin URL, not imgproxy.
- **Dual-source mode:** when `SIGNING_KEY` + `S3_*` envs are set, requests with a valid `exp`+`sig` HMAC-SHA256 signature route to a private S3/MinIO bucket (SigV4 presigned GetObject); invalid/expired → 403; unsigned → public source only. Private responses get `Cache-Control: private, max-age=900`. Startup fails fast if `SIGNING_KEY` is set without full S3 config. See README for the signing scheme.
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


<!-- headroom:rtk-instructions -->
# RTK (Rust Token Killer) - Token-Optimized Commands

When running shell commands, **always prefix with `rtk`**. This reduces context
usage by 60-90% with zero behavior change. If rtk has no filter for a command,
it passes through unchanged — so it is always safe to use.

## Key Commands
```bash
# Git (59-80% savings)
rtk git status          rtk git diff            rtk git log

# Files & Search (60-75% savings)
rtk ls <path>           rtk read <file>         rtk grep <pattern>
rtk find <pattern>      rtk diff <file>

# Test (90-99% savings) — shows failures only
rtk pytest tests/       rtk cargo test          rtk test <cmd>

# Build & Lint (80-90% savings) — shows errors only
rtk tsc                 rtk lint                rtk cargo build
rtk prettier --check    rtk mypy                rtk ruff check

# Analysis (70-90% savings)
rtk err <cmd>           rtk log <file>          rtk json <file>
rtk summary <cmd>       rtk deps                rtk env

# GitHub (26-87% savings)
rtk gh pr view <n>      rtk gh run list         rtk gh issue list

# Infrastructure (85% savings)
rtk docker ps           rtk kubectl get         rtk docker logs <c>

# Package managers (70-90% savings)
rtk pip list            rtk pnpm install        rtk npm run <script>
```

## Rules
- In command chains, prefix each segment: `rtk git add . && rtk git commit -m "msg"`
- For debugging, use raw command without rtk prefix
- `rtk proxy <cmd>` runs command without filtering but tracks usage
<!-- /headroom:rtk-instructions -->

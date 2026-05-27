---
name: project-context
description: Use when asked about how img-fwd works, its architecture, Go source files, query params, routing rules, env vars, Docker setup, or any general project question.
---

# Project Context — img-fwd

## What this is

`img-fwd` is a minimal Go HTTP proxy that transparently forwards image requests to an embedded [imgproxy](https://imgproxy.net) instance. It lets teams optimize images (format conversion, resize, quality, blur) **without changing existing URLs** — query params drive transformations.

## Architecture

```
Browser → img-fwd (Go) → imgproxy (:8080) → origin (Host header or SOURCE_BASE_URL)
```

- **No framework:** stdlib `net/http` only.
- **Stateless:** all config via env vars.
- **Two binaries in one container:** imgproxy starts first via `docker/entrypoint.sh`, then `img-fwd` binds to `:8888`.
- **Go is NOT installed locally.** Tests and builds must run inside a Docker container (see `go-testing` skill).

## Key source files

| File | Role |
|---|---|
| `app/main.go` | Entrypoint, request routing, param mapping, origin reconstruction |
| `app/main_test.go` | Unit tests using `httptest` to mock imgproxy and origin |
| `docker/entrypoint.sh` | Container startup — launches imgproxy then img-fwd |
| `Dockerfile` | Multi-stage: `golang:1.22-alpine` builder → `darthsim/imgproxy:latest` runtime |

## Query param → imgproxy mapping

| Param | imgproxy segment | Notes |
|---|---|---|
| `f=webp` | `f:webp` | Explicit format disables auto-format |
| `rs=600` | `rs:fit:600:0` | Width only |
| `rs=600:400` | `rs:fit:600:400` | Width:height |
| `g=sm` | `g:sm` | Gravity (crop focal point) |
| `q=80` | `q:80` | Quality 1–100 |
| `blur=5` | `bl:5` | Blur intensity |

Non-imgproxy params (e.g. `v=2`, `cache=1`) are forwarded to the origin URL, not imgproxy.

## Routing rules

1. `/healthz` → `200 ok` immediately.
2. Non-image extensions (`.html`, `.js`, `.css`, etc.) → origin directly, even if transformation params are present.
3. Image extensions without transformation params → origin directly (no imgproxy round-trip).
4. Image extensions with params → imgproxy with built processing options.

## Auto-format behavior

When no `f=` param is provided:

| Extension | Auto-format |
|---|---|
| `.jpg/.jpeg/.png/.webp/.tiff/.bmp` | AVIF |
| `.gif/.svg/.ico/.avif` | Passthrough (no conversion) |

## Env var behavior

- `PORT`: `8888` in Docker, `8000` fallback in Go binary alone.
- `ALLOWED_ORIGINS`: empty = allow all (warning logged).
- `SOURCE_BASE_URL`: overrides `Host` header reconstruction entirely.
- `IMGPROXY_LOG_LEVEL`: passed to imgproxy, defaults to `warn`.

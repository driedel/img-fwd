# Architecture

Single-binary Go proxy (`app/main.go`, `package main`) in front of imgproxy. Zero external dependencies (stdlib only).

## Request flow (handler in main.go)

1. `/healthz` → 200 ok.
2. Origin host from `Host` header (port stripped); `SOURCE_BASE_URL` overrides reconstruction entirely.
3. Path sanitized with `path.Clean` (traversal protection).
4. `ALLOWED_ORIGINS` check (empty = allow all, logged as warning).
5. Query params split: imgproxy params (`f,rs,g,q,blur`) become path segments; signing params (`exp,sig`) are consumed internally; everything else is forwarded to the origin URL.
6. Auto-format: `.jpg/.jpeg/.png/.webp/.tiff/.bmp` → AVIF unless `f=` is explicit. `.gif/.svg/.ico/.avif` passthrough.
7. Routes through imgproxy only when processing options exist AND path has an image extension; otherwise direct origin fetch. Response headers copied, body streamed.

## Dual-source mode (SIGNING_KEY + S3_*)

- `app/sign.go`: `sig = hex(HMAC-SHA256(SIGNING_KEY, "<path>\n<exp>"))`, constant-time compare. Absent sig → public route; invalid/expired → 403 hard-fail (never falls back to public).
- `app/s3.go`: hand-rolled AWS SigV4 presigned GetObject (path-style, MinIO + AWS compatible). Envs: `S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY, S3_USE_SSL (default false), S3_REGION (default us-east-1)`.
- Valid sig → presigned URL is the source (URL-escaped when embedded in the imgproxy path; fetched directly for passthrough). No sig → public source only.
- Private responses get `Cache-Control: private, max-age=900` (URLs expire; presign TTL 15 min).
- Startup fail-fast: `SIGNING_KEY` without full `S3_*` → `log.Fatal` (`validateConfig`).

## Ports

`PORT` env; default `8888` in docker-compose.yml, `8000` in Go fallback. imgproxy runs on `:8080` inside the container (see `docker/entrypoint.sh`).

<picture>
  <source srcset="demo/app/images/logo-dark.svg" media="(prefers-color-scheme: dark)">
  <source srcset="demo/app/images/logo-light.svg" media="(prefers-color-scheme: light)">
  <img src="web/static/images/logo-light.svg" width="300" alt="IMG FWD">
</picture>


HTTP proxy in Go that sits in front of [imgproxy](https://imgproxy.net), enabling **transparent image optimization** — no changes to existing URLs required.

Docker image: [`driedel/img-fwd`](https://hub.docker.com/r/driedel/img-fwd)

---

## How it works

The proxy uses the `Host` header from each request to reconstruct the original image URL automatically. Just point your domain's DNS to the proxy server.

```
Without proxy:
  Browser → cdn.examplesite.com/image.png

With proxy:
  Browser → img-fwd (Host: cdn.examplesite.com)
                 ↓ translates query params into imgproxy options
            imgproxy fetches → https://cdn.examplesite.com/image.png
```

**URL in your app (unchanged):**
```
https://cdn.examplesite.com/marketing/image.png?f=avif&rs=600
```

---

## Routing rules

1. `/healthz` → returns `200 ok` immediately.
2. Non-image paths (`.html`, `.js`, `.css`, `.json`, etc.) are forwarded to the origin directly, even if transformation params are present.
3. Image paths without transformation params are forwarded to the origin directly (no imgproxy round-trip).
4. Image paths with transformation params are processed by imgproxy with the requested options.

---

## Configuration

| Variable | Required | Description | Example |
|---|---|---|---|
| `ALLOWED_ORIGINS` | No* | Comma-separated list of allowed origin domains | `cdn.examplesite.com,assets.another.com` |
| `PORT` | No | Port the proxy listens on (default: `8888`) | `9010` |
| `SOURCE_BASE_URL` | No | Overrides automatic origin URL reconstruction | `http://internal-service:5000` |
| `SIGNING_KEY` | No | Enables dual-source mode: signed requests are routed to a private S3 bucket | random 32+ char secret |
| `S3_ENDPOINT` | No** | S3/MinIO endpoint (host:port) for the private bucket | `minio:9000`, `s3.amazonaws.com` |
| `S3_BUCKET` | No** | Private bucket name | `my-private-assets` |
| `S3_ACCESS_KEY` | No** | S3 access key | `AKIA...` |
| `S3_SECRET_KEY` | No** | S3 secret key | `...` |
| `S3_USE_SSL` | No | Use HTTPS for S3 (default: `false`) | `true` |
| `S3_REGION` | No | S3 region (default: `us-east-1`; MinIO accepts any value) | `sa-east-1` |
| `IMGPROXY_LOG_LEVEL` | No | Log level for the internal imgproxy (default: `warn`) | `info` |
| `EXTERNAL_NETWORK` | No | External Docker network to join (local dev only) | `my-project_default` |

> *If `ALLOWED_ORIGINS` is empty, all origins are accepted. **Not recommended in production.**
>
> **Required when `SIGNING_KEY` is set — the proxy refuses to start otherwise.

---

## Dual-source mode (public + private buckets)

When `SIGNING_KEY` and the `S3_*` variables are set, img-fwd runs in **dual-source mode**: a single instance serves a public origin *and* a private S3/MinIO bucket, routed by request signature:

```
Request with exp + sig:
    valid, unexpired signature → PRIVATE bucket (S3 presigned GetObject)
    invalid or expired         → 403 (never falls back to public)
Request without exp + sig:
    → PUBLIC source only (SOURCE_BASE_URL / Host header), never the private bucket
```

The private bucket never needs public access: img-fwd generates a **presigned S3 URL** (AWS Signature V4, 15 min TTL) and hands it to imgproxy — or fetches the object itself for passthrough requests. Responses from the private route always carry `Cache-Control: private, max-age=900`, since signed URLs expire.

### Signing scheme

The backend application signs image URLs with the same `SIGNING_KEY`:

```
sig = hex( HMAC-SHA256(SIGNING_KEY, "<path>\n<exp>") )
```

- `<path>` — the URL path only (e.g. `/photos/cat.jpg`); query params are **not** signed
- `<exp>` — Unix timestamp when the URL expires (a 15 min TTL is a good default)

Example in Go:

```go
func SignURL(path string, ttl time.Duration) string {
    exp := time.Now().Add(ttl).Unix()
    mac := hmac.New(sha256.New, []byte(signingKey))
    fmt.Fprintf(mac, "%s\n%d", path, exp)
    return fmt.Sprintf("%s?exp=%d&sig=%x", path, exp, mac.Sum(nil))
}
// → /photos/cat.jpg?exp=1750000900&sig=3f2a...
```

Transformation params compose freely with signed URLs (they are not part of the signature):

```
/photos/cat.jpg?exp=1750000900&sig=3f2a...&rs=800      → AVIF, resized, from the private bucket
/photos/cat.jpg?rs=800                                 → public source (no signature)
```

When `SIGNING_KEY` is **not** set, the proxy behaves exactly as before (open, public origin) — `exp`/`sig` params are simply ignored. Video/audio files are not served by img-fwd in either mode.

---

## Usage

```bash
docker compose up -d
```

Or with `docker run`:

```bash
docker run -d \
  -p 8888:8888 \
  -e ALLOWED_ORIGINS=cdn.examplesite.com \
  driedel/img-fwd:latest
```

### Demo

Run the full demo stack (nginx + proxy) locally:

```bash
docker compose -f demo/docker-compose.demo.yml up --build -d
```

Open [http://localhost:8888](http://localhost:8888).

### Local development

Used when img-fwd needs to reach services running inside another project's Docker network. Set the required variables in `.env`:

```env
SOURCE_BASE_URL=http://internal-service:80
EXTERNAL_NETWORK=my-project_default
```

Then start with the local overlay:

```bash
docker compose -f docker-compose.yml -f docker/docker-compose.local.yml up --build -d
```

---

## Automatic format conversion

img-fwd automatically converts images to a modern format when no explicit `f=` param is provided:

| Format | Auto-converts to | Reason |
|---|---|---|
| `.jpg`, `.jpeg`, `.png`, `.webp`, `.tiff`, `.bmp` | AVIF | Best compression for static images |
| `.gif`, `.svg`, `.ico`, `.avif` | — (passthrough) | Served as-is without conversion |

Setting `f=` explicitly always overrides the automatic behavior:
```
/image.jpg          → converted to AVIF automatically
/image.jpg?f=webp   → converted to WebP (explicit override)
/animation.gif      → served as-is (passthrough)
/icon.svg           → served as-is
```

---

## Supported transformation parameters

| Param | Description | Example | imgproxy equivalent |
|---|---|---|---|
| `f` | Output format | `f=avif`, `f=webp`, `f=png` | `f:avif` |
| `rs` | Resize (width or width:height) | `rs=600`, `rs=600:400` | `rs:fit:600:0` |
| `g` | Gravity (crop focal point) | `g=sm`, `g=ce` | `g:sm` |
| `q` | Quality (1–100) | `q=80` | `q:80` |
| `blur` | Blur intensity | `blur=5` | `bl:5` |

Non-transformation query params are forwarded as-is to the origin URL.

### Examples

```
/marketing/image.png?f=avif&rs=600
/photos/product.jpg?f=webp&rs=800:600&q=75
/banner.png?blur=3&f=avif
/image.png?f=avif&v=2   ← "v=2" is forwarded to the origin, not imgproxy
```

---

## Production setup

Point your domain's DNS to the server running img-fwd. The proxy detects the domain automatically from the `Host` header — no additional configuration is needed to add new domains.

---

## Health check

```
GET /healthz → 200 ok
```

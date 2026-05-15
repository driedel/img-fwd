```
                 _____________  ___________   ____________       _________ 
                 ____  _/__   |/  /_  ____/   ___  ____/_ |     / /__  __ \
                  __  / __  /|_/ /_  / __     __  /_   __ | /| / /__  / / /
                 __/ /  _  /  / / / /_/ /     _  __/   __ |/ |/ / _  /_/ / 
                 /___/  /_/  /_/  \____/      /_/      ____/|__/  /_____/  
```

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

## Configuration

| Variable | Required | Description | Example |
|---|---|---|---|
| `ALLOWED_ORIGINS` | No* | Comma-separated list of allowed origin domains | `cdn.examplesite.com,assets.another.com` |
| `PORT` | No | Port the proxy listens on (default: `8888`) | `9010` |
| `SOURCE_BASE_URL` | No | Overrides automatic origin URL reconstruction | `http://internal-service:5000` |
| `IMGPROXY_LOG_LEVEL` | No | Log level for the internal imgproxy (default: `warn`) | `info` |
| `EXTERNAL_NETWORK` | No | External Docker network to join (local dev only) | `my-project_default` |

> *If `ALLOWED_ORIGINS` is empty, all origins are accepted. **Not recommended in production.**

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
| `.gif` | WebP | Preserves animation — imgproxy does not support animated AVIF |
| `.svg`, `.ico`, `.avif` | — (passthrough) | Vector/already-modern formats, no conversion needed |

Setting `f=` explicitly always overrides the automatic behavior:
```
/image.jpg          → converted to AVIF automatically
/image.jpg?f=webp   → converted to WebP (explicit override)
/animation.gif      → converted to animated WebP automatically
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

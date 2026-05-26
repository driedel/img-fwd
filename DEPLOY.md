# Deploy Instructions

## Architecture

```
img-fwd.driedel.dev (Cloudflare)
         │
         ▼
┌─────────────────────┐
│  img-fwd-demo       │  ← Public proxy (driedel/img-fwd:latest)
│  Fly.io             │     SOURCE_BASE_URL = http://img-fwd-demo-static.internal
└────────┬────────────┘
         │ Fly.io internal DNS
         ▼
┌─────────────────────┐
│  img-fwd-demo-static│  ← Private nginx (nginx:alpine)
│  Fly.io             │     Serves demo/ folder
└─────────────────────┘
```

## Step 1: Deploy internal nginx app

```bash
cd /Users/daniloriedel/projects/github/img-fwd/demo-static
fly apps create img-fwd-demo-static
fly deploy
```

This creates a private app that only other Fly.io apps can reach.

## Step 2: Deploy public proxy app

```bash
cd /Users/daniloriedel/projects/github/img-fwd
fly deploy
```

The proxy is now configured to fetch originals from the internal nginx app.

## Step 3: Configure DNS

In Cloudflare, add:

| Type | Name | Target |
|------|------|--------|
| CNAME | img-fwd | img-fwd-demo.fly.dev |

## Verification

Test these URLs after DNS propagates:

- `https://img-fwd.driedel.dev/` → Demo page loads
- `https://img-fwd.driedel.dev/images/photo.jpg` → Original image
- `https://img-fwd.driedel.dev/images/photo.jpg?rs=800` → Optimized AVIF
- `https://img-fwd.driedel.dev/css/styles.css` → Stylesheet (pass-through)

## Troubleshooting

If images don't load:
1. Check proxy logs: `fly logs -a img-fwd-demo`
2. Check nginx is running: `fly status -a img-fwd-demo-static`
3. Verify internal DNS: `fly ssh console -a img-fwd-demo` then `curl http://img-fwd-demo-static.internal/`

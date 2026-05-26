# Setup Guide for New Developers

This guide helps you set up your local environment to deploy the img-fwd demo site.

## Prerequisites

### 1. Install flyctl CLI

**macOS:**
```bash
brew install flyctl
```

**Linux:**
```bash
curl -L https://fly.io/install.sh | sh
```

**Verify installation:**
```bash
flyctl version
```

### 2. Authenticate with Fly.io

```bash
flyctl auth login
```

This will open your browser for authentication.

**Verify authentication:**
```bash
flyctl auth whoami
```

### 3. Request Access to Fly.io Organization

You need to be added to the Fly.io organization that owns the `img-fwd` and `img-fwd-demo-static` apps.

**Contact the project maintainer** and provide:
- Your Fly.io email address
- Request "member" or "admin" access

**Verify access:**
```bash
flyctl apps list
```

You should see:
- `img-fwd` (proxy app)
- `img-fwd-demo-static` (nginx app)

## Deployment

### Automated Deployment (Recommended)

Use the deploy script:

```bash
# Deploy both apps
./scripts/deploy.sh

# Deploy only proxy (when fly.toml changes)
./scripts/deploy.sh --proxy

# Deploy only nginx (when demo/ files change)
./scripts/deploy.sh --nginx
```

### Manual Deployment

**Deploy proxy:**
```bash
flyctl deploy -a img-fwd
```

**Deploy nginx:**
```bash
cd demo
flyctl deploy -a img-fwd-demo-static
```

## Verification

After deployment, verify everything is working:

```bash
# Check proxy status
flyctl status -a img-fwd

# Check nginx status
flyctl status -a img-fwd-demo-static

# Test the site
curl -I https://img-fwd.driedel.dev/
```

## Troubleshooting

### "App not found" error

You don't have access to the Fly.io organization. Contact the maintainer.

### "Authentication required" error

Run:
```bash
flyctl auth login
```

### "flyctl: command not found"

Install flyctl (see Prerequisites section).

### Deployment fails with "no machines running"

The app might be scaled to zero. Deploy again:
```bash
flyctl deploy -a img-fwd
```

## Architecture Overview

```
img-fwd.driedel.dev (Cloudflare)
         │
         ▼
┌─────────────────────┐
│  img-fwd (proxy)    │  ← Public, scale-to-zero
│  driedel/img-fwd    │
└────────┬────────────┘
         │ .internal DNS
         ▼
┌─────────────────────┐
│  img-fwd-demo-static│  ← Internal, always on
│  nginx:alpine       │
└─────────────────────┘
```

**Two Fly.io apps:**
1. **img-fwd** - Public proxy using `driedel/img-fwd:latest` from Docker Hub
2. **img-fwd-demo-static** - Internal nginx serving static files from `demo/app/`

## When to Deploy

- **Proxy (img-fwd)**: When `fly.toml` in root changes
- **Nginx (img-fwd-demo-static)**: When files in `demo/` change (HTML, CSS, JS, images)

## Useful Commands

```bash
# View proxy logs
flyctl logs -a img-fwd

# View nginx logs
flyctl logs -a img-fwd-demo-static

# SSH to proxy
flyctl ssh console -a img-fwd

# SSH to nginx
flyctl ssh console -a img-fwd-demo-static

# Check machine status
flyctl machines list -a img-fwd
flyctl machines list -a img-fwd-demo-static
```

## Notes

- Both apps are in the `gru` region (São Paulo, Brazil)
- Proxy has scale-to-zero enabled (free tier)
- Nginx is always running (required for internal DNS resolution)
- The proxy uses `driedel/img-fwd:latest` from Docker Hub (no local build needed)
- The nginx app is built from `demo/Dockerfile` on each deploy

## Support

For issues or questions:
1. Check the [DEPLOY.md](DEPLOY.md) file
2. Review [AGENTS.md](AGENTS.md) for architecture details
3. Contact the project maintainer

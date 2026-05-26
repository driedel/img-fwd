---
name: deploy
description: Deploy img-fwd demo to Fly.io (proxy and nginx apps)
---

# Deploy to Fly.io

This skill deploys the img-fwd demo site to Fly.io. It deploys two apps:

1. **img-fwd** (proxy) - Uses `driedel/img-fwd:latest` from Docker Hub
2. **img-fwd-demo-static** (nginx) - Built from `demo/` folder

## Prerequisites

- `flyctl` CLI installed and authenticated
- Access to Fly.io organization with both apps

## Usage

When the user asks to deploy, run the following commands:

### 1. Deploy Proxy (if fly.toml changed)

```bash
cd /Users/daniloriedel/projects/github/img-fwd
flyctl deploy -a img-fwd
```

### 2. Deploy Nginx (if demo/ files changed)

```bash
cd /Users/daniloriedel/projects/github/img-fwd/demo
flyctl deploy -a img-fwd-demo-static
```

### 3. Verify Deployment

```bash
# Check proxy status
flyctl status -a img-fwd

# Check nginx status
flyctl status -a img-fwd-demo-static

# Test the site
curl -I https://img-fwd.driedel.dev/
```

## When to Deploy

- **Proxy (img-fwd)**: When `fly.toml` in root changes
- **Nginx (img-fwd-demo-static)**: When files in `demo/` change (HTML, CSS, JS, images)

## Troubleshooting

### flyctl not found

Install flyctl:
```bash
# macOS
brew install flyctl

# Linux
curl -L https://fly.io/install.sh | sh
```

### Authentication required

```bash
flyctl auth login
```

### App not found

Verify you have access to the Fly.io organization:
```bash
flyctl apps list
```

## Notes

- The proxy uses `driedel/img-fwd:latest` from Docker Hub (no build needed)
- The nginx app is built from `demo/Dockerfile` on each deploy
- Both apps are in the `gru` region (São Paulo)
- Proxy has scale-to-zero enabled (free tier)
- Nginx is always running (required for internal DNS)

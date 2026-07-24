# Deploy

- Docker build context: repo root (`Dockerfile` copies from `app/`). Base image `darthsim/imgproxy:latest`; entrypoint `docker/entrypoint.sh` starts imgproxy on `:8080`, then img-fwd.
- CI: tests run on every push to `main`. Image `driedel/img-fwd` (multi-platform amd64+arm64) is built/pushed **only on version tags `v*`**.
- Demo (https://img-fwd.driedel.dev): two Fly.io apps — `img-fwd-demo` (public proxy, repo root `fly.toml`) and `img-fwd-demo-static` (internal nginx serving `demo/app/`, no exposed ports). Proxy's `SOURCE_BASE_URL` points to the internal nginx app.
- Deploy via `./scripts/deploy.sh` (`--proxy`, `--nginx`, or both). Requires `flyctl` authenticated. Cloudflare CNAME points the domain at the proxy app.
- Local run: `docker compose up -d` (reads `.env`). External network overlay: `docker compose -f docker-compose.yml -f docker/docker-compose.local.yml up --build -d` with `EXTERNAL_NETWORK` in `.env`.

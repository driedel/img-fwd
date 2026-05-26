# Security Check

## Before committing changes, verify:

### 1. No secrets or credentials

```bash
grep -ri "password\|secret\|token\|key=" app/ docker/ --include="*.go" --include="*.sh" --include="*.yml" --include="*.md"
```

- Ensure no hardcoded API keys, passwords, or tokens in source.
- Ensure `.env` is in `.gitignore` (check: it should already be).

### 2. Dockerfile security

```bash
cat Dockerfile
```

- Multi-stage build is required (builder + runtime).
- Runtime base is `darthsim/imgproxy:latest` — verify it does not run as root if possible.
- `entrypoint.sh` should not leak env vars in logs.

### 3. Go dependencies

> Go is not installed locally — use Docker for these commands.

```bash
docker run --rm -v "$PWD/app:/app" -w /app golang:1.22-alpine go list -m all
```

- Verify no unexpected or deprecated dependencies.
- If `govulncheck` is available, run:
  ```bash
  docker run --rm -v "$PWD/app:/app" -w /app golang:1.22-alpine sh -c "go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./..."
  ```

### 4. Header / proxy behavior

- Ensure the proxy does not blindly forward dangerous headers (e.g. `Host`, `X-Forwarded-*`) in a way that breaks origin security.
- Ensure `ALLOWED_ORIGINS` restricts origins in production (empty = allow all, which logs a warning).
- Ensure `SOURCE_BASE_URL` cannot be used to open redirect vulnerabilities.

### 5. Input validation

- Query params are parsed with `url.Values` — ensure no path traversal in `r.URL.Path`.
- `splitHostPort` strips ports from `Host` header correctly.

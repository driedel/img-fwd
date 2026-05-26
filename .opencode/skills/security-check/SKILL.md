---
name: security-check
description: Use when reviewing a commit or change for security issues: secrets, Dockerfile hardening, Go dependencies, proxy header behavior, input validation.
---

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

> **Transparent proxy is intentional and safe.** `Cookie` and `Authorization` headers are forwarded intact to the downstream target (either imgproxy or origin). This is correct because:
> - When routed to **imgproxy**, it is a local trusted process inside the same container; imgproxy does **not** re-forward those headers to the origin — it fetches the origin image itself with a fresh HTTP request.
> - When routed **directly to origin** (passthrough), the client headers are meant to reach the origin unchanged.
> Therefore, stripping `Cookie`/`Authorization` is unnecessary and would break legitimate use cases.

### 5. Input validation

- Query params are parsed with `url.Values` — ensure no path traversal in `r.URL.Path`.
- `splitHostPort` strips ports from `Host` header correctly.

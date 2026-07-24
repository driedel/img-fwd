---
name: go-testing
description: Use when running, writing, or debugging Go tests in this repo. Go is NOT installed locally — all go commands must run inside Docker.
---

# Go Testing

> **Go is NOT installed locally.** All `go` commands must run inside a Docker container.

## Run all tests

```bash
docker run --rm -v "$PWD/app:/app" -w /app golang:1.25-alpine go test -v ./...
```

## Run a specific test

```bash
docker run --rm -v "$PWD/app:/app" -w /app golang:1.25-alpine go test -v -run TestHandlerHealthz ./...
```

## Check coverage

```bash
docker run --rm -v "$PWD/app:/app" -w /app golang:1.25-alpine go test -cover ./...
```

## Test patterns used in this repo

- **Table-driven tests** for pure functions (`parseAllowedOrigins`, `buildProcessingOptions`).
- **Mock servers** via `httptest.NewServer` for handler/integration-style tests.
- **Global var swap + defer restore** for config override in tests (e.g. `imgproxyURL`, `sourceBaseURL`, `allowedOrigins`).
- No external test framework — only stdlib `testing`.

## Adding tests

1. Keep tests in `app/main_test.go` (`package main`).
2. For handler tests, always restore global vars in `defer` so test order doesn't matter.
3. Example pattern for handler test:

```go
func TestHandlerSomething(t *testing.T) {
    // Mock imgproxy or origin
    mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // assertions on request
        w.WriteHeader(http.StatusOK)
    }))
    defer mock.Close()

    // Save and restore globals
    origImgproxy, origSource, origOrigins := imgproxyURL, sourceBaseURL, allowedOrigins
    defer func() { imgproxyURL, sourceBaseURL, allowedOrigins = origImgproxy, origSource, origOrigins }()

    imgproxyURL = mock.URL
    sourceBaseURL = "http://origin"
    allowedOrigins = map[string]bool{}

    req := httptest.NewRequest("GET", "/image.jpg?rs=600", nil)
    rr := httptest.NewRecorder()
    handler(rr, req)

    // assert response code, captured request URI, etc.
}
```

## No integration tests

There are no DB, network, or imgproxy integration tests. Everything is unit-testable with `httptest`. Do not add external test dependencies without discussion.

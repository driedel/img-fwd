package main

import (
	"strings"
	"testing"
	"time"
)

var testS3 = s3Config{
	endpoint:  "minio:9000",
	bucket:    "stimo-private",
	accessKey: "AKID",
	secretKey: "SECRET",
	region:    "us-east-1",
	useSSL:    false,
}

func TestPresignGetObjectFormat(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	u := testS3.presignGetObject("/photos/cat.jpg", 15*time.Minute, now)

	if !strings.HasPrefix(u, "http://minio:9000/stimo-private/photos/cat.jpg?") {
		t.Errorf("expected path-style URL, got %q", u)
	}
	for _, want := range []string{
		"X-Amz-Algorithm=AWS4-HMAC-SHA256",
		"X-Amz-Credential=AKID%2F20250615%2Fus-east-1%2Fs3%2Faws4_request",
		"X-Amz-Date=20250615T120000Z",
		"X-Amz-Expires=900",
		"X-Amz-SignedHeaders=host",
		"X-Amz-Signature=",
	} {
		if !strings.Contains(u, want) {
			t.Errorf("expected presigned URL to contain %q, got %q", want, u)
		}
	}
}

func TestPresignGetObjectDeterministic(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	a := testS3.presignGetObject("/photos/cat.jpg", 15*time.Minute, now)
	b := testS3.presignGetObject("/photos/cat.jpg", 15*time.Minute, now)
	if a != b {
		t.Errorf("presign must be deterministic:\n%s\n%s", a, b)
	}
}

func TestPresignGetObjectDifferentKeysDiffer(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	a := testS3.presignGetObject("/a.jpg", 15*time.Minute, now)
	b := testS3.presignGetObject("/b.jpg", 15*time.Minute, now)
	if a == b {
		t.Error("different keys must produce different presigned URLs")
	}
}

func TestPresignGetObjectSSL(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	cfg := testS3
	cfg.useSSL = true
	u := cfg.presignGetObject("/a.jpg", 15*time.Minute, now)
	if !strings.HasPrefix(u, "https://") {
		t.Errorf("expected https scheme with useSSL=true, got %q", u)
	}
}

func TestPresignGetObjectEscapesKey(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	u := testS3.presignGetObject("/fotos/café manhã.jpg", 15*time.Minute, now)
	if strings.Contains(u, "café manhã.jpg") {
		t.Errorf("expected special chars in key to be escaped, got %q", u)
	}
	if !strings.Contains(u, "/fotos/") {
		t.Errorf("path separators must be preserved, got %q", u)
	}
}

func TestS3Enabled(t *testing.T) {
	if testS3.enabled() != true {
		t.Error("config with endpoint must be enabled")
	}
	empty := s3Config{}
	if empty.enabled() != false {
		t.Error("config without endpoint must be disabled")
	}
}

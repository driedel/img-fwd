package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// --- parseAllowedOrigins ---

func TestParseAllowedOrigins(t *testing.T) {
	tests := []struct {
		input string
		want  map[string]bool
	}{
		{"", map[string]bool{}},
		{"cdn.example.com", map[string]bool{"cdn.example.com": true}},
		{"cdn.example.com,assets.example.com", map[string]bool{"cdn.example.com": true, "assets.example.com": true}},
		{" cdn.example.com , assets.example.com ", map[string]bool{"cdn.example.com": true, "assets.example.com": true}},
	}
	for _, tt := range tests {
		got := parseAllowedOrigins(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("parseAllowedOrigins(%q): got %d origins, want %d", tt.input, len(got), len(tt.want))
			continue
		}
		for k := range tt.want {
			if !got[k] {
				t.Errorf("parseAllowedOrigins(%q): missing %q", tt.input, k)
			}
		}
	}
}

// --- isAllowed ---

func TestIsAllowed(t *testing.T) {
	orig := allowedOrigins
	defer func() { allowedOrigins = orig }()

	allowedOrigins = map[string]bool{}
	if !isAllowed("anything.com") {
		t.Error("empty allowedOrigins should allow all origins")
	}

	allowedOrigins = map[string]bool{"cdn.example.com": true}
	if !isAllowed("cdn.example.com") {
		t.Error("should allow a listed origin")
	}
	if isAllowed("other.com") {
		t.Error("should deny an unlisted origin")
	}
}

// --- autoFormat ---

func TestAutoFormat(t *testing.T) {
	avif := []string{"/img.jpg", "/img.jpeg", "/img.png", "/img.webp", "/img.tiff", "/img.bmp", "/img.JPG"}
	for _, p := range avif {
		if got := autoFormat(p); got != "avif" {
			t.Errorf("autoFormat(%q) = %q, want \"avif\"", p, got)
		}
	}

	none := []string{"/img.svg", "/img.ico", "/img.avif", "/noext", "/page.html", "/animation.gif", "/animation.GIF"}
	for _, p := range none {
		if got := autoFormat(p); got != "" {
			t.Errorf("autoFormat(%q) = %q, want \"\"", p, got)
		}
	}
}

// --- hasImageExtension ---

func TestHasImageExtension(t *testing.T) {
	images := []string{"/img.jpg", "/img.jpeg", "/img.png", "/img.webp", "/img.avif", "/img.tiff", "/img.bmp", "/img.svg", "/img.ico"}
	for _, p := range images {
		if !hasImageExtension(p) {
			t.Errorf("hasImageExtension(%q) = false, want true", p)
		}
	}

	notImages := []string{"/page.html", "/style.css", "/script.js", "/noext"}
	for _, p := range notImages {
		if hasImageExtension(p) {
			t.Errorf("hasImageExtension(%q) = true, want false", p)
		}
	}
}

// --- buildProcessingOptions ---

func TestBuildProcessingOptions(t *testing.T) {
	tests := []struct {
		rawQuery   string
		autoFormat string
		want       string
	}{
		{"", "", ""},
		{"", "avif", "f:avif"},
		{"", "webp", "f:webp"},
		{"f=webp", "avif", "f:webp"},
		{"rs=600", "avif", "f:avif/rs:fit:600:0"},
		{"rs=600:400", "avif", "f:avif/rs:fit:600:400"},
		{"g=sm", "avif", "f:avif/g:sm"},
		{"q=80", "avif", "f:avif/q:80"},
		{"blur=5", "avif", "f:avif/bl:5"},
		{"f=webp&rs=800:600&q=75", "", "f:webp/rs:fit:800:600/q:75"},
	}
	for _, tt := range tests {
		q, _ := url.ParseQuery(tt.rawQuery)
		got := buildProcessingOptions(q, tt.autoFormat)
		if got != tt.want {
			t.Errorf("buildProcessingOptions(%q, %q) = %q, want %q", tt.rawQuery, tt.autoFormat, got, tt.want)
		}
	}
}

// --- handler ---

func TestHandlerHealthz(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "ok" {
		t.Errorf("expected body \"ok\", got %q", rr.Body.String())
	}
}

func TestHandlerForbiddenOrigin(t *testing.T) {
	orig := allowedOrigins
	defer func() { allowedOrigins = orig }()
	allowedOrigins = map[string]bool{"cdn.example.com": true}

	req := httptest.NewRequest("GET", "/image.jpg", nil)
	req.Host = "other.com"
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}

func TestHandlerRoutesImageThroughImgproxy(t *testing.T) {
	var capturedURI string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI
		w.WriteHeader(http.StatusOK)
	}))
	defer mock.Close()

	origImgproxy, origSource, origOrigins := imgproxyURL, sourceBaseURL, allowedOrigins
	defer func() { imgproxyURL, sourceBaseURL, allowedOrigins = origImgproxy, origSource, origOrigins }()

	imgproxyURL = mock.URL
	sourceBaseURL = "http://origin"
	allowedOrigins = map[string]bool{}

	req := httptest.NewRequest("GET", "/image.jpg?f=webp&rs=600", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !strings.HasPrefix(capturedURI, "/insecure/") {
		t.Errorf("expected request routed to imgproxy, got URI %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "f:webp") {
		t.Errorf("expected f:webp in imgproxy URI, got %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "rs:fit:600:0") {
		t.Errorf("expected rs:fit:600:0 in imgproxy URI, got %q", capturedURI)
	}
}

func TestHandlerAutoAvif(t *testing.T) {
	var capturedURI string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI
		w.WriteHeader(http.StatusOK)
	}))
	defer mock.Close()

	origImgproxy, origSource, origOrigins := imgproxyURL, sourceBaseURL, allowedOrigins
	defer func() { imgproxyURL, sourceBaseURL, allowedOrigins = origImgproxy, origSource, origOrigins }()

	imgproxyURL = mock.URL
	sourceBaseURL = "http://origin"
	allowedOrigins = map[string]bool{}

	req := httptest.NewRequest("GET", "/image.png?rs=600", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !strings.Contains(capturedURI, "f:avif") {
		t.Errorf("expected automatic f:avif in imgproxy URI, got %q", capturedURI)
	}
}

func TestHandlerGifPassesThrough(t *testing.T) {
	var originHit bool
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originHit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer origin.Close()

	origImgproxy, origSource, origOrigins := imgproxyURL, sourceBaseURL, allowedOrigins
	defer func() { imgproxyURL, sourceBaseURL, allowedOrigins = origImgproxy, origSource, origOrigins }()

	imgproxyURL = "http://imgproxy-should-not-be-called"
	sourceBaseURL = origin.URL
	allowedOrigins = map[string]bool{}

	req := httptest.NewRequest("GET", "/animation.gif?rs=600", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !originHit {
		t.Error("expected .gif to pass through to origin, not imgproxy")
	}
}

func TestHandlerDirectToOriginForNonConvertibleFormat(t *testing.T) {
	var originHit bool
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originHit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer origin.Close()

	origImgproxy, origSource, origOrigins := imgproxyURL, sourceBaseURL, allowedOrigins
	defer func() { imgproxyURL, sourceBaseURL, allowedOrigins = origImgproxy, origSource, origOrigins }()

	imgproxyURL = "http://imgproxy-should-not-be-called"
	sourceBaseURL = origin.URL
	allowedOrigins = map[string]bool{}

	// .svg has no auto-format and no explicit params → goes directly to origin
	req := httptest.NewRequest("GET", "/icon.svg", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !originHit {
		t.Error("expected non-convertible format (.svg) to go directly to origin")
	}
}

func TestHandlerNonImageSkipsImgproxy(t *testing.T) {
	var originHit bool
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originHit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer origin.Close()

	origImgproxy, origSource, origOrigins := imgproxyURL, sourceBaseURL, allowedOrigins
	defer func() { imgproxyURL, sourceBaseURL, allowedOrigins = origImgproxy, origSource, origOrigins }()

	imgproxyURL = "http://imgproxy-should-not-be-called"
	sourceBaseURL = origin.URL
	allowedOrigins = map[string]bool{}

	// HTML file with transformation params → should bypass imgproxy
	req := httptest.NewRequest("GET", "/page.html?f=avif&rs=600", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !originHit {
		t.Error("expected non-image path to bypass imgproxy even when transformation params are present")
	}
}

func TestHandlerForwardsNonImgproxyParams(t *testing.T) {
	var capturedURI string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI
		w.WriteHeader(http.StatusOK)
	}))
	defer mock.Close()

	origImgproxy, origSource, origOrigins := imgproxyURL, sourceBaseURL, allowedOrigins
	defer func() { imgproxyURL, sourceBaseURL, allowedOrigins = origImgproxy, origSource, origOrigins }()

	imgproxyURL = mock.URL
	sourceBaseURL = "http://origin"
	allowedOrigins = map[string]bool{}

	// v=42 and cache=1 are not imgproxy params — they must be forwarded in the origin URL
	req := httptest.NewRequest("GET", "/image.jpg?v=42&cache=1", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !strings.Contains(capturedURI, "v=42") {
		t.Errorf("expected non-imgproxy param v=42 forwarded in origin URL, got imgproxy URI %q", capturedURI)
	}
}

func TestHandlerStripsPortFromHost(t *testing.T) {
	var capturedURI string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI
		w.WriteHeader(http.StatusOK)
	}))
	defer mock.Close()

	origImgproxy, origSource, origOrigins := imgproxyURL, sourceBaseURL, allowedOrigins
	defer func() { imgproxyURL, sourceBaseURL, allowedOrigins = origImgproxy, origSource, origOrigins }()

	imgproxyURL = mock.URL
	sourceBaseURL = ""
	allowedOrigins = map[string]bool{"cdn.example.com": true}

	req := httptest.NewRequest("GET", "/image.jpg?rs=600", nil)
	req.Host = "cdn.example.com:443"
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code == http.StatusForbidden {
		t.Errorf("expected host with port to be allowed, got 403")
	}
	if !strings.Contains(capturedURI, "cdn.example.com") {
		t.Errorf("expected origin URL to contain clean host, got imgproxy URI %q", capturedURI)
	}
	if strings.Contains(capturedURI, ":443") {
		t.Errorf("expected port to be stripped from origin URL, got imgproxy URI %q", capturedURI)
	}
}

func TestHandlerSanitizesPathTraversal(t *testing.T) {
	var capturedURI string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI
		w.WriteHeader(http.StatusOK)
	}))
	defer mock.Close()

	origImgproxy, origSource, origOrigins := imgproxyURL, sourceBaseURL, allowedOrigins
	defer func() { imgproxyURL, sourceBaseURL, allowedOrigins = origImgproxy, origSource, origOrigins }()

	imgproxyURL = mock.URL
	sourceBaseURL = "http://origin"
	allowedOrigins = map[string]bool{}

	req := httptest.NewRequest("GET", "/../../../secret.jpg?rs=600", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if strings.Contains(capturedURI, "..") {
		t.Errorf("expected path traversal to be sanitized, got imgproxy URI %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "/secret.jpg") {
		t.Errorf("expected sanitized path /secret.jpg in imgproxy URI, got %q", capturedURI)
	}
}

// --- dual-source (SIGNING_KEY + S3_*) ---

// setupDualSource enables dual-source mode with a fake S3 backend and returns
// a cleanup that restores all globals.
func setupDualSource(t *testing.T, s3Handler http.HandlerFunc) (s3Server *httptest.Server, s3Hit *bool) {
	t.Helper()
	hit := false
	s3Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		s3Handler(w, r)
	}))

	origKey, origS3, origOrigins, origSource := signingKey, s3, allowedOrigins, sourceBaseURL
	t.Cleanup(func() {
		signingKey, s3, allowedOrigins, sourceBaseURL = origKey, origS3, origOrigins, origSource
		s3Server.Close()
	})

	signingKey = testSigningKey
	s3 = s3Config{
		endpoint:  strings.TrimPrefix(s3Server.URL, "http://"),
		bucket:    "private-bucket",
		accessKey: "AKID",
		secretKey: "SECRET",
		region:    "us-east-1",
		useSSL:    false,
	}
	allowedOrigins = map[string]bool{}
	sourceBaseURL = "http://public-origin"
	return s3Server, &hit
}

func signedURL(path string) string {
	exp := signedExp(15 * time.Minute)
	sig := computeSignature(testSigningKey, path, exp)
	return path + "?exp=" + exp + "&sig=" + sig
}

func TestValidateConfig(t *testing.T) {
	origKey, origS3 := signingKey, s3
	defer func() { signingKey, s3 = origKey, origS3 }()

	signingKey = ""
	s3 = s3Config{}
	if err := validateConfig(); err != nil {
		t.Errorf("no SIGNING_KEY must always be valid, got %v", err)
	}

	signingKey = "key"
	s3 = s3Config{}
	if err := validateConfig(); err == nil {
		t.Error("SIGNING_KEY without S3 config must fail validation")
	}

	s3 = s3Config{endpoint: "minio:9000", bucket: "b", accessKey: "a", secretKey: "s"}
	if err := validateConfig(); err != nil {
		t.Errorf("fully configured dual mode must be valid, got %v", err)
	}

	s3 = s3Config{endpoint: "minio:9000"}
	if err := validateConfig(); err == nil {
		t.Error("partial S3 config must fail validation")
	}
}

func TestHandlerValidSignatureRoutesToPrivateS3(t *testing.T) {
	var s3Query string
	_, s3Hit := setupDualSource(t, func(w http.ResponseWriter, r *http.Request) {
		s3Query = r.URL.RawQuery
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
	})

	// .gif has no auto-format → passthrough: img-fwd fetches the presigned URL directly.
	req := httptest.NewRequest("GET", signedURL("/private/photo.gif"), nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !*s3Hit {
		t.Fatal("expected private S3 backend to be hit")
	}
	if !strings.Contains(s3Query, "X-Amz-Signature=") {
		t.Errorf("expected presigned query at S3 backend, got %q", s3Query)
	}
	if !strings.Contains(s3Query, "X-Amz-Credential=AKID") {
		t.Errorf("expected credential in presigned query, got %q", s3Query)
	}
}

func TestHandlerInvalidSignature403(t *testing.T) {
	_, s3Hit := setupDualSource(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	exp := signedExp(15 * time.Minute)
	req := httptest.NewRequest("GET", "/private/photo.jpg?exp="+exp+"&sig=wrong", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
	if *s3Hit {
		t.Error("invalid signature must never reach the private backend")
	}
}

func TestHandlerExpiredSignature403(t *testing.T) {
	_, s3Hit := setupDualSource(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	exp := signedExp(-time.Minute)
	sig := computeSignature(testSigningKey, "/private/photo.jpg", exp)
	req := httptest.NewRequest("GET", "/private/photo.jpg?exp="+exp+"&sig="+sig, nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for expired signature, got %d", rr.Code)
	}
	if *s3Hit {
		t.Error("expired signature must never reach the private backend")
	}
}

func TestHandlerUnsignedStaysPublic(t *testing.T) {
	_, s3Hit := setupDualSource(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	var publicHit bool
	public := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publicHit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer public.Close()
	sourceBaseURL = public.URL

	// Unsigned request for a path that only exists "privately" → must go public only.
	// .svg has no auto-format → direct origin fetch.
	req := httptest.NewRequest("GET", "/private/icon.svg", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !publicHit {
		t.Error("unsigned request must hit the public origin")
	}
	if *s3Hit {
		t.Error("unsigned request must never hit the private bucket")
	}
}

func TestHandlerSignedTransformGoesThroughImgproxy(t *testing.T) {
	_, _ = setupDualSource(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	var capturedURI string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI
		w.WriteHeader(http.StatusOK)
	}))
	defer mock.Close()

	origImgproxy := imgproxyURL
	defer func() { imgproxyURL = origImgproxy }()
	imgproxyURL = mock.URL

	exp := signedExp(15 * time.Minute)
	sig := computeSignature(testSigningKey, "/private/photo.jpg", exp)
	req := httptest.NewRequest("GET", "/private/photo.jpg?rs=600&exp="+exp+"&sig="+sig, nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !strings.HasPrefix(capturedURI, "/insecure/") {
		t.Fatalf("expected imgproxy route, got %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "rs:fit:600:0") {
		t.Errorf("expected resize options in imgproxy URI, got %q", capturedURI)
	}
	if !strings.Contains(capturedURI, "X-Amz-Signature") {
		t.Errorf("expected escaped presigned URL in imgproxy URI, got %q", capturedURI)
	}
}

func TestHandlerPrivateResponseCacheControl(t *testing.T) {
	setupDualSource(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", signedURL("/photo.gif"), nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if got := rr.Header().Get("Cache-Control"); got != "private, max-age=900" {
		t.Errorf("expected short private cache on signed response, got %q", got)
	}
}

func TestHandlerPublicResponseCacheUntouched(t *testing.T) {
	setupDualSource(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	public := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.WriteHeader(http.StatusOK)
	}))
	defer public.Close()
	sourceBaseURL = public.URL

	req := httptest.NewRequest("GET", "/icon.svg", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if got := rr.Header().Get("Cache-Control"); got != "public, max-age=31536000" {
		t.Errorf("public response cache headers must pass through unchanged, got %q", got)
	}
}

func TestHandlerSigningParamsNotForwarded(t *testing.T) {
	var s3Query string
	setupDualSource(t, func(w http.ResponseWriter, r *http.Request) {
		s3Query = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", signedURL("/photo.gif"), nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if strings.Contains(s3Query, "sig=") && !strings.Contains(s3Query, "X-Amz-") {
		t.Errorf("app signing params must not leak upstream, got %q", s3Query)
	}
	if strings.Contains(s3Query, "exp=") {
		t.Errorf("exp param must not leak upstream, got %q", s3Query)
	}
}

func TestHandlerNoSigningKeyIgnoresSignature(t *testing.T) {
	var capturedURI string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI
		w.WriteHeader(http.StatusOK)
	}))
	defer mock.Close()

	origKey, origImgproxy, origSource, origOrigins := signingKey, imgproxyURL, sourceBaseURL, allowedOrigins
	defer func() { signingKey, imgproxyURL, sourceBaseURL, allowedOrigins = origKey, origImgproxy, origSource, origOrigins }()

	signingKey = ""
	imgproxyURL = mock.URL
	sourceBaseURL = "http://origin"
	allowedOrigins = map[string]bool{}

	// Legacy mode: exp/sig in query are ignored, request goes to imgproxy normally.
	req := httptest.NewRequest("GET", "/image.jpg?rs=600&exp=123&sig=abc", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected legacy mode to ignore exp/sig, got %d", rr.Code)
	}
	if !strings.Contains(capturedURI, "rs:fit:600:0") {
		t.Errorf("expected normal imgproxy routing, got %q", capturedURI)
	}
}

// --- security ---

func TestHandlerHostCaseInsensitive(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mock.Close()

	origImgproxy, origSource, origOrigins := imgproxyURL, sourceBaseURL, allowedOrigins
	defer func() { imgproxyURL, sourceBaseURL, allowedOrigins = origImgproxy, origSource, origOrigins }()

	imgproxyURL = mock.URL
	sourceBaseURL = ""
	allowedOrigins = map[string]bool{"cdn.example.com": true}

	// DNS names are case-insensitive — a mixed-case Host must not be rejected.
	req := httptest.NewRequest("GET", "/image.jpg?rs=600", nil)
	req.Host = "CDN.Example.COM"
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code == http.StatusForbidden {
		t.Error("mixed-case Host of an allowed origin must be accepted")
	}
}

func TestParseAllowedOriginsLowercasesEntries(t *testing.T) {
	got := parseAllowedOrigins("CDN.Example.COM")
	if !got["cdn.example.com"] {
		t.Error("allowed origins must be stored lowercased")
	}
}

func TestHandlerStripsSensitiveClientHeaders(t *testing.T) {
	var gotAuth, gotConnection, gotUA string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotConnection = r.Header.Get("Connection")
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	origImgproxy, origSource, origOrigins := imgproxyURL, sourceBaseURL, allowedOrigins
	defer func() { imgproxyURL, sourceBaseURL, allowedOrigins = origImgproxy, origSource, origOrigins }()

	imgproxyURL = "http://imgproxy-should-not-be-called"
	sourceBaseURL = upstream.URL
	allowedOrigins = map[string]bool{}

	req := httptest.NewRequest("GET", "/icon.svg", nil)
	req.Header.Set("Authorization", "Bearer client-secret-token")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "img-fwd-test")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if gotAuth != "" {
		t.Errorf("client Authorization header leaked upstream: %q", gotAuth)
	}
	if gotConnection != "" {
		t.Errorf("hop-by-hop Connection header leaked upstream: %q", gotConnection)
	}
	if gotUA != "img-fwd-test" {
		t.Errorf("regular headers must still be forwarded, got User-Agent %q", gotUA)
	}
}

func TestRedactURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"presigned s3 url",
			"http://minio:9000/b/k.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKID%2F20250615%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Signature=abc123",
			"X-Amz-Signature=REDACTED",
		},
		{
			"app signature",
			"http://origin/img.jpg?exp=1750000000&sig=deadbeef",
			"sig=REDACTED",
		},
	}
	for _, tt := range tests {
		got := redactURL(tt.input)
		if !strings.Contains(got, tt.want) {
			t.Errorf("%s: expected %q in redacted URL, got %q", tt.name, tt.want, got)
		}
		if strings.Contains(got, "abc123") || strings.Contains(got, "deadbeef") || strings.Contains(got, "AKID") {
			t.Errorf("%s: sensitive material still present in %q", tt.name, got)
		}
	}

	plain := "http://origin/img.jpg?v=2"
	if got := redactURL(plain); got != plain {
		t.Errorf("URL without sensitive params must be returned unchanged, got %q", got)
	}
}

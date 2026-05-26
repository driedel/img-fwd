package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
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

	webp := []string{"/animation.gif", "/animation.GIF"}
	for _, p := range webp {
		if got := autoFormat(p); got != "webp" {
			t.Errorf("autoFormat(%q) = %q, want \"webp\"", p, got)
		}
	}

	none := []string{"/img.svg", "/img.ico", "/img.avif", "/noext", "/page.html"}
	for _, p := range none {
		if got := autoFormat(p); got != "" {
			t.Errorf("autoFormat(%q) = %q, want \"\"", p, got)
		}
	}
}

// --- hasImageExtension ---

func TestHasImageExtension(t *testing.T) {
	images := []string{"/img.jpg", "/img.jpeg", "/img.png", "/img.webp", "/img.avif", "/img.tiff", "/img.bmp", "/img.svg", "/img.ico", "/img.gif"}
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

func TestHandlerGifAutoConvertToWebp(t *testing.T) {
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

	req := httptest.NewRequest("GET", "/animation.gif", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !strings.Contains(capturedURI, "f:webp") {
		t.Errorf("expected automatic f:webp for .gif, got %q", capturedURI)
	}
	if strings.Contains(capturedURI, "f:avif") {
		t.Errorf("expected .gif NOT to use f:avif, got %q", capturedURI)
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

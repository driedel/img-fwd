package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var (
	imgproxyURL    = getEnv("IMGPROXY_URL", "http://localhost:8080")
	port           = getEnv("PORT", "8000")
	allowedOrigins = parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS"))
	sourceBaseURL  = getEnv("SOURCE_BASE_URL", "")
	signingKey     = os.Getenv("SIGNING_KEY")
	s3             = s3Config{
		endpoint:  os.Getenv("S3_ENDPOINT"),
		bucket:    os.Getenv("S3_BUCKET"),
		accessKey: os.Getenv("S3_ACCESS_KEY"),
		secretKey: os.Getenv("S3_SECRET_KEY"),
		region:    getEnv("S3_REGION", "us-east-1"),
		useSSL:    getEnv("S3_USE_SSL", "false") == "true",
	}
)

// presignTTL is how long the S3 presigned URLs handed to imgproxy stay valid.
const presignTTL = 15 * time.Minute

// validateConfig enforces that SIGNING_KEY (private routing) only runs with a
// fully configured S3 source — a misconfigured private route must fail loudly
// at startup, never silently fall back to the public source.
func validateConfig() error {
	if signingKey == "" {
		return nil
	}
	missing := []string{}
	if s3.endpoint == "" {
		missing = append(missing, "S3_ENDPOINT")
	}
	if s3.bucket == "" {
		missing = append(missing, "S3_BUCKET")
	}
	if s3.accessKey == "" {
		missing = append(missing, "S3_ACCESS_KEY")
	}
	if s3.secretKey == "" {
		missing = append(missing, "S3_SECRET_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("SIGNING_KEY requires private S3 source; missing: %s", strings.Join(missing, ", "))
	}
	return nil
}

func parseAllowedOrigins(s string) map[string]bool {
	m := map[string]bool{}
	for _, origin := range strings.Split(s, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			m[origin] = true
		}
	}
	return m
}

func isAllowed(host string) bool {
	if len(allowedOrigins) == 0 {
		return true
	}
	return allowedOrigins[host]
}

// avifConvertibleExtensions lists raster formats that imgproxy can encode as AVIF.
// SVG, ICO and GIF are excluded: SVG/ICO are vector/icon formats; GIF passes through unchanged.
var avifConvertibleExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true,
	".webp": true, ".tiff": true, ".bmp": true,
}

func autoFormat(path string) string {
	dot := strings.LastIndex(path, ".")
	if dot == -1 {
		return ""
	}
	ext := strings.ToLower(path[dot:])
	if avifConvertibleExtensions[ext] {
		return "avif"
	}
	return ""
}

// buildProcessingOptions maps friendly query params to imgproxy path segments.
// Supported params:
//
//	f=webp      → f:webp  (explicit format; disables auto-format)
//	rs=600      → rs:fit:600:0
//	rs=600:400  → rs:fit:600:400
//	g=sm        → g:sm  (gravity)
//	q=80        → q:80  (quality)
//	blur=5      → bl:5
//
// When format is non-empty and no f= param is present, f:<format> is added automatically.
func buildProcessingOptions(q url.Values, format string) string {
	var parts []string

	if f := q.Get("f"); f != "" {
		parts = append(parts, "f:"+f)
	} else if format != "" {
		parts = append(parts, "f:"+format)
	}

	if rs := q.Get("rs"); rs != "" {
		dims := strings.SplitN(rs, ":", 2)
		w := dims[0]
		h := "0"
		if len(dims) == 2 {
			h = dims[1]
		}
		parts = append(parts, fmt.Sprintf("rs:fit:%s:%s", w, h))
	}

	if g := q.Get("g"); g != "" {
		parts = append(parts, "g:"+g)
	}

	if q2 := q.Get("q"); q2 != "" {
		parts = append(parts, "q:"+q2)
	}

	if blur := q.Get("blur"); blur != "" {
		parts = append(parts, "bl:"+blur)
	}

	return strings.Join(parts, "/")
}

var imgproxyParams = map[string]bool{
	"f": true, "rs": true, "g": true, "q": true, "blur": true,
}

// signingParams are consumed by the proxy itself and must never be forwarded
// to the origin or embedded in the imgproxy URL.
var signingParams = map[string]bool{
	"exp": true, "sig": true,
}

var imageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true,
	".webp": true, ".avif": true, ".tiff": true, ".bmp": true,
	".svg": true, ".ico": true,
}

func hasImageExtension(path string) bool {
	dot := strings.LastIndex(path, ".")
	if dot == -1 {
		return false
	}
	return imageExtensions[strings.ToLower(path[dot:])]
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
		return
	}

	// Use the Host header to reconstruct the original image URL transparently.
	// Strip port if present (e.g. "cdn.examplesite.com:443" → "cdn.examplesite.com").
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	// Sanitize path to prevent directory traversal (e.g. /../../../etc/passwd → /etc/passwd).
	cleanPath := path.Clean(r.URL.Path)
	if cleanPath == "." {
		cleanPath = "/"
	}

	if !isAllowed(host) {
		http.Error(w, "origin not allowed: "+host, http.StatusForbidden)
		return
	}

	q := r.URL.Query()

	// Dual-source routing: a valid exp+sig pair routes to the private S3
	// bucket; no signature routes to the public source; an invalid or
	// expired signature is a hard 403 — never a fallback to public.
	privateRoute := false
	if signingKey != "" {
		ok, present := verifySignature(signingKey, cleanPath, q.Get("exp"), q.Get("sig"))
		if present && !ok {
			http.Error(w, "invalid or expired signature", http.StatusForbidden)
			return
		}
		privateRoute = ok
	}

	var originalURL string
	if privateRoute {
		originalURL = s3.presignGetObject(cleanPath, presignTTL, time.Now())
	} else if sourceBaseURL != "" {
		originalURL = sourceBaseURL + cleanPath
	} else {
		originalURL = "https://" + host + cleanPath
	}

	// Strip imgproxy and signing params from query, forward the rest to the origin.
	remaining := url.Values{}
	for k, vs := range q {
		if !imgproxyParams[k] && !signingParams[k] {
			for _, v := range vs {
				remaining.Add(k, v)
			}
		}
	}
	if len(remaining) > 0 {
		sep := "?"
		if strings.Contains(originalURL, "?") {
			sep = "&"
		}
		originalURL += sep + remaining.Encode()
	}

	auto := ""
	if q.Get("f") == "" {
		auto = autoFormat(cleanPath)
	}
	processingOptions := buildProcessingOptions(q, auto)

	// Route through imgproxy only when transformation params are present on an image.
	// Everything else (HTML, JS, CSS, images without params) passes directly to the origin.
	var targetURL string
	if processingOptions != "" && hasImageExtension(cleanPath) {
		source := originalURL
		if privateRoute {
			// Presigned URLs contain ? and & — escape before embedding in the imgproxy path.
			source = url.QueryEscape(originalURL)
		}
		targetURL = imgproxyURL + fmt.Sprintf("/insecure/%s/plain/%s", processingOptions, source)
	} else {
		targetURL = originalURL
	}

	log.Printf("%s %s (host: %s) → %s", r.Method, r.URL.RequestURI(), host, targetURL)

	proxyReq, err := http.NewRequest(r.Method, targetURL, nil)
	if err != nil {
		http.Error(w, "failed to build request", http.StatusInternalServerError)
		return
	}
	proxyReq.Header = r.Header.Clone()

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, "failed to reach imgproxy", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	if privateRoute {
		// Signed URLs expire — responses must not be shared or cached long.
		w.Header().Set("Cache-Control", "private, max-age=900")
	}
	w.WriteHeader(resp.StatusCode)

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}

func main() {
	if err := validateConfig(); err != nil {
		log.Fatal(err)
	}

	if len(allowedOrigins) == 0 {
		log.Println("WARNING: ALLOWED_ORIGINS not set — all origins are allowed (not recommended in production)")
	} else {
		origins := make([]string, 0, len(allowedOrigins))
		for o := range allowedOrigins {
			origins = append(origins, o)
		}
		log.Printf("allowed origins: %s", strings.Join(origins, ", "))
	}

	log.Printf("img-fwd starting on :%s", port)
	log.Printf("  imgproxy → %s", imgproxyURL)
	if sourceBaseURL != "" {
		log.Printf("  source base URL → %s", sourceBaseURL)
	}
	if signingKey != "" {
		log.Printf("  dual-source mode → signed requests routed to s3://%s (via %s)", s3.bucket, s3.endpoint)
	}

	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

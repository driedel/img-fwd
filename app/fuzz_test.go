package main

import (
	"net/url"
	"testing"
	"time"
)

// FuzzVerifySignature ensures the signature verifier never panics and never
// validates a wrong signature, regardless of input shape.
func FuzzVerifySignature(f *testing.F) {
	validExp := signedExp(15 * time.Minute)
	validSig := computeSignature(testSigningKey, "/img.jpg", validExp)
	f.Add("/img.jpg", validExp, validSig)
	f.Add("/img.jpg", "not-a-number", "deadbeef")
	f.Add("", "", "")
	f.Add("/img.jpg", "99999999999999999999999999", validSig)
	f.Add("/a b/cñ.jpg", "-1", "00")
	f.Add("/img.jpg", validExp, "")

	f.Fuzz(func(t *testing.T, path, exp, sig string) {
		ok, _ := verifySignature(testSigningKey, path, exp, sig)
		if ok {
			// Whenever it validates, recompute to confirm it is genuinely correct.
			want := computeSignature(testSigningKey, path, exp)
			if sig != want {
				t.Fatalf("invalid signature accepted: path=%q exp=%q sig=%q", path, exp, sig)
			}
		}
	})
}

// FuzzBuildProcessingOptions ensures query-param parsing never panics.
func FuzzBuildProcessingOptions(f *testing.F) {
	f.Add("rs=600", "avif")
	f.Add("f=webp&rs=600:400&q=80&g=sm&blur=5", "")
	f.Add("rs=:", "avif")
	f.Add("rs=:::", "")
	f.Add("f==&q=%zz", "webp")

	f.Fuzz(func(t *testing.T, rawQuery, format string) {
		q, err := url.ParseQuery(rawQuery)
		if err != nil {
			q = url.Values{}
		}
		_ = buildProcessingOptions(q, format)
	})
}

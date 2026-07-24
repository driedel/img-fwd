package main

import (
	"fmt"
	"testing"
	"time"
)

const testSigningKey = "test-secret-key"

func signedExp(offset time.Duration) string {
	return fmt.Sprintf("%d", time.Now().Add(offset).Unix())
}

func TestVerifySignatureValid(t *testing.T) {
	exp := signedExp(15 * time.Minute)
	sig := computeSignature(testSigningKey, "/img.jpg", exp)
	ok, present := verifySignature(testSigningKey, "/img.jpg", exp, sig)
	if !ok || !present {
		t.Errorf("expected valid signature (ok=true, present=true), got ok=%v present=%v", ok, present)
	}
}

func TestVerifySignatureInvalid(t *testing.T) {
	exp := signedExp(15 * time.Minute)
	ok, present := verifySignature(testSigningKey, "/img.jpg", exp, "deadbeef")
	if ok || !present {
		t.Errorf("expected invalid signature (ok=false, present=true), got ok=%v present=%v", ok, present)
	}
}

func TestVerifySignatureWrongPath(t *testing.T) {
	exp := signedExp(15 * time.Minute)
	sig := computeSignature(testSigningKey, "/other.jpg", exp)
	ok, present := verifySignature(testSigningKey, "/img.jpg", exp, sig)
	if ok || !present {
		t.Errorf("signature for a different path must not validate, got ok=%v present=%v", ok, present)
	}
}

func TestVerifySignatureExpired(t *testing.T) {
	exp := signedExp(-time.Minute)
	sig := computeSignature(testSigningKey, "/img.jpg", exp)
	ok, present := verifySignature(testSigningKey, "/img.jpg", exp, sig)
	if ok || !present {
		t.Errorf("expected expired signature to fail (ok=false, present=true), got ok=%v present=%v", ok, present)
	}
}

func TestVerifySignatureAbsent(t *testing.T) {
	ok, present := verifySignature(testSigningKey, "/img.jpg", "", "")
	if ok || present {
		t.Errorf("expected absent signature (ok=false, present=false), got ok=%v present=%v", ok, present)
	}
}

func TestVerifySignaturePartial(t *testing.T) {
	exp := signedExp(15 * time.Minute)
	sig := computeSignature(testSigningKey, "/img.jpg", exp)

	if ok, present := verifySignature(testSigningKey, "/img.jpg", exp, ""); ok || !present {
		t.Errorf("sig missing with exp present must be present=true ok=false, got ok=%v present=%v", ok, present)
	}
	if ok, present := verifySignature(testSigningKey, "/img.jpg", "", sig); ok || !present {
		t.Errorf("exp missing with sig present must be present=true ok=false, got ok=%v present=%v", ok, present)
	}
}

func TestVerifySignatureMalformedExp(t *testing.T) {
	sig := computeSignature(testSigningKey, "/img.jpg", "not-a-number")
	ok, present := verifySignature(testSigningKey, "/img.jpg", "not-a-number", sig)
	if ok || !present {
		t.Errorf("expected malformed exp to fail (ok=false, present=true), got ok=%v present=%v", ok, present)
	}
}

func TestVerifySignatureOverflowExp(t *testing.T) {
	huge := "99999999999999999999999999"
	sig := computeSignature(testSigningKey, "/img.jpg", huge)
	ok, present := verifySignature(testSigningKey, "/img.jpg", huge, sig)
	if ok || !present {
		t.Errorf("expected overflowing exp to fail (ok=false, present=true), got ok=%v present=%v", ok, present)
	}
}

func TestVerifySignatureNegativeExp(t *testing.T) {
	sig := computeSignature(testSigningKey, "/img.jpg", "-1")
	ok, present := verifySignature(testSigningKey, "/img.jpg", "-1", sig)
	if ok || !present {
		t.Errorf("expected negative exp to fail (ok=false, present=true), got ok=%v present=%v", ok, present)
	}
}

func TestComputeSignatureDeterministic(t *testing.T) {
	a := computeSignature(testSigningKey, "/img.jpg", "1750000000")
	b := computeSignature(testSigningKey, "/img.jpg", "1750000000")
	if a != b {
		t.Errorf("same input must produce same signature: %q != %q", a, b)
	}
	c := computeSignature("other-key", "/img.jpg", "1750000000")
	if a == c {
		t.Error("different keys must produce different signatures")
	}
}

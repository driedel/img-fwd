package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

func computeSignature(key, path, exp string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(path + "\n" + exp))
	return hex.EncodeToString(mac.Sum(nil))
}

// verifySignature checks the exp/sig query params against the signing key.
// present reports whether signature params were supplied at all — callers use
// it to distinguish "unsigned request" (route to public source) from "invalid
// signature" (reject with 403).
func verifySignature(key, path string, exp, sig string) (ok bool, present bool) {
	if exp == "" && sig == "" {
		return false, false
	}
	if exp == "" || sig == "" {
		return false, true
	}
	expTS, err := strconv.ParseInt(exp, 10, 64)
	if err != nil || expTS < time.Now().Unix() {
		return false, true
	}
	expected := computeSignature(key, path, exp)
	return hmac.Equal([]byte(expected), []byte(sig)), true
}

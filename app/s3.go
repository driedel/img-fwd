package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type s3Config struct {
	endpoint  string
	bucket    string
	accessKey string
	secretKey string
	region    string
	useSSL    bool
}

func (c s3Config) enabled() bool {
	return c.endpoint != ""
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// encodeS3Key URI-encodes each path segment of an object key, preserving "/".
func encodeS3Key(key string) string {
	segments := strings.Split(key, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return strings.Join(segments, "/")
}

// presignGetObject builds an AWS Signature Version 4 presigned URL (path-style)
// for a GetObject request. Works with MinIO and AWS S3.
func (c s3Config) presignGetObject(key string, ttl time.Duration, now time.Time) string {
	scheme := "http"
	if c.useSSL {
		scheme = "https"
	}
	host := c.endpoint
	uri := "/" + c.bucket + encodeS3Key(key)

	amzDate := now.UTC().Format("20060102T150405Z")
	dateStamp := now.UTC().Format("20060102")
	scope := dateStamp + "/" + c.region + "/s3/aws4_request"
	credential := c.accessKey + "/" + scope

	q := url.Values{}
	q.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	q.Set("X-Amz-Credential", credential)
	q.Set("X-Amz-Date", amzDate)
	q.Set("X-Amz-Expires", fmt.Sprintf("%d", int64(ttl.Seconds())))
	q.Set("X-Amz-SignedHeaders", "host")
	canonicalQuery := q.Encode()

	canonicalRequest := "GET\n" +
		uri + "\n" +
		canonicalQuery + "\n" +
		"host:" + host + "\n" + "\n" +
		"host\n" +
		"UNSIGNED-PAYLOAD"

	requestHash := sha256.Sum256([]byte(canonicalRequest))
	stringToSign := "AWS4-HMAC-SHA256\n" +
		amzDate + "\n" +
		scope + "\n" +
		hex.EncodeToString(requestHash[:])

	kDate := hmacSHA256([]byte("AWS4"+c.secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, c.region)
	kService := hmacSHA256(kRegion, "s3")
	kSigning := hmacSHA256(kService, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))

	return fmt.Sprintf("%s://%s%s?%s&X-Amz-Signature=%s", scheme, host, uri, canonicalQuery, signature)
}

package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const s3ServiceName = "s3"

// S3Prober verifies conservative S3-compatible object operations.
type S3Prober struct {
	client *http.Client
	now    func() time.Time
}

const s3ProbeTimeout = 30 * time.Second

func NewS3Prober() *S3Prober {
	return &S3Prober{client: &http.Client{Timeout: s3ProbeTimeout}, now: time.Now}
}

func (prober *S3Prober) Probe(ctx context.Context, backend StorageBackend) error {
	config, ok := s3ProbeConfig(backend)
	if !ok {
		return fmt.Errorf("%w: backend type %q does not expose an s3-compatible adapter", ErrProbeUnsupported, backend.Type)
	}
	accessKey, secretKey, err := resolveS3ProbeCredentials(config)
	if err != nil {
		return err
	}
	region := strings.TrimSpace(config.Region)
	if region == "" {
		region = "auto"
	}

	objectKey := s3ProbeObjectKey(backend.ID, prober.now().UTC())
	cleanup := false
	if err := prober.do(ctx, http.MethodPut, config, region, accessKey, secretKey, objectKey, "", bytes.NewReader(probePayload), nil); err != nil {
		return err
	}
	cleanup = true
	defer func() {
		if cleanup {
			_ = prober.do(context.Background(), http.MethodDelete, config, region, accessKey, secretKey, objectKey, "", nil, nil)
		}
	}()

	var fullRead []byte
	if err := prober.do(ctx, http.MethodGet, config, region, accessKey, secretKey, objectKey, "", nil, &fullRead); err != nil {
		return err
	}
	if !bytes.Equal(fullRead, probePayload) {
		return fmt.Errorf("%w: s3 full read content mismatch", ErrProbeFailed)
	}

	var rangeRead []byte
	if err := prober.do(ctx, http.MethodGet, config, region, accessKey, secretKey, objectKey, "bytes=6-10", nil, &rangeRead); err != nil {
		return err
	}
	if !bytes.Equal(rangeRead, probePayload[6:11]) {
		return fmt.Errorf("%w: s3 range read content mismatch", ErrProbeFailed)
	}

	if err := prober.do(ctx, http.MethodDelete, config, region, accessKey, secretKey, objectKey, "", nil, nil); err != nil {
		return err
	}
	cleanup = false
	return nil
}

type s3CompatibleProbeConfig struct {
	Endpoint           string
	Region             string
	Bucket             string
	PathStyle          bool
	AccessKeySecretRef string
	SecretKeySecretRef string
}

func s3ProbeConfig(backend StorageBackend) (s3CompatibleProbeConfig, bool) {
	switch backend.Type {
	case BackendTypeS3:
		if backend.Config.S3 == nil {
			return s3CompatibleProbeConfig{}, false
		}
		return s3CompatibleProbeConfig{
			Endpoint:           backend.Config.S3.Endpoint,
			Region:             backend.Config.S3.Region,
			Bucket:             backend.Config.S3.Bucket,
			PathStyle:          backend.Config.S3.PathStyle,
			AccessKeySecretRef: backend.Config.S3.AccessKeySecretRef,
			SecretKeySecretRef: backend.Config.S3.SecretKeySecretRef,
		}, true
	case BackendTypeDistributed:
		if backend.Config.Distributed == nil || backend.Config.Distributed.Adapter != "s3-compatible" {
			return s3CompatibleProbeConfig{}, false
		}
		return s3CompatibleProbeConfig{
			Endpoint:           backend.Config.Distributed.Endpoint,
			Region:             backend.Config.Distributed.Region,
			Bucket:             backend.Config.Distributed.Bucket,
			PathStyle:          backend.Config.Distributed.PathStyle,
			AccessKeySecretRef: backend.Config.Distributed.AccessKeySecretRef,
			SecretKeySecretRef: backend.Config.Distributed.SecretKeySecretRef,
		}, true
	default:
		return s3CompatibleProbeConfig{}, false
	}
}

func resolveS3ProbeCredentials(config s3CompatibleProbeConfig) (string, string, error) {
	accessRef := strings.TrimSpace(config.AccessKeySecretRef)
	secretRef := strings.TrimSpace(config.SecretKeySecretRef)
	if accessRef == "" || secretRef == "" {
		return "", "", fmt.Errorf("%w: s3 credential secret references are required", ErrProbeFailed)
	}
	accessKey, ok := os.LookupEnv(accessRef)
	if !ok || accessKey == "" {
		return "", "", fmt.Errorf("%w: s3 access key secret %q is not configured", ErrProbeFailed, accessRef)
	}
	secretKey, ok := os.LookupEnv(secretRef)
	if !ok || secretKey == "" {
		return "", "", fmt.Errorf("%w: s3 secret key secret %q is not configured", ErrProbeFailed, secretRef)
	}
	return accessKey, secretKey, nil
}

func s3ProbeObjectKey(backendID string, now time.Time) string {
	cleanID := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, backendID)
	return fmt.Sprintf(".inori-music-probe/%s-%d", cleanID, now.UnixNano())
}

func (prober *S3Prober) do(ctx context.Context, method string, config s3CompatibleProbeConfig, region string, accessKey string, secretKey string, objectKey string, byteRange string, body io.Reader, responseBody *[]byte) error {
	payload := []byte(nil)
	if body != nil {
		var err error
		payload, err = io.ReadAll(body)
		if err != nil {
			return fmt.Errorf("%w: read s3 request payload: %v", ErrProbeFailed, err)
		}
	}
	requestURL, err := s3ObjectURL(config, objectKey)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("%w: build s3 request: %v", ErrProbeFailed, err)
	}
	if byteRange != "" {
		req.Header.Set("Range", byteRange)
	}
	signS3Request(req, region, accessKey, secretKey, payload, prober.now().UTC())

	resp, err := prober.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: s3 %s request: %v", ErrProbeFailed, strings.ToLower(method), err)
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: read s3 %s response: %v", ErrProbeFailed, strings.ToLower(method), err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: s3 %s returned status %d", ErrProbeFailed, strings.ToLower(method), resp.StatusCode)
	}
	if responseBody != nil {
		*responseBody = content
	}
	return nil
}

func s3ObjectURL(config s3CompatibleProbeConfig, objectKey string) (*url.URL, error) {
	endpoint, err := url.Parse(strings.TrimRight(config.Endpoint, "/"))
	if err != nil || endpoint.Scheme == "" || endpoint.Host == "" {
		return nil, fmt.Errorf("%w: s3 endpoint must be an absolute URL", ErrProbeFailed)
	}
	bucket := strings.TrimSpace(config.Bucket)
	if bucket == "" {
		return nil, fmt.Errorf("%w: s3 bucket is required", ErrProbeFailed)
	}
	objectPath := escapeS3Path(objectKey)
	if config.PathStyle {
		endpoint.Path = strings.TrimRight(endpoint.EscapedPath(), "/") + "/" + escapeS3Path(bucket) + "/" + objectPath
		return endpoint, nil
	}
	endpoint.Host = bucket + "." + endpoint.Host
	endpoint.Path = strings.TrimRight(endpoint.EscapedPath(), "/") + "/" + objectPath
	return endpoint, nil
}

func escapeS3Path(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

// presignS3URL generates an AWS Signature Version 4 presigned GET URL for a
// single object. The URL is valid for the given TTL and does not require the
// caller to supply any credentials at request time.
func presignS3URL(config s3CompatibleProbeConfig, objectKey string, accessKey string, secretKey string, ttl time.Duration, now time.Time) (string, error) {
	base, err := s3ObjectURL(config, objectKey)
	if err != nil {
		return "", err
	}
	region := strings.TrimSpace(config.Region)
	if region == "" {
		region = "us-east-1"
	}
	amzDate := now.UTC().Format("20060102T150405Z")
	dateStamp := now.UTC().Format("20060102")
	credentialScope := dateStamp + "/" + region + "/" + s3ServiceName + "/aws4_request"
	credential := accessKey + "/" + credentialScope
	expires := fmt.Sprintf("%d", int(ttl.Seconds()))
	host := base.Host

	// Build canonical query string (must be alphabetically sorted).
	q := url.Values{}
	q.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	q.Set("X-Amz-Credential", credential)
	q.Set("X-Amz-Date", amzDate)
	q.Set("X-Amz-Expires", expires)
	q.Set("X-Amz-SignedHeaders", "host")
	canonicalQueryString := q.Encode() // url.Values.Encode() sorts keys

	canonicalURI := base.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalHeaders := "host:" + host + "\n"
	signedHeaders := "host"
	payloadHash := "UNSIGNED-PAYLOAD"

	canonicalRequest := strings.Join([]string{
		http.MethodGet,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + credentialScope + "\n" + sha256Hex([]byte(canonicalRequest))
	signingKey := s3SigningKey(secretKey, dateStamp, region)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))

	base.RawQuery = canonicalQueryString + "&X-Amz-Signature=" + signature
	return base.String(), nil
}

func signS3Request(req *http.Request, region string, accessKey string, secretKey string, payload []byte, now time.Time) {
	payloadHash := sha256Hex(payload)
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	credentialScope := dateStamp + "/" + region + "/" + s3ServiceName + "/aws4_request"
	canonicalRequest, signedHeaders := canonicalS3Request(req, payloadHash)
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + credentialScope + "\n" + sha256Hex([]byte(canonicalRequest))
	signingKey := s3SigningKey(secretKey, dateStamp, region)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+accessKey+"/"+credentialScope+", SignedHeaders="+signedHeaders+", Signature="+signature)
}

func canonicalS3Request(req *http.Request, payloadHash string) (string, string) {
	headers := map[string]string{
		"host":                 req.URL.Host,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           req.Header.Get("X-Amz-Date"),
	}
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)
	var canonicalHeaders strings.Builder
	for _, name := range names {
		canonicalHeaders.WriteString(name)
		canonicalHeaders.WriteByte(':')
		canonicalHeaders.WriteString(strings.TrimSpace(headers[name]))
		canonicalHeaders.WriteByte('\n')
	}
	signedHeaders := strings.Join(names, ";")
	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	return req.Method + "\n" + canonicalURI + "\n" + req.URL.RawQuery + "\n" + canonicalHeaders.String() + "\n" + signedHeaders + "\n" + payloadHash, signedHeaders
}

func s3SigningKey(secretKey string, dateStamp string, region string) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	dateRegionKey := hmacSHA256(dateKey, region)
	dateRegionServiceKey := hmacSHA256(dateRegionKey, s3ServiceName)
	return hmacSHA256(dateRegionServiceKey, "aws4_request")
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write([]byte(data))
	return h.Sum(nil)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

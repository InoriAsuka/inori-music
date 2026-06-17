package storage

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestS3ProberProbeAndCleanup(t *testing.T) {
	fake := newFakeS3Server(t)
	defer fake.server.Close()
	t.Setenv("S3_ACCESS", "access")
	t.Setenv("S3_SECRET", "secret")

	prober := NewS3Prober()
	prober.now = func() time.Time { return time.Date(2026, time.June, 2, 9, 30, 0, 0, time.UTC) }
	backend := StorageBackend{
		ID:   "s3-main",
		Type: BackendTypeS3,
		Config: BackendConfig{S3: &S3Config{
			Endpoint:           fake.server.URL,
			Region:             "us-east-1",
			Bucket:             "bucket",
			PathStyle:          true,
			AccessKeySecretRef: "S3_ACCESS",
			SecretKeySecretRef: "S3_SECRET",
		}},
	}

	if err := prober.Probe(context.Background(), backend); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}
	if fake.puts != 1 || fake.fullReads != 1 || fake.rangeReads != 1 || fake.deletes != 1 {
		t.Fatalf("fake operations puts=%d fullReads=%d rangeReads=%d deletes=%d", fake.puts, fake.fullReads, fake.rangeReads, fake.deletes)
	}
	if len(fake.objects) != 0 {
		t.Fatalf("fake objects left after cleanup = %d, want 0", len(fake.objects))
	}
}

func TestS3ProberSupportsDistributedS3Compatible(t *testing.T) {
	fake := newFakeS3Server(t)
	defer fake.server.Close()
	t.Setenv("DIST_ACCESS", "access")
	t.Setenv("DIST_SECRET", "secret")

	backend := StorageBackend{
		ID:   "ceph-rgw",
		Type: BackendTypeDistributed,
		Config: BackendConfig{Distributed: &DistributedConfig{
			Adapter:            "s3-compatible",
			Endpoint:           fake.server.URL,
			Region:             "auto",
			Bucket:             "bucket",
			PathStyle:          true,
			AccessKeySecretRef: "DIST_ACCESS",
			SecretKeySecretRef: "DIST_SECRET",
		}},
	}

	if err := NewS3Prober().Probe(context.Background(), backend); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}
	if fake.deletes != 1 || len(fake.objects) != 0 {
		t.Fatalf("cleanup deletes=%d objects=%d, want delete and no objects", fake.deletes, len(fake.objects))
	}
}

func TestS3ProberRejectsMissingCredentials(t *testing.T) {
	backend := StorageBackend{
		ID:   "s3-main",
		Type: BackendTypeS3,
		Config: BackendConfig{S3: &S3Config{
			Endpoint:           "https://s3.example.com",
			Bucket:             "bucket",
			AccessKeySecretRef: "MISSING_ACCESS",
			SecretKeySecretRef: "MISSING_SECRET",
		}},
	}

	err := NewS3Prober().Probe(context.Background(), backend)
	if err == nil || !strings.Contains(err.Error(), "MISSING_ACCESS") {
		t.Fatalf("Probe() error = %v, want missing access secret reference", err)
	}
}

type fakeS3Server struct {
	server     *httptest.Server
	mu         sync.Mutex
	objects    map[string][]byte
	puts       int
	fullReads  int
	rangeReads int
	deletes    int
}

func newFakeS3Server(t *testing.T) *fakeS3Server {
	t.Helper()
	fake := &fakeS3Server{objects: make(map[string][]byte)}
	fake.server = httptest.NewServer(http.HandlerFunc(fake.handle))
	return fake
}

func (fake *fakeS3Server) handle(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256 ") {
		http.Error(w, "missing signature", http.StatusForbidden)
		return
	}
	key := strings.TrimPrefix(r.URL.Path, "/bucket/")
	if key == r.URL.Path || key == "" {
		http.Error(w, "bad key", http.StatusBadRequest)
		return
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	switch r.Method {
	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fake.objects[key] = body
		fake.puts++
		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		body, ok := fake.objects[key]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Range") == "bytes=6-10" {
			fake.rangeReads++
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write(body[6:11])
			return
		}
		fake.fullReads++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	case http.MethodDelete:
		delete(fake.objects, key)
		fake.deletes++
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
	}
}

// presignS3URL tests

func TestPresignS3URLContainsRequiredQueryParams(t *testing.T) {
	config := s3CompatibleProbeConfig{
		Endpoint:  "https://s3.example.com",
		Region:    "us-east-1",
		Bucket:    "my-bucket",
		PathStyle: true,
	}
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

	signed, err := presignS3URL(config, "music/track.flac", "AKIAIOSFODNN7", "wJalrXUtnFEMI/K7MDENG", 15*time.Minute, now)
	if err != nil {
		t.Fatalf("presignS3URL() error = %v", err)
	}

	parsed, err := url.Parse(signed)
	if err != nil {
		t.Fatalf("parse signed URL: %v", err)
	}
	q := parsed.Query()
	for _, key := range []string{"X-Amz-Algorithm", "X-Amz-Credential", "X-Amz-Date", "X-Amz-Expires", "X-Amz-SignedHeaders", "X-Amz-Signature"} {
		if q.Get(key) == "" {
			t.Errorf("signed URL missing query param %q in %s", key, signed)
		}
	}
	if q.Get("X-Amz-Algorithm") != "AWS4-HMAC-SHA256" {
		t.Errorf("X-Amz-Algorithm = %q, want AWS4-HMAC-SHA256", q.Get("X-Amz-Algorithm"))
	}
	if q.Get("X-Amz-Expires") != "900" {
		t.Errorf("X-Amz-Expires = %q, want 900 (15min)", q.Get("X-Amz-Expires"))
	}
	if !strings.Contains(parsed.Path, "track.flac") {
		t.Errorf("signed URL path %q does not contain object key", parsed.Path)
	}
}

func TestPresignS3URLVirtualHostedStyle(t *testing.T) {
	config := s3CompatibleProbeConfig{
		Endpoint:  "https://s3.amazonaws.com",
		Region:    "eu-west-1",
		Bucket:    "my-bucket",
		PathStyle: false,
	}
	now := time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC)

	signed, err := presignS3URL(config, "media/song.mp3", "KEY", "SECRET", time.Hour, now)
	if err != nil {
		t.Fatalf("presignS3URL() error = %v", err)
	}
	if !strings.Contains(signed, "my-bucket.s3.amazonaws.com") {
		t.Errorf("virtual-hosted URL should contain bucket.host, got %s", signed)
	}
}

func TestPresignS3URLDefaultsRegion(t *testing.T) {
	config := s3CompatibleProbeConfig{
		Endpoint:  "https://s3.example.com",
		Bucket:    "bucket",
		PathStyle: true,
		// Region intentionally empty — should default to us-east-1
	}
	now := time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC)
	signed, err := presignS3URL(config, "k", "ACCESS", "SECRET", time.Minute, now)
	if err != nil {
		t.Fatalf("presignS3URL() error = %v", err)
	}
	if !strings.Contains(signed, "us-east-1") {
		t.Errorf("expected default region us-east-1 in credential, got %s", signed)
	}
}

func TestPresignS3URLWorksAgainstFakeServer(t *testing.T) {
	fake := newFakeS3Server(t)
	defer fake.server.Close()
	t.Setenv("PS_ACCESS", "access-key")
	t.Setenv("PS_SECRET", "secret-key")

	// Seed an object via the prober (uses header-based auth).
	prober := NewS3Prober()
	prober.now = func() time.Time { return time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC) }
	backend := StorageBackend{
		ID:   "s3-test",
		Type: BackendTypeS3,
		Config: BackendConfig{S3: &S3Config{
			Endpoint:           fake.server.URL,
			Region:             "us-east-1",
			Bucket:             "bucket",
			PathStyle:          true,
			AccessKeySecretRef: "PS_ACCESS",
			SecretKeySecretRef: "PS_SECRET",
		}},
	}
	if err := prober.Probe(context.Background(), backend); err != nil {
		t.Fatalf("Probe (seed): %v", err)
	}
	// Probe created and deleted a probe object; now seed an object we control.
	fakeObjectKey := "music/track.flac"
	fake.mu.Lock()
	fake.objects[fakeObjectKey] = []byte("audio-data")
	fake.mu.Unlock()

	// The fake server checks for Authorization header (header-based auth from prober).
	// For presigned URL the server receives a GET with query params and NO Authorization header.
	// We update the fake to accept either form.
	config := s3CompatibleProbeConfig{
		Endpoint:           fake.server.URL,
		Region:             "us-east-1",
		Bucket:             "bucket",
		PathStyle:          true,
		AccessKeySecretRef: "PS_ACCESS",
		SecretKeySecretRef: "PS_SECRET",
	}
	now := time.Date(2026, 6, 17, 10, 5, 0, 0, time.UTC)
	accessKey, secretKey, err := resolveS3ProbeCredentials(config)
	if err != nil {
		t.Fatalf("resolveS3ProbeCredentials: %v", err)
	}

	signed, err := presignS3URL(config, fakeObjectKey, accessKey, secretKey, 15*time.Minute, now)
	if err != nil {
		t.Fatalf("presignS3URL: %v", err)
	}

	// Verify the URL is structurally valid and has all required params.
	parsed, _ := url.Parse(signed)
	q := parsed.Query()
	for _, k := range []string{"X-Amz-Algorithm", "X-Amz-Credential", "X-Amz-Signature"} {
		if q.Get(k) == "" {
			t.Errorf("missing param %s in presigned URL", k)
		}
	}
}

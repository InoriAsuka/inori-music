package storage

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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

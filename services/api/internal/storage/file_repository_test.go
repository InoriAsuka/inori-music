package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileRepositoryPersistsBackendsAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "nested", "storage-backends.json")
	repo, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("NewFileRepository() error = %v", err)
	}
	checkedAt := time.Date(2026, time.June, 3, 1, 2, 3, 0, time.UTC)
	backend := StorageBackend{
		ID:                "local-main",
		Type:              BackendTypeLocal,
		DisplayName:       "Local",
		Enabled:           true,
		IsDefault:         true,
		Priority:          10,
		HealthStatus:      HealthStatusHealthy,
		LastHealthCheckAt: &checkedAt,
		LastCapacity:      &CapacityReport{BackendID: "local-main", TotalBytes: 100, AvailableBytes: 25, UsedBytes: 75, CheckedAt: checkedAt},
		Config:            BackendConfig{Local: &LocalConfig{RootPath: "/srv/inori/media"}},
	}
	if err := repo.Save(ctx, backend); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	reopened, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("NewFileRepository(reopen) error = %v", err)
	}
	got, err := reopened.Get(ctx, "local-main")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != backend.ID || got.HealthStatus != HealthStatusHealthy || got.LastHealthCheckAt == nil || got.LastCapacity == nil {
		t.Fatalf("persisted backend = %+v, want health and capacity state", got)
	}
}

func TestFileRepositoryClearDefaultPersists(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "storage-backends.json")
	repo, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("NewFileRepository() error = %v", err)
	}
	if err := repo.Save(ctx, validPersistedBackend("b", 2, true)); err != nil {
		t.Fatalf("Save(b) error = %v", err)
	}
	if err := repo.Save(ctx, validPersistedBackend("a", 1, false)); err != nil {
		t.Fatalf("Save(a) error = %v", err)
	}
	if err := repo.ClearDefault(ctx); err != nil {
		t.Fatalf("ClearDefault() error = %v", err)
	}

	reopened, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("NewFileRepository(reopen) error = %v", err)
	}
	backends, err := reopened.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(backends) != 2 || backends[0].ID != "a" || backends[1].ID != "b" {
		t.Fatalf("List() = %+v, want stable priority order", backends)
	}
	for _, backend := range backends {
		if backend.IsDefault {
			t.Fatalf("backend %s is still default after persisted ClearDefault", backend.ID)
		}
	}
}

func TestFileRepositoryRejectsMalformedDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "storage-backends.json")
	if err := os.WriteFile(path, []byte(`{"version":`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := NewFileRepository(path); err == nil {
		t.Fatal("NewFileRepository() error = nil, want malformed JSON error")
	}
}

func TestFileRepositoryRejectsUnsupportedSchemaVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "storage-backends.json")
	if err := os.WriteFile(path, []byte(`{"version":99,"backends":[]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := NewFileRepository(path)
	if !errors.Is(err, ErrInvalidBackend) {
		t.Fatalf("NewFileRepository() error = %v, want ErrInvalidBackend", err)
	}
}

func TestFileRepositoryRejectsDuplicateBackendIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "storage-backends.json")
	document := fileRepositoryDocument{
		Version: fileRepositorySchemaVersion,
		Backends: []StorageBackend{
			validPersistedBackend(" local-main ", 1, false),
			validPersistedBackend("local-main", 2, false),
		},
	}
	content, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = NewFileRepository(path)
	if !errors.Is(err, ErrInvalidBackend) {
		t.Fatalf("NewFileRepository() error = %v, want ErrInvalidBackend", err)
	}
}

func TestFileRepositoryRejectsInvalidPersistedBackendConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "storage-backends.json")
	document := fileRepositoryDocument{
		Version: fileRepositorySchemaVersion,
		Backends: []StorageBackend{
			{ID: "local-main", Type: BackendTypeLocal, DisplayName: "Local", Enabled: true, Config: BackendConfig{S3: &S3Config{Endpoint: "https://s3.example.com", Bucket: "inori", AccessKeySecretRef: "access", SecretKeySecretRef: "secret"}}},
		},
	}
	content, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = NewFileRepository(path)
	if !errors.Is(err, ErrInvalidBackend) {
		t.Fatalf("NewFileRepository() error = %v, want ErrInvalidBackend", err)
	}
}

func TestFileRepositoryRejectsEmptyBackendID(t *testing.T) {
	repo, err := NewFileRepository(filepath.Join(t.TempDir(), "storage-backends.json"))
	if err != nil {
		t.Fatalf("NewFileRepository() error = %v", err)
	}
	err = repo.Save(context.Background(), StorageBackend{})
	if !errors.Is(err, ErrInvalidBackend) {
		t.Fatalf("Save(empty id) error = %v, want ErrInvalidBackend", err)
	}
}

func validPersistedBackend(id string, priority int, isDefault bool) StorageBackend {
	return StorageBackend{
		ID:          id,
		Type:        BackendTypeLocal,
		DisplayName: "Local " + id,
		Enabled:     true,
		IsDefault:   isDefault,
		Priority:    priority,
		Config:      BackendConfig{Local: &LocalConfig{RootPath: "/srv/inori/" + id}},
	}
}

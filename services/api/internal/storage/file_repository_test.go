package storage

import (
	"context"
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
	if err := repo.Save(ctx, StorageBackend{ID: "b", Type: BackendTypeLocal, IsDefault: true, Priority: 2}); err != nil {
		t.Fatalf("Save(b) error = %v", err)
	}
	if err := repo.Save(ctx, StorageBackend{ID: "a", Type: BackendTypeLocal, Priority: 1}); err != nil {
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

func TestFileRepositorySaveWithExclusiveDefaultPersistsSingleRewrite(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "storage-backends.json")
	repo, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("NewFileRepository() error = %v", err)
	}
	oldDefault := StorageBackend{ID: "local-main", Type: BackendTypeLocal, Enabled: true, IsDefault: true, Priority: 1}
	newDefault := StorageBackend{ID: "s3-prod", Type: BackendTypeS3, Enabled: true, Priority: 2}
	if err := repo.Save(ctx, oldDefault); err != nil {
		t.Fatalf("Save(local-main) error = %v", err)
	}
	if err := repo.Save(ctx, newDefault); err != nil {
		t.Fatalf("Save(s3-prod) error = %v", err)
	}

	if err := repo.SaveWithExclusiveDefault(ctx, newDefault); err != nil {
		t.Fatalf("SaveWithExclusiveDefault() error = %v", err)
	}

	reopened, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("NewFileRepository(reopen) error = %v", err)
	}
	backends, err := reopened.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	defaults := make([]string, 0)
	for _, backend := range backends {
		if backend.IsDefault {
			defaults = append(defaults, backend.ID)
		}
	}
	if len(defaults) != 1 || defaults[0] != "s3-prod" {
		t.Fatalf("defaults = %v, want [s3-prod]", defaults)
	}
}

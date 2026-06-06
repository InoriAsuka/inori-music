//go:build unix

package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSyncDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "repository.json")
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, []byte(`{"version":1}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if err := syncDirectory(filepath.Dir(path)); err != nil {
		t.Fatalf("syncDirectory() error = %v", err)
	}
}

func TestFileRepositoryPersistsAfterDirectorySync(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "nested", "storage-backends.json")
	repo, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("NewFileRepository() error = %v", err)
	}

	backend := StorageBackend{ID: "local-sync", Type: BackendTypeLocal, Priority: 1}
	if err := repo.Save(ctx, backend); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	reopened, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("NewFileRepository(reopen) error = %v", err)
	}
	got, err := reopened.Get(ctx, backend.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != backend.ID {
		t.Fatalf("persisted backend ID = %q, want %q", got.ID, backend.ID)
	}
}

func TestFileMediaObjectRepositoryPersistsAfterDirectorySync(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "nested", "media-objects.json")
	repo, err := NewFileMediaObjectRepository(path)
	if err != nil {
		t.Fatalf("NewFileMediaObjectRepository() error = %v", err)
	}

	object := validMediaObject()
	object.ID = "media-sync"
	if err := repo.SaveMediaObject(ctx, object); err != nil {
		t.Fatalf("SaveMediaObject() error = %v", err)
	}

	reopened, err := NewFileMediaObjectRepository(path)
	if err != nil {
		t.Fatalf("NewFileMediaObjectRepository(reopen) error = %v", err)
	}
	got, err := reopened.GetMediaObject(ctx, object.ID)
	if err != nil {
		t.Fatalf("GetMediaObject() error = %v", err)
	}
	if got.ID != object.ID || got.ObjectKey != object.ObjectKey {
		t.Fatalf("persisted object = %+v, want ID %q and object key %q", got, object.ID, object.ObjectKey)
	}
}

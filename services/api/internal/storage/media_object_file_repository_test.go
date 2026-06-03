package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileMediaObjectRepositoryPersistsObjectsAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "nested", "media-objects.json")
	repo, err := NewFileMediaObjectRepository(path)
	if err != nil {
		t.Fatalf("NewFileMediaObjectRepository() error = %v", err)
	}
	createdAt := time.Date(2026, time.June, 3, 7, 30, 0, 0, time.UTC)
	object := validMediaObject()
	object.CreatedAt = createdAt
	object.UpdatedAt = createdAt
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
	if got.ID != object.ID || got.ObjectKey != object.ObjectKey || !got.CreatedAt.Equal(createdAt) || !got.UpdatedAt.Equal(createdAt) {
		t.Fatalf("persisted object = %+v, want original metadata", got)
	}
}

func TestFileMediaObjectRepositoryListsPersistedObjectsInStableOrder(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "media-objects.json")
	repo, err := NewFileMediaObjectRepository(path)
	if err != nil {
		t.Fatalf("NewFileMediaObjectRepository() error = %v", err)
	}
	objects := []MediaObject{
		{ID: "z", BackendID: "local-main", ObjectKey: "z.flac", ContentHash: "sha256:same", MIMEType: "audio/flac", AssetKind: string(AssetKindOriginalAudio), LifecycleState: string(LifecycleStateActive)},
		{ID: "a", BackendID: "local-main", ObjectKey: "a.flac", ContentHash: "sha256:same", MIMEType: "audio/flac", AssetKind: string(AssetKindOriginalAudio), LifecycleState: string(LifecycleStateActive)},
		{ID: "other", BackendID: "archive", ObjectKey: "other.flac", ContentHash: "sha256:other", MIMEType: "audio/flac", AssetKind: string(AssetKindBackup), LifecycleState: string(LifecycleStateArchived)},
	}
	for _, object := range objects {
		if err := repo.SaveMediaObject(ctx, object); err != nil {
			t.Fatalf("SaveMediaObject(%s) error = %v", object.ID, err)
		}
	}

	reopened, err := NewFileMediaObjectRepository(path)
	if err != nil {
		t.Fatalf("NewFileMediaObjectRepository(reopen) error = %v", err)
	}
	byBackend, err := reopened.ListMediaObjectsByBackend(ctx, "local-main")
	if err != nil {
		t.Fatalf("ListMediaObjectsByBackend() error = %v", err)
	}
	if len(byBackend) != 2 || byBackend[0].ID != "a" || byBackend[1].ID != "z" {
		t.Fatalf("ListMediaObjectsByBackend() = %+v, want object-key order", byBackend)
	}
	byHash, err := reopened.ListMediaObjectsByContentHash(ctx, "sha256:same")
	if err != nil {
		t.Fatalf("ListMediaObjectsByContentHash() error = %v", err)
	}
	if len(byHash) != 2 || byHash[0].ID != "a" || byHash[1].ID != "z" {
		t.Fatalf("ListMediaObjectsByContentHash() = %+v, want object-key order", byHash)
	}
}

func TestFileMediaObjectRepositoryRejectsMalformedDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "media-objects.json")
	if err := os.WriteFile(path, []byte(`{"version":`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := NewFileMediaObjectRepository(path); err == nil {
		t.Fatal("NewFileMediaObjectRepository() error = nil, want malformed JSON error")
	}
}

func TestFileMediaObjectRepositoryRejectsUnsupportedSchemaVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "media-objects.json")
	if err := os.WriteFile(path, []byte(`{"version":99,"objects":[]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := NewFileMediaObjectRepository(path)
	if !errors.Is(err, ErrInvalidMediaObject) {
		t.Fatalf("NewFileMediaObjectRepository() error = %v, want ErrInvalidMediaObject", err)
	}
}

func TestFileMediaObjectRepositoryRejectsEmptyObjectID(t *testing.T) {
	repo, err := NewFileMediaObjectRepository(filepath.Join(t.TempDir(), "media-objects.json"))
	if err != nil {
		t.Fatalf("NewFileMediaObjectRepository() error = %v", err)
	}
	err = repo.SaveMediaObject(context.Background(), MediaObject{})
	if !errors.Is(err, ErrInvalidMediaObject) {
		t.Fatalf("SaveMediaObject(empty id) error = %v, want ErrInvalidMediaObject", err)
	}
}

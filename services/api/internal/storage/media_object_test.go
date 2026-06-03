package storage

import (
	"context"
	"errors"
	"testing"
)

func TestMediaObjectServiceRegistersObjectForEnabledBackend(t *testing.T) {
	ctx := context.Background()
	backendRepo := NewMemoryRepository()
	if err := backendRepo.Save(ctx, StorageBackend{ID: "local-main", Enabled: true}); err != nil {
		t.Fatalf("Save(backend) error = %v", err)
	}
	service := NewMediaObjectService(backendRepo, NewMemoryMediaObjectRepository())

	registered, err := service.RegisterMediaObject(ctx, validMediaObject())
	if err != nil {
		t.Fatalf("RegisterMediaObject() error = %v", err)
	}
	if registered.CreatedAt.IsZero() || registered.UpdatedAt.IsZero() {
		t.Fatalf("registered timestamps = %+v, want server-owned timestamps", registered)
	}
	got, err := service.GetMediaObject(ctx, registered.ID)
	if err != nil {
		t.Fatalf("GetMediaObject() error = %v", err)
	}
	if got.ID != registered.ID || got.BackendID != "local-main" {
		t.Fatalf("GetMediaObject() = %+v, want registered object", got)
	}
}

func TestMediaObjectServiceRejectsDisabledBackend(t *testing.T) {
	ctx := context.Background()
	backendRepo := NewMemoryRepository()
	if err := backendRepo.Save(ctx, StorageBackend{ID: "local-main", Enabled: false}); err != nil {
		t.Fatalf("Save(backend) error = %v", err)
	}
	service := NewMediaObjectService(backendRepo, NewMemoryMediaObjectRepository())

	_, err := service.RegisterMediaObject(ctx, validMediaObject())
	if !errors.Is(err, ErrBackendDisabled) {
		t.Fatalf("RegisterMediaObject() error = %v, want ErrBackendDisabled", err)
	}
}

func TestMediaObjectServiceRejectsMissingBackend(t *testing.T) {
	service := NewMediaObjectService(NewMemoryRepository(), NewMemoryMediaObjectRepository())

	_, err := service.RegisterMediaObject(context.Background(), validMediaObject())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("RegisterMediaObject() error = %v, want ErrNotFound", err)
	}
}

func TestValidateMediaObjectRejectsUnsafeObjectKeys(t *testing.T) {
	for _, key := range []string{"", "/absolute/audio.flac", "../escape.flac", "album/../escape.flac", "album\\track.flac"} {
		t.Run(key, func(t *testing.T) {
			object := validMediaObject()
			object.ObjectKey = key
			if err := ValidateMediaObject(&object); !errors.Is(err, ErrInvalidMediaObject) {
				t.Fatalf("ValidateMediaObject() error = %v, want ErrInvalidMediaObject", err)
			}
		})
	}
}

func TestValidateMediaObjectRejectsInvalidMetadata(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*MediaObject)
	}{
		{name: "content hash", mutate: func(object *MediaObject) { object.ContentHash = "sha256" }},
		{name: "negative size", mutate: func(object *MediaObject) { object.SizeBytes = -1 }},
		{name: "mime type", mutate: func(object *MediaObject) { object.MIMEType = "audio" }},
		{name: "asset kind", mutate: func(object *MediaObject) { object.AssetKind = "thumbnail" }},
		{name: "lifecycle", mutate: func(object *MediaObject) { object.LifecycleState = "missing" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			object := validMediaObject()
			tt.mutate(&object)
			if err := ValidateMediaObject(&object); !errors.Is(err, ErrInvalidMediaObject) {
				t.Fatalf("ValidateMediaObject() error = %v, want ErrInvalidMediaObject", err)
			}
		})
	}
}

func TestMediaObjectRepositoryListsByBackendInStableOrder(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryMediaObjectRepository()
	for _, object := range []MediaObject{
		{ID: "3", BackendID: "b", ObjectKey: "z.flac", ContentHash: "sha256:3", MIMEType: "audio/flac", AssetKind: string(AssetKindOriginalAudio), LifecycleState: string(LifecycleStateActive)},
		{ID: "2", BackendID: "b", ObjectKey: "a.flac", ContentHash: "sha256:2", MIMEType: "audio/flac", AssetKind: string(AssetKindOriginalAudio), LifecycleState: string(LifecycleStateActive)},
		{ID: "1", BackendID: "a", ObjectKey: "ignored.flac", ContentHash: "sha256:1", MIMEType: "audio/flac", AssetKind: string(AssetKindOriginalAudio), LifecycleState: string(LifecycleStateActive)},
	} {
		if err := repo.SaveMediaObject(ctx, object); err != nil {
			t.Fatalf("SaveMediaObject() error = %v", err)
		}
	}
	objects, err := repo.ListMediaObjectsByBackend(ctx, "b")
	if err != nil {
		t.Fatalf("ListMediaObjectsByBackend() error = %v", err)
	}
	if len(objects) != 2 || objects[0].ID != "2" || objects[1].ID != "3" {
		t.Fatalf("ListMediaObjectsByBackend() = %+v, want object-key order", objects)
	}
}

func TestMediaObjectRepositoryListsByContentHash(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryMediaObjectRepository()
	first := validMediaObject()
	first.ID = "first"
	second := validMediaObject()
	second.ID = "second"
	second.ObjectKey = "album/copy.flac"
	for _, object := range []MediaObject{second, first} {
		if err := repo.SaveMediaObject(ctx, object); err != nil {
			t.Fatalf("SaveMediaObject() error = %v", err)
		}
	}
	objects, err := repo.ListMediaObjectsByContentHash(ctx, first.ContentHash)
	if err != nil {
		t.Fatalf("ListMediaObjectsByContentHash() error = %v", err)
	}
	if len(objects) != 2 || objects[0].ID != "second" || objects[1].ID != "first" {
		t.Fatalf("ListMediaObjectsByContentHash() = %+v, want stable object-key order", objects)
	}
}

func TestMediaObjectServiceRejectsDuplicateObjectID(t *testing.T) {
	ctx := context.Background()
	backendRepo := NewMemoryRepository()
	if err := backendRepo.Save(ctx, StorageBackend{ID: "local-main", Enabled: true}); err != nil {
		t.Fatalf("Save(backend) error = %v", err)
	}
	service := NewMediaObjectService(backendRepo, NewMemoryMediaObjectRepository())
	object := validMediaObject()
	if _, err := service.RegisterMediaObject(ctx, object); err != nil {
		t.Fatalf("RegisterMediaObject(first) error = %v", err)
	}
	_, err := service.RegisterMediaObject(ctx, object)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("RegisterMediaObject(duplicate) error = %v, want ErrConflict", err)
	}
}

func validMediaObject() MediaObject {
	return MediaObject{
		ID:             "media-original-1",
		BackendID:      "local-main",
		ObjectKey:      "albums/inori/track-01.flac",
		ContentHash:    "sha256:0123456789abcdef",
		SizeBytes:      1234,
		MIMEType:       "audio/flac",
		AssetKind:      string(AssetKindOriginalAudio),
		LifecycleState: string(LifecycleStateActive),
	}
}

package storage

import (
	"context"
	"errors"
	"testing"
	"time"
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

func TestMediaObjectRepositoryListsByVerificationStatus(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryMediaObjectRepository()
	checkedAt := time.Date(2026, 6, 4, 2, 7, 0, 0, time.UTC)
	verified := validMediaObject()
	verified.ID = "verified"
	verified.ObjectKey = "album/verified.flac"
	verified.LastVerification = &MediaObjectVerificationResult{MediaObjectID: verified.ID, BackendID: verified.BackendID, ObjectKey: verified.ObjectKey, VerifiedAt: checkedAt, Status: "verified"}
	failed := validMediaObject()
	failed.ID = "failed"
	failed.ObjectKey = "album/failed.flac"
	failed.LastVerification = &MediaObjectVerificationResult{MediaObjectID: failed.ID, BackendID: failed.BackendID, ObjectKey: failed.ObjectKey, VerifiedAt: checkedAt, Status: "failed", Message: "sha256 mismatch"}
	unknown := validMediaObject()
	unknown.ID = "unknown"
	unknown.ObjectKey = "album/unknown.flac"
	for _, object := range []MediaObject{verified, failed, unknown} {
		if err := repo.SaveMediaObject(ctx, object); err != nil {
			t.Fatalf("SaveMediaObject() error = %v", err)
		}
	}

	verifiedObjects, err := repo.ListMediaObjectsByVerificationStatus(ctx, "verified")
	if err != nil {
		t.Fatalf("ListMediaObjectsByVerificationStatus(verified) error = %v", err)
	}
	if len(verifiedObjects) != 1 || verifiedObjects[0].ID != "verified" {
		t.Fatalf("verified objects = %+v, want only verified", verifiedObjects)
	}
	unknownObjects, err := repo.ListMediaObjectsByVerificationStatus(ctx, "unknown")
	if err != nil {
		t.Fatalf("ListMediaObjectsByVerificationStatus(unknown) error = %v", err)
	}
	if len(unknownObjects) != 1 || unknownObjects[0].ID != "unknown" {
		t.Fatalf("unknown objects = %+v, want only unverified", unknownObjects)
	}
}

func TestMediaObjectServiceRejectsInvalidVerificationStatusFilter(t *testing.T) {
	service := NewMediaObjectService(NewMemoryRepository(), NewMemoryMediaObjectRepository())
	_, err := service.ListMediaObjectsByVerificationStatus(context.Background(), "stale")
	if !errors.Is(err, ErrInvalidMediaObject) {
		t.Fatalf("ListMediaObjectsByVerificationStatus() error = %v, want ErrInvalidMediaObject", err)
	}
}

func TestMediaObjectServiceListsPaginatedObjects(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryMediaObjectRepository()
	for _, object := range []MediaObject{
		{ID: "3", BackendID: "local-main", ObjectKey: "c.flac", ContentHash: "sha256:same", MIMEType: "audio/flac", AssetKind: string(AssetKindOriginalAudio), LifecycleState: string(LifecycleStateActive)},
		{ID: "1", BackendID: "local-main", ObjectKey: "a.flac", ContentHash: "sha256:same", MIMEType: "audio/flac", AssetKind: string(AssetKindOriginalAudio), LifecycleState: string(LifecycleStateActive)},
		{ID: "2", BackendID: "local-main", ObjectKey: "b.flac", ContentHash: "sha256:same", MIMEType: "audio/flac", AssetKind: string(AssetKindOriginalAudio), LifecycleState: string(LifecycleStateActive)},
	} {
		if err := repo.SaveMediaObject(ctx, object); err != nil {
			t.Fatalf("SaveMediaObject() error = %v", err)
		}
	}
	service := NewMediaObjectService(NewMemoryRepository(), repo)

	page, err := service.ListMediaObjects(ctx, MediaObjectListFilter{BackendID: "local-main", Limit: 2, Offset: 1})
	if err != nil {
		t.Fatalf("ListMediaObjects() error = %v", err)
	}
	if len(page.Objects) != 2 || page.Objects[0].ID != "2" || page.Objects[1].ID != "3" {
		t.Fatalf("page objects = %+v, want second and third object", page.Objects)
	}
	if page.Pagination.Limit != 2 || page.Pagination.Offset != 1 || page.Pagination.Total != 3 || page.Pagination.HasMore {
		t.Fatalf("pagination = %+v, want limit/offset/total and no hasMore", page.Pagination)
	}
}

func TestMediaObjectServiceRejectsInvalidPagination(t *testing.T) {
	service := NewMediaObjectService(NewMemoryRepository(), NewMemoryMediaObjectRepository())
	for _, filter := range []MediaObjectListFilter{
		{BackendID: "local-main", Limit: -1},
		{BackendID: "local-main", Limit: MaxMediaObjectListLimit + 1},
		{BackendID: "local-main", Offset: -1},
		{BackendID: "local-main", ContentHash: "sha256:same"},
	} {
		if _, err := service.ListMediaObjects(context.Background(), filter); !errors.Is(err, ErrInvalidMediaObject) {
			t.Fatalf("ListMediaObjects(%+v) error = %v, want ErrInvalidMediaObject", filter, err)
		}
	}
}

func TestMediaObjectServiceBuildsMetadataStats(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryMediaObjectRepository()
	verified := validMediaObject()
	verified.ID = "verified"
	verified.BackendID = "local-main"
	verified.ObjectKey = "a.flac"
	verified.SizeBytes = 10
	verified.LastVerification = &MediaObjectVerificationResult{MediaObjectID: verified.ID, BackendID: verified.BackendID, ObjectKey: verified.ObjectKey, Status: "verified"}
	failed := validMediaObject()
	failed.ID = "failed"
	failed.BackendID = "archive"
	failed.ObjectKey = "b.flac"
	failed.SizeBytes = 20
	failed.AssetKind = string(AssetKindBackup)
	failed.LifecycleState = string(LifecycleStateArchived)
	failed.LastVerification = &MediaObjectVerificationResult{MediaObjectID: failed.ID, BackendID: failed.BackendID, ObjectKey: failed.ObjectKey, Status: "failed"}
	unknown := validMediaObject()
	unknown.ID = "unknown"
	unknown.BackendID = "local-main"
	unknown.ObjectKey = "c.flac"
	unknown.SizeBytes = 30
	for _, object := range []MediaObject{verified, failed, unknown} {
		if err := repo.SaveMediaObject(ctx, object); err != nil {
			t.Fatalf("SaveMediaObject() error = %v", err)
		}
	}
	service := NewMediaObjectService(NewMemoryRepository(), repo)

	stats, err := service.GetMediaObjectStats(ctx, "")
	if err != nil {
		t.Fatalf("GetMediaObjectStats() error = %v", err)
	}
	if stats.TotalObjects != 3 || stats.TotalSizeBytes != 60 || stats.ByVerificationStatus["verified"] != 1 || stats.ByVerificationStatus["failed"] != 1 || stats.ByVerificationStatus["unknown"] != 1 {
		t.Fatalf("stats = %+v, want total and verification buckets", stats)
	}
	if stats.ByBackendID["local-main"] != 2 || stats.ByAssetKind[string(AssetKindBackup)] != 1 || stats.ByLifecycleState[string(LifecycleStateArchived)] != 1 {
		t.Fatalf("stats buckets = %+v/%+v/%+v, want backend, asset kind, lifecycle counts", stats.ByBackendID, stats.ByAssetKind, stats.ByLifecycleState)
	}

	filtered, err := service.GetMediaObjectStats(ctx, "local-main")
	if err != nil {
		t.Fatalf("GetMediaObjectStats(filtered) error = %v", err)
	}
	if filtered.BackendID != "local-main" || filtered.TotalObjects != 2 || filtered.TotalSizeBytes != 40 || filtered.ByVerificationStatus["failed"] != 0 {
		t.Fatalf("filtered stats = %+v, want local-main only", filtered)
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

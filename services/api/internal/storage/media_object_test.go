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

func TestMediaObjectServiceSortsListBeforePagination(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryMediaObjectRepository()
	baseTime := time.Date(2026, 6, 4, 10, 20, 0, 0, time.UTC)
	for _, object := range []MediaObject{
		{ID: "small", BackendID: "local-main", ObjectKey: "small.flac", ContentHash: "sha256:same", SizeBytes: 10, MIMEType: "audio/flac", AssetKind: string(AssetKindOriginalAudio), LifecycleState: string(LifecycleStateActive), CreatedAt: baseTime.Add(2 * time.Minute), UpdatedAt: baseTime.Add(2 * time.Minute)},
		{ID: "large", BackendID: "local-main", ObjectKey: "large.flac", ContentHash: "sha256:same", SizeBytes: 30, MIMEType: "audio/flac", AssetKind: string(AssetKindOriginalAudio), LifecycleState: string(LifecycleStateActive), CreatedAt: baseTime, UpdatedAt: baseTime},
		{ID: "middle", BackendID: "local-main", ObjectKey: "middle.flac", ContentHash: "sha256:same", SizeBytes: 20, MIMEType: "audio/flac", AssetKind: string(AssetKindOriginalAudio), LifecycleState: string(LifecycleStateActive), CreatedAt: baseTime.Add(time.Minute), UpdatedAt: baseTime.Add(time.Minute)},
	} {
		if err := repo.SaveMediaObject(ctx, object); err != nil {
			t.Fatalf("SaveMediaObject() error = %v", err)
		}
	}
	service := NewMediaObjectService(NewMemoryRepository(), repo)

	page, err := service.ListMediaObjects(ctx, MediaObjectListFilter{BackendID: "local-main", SortBy: MediaObjectSortBySizeBytes, SortOrder: MediaObjectSortOrderDescending, Limit: 2})
	if err != nil {
		t.Fatalf("ListMediaObjects(size desc) error = %v", err)
	}
	if len(page.Objects) != 2 || page.Objects[0].ID != "large" || page.Objects[1].ID != "middle" {
		t.Fatalf("page objects = %+v, want size descending before pagination", page.Objects)
	}

	page, err = service.ListMediaObjects(ctx, MediaObjectListFilter{BackendID: "local-main", SortBy: MediaObjectSortByCreatedAt, SortOrder: MediaObjectSortOrderAscending, Limit: 3})
	if err != nil {
		t.Fatalf("ListMediaObjects(created asc) error = %v", err)
	}
	if page.Objects[0].ID != "large" || page.Objects[1].ID != "middle" || page.Objects[2].ID != "small" {
		t.Fatalf("page objects = %+v, want created_at ascending", page.Objects)
	}
}

func TestMediaObjectServiceRejectsInvalidPagination(t *testing.T) {
	service := NewMediaObjectService(NewMemoryRepository(), NewMemoryMediaObjectRepository())
	for _, filter := range []MediaObjectListFilter{
		{BackendID: "local-main", Limit: -1},
		{BackendID: "local-main", Limit: MaxMediaObjectListLimit + 1},
		{BackendID: "local-main", Offset: -1},
		{BackendID: "local-main", ContentHash: "sha256:same"},
		{BackendID: "local-main", SortBy: "missing"},
		{BackendID: "local-main", SortOrder: "sideways"},
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

func TestMediaObjectServiceUpdatesLifecycleState(t *testing.T) {
	ctx := context.Background()
	backendRepo := NewMemoryRepository()
	if err := backendRepo.Save(ctx, StorageBackend{ID: "local-main", Enabled: true}); err != nil {
		t.Fatalf("Save(backend) error = %v", err)
	}
	repo := NewMemoryMediaObjectRepository()
	service := NewMediaObjectService(backendRepo, repo)
	registered, err := service.RegisterMediaObject(ctx, validMediaObject())
	if err != nil {
		t.Fatalf("RegisterMediaObject() error = %v", err)
	}
	registered.LastVerification = &MediaObjectVerificationResult{MediaObjectID: registered.ID, BackendID: registered.BackendID, ObjectKey: registered.ObjectKey, Status: "verified"}
	if err := repo.SaveMediaObject(ctx, registered); err != nil {
		t.Fatalf("SaveMediaObject() error = %v", err)
	}

	updated, err := service.SetMediaObjectLifecycleState(ctx, registered.ID, string(LifecycleStateArchived))
	if err != nil {
		t.Fatalf("SetMediaObjectLifecycleState() error = %v", err)
	}
	if updated.LifecycleState != string(LifecycleStateArchived) || updated.LastVerification == nil || updated.LastVerification.Status != "verified" || !updated.UpdatedAt.After(registered.UpdatedAt) {
		t.Fatalf("updated object = %+v, want archived object preserving verification and updated timestamp", updated)
	}
}

func TestMediaObjectServiceRejectsInvalidLifecycleTransitions(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryMediaObjectRepository()
	deleted := validMediaObject()
	deleted.LifecycleState = string(LifecycleStateDeleted)
	if err := repo.SaveMediaObject(ctx, deleted); err != nil {
		t.Fatalf("SaveMediaObject() error = %v", err)
	}
	service := NewMediaObjectService(NewMemoryRepository(), repo)

	if _, err := service.SetMediaObjectLifecycleState(ctx, deleted.ID, string(LifecycleStateActive)); !errors.Is(err, ErrConflict) {
		t.Fatalf("SetMediaObjectLifecycleState(deleted->active) error = %v, want ErrConflict", err)
	}
	if _, err := service.SetMediaObjectLifecycleState(ctx, deleted.ID, "missing"); !errors.Is(err, ErrInvalidMediaObject) {
		t.Fatalf("SetMediaObjectLifecycleState(invalid) error = %v, want ErrInvalidMediaObject", err)
	}
}

func TestMediaObjectServiceListsByLifecycleState(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryMediaObjectRepository()
	active := validMediaObject()
	active.ID = "active"
	active.ObjectKey = "active.flac"
	archived := validMediaObject()
	archived.ID = "archived"
	archived.ObjectKey = "archived.flac"
	archived.LifecycleState = string(LifecycleStateArchived)
	for _, object := range []MediaObject{active, archived} {
		if err := repo.SaveMediaObject(ctx, object); err != nil {
			t.Fatalf("SaveMediaObject() error = %v", err)
		}
	}
	service := NewMediaObjectService(NewMemoryRepository(), repo)

	objects, err := service.ListMediaObjectsByLifecycleState(ctx, string(LifecycleStateArchived))
	if err != nil {
		t.Fatalf("ListMediaObjectsByLifecycleState() error = %v", err)
	}
	if len(objects) != 1 || objects[0].ID != "archived" {
		t.Fatalf("archived objects = %+v, want only archived", objects)
	}
	page, err := service.ListMediaObjects(ctx, MediaObjectListFilter{LifecycleState: string(LifecycleStateActive), Limit: 10})
	if err != nil {
		t.Fatalf("ListMediaObjects(lifecycle) error = %v", err)
	}
	if len(page.Objects) != 1 || page.Objects[0].ID != "active" || page.Pagination.Total != 1 {
		t.Fatalf("active page = %+v, want only active", page)
	}
}

func TestMediaObjectServiceRejectsInvalidLifecycleStateFilter(t *testing.T) {
	service := NewMediaObjectService(NewMemoryRepository(), NewMemoryMediaObjectRepository())
	_, err := service.ListMediaObjectsByLifecycleState(context.Background(), "missing")
	if !errors.Is(err, ErrInvalidMediaObject) {
		t.Fatalf("ListMediaObjectsByLifecycleState() error = %v, want ErrInvalidMediaObject", err)
	}
}

func TestMediaObjectServiceListsByAssetKind(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryMediaObjectRepository()
	audio := validMediaObject()
	audio.ID = "audio"
	audio.ObjectKey = "audio.flac"
	artwork := validMediaObject()
	artwork.ID = "artwork"
	artwork.ObjectKey = "cover.jpg"
	artwork.AssetKind = string(AssetKindArtwork)
	artwork.MIMEType = "image/jpeg"
	for _, object := range []MediaObject{audio, artwork} {
		if err := repo.SaveMediaObject(ctx, object); err != nil {
			t.Fatalf("SaveMediaObject() error = %v", err)
		}
	}
	service := NewMediaObjectService(NewMemoryRepository(), repo)

	objects, err := service.ListMediaObjectsByAssetKind(ctx, string(AssetKindArtwork))
	if err != nil {
		t.Fatalf("ListMediaObjectsByAssetKind() error = %v", err)
	}
	if len(objects) != 1 || objects[0].ID != "artwork" {
		t.Fatalf("artwork objects = %+v, want only artwork", objects)
	}
	page, err := service.ListMediaObjects(ctx, MediaObjectListFilter{AssetKind: string(AssetKindOriginalAudio), Limit: 10})
	if err != nil {
		t.Fatalf("ListMediaObjects(assetKind) error = %v", err)
	}
	if len(page.Objects) != 1 || page.Objects[0].ID != "audio" || page.Pagination.Total != 1 {
		t.Fatalf("audio page = %+v, want only original audio", page)
	}
}

func TestMediaObjectServiceRejectsInvalidAssetKindFilter(t *testing.T) {
	service := NewMediaObjectService(NewMemoryRepository(), NewMemoryMediaObjectRepository())
	_, err := service.ListMediaObjectsByAssetKind(context.Background(), "thumbnail")
	if !errors.Is(err, ErrInvalidMediaObject) {
		t.Fatalf("ListMediaObjectsByAssetKind() error = %v, want ErrInvalidMediaObject", err)
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

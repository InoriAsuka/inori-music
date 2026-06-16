//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"inori-music/services/api/internal/storage"
	pgstore "inori-music/services/api/internal/storage/postgres"
)

// setupTestDB starts a PostgreSQL container, runs migrations, and returns a pool.
// The container is terminated when t finishes.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("inori_test"),
		tcpostgres.WithUsername("inori"),
		tcpostgres.WithPassword("inori"),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	t.Cleanup(pool.Close)

	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}
	defer conn.Release()
	if err := pgstore.Migrate(ctx, conn.Conn()); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return pool
}

// ---- BackendRepository tests ----

func TestBackendRepository_SaveAndGet(t *testing.T) {
	pool := setupTestDB(t)
	repo := pgstore.NewBackendRepository(pool)
	ctx := context.Background()

	backend := storage.StorageBackend{
		ID:           "test-local-1",
		Type:         storage.BackendTypeLocal,
		DisplayName:  "Test Local",
		Enabled:      true,
		IsDefault:    false,
		Priority:     10,
		HealthStatus: storage.HealthStatusUnknown,
		Config:       storage.BackendConfig{Local: &storage.LocalConfig{RootPath: "/data/media"}},
		CreatedAt:    time.Now().UTC().Truncate(time.Millisecond),
		UpdatedAt:    time.Now().UTC().Truncate(time.Millisecond),
	}

	if err := repo.Save(ctx, backend); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.Get(ctx, backend.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != backend.ID {
		t.Errorf("ID: got %q want %q", got.ID, backend.ID)
	}
	if got.DisplayName != backend.DisplayName {
		t.Errorf("DisplayName: got %q want %q", got.DisplayName, backend.DisplayName)
	}
	if got.Config.Local == nil || got.Config.Local.RootPath != "/data/media" {
		t.Errorf("Config.Local.RootPath not preserved: %+v", got.Config)
	}
}

func TestBackendRepository_GetNotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := pgstore.NewBackendRepository(pool)
	_, err := repo.Get(context.Background(), "does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestBackendRepository_List(t *testing.T) {
	pool := setupTestDB(t)
	repo := pgstore.NewBackendRepository(pool)
	ctx := context.Background()

	for i, id := range []string{"b-1", "b-2", "b-3"} {
		b := storage.StorageBackend{
			ID: id, Type: storage.BackendTypeLocal, DisplayName: id,
			Enabled: true, Priority: i, HealthStatus: storage.HealthStatusUnknown,
			Config:    storage.BackendConfig{Local: &storage.LocalConfig{RootPath: "/data"}},
			CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		}
		if err := repo.Save(ctx, b); err != nil {
			t.Fatalf("Save %s: %v", id, err)
		}
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("List len: got %d want 3", len(list))
	}
}

func TestBackendRepository_ClearDefault(t *testing.T) {
	pool := setupTestDB(t)
	repo := pgstore.NewBackendRepository(pool)
	ctx := context.Background()

	b := storage.StorageBackend{
		ID: "default-backend", Type: storage.BackendTypeLocal, DisplayName: "Default",
		Enabled: true, IsDefault: true, Priority: 0, HealthStatus: storage.HealthStatusUnknown,
		Config:    storage.BackendConfig{Local: &storage.LocalConfig{RootPath: "/data"}},
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Save(ctx, b); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := repo.ClearDefault(ctx); err != nil {
		t.Fatalf("ClearDefault: %v", err)
	}
	got, err := repo.Get(ctx, b.ID)
	if err != nil {
		t.Fatalf("Get after ClearDefault: %v", err)
	}
	if got.IsDefault {
		t.Error("expected IsDefault=false after ClearDefault")
	}
}

func TestBackendRepository_Upsert(t *testing.T) {
	pool := setupTestDB(t)
	repo := pgstore.NewBackendRepository(pool)
	ctx := context.Background()

	b := storage.StorageBackend{
		ID: "upsert-backend", Type: storage.BackendTypeLocal, DisplayName: "Original",
		Enabled: true, Priority: 0, HealthStatus: storage.HealthStatusUnknown,
		Config:    storage.BackendConfig{Local: &storage.LocalConfig{RootPath: "/data"}},
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Save(ctx, b); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	b.DisplayName = "Updated"
	if err := repo.Save(ctx, b); err != nil {
		t.Fatalf("second Save: %v", err)
	}
	got, err := repo.Get(ctx, b.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.DisplayName != "Updated" {
		t.Errorf("DisplayName: got %q want %q", got.DisplayName, "Updated")
	}
}

// ---- MediaObjectRepository tests ----

func TestMediaObjectRepository_SaveAndGet(t *testing.T) {
	pool := setupTestDB(t)
	repo := pgstore.NewMediaObjectRepository(pool)
	ctx := context.Background()

	obj := storage.MediaObject{
		ID: "obj-1", BackendID: "backend-a", ObjectKey: "audio/track.flac",
		ContentHash: "sha256:abc123", SizeBytes: 1024, MIMEType: "audio/flac",
		AssetKind: "original_audio", LifecycleState: "active",
		CreatedAt: time.Now().UTC().Truncate(time.Millisecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Millisecond),
	}

	if err := repo.SaveMediaObject(ctx, obj); err != nil {
		t.Fatalf("SaveMediaObject: %v", err)
	}
	got, err := repo.GetMediaObject(ctx, obj.ID)
	if err != nil {
		t.Fatalf("GetMediaObject: %v", err)
	}
	if got.ObjectKey != obj.ObjectKey {
		t.Errorf("ObjectKey: got %q want %q", got.ObjectKey, obj.ObjectKey)
	}
	if got.ContentHash != obj.ContentHash {
		t.Errorf("ContentHash: got %q want %q", got.ContentHash, obj.ContentHash)
	}
}

func TestMediaObjectRepository_GetNotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := pgstore.NewMediaObjectRepository(pool)
	_, err := repo.GetMediaObject(context.Background(), "no-such-id")
	if err == nil {
		t.Fatal("expected error")
	}
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMediaObjectRepository_ListByBackend(t *testing.T) {
	pool := setupTestDB(t)
	repo := pgstore.NewMediaObjectRepository(pool)
	ctx := context.Background()

	for i, id := range []string{"o1", "o2", "o3"} {
		obj := storage.MediaObject{
			ID: id, BackendID: "backend-x", ObjectKey: "audio/" + id + ".flac",
			ContentHash: "sha256:" + id, SizeBytes: int64(i + 1),
			MIMEType: "audio/flac", AssetKind: "original_audio", LifecycleState: "active",
			CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		}
		if err := repo.SaveMediaObject(ctx, obj); err != nil {
			t.Fatalf("SaveMediaObject %s: %v", id, err)
		}
	}

	list, err := repo.ListMediaObjectsByBackend(ctx, "backend-x")
	if err != nil {
		t.Fatalf("ListMediaObjectsByBackend: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("len: got %d want 3", len(list))
	}
}

func TestMediaObjectRepository_ListByVerificationStatus(t *testing.T) {
	pool := setupTestDB(t)
	repo := pgstore.NewMediaObjectRepository(pool)
	ctx := context.Background()

	verifiedAt := time.Now().UTC()
	verified := storage.MediaObject{
		ID: "v1", BackendID: "b1", ObjectKey: "audio/v1.flac",
		ContentHash: "sha256:v1", SizeBytes: 100, MIMEType: "audio/flac",
		AssetKind: "original_audio", LifecycleState: "active",
		LastVerification: &storage.MediaObjectVerificationResult{Status: "verified", VerifiedAt: verifiedAt},
		CreatedAt:        time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	unknown := storage.MediaObject{
		ID: "v2", BackendID: "b1", ObjectKey: "audio/v2.flac",
		ContentHash: "sha256:v2", SizeBytes: 100, MIMEType: "audio/flac",
		AssetKind: "original_audio", LifecycleState: "active",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	for _, o := range []storage.MediaObject{verified, unknown} {
		if err := repo.SaveMediaObject(ctx, o); err != nil {
			t.Fatalf("SaveMediaObject %s: %v", o.ID, err)
		}
	}

	verifiedList, err := repo.ListMediaObjectsByVerificationStatus(ctx, "verified")
	if err != nil {
		t.Fatalf("list verified: %v", err)
	}
	if len(verifiedList) != 1 || verifiedList[0].ID != "v1" {
		t.Errorf("verified list: %+v", verifiedList)
	}

	unknownList, err := repo.ListMediaObjectsByVerificationStatus(ctx, "unknown")
	if err != nil {
		t.Fatalf("list unknown: %v", err)
	}
	if len(unknownList) != 1 || unknownList[0].ID != "v2" {
		t.Errorf("unknown list: %+v", unknownList)
	}
}

func TestMediaObjectRepository_Upsert(t *testing.T) {
	pool := setupTestDB(t)
	repo := pgstore.NewMediaObjectRepository(pool)
	ctx := context.Background()

	obj := storage.MediaObject{
		ID: "upsert-obj", BackendID: "b1", ObjectKey: "audio/x.flac",
		ContentHash: "sha256:x", SizeBytes: 50, MIMEType: "audio/flac",
		AssetKind: "original_audio", LifecycleState: "staged",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.SaveMediaObject(ctx, obj); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	obj.LifecycleState = "active"
	if err := repo.SaveMediaObject(ctx, obj); err != nil {
		t.Fatalf("second Save: %v", err)
	}
	got, err := repo.GetMediaObject(ctx, obj.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.LifecycleState != "active" {
		t.Errorf("LifecycleState: got %q want %q", got.LifecycleState, "active")
	}
}

// ---- helpers ----

func isNotFound(err error) bool {
	return err != nil && errors.Is(err, storage.ErrNotFound)
}

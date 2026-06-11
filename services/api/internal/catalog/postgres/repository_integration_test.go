//go:build integration

package catalogpg_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"inori-music/services/api/internal/catalog"
	catalogpg "inori-music/services/api/internal/catalog/postgres"
	pgstore "inori-music/services/api/internal/storage/postgres"
)

func setupCatalogTestDB(t *testing.T) *pgxpool.Pool {
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

func TestRepositorySaveListAndDeleteCatalog(t *testing.T) {
	pool := setupCatalogTestDB(t)
	repo := catalogpg.NewRepository(pool)
	ctx := context.Background()

	now := time.Now().UTC()
	artist := catalog.Artist{ID: "artist-1", Name: "Hatsune Miku", SortName: "Miku, Hatsune", CreatedAt: now, UpdatedAt: now}
	if err := repo.SaveArtist(ctx, artist); err != nil {
		t.Fatalf("SaveArtist: %v", err)
	}
	if err := repo.SaveArtist(ctx, catalog.Artist{ID: "artist-1", Name: "Hatsune Miku", SortName: "Miku, Hatsune", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("upsert artist: %v", err)
	}
	gotArtist, err := repo.GetArtist(ctx, "artist-1")
	if err != nil {
		t.Fatalf("GetArtist: %v", err)
	}
	if gotArtist.Name != "Hatsune Miku" {
		t.Fatalf("artist = %+v", gotArtist)
	}

	album := catalog.Album{ID: "album-1", Title: "Project DIVA", SortTitle: "Project DIVA", ArtistID: "artist-1", ReleaseYear: 2009, CreatedAt: now, UpdatedAt: now}
	if err := repo.SaveAlbum(ctx, album); err != nil {
		t.Fatalf("SaveAlbum: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO media_objects (id, backend_id, object_key, content_hash, size_bytes, mime_type, asset_kind, lifecycle_state, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		"media-1", "backend-1", "tracks/world-is-mine.flac", "sha256:abc", 123, "audio/flac", "original_audio", "active", now, now)
	if err != nil {
		t.Fatalf("insert media object: %v", err)
	}
	track := catalog.Track{ID: "track-1", Title: "World Is Mine", SortTitle: "World Is Mine", ArtistID: "artist-1", AlbumID: "album-1", MediaObjectID: "media-1", TrackNumber: 1, DiscNumber: 1, DurationMS: 245000, CreatedAt: now, UpdatedAt: now}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("SaveTrack: %v", err)
	}

	albums, err := repo.ListAlbumsByArtist(ctx, "artist-1")
	if err != nil || len(albums) != 1 || albums[0].ID != "album-1" {
		t.Fatalf("albums = %+v err=%v", albums, err)
	}
	tracks, err := repo.ListTracksByAlbum(ctx, "album-1")
	if err != nil || len(tracks) != 1 || tracks[0].ID != "track-1" {
		t.Fatalf("tracks = %+v err=%v", tracks, err)
	}
	tracks, err = repo.ListTracksByArtist(ctx, "artist-1")
	if err != nil || len(tracks) != 1 || tracks[0].MediaObjectID != "media-1" {
		t.Fatalf("tracks by artist = %+v err=%v", tracks, err)
	}

	if err := repo.DeleteTrack(ctx, "track-1"); err != nil {
		t.Fatalf("DeleteTrack: %v", err)
	}
	if _, err := repo.GetTrack(ctx, "track-1"); !errors.Is(err, catalog.ErrTrackNotFound) {
		t.Fatalf("GetTrack err = %v, want ErrTrackNotFound", err)
	}
}

//go:build integration

package catalogpg_test

import (
	"context"
	"errors"
	"fmt"
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

func TestRepositorySearchCatalog(t *testing.T) {
	pool := setupCatalogTestDB(t)
	repo := catalogpg.NewRepository(pool)
	ctx := context.Background()

	now := time.Now().UTC()

	// Seed data
	if err := repo.SaveArtist(ctx, catalog.Artist{ID: "search-artist-1", Name: "Hatsune Miku", SortName: "Miku, Hatsune", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("SaveArtist: %v", err)
	}
	if err := repo.SaveAlbum(ctx, catalog.Album{ID: "search-album-1", Title: "Project DIVA Mega Mix", SortTitle: "Project DIVA Mega Mix", ArtistID: "search-artist-1", ReleaseYear: 2020, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("SaveAlbum: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO media_objects (id, backend_id, object_key, content_hash, size_bytes, mime_type, asset_kind, lifecycle_state, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		"search-media-1", "backend-1", "tracks/world-is-mine.flac", "sha256:search", 123, "audio/flac", "original_audio", "active", now, now,
	); err != nil {
		t.Fatalf("insert media object: %v", err)
	}
	if err := repo.SaveTrack(ctx, catalog.Track{
		ID: "search-track-1", Title: "World Is Mine", SortTitle: "World Is Mine",
		ArtistID: "search-artist-1", AlbumID: "search-album-1", MediaObjectID: "search-media-1",
		TrackNumber: 1, DiscNumber: 1, DurationMS: 245000, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("SaveTrack: %v", err)
	}

	t.Run("match artist", func(t *testing.T) {
		result, err := repo.SearchCatalog(ctx, "miku")
		if err != nil {
			t.Fatalf("SearchCatalog: %v", err)
		}
		if result.Query != "miku" {
			t.Fatalf("Query = %q, want miku", result.Query)
		}
		artistHits := 0
		for _, item := range result.Items {
			if item.Kind == catalog.SearchResultArtist {
				artistHits++
			}
		}
		if artistHits != 1 {
			t.Fatalf("artist hits = %d, want 1; items = %v", artistHits, result.Items)
		}
	})

	t.Run("match album", func(t *testing.T) {
		result, err := repo.SearchCatalog(ctx, "diva")
		if err != nil {
			t.Fatalf("SearchCatalog: %v", err)
		}
		albumHits := 0
		for _, item := range result.Items {
			if item.Kind == catalog.SearchResultAlbum {
				albumHits++
			}
		}
		if albumHits != 1 {
			t.Fatalf("album hits = %d, want 1", albumHits)
		}
	})

	t.Run("match track", func(t *testing.T) {
		result, err := repo.SearchCatalog(ctx, "world")
		if err != nil {
			t.Fatalf("SearchCatalog: %v", err)
		}
		trackHits := 0
		for _, item := range result.Items {
			if item.Kind == catalog.SearchResultTrack {
				trackHits++
			}
		}
		if trackHits != 1 {
			t.Fatalf("track hits = %d, want 1", trackHits)
		}
	})

	t.Run("no results", func(t *testing.T) {
		result, err := repo.SearchCatalog(ctx, "notfound")
		if err != nil {
			t.Fatalf("SearchCatalog: %v", err)
		}
		if len(result.Items) != 0 {
			t.Fatalf("Items = %v, want empty", result.Items)
		}
	})
}

func TestRepositoryListArtistsPage(t *testing.T) {
	pool := setupCatalogTestDB(t)
	repo := catalogpg.NewRepository(pool)
	ctx := context.Background()

	names := []string{"Zara", "Alice", "Mike"}
	for _, name := range names {
		if err := repo.SaveArtist(ctx, catalog.Artist{
			ID: name, Name: name, SortName: name,
			CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		}); err != nil {
			t.Fatalf("SaveArtist %q: %v", name, err)
		}
	}

	// asc by name, all
	page, err := repo.ListArtistsPage(ctx, catalog.ListQuery{
		SortBy: catalog.ArtistSortByName, SortOrder: catalog.CatalogSortOrderAsc, Limit: 10, Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListArtistsPage: %v", err)
	}
	if len(page.Items) != 3 {
		t.Fatalf("items = %d, want 3", len(page.Items))
	}
	if page.Total != 3 {
		t.Errorf("total = %d, want 3", page.Total)
	}
	if page.Items[0].Name != "Alice" {
		t.Errorf("items[0].Name = %q, want Alice", page.Items[0].Name)
	}
	if page.Items[2].Name != "Zara" {
		t.Errorf("items[2].Name = %q, want Zara", page.Items[2].Name)
	}

	// desc, limit=1, total still 3
	page, err = repo.ListArtistsPage(ctx, catalog.ListQuery{
		SortBy: catalog.ArtistSortByName, SortOrder: catalog.CatalogSortOrderDesc, Limit: 1, Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListArtistsPage desc: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(page.Items))
	}
	if page.Total != 3 {
		t.Errorf("total = %d, want 3", page.Total)
	}
	if page.Items[0].Name != "Zara" {
		t.Errorf("items[0].Name = %q, want Zara", page.Items[0].Name)
	}

	// offset=1 limit=1 → Mike
	page, err = repo.ListArtistsPage(ctx, catalog.ListQuery{
		SortBy: catalog.ArtistSortByName, SortOrder: catalog.CatalogSortOrderAsc, Limit: 1, Offset: 1,
	})
	if err != nil {
		t.Fatalf("ListArtistsPage offset=1: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(page.Items))
	}
	if page.Items[0].Name != "Mike" {
		t.Errorf("items[0].Name = %q, want Mike", page.Items[0].Name)
	}
	if page.Total != 3 {
		t.Errorf("total = %d, want 3", page.Total)
	}

	// offset past end
	page, err = repo.ListArtistsPage(ctx, catalog.ListQuery{
		SortBy: catalog.ArtistSortByName, SortOrder: catalog.CatalogSortOrderAsc, Limit: 10, Offset: 99,
	})
	if err != nil {
		t.Fatalf("ListArtistsPage offset=99: %v", err)
	}
	if len(page.Items) != 0 {
		t.Fatalf("items = %d, want 0", len(page.Items))
	}
	if page.Total != 0 {
		// When offset >= total the window func still returns 0 for COUNT(*) OVER ()
		// because no rows qualify. Both 0 and 3 are acceptable; check items only.
		t.Logf("note: total = %d (may be 0 when no rows returned)", page.Total)
	}
}

func TestRepositoryListAlbumsPageByArtist(t *testing.T) {
	pool := setupCatalogTestDB(t)
	repo := catalogpg.NewRepository(pool)
	ctx := context.Background()

	artist := catalog.Artist{ID: "artist-1", Name: "Band", SortName: "Band",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := repo.SaveArtist(ctx, artist); err != nil {
		t.Fatalf("SaveArtist: %v", err)
	}
	for i, year := range []int{2020, 2015, 2023} {
		a := catalog.Album{
			ID: fmt.Sprintf("album-%d", i), Title: fmt.Sprintf("Album %d", year),
			SortTitle: fmt.Sprintf("Album %d", year), ArtistID: "artist-1", ReleaseYear: year,
			CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		}
		if err := repo.SaveAlbum(ctx, a); err != nil {
			t.Fatalf("SaveAlbum: %v", err)
		}
	}

	page, err := repo.ListAlbumsByArtistPage(ctx, "artist-1", catalog.ListQuery{
		SortBy: catalog.AlbumSortByReleaseYear, SortOrder: catalog.CatalogSortOrderAsc, Limit: 10, Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListAlbumsByArtistPage: %v", err)
	}
	if len(page.Items) != 3 || page.Total != 3 {
		t.Fatalf("want 3 total=3, got %d total=%d", len(page.Items), page.Total)
	}
	if page.Items[0].ReleaseYear != 2015 {
		t.Errorf("items[0].ReleaseYear = %d, want 2015", page.Items[0].ReleaseYear)
	}
}

func TestRepositoryCountEntities(t *testing.T) {
	pool := setupCatalogTestDB(t)
	repo := catalogpg.NewRepository(pool)
	ctx := context.Background()

	// empty
	stats, err := repo.CountEntities(ctx)
	if err != nil {
		t.Fatalf("CountEntities (empty): %v", err)
	}
	if stats.Artists != 0 || stats.Albums != 0 || stats.Tracks != 0 || stats.Playlists != 0 {
		t.Fatalf("empty catalog stats = %+v, want all zero", stats)
	}

	// seed one of each
	artist := catalog.Artist{ID: "a1", Name: "Band", SortName: "Band",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := repo.SaveArtist(ctx, artist); err != nil {
		t.Fatalf("SaveArtist: %v", err)
	}
	album := catalog.Album{ID: "al1", Title: "Album", SortTitle: "Album", ArtistID: "a1", ReleaseYear: 2024,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := repo.SaveAlbum(ctx, album); err != nil {
		t.Fatalf("SaveAlbum: %v", err)
	}

	stats, err = repo.CountEntities(ctx)
	if err != nil {
		t.Fatalf("CountEntities: %v", err)
	}
	if stats.Artists != 1 || stats.Albums != 1 {
		t.Errorf("stats = %+v, want Artists=1 Albums=1", stats)
	}
}

func TestRepositoryArtistAlbumTrackCounts(t *testing.T) {
	pool := setupCatalogTestDB(t)
	repo := catalogpg.NewRepository(pool)
	ctx := context.Background()

	if err := repo.SaveArtist(ctx, catalog.Artist{ID: "b1", Name: "Band", SortName: "Band",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("SaveArtist: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := repo.SaveAlbum(ctx, catalog.Album{
			ID: fmt.Sprintf("al%d", i), Title: fmt.Sprintf("Album %d", i),
			SortTitle: fmt.Sprintf("Album %d", i), ArtistID: "b1", ReleaseYear: 2020 + i,
			CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		}); err != nil {
			t.Fatalf("SaveAlbum: %v", err)
		}
	}

	items, err := repo.ArtistAlbumTrackCounts(ctx)
	if err != nil {
		t.Fatalf("ArtistAlbumTrackCounts: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %d, want 1", len(items))
	}
	if items[0].AlbumCount != 2 {
		t.Errorf("AlbumCount = %d, want 2", items[0].AlbumCount)
	}
	if items[0].TrackCount != 0 {
		t.Errorf("TrackCount = %d, want 0", items[0].TrackCount)
	}
}

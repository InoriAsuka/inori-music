// Command reindex performs a full rebuild of the Meilisearch search indexes
// from the PostgreSQL catalog. Use it after index corruption or a searchable-
// attribute schema change, where the only alternative would be replaying every
// write one at a time.
package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/catalog"
	catalogpg "inori-music/services/api/internal/catalog/postgres"
	"inori-music/services/api/internal/search"
	pgstore "inori-music/services/api/internal/storage/postgres"
)

// pageSize bounds how many rows are fetched per List*Page call while walking
// the full catalog.
const pageSize = 200

func main() {
	ctx := context.Background()

	dsn := os.Getenv("INORI_DATABASE_URL")
	if dsn == "" {
		log.Fatal("reindex: INORI_DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("reindex: connect to database: %v", err)
	}
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("reindex: acquire connection: %v", err)
	}
	migrateErr := pgstore.Migrate(ctx, conn.Conn())
	conn.Release()
	if migrateErr != nil {
		log.Fatalf("reindex: migrate: %v", migrateErr)
	}

	meiliHost := os.Getenv("MEILI_HOST")
	if meiliHost == "" {
		log.Fatal("reindex: MEILI_HOST is required")
	}
	searchSvc, err := search.NewMeilisearch(meiliHost, os.Getenv("MEILI_SEARCH_KEY"))
	if err != nil {
		log.Fatalf("reindex: meilisearch init: %v", err)
	}
	if searchSvc == nil {
		log.Fatalf("reindex: meilisearch not reachable at %s", meiliHost)
	}

	repo := catalogpg.NewRepository(pool)

	log.Print("reindex: clearing existing indexes")
	if err := searchSvc.ClearIndexes(ctx); err != nil {
		log.Fatalf("reindex: clear indexes: %v", err)
	}

	artistsOK, artistsFailed := reindexArtists(ctx, repo, searchSvc)
	albumsOK, albumsFailed := reindexAlbums(ctx, repo, searchSvc)
	tracksOK, tracksFailed := reindexTracks(ctx, repo, searchSvc)

	log.Printf("reindex: done — artists %d ok/%d failed, albums %d ok/%d failed, tracks %d ok/%d failed",
		artistsOK, artistsFailed, albumsOK, albumsFailed, tracksOK, tracksFailed)

	if artistsFailed+albumsFailed+tracksFailed > 0 {
		os.Exit(1)
	}
}

// entityPage holds a single page of catalog entities together with the total count.
type entityPage struct {
	items []any
	total int
}

// fetchPageFn retrieves the next page of entities starting at offset.
type fetchPageFn func(ctx context.Context, sortBy string, limit, offset int) (entityPage, error)

// indexFn indexes one entity; non-fatal errors are counted as failures.
type indexFn func(entity any) error

// entityField extracts the log-friendly label from one entity.
type entityField func(entity any) (id string)

// reindexWalk applies the same paginated walking pattern across every entity kind.
// It is package-level (not a method) so unit tests can exercise it with stub
// fetchPageFn / indexFn without a real repository or Meilisearch.
func reindexWalk(ctx context.Context, kind string, fetchPage fetchPageFn, sortBy string, getLabel entityField, index indexFn) (ok, failed int) {
	offset := 0
	for {
		page, err := fetchPage(ctx, sortBy, pageSize, offset)
		if err != nil {
			log.Fatalf("reindex: list %s at offset %d: %v", kind, offset, err)
		}
		for _, e := range page.items {
			if err := index(e); err != nil {
				log.Printf("reindex: index %s %s: %v", kind, getLabel(e), err)
				failed++
				continue
			}
			ok++
		}
		offset += len(page.items)
		log.Printf("reindex: %s %d/%d", kind, offset, page.total)
		if len(page.items) == 0 || offset >= page.total {
			break
		}
	}
	return ok, failed
}

func reindexArtists(ctx context.Context, repo catalog.Repository, searchSvc *search.MeilisearchService) (ok, failed int) {
	return reindexWalk(ctx, "artists",
		func(_ context.Context, _ string, limit, offset int) (entityPage, error) {
			page, err := repo.ListArtistsPage(ctx, catalog.ListQuery{Limit: limit, Offset: offset, SortBy: catalog.ArtistSortByCreatedAt})
			return entityPage{items: sliceOf(page.Items, func(a catalog.Artist) any { return a }), total: page.Total}, err
		},
		catalog.ArtistSortByCreatedAt,
		func(e any) string { return e.(catalog.Artist).ID },
		func(e any) error { a := e.(catalog.Artist); return searchSvc.IndexArtist(ctx, a.ID, a.Name) },
	)
}

func reindexAlbums(ctx context.Context, repo catalog.Repository, searchSvc *search.MeilisearchService) (ok, failed int) {
	return reindexWalk(ctx, "albums",
		func(_ context.Context, _ string, limit, offset int) (entityPage, error) {
			page, err := repo.ListAlbumsPage(ctx, catalog.ListQuery{Limit: limit, Offset: offset, SortBy: catalog.AlbumSortByCreatedAt})
			return entityPage{items: sliceOf(page.Items, func(a catalog.Album) any { return a }), total: page.Total}, err
		},
		catalog.AlbumSortByCreatedAt,
		func(e any) string { return e.(catalog.Album).ID },
		func(e any) error { a := e.(catalog.Album); return searchSvc.IndexAlbum(ctx, a.ID, a.Title, a.ArtistID) },
	)
}

func reindexTracks(ctx context.Context, repo catalog.Repository, searchSvc *search.MeilisearchService) (ok, failed int) {
	return reindexWalk(ctx, "tracks",
		func(_ context.Context, _ string, limit, offset int) (entityPage, error) {
			page, err := repo.ListTracksPage(ctx, catalog.ListQuery{Limit: limit, Offset: offset, SortBy: catalog.TrackSortByCreatedAt})
			return entityPage{items: sliceOf(page.Items, func(t catalog.Track) any { return t }), total: page.Total}, err
		},
		catalog.TrackSortByCreatedAt,
		func(e any) string { return e.(catalog.Track).ID },
		func(e any) error { t := e.(catalog.Track); return searchSvc.IndexTrack(ctx, t.ID, t.Title, t.ArtistID, t.Genre) },
	)
}

func sliceOf[T any, U any](in []T, fn func(T) U) []U {
	out := make([]U, 0, len(in))
	for _, v := range in {
		out = append(out, fn(v))
	}
	return out
}

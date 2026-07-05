package search

import (
	"context"
)

// CatalogSearcher is the minimal interface needed by PostgresService.
// catalog.Service implements this.
type CatalogSearcher interface {
	SearchCatalog(ctx context.Context, query string) (CatalogResult, error)
}

// CatalogResult mirrors catalog.CatalogSearchResult without importing the catalog package.
type CatalogResult struct {
	Query string
	Items []CatalogResultItem
}

// CatalogResultItem mirrors catalog.SearchResultItem.
type CatalogResultItem struct {
	ArtistID string // non-empty when item is an artist
	AlbumID  string // non-empty when item is an album
	TrackID  string // non-empty when item is a track
}

// PostgresService wraps a CatalogSearcher as a search.Service fallback.
type PostgresService struct {
	searcher func(ctx context.Context, q string) (CatalogResult, error)
}

// NewPostgresService accepts a search function so the search package does not
// need to import the catalog package (avoiding import cycles).
func NewPostgresService(fn func(ctx context.Context, q string) (CatalogResult, error)) *PostgresService {
	return &PostgresService{searcher: fn}
}

func (s *PostgresService) Search(ctx context.Context, q string, limit int) (SearchResult, error) {
	result, err := s.searcher(ctx, q)
	if err != nil {
		return SearchResult{}, err
	}
	// Highlights stays an empty (non-nil) map: PG full-text search has no
	// snippet/highlight capability, so callers see "no highlight" uniformly
	// rather than having to distinguish nil-vs-empty across backends.
	out := SearchResult{Highlights: map[string]string{}}
	for _, item := range result.Items {
		switch {
		case item.ArtistID != "":
			out.Artists = append(out.Artists, item.ArtistID)
		case item.AlbumID != "":
			out.Albums = append(out.Albums, item.AlbumID)
		case item.TrackID != "":
			out.Tracks = append(out.Tracks, item.TrackID)
		}
	}
	return out, nil
}

func (s *PostgresService) IndexTrack(_ context.Context, _, _, _, _ string) error {
	return nil // PG indexes via tsvector, no-op
}

func (s *PostgresService) IndexAlbum(_ context.Context, _, _, _ string) error {
	return nil
}

func (s *PostgresService) IndexArtist(_ context.Context, _, _ string) error {
	return nil
}

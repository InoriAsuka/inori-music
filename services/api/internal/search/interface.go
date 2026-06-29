package search

import "context"

// SearchResult holds categorized result IDs.
type SearchResult struct {
	Artists []string
	Albums  []string
	Tracks  []string
}

// Service abstracts the search backend.
type Service interface {
	Search(ctx context.Context, q string, limit int) (SearchResult, error)
	IndexTrack(ctx context.Context, trackID, title, artistName, genre string) error
	IndexAlbum(ctx context.Context, albumID, title, artistName string) error
	IndexArtist(ctx context.Context, artistID, name string) error
}

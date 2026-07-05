package search

import "context"

// SearchResult holds categorized result IDs.
type SearchResult struct {
	Artists []string
	Albums  []string
	Tracks  []string
	// Highlights maps an entity ID to its highlighted match snippet (HTML with
	// <mark> tags), when the backend supports highlighting. Empty when the
	// backend does not support it (e.g. PG fallback) or an entity had no
	// highlightable match.
	Highlights map[string]string
}

// Service abstracts the search backend.
type Service interface {
	Search(ctx context.Context, q string, limit int) (SearchResult, error)
	IndexTrack(ctx context.Context, trackID, title, artistName, genre string) error
	IndexAlbum(ctx context.Context, albumID, title, artistName string) error
	IndexArtist(ctx context.Context, artistID, name string) error
}

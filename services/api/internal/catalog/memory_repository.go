package catalog

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// MemoryRepository is an in-memory catalog repository for tests and development.
type MemoryRepository struct {
	mu      sync.RWMutex
	artists map[string]Artist
	albums  map[string]Album
	tracks  map[string]Track
}

// NewMemoryRepository returns an empty in-memory catalog repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		artists: map[string]Artist{},
		albums:  map[string]Album{},
		tracks:  map[string]Track{},
	}
}

func (r *MemoryRepository) SaveArtist(_ context.Context, a Artist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.artists[a.ID] = a
	return nil
}

func (r *MemoryRepository) GetArtist(_ context.Context, id string) (Artist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.artists[id]
	if !ok {
		return Artist{}, fmt.Errorf("%w: %s", ErrArtistNotFound, id)
	}
	return a, nil
}

func (r *MemoryRepository) ListArtists(_ context.Context) ([]Artist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Artist, 0, len(r.artists))
	for _, a := range r.artists {
		out = append(out, a)
	}
	return out, nil
}

func (r *MemoryRepository) DeleteArtist(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.artists[id]; !ok {
		return fmt.Errorf("%w: %s", ErrArtistNotFound, id)
	}
	delete(r.artists, id)
	return nil
}

func (r *MemoryRepository) SaveAlbum(_ context.Context, a Album) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.albums[a.ID] = a
	return nil
}

func (r *MemoryRepository) GetAlbum(_ context.Context, id string) (Album, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.albums[id]
	if !ok {
		return Album{}, fmt.Errorf("%w: %s", ErrAlbumNotFound, id)
	}
	return a, nil
}

func (r *MemoryRepository) ListAlbums(_ context.Context) ([]Album, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Album, 0, len(r.albums))
	for _, a := range r.albums {
		out = append(out, a)
	}
	return out, nil
}

func (r *MemoryRepository) ListAlbumsByArtist(_ context.Context, artistID string) ([]Album, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Album
	for _, a := range r.albums {
		if a.ArtistID == artistID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *MemoryRepository) DeleteAlbum(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.albums[id]; !ok {
		return fmt.Errorf("%w: %s", ErrAlbumNotFound, id)
	}
	delete(r.albums, id)
	return nil
}

func (r *MemoryRepository) SaveTrack(_ context.Context, t Track) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tracks[t.ID] = t
	return nil
}

func (r *MemoryRepository) GetTrack(_ context.Context, id string) (Track, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tracks[id]
	if !ok {
		return Track{}, fmt.Errorf("%w: %s", ErrTrackNotFound, id)
	}
	return t, nil
}

func (r *MemoryRepository) ListTracks(_ context.Context) ([]Track, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Track, 0, len(r.tracks))
	for _, t := range r.tracks {
		out = append(out, t)
	}
	return out, nil
}

func (r *MemoryRepository) ListTracksByAlbum(_ context.Context, albumID string) ([]Track, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Track
	for _, t := range r.tracks {
		if t.AlbumID == albumID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *MemoryRepository) ListTracksByArtist(_ context.Context, artistID string) ([]Track, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Track
	for _, t := range r.tracks {
		if t.ArtistID == artistID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *MemoryRepository) DeleteTrack(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tracks[id]; !ok {
		return fmt.Errorf("%w: %s", ErrTrackNotFound, id)
	}
	delete(r.tracks, id)
	return nil
}

// SearchCatalog is a case-insensitive substring fallback for environments that do not
// have PostgreSQL full-text search available (e.g. unit tests).
func (r *MemoryRepository) SearchCatalog(_ context.Context, query string) (CatalogSearchResult, error) {
	q := strings.ToLower(query)
	result := CatalogSearchResult{Query: query}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.artists {
		if strings.Contains(strings.ToLower(a.Name), q) || strings.Contains(strings.ToLower(a.SortName), q) {
			ac := a
			result.Items = append(result.Items, SearchResultItem{Kind: SearchResultArtist, Artist: &ac})
		}
	}
	for _, a := range r.albums {
		if strings.Contains(strings.ToLower(a.Title), q) || strings.Contains(strings.ToLower(a.SortTitle), q) {
			ac := a
			result.Items = append(result.Items, SearchResultItem{Kind: SearchResultAlbum, Album: &ac})
		}
	}
	for _, t := range r.tracks {
		if strings.Contains(strings.ToLower(t.Title), q) || strings.Contains(strings.ToLower(t.SortTitle), q) {
			tc := t
			result.Items = append(result.Items, SearchResultItem{Kind: SearchResultTrack, Track: &tc})
		}
	}
	return result, nil
}

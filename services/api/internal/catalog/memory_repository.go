package catalog

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// MemoryRepository is an in-memory catalog repository for tests and development.
type MemoryRepository struct {
	mu        sync.RWMutex
	artists   map[string]Artist
	albums    map[string]Album
	tracks    map[string]Track
	playlists map[string]Playlist
}

// NewMemoryRepository returns an empty in-memory catalog repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		artists:   map[string]Artist{},
		albums:    map[string]Album{},
		tracks:    map[string]Track{},
		playlists: map[string]Playlist{},
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

func (r *MemoryRepository) SavePlaylist(_ context.Context, p Playlist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Store a defensive copy of the TrackIDs slice.
	cp := p
	cp.TrackIDs = make([]string, len(p.TrackIDs))
	copy(cp.TrackIDs, p.TrackIDs)
	r.playlists[p.ID] = cp
	return nil
}

func (r *MemoryRepository) GetPlaylist(_ context.Context, id string) (Playlist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.playlists[id]
	if !ok {
		return Playlist{}, fmt.Errorf("%w: %s", ErrPlaylistNotFound, id)
	}
	cp := p
	cp.TrackIDs = make([]string, len(p.TrackIDs))
	copy(cp.TrackIDs, p.TrackIDs)
	return cp, nil
}

func (r *MemoryRepository) ListPlaylists(_ context.Context) ([]Playlist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Playlist, 0, len(r.playlists))
	for _, p := range r.playlists {
		cp := p
		cp.TrackIDs = make([]string, len(p.TrackIDs))
		copy(cp.TrackIDs, p.TrackIDs)
		out = append(out, cp)
	}
	return out, nil
}

func (r *MemoryRepository) DeletePlaylist(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.playlists[id]; !ok {
		return fmt.Errorf("%w: %s", ErrPlaylistNotFound, id)
	}
	delete(r.playlists, id)
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

// ---- ListPage methods (in-memory sort + slice to satisfy the Repository interface) ----

func memPage[T any](items []T, q ListQuery, less func(a, b T) bool) (ListPage[T], error) {
	sorted := make([]T, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool {
		if q.SortOrder == CatalogSortOrderDesc {
			return less(sorted[j], sorted[i])
		}
		return less(sorted[i], sorted[j])
	})
	total := len(sorted)
	start := q.Offset
	if start >= total {
		return ListPage[T]{Items: []T{}, Total: total}, nil
	}
	end := start + q.Limit
	if end > total {
		end = total
	}
	return ListPage[T]{Items: sorted[start:end], Total: total}, nil
}

func (r *MemoryRepository) ListArtistsPage(_ context.Context, q ListQuery) (ListPage[Artist], error) {
	r.mu.RLock()
	all := make([]Artist, 0, len(r.artists))
	for _, a := range r.artists {
		all = append(all, a)
	}
	r.mu.RUnlock()
	return memPage(all, q, func(a, b Artist) bool {
		switch q.SortBy {
		case ArtistSortBySortName:
			return strings.ToLower(a.SortName) < strings.ToLower(b.SortName)
		case ArtistSortByCreatedAt:
			return a.CreatedAt.Before(b.CreatedAt)
		case ArtistSortByUpdatedAt:
			return a.UpdatedAt.Before(b.UpdatedAt)
		default:
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		}
	})
}

func (r *MemoryRepository) ListAlbumsPage(_ context.Context, q ListQuery) (ListPage[Album], error) {
	r.mu.RLock()
	all := make([]Album, 0, len(r.albums))
	for _, a := range r.albums {
		if (q.ReleaseYearMin == 0 || a.ReleaseYear >= q.ReleaseYearMin) &&
			(q.ReleaseYearMax == 0 || a.ReleaseYear <= q.ReleaseYearMax) {
			all = append(all, a)
		}
	}
	r.mu.RUnlock()
	return memPage(all, q, albumLess(q.SortBy))
}

func (r *MemoryRepository) ListAlbumsByArtistPage(_ context.Context, artistID string, q ListQuery) (ListPage[Album], error) {
	r.mu.RLock()
	var all []Album
	for _, a := range r.albums {
		if a.ArtistID == artistID &&
			(q.ReleaseYearMin == 0 || a.ReleaseYear >= q.ReleaseYearMin) &&
			(q.ReleaseYearMax == 0 || a.ReleaseYear <= q.ReleaseYearMax) {
			all = append(all, a)
		}
	}
	r.mu.RUnlock()
	return memPage(all, q, albumLess(q.SortBy))
}

func albumLess(sortBy string) func(a, b Album) bool {
	return func(a, b Album) bool {
		switch sortBy {
		case AlbumSortBySortTitle:
			return strings.ToLower(a.SortTitle) < strings.ToLower(b.SortTitle)
		case AlbumSortByReleaseYear:
			return a.ReleaseYear < b.ReleaseYear
		case AlbumSortByCreatedAt:
			return a.CreatedAt.Before(b.CreatedAt)
		case AlbumSortByUpdatedAt:
			return a.UpdatedAt.Before(b.UpdatedAt)
		default:
			return strings.ToLower(a.Title) < strings.ToLower(b.Title)
		}
	}
}

func (r *MemoryRepository) ListTracksPage(_ context.Context, q ListQuery) (ListPage[Track], error) {
	r.mu.RLock()
	all := make([]Track, 0, len(r.tracks))
	for _, t := range r.tracks {
		if q.Genre == "" || strings.EqualFold(t.Genre, q.Genre) {
			all = append(all, t)
		}
	}
	r.mu.RUnlock()
	return memPage(all, q, trackLess(q.SortBy))
}

func (r *MemoryRepository) ListTracksByAlbumPage(_ context.Context, albumID string, q ListQuery) (ListPage[Track], error) {
	r.mu.RLock()
	var all []Track
	for _, t := range r.tracks {
		if t.AlbumID == albumID && (q.Genre == "" || strings.EqualFold(t.Genre, q.Genre)) {
			all = append(all, t)
		}
	}
	r.mu.RUnlock()
	return memPage(all, q, trackLess(q.SortBy))
}

func (r *MemoryRepository) ListTracksByArtistPage(_ context.Context, artistID string, q ListQuery) (ListPage[Track], error) {
	r.mu.RLock()
	var all []Track
	for _, t := range r.tracks {
		if t.ArtistID == artistID && (q.Genre == "" || strings.EqualFold(t.Genre, q.Genre)) {
			all = append(all, t)
		}
	}
	r.mu.RUnlock()
	return memPage(all, q, trackLess(q.SortBy))
}

func trackLess(sortBy string) func(a, b Track) bool {
	return func(a, b Track) bool {
		switch sortBy {
		case TrackSortBySortTitle:
			return strings.ToLower(a.SortTitle) < strings.ToLower(b.SortTitle)
		case TrackSortByTrackNumber:
			return a.TrackNumber < b.TrackNumber
		case TrackSortByDiscNumber:
			return a.DiscNumber < b.DiscNumber
		case TrackSortByDurationMS:
			return a.DurationMS < b.DurationMS
		case TrackSortByGenre:
			return strings.ToLower(a.Genre) < strings.ToLower(b.Genre)
		case TrackSortByCreatedAt:
			return a.CreatedAt.Before(b.CreatedAt)
		case TrackSortByUpdatedAt:
			return a.UpdatedAt.Before(b.UpdatedAt)
		default:
			return strings.ToLower(a.Title) < strings.ToLower(b.Title)
		}
	}
}

func (r *MemoryRepository) ListPlaylistsPage(_ context.Context, q ListQuery) (ListPage[Playlist], error) {
	r.mu.RLock()
	all := make([]Playlist, 0, len(r.playlists))
	for _, p := range r.playlists {
		all = append(all, p)
	}
	r.mu.RUnlock()
	return memPage(all, q, func(a, b Playlist) bool {
		switch q.SortBy {
		case PlaylistSortByCreatedAt:
			return a.CreatedAt.Before(b.CreatedAt)
		case PlaylistSortByUpdatedAt:
			return a.UpdatedAt.Before(b.UpdatedAt)
		default:
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		}
	})
}

// ---- Aggregate stats methods ----

func (r *MemoryRepository) CountEntities(_ context.Context) (CatalogStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return CatalogStats{
		Artists:   len(r.artists),
		Albums:    len(r.albums),
		Tracks:    len(r.tracks),
		Playlists: len(r.playlists),
	}, nil
}

func (r *MemoryRepository) ArtistAlbumTrackCounts(ctx context.Context) ([]ArtistStatItem, error) {
	r.mu.RLock()
	artists := make([]Artist, 0, len(r.artists))
	for _, a := range r.artists {
		artists = append(artists, a)
	}
	r.mu.RUnlock()
	items := make([]ArtistStatItem, 0, len(artists))
	for _, a := range artists {
		albumCount := 0
		trackCount := 0
		r.mu.RLock()
		for _, al := range r.albums {
			if al.ArtistID == a.ID {
				albumCount++
			}
		}
		for _, t := range r.tracks {
			if t.ArtistID == a.ID {
				trackCount++
			}
		}
		r.mu.RUnlock()
		items = append(items, ArtistStatItem{
			ArtistID:   a.ID,
			Name:       a.Name,
			AlbumCount: albumCount,
			TrackCount: trackCount,
		})
	}
	return items, nil
}

func (r *MemoryRepository) AlbumTrackCounts(ctx context.Context) ([]AlbumStatItem, error) {
	r.mu.RLock()
	albums := make([]Album, 0, len(r.albums))
	for _, a := range r.albums {
		albums = append(albums, a)
	}
	r.mu.RUnlock()
	items := make([]AlbumStatItem, 0, len(albums))
	for _, al := range albums {
		count := 0
		r.mu.RLock()
		for _, t := range r.tracks {
			if t.AlbumID == al.ID {
				count++
			}
		}
		r.mu.RUnlock()
		items = append(items, AlbumStatItem{
			AlbumID:    al.ID,
			Title:      al.Title,
			ArtistID:   al.ArtistID,
			TrackCount: count,
		})
	}
	return items, nil
}

func (r *MemoryRepository) PlaylistTrackCounts(_ context.Context) ([]PlaylistStatItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]PlaylistStatItem, 0, len(r.playlists))
	for _, p := range r.playlists {
		items = append(items, PlaylistStatItem{
			PlaylistID: p.ID,
			Name:       p.Name,
			TrackCount: len(p.TrackIDs),
		})
	}
	return items, nil
}

func (r *MemoryRepository) RecentlyAdded(ctx context.Context, kind string, limit int) ([]RecentCatalogItem, error) {
	items := make([]RecentCatalogItem, 0)
	r.mu.RLock()
	if kind == "" || kind == string(RecentItemArtist) {
		for _, a := range r.artists {
			ac := a
			items = append(items, RecentCatalogItem{Kind: RecentItemArtist, Artist: &ac, AddedAt: a.CreatedAt})
		}
	}
	if kind == "" || kind == string(RecentItemAlbum) {
		for _, a := range r.albums {
			ac := a
			items = append(items, RecentCatalogItem{Kind: RecentItemAlbum, Album: &ac, AddedAt: a.CreatedAt})
		}
	}
	if kind == "" || kind == string(RecentItemTrack) {
		for _, t := range r.tracks {
			tc := t
			items = append(items, RecentCatalogItem{Kind: RecentItemTrack, Track: &tc, AddedAt: t.CreatedAt})
		}
	}
	if kind == "" || kind == string(RecentItemPlaylist) {
		for _, p := range r.playlists {
			pc := p
			pc.TrackIDs = make([]string, len(p.TrackIDs))
			copy(pc.TrackIDs, p.TrackIDs)
			items = append(items, RecentCatalogItem{Kind: RecentItemPlaylist, Playlist: &pc, AddedAt: p.CreatedAt})
		}
	}
	r.mu.RUnlock()
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].AddedAt.After(items[j].AddedAt)
	})
	if limit < len(items) {
		items = items[:limit]
	}
	return items, nil
}

func (r *MemoryRepository) RecentlyUpdated(ctx context.Context, kind string, limit int) ([]UpdatedCatalogItem, error) {
	items := make([]UpdatedCatalogItem, 0)
	r.mu.RLock()
	if kind == "" || kind == string(RecentItemArtist) {
		for _, a := range r.artists {
			ac := a
			items = append(items, UpdatedCatalogItem{Kind: RecentItemArtist, Artist: &ac, UpdatedAt: a.UpdatedAt})
		}
	}
	if kind == "" || kind == string(RecentItemAlbum) {
		for _, a := range r.albums {
			ac := a
			items = append(items, UpdatedCatalogItem{Kind: RecentItemAlbum, Album: &ac, UpdatedAt: a.UpdatedAt})
		}
	}
	if kind == "" || kind == string(RecentItemTrack) {
		for _, t := range r.tracks {
			tc := t
			items = append(items, UpdatedCatalogItem{Kind: RecentItemTrack, Track: &tc, UpdatedAt: t.UpdatedAt})
		}
	}
	if kind == "" || kind == string(RecentItemPlaylist) {
		for _, p := range r.playlists {
			pc := p
			pc.TrackIDs = make([]string, len(p.TrackIDs))
			copy(pc.TrackIDs, p.TrackIDs)
			items = append(items, UpdatedCatalogItem{Kind: RecentItemPlaylist, Playlist: &pc, UpdatedAt: p.UpdatedAt})
		}
	}
	r.mu.RUnlock()
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	if limit < len(items) {
		items = items[:limit]
	}
	return items, nil
}

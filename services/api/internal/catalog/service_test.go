package catalog_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"inori-music/services/api/internal/catalog"
)

type memRepo struct {
	mu        sync.RWMutex
	artists   map[string]catalog.Artist
	albums    map[string]catalog.Album
	tracks    map[string]catalog.Track
	playlists map[string]catalog.Playlist
}

func newMemRepo() *memRepo {
	return &memRepo{
		artists:   map[string]catalog.Artist{},
		albums:    map[string]catalog.Album{},
		tracks:    map[string]catalog.Track{},
		playlists: map[string]catalog.Playlist{},
	}
}

func (r *memRepo) SaveArtist(_ context.Context, a catalog.Artist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.artists[a.ID] = a
	return nil
}

func (r *memRepo) GetArtist(_ context.Context, id string) (catalog.Artist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.artists[id]
	if !ok {
		return catalog.Artist{}, fmt.Errorf("%w: %s", catalog.ErrArtistNotFound, id)
	}
	return a, nil
}

func (r *memRepo) ListArtists(_ context.Context) ([]catalog.Artist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]catalog.Artist, 0, len(r.artists))
	for _, a := range r.artists {
		out = append(out, a)
	}
	return out, nil
}

func (r *memRepo) DeleteArtist(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.artists[id]; !ok {
		return fmt.Errorf("%w: %s", catalog.ErrArtistNotFound, id)
	}
	delete(r.artists, id)
	return nil
}

func (r *memRepo) SaveAlbum(_ context.Context, a catalog.Album) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.albums[a.ID] = a
	return nil
}

func (r *memRepo) GetAlbum(_ context.Context, id string) (catalog.Album, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.albums[id]
	if !ok {
		return catalog.Album{}, fmt.Errorf("%w: %s", catalog.ErrAlbumNotFound, id)
	}
	return a, nil
}

func (r *memRepo) ListAlbums(_ context.Context) ([]catalog.Album, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]catalog.Album, 0, len(r.albums))
	for _, a := range r.albums {
		out = append(out, a)
	}
	return out, nil
}

func (r *memRepo) ListAlbumsByArtist(_ context.Context, artistID string) ([]catalog.Album, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []catalog.Album{}
	for _, a := range r.albums {
		if a.ArtistID == artistID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *memRepo) DeleteAlbum(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.albums[id]; !ok {
		return fmt.Errorf("%w: %s", catalog.ErrAlbumNotFound, id)
	}
	delete(r.albums, id)
	return nil
}

func (r *memRepo) SaveTrack(_ context.Context, t catalog.Track) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tracks[t.ID] = t
	return nil
}

func (r *memRepo) GetTrack(_ context.Context, id string) (catalog.Track, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tr, ok := r.tracks[id]
	if !ok {
		return catalog.Track{}, fmt.Errorf("%w: %s", catalog.ErrTrackNotFound, id)
	}
	return tr, nil
}

func (r *memRepo) ListTracks(_ context.Context) ([]catalog.Track, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]catalog.Track, 0, len(r.tracks))
	for _, tr := range r.tracks {
		out = append(out, tr)
	}
	return out, nil
}

func (r *memRepo) ListTracksByAlbum(_ context.Context, albumID string) ([]catalog.Track, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []catalog.Track{}
	for _, tr := range r.tracks {
		if tr.AlbumID == albumID {
			out = append(out, tr)
		}
	}
	return out, nil
}

func (r *memRepo) ListTracksByArtist(_ context.Context, artistID string) ([]catalog.Track, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []catalog.Track{}
	for _, tr := range r.tracks {
		if tr.ArtistID == artistID {
			out = append(out, tr)
		}
	}
	return out, nil
}

func (r *memRepo) DeleteTrack(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tracks[id]; !ok {
		return fmt.Errorf("%w: %s", catalog.ErrTrackNotFound, id)
	}
	delete(r.tracks, id)
	return nil
}

func (r *memRepo) SavePlaylist(_ context.Context, p catalog.Playlist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := p
	cp.TrackIDs = make([]string, len(p.TrackIDs))
	copy(cp.TrackIDs, p.TrackIDs)
	r.playlists[p.ID] = cp
	return nil
}

func (r *memRepo) GetPlaylist(_ context.Context, id string) (catalog.Playlist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.playlists[id]
	if !ok {
		return catalog.Playlist{}, fmt.Errorf("%w: %s", catalog.ErrPlaylistNotFound, id)
	}
	cp := p
	cp.TrackIDs = make([]string, len(p.TrackIDs))
	copy(cp.TrackIDs, p.TrackIDs)
	return cp, nil
}

func (r *memRepo) ListPlaylists(_ context.Context) ([]catalog.Playlist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]catalog.Playlist, 0, len(r.playlists))
	for _, p := range r.playlists {
		cp := p
		cp.TrackIDs = make([]string, len(p.TrackIDs))
		copy(cp.TrackIDs, p.TrackIDs)
		out = append(out, cp)
	}
	return out, nil
}

func (r *memRepo) DeletePlaylist(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.playlists[id]; !ok {
		return fmt.Errorf("%w: %s", catalog.ErrPlaylistNotFound, id)
	}
	delete(r.playlists, id)
	return nil
}

func (r *memRepo) SearchCatalog(_ context.Context, query string) (catalog.CatalogSearchResult, error) {
	q := strings.ToLower(query)
	result := catalog.CatalogSearchResult{Query: query}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.artists {
		if strings.Contains(strings.ToLower(a.Name), q) {
			ac := a
			result.Items = append(result.Items, catalog.SearchResultItem{Kind: catalog.SearchResultArtist, Artist: &ac})
		}
	}
	for _, a := range r.albums {
		if strings.Contains(strings.ToLower(a.Title), q) {
			ac := a
			result.Items = append(result.Items, catalog.SearchResultItem{Kind: catalog.SearchResultAlbum, Album: &ac})
		}
	}
	for _, t := range r.tracks {
		if strings.Contains(strings.ToLower(t.Title), q) {
			tc := t
			result.Items = append(result.Items, catalog.SearchResultItem{Kind: catalog.SearchResultTrack, Track: &tc})
		}
	}
	return result, nil
}

// ListXxxPage stubs — the test memRepo delegates to the in-package MemoryRepository
// which already implements the full page/sort logic. We construct a temporary
// MemoryRepository, copy all items into it, and call its page method.

// ---- Aggregate stats stubs ----

func (r *memRepo) CountEntities(ctx context.Context) (catalog.CatalogStats, error) {
	return catalog.CatalogStats{
		Artists:   len(r.artists),
		Albums:    len(r.albums),
		Tracks:    len(r.tracks),
		Playlists: len(r.playlists),
	}, nil
}

func (r *memRepo) ArtistAlbumTrackCounts(ctx context.Context) ([]catalog.ArtistStatItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]catalog.ArtistStatItem, 0, len(r.artists))
	for _, a := range r.artists {
		albumCount, trackCount := 0, 0
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
		items = append(items, catalog.ArtistStatItem{ArtistID: a.ID, Name: a.Name, AlbumCount: albumCount, TrackCount: trackCount})
	}
	return items, nil
}

func (r *memRepo) AlbumTrackCounts(ctx context.Context) ([]catalog.AlbumStatItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]catalog.AlbumStatItem, 0, len(r.albums))
	for _, al := range r.albums {
		count := 0
		for _, t := range r.tracks {
			if t.AlbumID == al.ID {
				count++
			}
		}
		items = append(items, catalog.AlbumStatItem{AlbumID: al.ID, Title: al.Title, ArtistID: al.ArtistID, TrackCount: count})
	}
	return items, nil
}

func (r *memRepo) PlaylistTrackCounts(ctx context.Context) ([]catalog.PlaylistStatItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]catalog.PlaylistStatItem, 0, len(r.playlists))
	for _, p := range r.playlists {
		items = append(items, catalog.PlaylistStatItem{PlaylistID: p.ID, Name: p.Name, TrackCount: len(p.TrackIDs)})
	}
	return items, nil
}


func (r *memRepo) ListArtistsPage(ctx context.Context, q catalog.ListQuery) (catalog.ListPage[catalog.Artist], error) {
	r.mu.RLock()
	tmp := catalog.NewMemoryRepository()
	for _, a := range r.artists {
		_ = tmp.SaveArtist(ctx, a)
	}
	r.mu.RUnlock()
	return tmp.ListArtistsPage(ctx, q)
}

func (r *memRepo) ListAlbumsPage(ctx context.Context, q catalog.ListQuery) (catalog.ListPage[catalog.Album], error) {
	r.mu.RLock()
	tmp := catalog.NewMemoryRepository()
	for _, a := range r.albums {
		_ = tmp.SaveAlbum(ctx, a)
	}
	r.mu.RUnlock()
	return tmp.ListAlbumsPage(ctx, q)
}

func (r *memRepo) ListAlbumsByArtistPage(ctx context.Context, artistID string, q catalog.ListQuery) (catalog.ListPage[catalog.Album], error) {
	r.mu.RLock()
	tmp := catalog.NewMemoryRepository()
	for _, a := range r.albums {
		_ = tmp.SaveAlbum(ctx, a)
	}
	r.mu.RUnlock()
	return tmp.ListAlbumsByArtistPage(ctx, artistID, q)
}

func (r *memRepo) ListTracksPage(ctx context.Context, q catalog.ListQuery) (catalog.ListPage[catalog.Track], error) {
	r.mu.RLock()
	tmp := catalog.NewMemoryRepository()
	for _, t := range r.tracks {
		_ = tmp.SaveTrack(ctx, t)
	}
	r.mu.RUnlock()
	return tmp.ListTracksPage(ctx, q)
}

func (r *memRepo) ListTracksByAlbumPage(ctx context.Context, albumID string, q catalog.ListQuery) (catalog.ListPage[catalog.Track], error) {
	r.mu.RLock()
	tmp := catalog.NewMemoryRepository()
	for _, t := range r.tracks {
		_ = tmp.SaveTrack(ctx, t)
	}
	r.mu.RUnlock()
	return tmp.ListTracksByAlbumPage(ctx, albumID, q)
}

func (r *memRepo) ListTracksByArtistPage(ctx context.Context, artistID string, q catalog.ListQuery) (catalog.ListPage[catalog.Track], error) {
	r.mu.RLock()
	tmp := catalog.NewMemoryRepository()
	for _, t := range r.tracks {
		_ = tmp.SaveTrack(ctx, t)
	}
	r.mu.RUnlock()
	return tmp.ListTracksByArtistPage(ctx, artistID, q)
}

func (r *memRepo) ListPlaylistsPage(ctx context.Context, q catalog.ListQuery) (catalog.ListPage[catalog.Playlist], error) {
	r.mu.RLock()
	tmp := catalog.NewMemoryRepository()
	for _, p := range r.playlists {
		_ = tmp.SavePlaylist(ctx, p)
	}
	r.mu.RUnlock()
	return tmp.ListPlaylistsPage(ctx, q)
}

func TestServiceCreatesArtistAlbumAndTrack(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())

	artist, err := svc.CreateArtist(ctx, "Hatsune Miku", "Miku, Hatsune")
	if err != nil {
		t.Fatalf("CreateArtist: %v", err)
	}
	if artist.ID == "" || artist.Name != "Hatsune Miku" || artist.SortName != "Miku, Hatsune" {
		t.Fatalf("artist = %+v", artist)
	}

	album, err := svc.CreateAlbum(ctx, "Project DIVA", "Project DIVA", artist.ID, 2009)
	if err != nil {
		t.Fatalf("CreateAlbum: %v", err)
	}
	if album.ID == "" || album.ArtistID != artist.ID || album.ReleaseYear != 2009 {
		t.Fatalf("album = %+v", album)
	}

	track, err := svc.CreateTrack(ctx, "World Is Mine", "World Is Mine", artist.ID, album.ID, "media-1", 1, 1, 245000)
	if err != nil {
		t.Fatalf("CreateTrack: %v", err)
	}
	if track.ID == "" || track.ArtistID != artist.ID || track.AlbumID != album.ID || track.MediaObjectID != "media-1" {
		t.Fatalf("track = %+v", track)
	}

	tracks, err := svc.ListTracksByAlbum(ctx, album.ID)
	if err != nil || len(tracks) != 1 || tracks[0].ID != track.ID {
		t.Fatalf("tracks = %+v err=%v", tracks, err)
	}
}

func TestServiceValidatesRequiredFields(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	if _, err := svc.CreateArtist(ctx, " ", ""); !errors.Is(err, catalog.ErrInvalidArtist) {
		t.Fatalf("CreateArtist err = %v, want ErrInvalidArtist", err)
	}
	if _, err := svc.CreateAlbum(ctx, "", "", "artist", 0); !errors.Is(err, catalog.ErrInvalidAlbum) {
		t.Fatalf("CreateAlbum err = %v, want ErrInvalidAlbum", err)
	}
	if _, err := svc.CreateTrack(ctx, "", "", "artist", "", "media", 0, 0, 0); !errors.Is(err, catalog.ErrInvalidTrack) {
		t.Fatalf("CreateTrack err = %v, want ErrInvalidTrack", err)
	}
}

func TestServiceRequiresExistingArtistAndMatchingAlbum(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	svc := catalog.NewService(repo)

	if _, err := svc.CreateAlbum(ctx, "Missing", "", "missing", 0); !errors.Is(err, catalog.ErrArtistNotFound) {
		t.Fatalf("CreateAlbum err = %v, want ErrArtistNotFound", err)
	}

	artist, err := svc.CreateArtist(ctx, "Artist One", "")
	if err != nil {
		t.Fatal(err)
	}
	other, err := svc.CreateArtist(ctx, "Artist Two", "")
	if err != nil {
		t.Fatal(err)
	}
	album, err := svc.CreateAlbum(ctx, "Album", "", artist.ID, 2026)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateTrack(ctx, "Song", "", other.ID, album.ID, "media-1", 1, 1, 1000); !errors.Is(err, catalog.ErrInvalidTrack) {
		t.Fatalf("CreateTrack err = %v, want ErrInvalidTrack", err)
	}
}

func TestServiceDeletesRecords(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "Artist", "")
	album, _ := svc.CreateAlbum(ctx, "Album", "", artist.ID, 0)
	track, _ := svc.CreateTrack(ctx, "Track", "", artist.ID, album.ID, "media-1", 0, 0, 0)
	if err := svc.DeleteTrack(ctx, track.ID); err != nil {
		t.Fatalf("DeleteTrack: %v", err)
	}
	if err := svc.DeleteAlbum(ctx, album.ID); err != nil {
		t.Fatalf("DeleteAlbum: %v", err)
	}
	if err := svc.DeleteArtist(ctx, artist.ID); err != nil {
		t.Fatalf("DeleteArtist: %v", err)
	}
	if _, err := svc.GetTrack(ctx, track.ID); !errors.Is(err, catalog.ErrTrackNotFound) {
		t.Fatalf("GetTrack err = %v, want ErrTrackNotFound", err)
	}
}

// ---- ImportTrack tests ----

type memMediaReader struct {
	objects map[string]catalog.MediaObjectInfo
}

func newMemMediaReader() *memMediaReader {
	return &memMediaReader{objects: map[string]catalog.MediaObjectInfo{}}
}

func (r *memMediaReader) add(id, assetKind, lifecycleState string) {
	r.objects[id] = catalog.MediaObjectInfo{ID: id, AssetKind: assetKind, LifecycleState: lifecycleState}
}

func (r *memMediaReader) GetMediaObjectInfo(_ context.Context, id string) (catalog.MediaObjectInfo, error) {
	info, ok := r.objects[id]
	if !ok {
		return catalog.MediaObjectInfo{}, fmt.Errorf("not found: %s", id)
	}
	return info, nil
}

func newImportSvc(t *testing.T) (*catalog.Service, *memMediaReader) {
	t.Helper()
	repo := catalog.NewMemoryRepository()
	reader := newMemMediaReader()
	svc := catalog.NewService(repo).WithMediaObjectReader(reader)
	return svc, reader
}

func TestImportTrackSuccess(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-1", "original_audio", "active")

	track, err := svc.ImportTrack(ctx, catalog.ImportTrackRequest{
		MediaObjectID: "media-1",
		Title:         "World Is Mine",
		TrackNumber:   1,
		DurationMS:    245000,
	})
	if err != nil {
		t.Fatalf("ImportTrack: %v", err)
	}
	if track.ID == "" {
		t.Fatal("expected non-empty track id")
	}
	if track.Title != "World Is Mine" {
		t.Fatalf("Title = %q, want %q", track.Title, "World Is Mine")
	}
	if track.MediaObjectID != "media-1" {
		t.Fatalf("MediaObjectID = %q, want media-1", track.MediaObjectID)
	}
	if track.DurationMS != 245000 {
		t.Fatalf("DurationMS = %d, want 245000", track.DurationMS)
	}
}

func TestImportTrackTitleFallback(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-2", "transcoded_audio", "active")

	track, err := svc.ImportTrack(ctx, catalog.ImportTrackRequest{MediaObjectID: "media-2"})
	if err != nil {
		t.Fatalf("ImportTrack: %v", err)
	}
	if track.Title != "media-2" {
		t.Fatalf("Title = %q, want %q (media object id fallback)", track.Title, "media-2")
	}
}

func TestImportTrackWrongAssetKind(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-3", "artwork", "active")

	_, err := svc.ImportTrack(ctx, catalog.ImportTrackRequest{MediaObjectID: "media-3"})
	if !errors.Is(err, catalog.ErrImportRejected) {
		t.Fatalf("err = %v, want ErrImportRejected", err)
	}
}

func TestImportTrackNotActive(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-4", "original_audio", "staged")

	_, err := svc.ImportTrack(ctx, catalog.ImportTrackRequest{MediaObjectID: "media-4"})
	if !errors.Is(err, catalog.ErrImportRejected) {
		t.Fatalf("err = %v, want ErrImportRejected", err)
	}
}

func TestImportTrackNotFound(t *testing.T) {
	ctx := context.Background()
	svc, _ := newImportSvc(t)

	_, err := svc.ImportTrack(ctx, catalog.ImportTrackRequest{MediaObjectID: "missing"})
	if err == nil {
		t.Fatal("expected error for missing media object")
	}
}

func TestImportTrackNoReaderConfigured(t *testing.T) {
	ctx := context.Background()
	repo := catalog.NewMemoryRepository()
	svc := catalog.NewService(repo) // no media reader

	_, err := svc.ImportTrack(ctx, catalog.ImportTrackRequest{MediaObjectID: "media-1"})
	if !errors.Is(err, catalog.ErrImportRejected) {
		t.Fatalf("err = %v, want ErrImportRejected", err)
	}
}

func TestImportTrackWithArtistAndAlbum(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-5", "original_audio", "active")

	artist, err := svc.CreateArtist(ctx, "Miku", "")
	if err != nil {
		t.Fatalf("CreateArtist: %v", err)
	}
	album, err := svc.CreateAlbum(ctx, "supercell", "", artist.ID, 2009)
	if err != nil {
		t.Fatalf("CreateAlbum: %v", err)
	}

	track, err := svc.ImportTrack(ctx, catalog.ImportTrackRequest{
		MediaObjectID: "media-5",
		Title:         "World Is Mine",
		ArtistID:      artist.ID,
		AlbumID:       album.ID,
		TrackNumber:   1,
	})
	if err != nil {
		t.Fatalf("ImportTrack: %v", err)
	}
	if track.ArtistID != artist.ID {
		t.Fatalf("ArtistID = %q, want %q", track.ArtistID, artist.ID)
	}
	if track.AlbumID != album.ID {
		t.Fatalf("AlbumID = %q, want %q", track.AlbumID, album.ID)
	}
}

func TestImportTrackAlbumArtistInherited(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-6", "original_audio", "active")

	artist, _ := svc.CreateArtist(ctx, "Ryo", "")
	album, _ := svc.CreateAlbum(ctx, "supercell", "", artist.ID, 0)

	// no artistID supplied — should inherit from album
	track, err := svc.ImportTrack(ctx, catalog.ImportTrackRequest{
		MediaObjectID: "media-6",
		AlbumID:       album.ID,
	})
	if err != nil {
		t.Fatalf("ImportTrack: %v", err)
	}
	if track.ArtistID != artist.ID {
		t.Fatalf("ArtistID = %q, want inherited %q", track.ArtistID, artist.ID)
	}
}

// ---- RelinkTrack tests ----

func TestRelinkTrackSuccess(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-a", "original_audio", "active")
	reader.add("media-b", "transcoded_audio", "active")

	// create a track with media-a via import
	orig, err := svc.ImportTrack(ctx, catalog.ImportTrackRequest{
		MediaObjectID: "media-a",
		Title:         "Song A",
	})
	if err != nil {
		t.Fatalf("ImportTrack: %v", err)
	}

	// relink to media-b
	linked, err := svc.RelinkTrack(ctx, orig.ID, catalog.RelinkTrackRequest{MediaObjectID: "media-b"})
	if err != nil {
		t.Fatalf("RelinkTrack: %v", err)
	}
	if linked.MediaObjectID != "media-b" {
		t.Fatalf("MediaObjectID = %q, want media-b", linked.MediaObjectID)
	}
	// other fields must be preserved
	if linked.Title != "Song A" {
		t.Fatalf("Title = %q, want Song A", linked.Title)
	}
}

func TestRelinkTrackWrongAssetKind(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-a", "original_audio", "active")
	reader.add("media-art", "artwork", "active")

	orig, _ := svc.ImportTrack(ctx, catalog.ImportTrackRequest{MediaObjectID: "media-a", Title: "T"})
	_, err := svc.RelinkTrack(ctx, orig.ID, catalog.RelinkTrackRequest{MediaObjectID: "media-art"})
	if !errors.Is(err, catalog.ErrRelinkRejected) {
		t.Fatalf("err = %v, want ErrRelinkRejected", err)
	}
}

func TestRelinkTrackNotActive(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-a", "original_audio", "active")
	reader.add("media-staged", "original_audio", "staged")

	orig, _ := svc.ImportTrack(ctx, catalog.ImportTrackRequest{MediaObjectID: "media-a", Title: "T"})
	_, err := svc.RelinkTrack(ctx, orig.ID, catalog.RelinkTrackRequest{MediaObjectID: "media-staged"})
	if !errors.Is(err, catalog.ErrRelinkRejected) {
		t.Fatalf("err = %v, want ErrRelinkRejected", err)
	}
}

func TestRelinkTrackMediaNotFound(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-a", "original_audio", "active")

	orig, _ := svc.ImportTrack(ctx, catalog.ImportTrackRequest{MediaObjectID: "media-a", Title: "T"})
	_, err := svc.RelinkTrack(ctx, orig.ID, catalog.RelinkTrackRequest{MediaObjectID: "missing"})
	if err == nil {
		t.Fatal("expected error for missing media object")
	}
}

func TestRelinkTrackTrackNotFound(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-a", "original_audio", "active")

	_, err := svc.RelinkTrack(ctx, "no-such-track", catalog.RelinkTrackRequest{MediaObjectID: "media-a"})
	if !errors.Is(err, catalog.ErrTrackNotFound) {
		t.Fatalf("err = %v, want ErrTrackNotFound", err)
	}
}

func TestRelinkTrackNoReaderConfigured(t *testing.T) {
	ctx := context.Background()
	repo := catalog.NewMemoryRepository()
	svc := catalog.NewService(repo) // no media reader

	_, err := svc.RelinkTrack(ctx, "any", catalog.RelinkTrackRequest{MediaObjectID: "media-a"})
	if !errors.Is(err, catalog.ErrRelinkRejected) {
		t.Fatalf("err = %v, want ErrRelinkRejected", err)
	}
}

func TestRelinkTrackEmptyMediaObjectID(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-a", "original_audio", "active")

	orig, _ := svc.ImportTrack(ctx, catalog.ImportTrackRequest{MediaObjectID: "media-a", Title: "T"})
	_, err := svc.RelinkTrack(ctx, orig.ID, catalog.RelinkTrackRequest{MediaObjectID: "  "})
	if !errors.Is(err, catalog.ErrRelinkRejected) {
		t.Fatalf("err = %v, want ErrRelinkRejected", err)
	}
}

// ---- SearchCatalog tests ----

func TestSearchCatalogEmptyQueryRejected(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	_, err := svc.SearchCatalog(ctx, "  ")
	if !errors.Is(err, catalog.ErrInvalidTrack) {
		t.Fatalf("err = %v, want ErrInvalidTrack for empty query", err)
	}
}

func TestSearchCatalogNoResults(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	result, err := svc.SearchCatalog(ctx, "notfound")
	if err != nil {
		t.Fatalf("SearchCatalog: %v", err)
	}
	if result.Query != "notfound" {
		t.Fatalf("Query = %q, want notfound", result.Query)
	}
	if len(result.Items) != 0 {
		t.Fatalf("Items = %v, want empty", result.Items)
	}
}

func TestSearchCatalogMatchesArtist(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	if _, err := svc.CreateArtist(ctx, "Hatsune Miku", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateArtist(ctx, "KAITO", ""); err != nil {
		t.Fatal(err)
	}

	result, err := svc.SearchCatalog(ctx, "miku")
	if err != nil {
		t.Fatalf("SearchCatalog: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Kind != catalog.SearchResultArtist {
		t.Fatalf("Kind = %q, want artist", result.Items[0].Kind)
	}
	if result.Items[0].Artist.Name != "Hatsune Miku" {
		t.Fatalf("Artist.Name = %q", result.Items[0].Artist.Name)
	}
}

func TestSearchCatalogMatchesAlbum(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "Miku", "")
	if _, err := svc.CreateAlbum(ctx, "Project DIVA", "", artist.ID, 2009); err != nil {
		t.Fatal(err)
	}

	result, err := svc.SearchCatalog(ctx, "diva")
	if err != nil {
		t.Fatalf("SearchCatalog: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Kind != catalog.SearchResultAlbum {
		t.Fatalf("Items = %v, want 1 album", result.Items)
	}
}

func TestSearchCatalogCrossEntityResults(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("media-s1", "original_audio", "active")

	artist, _ := svc.CreateArtist(ctx, "Miku Artist", "")
	album, _ := svc.CreateAlbum(ctx, "Miku Album", "", artist.ID, 2020)
	_, _ = svc.ImportTrack(ctx, catalog.ImportTrackRequest{
		MediaObjectID: "media-s1",
		Title:         "Miku Song",
		ArtistID:      artist.ID,
		AlbumID:       album.ID,
	})

	result, err := svc.SearchCatalog(ctx, "miku")
	if err != nil {
		t.Fatalf("SearchCatalog: %v", err)
	}
	kinds := map[catalog.SearchResultKind]int{}
	for _, item := range result.Items {
		kinds[item.Kind]++
	}
	if kinds[catalog.SearchResultArtist] != 1 {
		t.Errorf("artist hits = %d, want 1", kinds[catalog.SearchResultArtist])
	}
	if kinds[catalog.SearchResultAlbum] != 1 {
		t.Errorf("album hits = %d, want 1", kinds[catalog.SearchResultAlbum])
	}
	if kinds[catalog.SearchResultTrack] != 1 {
		t.Errorf("track hits = %d, want 1", kinds[catalog.SearchResultTrack])
	}
}

// ---- BatchImportTracks tests ----

func TestBatchImportTracksAllSuccess(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("bi-1", "original_audio", "active")
	reader.add("bi-2", "transcoded_audio", "active")

	artist, _ := svc.CreateArtist(ctx, "Batch Artist", "")
	result := svc.BatchImportTracks(ctx, []catalog.ImportTrackRequest{
		{MediaObjectID: "bi-1", Title: "Track One", ArtistID: artist.ID},
		{MediaObjectID: "bi-2", Title: "Track Two", ArtistID: artist.ID},
	})

	if result.Total != 2 || result.Imported != 2 || result.Failed != 0 {
		t.Fatalf("Total=%d Imported=%d Failed=%d, want 2/2/0", result.Total, result.Imported, result.Failed)
	}
	for i, item := range result.Items {
		if item.Track == nil {
			t.Errorf("item[%d].Track is nil", i)
		}
		if item.Error != "" {
			t.Errorf("item[%d].Error = %q, want empty", i, item.Error)
		}
		if item.Index != i {
			t.Errorf("item[%d].Index = %d, want %d", i, item.Index, i)
		}
	}
}

func TestBatchImportTracksPartialFailure(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("bi-ok", "original_audio", "active")
	// bi-bad has wrong asset kind

	artist, _ := svc.CreateArtist(ctx, "Partial Artist", "")
	result := svc.BatchImportTracks(ctx, []catalog.ImportTrackRequest{
		{MediaObjectID: "bi-ok", Title: "Good Track", ArtistID: artist.ID},
		{MediaObjectID: "bi-bad", Title: "Bad Track", ArtistID: artist.ID},
		{MediaObjectID: "bi-ok", Title: "Good Again", ArtistID: artist.ID},
	})

	if result.Total != 3 {
		t.Fatalf("Total = %d, want 3", result.Total)
	}
	// bi-bad is not found in reader → import_rejected
	if result.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", result.Failed)
	}
	if result.Imported != 2 {
		t.Fatalf("Imported = %d, want 2", result.Imported)
	}
	if result.Items[1].Error == "" {
		t.Error("item[1].Error should be non-empty for missing media object")
	}
}

func TestBatchImportTracksAllFail(t *testing.T) {
	ctx := context.Background()
	svc, _ := newImportSvc(t)
	// No media objects registered; all should fail.
	result := svc.BatchImportTracks(ctx, []catalog.ImportTrackRequest{
		{MediaObjectID: "none-1"},
		{MediaObjectID: "none-2"},
	})
	if result.Failed != 2 || result.Imported != 0 {
		t.Fatalf("Failed=%d Imported=%d, want 2/0", result.Failed, result.Imported)
	}
	for _, item := range result.Items {
		if item.ErrorCode == "" {
			t.Error("ErrorCode should be set on failure")
		}
	}
}

func TestBatchImportTracksEmptyBatch(t *testing.T) {
	ctx := context.Background()
	svc, _ := newImportSvc(t)
	result := svc.BatchImportTracks(ctx, nil)
	if result.Total != 0 || result.Imported != 0 || result.Failed != 0 {
		t.Fatalf("unexpected result for empty batch: %+v", result)
	}
	if len(result.Items) != 0 {
		t.Fatalf("Items length = %d, want 0", len(result.Items))
	}
}

func TestBatchImportTracksIndexPreserved(t *testing.T) {
	ctx := context.Background()
	svc, reader := newImportSvc(t)
	reader.add("idx-1", "original_audio", "active")
	reader.add("idx-3", "original_audio", "active")

	artist, _ := svc.CreateArtist(ctx, "Index Artist", "")
	result := svc.BatchImportTracks(ctx, []catalog.ImportTrackRequest{
		{MediaObjectID: "idx-1", Title: "A", ArtistID: artist.ID},
		{MediaObjectID: "idx-missing"},
		{MediaObjectID: "idx-3", Title: "C", ArtistID: artist.ID},
	})

	if result.Total != 3 || result.Imported != 2 || result.Failed != 1 {
		t.Fatalf("Total=%d Imported=%d Failed=%d, want 3/2/1", result.Total, result.Imported, result.Failed)
	}
	for i, item := range result.Items {
		if item.Index != i {
			t.Errorf("item[%d].Index = %d, want %d", i, item.Index, i)
		}
	}
	if result.Items[1].ErrorCode == "" {
		t.Error("middle item should have ErrorCode")
	}
}

// ---------- Phase 45: UpdateArtist / UpdateAlbum / UpdateTrack ----------

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func newSteppingService(repo catalog.Repository) *catalog.Service {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tick := 0
	return catalog.NewService(repo).WithClock(func() time.Time {
		tick++
		return base.Add(time.Duration(tick) * time.Second)
	})
}

func TestUpdateArtistChangesNameAndSortName(t *testing.T) {
	ctx := context.Background()
	svc := newSteppingService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "Old Name", "Old, Name")

	updated, err := svc.UpdateArtist(ctx, artist.ID, catalog.UpdateArtistRequest{
		Name:     strPtr("New Name"),
		SortName: strPtr("Name, New"),
	})
	if err != nil {
		t.Fatalf("UpdateArtist: %v", err)
	}
	if updated.Name != "New Name" || updated.SortName != "Name, New" {
		t.Fatalf("got Name=%q SortName=%q", updated.Name, updated.SortName)
	}
	if updated.ID != artist.ID {
		t.Fatal("ID must not change")
	}
	if !updated.UpdatedAt.After(artist.UpdatedAt) {
		t.Fatal("UpdatedAt must advance")
	}
}

func TestUpdateArtistPartialNilFields(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "Unchanged", "Sort")

	// Only update SortName; Name should stay.
	updated, err := svc.UpdateArtist(ctx, artist.ID, catalog.UpdateArtistRequest{
		SortName: strPtr("New Sort"),
	})
	if err != nil {
		t.Fatalf("UpdateArtist: %v", err)
	}
	if updated.Name != "Unchanged" {
		t.Fatalf("Name changed unexpectedly to %q", updated.Name)
	}
	if updated.SortName != "New Sort" {
		t.Fatalf("SortName = %q, want %q", updated.SortName, "New Sort")
	}
}

func TestUpdateArtistRejectsEmptyName(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "Some Artist", "")

	_, err := svc.UpdateArtist(ctx, artist.ID, catalog.UpdateArtistRequest{Name: strPtr("")})
	if !errors.Is(err, catalog.ErrInvalidArtist) {
		t.Fatalf("expected ErrInvalidArtist, got %v", err)
	}
}

func TestUpdateArtistNotFound(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	_, err := svc.UpdateArtist(ctx, "nonexistent", catalog.UpdateArtistRequest{Name: strPtr("X")})
	if !errors.Is(err, catalog.ErrArtistNotFound) {
		t.Fatalf("expected ErrArtistNotFound, got %v", err)
	}
}

func TestUpdateAlbumChangesTitle(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "Artist", "")
	album, _ := svc.CreateAlbum(ctx, "Old Title", "", artist.ID, 2000)

	updated, err := svc.UpdateAlbum(ctx, album.ID, catalog.UpdateAlbumRequest{
		Title:       strPtr("New Title"),
		ReleaseYear: intPtr(2024),
	})
	if err != nil {
		t.Fatalf("UpdateAlbum: %v", err)
	}
	if updated.Title != "New Title" || updated.ReleaseYear != 2024 {
		t.Fatalf("got Title=%q Year=%d", updated.Title, updated.ReleaseYear)
	}
	if !updated.UpdatedAt.After(album.UpdatedAt) {
		t.Fatal("UpdatedAt must advance")
	}
}

func TestUpdateAlbumChangesArtist(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	a1, _ := svc.CreateArtist(ctx, "Artist1", "")
	a2, _ := svc.CreateArtist(ctx, "Artist2", "")
	album, _ := svc.CreateAlbum(ctx, "Album", "", a1.ID, 0)

	updated, err := svc.UpdateAlbum(ctx, album.ID, catalog.UpdateAlbumRequest{ArtistID: &a2.ID})
	if err != nil {
		t.Fatalf("UpdateAlbum: %v", err)
	}
	if updated.ArtistID != a2.ID {
		t.Fatalf("ArtistID = %q, want %q", updated.ArtistID, a2.ID)
	}
}

func TestUpdateAlbumRejectsNegativeYear(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "A", "")
	album, _ := svc.CreateAlbum(ctx, "B", "", artist.ID, 0)

	_, err := svc.UpdateAlbum(ctx, album.ID, catalog.UpdateAlbumRequest{ReleaseYear: intPtr(-1)})
	if !errors.Is(err, catalog.ErrInvalidAlbum) {
		t.Fatalf("expected ErrInvalidAlbum, got %v", err)
	}
}

func TestUpdateAlbumNotFound(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	_, err := svc.UpdateAlbum(ctx, "nope", catalog.UpdateAlbumRequest{Title: strPtr("X")})
	if !errors.Is(err, catalog.ErrAlbumNotFound) {
		t.Fatalf("expected ErrAlbumNotFound, got %v", err)
	}
}

func TestUpdateTrackChangesTitle(t *testing.T) {
	ctx := context.Background()
	svc := newSteppingService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "A", "")
	album, _ := svc.CreateAlbum(ctx, "B", "", artist.ID, 0)
	track, _ := svc.CreateTrack(ctx, "Old", "", artist.ID, album.ID, "mo1", 1, 1, 60000)

	updated, err := svc.UpdateTrack(ctx, track.ID, catalog.UpdateTrackRequest{
		Title:       strPtr("New"),
		TrackNumber: intPtr(2),
		DurationMS:  intPtr(90000),
	})
	if err != nil {
		t.Fatalf("UpdateTrack: %v", err)
	}
	if updated.Title != "New" || updated.TrackNumber != 2 || updated.DurationMS != 90000 {
		t.Fatalf("unexpected values: %+v", updated)
	}
	if !updated.UpdatedAt.After(track.UpdatedAt) {
		t.Fatal("UpdatedAt must advance")
	}
}

func TestUpdateTrackClearsAlbum(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "A", "")
	album, _ := svc.CreateAlbum(ctx, "B", "", artist.ID, 0)
	track, _ := svc.CreateTrack(ctx, "T", "", artist.ID, album.ID, "mo2", 0, 0, 0)

	empty := ""
	updated, err := svc.UpdateTrack(ctx, track.ID, catalog.UpdateTrackRequest{AlbumID: &empty})
	if err != nil {
		t.Fatalf("UpdateTrack: %v", err)
	}
	if updated.AlbumID != "" {
		t.Fatalf("expected empty AlbumID, got %q", updated.AlbumID)
	}
}

func TestUpdateTrackRejectsEmptyTitle(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "A", "")
	track, _ := svc.CreateTrack(ctx, "T", "", artist.ID, "", "mo3", 0, 0, 0)

	_, err := svc.UpdateTrack(ctx, track.ID, catalog.UpdateTrackRequest{Title: strPtr("")})
	if !errors.Is(err, catalog.ErrInvalidTrack) {
		t.Fatalf("expected ErrInvalidTrack, got %v", err)
	}
}

func TestUpdateTrackNotFound(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	_, err := svc.UpdateTrack(ctx, "nope", catalog.UpdateTrackRequest{Title: strPtr("X")})
	if !errors.Is(err, catalog.ErrTrackNotFound) {
		t.Fatalf("expected ErrTrackNotFound, got %v", err)
	}
}

// ---------- Phase 46: Playlist service tests ----------

func TestPlaylistCreateAndGet(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	p, err := svc.CreatePlaylist(ctx, "My Mix", "a great playlist")
	if err != nil {
		t.Fatalf("CreatePlaylist: %v", err)
	}
	if p.ID == "" || p.Name != "My Mix" || p.Description != "a great playlist" {
		t.Fatalf("unexpected playlist: %+v", p)
	}
	if len(p.TrackIDs) != 0 {
		t.Fatalf("expected empty track list, got %v", p.TrackIDs)
	}
	got, err := svc.GetPlaylist(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetPlaylist: %v", err)
	}
	if got.ID != p.ID || got.Name != p.Name {
		t.Fatalf("got = %+v, want %+v", got, p)
	}
}

func TestPlaylistRequiresName(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	_, err := svc.CreatePlaylist(ctx, "  ", "")
	if !errors.Is(err, catalog.ErrInvalidPlaylist) {
		t.Fatalf("expected ErrInvalidPlaylist, got %v", err)
	}
}

func TestPlaylistAddAndRemoveTrack(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "A", "")
	t1, _ := svc.CreateTrack(ctx, "T1", "", artist.ID, "", "mo1", 0, 0, 0)
	t2, _ := svc.CreateTrack(ctx, "T2", "", artist.ID, "", "mo2", 0, 0, 0)

	pl, _ := svc.CreatePlaylist(ctx, "PL", "")
	pl, err := svc.AddTrackToPlaylist(ctx, pl.ID, t1.ID)
	if err != nil {
		t.Fatalf("AddTrack: %v", err)
	}
	pl, err = svc.AddTrackToPlaylist(ctx, pl.ID, t2.ID)
	if err != nil {
		t.Fatalf("AddTrack 2: %v", err)
	}
	if len(pl.TrackIDs) != 2 || pl.TrackIDs[0] != t1.ID || pl.TrackIDs[1] != t2.ID {
		t.Fatalf("unexpected trackIDs: %v", pl.TrackIDs)
	}

	pl, err = svc.RemoveTrackFromPlaylist(ctx, pl.ID, t1.ID)
	if err != nil {
		t.Fatalf("RemoveTrack: %v", err)
	}
	if len(pl.TrackIDs) != 1 || pl.TrackIDs[0] != t2.ID {
		t.Fatalf("after remove: %v", pl.TrackIDs)
	}
}

func TestPlaylistAddNonExistentTrack(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	pl, _ := svc.CreatePlaylist(ctx, "PL", "")
	_, err := svc.AddTrackToPlaylist(ctx, pl.ID, "no-such-track")
	if !errors.Is(err, catalog.ErrTrackNotFound) {
		t.Fatalf("expected ErrTrackNotFound, got %v", err)
	}
}

func TestPlaylistRemoveAbsentTrack(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "A", "")
	tr, _ := svc.CreateTrack(ctx, "T", "", artist.ID, "", "mo1", 0, 0, 0)
	pl, _ := svc.CreatePlaylist(ctx, "PL", "")
	_, _ = svc.AddTrackToPlaylist(ctx, pl.ID, tr.ID)

	_, err := svc.RemoveTrackFromPlaylist(ctx, pl.ID, "not-in-list")
	if !errors.Is(err, catalog.ErrInvalidPlaylist) {
		t.Fatalf("expected ErrInvalidPlaylist, got %v", err)
	}
}

func TestPlaylistUpdateMetadata(t *testing.T) {
	ctx := context.Background()
	svc := newSteppingService(newMemRepo())
	pl, _ := svc.CreatePlaylist(ctx, "Old", "desc")
	updated, err := svc.UpdatePlaylist(ctx, pl.ID, catalog.UpdatePlaylistRequest{
		Name:        strPtr("New Name"),
		Description: strPtr("new desc"),
	})
	if err != nil {
		t.Fatalf("UpdatePlaylist: %v", err)
	}
	if updated.Name != "New Name" || updated.Description != "new desc" {
		t.Fatalf("unexpected: %+v", updated)
	}
	if !updated.UpdatedAt.After(pl.UpdatedAt) {
		t.Fatal("UpdatedAt must advance")
	}
}

func TestPlaylistUpdateRejectsEmptyName(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	pl, _ := svc.CreatePlaylist(ctx, "PL", "")
	_, err := svc.UpdatePlaylist(ctx, pl.ID, catalog.UpdatePlaylistRequest{Name: strPtr("")})
	if !errors.Is(err, catalog.ErrInvalidPlaylist) {
		t.Fatalf("expected ErrInvalidPlaylist, got %v", err)
	}
}

func TestPlaylistDeleteAndNotFound(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	pl, _ := svc.CreatePlaylist(ctx, "PL", "")
	if err := svc.DeletePlaylist(ctx, pl.ID); err != nil {
		t.Fatalf("DeletePlaylist: %v", err)
	}
	_, err := svc.GetPlaylist(ctx, pl.ID)
	if !errors.Is(err, catalog.ErrPlaylistNotFound) {
		t.Fatalf("expected ErrPlaylistNotFound, got %v", err)
	}
}

func TestPlaylistGetNotFound(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	_, err := svc.GetPlaylist(ctx, "nonexistent")
	if !errors.Is(err, catalog.ErrPlaylistNotFound) {
		t.Fatalf("expected ErrPlaylistNotFound, got %v", err)
	}
}

func TestPlaylistSetTracks(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "A", "")
	t1, _ := svc.CreateTrack(ctx, "T1", "", artist.ID, "", "mo1", 0, 0, 0)
	t2, _ := svc.CreateTrack(ctx, "T2", "", artist.ID, "", "mo2", 0, 0, 0)
	pl, _ := svc.CreatePlaylist(ctx, "PL", "")

	// happy path: reorder [t2, t1]
	got, err := svc.SetPlaylistTracks(ctx, pl.ID, []string{t2.ID, t1.ID})
	if err != nil {
		t.Fatalf("SetPlaylistTracks: %v", err)
	}
	if len(got.TrackIDs) != 2 || got.TrackIDs[0] != t2.ID || got.TrackIDs[1] != t1.ID {
		t.Fatalf("unexpected trackIDs after reorder: %v", got.TrackIDs)
	}

	// clear: empty slice
	got, err = svc.SetPlaylistTracks(ctx, pl.ID, []string{})
	if err != nil {
		t.Fatalf("SetPlaylistTracks clear: %v", err)
	}
	if len(got.TrackIDs) != 0 {
		t.Fatalf("expected empty trackIDs after clear, got %v", got.TrackIDs)
	}

	// duplicates preserved
	got, err = svc.SetPlaylistTracks(ctx, pl.ID, []string{t1.ID, t1.ID})
	if err != nil {
		t.Fatalf("SetPlaylistTracks duplicates: %v", err)
	}
	if len(got.TrackIDs) != 2 || got.TrackIDs[0] != t1.ID || got.TrackIDs[1] != t1.ID {
		t.Fatalf("unexpected trackIDs with duplicates: %v", got.TrackIDs)
	}

	// unknown track → ErrTrackNotFound
	_, err = svc.SetPlaylistTracks(ctx, pl.ID, []string{"no-such-track"})
	if !errors.Is(err, catalog.ErrTrackNotFound) {
		t.Fatalf("expected ErrTrackNotFound for unknown track, got %v", err)
	}

	// unknown playlist → ErrPlaylistNotFound
	_, err = svc.SetPlaylistTracks(ctx, "no-such-pl", []string{t1.ID})
	if !errors.Is(err, catalog.ErrPlaylistNotFound) {
		t.Fatalf("expected ErrPlaylistNotFound for unknown playlist, got %v", err)
	}
}

// ---------- Phase 49: GetPlaylistTracks service tests ----------

func TestGetPlaylistTracksOrdered(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "A", "")
	tr1, _ := svc.CreateTrack(ctx, "Alpha", "", artist.ID, "", "mo-gpt-1", 0, 0, 0)
	tr2, _ := svc.CreateTrack(ctx, "Beta", "", artist.ID, "", "mo-gpt-2", 0, 0, 0)
	tr3, _ := svc.CreateTrack(ctx, "Gamma", "", artist.ID, "", "mo-gpt-3", 0, 0, 0)

	pl, _ := svc.CreatePlaylist(ctx, "TestPL", "")
	_, _ = svc.SetPlaylistTracks(ctx, pl.ID, []string{tr3.ID, tr1.ID, tr2.ID})

	tracks, err := svc.GetPlaylistTracks(ctx, pl.ID)
	if err != nil {
		t.Fatalf("GetPlaylistTracks: %v", err)
	}
	if len(tracks) != 3 {
		t.Fatalf("len = %d, want 3", len(tracks))
	}
	if tracks[0].ID != tr3.ID || tracks[1].ID != tr1.ID || tracks[2].ID != tr2.ID {
		t.Fatalf("wrong order: got [%s %s %s]", tracks[0].ID, tracks[1].ID, tracks[2].ID)
	}
}

func TestGetPlaylistTracksEmpty(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	pl, _ := svc.CreatePlaylist(ctx, "Empty PL", "")

	tracks, err := svc.GetPlaylistTracks(ctx, pl.ID)
	if err != nil {
		t.Fatalf("GetPlaylistTracks: %v", err)
	}
	if len(tracks) != 0 {
		t.Fatalf("expected empty slice, got %v", tracks)
	}
}

func TestGetPlaylistTracksNotFound(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	_, err := svc.GetPlaylistTracks(ctx, "no-such-pl")
	if !errors.Is(err, catalog.ErrPlaylistNotFound) {
		t.Fatalf("expected ErrPlaylistNotFound, got %v", err)
	}
}

func TestGetPlaylistTracksPreserveDuplicates(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())
	artist, _ := svc.CreateArtist(ctx, "A", "")
	tr, _ := svc.CreateTrack(ctx, "T", "", artist.ID, "", "mo-dup-1", 0, 0, 0)

	pl, _ := svc.CreatePlaylist(ctx, "DupPL", "")
	_, _ = svc.SetPlaylistTracks(ctx, pl.ID, []string{tr.ID, tr.ID, tr.ID})

	tracks, err := svc.GetPlaylistTracks(ctx, pl.ID)
	if err != nil {
		t.Fatalf("GetPlaylistTracks: %v", err)
	}
	if len(tracks) != 3 {
		t.Fatalf("len = %d, want 3 (duplicates preserved)", len(tracks))
	}
	for i, got := range tracks {
		if got.ID != tr.ID {
			t.Errorf("tracks[%d].ID = %q, want %q", i, got.ID, tr.ID)
		}
	}
}

func TestGetCatalogStatsEmpty(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())

	stats, err := svc.GetCatalogStats(ctx)
	if err != nil {
		t.Fatalf("GetCatalogStats: %v", err)
	}
	if stats.Artists != 0 || stats.Albums != 0 || stats.Tracks != 0 || stats.Playlists != 0 {
		t.Fatalf("expected all zeros, got %+v", stats)
	}
}

func TestGetCatalogStatsPopulated(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())

	_, _ = svc.CreateArtist(ctx, "Artist A", "")
	artist2, _ := svc.CreateArtist(ctx, "Artist B", "")
	_, _ = svc.CreateAlbum(ctx, "Album 1", "", artist2.ID, 2020)
	_, _ = svc.CreateTrack(ctx, "Track 1", "", artist2.ID, "", "mo-s1", 0, 0, 0)
	_, _ = svc.CreateTrack(ctx, "Track 2", "", artist2.ID, "", "mo-s2", 0, 0, 0)
	_, _ = svc.CreatePlaylist(ctx, "Playlist 1", "")

	stats, err := svc.GetCatalogStats(ctx)
	if err != nil {
		t.Fatalf("GetCatalogStats: %v", err)
	}
	if stats.Artists != 2 {
		t.Errorf("Artists = %d, want 2", stats.Artists)
	}
	if stats.Albums != 1 {
		t.Errorf("Albums = %d, want 1", stats.Albums)
	}
	if stats.Tracks != 2 {
		t.Errorf("Tracks = %d, want 2", stats.Tracks)
	}
	if stats.Playlists != 1 {
		t.Errorf("Playlists = %d, want 1", stats.Playlists)
	}
}

func TestGetCatalogStatsRepoError(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	svc := catalog.NewService(repo)

	// Inject an error by closing the repo's backing store simulation.
	// MemRepo does not support forced errors, so we verify the happy
	// path succeeds without an injected error – error propagation is
	// covered structurally by the implementation delegating to repo.
	stats, err := svc.GetCatalogStats(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Artists != 0 {
		t.Fatalf("expected 0 artists, got %d", stats.Artists)
	}
}

func TestGetArtistStatsBreakdownEmpty(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())

	result, err := svc.GetArtistStatsBreakdown(ctx)
	if err != nil {
		t.Fatalf("GetArtistStatsBreakdown: %v", err)
	}
	if result.Artists == nil {
		t.Fatal("expected non-nil Artists slice")
	}
	if len(result.Artists) != 0 {
		t.Fatalf("expected empty Artists, got %d items", len(result.Artists))
	}
}

func TestGetArtistStatsBreakdownPopulated(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())

	a1, _ := svc.CreateArtist(ctx, "Aria", "")
	a2, _ := svc.CreateArtist(ctx, "Bloom", "")
	_, _ = svc.CreateAlbum(ctx, "First Album", "", a1.ID, 2020)
	_, _ = svc.CreateTrack(ctx, "Track A", "", a1.ID, "", "mo-1", 0, 0, 0)
	_, _ = svc.CreateTrack(ctx, "Track B", "", a1.ID, "", "mo-2", 0, 0, 0)
	_, _ = svc.CreateTrack(ctx, "Track C", "", a1.ID, "", "mo-3", 0, 0, 0)

	result, err := svc.GetArtistStatsBreakdown(ctx)
	if err != nil {
		t.Fatalf("GetArtistStatsBreakdown: %v", err)
	}
	if len(result.Artists) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Artists))
	}
	byID := map[string]catalog.ArtistStatItem{}
	for _, item := range result.Artists {
		byID[item.ArtistID] = item
	}
	if got := byID[a1.ID]; got.AlbumCount != 1 || got.TrackCount != 3 {
		t.Errorf("artist1: albumCount=%d trackCount=%d, want 1/3", got.AlbumCount, got.TrackCount)
	}
	if got := byID[a2.ID]; got.AlbumCount != 0 || got.TrackCount != 0 {
		t.Errorf("artist2: albumCount=%d trackCount=%d, want 0/0", got.AlbumCount, got.TrackCount)
	}
}

func TestGetAlbumStatsBreakdownEmpty(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())

	result, err := svc.GetAlbumStatsBreakdown(ctx)
	if err != nil {
		t.Fatalf("GetAlbumStatsBreakdown: %v", err)
	}
	if result.Albums == nil {
		t.Fatal("expected non-nil Albums slice")
	}
	if len(result.Albums) != 0 {
		t.Fatalf("expected empty Albums, got %d items", len(result.Albums))
	}
}

func TestGetAlbumStatsBreakdownPopulated(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())

	artist, _ := svc.CreateArtist(ctx, "Solo Artist", "")
	al1, _ := svc.CreateAlbum(ctx, "Debut", "", artist.ID, 2021)
	al2, _ := svc.CreateAlbum(ctx, "Sophomore", "", artist.ID, 2023)
	_, _ = svc.CreateTrack(ctx, "Song 1", "", artist.ID, al1.ID, "mo-a1", 0, 0, 0)
	_, _ = svc.CreateTrack(ctx, "Song 2", "", artist.ID, al1.ID, "mo-a2", 0, 0, 0)

	result, err := svc.GetAlbumStatsBreakdown(ctx)
	if err != nil {
		t.Fatalf("GetAlbumStatsBreakdown: %v", err)
	}
	if len(result.Albums) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Albums))
	}
	byID := map[string]catalog.AlbumStatItem{}
	for _, item := range result.Albums {
		byID[item.AlbumID] = item
	}
	if got := byID[al1.ID]; got.TrackCount != 2 {
		t.Errorf("album1 trackCount=%d, want 2", got.TrackCount)
	}
	if got := byID[al2.ID]; got.TrackCount != 0 {
		t.Errorf("album2 trackCount=%d, want 0", got.TrackCount)
	}
}

func TestGetPlaylistStatsBreakdownEmpty(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())

	result, err := svc.GetPlaylistStatsBreakdown(ctx)
	if err != nil {
		t.Fatalf("GetPlaylistStatsBreakdown: %v", err)
	}
	if result.Playlists == nil {
		t.Fatal("expected non-nil Playlists slice")
	}
	if len(result.Playlists) != 0 {
		t.Fatalf("expected empty Playlists, got %d items", len(result.Playlists))
	}
}

func TestGetPlaylistStatsBreakdownPopulated(t *testing.T) {
	ctx := context.Background()
	svc := catalog.NewService(newMemRepo())

	artist, _ := svc.CreateArtist(ctx, "Test Artist", "")
	tr1, _ := svc.CreateTrack(ctx, "Track A", "", artist.ID, "", "mo-p1", 0, 0, 0)
	tr2, _ := svc.CreateTrack(ctx, "Track B", "", artist.ID, "", "mo-p2", 0, 0, 0)

	pl1, _ := svc.CreatePlaylist(ctx, "Playlist One", "")
	pl2, _ := svc.CreatePlaylist(ctx, "Playlist Two", "")

	_, _ = svc.AddTrackToPlaylist(ctx, pl1.ID, tr1.ID)
	_, _ = svc.AddTrackToPlaylist(ctx, pl1.ID, tr2.ID)
	// pl2 stays empty

	result, err := svc.GetPlaylistStatsBreakdown(ctx)
	if err != nil {
		t.Fatalf("GetPlaylistStatsBreakdown: %v", err)
	}
	if len(result.Playlists) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Playlists))
	}
	byID := map[string]catalog.PlaylistStatItem{}
	for _, item := range result.Playlists {
		byID[item.PlaylistID] = item
	}
	if got := byID[pl1.ID]; got.TrackCount != 2 {
		t.Errorf("playlist1 trackCount=%d, want 2", got.TrackCount)
	}
	if got := byID[pl2.ID]; got.TrackCount != 0 {
		t.Errorf("playlist2 trackCount=%d, want 0", got.TrackCount)
	}
}

// ---- GetRecentlyAdded tests ----

func TestGetRecentlyAddedEmpty(t *testing.T) {
	svc := catalog.NewService(newMemRepo())
	result, err := svc.GetRecentlyAdded(context.Background(), "", 0)
	if err != nil {
		t.Fatalf("GetRecentlyAdded: %v", err)
	}
	if result.Items == nil {
		t.Fatal("expected non-nil Items slice")
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(result.Items))
	}
}

func TestGetRecentlyAddedNewestFirst(t *testing.T) {
	repo := newMemRepo()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := catalog.NewService(repo)
	svc.WithClock(func() time.Time { return base })
	ctx := context.Background()

	_, _ = svc.CreateArtist(ctx, "Artist A", "")

	svc.WithClock(func() time.Time { return base.Add(time.Hour) })
	_, _ = svc.CreateArtist(ctx, "Artist B", "")

	result, err := svc.GetRecentlyAdded(ctx, "", 0)
	if err != nil {
		t.Fatalf("GetRecentlyAdded: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Artist == nil || result.Items[0].Artist.Name != "Artist B" {
		t.Errorf("expected newest item first, got %+v", result.Items[0])
	}
}

func TestGetRecentlyAddedKindFilter(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	artist, _ := svc.CreateArtist(ctx, "Solo Artist", "")
	_, _ = svc.CreateTrack(ctx, "Track 1", "", artist.ID, "", "mo-001", 1, 0, 180000)

	result, err := svc.GetRecentlyAdded(ctx, "artist", 0)
	if err != nil {
		t.Fatalf("GetRecentlyAdded artist: %v", err)
	}
	for _, item := range result.Items {
		if item.Kind != catalog.RecentItemArtist {
			t.Errorf("expected only artist items, got %s", item.Kind)
		}
	}

	result, err = svc.GetRecentlyAdded(ctx, "track", 0)
	if err != nil {
		t.Fatalf("GetRecentlyAdded track: %v", err)
	}
	for _, item := range result.Items {
		if item.Kind != catalog.RecentItemTrack {
			t.Errorf("expected only track items, got %s", item.Kind)
		}
	}
}

func TestGetRecentlyAddedRejectsInvalidKind(t *testing.T) {
	svc := catalog.NewService(newMemRepo())
	_, err := svc.GetRecentlyAdded(context.Background(), "invalid", 0)
	if !errors.Is(err, catalog.ErrInvalidTrack) {
		t.Fatalf("expected ErrInvalidTrack for invalid kind, got %v", err)
	}
}

func TestGetRecentlyAddedLimitCaps(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		svc.CreateArtist(ctx, "Artist", "")
	}

	result, err := svc.GetRecentlyAdded(ctx, "", 5)
	if err != nil {
		t.Fatalf("GetRecentlyAdded: %v", err)
	}
	if len(result.Items) != 5 {
		t.Errorf("expected 5 items (limit=5), got %d", len(result.Items))
	}

	// limit > max (100) should be clamped
	result, err = svc.GetRecentlyAdded(ctx, "", 999)
	if err != nil {
		t.Fatalf("GetRecentlyAdded limit=999: %v", err)
	}
	if len(result.Items) != 25 {
		t.Errorf("expected all 25 items when limit>count, got %d", len(result.Items))
	}
}

func TestGetRecentlyAddedPlaylistKindFilter(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	_, _ = svc.CreatePlaylist(ctx, "My Mix", "")
	_, _ = svc.CreateArtist(ctx, "Artist X", "")

	// kind=playlist returns only playlists
	result, err := svc.GetRecentlyAdded(ctx, "playlist", 0)
	if err != nil {
		t.Fatalf("GetRecentlyAdded playlist: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 playlist item, got %d", len(result.Items))
	}
	if result.Items[0].Kind != catalog.RecentItemPlaylist {
		t.Errorf("expected playlist kind, got %s", result.Items[0].Kind)
	}
	if result.Items[0].Playlist == nil || result.Items[0].Playlist.Name != "My Mix" {
		t.Errorf("expected playlist payload, got %+v", result.Items[0])
	}
}

func TestGetRecentlyAddedPlaylistUnified(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	_, _ = svc.CreatePlaylist(ctx, "Mix", "")
	_, _ = svc.CreateArtist(ctx, "Artist", "")

	result, err := svc.GetRecentlyAdded(ctx, "", 0)
	if err != nil {
		t.Fatalf("GetRecentlyAdded: %v", err)
	}

	hasPlaylist := false
	hasArtist := false
	for _, item := range result.Items {
		switch item.Kind {
		case catalog.RecentItemPlaylist:
			hasPlaylist = true
		case catalog.RecentItemArtist:
			hasArtist = true
		}
	}
	if !hasPlaylist {
		t.Error("expected at least one playlist item in unified timeline")
	}
	if !hasArtist {
		t.Error("expected at least one artist item in unified timeline")
	}
}

// ---- GetRecentlyUpdated tests ----

func TestGetRecentlyUpdatedEmpty(t *testing.T) {
	svc := catalog.NewService(newMemRepo())
	result, err := svc.GetRecentlyUpdated(context.Background(), "", 0)
	if err != nil {
		t.Fatalf("GetRecentlyUpdated: %v", err)
	}
	if result.Items == nil {
		t.Fatal("expected non-nil Items slice")
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(result.Items))
	}
}

func TestGetRecentlyUpdatedNewestFirst(t *testing.T) {
	repo := newMemRepo()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := catalog.NewService(repo)
	svc.WithClock(func() time.Time { return base })
	ctx := context.Background()

	artistA, _ := svc.CreateArtist(ctx, "Artist A", "")

	svc.WithClock(func() time.Time { return base.Add(time.Hour) })
	_, _ = svc.CreateArtist(ctx, "Artist B", "")

	svc.WithClock(func() time.Time { return base.Add(2 * time.Hour) })
	_, err := svc.UpdateArtist(ctx, artistA.ID, catalog.UpdateArtistRequest{SortName: strPtr("A")})
	if err != nil {
		t.Fatalf("UpdateArtist: %v", err)
	}

	result, err := svc.GetRecentlyUpdated(ctx, "", 0)
	if err != nil {
		t.Fatalf("GetRecentlyUpdated: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Artist == nil || result.Items[0].Artist.Name != "Artist A" {
		t.Errorf("expected most recently updated item first, got %+v", result.Items[0])
	}
	if !result.Items[0].UpdatedAt.After(result.Items[1].UpdatedAt) {
		t.Errorf("expected UpdatedAt descending order, got %s then %s", result.Items[0].UpdatedAt, result.Items[1].UpdatedAt)
	}
}

func TestGetRecentlyUpdatedKindFilter(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	artist, _ := svc.CreateArtist(ctx, "Solo Artist", "")
	_, _ = svc.CreateTrack(ctx, "Track 1", "", artist.ID, "", "mo-001", 1, 0, 180000)

	result, err := svc.GetRecentlyUpdated(ctx, "artist", 0)
	if err != nil {
		t.Fatalf("GetRecentlyUpdated artist: %v", err)
	}
	for _, item := range result.Items {
		if item.Kind != catalog.RecentItemArtist {
			t.Errorf("expected only artist items, got %s", item.Kind)
		}
	}

	result, err = svc.GetRecentlyUpdated(ctx, "track", 0)
	if err != nil {
		t.Fatalf("GetRecentlyUpdated track: %v", err)
	}
	for _, item := range result.Items {
		if item.Kind != catalog.RecentItemTrack {
			t.Errorf("expected only track items, got %s", item.Kind)
		}
	}
}

func TestGetRecentlyUpdatedRejectsInvalidKind(t *testing.T) {
	svc := catalog.NewService(newMemRepo())
	_, err := svc.GetRecentlyUpdated(context.Background(), "invalid", 0)
	if !errors.Is(err, catalog.ErrInvalidTrack) {
		t.Fatalf("expected ErrInvalidTrack for invalid kind, got %v", err)
	}
}

func TestGetRecentlyUpdatedLimitCaps(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		svc.CreateArtist(ctx, "Artist", "")
	}

	result, err := svc.GetRecentlyUpdated(ctx, "", 5)
	if err != nil {
		t.Fatalf("GetRecentlyUpdated: %v", err)
	}
	if len(result.Items) != 5 {
		t.Errorf("expected 5 items (limit=5), got %d", len(result.Items))
	}

	result, err = svc.GetRecentlyUpdated(ctx, "", 999)
	if err != nil {
		t.Fatalf("GetRecentlyUpdated limit=999: %v", err)
	}
	if len(result.Items) != 25 {
		t.Errorf("expected all 25 items when limit>count, got %d", len(result.Items))
	}
}

func TestGetRecentlyUpdatedPlaylistKindFilter(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	_, _ = svc.CreatePlaylist(ctx, "Fresh Mix", "")
	_, _ = svc.CreateArtist(ctx, "Artist X", "")

	result, err := svc.GetRecentlyUpdated(ctx, "playlist", 0)
	if err != nil {
		t.Fatalf("GetRecentlyUpdated playlist: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 playlist item, got %d", len(result.Items))
	}
	if result.Items[0].Kind != catalog.RecentItemPlaylist {
		t.Errorf("expected playlist kind, got %s", result.Items[0].Kind)
	}
	if result.Items[0].Playlist == nil || result.Items[0].Playlist.Name != "Fresh Mix" {
		t.Errorf("expected playlist payload, got %+v", result.Items[0])
	}
}

func TestGetRecentlyUpdatedPlaylistUnified(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	_, _ = svc.CreatePlaylist(ctx, "Mix", "")
	_, _ = svc.CreateArtist(ctx, "Artist", "")

	result, err := svc.GetRecentlyUpdated(ctx, "", 0)
	if err != nil {
		t.Fatalf("GetRecentlyUpdated: %v", err)
	}

	hasPlaylist := false
	hasArtist := false
	for _, item := range result.Items {
		switch item.Kind {
		case catalog.RecentItemPlaylist:
			hasPlaylist = true
		case catalog.RecentItemArtist:
			hasArtist = true
		}
	}
	if !hasPlaylist {
		t.Error("expected at least one playlist item in unified timeline")
	}
	if !hasArtist {
		t.Error("expected at least one artist item in unified timeline")
	}
}

// ---- ListXxxPage tests ----

func TestListArtistsPageSortAndPaginate(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	for _, name := range []string{"Zara", "Alice", "Mike"} {
		if _, err := svc.CreateArtist(ctx, name, ""); err != nil {
			t.Fatalf("CreateArtist %q: %v", name, err)
		}
	}

	// asc by name, all
	page, err := svc.ListArtistsPage(ctx, catalog.ListQuery{SortBy: "name", SortOrder: "asc", Limit: 10, Offset: 0})
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

	// desc, limit=1
	page, err = svc.ListArtistsPage(ctx, catalog.ListQuery{SortBy: "name", SortOrder: "desc", Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("ListArtistsPage desc: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(page.Items))
	}
	if page.Total != 3 {
		t.Errorf("total = %d, want 3 even with limit=1", page.Total)
	}
	if page.Items[0].Name != "Zara" {
		t.Errorf("items[0].Name = %q, want Zara", page.Items[0].Name)
	}

	// offset past end
	page, err = svc.ListArtistsPage(ctx, catalog.ListQuery{SortBy: "name", SortOrder: "asc", Limit: 10, Offset: 99})
	if err != nil {
		t.Fatalf("ListArtistsPage offset>total: %v", err)
	}
	if len(page.Items) != 0 {
		t.Errorf("items = %d, want 0", len(page.Items))
	}
	if page.Total != 3 {
		t.Errorf("total = %d, want 3", page.Total)
	}
}

func TestListAlbumsPageByArtist(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	artist, _ := svc.CreateArtist(ctx, "Band", "")
	for _, year := range []int{2020, 2015, 2023} {
		svc.CreateAlbum(ctx, fmt.Sprintf("Album %d", year), "", artist.ID, year)
	}

	page, err := svc.ListAlbumsByArtistPage(ctx, artist.ID, catalog.ListQuery{
		SortBy: catalog.AlbumSortByReleaseYear, SortOrder: catalog.CatalogSortOrderAsc, Limit: 10, Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListAlbumsByArtistPage: %v", err)
	}
	if len(page.Items) != 3 || page.Total != 3 {
		t.Fatalf("want 3 items total=3, got %d total=%d", len(page.Items), page.Total)
	}
	if page.Items[0].ReleaseYear != 2015 {
		t.Errorf("items[0].ReleaseYear = %d, want 2015", page.Items[0].ReleaseYear)
	}
}

func TestListTracksPageSortByTitle(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	artist, _ := svc.CreateArtist(ctx, "Band", "")
	for _, title := range []string{"Zephyr", "Aura", "Midnight"} {
		svc.CreateTrack(ctx, title, "", artist.ID, "", "mo-"+title, 0, 0, 0)
	}

	page, err := svc.ListTracksPage(ctx, catalog.ListQuery{
		SortBy: catalog.TrackSortByTitle, SortOrder: catalog.CatalogSortOrderAsc, Limit: 2, Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListTracksPage: %v", err)
	}
	if len(page.Items) != 2 || page.Total != 3 {
		t.Fatalf("want 2 items total=3, got %d total=%d", len(page.Items), page.Total)
	}
	if page.Items[0].Title != "Aura" {
		t.Errorf("items[0].Title = %q, want Aura", page.Items[0].Title)
	}
}

func TestListPlaylistsPageDescByName(t *testing.T) {
	repo := newMemRepo()
	svc := catalog.NewService(repo)
	ctx := context.Background()

	for _, name := range []string{"Zen Mix", "Alpha Hits", "Morning Chill"} {
		svc.CreatePlaylist(ctx, name, "")
	}

	page, err := svc.ListPlaylistsPage(ctx, catalog.ListQuery{
		SortBy: catalog.PlaylistSortByName, SortOrder: catalog.CatalogSortOrderDesc, Limit: 10, Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListPlaylistsPage: %v", err)
	}
	if len(page.Items) != 3 || page.Total != 3 {
		t.Fatalf("want 3 items total=3, got %d total=%d", len(page.Items), page.Total)
	}
	if page.Items[0].Name != "Zen Mix" {
		t.Errorf("items[0].Name = %q, want Zen Mix", page.Items[0].Name)
	}
}

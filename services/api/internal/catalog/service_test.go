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

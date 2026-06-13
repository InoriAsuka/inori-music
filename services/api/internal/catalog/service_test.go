package catalog_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"inori-music/services/api/internal/catalog"
)

type memRepo struct {
	mu      sync.RWMutex
	artists map[string]catalog.Artist
	albums  map[string]catalog.Album
	tracks  map[string]catalog.Track
}

func newMemRepo() *memRepo {
	return &memRepo{artists: map[string]catalog.Artist{}, albums: map[string]catalog.Album{}, tracks: map[string]catalog.Track{}}
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

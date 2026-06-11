package catalog_test

import (
	"context"
	"errors"
	"fmt"
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

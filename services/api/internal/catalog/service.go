package catalog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Service coordinates music catalog metadata validation and persistence.
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) CreateArtist(ctx context.Context, name, sortName string) (Artist, error) {
	name = strings.TrimSpace(name)
	sortName = strings.TrimSpace(sortName)
	if name == "" {
		return Artist{}, fmt.Errorf("%w: name is required", ErrInvalidArtist)
	}
	id, err := newID()
	if err != nil {
		return Artist{}, fmt.Errorf("generate artist id: %w", err)
	}
	now := s.now().UTC()
	artist := Artist{ID: id, Name: name, SortName: sortName, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.SaveArtist(ctx, artist); err != nil {
		return Artist{}, err
	}
	return artist, nil
}

func (s *Service) ListArtists(ctx context.Context) ([]Artist, error) {
	return s.repo.ListArtists(ctx)
}

func (s *Service) GetArtist(ctx context.Context, id string) (Artist, error) {
	return s.repo.GetArtist(ctx, strings.TrimSpace(id))
}

func (s *Service) DeleteArtist(ctx context.Context, id string) error {
	return s.repo.DeleteArtist(ctx, strings.TrimSpace(id))
}

func (s *Service) CreateAlbum(ctx context.Context, title, sortTitle, artistID string, releaseYear int) (Album, error) {
	title = strings.TrimSpace(title)
	sortTitle = strings.TrimSpace(sortTitle)
	artistID = strings.TrimSpace(artistID)
	if title == "" {
		return Album{}, fmt.Errorf("%w: title is required", ErrInvalidAlbum)
	}
	if artistID == "" {
		return Album{}, fmt.Errorf("%w: artist_id is required", ErrInvalidAlbum)
	}
	if releaseYear < 0 {
		return Album{}, fmt.Errorf("%w: release_year must be non-negative", ErrInvalidAlbum)
	}
	if _, err := s.repo.GetArtist(ctx, artistID); err != nil {
		return Album{}, err
	}
	id, err := newID()
	if err != nil {
		return Album{}, fmt.Errorf("generate album id: %w", err)
	}
	now := s.now().UTC()
	album := Album{ID: id, Title: title, SortTitle: sortTitle, ArtistID: artistID, ReleaseYear: releaseYear, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.SaveAlbum(ctx, album); err != nil {
		return Album{}, err
	}
	return album, nil
}

func (s *Service) ListAlbums(ctx context.Context) ([]Album, error) {
	return s.repo.ListAlbums(ctx)
}

func (s *Service) ListAlbumsByArtist(ctx context.Context, artistID string) ([]Album, error) {
	return s.repo.ListAlbumsByArtist(ctx, strings.TrimSpace(artistID))
}

func (s *Service) GetAlbum(ctx context.Context, id string) (Album, error) {
	return s.repo.GetAlbum(ctx, strings.TrimSpace(id))
}

func (s *Service) DeleteAlbum(ctx context.Context, id string) error {
	return s.repo.DeleteAlbum(ctx, strings.TrimSpace(id))
}

func (s *Service) CreateTrack(ctx context.Context, title, sortTitle, artistID, albumID, mediaObjectID string, trackNumber, discNumber, durationMS int) (Track, error) {
	title = strings.TrimSpace(title)
	sortTitle = strings.TrimSpace(sortTitle)
	artistID = strings.TrimSpace(artistID)
	albumID = strings.TrimSpace(albumID)
	mediaObjectID = strings.TrimSpace(mediaObjectID)
	if title == "" {
		return Track{}, fmt.Errorf("%w: title is required", ErrInvalidTrack)
	}
	if artistID == "" {
		return Track{}, fmt.Errorf("%w: artist_id is required", ErrInvalidTrack)
	}
	if mediaObjectID == "" {
		return Track{}, fmt.Errorf("%w: media_object_id is required", ErrInvalidTrack)
	}
	if trackNumber < 0 || discNumber < 0 || durationMS < 0 {
		return Track{}, fmt.Errorf("%w: numeric fields must be non-negative", ErrInvalidTrack)
	}
	if _, err := s.repo.GetArtist(ctx, artistID); err != nil {
		return Track{}, err
	}
	if albumID != "" {
		album, err := s.repo.GetAlbum(ctx, albumID)
		if err != nil {
			return Track{}, err
		}
		if album.ArtistID != artistID {
			return Track{}, fmt.Errorf("%w: album artist mismatch", ErrInvalidTrack)
		}
	}
	id, err := newID()
	if err != nil {
		return Track{}, fmt.Errorf("generate track id: %w", err)
	}
	now := s.now().UTC()
	track := Track{ID: id, Title: title, SortTitle: sortTitle, ArtistID: artistID, AlbumID: albumID, MediaObjectID: mediaObjectID, TrackNumber: trackNumber, DiscNumber: discNumber, DurationMS: durationMS, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.SaveTrack(ctx, track); err != nil {
		return Track{}, err
	}
	return track, nil
}

func (s *Service) ListTracks(ctx context.Context) ([]Track, error) {
	return s.repo.ListTracks(ctx)
}

func (s *Service) ListTracksByAlbum(ctx context.Context, albumID string) ([]Track, error) {
	return s.repo.ListTracksByAlbum(ctx, strings.TrimSpace(albumID))
}

func (s *Service) ListTracksByArtist(ctx context.Context, artistID string) ([]Track, error) {
	return s.repo.ListTracksByArtist(ctx, strings.TrimSpace(artistID))
}

func (s *Service) GetTrack(ctx context.Context, id string) (Track, error) {
	return s.repo.GetTrack(ctx, strings.TrimSpace(id))
}

func (s *Service) DeleteTrack(ctx context.Context, id string) error {
	return s.repo.DeleteTrack(ctx, strings.TrimSpace(id))
}

func newID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

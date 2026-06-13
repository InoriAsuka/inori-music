package catalog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Service coordinates music catalog metadata validation and persistence.
type Service struct {
	repo        Repository
	mediaReader MediaObjectReader // optional; required only for ImportTrack
	now         func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// WithClock replaces the time source used by the service. Intended for tests
// that need deterministic or advancing timestamps.
func (s *Service) WithClock(fn func() time.Time) *Service {
	s.now = fn
	return s
}

// WithMediaObjectReader attaches a media object reader to the service so that
// ImportTrack can validate media objects before creating track records.
func (s *Service) WithMediaObjectReader(r MediaObjectReader) *Service {
	s.mediaReader = r
	return s
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

// UpdateArtist applies a partial update to an existing artist.
// Only non-nil fields in req are applied. Name may not be set to an empty string.
func (s *Service) UpdateArtist(ctx context.Context, id string, req UpdateArtistRequest) (Artist, error) {
	id = strings.TrimSpace(id)
	artist, err := s.repo.GetArtist(ctx, id)
	if err != nil {
		return Artist{}, err
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return Artist{}, fmt.Errorf("%w: name must not be empty", ErrInvalidArtist)
		}
		artist.Name = name
	}
	if req.SortName != nil {
		artist.SortName = strings.TrimSpace(*req.SortName)
	}
	artist.UpdatedAt = s.now().UTC()
	if err := s.repo.SaveArtist(ctx, artist); err != nil {
		return Artist{}, err
	}
	return artist, nil
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

// UpdateAlbum applies a partial update to an existing album.
// Only non-nil fields in req are applied. Title and ArtistID may not be set to empty strings.
// When ArtistID changes, the referenced artist must exist.
func (s *Service) UpdateAlbum(ctx context.Context, id string, req UpdateAlbumRequest) (Album, error) {
	id = strings.TrimSpace(id)
	album, err := s.repo.GetAlbum(ctx, id)
	if err != nil {
		return Album{}, err
	}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return Album{}, fmt.Errorf("%w: title must not be empty", ErrInvalidAlbum)
		}
		album.Title = title
	}
	if req.SortTitle != nil {
		album.SortTitle = strings.TrimSpace(*req.SortTitle)
	}
	if req.ArtistID != nil {
		artistID := strings.TrimSpace(*req.ArtistID)
		if artistID == "" {
			return Album{}, fmt.Errorf("%w: artist_id must not be empty", ErrInvalidAlbum)
		}
		if _, err := s.repo.GetArtist(ctx, artistID); err != nil {
			return Album{}, err
		}
		album.ArtistID = artistID
	}
	if req.ReleaseYear != nil {
		if *req.ReleaseYear < 0 {
			return Album{}, fmt.Errorf("%w: release_year must be non-negative", ErrInvalidAlbum)
		}
		album.ReleaseYear = *req.ReleaseYear
	}
	album.UpdatedAt = s.now().UTC()
	if err := s.repo.SaveAlbum(ctx, album); err != nil {
		return Album{}, err
	}
	return album, nil
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

// UpdateTrack applies a partial update to an existing track.
// Only non-nil fields in req are applied. Title and ArtistID may not be set to empty strings.
// When ArtistID or AlbumID changes, the referenced entities must exist; artist ownership of
// the album is enforced using the final ArtistID (after applying any ArtistID change).
func (s *Service) UpdateTrack(ctx context.Context, id string, req UpdateTrackRequest) (Track, error) {
	id = strings.TrimSpace(id)
	track, err := s.repo.GetTrack(ctx, id)
	if err != nil {
		return Track{}, err
	}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return Track{}, fmt.Errorf("%w: title must not be empty", ErrInvalidTrack)
		}
		track.Title = title
	}
	if req.SortTitle != nil {
		track.SortTitle = strings.TrimSpace(*req.SortTitle)
	}
	if req.ArtistID != nil {
		artistID := strings.TrimSpace(*req.ArtistID)
		if artistID == "" {
			return Track{}, fmt.Errorf("%w: artist_id must not be empty", ErrInvalidTrack)
		}
		if _, err := s.repo.GetArtist(ctx, artistID); err != nil {
			return Track{}, err
		}
		track.ArtistID = artistID
	}
	if req.AlbumID != nil {
		albumID := strings.TrimSpace(*req.AlbumID)
		if albumID != "" {
			album, err := s.repo.GetAlbum(ctx, albumID)
			if err != nil {
				return Track{}, err
			}
			if album.ArtistID != track.ArtistID {
				return Track{}, fmt.Errorf("%w: album artist mismatch", ErrInvalidTrack)
			}
		}
		track.AlbumID = albumID
	}
	if req.TrackNumber != nil {
		if *req.TrackNumber < 0 {
			return Track{}, fmt.Errorf("%w: track_number must be non-negative", ErrInvalidTrack)
		}
		track.TrackNumber = *req.TrackNumber
	}
	if req.DiscNumber != nil {
		if *req.DiscNumber < 0 {
			return Track{}, fmt.Errorf("%w: disc_number must be non-negative", ErrInvalidTrack)
		}
		track.DiscNumber = *req.DiscNumber
	}
	if req.DurationMS != nil {
		if *req.DurationMS < 0 {
			return Track{}, fmt.Errorf("%w: duration_ms must be non-negative", ErrInvalidTrack)
		}
		track.DurationMS = *req.DurationMS
	}
	track.UpdatedAt = s.now().UTC()
	if err := s.repo.SaveTrack(ctx, track); err != nil {
		return Track{}, err
	}
	return track, nil
}

// ImportTrack validates a media object and creates a Track record that references it.
// The media object must exist, be of kind original_audio or transcoded_audio, and
// have lifecycle state active. A title is derived from the request or falls back to
// the media object ID when none is supplied.
func (s *Service) ImportTrack(ctx context.Context, req ImportTrackRequest) (Track, error) {
	if s.mediaReader == nil {
		return Track{}, fmt.Errorf("%w: media object reader not configured", ErrImportRejected)
	}
	req.MediaObjectID = strings.TrimSpace(req.MediaObjectID)
	if req.MediaObjectID == "" {
		return Track{}, fmt.Errorf("%w: media_object_id is required", ErrImportRejected)
	}
	info, err := s.mediaReader.GetMediaObjectInfo(ctx, req.MediaObjectID)
	if err != nil {
		return Track{}, err
	}
	if info.AssetKind != "original_audio" && info.AssetKind != "transcoded_audio" {
		return Track{}, fmt.Errorf("%w: media object asset_kind must be original_audio or transcoded_audio, got %q", ErrImportRejected, info.AssetKind)
	}
	if info.LifecycleState != "active" {
		return Track{}, fmt.Errorf("%w: media object lifecycle_state must be active, got %q", ErrImportRejected, info.LifecycleState)
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = req.MediaObjectID
	}
	sortTitle := strings.TrimSpace(req.SortTitle)
	artistID := strings.TrimSpace(req.ArtistID)
	if artistID != "" {
		if _, err := s.repo.GetArtist(ctx, artistID); err != nil {
			return Track{}, err
		}
	}
	albumID := strings.TrimSpace(req.AlbumID)
	if albumID != "" {
		album, err := s.repo.GetAlbum(ctx, albumID)
		if err != nil {
			return Track{}, err
		}
		if artistID != "" && album.ArtistID != artistID {
			return Track{}, fmt.Errorf("%w: album artist mismatch", ErrInvalidTrack)
		}
		if artistID == "" {
			artistID = album.ArtistID
		}
	}
	if req.TrackNumber < 0 || req.DiscNumber < 0 || req.DurationMS < 0 {
		return Track{}, fmt.Errorf("%w: numeric fields must be non-negative", ErrInvalidTrack)
	}
	id, err := newID()
	if err != nil {
		return Track{}, fmt.Errorf("generate track id: %w", err)
	}
	now := s.now().UTC()
	track := Track{
		ID:            id,
		Title:         title,
		SortTitle:     sortTitle,
		ArtistID:      artistID,
		AlbumID:       albumID,
		MediaObjectID: req.MediaObjectID,
		TrackNumber:   req.TrackNumber,
		DiscNumber:    req.DiscNumber,
		DurationMS:    req.DurationMS,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.SaveTrack(ctx, track); err != nil {
		return Track{}, err
	}
	return track, nil
}

// BatchImportTracks imports each item in req independently. Failures do not abort
// subsequent items. The returned BatchImportResult records both successes and
// per-item error details so callers can present partial results to the user.
func (s *Service) BatchImportTracks(ctx context.Context, items []ImportTrackRequest) BatchImportResult {
	result := BatchImportResult{
		Total: len(items),
		Items: make([]BatchImportResultItem, 0, len(items)),
	}
	for i, req := range items {
		mediaID := strings.TrimSpace(req.MediaObjectID)
		track, err := s.ImportTrack(ctx, req)
		if err != nil {
			code := "import_rejected"
			switch {
			case errors.Is(err, ErrArtistNotFound):
				code = "artist_not_found"
			case errors.Is(err, ErrAlbumNotFound):
				code = "album_not_found"
			case errors.Is(err, ErrInvalidTrack):
				code = "invalid_catalog_entity"
			}
			result.Failed++
			result.Items = append(result.Items, BatchImportResultItem{
				Index:         i,
				MediaObjectID: mediaID,
				Error:         err.Error(),
				ErrorCode:     code,
			})
			continue
		}
		result.Imported++
		t := track
		result.Items = append(result.Items, BatchImportResultItem{
			Index:         i,
			MediaObjectID: mediaID,
			Track:         &t,
		})
	}
	return result
}

// SearchCatalog delegates a full-text catalog search to the underlying repository.
// An empty query string is rejected with a validation error.
func (s *Service) SearchCatalog(ctx context.Context, query string) (CatalogSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return CatalogSearchResult{}, fmt.Errorf("%w: query must not be empty", ErrInvalidTrack)
	}
	return s.repo.SearchCatalog(ctx, query)
}

func newID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

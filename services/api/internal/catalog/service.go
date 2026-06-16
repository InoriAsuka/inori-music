package catalog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
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

// RelinkTrack replaces the media object reference on an existing track.
// The new media object must exist, be of kind original_audio or transcoded_audio,
// and have lifecycle state active.
func (s *Service) RelinkTrack(ctx context.Context, id string, req RelinkTrackRequest) (Track, error) {
	if s.mediaReader == nil {
		return Track{}, fmt.Errorf("%w: media object reader not configured", ErrRelinkRejected)
	}
	id = strings.TrimSpace(id)
	track, err := s.repo.GetTrack(ctx, id)
	if err != nil {
		return Track{}, err
	}
	req.MediaObjectID = strings.TrimSpace(req.MediaObjectID)
	if req.MediaObjectID == "" {
		return Track{}, fmt.Errorf("%w: media_object_id is required", ErrRelinkRejected)
	}
	info, err := s.mediaReader.GetMediaObjectInfo(ctx, req.MediaObjectID)
	if err != nil {
		return Track{}, err
	}
	if info.AssetKind != "original_audio" && info.AssetKind != "transcoded_audio" {
		return Track{}, fmt.Errorf("%w: media object asset_kind must be original_audio or transcoded_audio, got %q", ErrRelinkRejected, info.AssetKind)
	}
	if info.LifecycleState != "active" {
		return Track{}, fmt.Errorf("%w: media object lifecycle_state must be active, got %q", ErrRelinkRejected, info.LifecycleState)
	}
	track.MediaObjectID = req.MediaObjectID
	track.UpdatedAt = s.now().UTC()
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

// CreatePlaylist creates a new empty playlist with the given name and optional description.
func (s *Service) CreatePlaylist(ctx context.Context, name, description string) (Playlist, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Playlist{}, fmt.Errorf("%w: name is required", ErrInvalidPlaylist)
	}
	id, err := newID()
	if err != nil {
		return Playlist{}, fmt.Errorf("generate playlist id: %w", err)
	}
	now := s.now().UTC()
	p := Playlist{
		ID:          id,
		Name:        name,
		Description: strings.TrimSpace(description),
		TrackIDs:    []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.SavePlaylist(ctx, p); err != nil {
		return Playlist{}, err
	}
	return p, nil
}

// ListPlaylists returns all playlists in the repository.
func (s *Service) ListPlaylists(ctx context.Context) ([]Playlist, error) {
	return s.repo.ListPlaylists(ctx)
}

// GetPlaylist returns a single playlist by ID.
func (s *Service) GetPlaylist(ctx context.Context, id string) (Playlist, error) {
	return s.repo.GetPlaylist(ctx, strings.TrimSpace(id))
}

// DeletePlaylist removes a playlist by ID.
func (s *Service) DeletePlaylist(ctx context.Context, id string) error {
	return s.repo.DeletePlaylist(ctx, strings.TrimSpace(id))
}

// UpdatePlaylist applies a partial metadata update to an existing playlist.
// Only non-nil fields in req are applied. Name may not be set to an empty string.
func (s *Service) UpdatePlaylist(ctx context.Context, id string, req UpdatePlaylistRequest) (Playlist, error) {
	id = strings.TrimSpace(id)
	p, err := s.repo.GetPlaylist(ctx, id)
	if err != nil {
		return Playlist{}, err
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return Playlist{}, fmt.Errorf("%w: name must not be empty", ErrInvalidPlaylist)
		}
		p.Name = name
	}
	if req.Description != nil {
		p.Description = strings.TrimSpace(*req.Description)
	}
	p.UpdatedAt = s.now().UTC()
	if err := s.repo.SavePlaylist(ctx, p); err != nil {
		return Playlist{}, err
	}
	return p, nil
}

// AddTrackToPlaylist appends a track to the end of the playlist's ordered track list.
// The track must exist. Duplicate entries are permitted (same track may appear multiple times).
func (s *Service) AddTrackToPlaylist(ctx context.Context, playlistID, trackID string) (Playlist, error) {
	playlistID = strings.TrimSpace(playlistID)
	trackID = strings.TrimSpace(trackID)
	if trackID == "" {
		return Playlist{}, fmt.Errorf("%w: track_id is required", ErrInvalidPlaylist)
	}
	if _, err := s.repo.GetTrack(ctx, trackID); err != nil {
		return Playlist{}, err
	}
	p, err := s.repo.GetPlaylist(ctx, playlistID)
	if err != nil {
		return Playlist{}, err
	}
	p.TrackIDs = append(p.TrackIDs, trackID)
	p.UpdatedAt = s.now().UTC()
	if err := s.repo.SavePlaylist(ctx, p); err != nil {
		return Playlist{}, err
	}
	return p, nil
}

// RemoveTrackFromPlaylist removes the first occurrence of trackID from the playlist.
// Returns ErrInvalidPlaylist when trackID is not found in the list.
func (s *Service) RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) (Playlist, error) {
	playlistID = strings.TrimSpace(playlistID)
	trackID = strings.TrimSpace(trackID)
	if trackID == "" {
		return Playlist{}, fmt.Errorf("%w: track_id is required", ErrInvalidPlaylist)
	}
	p, err := s.repo.GetPlaylist(ctx, playlistID)
	if err != nil {
		return Playlist{}, err
	}
	idx := -1
	for i, tid := range p.TrackIDs {
		if tid == trackID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return Playlist{}, fmt.Errorf("%w: track %s not found in playlist", ErrInvalidPlaylist, trackID)
	}
	p.TrackIDs = append(p.TrackIDs[:idx], p.TrackIDs[idx+1:]...)
	p.UpdatedAt = s.now().UTC()
	if err := s.repo.SavePlaylist(ctx, p); err != nil {
		return Playlist{}, err
	}
	return p, nil
}

// SetPlaylistTracks atomically replaces the ordered track list of a playlist.
// Every trackID in the supplied slice must exist; an unknown ID returns ErrTrackNotFound.
// An empty slice is valid and clears the track list.
// GetPlaylistTracks returns the full Track objects for a playlist in the
// playlist's defined order. If the playlist has no tracks an empty slice is
// returned. An unknown playlist ID returns ErrPlaylistNotFound.
func (s *Service) GetPlaylistTracks(ctx context.Context, playlistID string) ([]Track, error) {
	playlistID = strings.TrimSpace(playlistID)
	p, err := s.repo.GetPlaylist(ctx, playlistID)
	if err != nil {
		return nil, err
	}
	tracks := make([]Track, 0, len(p.TrackIDs))
	for _, tid := range p.TrackIDs {
		t, err := s.repo.GetTrack(ctx, tid)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	return tracks, nil
}

func (s *Service) SetPlaylistTracks(ctx context.Context, playlistID string, trackIDs []string) (Playlist, error) {
	playlistID = strings.TrimSpace(playlistID)
	p, err := s.repo.GetPlaylist(ctx, playlistID)
	if err != nil {
		return Playlist{}, err
	}
	for _, tid := range trackIDs {
		if _, err := s.repo.GetTrack(ctx, tid); err != nil {
			return Playlist{}, err
		}
	}
	p.TrackIDs = trackIDs
	p.UpdatedAt = s.now().UTC()
	if err := s.repo.SavePlaylist(ctx, p); err != nil {
		return Playlist{}, err
	}
	return p, nil
}

// GetCatalogStats returns metadata-only aggregate entity counts.
func (s *Service) GetCatalogStats(ctx context.Context) (CatalogStats, error) {
	artists, err := s.repo.ListArtists(ctx)
	if err != nil {
		return CatalogStats{}, err
	}
	albums, err := s.repo.ListAlbums(ctx)
	if err != nil {
		return CatalogStats{}, err
	}
	tracks, err := s.repo.ListTracks(ctx)
	if err != nil {
		return CatalogStats{}, err
	}
	playlists, err := s.repo.ListPlaylists(ctx)
	if err != nil {
		return CatalogStats{}, err
	}
	return CatalogStats{
		Artists:   len(artists),
		Albums:    len(albums),
		Tracks:    len(tracks),
		Playlists: len(playlists),
	}, nil
}

// GetArtistStatsBreakdown returns per-artist album and track counts.
// Counts are derived from existing list calls; no new Repository interface methods are required.
func (s *Service) GetArtistStatsBreakdown(ctx context.Context) (ArtistStatsBreakdown, error) {
	artists, err := s.repo.ListArtists(ctx)
	if err != nil {
		return ArtistStatsBreakdown{}, err
	}
	items := make([]ArtistStatItem, 0, len(artists))
	for _, a := range artists {
		albums, err := s.repo.ListAlbumsByArtist(ctx, a.ID)
		if err != nil {
			return ArtistStatsBreakdown{}, err
		}
		tracks, err := s.repo.ListTracksByArtist(ctx, a.ID)
		if err != nil {
			return ArtistStatsBreakdown{}, err
		}
		items = append(items, ArtistStatItem{
			ArtistID:   a.ID,
			Name:       a.Name,
			AlbumCount: len(albums),
			TrackCount: len(tracks),
		})
	}
	return ArtistStatsBreakdown{Artists: items}, nil
}

// GetAlbumStatsBreakdown returns per-album track counts.
// Counts are derived from existing list calls; no new Repository interface methods are required.
func (s *Service) GetAlbumStatsBreakdown(ctx context.Context) (AlbumStatsBreakdown, error) {
	albums, err := s.repo.ListAlbums(ctx)
	if err != nil {
		return AlbumStatsBreakdown{}, err
	}
	items := make([]AlbumStatItem, 0, len(albums))
	for _, al := range albums {
		tracks, err := s.repo.ListTracksByAlbum(ctx, al.ID)
		if err != nil {
			return AlbumStatsBreakdown{}, err
		}
		items = append(items, AlbumStatItem{
			AlbumID:    al.ID,
			Title:      al.Title,
			ArtistID:   al.ArtistID,
			TrackCount: len(tracks),
		})
	}
	return AlbumStatsBreakdown{Albums: items}, nil
}

// GetPlaylistStatsBreakdown returns per-playlist track counts.
// Counts are derived from the playlist's TrackIDs length; no new Repository interface methods are required.
func (s *Service) GetPlaylistStatsBreakdown(ctx context.Context) (PlaylistStatsBreakdown, error) {
	playlists, err := s.repo.ListPlaylists(ctx)
	if err != nil {
		return PlaylistStatsBreakdown{}, err
	}
	items := make([]PlaylistStatItem, 0, len(playlists))
	for _, p := range playlists {
		items = append(items, PlaylistStatItem{
			PlaylistID: p.ID,
			Name:       p.Name,
			TrackCount: len(p.TrackIDs),
		})
	}
	return PlaylistStatsBreakdown{Playlists: items}, nil
}

// GetRecentlyAdded returns the most recently created catalog items across artists, albums,
// and tracks in a unified newest-first timeline. limit caps the result set (1–100; default 20).
// kind filters to a single entity type ("artist", "album", or "track"); an empty string
// returns all kinds.
func (s *Service) GetRecentlyAdded(ctx context.Context, kind string, limit int) (RecentCatalogResult, error) {
	kind = strings.TrimSpace(kind)
	if err := validateRecentItemKind(kind); err != nil {
		return RecentCatalogResult{}, err
	}
	limit = normalizeRecentLimit(limit)

	items := make([]RecentCatalogItem, 0)

	if kind == "" || kind == string(RecentItemArtist) {
		artists, err := s.repo.ListArtists(ctx)
		if err != nil {
			return RecentCatalogResult{}, err
		}
		for i := range artists {
			a := artists[i]
			items = append(items, RecentCatalogItem{
				Kind:    RecentItemArtist,
				Artist:  &a,
				AddedAt: a.CreatedAt,
			})
		}
	}

	if kind == "" || kind == string(RecentItemAlbum) {
		albums, err := s.repo.ListAlbums(ctx)
		if err != nil {
			return RecentCatalogResult{}, err
		}
		for i := range albums {
			al := albums[i]
			items = append(items, RecentCatalogItem{
				Kind:    RecentItemAlbum,
				Album:   &al,
				AddedAt: al.CreatedAt,
			})
		}
	}

	if kind == "" || kind == string(RecentItemTrack) {
		tracks, err := s.repo.ListTracks(ctx)
		if err != nil {
			return RecentCatalogResult{}, err
		}
		for i := range tracks {
			t := tracks[i]
			items = append(items, RecentCatalogItem{
				Kind:    RecentItemTrack,
				Track:   &t,
				AddedAt: t.CreatedAt,
			})
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].AddedAt.After(items[j].AddedAt)
	})

	if limit < len(items) {
		items = items[:limit]
	}

	return RecentCatalogResult{Items: items}, nil
}

// GetRecentlyUpdated returns the most recently updated catalog items across artists, albums,
// and tracks in a unified newest-first timeline. limit caps the result set (1–100; default 20).
// kind filters to a single entity type ("artist", "album", or "track"); an empty string
// returns all kinds.
func (s *Service) GetRecentlyUpdated(ctx context.Context, kind string, limit int) (UpdatedCatalogResult, error) {
	kind = strings.TrimSpace(kind)
	if err := validateRecentItemKind(kind); err != nil {
		return UpdatedCatalogResult{}, err
	}
	limit = normalizeRecentLimit(limit)

	items := make([]UpdatedCatalogItem, 0)

	if kind == "" || kind == string(RecentItemArtist) {
		artists, err := s.repo.ListArtists(ctx)
		if err != nil {
			return UpdatedCatalogResult{}, err
		}
		for i := range artists {
			a := artists[i]
			items = append(items, UpdatedCatalogItem{
				Kind:      RecentItemArtist,
				Artist:    &a,
				UpdatedAt: a.UpdatedAt,
			})
		}
	}

	if kind == "" || kind == string(RecentItemAlbum) {
		albums, err := s.repo.ListAlbums(ctx)
		if err != nil {
			return UpdatedCatalogResult{}, err
		}
		for i := range albums {
			al := albums[i]
			items = append(items, UpdatedCatalogItem{
				Kind:      RecentItemAlbum,
				Album:     &al,
				UpdatedAt: al.UpdatedAt,
			})
		}
	}

	if kind == "" || kind == string(RecentItemTrack) {
		tracks, err := s.repo.ListTracks(ctx)
		if err != nil {
			return UpdatedCatalogResult{}, err
		}
		for i := range tracks {
			t := tracks[i]
			items = append(items, UpdatedCatalogItem{
				Kind:      RecentItemTrack,
				Track:     &t,
				UpdatedAt: t.UpdatedAt,
			})
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})

	if limit < len(items) {
		items = items[:limit]
	}

	return UpdatedCatalogResult{Items: items}, nil
}

func validateRecentItemKind(kind string) error {
	switch kind {
	case "", string(RecentItemArtist), string(RecentItemAlbum), string(RecentItemTrack):
		return nil
	default:
		return fmt.Errorf("%w: kind must be artist, album, or track", ErrInvalidTrack)
	}
}

func normalizeRecentLimit(limit int) int {
	const defaultLimit = 20
	const maxLimit = 100
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func newID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

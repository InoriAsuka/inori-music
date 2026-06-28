package userplaylist

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

// Service coordinates user playlist business rules.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService creates a Service backed by the given repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// WithClock replaces the time source. Intended for tests.
func (s *Service) WithClock(fn func() time.Time) *Service {
	s.now = fn
	return s
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("userplaylist: rand.Read failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// CreatePlaylist creates a new playlist owned by userID.
func (s *Service) CreatePlaylist(ctx context.Context, userID string, req CreateRequest) (UserPlaylist, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return UserPlaylist{}, ErrInvalidInput
	}
	now := s.now().UTC()
	p := UserPlaylist{
		ID:          newID(),
		UserID:      strings.TrimSpace(userID),
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		TrackIDs:    []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Save(ctx, p); err != nil {
		return UserPlaylist{}, err
	}
	return p, nil
}

// ListPlaylists returns all playlists owned by userID.
func (s *Service) ListPlaylists(ctx context.Context, userID string) ([]UserPlaylist, error) {
	return s.repo.ListByUser(ctx, strings.TrimSpace(userID))
}

// GetPlaylist returns a playlist by id, verifying ownership.
func (s *Service) GetPlaylist(ctx context.Context, userID, id string) (UserPlaylist, error) {
	p, err := s.repo.Get(ctx, strings.TrimSpace(id))
	if err != nil {
		return UserPlaylist{}, err
	}
	if p.UserID != strings.TrimSpace(userID) {
		return UserPlaylist{}, ErrForbidden
	}
	return p, nil
}

// UpdatePlaylist patches name and/or description for a playlist owned by userID.
func (s *Service) UpdatePlaylist(ctx context.Context, userID, id string, req UpdateRequest) (UserPlaylist, error) {
	p, err := s.GetPlaylist(ctx, userID, id)
	if err != nil {
		return UserPlaylist{}, err
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return UserPlaylist{}, ErrInvalidInput
		}
		p.Name = name
	}
	if req.Description != nil {
		p.Description = strings.TrimSpace(*req.Description)
	}
	p.UpdatedAt = s.now().UTC()
	if err := s.repo.Save(ctx, p); err != nil {
		return UserPlaylist{}, err
	}
	return p, nil
}

// DeletePlaylist deletes a playlist owned by userID.
func (s *Service) DeletePlaylist(ctx context.Context, userID, id string) error {
	if _, err := s.GetPlaylist(ctx, userID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, strings.TrimSpace(id))
}

// AddTrack appends trackID to the playlist if not already present.
func (s *Service) AddTrack(ctx context.Context, userID, playlistID, trackID string) error {
	p, err := s.GetPlaylist(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	trackID = strings.TrimSpace(trackID)
	for _, tid := range p.TrackIDs {
		if tid == trackID {
			return nil // already present — idempotent
		}
	}
	p.TrackIDs = append(p.TrackIDs, trackID)
	p.UpdatedAt = s.now().UTC()
	return s.repo.Save(ctx, p)
}

// RemoveTrack removes trackID from the playlist. Idempotent.
func (s *Service) RemoveTrack(ctx context.Context, userID, playlistID, trackID string) error {
	p, err := s.GetPlaylist(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	trackID = strings.TrimSpace(trackID)
	filtered := p.TrackIDs[:0]
	for _, tid := range p.TrackIDs {
		if tid != trackID {
			filtered = append(filtered, tid)
		}
	}
	p.TrackIDs = filtered
	p.UpdatedAt = s.now().UTC()
	return s.repo.Save(ctx, p)
}

// SetTracks replaces the playlist track list entirely.
func (s *Service) SetTracks(ctx context.Context, userID, playlistID string, trackIDs []string) error {
	p, err := s.GetPlaylist(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(trackIDs))
	for _, tid := range trackIDs {
		if t := strings.TrimSpace(tid); t != "" {
			ids = append(ids, t)
		}
	}
	p.TrackIDs = ids
	p.UpdatedAt = s.now().UTC()
	return s.repo.Save(ctx, p)
}

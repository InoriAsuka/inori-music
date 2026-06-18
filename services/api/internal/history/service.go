package history

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	DefaultListLimit = 50
	MaxListLimit     = 500
)

// Service coordinates playback history persistence.
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// RecordPlay saves one play event for the given user and track.
// playedAt is the client-reported time; it defaults to now when zero.
func (s *Service) RecordPlay(ctx context.Context, userID, trackID string, playedAt time.Time) (PlayEvent, error) {
	if userID == "" {
		return PlayEvent{}, fmt.Errorf("userID is required")
	}
	if trackID == "" {
		return PlayEvent{}, fmt.Errorf("trackID is required")
	}
	if playedAt.IsZero() {
		playedAt = s.now()
	}
	id, err := newID()
	if err != nil {
		return PlayEvent{}, fmt.Errorf("generate event id: %w", err)
	}
	e := PlayEvent{
		ID:        id,
		UserID:    userID,
		TrackID:   trackID,
		PlayedAt:  playedAt.UTC(),
		CreatedAt: s.now().UTC(),
	}
	if err := s.repo.SavePlayEvent(ctx, e); err != nil {
		return PlayEvent{}, err
	}
	return e, nil
}

// ListPlays returns the play events for a user in reverse-chronological order.
func (s *Service) ListPlays(ctx context.Context, f PlayEventFilter) ([]PlayEvent, int, error) {
	if f.Limit <= 0 {
		f.Limit = DefaultListLimit
	}
	if f.Limit > MaxListLimit {
		f.Limit = MaxListLimit
	}
	return s.repo.ListPlayEvents(ctx, f)
}

// ClearHistory deletes all play events for the given user.
func (s *Service) ClearHistory(ctx context.Context, userID string) error {
	return s.repo.DeletePlayEventsByUser(ctx, userID)
}

// GetUserHistory returns paginated play events for any user; intended for admin use.
func (s *Service) GetUserHistory(ctx context.Context, f PlayEventFilter) ([]PlayEvent, int, error) {
	if f.Limit <= 0 {
		f.Limit = DefaultListLimit
	}
	if f.Limit > MaxListLimit {
		f.Limit = MaxListLimit
	}
	return s.repo.ListPlayEvents(ctx, f)
}

// GetTrackHistory returns paginated play events for a specific track across all users; intended for admin use.
func (s *Service) GetTrackHistory(ctx context.Context, f AdminPlayEventFilter) ([]PlayEvent, int, error) {
	if f.Limit <= 0 {
		f.Limit = DefaultListLimit
	}
	if f.Limit > MaxListLimit {
		f.Limit = MaxListLimit
	}
	return s.repo.ListPlayEventsByTrack(ctx, f)
}

// GetHistoryStats returns system-wide aggregate counts for admin use.
// f.Since optionally bounds the query to events on or after that time.
func (s *Service) GetHistoryStats(ctx context.Context, f StatsFilter) (HistoryStats, error) {
	return s.repo.HistoryStats(ctx, f)
}

// GetTopTracks returns the most-played tracks across all users.
// limit ≤ 0 defaults to 10 and is clamped to 100.
// f.Since optionally bounds the query to events on or after that time.
func (s *Service) GetTopTracks(ctx context.Context, f StatsFilter, limit int) ([]TrackPlayCount, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.TopTracks(ctx, f, limit)
}

// GetTopUsers returns the users with the most play events.
// limit ≤ 0 defaults to 10 and is clamped to 100.
// f.Since optionally bounds the query to events on or after that time.
func (s *Service) GetTopUsers(ctx context.Context, f StatsFilter, limit int) ([]UserPlayCount, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.TopUsers(ctx, f, limit)
}

func newID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

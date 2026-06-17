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

func newID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

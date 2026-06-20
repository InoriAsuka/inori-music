package favorites

import (
	"context"
	"strings"
	"time"
)

const (
	DefaultListLimit = 50
	MaxListLimit     = 200
)

// Service coordinates user favorites persistence rules.
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// WithClock replaces the time source. Intended for tests.
func (s *Service) WithClock(fn func() time.Time) *Service {
	s.now = fn
	return s
}

// AddFavorite records a user–track favorite. Idempotent when the entry already exists.
func (s *Service) AddFavorite(ctx context.Context, userID, trackID string) error {
	userID = strings.TrimSpace(userID)
	trackID = strings.TrimSpace(trackID)
	return s.repo.AddFavorite(ctx, userID, trackID, s.now().UTC())
}

// RemoveFavorite deletes a user–track favorite. Idempotent when the entry is absent.
func (s *Service) RemoveFavorite(ctx context.Context, userID, trackID string) error {
	userID = strings.TrimSpace(userID)
	trackID = strings.TrimSpace(trackID)
	return s.repo.RemoveFavorite(ctx, userID, trackID)
}

// ListFavorites returns a paginated page of favorite track IDs for userID (newest first).
// Limit is clamped to [1, MaxListLimit]; a zero limit defaults to DefaultListLimit.
func (s *Service) ListFavorites(ctx context.Context, userID string, limit, offset int) (FavoritesPage, error) {
	if limit <= 0 {
		limit = DefaultListLimit
	}
	if limit > MaxListLimit {
		limit = MaxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListFavorites(ctx, strings.TrimSpace(userID), limit, offset)
}

// IsFavorite reports whether userID has favorited trackID.
func (s *Service) IsFavorite(ctx context.Context, userID, trackID string) (bool, error) {
	return s.repo.IsFavorite(ctx, strings.TrimSpace(userID), strings.TrimSpace(trackID))
}

// AreFavorites returns a map of trackID → bool for the given slice, indicating
// which are favorited by userID. Used to annotate catalog track lists.
func (s *Service) AreFavorites(ctx context.Context, userID string, trackIDs []string) (map[string]bool, error) {
	if len(trackIDs) == 0 {
		return map[string]bool{}, nil
	}
	return s.repo.AreFavorites(ctx, strings.TrimSpace(userID), trackIDs)
}

package playerstate

import (
	"context"
	"strings"
	"time"
)

// MaxQueueLength is the server-side cap on the number of tracks a player state
// queue may contain. Writes exceeding this are rejected with ErrInvalidInput.
const MaxQueueLength = 500

// Service coordinates cross-device player state persistence rules.
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

// Get returns the stored player state for userID, or ErrNotFound if none exists.
func (s *Service) Get(ctx context.Context, userID string) (PlayerState, error) {
	return s.repo.Get(ctx, strings.TrimSpace(userID))
}

// Put validates and upserts the player state for userID. The server assigns
// UpdatedAt from its own clock (last-write-wins); any client-supplied value is
// ignored. Returns the stored state including the server timestamp.
func (s *Service) Put(ctx context.Context, userID string, state PlayerState) (PlayerState, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return PlayerState{}, ErrInvalidInput
	}

	if state.Queue == nil {
		state.Queue = []string{}
	}
	if len(state.Queue) > MaxQueueLength {
		return PlayerState{}, ErrInvalidInput
	}

	// Clamp currentIndex to a valid range. An empty queue pins it to 0; a
	// non-empty queue requires the index to reference an existing entry.
	if len(state.Queue) == 0 {
		state.CurrentIndex = 0
	} else if state.CurrentIndex < 0 || state.CurrentIndex >= len(state.Queue) {
		return PlayerState{}, ErrInvalidInput
	}

	if state.PositionSeconds < 0 {
		state.PositionSeconds = 0
	}

	state.UpdatedAt = s.now().UTC()
	if err := s.repo.Put(ctx, userID, state); err != nil {
		return PlayerState{}, err
	}
	return state, nil
}

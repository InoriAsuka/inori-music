package playerstate

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrNotFound is returned when a user has never reported a player state.
	ErrNotFound = errors.New("player state not found")
	// ErrInvalidInput is returned when a submitted player state violates the
	// server-side constraints (e.g. queue too long, index out of bounds).
	ErrInvalidInput = errors.New("invalid player state input")
)

// PlayerState is a single user's cross-device playback snapshot. There is at
// most one row per user (last-write-wins); no version history is kept.
type PlayerState struct {
	// Queue is the ordered list of track IDs in the current playback queue.
	Queue []string `json:"queue"`
	// CurrentIndex is the position within Queue of the active track.
	CurrentIndex int `json:"currentIndex"`
	// PositionSeconds is the playback offset into the current track, in seconds.
	PositionSeconds float64 `json:"positionSeconds"`
	// Repeat is the repeat mode ("off", "one", "all").
	Repeat string `json:"repeat"`
	// Shuffle reports whether shuffle is enabled.
	Shuffle bool `json:"shuffle"`
	// Volume is the playback volume in [0, 1].
	Volume float64 `json:"volume"`
	// Speed is the playback speed multiplier.
	Speed float64 `json:"speed"`
	// Status is the transport state ("playing", "paused", "stopped").
	Status string `json:"status"`
	// UpdatedAt is the server-assigned write time (last-write-wins clock).
	UpdatedAt time.Time `json:"updatedAt"`
}

// Repository persists the single-row-per-user player state.
type Repository interface {
	// Get returns the stored player state for userID, or ErrNotFound if the
	// user has never reported one.
	Get(ctx context.Context, userID string) (PlayerState, error)
	// Put upserts the player state for userID (single row, last-write-wins).
	Put(ctx context.Context, userID string, state PlayerState) error
}

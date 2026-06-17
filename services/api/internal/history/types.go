package history

import (
	"context"
	"time"
)

// PlayEvent records one listening event: a user played a track at a point in time.
type PlayEvent struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	TrackID   string    `json:"trackId"`
	PlayedAt  time.Time `json:"playedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// PlayEventFilter scopes a list of play events.
type PlayEventFilter struct {
	UserID  string // required
	TrackID string // optional — filter to a single track
	Limit   int    // 0 → default (50); clamped to 500
	Offset  int
}

// Repository persists play events.
type Repository interface {
	SavePlayEvent(ctx context.Context, e PlayEvent) error
	ListPlayEvents(ctx context.Context, f PlayEventFilter) ([]PlayEvent, int, error)
	DeletePlayEventsByUser(ctx context.Context, userID string) error

	// Aggregate stats — admin-facing queries.
	HistoryStats(ctx context.Context) (HistoryStats, error)
	TopTracks(ctx context.Context, limit int) ([]TrackPlayCount, error)
	TopUsers(ctx context.Context, limit int) ([]UserPlayCount, error)
}

// HistoryStats holds system-wide playback aggregate counts.
type HistoryStats struct {
	TotalEvents  int `json:"totalEvents"`
	UniqueUsers  int `json:"uniqueUsers"`
	UniqueTracks int `json:"uniqueTracks"`
}

// TrackPlayCount holds a track's total play count across all users.
type TrackPlayCount struct {
	TrackID   string `json:"trackId"`
	PlayCount int    `json:"playCount"`
}

// UserPlayCount holds a user's total play event count.
type UserPlayCount struct {
	UserID    string `json:"userId"`
	PlayCount int    `json:"playCount"`
}

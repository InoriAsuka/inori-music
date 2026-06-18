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

// PlayEventFilter scopes a list of play events for a specific user.
type PlayEventFilter struct {
	UserID  string // required
	TrackID string // optional — filter to a single track
	Limit   int    // 0 → default (50); clamped to 500
	Offset  int
}

// AdminPlayEventFilter scopes admin list queries that are not user-scoped.
type AdminPlayEventFilter struct {
	TrackID string // required for ListPlayEventsByTrack
	UserID  string // required for ListPlayEventsByUser (admin view)
	Limit   int    // 0 → default (50); clamped to 500
	Offset  int
}

// GlobalPlayEventFilter scopes admin queries that list all events across every user and track.
// All fields are optional filters; zero values mean "no restriction".
type GlobalPlayEventFilter struct {
	UserID  string    // optional — filter to a single user
	TrackID string    // optional — filter to a single track
	Since   time.Time // optional lower bound on played_at (inclusive)
	Until   time.Time // optional upper bound on played_at (exclusive)
	Limit   int       // 0 → default (50); clamped to 500
	Offset  int
}

// StatsFilter scopes admin aggregate queries.
// Zero-value Since/Until means no bound on that side.
type StatsFilter struct {
	Since time.Time // optional lower bound on played_at (inclusive)
	Until time.Time // optional upper bound on played_at (exclusive)
}

// UserStatsFilter scopes viewer aggregate queries to a single authenticated user.
type UserStatsFilter struct {
	UserID string    // required — injected from auth context
	Since  time.Time // optional lower bound on played_at (inclusive)
	Until  time.Time // optional upper bound on played_at (exclusive)
}

// Repository persists play events.
type Repository interface {
	SavePlayEvent(ctx context.Context, e PlayEvent) error
	ListPlayEvents(ctx context.Context, f PlayEventFilter) ([]PlayEvent, int, error)
	DeletePlayEventsByUser(ctx context.Context, userID string) error

	// Admin detail queries — not scoped to the requesting user.
	ListPlayEventsByTrack(ctx context.Context, f AdminPlayEventFilter) ([]PlayEvent, int, error)

	// Admin global list — unscoped, with optional user/track/time filters.
	ListAllPlayEvents(ctx context.Context, f GlobalPlayEventFilter) ([]PlayEvent, int, error)

	// Admin bulk-delete queries.
	DeletePlayEventsByUserAdmin(ctx context.Context, userID string) error
	DeletePlayEventsByTrack(ctx context.Context, trackID string) error
	DeletePlayEventsInWindow(ctx context.Context, f StatsFilter) error

	// Aggregate stats — admin-facing queries.
	HistoryStats(ctx context.Context, f StatsFilter) (HistoryStats, error)
	TopTracks(ctx context.Context, f StatsFilter, limit int) ([]TrackPlayCount, error)
	TopUsers(ctx context.Context, f StatsFilter, limit int) ([]UserPlayCount, error)

	// Viewer-scoped aggregate stats — restricted to the authenticated user.
	UserTopTracks(ctx context.Context, f UserStatsFilter, limit int) ([]TrackPlayCount, error)
	UserHistoryStats(ctx context.Context, f UserStatsFilter) (UserHistoryStats, error)
}

// HistoryStats holds system-wide playback aggregate counts.
type HistoryStats struct {
	TotalEvents  int `json:"totalEvents"`
	UniqueUsers  int `json:"uniqueUsers"`
	UniqueTracks int `json:"uniqueTracks"`
}

// UserHistoryStats holds per-user playback aggregate counts.
// UniqueUsers is always 1 for a single user, so it is omitted.
type UserHistoryStats struct {
	TotalEvents  int `json:"totalEvents"`
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

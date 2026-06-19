package history

import (
	"context"
	"errors"
	"time"
)

// ErrEventNotFound is returned when a play event with the requested ID does not exist.
var ErrEventNotFound = errors.New("play event not found")

// ErrEventForbidden is returned when a viewer tries to access an event they do not own.
var ErrEventForbidden = errors.New("play event belongs to another user")

// ErrInvalidTimeRange is returned when Since >= Until or a required time bound is missing.
var ErrInvalidTimeRange = errors.New("invalid time range")

// TimelineGranularity controls how play events are bucketed in a timeline query.
type TimelineGranularity string

const (
	GranularityDay   TimelineGranularity = "day"
	GranularityWeek  TimelineGranularity = "week"
	GranularityMonth TimelineGranularity = "month"
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
	UserID  string    // required
	TrackID string    // optional — filter to a single track
	Since   time.Time // optional lower bound on played_at (inclusive)
	Until   time.Time // optional upper bound on played_at (exclusive)
	Limit   int       // 0 → default (50); clamped to 500
	Offset  int
	Asc     bool // false (default) → played_at DESC; true → played_at ASC
}

// AdminPlayEventFilter scopes admin list queries that are not user-scoped.
type AdminPlayEventFilter struct {
	TrackID string    // required for ListPlayEventsByTrack
	UserID  string    // required for ListPlayEventsByUser (admin view)
	Since   time.Time // optional lower bound on played_at (inclusive)
	Until   time.Time // optional upper bound on played_at (exclusive)
	Limit   int       // 0 → default (50); clamped to 500
	Offset  int
	Asc     bool // false (default) → played_at DESC; true → played_at ASC
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
	Asc     bool // false (default) → played_at DESC; true → played_at ASC
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

// TrackStatsFilter scopes admin aggregate queries to a single track.
type TrackStatsFilter struct {
	TrackID string    // required
	Since   time.Time // optional lower bound on played_at (inclusive)
	Until   time.Time // optional upper bound on played_at (exclusive)
}

// TimelineFilter scopes a timeline (bucket) query.
// Since and Until are both required; Granularity defaults to GranularityDay.
type TimelineFilter struct {
	Since       time.Time           // required lower bound (inclusive)
	Until       time.Time           // required upper bound (exclusive)
	Granularity TimelineGranularity // day | week | month; default day
	UserID      string              // optional — restrict to a single user
	TrackID     string              // optional — restrict to a single track
}

// TimelineBucket holds the event count for a single time bucket.
type TimelineBucket struct {
	BucketStart time.Time `json:"bucketStart"`
	EventCount  int       `json:"eventCount"`
}

// Repository persists play events.
type Repository interface {
	SavePlayEvent(ctx context.Context, e PlayEvent) error
	ListPlayEvents(ctx context.Context, f PlayEventFilter) ([]PlayEvent, int, error)
	DeletePlayEventsByUser(ctx context.Context, userID string) error

	// Per-event operations — used by both admin and viewer single-event endpoints.
	GetPlayEventByID(ctx context.Context, id string) (PlayEvent, error)
	UpdatePlayEventByID(ctx context.Context, id string, playedAt time.Time) (PlayEvent, error)
	DeletePlayEventByID(ctx context.Context, id string) error

	// Batch-delete by explicit ID list.
	DeletePlayEventsByIDs(ctx context.Context, ids []string) (int, error)

	// Viewer-scoped batch-delete — only deletes events that belong to userID.
	DeletePlayEventsByIDsForUser(ctx context.Context, userID string, ids []string) (int, error)

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
	UserTrackPlayStats(ctx context.Context, userID, trackID string) (UserTrackStats, error)

	// Track-scoped aggregate stats — admin-facing queries for a single track.
	TrackHistoryStats(ctx context.Context, f TrackStatsFilter) (TrackHistoryStatsResult, error)
	TrackTopListeners(ctx context.Context, f TrackStatsFilter, limit int) ([]UserPlayCount, error)

	// Timeline — group event counts by time bucket (day/week/month).
	HistoryTimeline(ctx context.Context, f TimelineFilter) ([]TimelineBucket, error)
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

// TrackHistoryStatsResult holds per-track playback aggregate counts.
type TrackHistoryStatsResult struct {
	TotalEvents     int `json:"totalEvents"`
	UniqueListeners int `json:"uniqueListeners"`
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

// UserTrackStats holds aggregate play counts for one (user, track) pair.
// FirstPlayedAt and LastPlayedAt are zero when TotalPlays is zero.
type UserTrackStats struct {
	TrackID       string    `json:"trackId"`
	TotalPlays    int       `json:"totalPlays"`
	FirstPlayedAt time.Time `json:"firstPlayedAt,omitempty"`
	LastPlayedAt  time.Time `json:"lastPlayedAt,omitempty"`
}

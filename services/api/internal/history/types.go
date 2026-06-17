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
}

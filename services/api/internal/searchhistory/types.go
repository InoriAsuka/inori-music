package searchhistory

import (
	"context"
	"time"
)

// MaxEntries is the number of recent search queries retained per user. Both
// reads and writes are capped to this many entries (newest first).
const MaxEntries = 20

// Entry records that a user issued a search query at a point in time.
type Entry struct {
	Query      string    `json:"query"`
	SearchedAt time.Time `json:"searchedAt"`
}

// Repository persists per-user recent search queries.
type Repository interface {
	// List returns the user's recent queries, newest first, capped to MaxEntries.
	List(ctx context.Context, userID string) ([]Entry, error)
	// Replace overwrites the user's entire search history with entries. An empty
	// slice clears the history.
	Replace(ctx context.Context, userID string, entries []Entry) error
	// Clear removes all search history for the user.
	Clear(ctx context.Context, userID string) error
}

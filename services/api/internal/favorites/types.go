package favorites

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrFavoriteNotFound is returned when a favorite entry does not exist.
	ErrFavoriteNotFound = errors.New("favorite not found")
)

// FavoriteEntry records that a user has favorited a specific track.
type FavoriteEntry struct {
	UserID    string    `json:"userId"`
	TrackID   string    `json:"trackId"`
	CreatedAt time.Time `json:"createdAt"`
}

// FavoritesPage carries a paginated list of track IDs from a user's favorites,
// ordered by creation time (newest first).
type FavoritesPage struct {
	TrackIDs []string `json:"trackIds"`
	Total    int      `json:"total"`
}

// Repository persists user–track favorite relationships.
type Repository interface {
	// AddFavorite records the favorite. Idempotent: no-op if the entry already exists.
	AddFavorite(ctx context.Context, userID, trackID string, now time.Time) error
	// RemoveFavorite deletes the favorite. Idempotent: no error if the entry does not exist.
	RemoveFavorite(ctx context.Context, userID, trackID string) error
	// ListFavorites returns the ordered track IDs for a user with pagination (newest first).
	ListFavorites(ctx context.Context, userID string, limit, offset int) (FavoritesPage, error)
	// IsFavorite reports whether userID has favorited trackID.
	IsFavorite(ctx context.Context, userID, trackID string) (bool, error)
	// AreFavorites returns the subset of trackIDs that userID has favorited.
	AreFavorites(ctx context.Context, userID string, trackIDs []string) (map[string]bool, error)
}

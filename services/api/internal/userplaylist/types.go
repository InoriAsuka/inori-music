package userplaylist

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("user playlist not found")
	ErrInvalidInput = errors.New("invalid user playlist input")
	ErrForbidden    = errors.New("access forbidden")
)

// UserPlaylist is a named collection of tracks owned by a user.
type UserPlaylist struct {
	ID          string
	UserID      string
	Name        string
	Description string
	TrackIDs    []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Repository persists user playlists.
type Repository interface {
	Save(ctx context.Context, p UserPlaylist) error
	Get(ctx context.Context, id string) (UserPlaylist, error)
	ListByUser(ctx context.Context, userID string) ([]UserPlaylist, error)
	Delete(ctx context.Context, id string) error
}

// CreateRequest carries the fields needed to create a new playlist.
type CreateRequest struct {
	Name        string
	Description string
}

// UpdateRequest carries optional fields to patch a playlist.
type UpdateRequest struct {
	Name        *string
	Description *string
}

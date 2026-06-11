package catalog

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidArtist  = errors.New("invalid artist")
	ErrArtistNotFound = errors.New("artist not found")
	ErrArtistConflict = errors.New("artist conflict")
	ErrInvalidAlbum   = errors.New("invalid album")
	ErrAlbumNotFound  = errors.New("album not found")
	ErrAlbumConflict  = errors.New("album conflict")
	ErrInvalidTrack   = errors.New("invalid track")
	ErrTrackNotFound  = errors.New("track not found")
	ErrTrackConflict  = errors.New("track conflict")
)

// Artist represents a music library artist or performer.
type Artist struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SortName  string    `json:"sortName,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Album represents a music library release.
type Album struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	SortTitle   string    `json:"sortTitle,omitempty"`
	ArtistID    string    `json:"artistId"`
	ReleaseYear int       `json:"releaseYear,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Track represents a playable music item and links catalog metadata to a media object.
type Track struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	SortTitle     string    `json:"sortTitle,omitempty"`
	ArtistID      string    `json:"artistId"`
	AlbumID       string    `json:"albumId,omitempty"`
	MediaObjectID string    `json:"mediaObjectId"`
	TrackNumber   int       `json:"trackNumber,omitempty"`
	DiscNumber    int       `json:"discNumber,omitempty"`
	DurationMS    int       `json:"durationMs,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Repository persists catalog metadata records.
type Repository interface {
	SaveArtist(ctx context.Context, artist Artist) error
	GetArtist(ctx context.Context, id string) (Artist, error)
	ListArtists(ctx context.Context) ([]Artist, error)
	DeleteArtist(ctx context.Context, id string) error

	SaveAlbum(ctx context.Context, album Album) error
	GetAlbum(ctx context.Context, id string) (Album, error)
	ListAlbums(ctx context.Context) ([]Album, error)
	ListAlbumsByArtist(ctx context.Context, artistID string) ([]Album, error)
	DeleteAlbum(ctx context.Context, id string) error

	SaveTrack(ctx context.Context, track Track) error
	GetTrack(ctx context.Context, id string) (Track, error)
	ListTracks(ctx context.Context) ([]Track, error)
	ListTracksByAlbum(ctx context.Context, albumID string) ([]Track, error)
	ListTracksByArtist(ctx context.Context, artistID string) ([]Track, error)
	DeleteTrack(ctx context.Context, id string) error
}

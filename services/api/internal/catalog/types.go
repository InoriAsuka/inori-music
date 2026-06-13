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
	// ErrImportRejected is returned when a media object cannot be imported as a track
	// (e.g. wrong asset kind or lifecycle state).
	ErrImportRejected = errors.New("import rejected")
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

// MediaObjectInfo carries the subset of media object metadata that the catalog
// import workflow needs without importing the storage package.
type MediaObjectInfo struct {
	ID             string
	AssetKind      string
	LifecycleState string
	MIMEType       string
}

// MediaObjectReader fetches a single media object's metadata by ID.
// It is satisfied by *storage.MediaObjectService without introducing a hard
// import cycle between the catalog and storage packages.
type MediaObjectReader interface {
	GetMediaObjectInfo(ctx context.Context, id string) (MediaObjectInfo, error)
}

// ImportTrackRequest carries caller-supplied metadata for the import workflow.
type ImportTrackRequest struct {
	MediaObjectID string
	Title         string
	SortTitle     string
	ArtistID      string
	AlbumID       string
	TrackNumber   int
	DiscNumber    int
	DurationMS    int
}

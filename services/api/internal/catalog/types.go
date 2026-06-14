package catalog

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidArtist    = errors.New("invalid artist")
	ErrArtistNotFound   = errors.New("artist not found")
	ErrArtistConflict   = errors.New("artist conflict")
	ErrInvalidAlbum     = errors.New("invalid album")
	ErrAlbumNotFound    = errors.New("album not found")
	ErrAlbumConflict    = errors.New("album conflict")
	ErrInvalidTrack     = errors.New("invalid track")
	ErrTrackNotFound    = errors.New("track not found")
	ErrTrackConflict    = errors.New("track conflict")
	ErrInvalidPlaylist  = errors.New("invalid playlist")
	ErrPlaylistNotFound = errors.New("playlist not found")
	// ErrImportRejected is returned when a media object cannot be imported as a track
	// (e.g. wrong asset kind or lifecycle state).
	ErrImportRejected = errors.New("import rejected")
	// ErrRelinkRejected is returned when a media object cannot be used to relink an
	// existing track (e.g. wrong asset kind or lifecycle state).
	ErrRelinkRejected = errors.New("relink rejected")
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

// SearchResultKind identifies which entity kind a search hit belongs to.
type SearchResultKind string

const (
	SearchResultArtist SearchResultKind = "artist"
	SearchResultAlbum  SearchResultKind = "album"
	SearchResultTrack  SearchResultKind = "track"
)

// SearchResultItem is a single catalog entity returned from a full-text search.
type SearchResultItem struct {
	Kind   SearchResultKind `json:"kind"`
	Artist *Artist          `json:"artist,omitempty"`
	Album  *Album           `json:"album,omitempty"`
	Track  *Track           `json:"track,omitempty"`
}

// CatalogSearchResult holds the ordered result set from a catalog full-text search.
type CatalogSearchResult struct {
	Query string             `json:"query"`
	Items []SearchResultItem `json:"items"`
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

	// SearchCatalog performs a full-text search across artists, albums, and tracks.
	// The query string is tokenised using the simple text-search dictionary.
	// Implementations that do not support full-text search may fall back to
	// case-insensitive substring matching.
	SearchCatalog(ctx context.Context, query string) (CatalogSearchResult, error)

	SavePlaylist(ctx context.Context, playlist Playlist) error
	GetPlaylist(ctx context.Context, id string) (Playlist, error)
	ListPlaylists(ctx context.Context) ([]Playlist, error)
	DeletePlaylist(ctx context.Context, id string) error
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

// RelinkTrackRequest carries the new media object reference for an existing track.
type RelinkTrackRequest struct {
	MediaObjectID string
}


// UpdateArtistRequest carries the fields that may be changed via a PATCH request.
// Nil pointer fields are left unchanged; a pointer to an empty string clears the field.
type UpdateArtistRequest struct {
	Name     *string
	SortName *string
}

// UpdateAlbumRequest carries the fields that may be changed via a PATCH request.
// Nil pointer fields are left unchanged.
type UpdateAlbumRequest struct {
	Title       *string
	SortTitle   *string
	ArtistID    *string
	ReleaseYear *int
}

// UpdateTrackRequest carries the fields that may be changed via a PATCH request.
// Nil pointer fields are left unchanged.
type UpdateTrackRequest struct {
	Title       *string
	SortTitle   *string
	ArtistID    *string
	AlbumID     *string
	TrackNumber *int
	DiscNumber  *int
	DurationMS  *int
}

// BatchImportResultItem holds the outcome of a single import within a batch request.
// Exactly one of Track or Error is set.
type BatchImportResultItem struct {
	Index         int    `json:"index"`
	MediaObjectID string `json:"mediaObjectId"`
	// Track is populated on success.
	Track *Track `json:"track,omitempty"`
	// Error is populated on failure.
	Error string `json:"error,omitempty"`
	// ErrorCode is a machine-readable error code populated on failure.
	ErrorCode string `json:"errorCode,omitempty"`
}

// BatchImportResult is the aggregate result of a batch-import request.
type BatchImportResult struct {
	Total    int                     `json:"total"`
	Imported int                     `json:"imported"`
	Failed   int                     `json:"failed"`
	Items    []BatchImportResultItem `json:"items"`
}

// Playlist is an ordered collection of tracks curated by a user or administrator.
type Playlist struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	TrackIDs    []string  `json:"trackIds"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// UpdatePlaylistRequest carries the fields that may be changed via a PATCH request.
// Nil pointer fields are left unchanged.
type UpdatePlaylistRequest struct {
	Name        *string
	Description *string
}

// CatalogStats holds metadata-only aggregate counts for admin dashboards.
type CatalogStats struct {
	Artists   int `json:"artists"`
	Albums    int `json:"albums"`
	Tracks    int `json:"tracks"`
	Playlists int `json:"playlists"`
}

// ArtistStatItem holds per-artist album and track counts for the breakdown stats endpoint.
type ArtistStatItem struct {
	ArtistID   string `json:"artistId"`
	Name       string `json:"name"`
	AlbumCount int    `json:"albumCount"`
	TrackCount int    `json:"trackCount"`
}

// ArtistStatsBreakdown holds the per-artist breakdown returned by the stats/artists endpoint.
type ArtistStatsBreakdown struct {
	Artists []ArtistStatItem `json:"artists"`
}

// AlbumStatItem holds per-album track counts for the breakdown stats endpoint.
type AlbumStatItem struct {
	AlbumID    string `json:"albumId"`
	Title      string `json:"title"`
	ArtistID   string `json:"artistId"`
	TrackCount int    `json:"trackCount"`
}

// AlbumStatsBreakdown holds the per-album breakdown returned by the stats/albums endpoint.
type AlbumStatsBreakdown struct {
	Albums []AlbumStatItem `json:"albums"`
}

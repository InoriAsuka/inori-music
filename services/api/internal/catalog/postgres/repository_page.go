package catalogpg

import (
	"context"
	"fmt"
	"strings"

	"inori-music/services/api/internal/catalog"
)

// artistOrderBy returns the ORDER BY clause for artist page queries.
// A tiebreak on id ensures stable pagination.
func artistOrderBy(q catalog.ListQuery) string {
	col := "lower(sort_name), lower(name)"
	switch q.SortBy {
	case catalog.ArtistSortByName:
		col = "lower(name), lower(sort_name)"
	case catalog.ArtistSortBySortName:
		col = "lower(sort_name), lower(name)"
	case catalog.ArtistSortByCreatedAt:
		col = "created_at"
	case catalog.ArtistSortByUpdatedAt:
		col = "updated_at"
	}
	dir := "ASC"
	if strings.ToLower(strings.TrimSpace(q.SortOrder)) == catalog.CatalogSortOrderDesc {
		dir = "DESC"
	}
	return col + " " + dir + ", id " + dir
}

func albumOrderBy(q catalog.ListQuery) string {
	col := "lower(sort_title), lower(title)"
	switch q.SortBy {
	case catalog.AlbumSortByTitle:
		col = "lower(title), lower(sort_title)"
	case catalog.AlbumSortBySortTitle:
		col = "lower(sort_title), lower(title)"
	case catalog.AlbumSortByReleaseYear:
		col = "release_year"
	case catalog.AlbumSortByCreatedAt:
		col = "created_at"
	case catalog.AlbumSortByUpdatedAt:
		col = "updated_at"
	}
	dir := "ASC"
	if strings.ToLower(strings.TrimSpace(q.SortOrder)) == catalog.CatalogSortOrderDesc {
		dir = "DESC"
	}
	return col + " " + dir + ", id " + dir
}

func trackOrderBy(q catalog.ListQuery) string {
	col := "lower(sort_title), lower(title)"
	switch q.SortBy {
	case catalog.TrackSortByTitle:
		col = "lower(title), lower(sort_title)"
	case catalog.TrackSortBySortTitle:
		col = "lower(sort_title), lower(title)"
	case catalog.TrackSortByTrackNumber:
		col = "track_number"
	case catalog.TrackSortByDiscNumber:
		col = "disc_number"
	case catalog.TrackSortByDurationMS:
		col = "duration_ms"
	case catalog.TrackSortByCreatedAt:
		col = "created_at"
	case catalog.TrackSortByUpdatedAt:
		col = "updated_at"
	}
	dir := "ASC"
	if strings.ToLower(strings.TrimSpace(q.SortOrder)) == catalog.CatalogSortOrderDesc {
		dir = "DESC"
	}
	return col + " " + dir + ", id " + dir
}

func playlistOrderBy(q catalog.ListQuery) string {
	col := "lower(name)"
	switch q.SortBy {
	case catalog.PlaylistSortByCreatedAt:
		col = "created_at"
	case catalog.PlaylistSortByUpdatedAt:
		col = "updated_at"
	}
	dir := "ASC"
	if strings.ToLower(strings.TrimSpace(q.SortOrder)) == catalog.CatalogSortOrderDesc {
		dir = "DESC"
	}
	return col + " " + dir + ", id " + dir
}

func (r *Repository) ListArtistsPage(ctx context.Context, q catalog.ListQuery) (catalog.ListPage[catalog.Artist], error) {
	sql := fmt.Sprintf(`
		SELECT id, name, sort_name, created_at, updated_at,
		       COUNT(*) OVER () AS total_count
		FROM artists
		ORDER BY %s
		LIMIT $1 OFFSET $2`, artistOrderBy(q))
	rows, err := r.pool.Query(ctx, sql, q.Limit, q.Offset)
	if err != nil {
		return catalog.ListPage[catalog.Artist]{}, err
	}
	defer rows.Close()
	var artists []catalog.Artist
	total := 0
	for rows.Next() {
		var a catalog.Artist
		if err := rows.Scan(&a.ID, &a.Name, &a.SortName, &a.CreatedAt, &a.UpdatedAt, &total); err != nil {
			return catalog.ListPage[catalog.Artist]{}, err
		}
		artists = append(artists, a)
	}
	if err := rows.Err(); err != nil {
		return catalog.ListPage[catalog.Artist]{}, err
	}
	if artists == nil {
		artists = []catalog.Artist{}
	}
	return catalog.ListPage[catalog.Artist]{Items: artists, Total: total}, nil
}

func (r *Repository) ListAlbumsPage(ctx context.Context, q catalog.ListQuery) (catalog.ListPage[catalog.Album], error) {
	sql := fmt.Sprintf(`
		SELECT id, title, sort_title, artist_id, release_year, created_at, updated_at,
		       COUNT(*) OVER () AS total_count
		FROM albums
		ORDER BY %s
		LIMIT $1 OFFSET $2`, albumOrderBy(q))
	return r.queryAlbumsPage(ctx, sql, q.Limit, q.Offset)
}

func (r *Repository) ListAlbumsByArtistPage(ctx context.Context, artistID string, q catalog.ListQuery) (catalog.ListPage[catalog.Album], error) {
	sql := fmt.Sprintf(`
		SELECT id, title, sort_title, artist_id, release_year, created_at, updated_at,
		       COUNT(*) OVER () AS total_count
		FROM albums
		WHERE artist_id = $3
		ORDER BY %s
		LIMIT $1 OFFSET $2`, albumOrderBy(q))
	return r.queryAlbumsPage(ctx, sql, q.Limit, q.Offset, artistID)
}

func (r *Repository) queryAlbumsPage(ctx context.Context, sql string, args ...any) (catalog.ListPage[catalog.Album], error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return catalog.ListPage[catalog.Album]{}, err
	}
	defer rows.Close()
	var albums []catalog.Album
	total := 0
	for rows.Next() {
		var a catalog.Album
		if err := rows.Scan(&a.ID, &a.Title, &a.SortTitle, &a.ArtistID, &a.ReleaseYear, &a.CreatedAt, &a.UpdatedAt, &total); err != nil {
			return catalog.ListPage[catalog.Album]{}, err
		}
		albums = append(albums, a)
	}
	if err := rows.Err(); err != nil {
		return catalog.ListPage[catalog.Album]{}, err
	}
	if albums == nil {
		albums = []catalog.Album{}
	}
	return catalog.ListPage[catalog.Album]{Items: albums, Total: total}, nil
}

func (r *Repository) ListTracksPage(ctx context.Context, q catalog.ListQuery) (catalog.ListPage[catalog.Track], error) {
	sql := fmt.Sprintf(`
		SELECT id, title, sort_title, artist_id, COALESCE(album_id,''), media_object_id,
		       track_number, disc_number, duration_ms, created_at, updated_at,
		       COUNT(*) OVER () AS total_count
		FROM tracks
		ORDER BY %s
		LIMIT $1 OFFSET $2`, trackOrderBy(q))
	return r.queryTracksPage(ctx, sql, q.Limit, q.Offset)
}

func (r *Repository) ListTracksByAlbumPage(ctx context.Context, albumID string, q catalog.ListQuery) (catalog.ListPage[catalog.Track], error) {
	sql := fmt.Sprintf(`
		SELECT id, title, sort_title, artist_id, COALESCE(album_id,''), media_object_id,
		       track_number, disc_number, duration_ms, created_at, updated_at,
		       COUNT(*) OVER () AS total_count
		FROM tracks
		WHERE album_id = $3
		ORDER BY %s
		LIMIT $1 OFFSET $2`, trackOrderBy(q))
	return r.queryTracksPage(ctx, sql, q.Limit, q.Offset, albumID)
}

func (r *Repository) ListTracksByArtistPage(ctx context.Context, artistID string, q catalog.ListQuery) (catalog.ListPage[catalog.Track], error) {
	sql := fmt.Sprintf(`
		SELECT id, title, sort_title, artist_id, COALESCE(album_id,''), media_object_id,
		       track_number, disc_number, duration_ms, created_at, updated_at,
		       COUNT(*) OVER () AS total_count
		FROM tracks
		WHERE artist_id = $3
		ORDER BY %s
		LIMIT $1 OFFSET $2`, trackOrderBy(q))
	return r.queryTracksPage(ctx, sql, q.Limit, q.Offset, artistID)
}

func (r *Repository) queryTracksPage(ctx context.Context, sql string, args ...any) (catalog.ListPage[catalog.Track], error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return catalog.ListPage[catalog.Track]{}, err
	}
	defer rows.Close()
	var tracks []catalog.Track
	total := 0
	for rows.Next() {
		var t catalog.Track
		if err := rows.Scan(
			&t.ID, &t.Title, &t.SortTitle, &t.ArtistID, &t.AlbumID, &t.MediaObjectID,
			&t.TrackNumber, &t.DiscNumber, &t.DurationMS, &t.CreatedAt, &t.UpdatedAt, &total,
		); err != nil {
			return catalog.ListPage[catalog.Track]{}, err
		}
		tracks = append(tracks, t)
	}
	if err := rows.Err(); err != nil {
		return catalog.ListPage[catalog.Track]{}, err
	}
	if tracks == nil {
		tracks = []catalog.Track{}
	}
	return catalog.ListPage[catalog.Track]{Items: tracks, Total: total}, nil
}

func (r *Repository) ListPlaylistsPage(ctx context.Context, q catalog.ListQuery) (catalog.ListPage[catalog.Playlist], error) {
	sql := fmt.Sprintf(`
		SELECT id, name, description, created_at, updated_at,
		       COUNT(*) OVER () AS total_count
		FROM playlists
		ORDER BY %s
		LIMIT $1 OFFSET $2`, playlistOrderBy(q))
	rows, err := r.pool.Query(ctx, sql, q.Limit, q.Offset)
	if err != nil {
		return catalog.ListPage[catalog.Playlist]{}, err
	}
	defer rows.Close()
	var playlists []catalog.Playlist
	total := 0
	for rows.Next() {
		var p catalog.Playlist
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt, &total); err != nil {
			return catalog.ListPage[catalog.Playlist]{}, err
		}
		p.TrackIDs = []string{}
		playlists = append(playlists, p)
	}
	if err := rows.Err(); err != nil {
		return catalog.ListPage[catalog.Playlist]{}, err
	}
	if playlists == nil {
		playlists = []catalog.Playlist{}
	}
	// Load track IDs for each playlist
	for i := range playlists {
		ids, err := r.loadPlaylistTracks(ctx, playlists[i].ID)
		if err != nil {
			return catalog.ListPage[catalog.Playlist]{}, err
		}
		playlists[i].TrackIDs = ids
	}
	return catalog.ListPage[catalog.Playlist]{Items: playlists, Total: total}, nil
}

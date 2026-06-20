package catalogpg

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	case catalog.TrackSortByGenre:
		col = "lower(COALESCE(genre,''))"
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
	if q.ReleaseYearMin > 0 || q.ReleaseYearMax > 0 {
		var clauses []string
		args := []any{q.Limit, q.Offset}
		if q.ReleaseYearMin > 0 {
			args = append(args, q.ReleaseYearMin)
			clauses = append(clauses, fmt.Sprintf("release_year >= $%d", len(args)))
		}
		if q.ReleaseYearMax > 0 {
			args = append(args, q.ReleaseYearMax)
			clauses = append(clauses, fmt.Sprintf("release_year <= $%d", len(args)))
		}
		where := " WHERE " + strings.Join(clauses, " AND ")
		sql := fmt.Sprintf(`
			SELECT id, title, sort_title, artist_id, release_year, created_at, updated_at,
			       COUNT(*) OVER () AS total_count
			FROM albums%s
			ORDER BY %s
			LIMIT $1 OFFSET $2`, where, albumOrderBy(q))
		return r.queryAlbumsPage(ctx, sql, args...)
	}
	sql := fmt.Sprintf(`
		SELECT id, title, sort_title, artist_id, release_year, created_at, updated_at,
		       COUNT(*) OVER () AS total_count
		FROM albums
		ORDER BY %s
		LIMIT $1 OFFSET $2`, albumOrderBy(q))
	return r.queryAlbumsPage(ctx, sql, q.Limit, q.Offset)
}

func (r *Repository) ListAlbumsByArtistPage(ctx context.Context, artistID string, q catalog.ListQuery) (catalog.ListPage[catalog.Album], error) {
	args := []any{q.Limit, q.Offset, artistID}
	clauses := []string{"artist_id = $3"}
	if q.ReleaseYearMin > 0 {
		args = append(args, q.ReleaseYearMin)
		clauses = append(clauses, fmt.Sprintf("release_year >= $%d", len(args)))
	}
	if q.ReleaseYearMax > 0 {
		args = append(args, q.ReleaseYearMax)
		clauses = append(clauses, fmt.Sprintf("release_year <= $%d", len(args)))
	}
	where := strings.Join(clauses, " AND ")
	sql := fmt.Sprintf(`
		SELECT id, title, sort_title, artist_id, release_year, created_at, updated_at,
		       COUNT(*) OVER () AS total_count
		FROM albums
		WHERE %s
		ORDER BY %s
		LIMIT $1 OFFSET $2`, where, albumOrderBy(q))
	return r.queryAlbumsPage(ctx, sql, args...)
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
	if q.Genre != "" {
		sql := fmt.Sprintf(`
			SELECT id, title, sort_title, artist_id, COALESCE(album_id,''), media_object_id,
			       track_number, disc_number, duration_ms, COALESCE(genre,''), created_at, updated_at,
			       COUNT(*) OVER () AS total_count
			FROM tracks
			WHERE lower(COALESCE(genre,'')) = lower($3)
			ORDER BY %s
			LIMIT $1 OFFSET $2`, trackOrderBy(q))
		return r.queryTracksPage(ctx, sql, q.Limit, q.Offset, q.Genre)
	}
	sql := fmt.Sprintf(`
		SELECT id, title, sort_title, artist_id, COALESCE(album_id,''), media_object_id,
		       track_number, disc_number, duration_ms, COALESCE(genre,''), created_at, updated_at,
		       COUNT(*) OVER () AS total_count
		FROM tracks
		ORDER BY %s
		LIMIT $1 OFFSET $2`, trackOrderBy(q))
	return r.queryTracksPage(ctx, sql, q.Limit, q.Offset)
}

func (r *Repository) ListTracksByAlbumPage(ctx context.Context, albumID string, q catalog.ListQuery) (catalog.ListPage[catalog.Track], error) {
	if q.Genre != "" {
		sql := fmt.Sprintf(`
			SELECT id, title, sort_title, artist_id, COALESCE(album_id,''), media_object_id,
			       track_number, disc_number, duration_ms, COALESCE(genre,''), created_at, updated_at,
			       COUNT(*) OVER () AS total_count
			FROM tracks
			WHERE album_id = $3 AND lower(COALESCE(genre,'')) = lower($4)
			ORDER BY %s
			LIMIT $1 OFFSET $2`, trackOrderBy(q))
		return r.queryTracksPage(ctx, sql, q.Limit, q.Offset, albumID, q.Genre)
	}
	sql := fmt.Sprintf(`
		SELECT id, title, sort_title, artist_id, COALESCE(album_id,''), media_object_id,
		       track_number, disc_number, duration_ms, COALESCE(genre,''), created_at, updated_at,
		       COUNT(*) OVER () AS total_count
		FROM tracks
		WHERE album_id = $3
		ORDER BY %s
		LIMIT $1 OFFSET $2`, trackOrderBy(q))
	return r.queryTracksPage(ctx, sql, q.Limit, q.Offset, albumID)
}

func (r *Repository) ListTracksByArtistPage(ctx context.Context, artistID string, q catalog.ListQuery) (catalog.ListPage[catalog.Track], error) {
	if q.Genre != "" {
		sql := fmt.Sprintf(`
			SELECT id, title, sort_title, artist_id, COALESCE(album_id,''), media_object_id,
			       track_number, disc_number, duration_ms, COALESCE(genre,''), created_at, updated_at,
			       COUNT(*) OVER () AS total_count
			FROM tracks
			WHERE artist_id = $3 AND lower(COALESCE(genre,'')) = lower($4)
			ORDER BY %s
			LIMIT $1 OFFSET $2`, trackOrderBy(q))
		return r.queryTracksPage(ctx, sql, q.Limit, q.Offset, artistID, q.Genre)
	}
	sql := fmt.Sprintf(`
		SELECT id, title, sort_title, artist_id, COALESCE(album_id,''), media_object_id,
		       track_number, disc_number, duration_ms, COALESCE(genre,''), created_at, updated_at,
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
			&t.TrackNumber, &t.DiscNumber, &t.DurationMS, &t.Genre, &t.CreatedAt, &t.UpdatedAt, &total,
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

// ---- Aggregate stats methods ----

func (r *Repository) CountEntities(ctx context.Context) (catalog.CatalogStats, error) {
	var stats catalog.CatalogStats
	row := r.pool.QueryRow(ctx, `
		SELECT
		    (SELECT COUNT(*) FROM artists)  AS artists,
		    (SELECT COUNT(*) FROM albums)   AS albums,
		    (SELECT COUNT(*) FROM tracks)   AS tracks,
		    (SELECT COUNT(*) FROM playlists) AS playlists`)
	if err := row.Scan(&stats.Artists, &stats.Albums, &stats.Tracks, &stats.Playlists); err != nil {
		return catalog.CatalogStats{}, err
	}
	return stats, nil
}

func (r *Repository) ArtistAlbumTrackCounts(ctx context.Context) ([]catalog.ArtistStatItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.name,
		       COUNT(DISTINCT al.id) AS album_count,
		       COUNT(DISTINCT t.id)  AS track_count
		FROM artists a
		LEFT JOIN albums al ON al.artist_id = a.id
		LEFT JOIN tracks t  ON t.artist_id  = a.id
		GROUP BY a.id, a.name
		ORDER BY lower(a.sort_name), lower(a.name), a.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []catalog.ArtistStatItem
	for rows.Next() {
		var item catalog.ArtistStatItem
		if err := rows.Scan(&item.ArtistID, &item.Name, &item.AlbumCount, &item.TrackCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		items = []catalog.ArtistStatItem{}
	}
	return items, nil
}

func (r *Repository) AlbumTrackCounts(ctx context.Context) ([]catalog.AlbumStatItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT al.id, al.title, al.artist_id,
		       COUNT(t.id) AS track_count
		FROM albums al
		LEFT JOIN tracks t ON t.album_id = al.id
		GROUP BY al.id, al.title, al.artist_id
		ORDER BY lower(al.sort_title), lower(al.title), al.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []catalog.AlbumStatItem
	for rows.Next() {
		var item catalog.AlbumStatItem
		if err := rows.Scan(&item.AlbumID, &item.Title, &item.ArtistID, &item.TrackCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		items = []catalog.AlbumStatItem{}
	}
	return items, nil
}

func (r *Repository) PlaylistTrackCounts(ctx context.Context) ([]catalog.PlaylistStatItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.name,
		       COUNT(pt.track_id) AS track_count
		FROM playlists p
		LEFT JOIN playlist_tracks pt ON pt.playlist_id = p.id
		GROUP BY p.id, p.name
		ORDER BY lower(p.name), p.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []catalog.PlaylistStatItem
	for rows.Next() {
		var item catalog.PlaylistStatItem
		if err := rows.Scan(&item.PlaylistID, &item.Name, &item.TrackCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		items = []catalog.PlaylistStatItem{}
	}
	return items, nil
}

// ---- Recent timeline methods using UNION ALL + ORDER BY + LIMIT ----

// recentAddedSQL builds a UNION ALL query that returns the top-N most recently
// created items across the requested kinds, tagged with their entity data as
// JSON so a single query suffices.
// Each branch returns: (kind text, ts timestamptz, id text, json_payload text)
func recentAddedSQL(kind string, limit int) (string, []any) {
	branches := []string{}
	if kind == "" || kind == "artist" {
		branches = append(branches, `
			SELECT 'artist' AS kind, created_at AS ts, id, name, '' AS sort_title, '' AS artist_id,
			       0 AS release_year, '' AS sort_name, 0 AS track_number, 0 AS disc_number, 0 AS duration_ms,
			       '' AS album_id, '' AS media_object_id, '' AS description, ARRAY[]::text[] AS track_ids, updated_at
			FROM artists`)
	}
	if kind == "" || kind == "album" {
		branches = append(branches, `
			SELECT 'album' AS kind, created_at AS ts, id, title AS name, sort_title, artist_id,
			       release_year, '' AS sort_name, 0, 0, 0, '' AS album_id, '' AS media_object_id, '' AS description, ARRAY[]::text[], updated_at
			FROM albums`)
	}
	if kind == "" || kind == "track" {
		branches = append(branches, `
			SELECT 'track' AS kind, created_at AS ts, id, title AS name, sort_title, artist_id,
			       0 AS release_year, '' AS sort_name, track_number, disc_number, duration_ms,
			       COALESCE(album_id,'') AS album_id, media_object_id, '' AS description, ARRAY[]::text[], updated_at
			FROM tracks`)
	}
	if kind == "" || kind == "playlist" {
		branches = append(branches, `
			SELECT 'playlist' AS kind, p.created_at AS ts, p.id, p.name,
			       '' AS sort_title, '' AS artist_id, 0, '' AS sort_name, 0, 0, 0, '', '' AS media_object_id,
			       p.description,
			       COALESCE(ARRAY(SELECT pt.track_id FROM playlist_tracks pt WHERE pt.playlist_id = p.id ORDER BY pt.position), ARRAY[]::text[]) AS track_ids,
			       p.updated_at
			FROM playlists p`)
	}
	union := strings.Join(branches, " UNION ALL ")
	sql := fmt.Sprintf("SELECT * FROM (%s) combined ORDER BY ts DESC, id DESC LIMIT $1", union)
	return sql, []any{limit}
}

func (r *Repository) RecentlyAdded(ctx context.Context, kind string, limit int) ([]catalog.RecentCatalogItem, error) {
	sql, args := recentAddedSQL(kind, limit)
	return r.queryRecentRows(ctx, sql, args, true)
}

func (r *Repository) RecentlyUpdated(ctx context.Context, kind string, limit int) ([]catalog.UpdatedCatalogItem, error) {
	// For recently-updated we sort on updated_at instead of created_at.
	// We reuse the same column layout but substitute updated_at in the ts slot.
	branches := []string{}
	if kind == "" || kind == "artist" {
		branches = append(branches, `
			SELECT 'artist', updated_at AS ts, id, name, '' AS sort_title, '' AS artist_id,
			       0, '' AS sort_name, 0, 0, 0, '', '', '' AS description, ARRAY[]::text[]::text[], created_at
			FROM artists`)
	}
	if kind == "" || kind == "album" {
		branches = append(branches, `
			SELECT 'album', updated_at AS ts, id, title, sort_title, artist_id,
			       release_year, '', 0, 0, 0, '', '', '', ARRAY[]::text[], created_at
			FROM albums`)
	}
	if kind == "" || kind == "track" {
		branches = append(branches, `
			SELECT 'track', updated_at AS ts, id, title, sort_title, artist_id,
			       0, '', track_number, disc_number, duration_ms,
			       COALESCE(album_id,''), media_object_id, '', ARRAY[]::text[], created_at
			FROM tracks`)
	}
	if kind == "" || kind == "playlist" {
		branches = append(branches, `
			SELECT 'playlist', p.updated_at AS ts, p.id, p.name,
			       '', '', 0, '', 0, 0, 0, '', '', p.description,
			       COALESCE(ARRAY(SELECT pt.track_id FROM playlist_tracks pt WHERE pt.playlist_id = p.id ORDER BY pt.position), ARRAY[]::text[]),
			       p.created_at
			FROM playlists p`)
	}
	union := strings.Join(branches, " UNION ALL ")
	sql := fmt.Sprintf("SELECT * FROM (%s) combined ORDER BY ts DESC, id DESC LIMIT $1", union)
	rows, err := r.pool.Query(ctx, sql, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []catalog.UpdatedCatalogItem
	for rows.Next() {
		var kindStr string
		var ts, createdAt interface{}
		var id, name, sortTitle, artistID, sortName, albumID, mediaObjectID, description string
		var releaseYear, trackNumber, discNumber, durationMS int
		var trackIDs []string
		if err := rows.Scan(&kindStr, &ts, &id, &name, &sortTitle, &artistID,
			&releaseYear, &sortName, &trackNumber, &discNumber, &durationMS,
			&albumID, &mediaObjectID, &description, &trackIDs, &createdAt); err != nil {
			return nil, err
		}
		updatedAt := tsTime(ts)
		item := catalog.UpdatedCatalogItem{UpdatedAt: updatedAt}
		switch kindStr {
		case "artist":
			item.Kind = catalog.RecentItemArtist
			a := catalog.Artist{ID: id, Name: name, SortName: sortName, CreatedAt: tsTime(createdAt), UpdatedAt: updatedAt}
			item.Artist = &a
		case "album":
			item.Kind = catalog.RecentItemAlbum
			al := catalog.Album{ID: id, Title: name, SortTitle: sortTitle, ArtistID: artistID, ReleaseYear: releaseYear, CreatedAt: tsTime(createdAt), UpdatedAt: updatedAt}
			item.Album = &al
		case "track":
			item.Kind = catalog.RecentItemTrack
			t := catalog.Track{ID: id, Title: name, SortTitle: sortTitle, ArtistID: artistID, AlbumID: albumID, MediaObjectID: mediaObjectID, TrackNumber: trackNumber, DiscNumber: discNumber, DurationMS: durationMS, CreatedAt: tsTime(createdAt), UpdatedAt: updatedAt}
			item.Track = &t
		case "playlist":
			item.Kind = catalog.RecentItemPlaylist
			if trackIDs == nil {
				trackIDs = []string{}
			}
			p := catalog.Playlist{ID: id, Name: name, Description: description, TrackIDs: trackIDs, CreatedAt: tsTime(createdAt), UpdatedAt: updatedAt}
			item.Playlist = &p
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		items = []catalog.UpdatedCatalogItem{}
	}
	return items, nil
}

func (r *Repository) queryRecentRows(ctx context.Context, sql string, args []any, useCreatedAt bool) ([]catalog.RecentCatalogItem, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []catalog.RecentCatalogItem
	for rows.Next() {
		var kindStr string
		var ts, updatedAt interface{}
		var id, name, sortTitle, artistID, sortName, albumID, mediaObjectID, description string
		var releaseYear, trackNumber, discNumber, durationMS int
		var trackIDs []string
		if err := rows.Scan(&kindStr, &ts, &id, &name, &sortTitle, &artistID,
			&releaseYear, &sortName, &trackNumber, &discNumber, &durationMS,
			&albumID, &mediaObjectID, &description, &trackIDs, &updatedAt); err != nil {
			return nil, err
		}
		addedAt := tsTime(ts)
		item := catalog.RecentCatalogItem{AddedAt: addedAt}
		switch kindStr {
		case "artist":
			item.Kind = catalog.RecentItemArtist
			a := catalog.Artist{ID: id, Name: name, SortName: sortName, CreatedAt: addedAt, UpdatedAt: tsTime(updatedAt)}
			item.Artist = &a
		case "album":
			item.Kind = catalog.RecentItemAlbum
			al := catalog.Album{ID: id, Title: name, SortTitle: sortTitle, ArtistID: artistID, ReleaseYear: releaseYear, CreatedAt: addedAt, UpdatedAt: tsTime(updatedAt)}
			item.Album = &al
		case "track":
			item.Kind = catalog.RecentItemTrack
			t := catalog.Track{ID: id, Title: name, SortTitle: sortTitle, ArtistID: artistID, AlbumID: albumID, MediaObjectID: mediaObjectID, TrackNumber: trackNumber, DiscNumber: discNumber, DurationMS: durationMS, CreatedAt: addedAt, UpdatedAt: tsTime(updatedAt)}
			item.Track = &t
		case "playlist":
			item.Kind = catalog.RecentItemPlaylist
			if trackIDs == nil {
				trackIDs = []string{}
			}
			p := catalog.Playlist{ID: id, Name: name, Description: description, TrackIDs: trackIDs, CreatedAt: addedAt, UpdatedAt: tsTime(updatedAt)}
			item.Playlist = &p
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		items = []catalog.RecentCatalogItem{}
	}
	return items, nil
}

func tsTime(v interface{}) (t time.Time) {
	if v == nil {
		return
	}
	switch val := v.(type) {
	case time.Time:
		return val
	}
	return
}

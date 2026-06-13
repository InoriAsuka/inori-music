package catalogpg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/catalog"
)

// Repository implements catalog.Repository using a PostgreSQL connection pool.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveArtist(ctx context.Context, artist catalog.Artist) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO artists (id, name, sort_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
		    name       = EXCLUDED.name,
		    sort_name  = EXCLUDED.sort_name,
		    updated_at = EXCLUDED.updated_at`,
		artist.ID, artist.Name, artist.SortName, artist.CreatedAt.UTC(), artist.UpdatedAt.UTC(),
	)
	return err
}

func (r *Repository) GetArtist(ctx context.Context, id string) (catalog.Artist, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, name, sort_name, created_at, updated_at FROM artists WHERE id = $1`, id)
	artist, err := scanArtist(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return catalog.Artist{}, fmt.Errorf("%w: %s", catalog.ErrArtistNotFound, id)
		}
		return catalog.Artist{}, err
	}
	return artist, nil
}

func (r *Repository) ListArtists(ctx context.Context) ([]catalog.Artist, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, sort_name, created_at, updated_at FROM artists ORDER BY lower(sort_name), lower(name), id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var artists []catalog.Artist
	for rows.Next() {
		artist, err := scanArtist(rows)
		if err != nil {
			return nil, err
		}
		artists = append(artists, artist)
	}
	return artists, rows.Err()
}

func (r *Repository) DeleteArtist(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM artists WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", catalog.ErrArtistNotFound, id)
	}
	return nil
}

func (r *Repository) SaveAlbum(ctx context.Context, album catalog.Album) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO albums (id, title, sort_title, artist_id, release_year, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
		    title        = EXCLUDED.title,
		    sort_title   = EXCLUDED.sort_title,
		    artist_id    = EXCLUDED.artist_id,
		    release_year = EXCLUDED.release_year,
		    updated_at   = EXCLUDED.updated_at`,
		album.ID, album.Title, album.SortTitle, album.ArtistID, album.ReleaseYear, album.CreatedAt.UTC(), album.UpdatedAt.UTC(),
	)
	return err
}

func (r *Repository) GetAlbum(ctx context.Context, id string) (catalog.Album, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, title, sort_title, artist_id, release_year, created_at, updated_at FROM albums WHERE id = $1`, id)
	album, err := scanAlbum(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return catalog.Album{}, fmt.Errorf("%w: %s", catalog.ErrAlbumNotFound, id)
		}
		return catalog.Album{}, err
	}
	return album, nil
}

func (r *Repository) ListAlbums(ctx context.Context) ([]catalog.Album, error) {
	return r.queryAlbums(ctx, `SELECT id, title, sort_title, artist_id, release_year, created_at, updated_at FROM albums ORDER BY lower(sort_title), lower(title), id`)
}

func (r *Repository) ListAlbumsByArtist(ctx context.Context, artistID string) ([]catalog.Album, error) {
	return r.queryAlbums(ctx, `SELECT id, title, sort_title, artist_id, release_year, created_at, updated_at FROM albums WHERE artist_id = $1 ORDER BY release_year, lower(sort_title), lower(title), id`, artistID)
}

func (r *Repository) queryAlbums(ctx context.Context, sql string, args ...any) ([]catalog.Album, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var albums []catalog.Album
	for rows.Next() {
		album, err := scanAlbum(rows)
		if err != nil {
			return nil, err
		}
		albums = append(albums, album)
	}
	return albums, rows.Err()
}

func (r *Repository) DeleteAlbum(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM albums WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", catalog.ErrAlbumNotFound, id)
	}
	return nil
}

func (r *Repository) SaveTrack(ctx context.Context, track catalog.Track) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tracks (id, title, sort_title, artist_id, album_id, media_object_id, track_number, disc_number, duration_ms, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
		    title           = EXCLUDED.title,
		    sort_title      = EXCLUDED.sort_title,
		    artist_id       = EXCLUDED.artist_id,
		    album_id        = EXCLUDED.album_id,
		    media_object_id = EXCLUDED.media_object_id,
		    track_number    = EXCLUDED.track_number,
		    disc_number     = EXCLUDED.disc_number,
		    duration_ms     = EXCLUDED.duration_ms,
		    updated_at      = EXCLUDED.updated_at`,
		track.ID, track.Title, track.SortTitle, track.ArtistID, track.AlbumID, track.MediaObjectID, track.TrackNumber, track.DiscNumber, track.DurationMS, track.CreatedAt.UTC(), track.UpdatedAt.UTC(),
	)
	return err
}

func (r *Repository) GetTrack(ctx context.Context, id string) (catalog.Track, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, title, sort_title, artist_id, COALESCE(album_id, ''), media_object_id, track_number, disc_number, duration_ms, created_at, updated_at FROM tracks WHERE id = $1`, id)
	track, err := scanTrack(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return catalog.Track{}, fmt.Errorf("%w: %s", catalog.ErrTrackNotFound, id)
		}
		return catalog.Track{}, err
	}
	return track, nil
}

func (r *Repository) ListTracks(ctx context.Context) ([]catalog.Track, error) {
	return r.queryTracks(ctx, `SELECT id, title, sort_title, artist_id, COALESCE(album_id, ''), media_object_id, track_number, disc_number, duration_ms, created_at, updated_at FROM tracks ORDER BY lower(sort_title), lower(title), id`)
}

func (r *Repository) ListTracksByAlbum(ctx context.Context, albumID string) ([]catalog.Track, error) {
	return r.queryTracks(ctx, `SELECT id, title, sort_title, artist_id, COALESCE(album_id, ''), media_object_id, track_number, disc_number, duration_ms, created_at, updated_at FROM tracks WHERE album_id = $1 ORDER BY disc_number, track_number, lower(title), id`, albumID)
}

func (r *Repository) ListTracksByArtist(ctx context.Context, artistID string) ([]catalog.Track, error) {
	return r.queryTracks(ctx, `SELECT id, title, sort_title, artist_id, COALESCE(album_id, ''), media_object_id, track_number, disc_number, duration_ms, created_at, updated_at FROM tracks WHERE artist_id = $1 ORDER BY lower(sort_title), lower(title), id`, artistID)
}

func (r *Repository) queryTracks(ctx context.Context, sql string, args ...any) ([]catalog.Track, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tracks []catalog.Track
	for rows.Next() {
		track, err := scanTrack(rows)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	return tracks, rows.Err()
}

func (r *Repository) DeleteTrack(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM tracks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", catalog.ErrTrackNotFound, id)
	}
	return nil
}

// SearchCatalog uses PostgreSQL full-text search (tsvector/GIN indexes from migration
// 006_catalog_fts) to find artists, albums, and tracks matching the query.
// Results are ordered: artists first (rank desc), then albums, then tracks.
func (r *Repository) SearchCatalog(ctx context.Context, query string) (catalog.CatalogSearchResult, error) {
	result := catalog.CatalogSearchResult{Query: query}

	// Artists
	artistRows, err := r.pool.Query(ctx, `
		SELECT id, name, sort_name, created_at, updated_at
		FROM artists
		WHERE search_vector @@ plainto_tsquery('simple', $1)
		ORDER BY ts_rank(search_vector, plainto_tsquery('simple', $1)) DESC, lower(sort_name), lower(name), id`,
		query)
	if err != nil {
		return result, fmt.Errorf("search artists: %w", err)
	}
	defer artistRows.Close()
	for artistRows.Next() {
		a, err := scanArtist(artistRows)
		if err != nil {
			return result, err
		}
		ac := a
		result.Items = append(result.Items, catalog.SearchResultItem{Kind: catalog.SearchResultArtist, Artist: &ac})
	}
	if err := artistRows.Err(); err != nil {
		return result, err
	}

	// Albums
	albumRows, err := r.pool.Query(ctx, `
		SELECT id, title, sort_title, artist_id, release_year, created_at, updated_at
		FROM albums
		WHERE search_vector @@ plainto_tsquery('simple', $1)
		ORDER BY ts_rank(search_vector, plainto_tsquery('simple', $1)) DESC, release_year, lower(sort_title), lower(title), id`,
		query)
	if err != nil {
		return result, fmt.Errorf("search albums: %w", err)
	}
	defer albumRows.Close()
	for albumRows.Next() {
		a, err := scanAlbum(albumRows)
		if err != nil {
			return result, err
		}
		ac := a
		result.Items = append(result.Items, catalog.SearchResultItem{Kind: catalog.SearchResultAlbum, Album: &ac})
	}
	if err := albumRows.Err(); err != nil {
		return result, err
	}

	// Tracks
	trackRows, err := r.pool.Query(ctx, `
		SELECT id, title, sort_title, artist_id, COALESCE(album_id, ''), media_object_id, track_number, disc_number, duration_ms, created_at, updated_at
		FROM tracks
		WHERE search_vector @@ plainto_tsquery('simple', $1)
		ORDER BY ts_rank(search_vector, plainto_tsquery('simple', $1)) DESC, lower(sort_title), lower(title), id`,
		query)
	if err != nil {
		return result, fmt.Errorf("search tracks: %w", err)
	}
	defer trackRows.Close()
	for trackRows.Next() {
		t, err := scanTrack(trackRows)
		if err != nil {
			return result, err
		}
		tc := t
		result.Items = append(result.Items, catalog.SearchResultItem{Kind: catalog.SearchResultTrack, Track: &tc})
	}
	if err := trackRows.Err(); err != nil {
		return result, err
	}

	return result, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanArtist(s scanner) (catalog.Artist, error) {
	var a catalog.Artist
	if err := s.Scan(&a.ID, &a.Name, &a.SortName, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return catalog.Artist{}, err
	}
	return a, nil
}

func scanAlbum(s scanner) (catalog.Album, error) {
	var a catalog.Album
	if err := s.Scan(&a.ID, &a.Title, &a.SortTitle, &a.ArtistID, &a.ReleaseYear, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return catalog.Album{}, err
	}
	return a, nil
}

func scanTrack(s scanner) (catalog.Track, error) {
	var t catalog.Track
	if err := s.Scan(&t.ID, &t.Title, &t.SortTitle, &t.ArtistID, &t.AlbumID, &t.MediaObjectID, &t.TrackNumber, &t.DiscNumber, &t.DurationMS, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return catalog.Track{}, err
	}
	return t, nil
}

// SavePlaylist upserts a playlist record and replaces its ordered track list.
func (r *Repository) SavePlaylist(ctx context.Context, p catalog.Playlist) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx, `
		INSERT INTO playlists (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
		    name        = EXCLUDED.name,
		    description = EXCLUDED.description,
		    updated_at  = EXCLUDED.updated_at`,
		p.ID, p.Name, p.Description, p.CreatedAt.UTC(), p.UpdatedAt.UTC(),
	)
	if err != nil {
		return err
	}
	// Replace the ordered track list atomically.
	if _, err = tx.Exec(ctx, `DELETE FROM playlist_tracks WHERE playlist_id = $1`, p.ID); err != nil {
		return err
	}
	for pos, trackID := range p.TrackIDs {
		if _, err = tx.Exec(ctx,
			`INSERT INTO playlist_tracks (playlist_id, track_id, position) VALUES ($1, $2, $3)`,
			p.ID, trackID, pos,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) GetPlaylist(ctx context.Context, id string) (catalog.Playlist, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, name, description, created_at, updated_at FROM playlists WHERE id = $1`, id)
	var p catalog.Playlist
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return catalog.Playlist{}, fmt.Errorf("%w: %s", catalog.ErrPlaylistNotFound, id)
		}
		return catalog.Playlist{}, err
	}
	trackIDs, err := r.loadPlaylistTracks(ctx, id)
	if err != nil {
		return catalog.Playlist{}, err
	}
	p.TrackIDs = trackIDs
	return p, nil
}

func (r *Repository) ListPlaylists(ctx context.Context) ([]catalog.Playlist, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, description, created_at, updated_at FROM playlists ORDER BY lower(name), id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var playlists []catalog.Playlist
	for rows.Next() {
		var p catalog.Playlist
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		playlists = append(playlists, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Load track lists for each playlist.
	for i := range playlists {
		trackIDs, err := r.loadPlaylistTracks(ctx, playlists[i].ID)
		if err != nil {
			return nil, err
		}
		playlists[i].TrackIDs = trackIDs
	}
	return playlists, nil
}

func (r *Repository) DeletePlaylist(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM playlists WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", catalog.ErrPlaylistNotFound, id)
	}
	return nil
}

// loadPlaylistTracks returns the ordered list of track IDs for a playlist.
func (r *Repository) loadPlaylistTracks(ctx context.Context, playlistID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT track_id FROM playlist_tracks WHERE playlist_id = $1 ORDER BY position`, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return nil, err
		}
		ids = append(ids, tid)
	}
	if ids == nil {
		ids = []string{}
	}
	return ids, rows.Err()
}

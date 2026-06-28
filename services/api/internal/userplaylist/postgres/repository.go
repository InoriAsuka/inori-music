package userplaylistpg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/userplaylist"
)

// Repository implements userplaylist.Repository using PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Migrate creates the required tables if they do not exist.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user_playlists (
			id          TEXT PRIMARY KEY,
			user_id     TEXT NOT NULL,
			name        TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at  TIMESTAMPTZ NOT NULL,
			updated_at  TIMESTAMPTZ NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_user_playlists_user_id ON user_playlists(user_id);

		CREATE TABLE IF NOT EXISTS user_playlist_tracks (
			playlist_id TEXT NOT NULL REFERENCES user_playlists(id) ON DELETE CASCADE,
			position    INT  NOT NULL,
			track_id    TEXT NOT NULL,
			PRIMARY KEY (playlist_id, position)
		);
	`)
	return err
}

// Save upserts a playlist and replaces its track list atomically.
func (r *Repository) Save(ctx context.Context, p userplaylist.UserPlaylist) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx, `
		INSERT INTO user_playlists (id, user_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			name        = EXCLUDED.name,
			description = EXCLUDED.description,
			updated_at  = EXCLUDED.updated_at`,
		p.ID, p.UserID, p.Name, p.Description, p.CreatedAt.UTC(), p.UpdatedAt.UTC(),
	)
	if err != nil {
		return err
	}

	// Replace tracks: delete all then insert.
	if _, err = tx.Exec(ctx, `DELETE FROM user_playlist_tracks WHERE playlist_id = $1`, p.ID); err != nil {
		return err
	}
	for i, tid := range p.TrackIDs {
		if _, err = tx.Exec(ctx, `
			INSERT INTO user_playlist_tracks (playlist_id, position, track_id) VALUES ($1, $2, $3)`,
			p.ID, i, tid,
		); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Get retrieves a playlist and its ordered track IDs.
func (r *Repository) Get(ctx context.Context, id string) (userplaylist.UserPlaylist, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM user_playlists WHERE id = $1`, id)

	var p userplaylist.UserPlaylist
	if err := row.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return userplaylist.UserPlaylist{}, userplaylist.ErrNotFound
	}

	rows, err := r.pool.Query(ctx, `
		SELECT track_id FROM user_playlist_tracks
		WHERE playlist_id = $1 ORDER BY position ASC`, id)
	if err != nil {
		return userplaylist.UserPlaylist{}, err
	}
	defer rows.Close()

	p.TrackIDs = []string{}
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return userplaylist.UserPlaylist{}, err
		}
		p.TrackIDs = append(p.TrackIDs, tid)
	}
	return p, rows.Err()
}

// ListByUser returns all playlists for a user, newest first.
func (r *Repository) ListByUser(ctx context.Context, userID string) ([]userplaylist.UserPlaylist, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM user_playlists WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []userplaylist.UserPlaylist
	for rows.Next() {
		var p userplaylist.UserPlaylist
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		playlists = append(playlists, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if playlists == nil {
		playlists = []userplaylist.UserPlaylist{}
	}

	// Fetch track IDs for each playlist.
	for i, pl := range playlists {
		trows, err := r.pool.Query(ctx, `
			SELECT track_id FROM user_playlist_tracks
			WHERE playlist_id = $1 ORDER BY position ASC`, pl.ID)
		if err != nil {
			return nil, err
		}
		playlists[i].TrackIDs = []string{}
		for trows.Next() {
			var tid string
			if err := trows.Scan(&tid); err != nil {
				trows.Close()
				return nil, err
			}
			playlists[i].TrackIDs = append(playlists[i].TrackIDs, tid)
		}
		trows.Close()
		if err := trows.Err(); err != nil {
			return nil, err
		}
	}
	return playlists, nil
}

// Delete removes a playlist (cascade removes tracks via FK).
func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_playlists WHERE id = $1`, id)
	return err
}

package searchhistorypg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/searchhistory"
)

// Repository implements searchhistory.Repository using a PostgreSQL connection pool.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Repository backed by the given pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Migrate creates the user_search_history table if it does not exist. Each
// (user_id, query) pair is unique; re-searching a query refreshes its timestamp.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user_search_history (
			user_id     TEXT NOT NULL,
			query       TEXT NOT NULL,
			searched_at TIMESTAMPTZ NOT NULL,
			PRIMARY KEY (user_id, query)
		);
		CREATE INDEX IF NOT EXISTS idx_user_search_history_user_time
			ON user_search_history(user_id, searched_at DESC);
	`)
	return err
}

func (r *Repository) List(ctx context.Context, userID string) ([]searchhistory.Entry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT query, searched_at FROM user_search_history
		WHERE user_id = $1
		ORDER BY searched_at DESC, query ASC
		LIMIT $2`, userID, searchhistory.MaxEntries)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]searchhistory.Entry, 0, searchhistory.MaxEntries)
	for rows.Next() {
		var e searchhistory.Entry
		if err := rows.Scan(&e.Query, &e.SearchedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// Replace overwrites the user's entire search history atomically.
func (r *Repository) Replace(ctx context.Context, userID string, entries []searchhistory.Entry) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err = tx.Exec(ctx, `DELETE FROM user_search_history WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, e := range entries {
		if _, err = tx.Exec(ctx, `
			INSERT INTO user_search_history (user_id, query, searched_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, query) DO UPDATE SET searched_at = EXCLUDED.searched_at`,
			userID, e.Query, e.SearchedAt.UTC(),
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) Clear(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_search_history WHERE user_id = $1`, userID)
	return err
}

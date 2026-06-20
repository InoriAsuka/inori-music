package favoritespg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/favorites"
)

// Repository implements favorites.Repository using a PostgreSQL connection pool.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Repository backed by the given pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) AddFavorite(ctx context.Context, userID, trackID string, now time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_track_favorites (user_id, track_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, track_id) DO NOTHING`,
		userID, trackID, now.UTC(),
	)
	return err
}

func (r *Repository) RemoveFavorite(ctx context.Context, userID, trackID string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_track_favorites WHERE user_id = $1 AND track_id = $2`,
		userID, trackID,
	)
	return err
}

func (r *Repository) ListFavorites(ctx context.Context, userID string, limit, offset int) (favorites.FavoritesPage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT track_id, COUNT(*) OVER () AS total_count
		FROM user_track_favorites
		WHERE user_id = $1
		ORDER BY created_at DESC, track_id ASC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return favorites.FavoritesPage{TrackIDs: []string{}}, err
	}
	defer rows.Close()

	var ids []string
	total := 0
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid, &total); err != nil {
			return favorites.FavoritesPage{}, err
		}
		ids = append(ids, tid)
	}
	if err := rows.Err(); err != nil {
		return favorites.FavoritesPage{}, err
	}
	if ids == nil {
		ids = []string{}
	}
	return favorites.FavoritesPage{TrackIDs: ids, Total: total}, nil
}

func (r *Repository) IsFavorite(ctx context.Context, userID, trackID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_track_favorites WHERE user_id = $1 AND track_id = $2)`,
		userID, trackID,
	).Scan(&exists)
	return exists, err
}

func (r *Repository) AreFavorites(ctx context.Context, userID string, trackIDs []string) (map[string]bool, error) {
	result := make(map[string]bool, len(trackIDs))
	for _, tid := range trackIDs {
		result[tid] = false
	}
	rows, err := r.pool.Query(ctx, `
		SELECT track_id FROM user_track_favorites
		WHERE user_id = $1 AND track_id = ANY($2)`,
		userID, trackIDs,
	)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return result, err
		}
		result[tid] = true
	}
	return result, rows.Err()
}

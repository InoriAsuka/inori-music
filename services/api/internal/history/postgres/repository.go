package historypg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/history"
)

// Repository implements history.Repository using a PostgreSQL connection pool.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SavePlayEvent(ctx context.Context, e history.PlayEvent) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO play_events (id, user_id, track_id, played_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING`,
		e.ID, e.UserID, e.TrackID, e.PlayedAt.UTC(), e.CreatedAt.UTC(),
	)
	return err
}

func (r *Repository) ListPlayEvents(ctx context.Context, f history.PlayEventFilter) ([]history.PlayEvent, int, error) {
	if f.UserID == "" {
		return nil, 0, fmt.Errorf("userID is required")
	}

	var rows pgx.Rows
	var err error
	if f.TrackID != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, user_id, track_id, played_at, created_at,
			       COUNT(*) OVER () AS total_count
			FROM play_events
			WHERE user_id = $3 AND track_id = $4
			ORDER BY played_at DESC, id DESC
			LIMIT $1 OFFSET $2`,
			f.Limit, f.Offset, f.UserID, f.TrackID)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, user_id, track_id, played_at, created_at,
			       COUNT(*) OVER () AS total_count
			FROM play_events
			WHERE user_id = $3
			ORDER BY played_at DESC, id DESC
			LIMIT $1 OFFSET $2`,
			f.Limit, f.Offset, f.UserID)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []history.PlayEvent{}, 0, nil
		}
		return nil, 0, err
	}
	defer rows.Close()

	var events []history.PlayEvent
	total := 0
	for rows.Next() {
		var e history.PlayEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.TrackID, &e.PlayedAt, &e.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if events == nil {
		events = []history.PlayEvent{}
	}
	return events, total, nil
}

func (r *Repository) DeletePlayEventsByUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM play_events WHERE user_id = $1`, userID)
	return err
}

func (r *Repository) HistoryStats(ctx context.Context, f history.StatsFilter) (history.HistoryStats, error) {
	var row pgx.Row
	if !f.Since.IsZero() {
		row = r.pool.QueryRow(ctx, `
			SELECT
				COUNT(*)                    AS total_events,
				COUNT(DISTINCT user_id)     AS unique_users,
				COUNT(DISTINCT track_id)    AS unique_tracks
			FROM play_events
			WHERE played_at >= $1`, f.Since.UTC())
	} else {
		row = r.pool.QueryRow(ctx, `
			SELECT
				COUNT(*)                    AS total_events,
				COUNT(DISTINCT user_id)     AS unique_users,
				COUNT(DISTINCT track_id)    AS unique_tracks
			FROM play_events`)
	}
	var s history.HistoryStats
	if err := row.Scan(&s.TotalEvents, &s.UniqueUsers, &s.UniqueTracks); err != nil {
		return history.HistoryStats{}, err
	}
	return s, nil
}

func (r *Repository) TopTracks(ctx context.Context, f history.StatsFilter, limit int) ([]history.TrackPlayCount, error) {
	if limit <= 0 {
		limit = 10
	}
	var rows pgx.Rows
	var err error
	if !f.Since.IsZero() {
		rows, err = r.pool.Query(ctx, `
			SELECT track_id, COUNT(*) AS play_count
			FROM play_events
			WHERE played_at >= $2
			GROUP BY track_id
			ORDER BY play_count DESC, track_id ASC
			LIMIT $1`, limit, f.Since.UTC())
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT track_id, COUNT(*) AS play_count
			FROM play_events
			GROUP BY track_id
			ORDER BY play_count DESC, track_id ASC
			LIMIT $1`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []history.TrackPlayCount
	for rows.Next() {
		var item history.TrackPlayCount
		if err := rows.Scan(&item.TrackID, &item.PlayCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []history.TrackPlayCount{}
	}
	return result, nil
}

func (r *Repository) TopUsers(ctx context.Context, f history.StatsFilter, limit int) ([]history.UserPlayCount, error) {
	if limit <= 0 {
		limit = 10
	}
	var rows pgx.Rows
	var err error
	if !f.Since.IsZero() {
		rows, err = r.pool.Query(ctx, `
			SELECT user_id, COUNT(*) AS play_count
			FROM play_events
			WHERE played_at >= $2
			GROUP BY user_id
			ORDER BY play_count DESC, user_id ASC
			LIMIT $1`, limit, f.Since.UTC())
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT user_id, COUNT(*) AS play_count
			FROM play_events
			GROUP BY user_id
			ORDER BY play_count DESC, user_id ASC
			LIMIT $1`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []history.UserPlayCount
	for rows.Next() {
		var item history.UserPlayCount
		if err := rows.Scan(&item.UserID, &item.PlayCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []history.UserPlayCount{}
	}
	return result, nil
}

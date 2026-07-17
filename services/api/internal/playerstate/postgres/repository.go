package playerstatepg

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/playerstate"
)

// Repository implements playerstate.Repository using a PostgreSQL connection pool.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Repository backed by the given pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Migrate creates the user_player_state table if it does not exist. There is a
// single row per user (user_id primary key); writes upsert last-write-wins.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user_player_state (
			user_id          TEXT PRIMARY KEY,
			queue            JSONB NOT NULL DEFAULT '[]'::jsonb,
			current_index    INT NOT NULL DEFAULT 0,
			position_seconds DOUBLE PRECISION NOT NULL DEFAULT 0,
			repeat_mode      TEXT NOT NULL DEFAULT '',
			shuffle          BOOLEAN NOT NULL DEFAULT FALSE,
			volume           DOUBLE PRECISION NOT NULL DEFAULT 1,
			speed            DOUBLE PRECISION NOT NULL DEFAULT 1,
			status           TEXT NOT NULL DEFAULT '',
			updated_at       TIMESTAMPTZ NOT NULL
		);
	`)
	return err
}

func (r *Repository) Get(ctx context.Context, userID string) (playerstate.PlayerState, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT queue, current_index, position_seconds, repeat_mode, shuffle, volume, speed, status, updated_at
		FROM user_player_state WHERE user_id = $1`, userID)

	var (
		state    playerstate.PlayerState
		queueRaw []byte
	)
	if err := row.Scan(&queueRaw, &state.CurrentIndex, &state.PositionSeconds, &state.Repeat,
		&state.Shuffle, &state.Volume, &state.Speed, &state.Status, &state.UpdatedAt); err != nil {
		// Map only the missing-row case to the domain sentinel; propagate all
		// other errors so callers do not conflate "never reported" with an outage.
		if errors.Is(err, pgx.ErrNoRows) {
			return playerstate.PlayerState{}, playerstate.ErrNotFound
		}
		return playerstate.PlayerState{}, err
	}

	state.Queue = []string{}
	if len(queueRaw) > 0 {
		if err := json.Unmarshal(queueRaw, &state.Queue); err != nil {
			return playerstate.PlayerState{}, err
		}
		if state.Queue == nil {
			state.Queue = []string{}
		}
	}
	return state, nil
}

func (r *Repository) Put(ctx context.Context, userID string, state playerstate.PlayerState) error {
	queue := state.Queue
	if queue == nil {
		queue = []string{}
	}
	queueRaw, err := json.Marshal(queue)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO user_player_state
			(user_id, queue, current_index, position_seconds, repeat_mode, shuffle, volume, speed, status, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id) DO UPDATE SET
			queue            = EXCLUDED.queue,
			current_index    = EXCLUDED.current_index,
			position_seconds = EXCLUDED.position_seconds,
			repeat_mode      = EXCLUDED.repeat_mode,
			shuffle          = EXCLUDED.shuffle,
			volume           = EXCLUDED.volume,
			speed            = EXCLUDED.speed,
			status           = EXCLUDED.status,
			updated_at       = EXCLUDED.updated_at`,
		userID, queueRaw, state.CurrentIndex, state.PositionSeconds, state.Repeat,
		state.Shuffle, state.Volume, state.Speed, state.Status, state.UpdatedAt.UTC(),
	)
	return err
}

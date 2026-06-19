package authpg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/auth"
)

// UserRepository implements auth.UserRepository using a PostgreSQL connection pool.
type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) SaveUser(ctx context.Context, user auth.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, role, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
		    username      = EXCLUDED.username,
		    password_hash = EXCLUDED.password_hash,
		    role          = EXCLUDED.role,
		    enabled       = EXCLUDED.enabled,
		    updated_at    = EXCLUDED.updated_at`,
		user.ID, user.Username, user.PasswordHash, string(user.Role),
		user.Enabled, user.CreatedAt.UTC(), user.UpdatedAt.UTC(),
	)
	return err
}

func (r *UserRepository) GetUser(ctx context.Context, id string) (auth.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, role, enabled, created_at, updated_at
		FROM users WHERE id = $1`, id)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.User{}, fmt.Errorf("%w: %s", auth.ErrUserNotFound, id)
		}
		return auth.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (auth.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, role, enabled, created_at, updated_at
		FROM users WHERE lower(username) = lower($1)`, username)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.User{}, fmt.Errorf("%w: %s", auth.ErrUserNotFound, username)
		}
		return auth.User{}, err
	}
	return user, nil
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]auth.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, username, password_hash, role, enabled, created_at, updated_at
		FROM users ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []auth.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", auth.ErrUserNotFound, id)
	}
	return nil
}

func (r *UserRepository) CountAdminUsers(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin' AND enabled = TRUE`).Scan(&count)
	return count, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(s scanner) (auth.User, error) {
	var u auth.User
	var role string
	err := s.Scan(&u.ID, &u.Username, &u.PasswordHash, &role, &u.Enabled, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return auth.User{}, err
	}
	u.Role = auth.Role(role)
	return u, nil
}

// SessionRepository implements auth.SessionRepository using a PostgreSQL connection pool.
type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) SaveSession(ctx context.Context, session auth.Session) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (token_hash, user_id, expires_at, created_at, revoked_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (token_hash) DO UPDATE SET
		    expires_at = EXCLUDED.expires_at,
		    revoked_at = EXCLUDED.revoked_at`,
		session.TokenHash, session.UserID,
		session.ExpiresAt.UTC(), session.CreatedAt.UTC(), session.RevokedAt,
	)
	return err
}

func (r *SessionRepository) GetSession(ctx context.Context, tokenHash string) (auth.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT token_hash, user_id, expires_at, created_at, revoked_at
		FROM sessions WHERE token_hash = $1`, tokenHash)
	var s auth.Session
	var revokedAt *time.Time
	err := row.Scan(&s.TokenHash, &s.UserID, &s.ExpiresAt, &s.CreatedAt, &revokedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.Session{}, auth.ErrSessionNotFound
		}
		return auth.Session{}, err
	}
	s.RevokedAt = revokedAt
	return s, nil
}

func (r *SessionRepository) RevokeSession(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE sessions SET revoked_at = $1 WHERE token_hash = $2 AND revoked_at IS NULL`,
		revokedAt.UTC(), tokenHash,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return auth.ErrSessionNotFound
	}
	return nil
}

func (r *SessionRepository) ListSessionsByUser(ctx context.Context, userID string) ([]auth.Session, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT token_hash, user_id, expires_at, created_at, revoked_at
		FROM sessions WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []auth.Session
	for rows.Next() {
		var s auth.Session
		var revokedAt *time.Time
		if err := rows.Scan(&s.TokenHash, &s.UserID, &s.ExpiresAt, &s.CreatedAt, &revokedAt); err != nil {
			return nil, err
		}
		s.RevokedAt = revokedAt
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *SessionRepository) RevokeAllSessionsByUser(ctx context.Context, userID string, revokedAt time.Time) (int, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE sessions SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL AND expires_at > $1`,
		revokedAt.UTC(), userID,
	)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context, before time.Time) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM sessions WHERE expires_at < $1`, before.UTC())
	return err
}

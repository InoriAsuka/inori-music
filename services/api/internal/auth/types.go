package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidUser     = errors.New("invalid user")
	ErrUserNotFound    = errors.New("user not found")
	ErrUserConflict    = errors.New("user conflict")
	ErrInvalidSession  = errors.New("invalid session")
	ErrSessionNotFound = errors.New("session not found")
	ErrBadCredentials  = errors.New("bad credentials")
	ErrUserDisabled    = errors.New("user disabled")
)

// Role represents the authorization level of a user account.
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleViewer Role = "viewer"
)

// User is a server-managed account record.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// UserView is the safe public projection of a User (no password hash).
type UserView struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      Role      `json:"role"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Session is a server-issued opaque bearer token record.
// TokenHash is the SHA-256 hex digest of the plaintext token.
// The plaintext token is never stored.
type Session struct {
	TokenHash string     `json:"-"`
	UserID    string     `json:"userId"`
	ExpiresAt time.Time  `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
}

// UserRepository persists user account records.
type UserRepository interface {
	SaveUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, id string) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	ListUsers(ctx context.Context) ([]User, error)
	DeleteUser(ctx context.Context, id string) error
	CountAdminUsers(ctx context.Context) (int, error)
}

// SessionRepository persists session token records.
type SessionRepository interface {
	SaveSession(ctx context.Context, session Session) error
	GetSession(ctx context.Context, tokenHash string) (Session, error)
	RevokeSession(ctx context.Context, tokenHash string, revokedAt time.Time) error
	DeleteExpiredSessions(ctx context.Context, before time.Time) error
}

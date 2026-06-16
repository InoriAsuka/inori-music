package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,64}$`)

// ServiceConfig holds runtime-configurable auth parameters.
type ServiceConfig struct {
	SessionTTL time.Duration // default 24h
}

// Service coordinates user account and session management.
type Service struct {
	users    UserRepository
	sessions SessionRepository
	cfg      ServiceConfig
	now      func() time.Time
}

func NewService(users UserRepository, sessions SessionRepository, cfg ServiceConfig) *Service {
	if cfg.SessionTTL <= 0 {
		cfg.SessionTTL = 24 * time.Hour
	}
	return &Service{users: users, sessions: sessions, cfg: cfg, now: time.Now}
}

// EnsureInitialAdmin creates an admin user if no admin accounts exist.
// It is safe to call on every startup — it is a no-op when an admin already exists.
func (s *Service) EnsureInitialAdmin(ctx context.Context, username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil
	}
	count, err := s.users.CountAdminUsers(ctx)
	if err != nil {
		return fmt.Errorf("check admin count: %w", err)
	}
	if count > 0 {
		return nil
	}
	_, err = s.CreateUser(ctx, username, password, RoleAdmin)
	if err != nil && !errors.Is(err, ErrUserConflict) {
		return fmt.Errorf("create initial admin: %w", err)
	}
	return nil
}

// CreateUser validates, hashes the password, and persists a new user.
func (s *Service) CreateUser(ctx context.Context, username, password string, role Role) (UserView, error) {
	username = strings.TrimSpace(username)
	if !usernameRe.MatchString(username) {
		return UserView{}, fmt.Errorf("%w: username must be 3-64 characters, letters/digits/underscore only", ErrInvalidUser)
	}
	if len(password) < 8 {
		return UserView{}, fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidUser)
	}
	if role != RoleAdmin && role != RoleViewer {
		return UserView{}, fmt.Errorf("%w: role must be admin or viewer", ErrInvalidUser)
	}

	if _, err := s.users.GetUserByUsername(ctx, username); err == nil {
		return UserView{}, fmt.Errorf("%w: username %q already exists", ErrUserConflict, username)
	} else if !errors.Is(err, ErrUserNotFound) {
		return UserView{}, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return UserView{}, fmt.Errorf("hash password: %w", err)
	}

	now := s.now().UTC()
	id, err := GenerateToken()
	if err != nil {
		return UserView{}, fmt.Errorf("generate id: %w", err)
	}
	// Use first 16 hex chars as a short opaque user ID.
	id = id[:16]

	user := User{
		ID:           id,
		Username:     username,
		PasswordHash: hash,
		Role:         role,
		Enabled:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.users.SaveUser(ctx, user); err != nil {
		return UserView{}, err
	}
	return toView(user), nil
}

// Login verifies credentials and issues a new session token.
// Returns the plaintext token (stored only as a hash).
func (s *Service) Login(ctx context.Context, username, password string) (string, Session, error) {
	username = strings.TrimSpace(username)
	user, err := s.users.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Constant-time-ish: still hash to avoid short-circuit timing leak.
			CheckPassword("$2a$12$invalid", password)
			return "", Session{}, ErrBadCredentials
		}
		return "", Session{}, err
	}
	if !user.Enabled {
		return "", Session{}, ErrUserDisabled
	}
	if !CheckPassword(user.PasswordHash, password) {
		return "", Session{}, ErrBadCredentials
	}

	token, err := GenerateToken()
	if err != nil {
		return "", Session{}, fmt.Errorf("generate token: %w", err)
	}
	now := s.now().UTC()
	session := Session{
		TokenHash: HashToken(token),
		UserID:    user.ID,
		ExpiresAt: now.Add(s.cfg.SessionTTL),
		CreatedAt: now,
	}
	if err := s.sessions.SaveSession(ctx, session); err != nil {
		return "", Session{}, fmt.Errorf("save session: %w", err)
	}
	return token, session, nil
}

// ValidateToken looks up the session by token hash and returns the owning user.
// Returns ErrSessionNotFound if expired, revoked, or unknown.
func (s *Service) ValidateToken(ctx context.Context, token string) (User, error) {
	session, err := s.sessions.GetSession(ctx, HashToken(token))
	if err != nil {
		return User{}, err
	}
	if session.RevokedAt != nil {
		return User{}, ErrSessionNotFound
	}
	if s.now().UTC().After(session.ExpiresAt) {
		return User{}, ErrSessionNotFound
	}
	return s.users.GetUser(ctx, session.UserID)
}

// Logout revokes the session associated with the plaintext token.
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.sessions.RevokeSession(ctx, HashToken(token), s.now().UTC())
}

// ListUsers returns all user accounts.
func (s *Service) ListUsers(ctx context.Context) ([]UserView, error) {
	users, err := s.users.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]UserView, len(users))
	for i, u := range users {
		views[i] = toView(u)
	}
	return views, nil
}

// DisableUser marks the user as disabled, preventing future logins.
func (s *Service) DisableUser(ctx context.Context, id string) (UserView, error) {
	user, err := s.users.GetUser(ctx, id)
	if err != nil {
		return UserView{}, err
	}
	user.Enabled = false
	user.UpdatedAt = s.now().UTC()
	if err := s.users.SaveUser(ctx, user); err != nil {
		return UserView{}, err
	}
	return toView(user), nil
}

// DeleteUser removes a user record permanently.
func (s *Service) DeleteUser(ctx context.Context, id string) error {
	if _, err := s.users.GetUser(ctx, id); err != nil {
		return err
	}
	return s.users.DeleteUser(ctx, id)
}

func toView(u User) UserView {
	return UserView{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		Enabled:   u.Enabled,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryUserRepository is a development/CI repository for user account records.
// It requires no external database, mirroring the in-memory fallback pattern
// used by the storage/catalog/history/favorites/userplaylist services.
type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{users: make(map[string]User)}
}

func (r *MemoryUserRepository) SaveUser(_ context.Context, u User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}

func (r *MemoryUserRepository) GetUser(_ context.Context, id string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return User{}, fmt.Errorf("%w: %s", ErrUserNotFound, id)
	}
	return u, nil
}

func (r *MemoryUserRepository) GetUserByUsername(_ context.Context, username string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return User{}, fmt.Errorf("%w: %s", ErrUserNotFound, username)
}

func (r *MemoryUserRepository) ListUsers(_ context.Context) ([]User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]User, 0, len(r.users))
	for _, u := range r.users {
		list = append(list, u)
	}
	return list, nil
}

func (r *MemoryUserRepository) DeleteUser(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return fmt.Errorf("%w: %s", ErrUserNotFound, id)
	}
	delete(r.users, id)
	return nil
}

func (r *MemoryUserRepository) CountAdminUsers(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, u := range r.users {
		if u.Role == RoleAdmin && u.Enabled {
			n++
		}
	}
	return n, nil
}

// MemorySessionRepository is a development/CI repository for session token records.
type MemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{sessions: make(map[string]Session)}
}

func (r *MemorySessionRepository) SaveSession(_ context.Context, s Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.TokenHash] = s
	return nil
}

func (r *MemorySessionRepository) GetSession(_ context.Context, tokenHash string) (Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[tokenHash]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	return s, nil
}

func (r *MemorySessionRepository) RevokeSession(_ context.Context, tokenHash string, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[tokenHash]
	if !ok {
		return ErrSessionNotFound
	}
	t := revokedAt
	s.RevokedAt = &t
	r.sessions[tokenHash] = s
	return nil
}

func (r *MemorySessionRepository) ListSessionsByUser(_ context.Context, userID string) ([]Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []Session
	for _, s := range r.sessions {
		if s.UserID == userID {
			list = append(list, s)
		}
	}
	return list, nil
}

func (r *MemorySessionRepository) RevokeAllSessionsByUser(_ context.Context, userID string, revokedAt time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for k, s := range r.sessions {
		if s.UserID == userID && s.RevokedAt == nil {
			t := revokedAt
			s.RevokedAt = &t
			r.sessions[k] = s
			count++
		}
	}
	return count, nil
}

func (r *MemorySessionRepository) DeleteExpiredSessions(_ context.Context, before time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, s := range r.sessions {
		if s.ExpiresAt.Before(before) {
			delete(r.sessions, k)
		}
	}
	return nil
}

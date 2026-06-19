package auth_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"inori-music/services/api/internal/auth"
)

// ---- in-memory stubs ----

type memUserRepo struct {
	mu    sync.RWMutex
	users map[string]auth.User
}

func newMemUserRepo() *memUserRepo { return &memUserRepo{users: map[string]auth.User{}} }

func (r *memUserRepo) SaveUser(_ context.Context, u auth.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}
func (r *memUserRepo) GetUser(_ context.Context, id string) (auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return auth.User{}, fmt.Errorf("%w: %s", auth.ErrUserNotFound, id)
	}
	return u, nil
}
func (r *memUserRepo) GetUserByUsername(_ context.Context, username string) (auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return auth.User{}, fmt.Errorf("%w: %s", auth.ErrUserNotFound, username)
}
func (r *memUserRepo) ListUsers(_ context.Context) ([]auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]auth.User, 0, len(r.users))
	for _, u := range r.users {
		list = append(list, u)
	}
	return list, nil
}
func (r *memUserRepo) DeleteUser(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return fmt.Errorf("%w: %s", auth.ErrUserNotFound, id)
	}
	delete(r.users, id)
	return nil
}
func (r *memUserRepo) CountAdminUsers(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, u := range r.users {
		if u.Role == auth.RoleAdmin && u.Enabled {
			n++
		}
	}
	return n, nil
}

type memSessionRepo struct {
	mu       sync.RWMutex
	sessions map[string]auth.Session
}

func newMemSessionRepo() *memSessionRepo { return &memSessionRepo{sessions: map[string]auth.Session{}} }

func (r *memSessionRepo) SaveSession(_ context.Context, s auth.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.TokenHash] = s
	return nil
}
func (r *memSessionRepo) GetSession(_ context.Context, tokenHash string) (auth.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[tokenHash]
	if !ok {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	return s, nil
}
func (r *memSessionRepo) RevokeSession(_ context.Context, tokenHash string, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[tokenHash]
	if !ok {
		return auth.ErrSessionNotFound
	}
	s.RevokedAt = &revokedAt
	r.sessions[tokenHash] = s
	return nil
}
func (r *memSessionRepo) DeleteExpiredSessions(_ context.Context, before time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, s := range r.sessions {
		if s.ExpiresAt.Before(before) {
			delete(r.sessions, k)
		}
	}
	return nil
}

func newTestService(ttl time.Duration) *auth.Service {
	return auth.NewService(newMemUserRepo(), newMemSessionRepo(), auth.ServiceConfig{SessionTTL: ttl})
}

// ---- tests ----

func TestCreateUser_Valid(t *testing.T) {
	svc := newTestService(time.Hour)
	view, err := svc.CreateUser(context.Background(), "alice", "password123", auth.RoleAdmin)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if view.Username != "alice" {
		t.Errorf("Username: got %q want %q", view.Username, "alice")
	}
	if view.Role != auth.RoleAdmin {
		t.Errorf("Role: got %q want %q", view.Role, auth.RoleAdmin)
	}
	if !view.Enabled {
		t.Error("expected Enabled=true")
	}
}

func TestCreateUser_InvalidUsername(t *testing.T) {
	svc := newTestService(time.Hour)
	cases := []string{"ab", "has space", "has-dash", "toolongusernamethatexceedssixtyfourcharacterslimitforthistest123456"}
	for _, name := range cases {
		_, err := svc.CreateUser(context.Background(), name, "password123", auth.RoleAdmin)
		if !errors.Is(err, auth.ErrInvalidUser) {
			t.Errorf("username %q: expected ErrInvalidUser, got %v", name, err)
		}
	}
}

func TestCreateUser_ShortPassword(t *testing.T) {
	svc := newTestService(time.Hour)
	_, err := svc.CreateUser(context.Background(), "alice", "short", auth.RoleAdmin)
	if !errors.Is(err, auth.ErrInvalidUser) {
		t.Errorf("expected ErrInvalidUser, got %v", err)
	}
}

func TestCreateUser_Conflict(t *testing.T) {
	svc := newTestService(time.Hour)
	if _, err := svc.CreateUser(context.Background(), "alice", "password123", auth.RoleAdmin); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := svc.CreateUser(context.Background(), "alice", "password456", auth.RoleAdmin)
	if !errors.Is(err, auth.ErrUserConflict) {
		t.Errorf("expected ErrUserConflict, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newTestService(time.Hour)
	if _, err := svc.CreateUser(context.Background(), "bob", "securepass", auth.RoleAdmin); err != nil {
		t.Fatal(err)
	}
	token, session, err := svc.Login(context.Background(), "bob", "securepass")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if session.UserID == "" {
		t.Error("expected non-empty UserID in session")
	}
}

func TestLogin_BadPassword(t *testing.T) {
	svc := newTestService(time.Hour)
	if _, err := svc.CreateUser(context.Background(), "bob", "securepass", auth.RoleAdmin); err != nil {
		t.Fatal(err)
	}
	_, _, err := svc.Login(context.Background(), "bob", "wrongpass")
	if !errors.Is(err, auth.ErrBadCredentials) {
		t.Errorf("expected ErrBadCredentials, got %v", err)
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	svc := newTestService(time.Hour)
	_, _, err := svc.Login(context.Background(), "nobody", "password123")
	if !errors.Is(err, auth.ErrBadCredentials) {
		t.Errorf("expected ErrBadCredentials, got %v", err)
	}
}

func TestLogin_DisabledUser(t *testing.T) {
	svc := newTestService(time.Hour)
	view, err := svc.CreateUser(context.Background(), "bob", "securepass", auth.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DisableUser(context.Background(), view.ID); err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.Login(context.Background(), "bob", "securepass")
	if !errors.Is(err, auth.ErrUserDisabled) {
		t.Errorf("expected ErrUserDisabled, got %v", err)
	}
}

func TestValidateToken_Valid(t *testing.T) {
	svc := newTestService(time.Hour)
	if _, err := svc.CreateUser(context.Background(), "carol", "password123", auth.RoleAdmin); err != nil {
		t.Fatal(err)
	}
	token, _, err := svc.Login(context.Background(), "carol", "password123")
	if err != nil {
		t.Fatal(err)
	}
	user, err := svc.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if user.Username != "carol" {
		t.Errorf("Username: got %q want %q", user.Username, "carol")
	}
}

func TestValidateToken_ExpiredSession(t *testing.T) {
	svc := newTestService(time.Millisecond) // 1ms TTL
	if _, err := svc.CreateUser(context.Background(), "carol", "password123", auth.RoleAdmin); err != nil {
		t.Fatal(err)
	}
	token, _, err := svc.Login(context.Background(), "carol", "password123")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	_, err = svc.ValidateToken(context.Background(), token)
	if !errors.Is(err, auth.ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestLogout_RevokesSession(t *testing.T) {
	svc := newTestService(time.Hour)
	if _, err := svc.CreateUser(context.Background(), "dave", "password123", auth.RoleAdmin); err != nil {
		t.Fatal(err)
	}
	token, _, err := svc.Login(context.Background(), "dave", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(context.Background(), token); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	_, err = svc.ValidateToken(context.Background(), token)
	if !errors.Is(err, auth.ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound after logout, got %v", err)
	}
}

func TestEnsureInitialAdmin_CreatesWhenEmpty(t *testing.T) {
	svc := newTestService(time.Hour)
	if err := svc.EnsureInitialAdmin(context.Background(), "admin", "adminpass"); err != nil {
		t.Fatalf("EnsureInitialAdmin: %v", err)
	}
	_, _, err := svc.Login(context.Background(), "admin", "adminpass")
	if err != nil {
		t.Fatalf("Login after EnsureInitialAdmin: %v", err)
	}
}

func TestEnsureInitialAdmin_SkipsWhenAdminExists(t *testing.T) {
	svc := newTestService(time.Hour)
	if _, err := svc.CreateUser(context.Background(), "existing_admin", "password123", auth.RoleAdmin); err != nil {
		t.Fatal(err)
	}
	// Should be a no-op — not create a second admin.
	if err := svc.EnsureInitialAdmin(context.Background(), "new_admin", "password456"); err != nil {
		t.Fatalf("EnsureInitialAdmin: %v", err)
	}
	users, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}

func TestDeleteUser(t *testing.T) {
	svc := newTestService(time.Hour)
	view, err := svc.CreateUser(context.Background(), "eve", "password123", auth.RoleViewer)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteUser(context.Background(), view.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	users, _ := svc.ListUsers(context.Background())
	if len(users) != 0 {
		t.Errorf("expected 0 users after delete, got %d", len(users))
	}
}

func TestGetUser(t *testing.T) {
	svc := newTestService(time.Hour)
	created, err := svc.CreateUser(context.Background(), "fred", "password123", auth.RoleViewer)
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetUser(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got.ID != created.ID || got.Username != "fred" || got.Role != auth.RoleViewer {
		t.Errorf("GetUser: got %+v, want id=%s username=fred role=viewer", got, created.ID)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	svc := newTestService(time.Hour)
	_, err := svc.GetUser(context.Background(), "no-such-id")
	if !errors.Is(err, auth.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

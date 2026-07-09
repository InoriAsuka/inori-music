package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"inori-music/services/api/internal/auth"
	"inori-music/services/api/internal/storage"
)

// ---- auth test helpers ----

type memAuthUserRepo struct {
	mu    sync.RWMutex
	users map[string]auth.User
}

func newMemAuthUserRepo() *memAuthUserRepo { return &memAuthUserRepo{users: map[string]auth.User{}} }

func (r *memAuthUserRepo) SaveUser(_ context.Context, u auth.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}
func (r *memAuthUserRepo) GetUser(_ context.Context, id string) (auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return auth.User{}, fmt.Errorf("%w: %s", auth.ErrUserNotFound, id)
	}
	return u, nil
}
func (r *memAuthUserRepo) GetUserByUsername(_ context.Context, username string) (auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return auth.User{}, fmt.Errorf("%w: %s", auth.ErrUserNotFound, username)
}
func (r *memAuthUserRepo) ListUsers(_ context.Context) ([]auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]auth.User, 0, len(r.users))
	for _, u := range r.users {
		list = append(list, u)
	}
	return list, nil
}
func (r *memAuthUserRepo) DeleteUser(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return fmt.Errorf("%w: %s", auth.ErrUserNotFound, id)
	}
	delete(r.users, id)
	return nil
}
func (r *memAuthUserRepo) CountAdminUsers(_ context.Context) (int, error) {
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

type memAuthSessionRepo struct {
	mu       sync.RWMutex
	sessions map[string]auth.Session
}

func newMemAuthSessionRepo() *memAuthSessionRepo {
	return &memAuthSessionRepo{sessions: map[string]auth.Session{}}
}
func (r *memAuthSessionRepo) SaveSession(_ context.Context, s auth.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.TokenHash] = s
	return nil
}
func (r *memAuthSessionRepo) GetSession(_ context.Context, h string) (auth.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[h]
	if !ok {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	return s, nil
}
func (r *memAuthSessionRepo) RevokeSession(_ context.Context, h string, t time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[h]
	if !ok {
		return auth.ErrSessionNotFound
	}
	s.RevokedAt = &t
	r.sessions[h] = s
	return nil
}
func (r *memAuthSessionRepo) ListSessionsByUser(_ context.Context, userID string) ([]auth.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []auth.Session
	for _, s := range r.sessions {
		if s.UserID == userID {
			list = append(list, s)
		}
	}
	return list, nil
}
func (r *memAuthSessionRepo) RevokeAllSessionsByUser(_ context.Context, userID string, revokedAt time.Time) (int, error) {
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
func (r *memAuthSessionRepo) DeleteExpiredSessions(_ context.Context, before time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, s := range r.sessions {
		if s.ExpiresAt.Before(before) {
			delete(r.sessions, k)
		}
	}
	return nil
}

func newAuthTestHandler() (http.Handler, *auth.Service) {
	repo := storage.NewMemoryRepository()
	svc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	if _, err := svc.CreateUser(context.Background(), "admin", "adminpass1", auth.RoleAdmin); err != nil {
		panic(err)
	}
	h := NewHandler(
		storage.NewService(repo),
		WithAuthService(svc),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test-version", Commit: "test-commit", BuildTime: "2026-06-05T12:30:00Z"}),
	).Routes()
	return h, svc
}

// ---- auth endpoint tests ----

func TestLoginSuccess(t *testing.T) {
	h, _ := newAuthTestHandler()
	body := `{"username":"admin","password":"adminpass1"}`
	resp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", body)
	if resp.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["token"] == "" || result["token"] == nil {
		t.Error("expected non-empty token in response")
	}
}

func TestLoginBadCredentials(t *testing.T) {
	h, _ := newAuthTestHandler()
	body := `{"username":"admin","password":"wrongpass"}`
	resp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", body)
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestLoginUnknownUser(t *testing.T) {
	h, _ := newAuthTestHandler()
	body := `{"username":"nobody","password":"adminpass1"}`
	resp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", body)
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestLoginNotConfigured(t *testing.T) {
	repo := storage.NewMemoryRepository()
	h := NewHandler(storage.NewService(repo), WithAdminToken(testAdminToken)).Routes()
	body := `{"username":"admin","password":"adminpass1"}`
	resp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", body)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

func TestLogoutSuccess(t *testing.T) {
	h, _ := newAuthTestHandler()
	// Login first.
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"adminpass1"}`)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d", loginResp.Code)
	}
	var loginResult map[string]any
	json.Unmarshal(loginResp.Body.Bytes(), &loginResult)
	token := loginResult["token"].(string)

	// Logout.
	logoutResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/auth/logout", "", "Bearer "+token)
	if logoutResp.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, body = %s", logoutResp.Code, logoutResp.Body.String())
	}
}

func TestLogoutInvalidToken(t *testing.T) {
	h, _ := newAuthTestHandler()
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/auth/logout", "", "Bearer invalidtoken")
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestSessionTokenGrantsAdminAccess(t *testing.T) {
	h, _ := newAuthTestHandler()
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"adminpass1"}`)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d", loginResp.Code)
	}
	var loginResult map[string]any
	json.Unmarshal(loginResp.Body.Bytes(), &loginResult)
	token := loginResult["token"].(string)

	// Use session token to access an admin route.
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/storage/backends", "", "Bearer "+token)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin route status = %d, body = %s", resp.Code, resp.Body.String())
	}
}

func TestRevokedSessionDeniesAccess(t *testing.T) {
	h, _ := newAuthTestHandler()
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"adminpass1"}`)
	var loginResult map[string]any
	json.Unmarshal(loginResp.Body.Bytes(), &loginResult)
	token := loginResult["token"].(string)

	// Logout (revoke).
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/auth/logout", "", "Bearer "+token)

	// Revoked token should now be denied.
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/storage/backends", "", "Bearer "+token)
	if resp.Code == http.StatusOK {
		t.Fatal("expected denied access after logout, got 200")
	}
}

func TestUserManagementWorkflow(t *testing.T) {
	h, _ := newAuthTestHandler()
	token := loginAdminToken(t, h)

	listed := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/users", "", "Bearer "+token)
	if listed.Code != http.StatusOK {
		t.Fatalf("list users status = %d, body = %s", listed.Code, listed.Body.String())
	}
	assertUserListLength(t, listed, 1)

	created := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users", `{"username":"viewer1","password":"viewerpass1","role":"viewer"}`, "Bearer "+token)
	if created.Code != http.StatusCreated {
		t.Fatalf("create user status = %d, body = %s", created.Code, created.Body.String())
	}
	var createdUser auth.UserView
	decodeResponse(t, created, &createdUser)
	if createdUser.ID == "" || createdUser.Username != "viewer1" || createdUser.Role != auth.RoleViewer || !createdUser.Enabled {
		t.Fatalf("created user = %+v", createdUser)
	}
	if strings.Contains(created.Body.String(), "password") {
		t.Fatalf("response leaked password material: %s", created.Body.String())
	}

	listed = performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/users", "", "Bearer "+token)
	assertUserListLength(t, listed, 2)

	disabled := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users/"+createdUser.ID+"/disable", "", "Bearer "+token)
	if disabled.Code != http.StatusOK {
		t.Fatalf("disable user status = %d, body = %s", disabled.Code, disabled.Body.String())
	}
	var disabledUser auth.UserView
	decodeResponse(t, disabled, &disabledUser)
	if disabledUser.Enabled {
		t.Fatalf("disabled user still enabled: %+v", disabledUser)
	}

	deleted := performRequestWithAuthHeader(t, h, http.MethodDelete, "/api/v1/admin/users/"+createdUser.ID, "", "Bearer "+token)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete user status = %d, body = %s", deleted.Code, deleted.Body.String())
	}
	listed = performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/users", "", "Bearer "+token)
	assertUserListLength(t, listed, 1)
}

func TestUserManagementCreateValidation(t *testing.T) {
	h, _ := newAuthTestHandler()
	token := loginAdminToken(t, h)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users", `{"username":"bad-name","password":"viewerpass1","role":"viewer"}`, "Bearer "+token)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_user")
}

func TestUserManagementCreateConflict(t *testing.T) {
	h, _ := newAuthTestHandler()
	token := loginAdminToken(t, h)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users", `{"username":"admin","password":"viewerpass1","role":"admin"}`, "Bearer "+token)
	assertAPIError(t, resp, http.StatusConflict, "conflict")
}

func TestUserManagementRequiresAdminRole(t *testing.T) {
	h, svc := newAuthTestHandler()
	if _, err := svc.CreateUser(context.Background(), "viewer2", "viewerpass2", auth.RoleViewer); err != nil {
		t.Fatal(err)
	}
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"viewer2","password":"viewerpass2"}`)
	var loginResult map[string]any
	decodeResponse(t, loginResp, &loginResult)
	token := loginResult["token"].(string)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/users", "", "Bearer "+token)
	assertAPIError(t, resp, http.StatusForbidden, "unauthorized")
}

func TestUserManagementNotConfigured(t *testing.T) {
	repo := storage.NewMemoryRepository()
	h := NewHandler(storage.NewService(repo), WithAdminToken(testAdminToken)).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

func TestGetMe(t *testing.T) {
	h, svc := newAuthTestHandler()
	// create a viewer and log in
	if _, err := svc.CreateUser(context.Background(), "viewer_me", "viewpass1", auth.RoleViewer); err != nil {
		t.Fatal(err)
	}
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"viewer_me","password":"viewpass1"}`)
	var loginResult map[string]any
	decodeResponse(t, loginResp, &loginResult)
	token := loginResult["token"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me", "", "Bearer "+token)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /me status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var view auth.UserView
	decodeResponse(t, resp, &view)
	if view.Username != "viewer_me" || view.Role != auth.RoleViewer || !view.Enabled {
		t.Errorf("GET /me: got %+v", view)
	}
	if strings.Contains(resp.Body.String(), "password") {
		t.Errorf("GET /me leaked password material")
	}
}

func TestGetMeUnauthenticated(t *testing.T) {
	h, _ := newAuthTestHandler()
	resp := performRequestWithoutAuth(t, h, http.MethodGet, "/api/v1/me", "")
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestGetMeNotConfigured(t *testing.T) {
	repo := storage.NewMemoryRepository()
	h := NewHandler(storage.NewService(repo), WithAdminToken(testAdminToken)).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/me", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

func TestAdminGetUser(t *testing.T) {
	h, svc := newAuthTestHandler()
	token := loginAdminToken(t, h)
	// create a viewer
	view, err := svc.CreateUser(context.Background(), "target_user", "tpass1234", auth.RoleViewer)
	if err != nil {
		t.Fatal(err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/users/"+view.ID, "", "Bearer "+token)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /admin/users/{id} status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got auth.UserView
	decodeResponse(t, resp, &got)
	if got.ID != view.ID || got.Username != "target_user" || got.Role != auth.RoleViewer {
		t.Errorf("GET /admin/users/{id}: got %+v", got)
	}
}

func TestAdminGetUserNotFound(t *testing.T) {
	h, _ := newAuthTestHandler()
	token := loginAdminToken(t, h)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/users/no-such-id", "", "Bearer "+token)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestAdminGetUserNotConfigured(t *testing.T) {
	repo := storage.NewMemoryRepository()
	h := NewHandler(storage.NewService(repo), WithAdminToken(testAdminToken)).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users/some-id", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

func TestChangePassword(t *testing.T) {
	h, svc := newAuthTestHandler()
	if _, err := svc.CreateUser(context.Background(), "passusr", "oldpasswd1", auth.RoleViewer); err != nil {
		t.Fatal(err)
	}
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"passusr","password":"oldpasswd1"}`)
	var loginResult map[string]any
	decodeResponse(t, loginResp, &loginResult)
	token := loginResult["token"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/change-password",
		`{"currentPassword":"oldpasswd1","newPassword":"newpasswd2"}`, "Bearer "+token)
	if resp.Code != http.StatusNoContent {
		t.Fatalf("POST /me/change-password status = %d, body = %s", resp.Code, resp.Body.String())
	}

	// old password should be rejected
	badLogin := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"passusr","password":"oldpasswd1"}`)
	assertAPIError(t, badLogin, http.StatusUnauthorized, "unauthorized")

	// new password should be accepted
	goodLogin := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"passusr","password":"newpasswd2"}`)
	if goodLogin.Code != http.StatusOK {
		t.Fatalf("login with new password status = %d", goodLogin.Code)
	}
}

func TestChangePasswordWrongCurrent(t *testing.T) {
	h, svc := newAuthTestHandler()
	if _, err := svc.CreateUser(context.Background(), "passusr2", "correctpass", auth.RoleViewer); err != nil {
		t.Fatal(err)
	}
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"passusr2","password":"correctpass"}`)
	var loginResult map[string]any
	decodeResponse(t, loginResp, &loginResult)
	token := loginResult["token"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/change-password",
		`{"currentPassword":"wrongpass","newPassword":"newpasswd2"}`, "Bearer "+token)
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestChangePasswordUnauthenticated(t *testing.T) {
	h, _ := newAuthTestHandler()
	resp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/me/change-password",
		`{"currentPassword":"old","newPassword":"new"}`)
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestChangePasswordNotConfigured(t *testing.T) {
	repo := storage.NewMemoryRepository()
	h := NewHandler(storage.NewService(repo), WithAdminToken(testAdminToken)).Routes()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/me/change-password",
		`{"currentPassword":"old","newPassword":"new"}`)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

func TestEnableUser(t *testing.T) {
	h, svc := newAuthTestHandler()
	token := loginAdminToken(t, h)
	// create and disable a viewer
	view, err := svc.CreateUser(context.Background(), "reenable_usr", "pass1234", auth.RoleViewer)
	if err != nil {
		t.Fatal(err)
	}
	disableResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users/"+view.ID+"/disable", "", "Bearer "+token)
	if disableResp.Code != http.StatusOK {
		t.Fatalf("disable status = %d", disableResp.Code)
	}
	// enable again
	enableResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users/"+view.ID+"/enable", "", "Bearer "+token)
	if enableResp.Code != http.StatusOK {
		t.Fatalf("enable status = %d, body = %s", enableResp.Code, enableResp.Body.String())
	}
	var enabled auth.UserView
	decodeResponse(t, enableResp, &enabled)
	if !enabled.Enabled {
		t.Errorf("enable response: Enabled = false, want true")
	}
}

func TestEnableUserNotFound(t *testing.T) {
	h, _ := newAuthTestHandler()
	token := loginAdminToken(t, h)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users/no-such-id/enable", "", "Bearer "+token)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestEnableUserNotConfigured(t *testing.T) {
	repo := storage.NewMemoryRepository()
	h := NewHandler(storage.NewService(repo), WithAdminToken(testAdminToken)).Routes()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/users/some-id/enable", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

func TestAdminPatchUserRole(t *testing.T) {
	h, svc := newAuthTestHandler()
	token := loginAdminToken(t, h)
	view, err := svc.CreateUser(context.Background(), "patchrole", "pass1234", auth.RoleViewer)
	if err != nil {
		t.Fatal(err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/admin/users/"+view.ID,
		`{"role":"admin"}`, "Bearer "+token)
	if resp.Code != http.StatusOK {
		t.Fatalf("PATCH /admin/users/{id} status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got auth.UserView
	decodeResponse(t, resp, &got)
	if got.Role != auth.RoleAdmin {
		t.Errorf("PATCH role: got %q, want admin", got.Role)
	}
}

func TestAdminPatchUserUsernameConflict(t *testing.T) {
	h, svc := newAuthTestHandler()
	token := loginAdminToken(t, h)
	if _, err := svc.CreateUser(context.Background(), "takenname", "pass1234", auth.RoleViewer); err != nil {
		t.Fatal(err)
	}
	view2, err := svc.CreateUser(context.Background(), "other2", "pass1234", auth.RoleViewer)
	if err != nil {
		t.Fatal(err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/admin/users/"+view2.ID,
		`{"username":"takenname"}`, "Bearer "+token)
	assertAPIError(t, resp, http.StatusConflict, "conflict")
}

func TestAdminPatchUserEmpty(t *testing.T) {
	h, svc := newAuthTestHandler()
	token := loginAdminToken(t, h)
	view, err := svc.CreateUser(context.Background(), "nopatch", "pass1234", auth.RoleViewer)
	if err != nil {
		t.Fatal(err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/admin/users/"+view.ID,
		`{}`, "Bearer "+token)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_user")
}

func TestAdminPatchUserNotConfigured(t *testing.T) {
	repo := storage.NewMemoryRepository()
	h := NewHandler(storage.NewService(repo), WithAdminToken(testAdminToken)).Routes()
	resp := performRequest(t, h, http.MethodPatch, "/api/v1/admin/users/some-id", `{"role":"admin"}`)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

func loginAdminToken(t *testing.T, h http.Handler) string {
	t.Helper()
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"adminpass1"}`)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", loginResp.Code, loginResp.Body.String())
	}
	var loginResult map[string]any
	decodeResponse(t, loginResp, &loginResult)
	token, ok := loginResult["token"].(string)
	if !ok || token == "" {
		t.Fatalf("missing token in login response: %+v", loginResult)
	}
	return token
}

func assertUserListLength(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("list users status = %d, body = %s", response.Code, response.Body.String())
	}
	var body struct {
		Users []auth.UserView `json:"users"`
	}
	decodeResponse(t, response, &body)
	if len(body.Users) != want {
		t.Fatalf("users length = %d, want %d, body = %+v", len(body.Users), want, body.Users)
	}
}

// ---- Phase 91: GET /api/v1/admin/users/{id}/sessions ----

func TestAdminGetUserSessionsEmpty(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	view, err := authSvc.CreateUser(context.Background(), "sess_alice", "pass1234!", auth.RoleViewer)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users/"+view.ID+"/sessions", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("GET sessions: %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if body["count"].(float64) != 0 {
		t.Errorf("count = %v, want 0", body["count"])
	}
}

func TestAdminGetUserSessionsActive(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	view, err := authSvc.CreateUser(context.Background(), "sess_bob", "pass1234!", auth.RoleViewer)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, _, err := authSvc.Login(context.Background(), "sess_bob", "pass1234!"); err != nil {
		t.Fatalf("login: %v", err)
	}
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users/"+view.ID+"/sessions", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("GET sessions: %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if body["count"].(float64) != 1 {
		t.Errorf("count = %v, want 1", body["count"])
	}
}

func TestAdminGetUserSessionsNotFound(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users/does-not-exist/sessions", "")
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestAdminGetUserSessionsNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users/any-id/sessions", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

// ---- Phase 92: DELETE /api/v1/admin/users/{id}/sessions ----

func TestAdminDeleteUserSessionsRevokeActive(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	view, err := authSvc.CreateUser(context.Background(), "revoke_alice", "pass1234!", auth.RoleViewer)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	// Login twice → 2 active sessions.
	for range 2 {
		if _, _, err := authSvc.Login(context.Background(), "revoke_alice", "pass1234!"); err != nil {
			t.Fatalf("login: %v", err)
		}
	}
	resp := performRequest(t, h, http.MethodDelete, "/api/v1/admin/users/"+view.ID+"/sessions", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("DELETE sessions: %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if body["revoked"].(float64) != 2 {
		t.Errorf("revoked = %v, want 2", body["revoked"])
	}
	// Verify sessions are gone.
	sessions, _ := authSvc.ListActiveSessions(context.Background(), view.ID)
	if len(sessions) != 0 {
		t.Errorf("expected 0 active sessions after revoke-all, got %d", len(sessions))
	}
}

func TestAdminDeleteUserSessionsNotFound(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodDelete, "/api/v1/admin/users/no-such-user/sessions", "")
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestAdminDeleteUserSessionsNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodDelete, "/api/v1/admin/users/any-id/sessions", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

// ---- Phase 93: GET /api/v1/me/sessions ----

func TestViewerGetMySessionsFiltersRevoked(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "mysess_alice", "pass1234!", auth.RoleViewer); err != nil {
		t.Fatalf("create user: %v", err)
	}
	token, _, err := authSvc.Login(context.Background(), "mysess_alice", "pass1234!")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	// Revoke own session, then login again to get a valid session for the request.
	if err := authSvc.Logout(context.Background(), token); err != nil {
		t.Fatalf("logout: %v", err)
	}
	token2, _, err := authSvc.Login(context.Background(), "mysess_alice", "pass1234!")
	if err != nil {
		t.Fatalf("login2: %v", err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/sessions", "", "Bearer "+token2)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /me/sessions: %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if body["count"].(float64) != 1 {
		t.Errorf("count = %v, want 1 (only active token2)", body["count"])
	}
}

func TestViewerGetMySessionsActive(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "mysess_bob", "pass1234!", auth.RoleViewer); err != nil {
		t.Fatalf("create user: %v", err)
	}
	var tokens []string
	for range 3 {
		tok, _, err := authSvc.Login(context.Background(), "mysess_bob", "pass1234!")
		if err != nil {
			t.Fatalf("login: %v", err)
		}
		tokens = append(tokens, tok)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/sessions", "", "Bearer "+tokens[0])
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /me/sessions: %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if body["count"].(float64) != 3 {
		t.Errorf("count = %v, want 3", body["count"])
	}
}

func TestViewerGetMySessionsNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/me/sessions", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

// ---- Phase 94: POST /api/v1/me/sessions/revoke-all ----

func TestViewerRevokeMyOtherSessions(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "revoke_me", "pass1234!", auth.RoleViewer); err != nil {
		t.Fatalf("create user: %v", err)
	}
	var tokens []string
	for range 3 {
		tok, _, err := authSvc.Login(context.Background(), "revoke_me", "pass1234!")
		if err != nil {
			t.Fatalf("login: %v", err)
		}
		tokens = append(tokens, tok)
	}
	// Use token[0] as current — should revoke token[1] and token[2].
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/sessions/revoke-all", "", "Bearer "+tokens[0])
	if resp.Code != http.StatusOK {
		t.Fatalf("POST /me/sessions/revoke-all: %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if body["revoked"].(float64) != 2 {
		t.Errorf("revoked = %v, want 2", body["revoked"])
	}
	// Verify only token[0] remains active.
	view, _ := authSvc.ValidateToken(context.Background(), tokens[0])
	if view.Username != "revoke_me" {
		t.Errorf("current session should still be valid after revoke-all")
	}
	for _, tok := range tokens[1:] {
		if _, err := authSvc.ValidateToken(context.Background(), tok); err == nil {
			t.Error("other session should have been revoked")
		}
	}
}

func TestViewerRevokeMyOtherSessionsNoneOther(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "revoke_solo", "pass1234!", auth.RoleViewer); err != nil {
		t.Fatalf("create user: %v", err)
	}
	tok, _, err := authSvc.Login(context.Background(), "revoke_solo", "pass1234!")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/sessions/revoke-all", "", "Bearer "+tok)
	if resp.Code != http.StatusOK {
		t.Fatalf("POST /me/sessions/revoke-all: %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if body["revoked"].(float64) != 0 {
		t.Errorf("revoked = %v, want 0 (only current session exists)", body["revoked"])
	}
}

func TestViewerRevokeMyOtherSessionsNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/me/sessions/revoke-all", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

// ---- Phase 95: GET /api/v1/admin/users pagination + sorting ----

func TestAdminListUsersPagination(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	for _, u := range []struct{ name, pass string }{
		{"alice_pag", "pass1234!"}, {"bob_pag", "pass1234!"}, {"carol_pag", "pass1234!"},
	} {
		if _, err := authSvc.CreateUser(context.Background(), u.name, u.pass, auth.RoleViewer); err != nil {
			t.Fatalf("create user %s: %v", u.name, err)
		}
	}
	// limit=2, offset=0 → first 2 of 4 (admin + 3 viewers)
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users?limit=2&offset=0", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /admin/users?limit=2: %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	users := body["users"].([]any)
	if len(users) != 2 {
		t.Errorf("users len = %d, want 2", len(users))
	}
	pg := body["pagination"].(map[string]any)
	if pg["total"].(float64) < 3 {
		t.Errorf("total = %v, want >= 3", pg["total"])
	}
	if !pg["hasMore"].(bool) {
		t.Error("hasMore should be true")
	}
}

func TestAdminListUsersSortByUsername(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	for _, u := range []string{"zorro_sort", "alpha_sort", "middle_sort"} {
		if _, err := authSvc.CreateUser(context.Background(), u, "pass1234!", auth.RoleViewer); err != nil {
			t.Fatalf("create %s: %v", u, err)
		}
	}
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users?sortBy=username&sortOrder=asc", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d", resp.Code)
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	users := body["users"].([]any)
	// Verify ascending order across the returned slice
	for i := 1; i < len(users); i++ {
		prev := users[i-1].(map[string]any)["username"].(string)
		cur := users[i].(map[string]any)["username"].(string)
		if prev > cur {
			t.Errorf("users not sorted asc at index %d: %q > %q", i, prev, cur)
		}
	}
}

func TestAdminListUsersSortDesc(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	for _, u := range []string{"aaa_desc", "bbb_desc", "ccc_desc"} {
		if _, err := authSvc.CreateUser(context.Background(), u, "pass1234!", auth.RoleViewer); err != nil {
			t.Fatalf("create %s: %v", u, err)
		}
	}
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users?sortBy=username&sortOrder=desc", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d", resp.Code)
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	users := body["users"].([]any)
	for i := 1; i < len(users); i++ {
		prev := users[i-1].(map[string]any)["username"].(string)
		cur := users[i].(map[string]any)["username"].(string)
		if prev < cur {
			t.Errorf("users not sorted desc at index %d: %q < %q", i, prev, cur)
		}
	}
}

func TestAdminListUsersInvalidSortOrder(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users?sortOrder=invalid", "")
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_sort_order")
}

func TestAdminListUsersInvalidLimit(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users?limit=0", "")
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_limit")
}

// ---- Phase 96: GET /api/v1/admin/users filter by username/role/enabled ----

func TestAdminListUsersFilterByRole(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "viewer_filter1", "pass1234!", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "admin_filter1", "pass1234!", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users?role=viewer", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	for _, u := range body["users"].([]any) {
		if u.(map[string]any)["role"].(string) != "viewer" {
			t.Errorf("expected only viewer users; got role=%s", u.(map[string]any)["role"])
		}
	}
}

func TestAdminListUsersFilterByEnabled(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "enabled_filt", "pass1234!", auth.RoleViewer); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "disabled_filt", "pass1234!", auth.RoleViewer); err != nil {
		t.Fatalf("create user: %v", err)
	}
	// Disable the second user
	all, _ := authSvc.ListUsers(context.Background())
	for _, u := range all {
		if u.Username == "disabled_filt" {
			if _, err := authSvc.DisableUser(context.Background(), u.ID); err != nil {
				t.Fatalf("disable: %v", err)
			}
		}
	}
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users?enabled=false", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	for _, u := range body["users"].([]any) {
		if u.(map[string]any)["enabled"].(bool) {
			t.Errorf("expected only disabled users; got enabled=true for %s", u.(map[string]any)["username"])
		}
	}
}

func TestAdminListUsersFilterByUsername(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "unique_user_xyz", "pass1234!", auth.RoleViewer); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "other_user_abc", "pass1234!", auth.RoleViewer); err != nil {
		t.Fatalf("create user: %v", err)
	}
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users?username=unique_user_xyz", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	users := body["users"].([]any)
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
	if len(users) > 0 && users[0].(map[string]any)["username"].(string) != "unique_user_xyz" {
		t.Errorf("unexpected username: %s", users[0].(map[string]any)["username"])
	}
}

func TestAdminListUsersFilterInvalidRole(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users?role=superuser", "")
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_role")
}

func TestAdminListUsersFilterInvalidEnabled(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users?enabled=yes", "")
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_enabled")
}

// ---- Phase 97: POST /api/v1/admin/users/{id}/change-password ----

func TestAdminForceChangePassword(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	view, err := authSvc.CreateUser(context.Background(), "forcepw_user", "oldpass1234", auth.RoleViewer)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	body := `{"newPassword":"newSecurePass9"}`
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/users/"+view.ID+"/change-password", body)
	if resp.Code != http.StatusNoContent {
		t.Fatalf("POST /admin/users/{id}/change-password: %d %s", resp.Code, resp.Body.String())
	}
}

func TestAdminForceChangePasswordWeakPassword(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	view, err := authSvc.CreateUser(context.Background(), "forcepw_weak", "oldpass1234", auth.RoleViewer)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/users/"+view.ID+"/change-password", `{"newPassword":"short"}`)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_user")
}

func TestAdminForceChangePasswordNotFound(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/users/no-such-user/change-password", `{"newPassword":"newSecurePass9"}`)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestAdminForceChangePasswordNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/users/any-id/change-password", `{"newPassword":"newSecurePass9"}`)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

// ---- Phase 98: DELETE /admin/users/{id} cascades session revocation ----

func TestAdminDeleteUserRevokesSessionsFirst(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()

	// Create a viewer and log in to get a session.
	view, err := authSvc.CreateUser(context.Background(), "delete_cascade_http", "pass1234!", auth.RoleViewer)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	tok, _, err := authSvc.Login(context.Background(), "delete_cascade_http", "pass1234!")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	// Delete the user via HTTP.
	resp := performRequest(t, h, http.MethodDelete, "/api/v1/admin/users/"+view.ID, "")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("DELETE /admin/users/{id}: %d %s", resp.Code, resp.Body.String())
	}

	// The former session token should now be invalid.
	if _, err := authSvc.ValidateToken(context.Background(), tok); err == nil {
		t.Error("session should be revoked after user deletion via HTTP")
	}
}

// ---- Phase 99: POST /api/v1/me/sessions/revoke-all-devices ----

func TestViewerRevokeAllMySessions(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "revoke_all_dev", "pass1234!", auth.RoleViewer); err != nil {
		t.Fatalf("create user: %v", err)
	}
	var tokens []string
	for range 3 {
		tok, _, err := authSvc.Login(context.Background(), "revoke_all_dev", "pass1234!")
		if err != nil {
			t.Fatalf("login: %v", err)
		}
		tokens = append(tokens, tok)
	}
	// Use token[0] to invoke revoke-all-devices — it should also be revoked.
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/sessions/revoke-all-devices", "", "Bearer "+tokens[0])
	if resp.Code != http.StatusOK {
		t.Fatalf("POST /me/sessions/revoke-all-devices: %d %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if body["revoked"].(float64) != 3 {
		t.Errorf("revoked = %v, want 3 (including current)", body["revoked"])
	}
	// All tokens should now be invalid.
	for _, tok := range tokens {
		if _, err := authSvc.ValidateToken(context.Background(), tok); err == nil {
			t.Error("token should be revoked after revoke-all-devices")
		}
	}
}

func TestViewerRevokeAllMySessionsIncludesCurrent(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "revoke_all_current", "pass1234!", auth.RoleViewer); err != nil {
		t.Fatalf("create user: %v", err)
	}
	tok, _, err := authSvc.Login(context.Background(), "revoke_all_current", "pass1234!")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/sessions/revoke-all-devices", "", "Bearer "+tok)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d", resp.Code)
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	// Should revoke the single current session too.
	if body["revoked"].(float64) != 1 {
		t.Errorf("revoked = %v, want 1", body["revoked"])
	}
}

func TestViewerRevokeAllMySessionsNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/me/sessions/revoke-all-devices", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

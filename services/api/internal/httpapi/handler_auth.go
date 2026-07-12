package httpapi

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"inori-music/services/api/internal/auth"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	UserID    string    `json:"userId"`
}

func (handler *Handler) login(w http.ResponseWriter, r *http.Request) {
	if handler.authService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "auth_not_configured", "authentication service is not configured")
		return
	}
	var req loginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	// Rate limit: per-IP + per-username.
	clientIP := clientIPFromRequest(r)
	if handler.loginLimiter != nil {
		if handler.loginLimiter.IsLocked("ip:"+clientIP) || handler.loginLimiter.IsLocked("user:"+req.Username) {
			writeAPIError(w, http.StatusTooManyRequests, "too_many_requests", "too many failed login attempts, please try again later")
			return
		}
	}
	token, session, err := handler.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if handler.loginLimiter != nil {
			handler.loginLimiter.RecordFailure("ip:" + clientIP)
			handler.loginLimiter.RecordFailure("user:" + req.Username)
		}
		switch {
		case errors.Is(err, auth.ErrBadCredentials), errors.Is(err, auth.ErrUserDisabled):
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		default:
			writeError(w, err)
		}
		return
	}
	if handler.loginLimiter != nil {
		handler.loginLimiter.ResetFailures("ip:" + clientIP)
		handler.loginLimiter.ResetFailures("user:" + req.Username)
	}
	writeJSON(w, http.StatusOK, loginResponse{
		Token:     token,
		ExpiresAt: session.ExpiresAt,
		UserID:    session.UserID,
	})
}

// clientIPFromRequest extracts the client IP from the request.
// It trusts X-Forwarded-For only when behind a reverse proxy (set via
// INORI_TRUST_PROXY). Falls back to RemoteAddr.
func clientIPFromRequest(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		if idx := strings.IndexByte(xff, ','); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if idx := strings.LastIndexByte(r.RemoteAddr, ':'); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

func (handler *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if handler.authService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "auth_not_configured", "authentication service is not configured")
		return
	}
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		w.Header().Set("WWW-Authenticate", `Bearer realm="inori-admin"`)
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	if err := handler.authService.Logout(r.Context(), token); err != nil {
		if errors.Is(err, auth.ErrSessionNotFound) {
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "token not found or already revoked")
			return
		}
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type createUserRequest struct {
	Username string    `json:"username"`
	Password string    `json:"password"`
	Role     auth.Role `json:"role"`
}

func (handler *Handler) requireAuthService(w http.ResponseWriter) bool {
	if handler.authService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "auth_not_configured", "authentication service is not configured")
		return false
	}
	return true
}

func (handler *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	view, err := handler.authService.GetUser(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (handler *Handler) getMyActiveSessions(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	sessions, err := handler.authService.ListActiveSessions(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions, "count": len(sessions)})
}

func (handler *Handler) revokeMyOtherSessions(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	token, _ := bearerToken(r.Header.Get("Authorization"))
	exceptHash := auth.HashToken(token)
	revoked, err := handler.authService.RevokeAllExcept(r.Context(), user.ID, exceptHash)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"revoked": revoked})
}

func (handler *Handler) revokeAllMySessions(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	revoked, err := handler.authService.RevokeAllSessionsForUser(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"revoked": revoked})
}

func (handler *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	var body struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := decodeJSONWithSentinel(w, r, &body, auth.ErrInvalidUser); err != nil {
		writeError(w, err)
		return
	}
	if body.CurrentPassword == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_user", "currentPassword is required")
		return
	}
	if body.NewPassword == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_user", "newPassword is required")
		return
	}
	if err := handler.authService.ChangePassword(r.Context(), user.ID, body.CurrentPassword, body.NewPassword); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) forceChangePassword(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	var body struct {
		NewPassword string `json:"newPassword"`
	}
	if err := decodeJSONWithSentinel(w, r, &body, auth.ErrInvalidUser); err != nil {
		writeError(w, err)
		return
	}
	if body.NewPassword == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_user", "newPassword is required")
		return
	}
	if err := handler.authService.ForceChangePassword(r.Context(), r.PathValue("id"), body.NewPassword); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	users, err := handler.authService.ListUsers(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	q := r.URL.Query()

	// -- filter --
	if rawUsername := strings.TrimSpace(q.Get("username")); rawUsername != "" {
		filtered := users[:0]
		for _, u := range users {
			if u.Username == rawUsername {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}
	if rawRole := strings.TrimSpace(q.Get("role")); rawRole != "" {
		role := auth.Role(rawRole)
		if role != auth.RoleAdmin && role != auth.RoleViewer {
			writeAPIError(w, http.StatusBadRequest, "invalid_role", "role must be admin or viewer")
			return
		}
		filtered := users[:0]
		for _, u := range users {
			if u.Role == role {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}
	if rawEnabled := strings.TrimSpace(q.Get("enabled")); rawEnabled != "" {
		if rawEnabled != "true" && rawEnabled != "false" {
			writeAPIError(w, http.StatusBadRequest, "invalid_enabled", "enabled must be true or false")
			return
		}
		wantEnabled := rawEnabled == "true"
		filtered := users[:0]
		for _, u := range users {
			if u.Enabled == wantEnabled {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}

	// -- sort --
	sortBy := strings.ToLower(strings.TrimSpace(q.Get("sortBy")))
	sortOrder := strings.ToLower(strings.TrimSpace(q.Get("sortOrder")))
	if sortOrder == "" {
		sortOrder = "asc"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	desc := sortOrder == "desc"
	sort.SliceStable(users, func(i, j int) bool {
		a, b := users[i], users[j]
		var less bool
		switch sortBy {
		case "role":
			less = string(a.Role) < string(b.Role)
		case "createdat":
			less = a.CreatedAt.Before(b.CreatedAt)
		case "updatedat":
			less = a.UpdatedAt.Before(b.UpdatedAt)
		default: // "username"
			less = a.Username < b.Username
		}
		if desc {
			return !less
		}
		return less
	})

	// -- paginate --
	total := len(users)
	limit := 0
	offset := 0
	if raw := q.Get("limit"); raw != "" {
		v, err2 := strconv.Atoi(raw)
		if err2 != nil || v < 1 {
			writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return
		}
		limit = v
	}
	if raw := q.Get("offset"); raw != "" {
		v, err2 := strconv.Atoi(raw)
		if err2 != nil || v < 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_offset", "offset must be a non-negative integer")
			return
		}
		offset = v
	}

	var page []auth.UserView
	if limit > 0 {
		if offset >= total {
			page = []auth.UserView{}
		} else {
			end := offset + limit
			if end > total {
				end = total
			}
			page = users[offset:end]
		}
	} else {
		if offset >= total {
			page = []auth.UserView{}
		} else {
			page = users[offset:]
		}
	}

	hasMore := false
	if limit > 0 {
		hasMore = offset+limit < total
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"users": page,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"hasMore": hasMore,
		},
	})
}

func (handler *Handler) getAdminUser(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	view, err := handler.authService.GetUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (handler *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	var req createUserRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	view, err := handler.authService.CreateUser(r.Context(), req.Username, req.Password, req.Role)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, view)
}

func (handler *Handler) disableUser(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	view, err := handler.authService.DisableUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (handler *Handler) enableUser(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	view, err := handler.authService.EnableUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

type patchUserRequest struct {
	Role     string `json:"role"`
	Username string `json:"username"`
}

func (handler *Handler) patchAdminUser(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	var req patchUserRequest
	if err := decodeJSONWithSentinel(w, r, &req, auth.ErrInvalidUser); err != nil {
		writeError(w, err)
		return
	}
	if req.Role == "" && req.Username == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_user", "at least one of role or username must be set")
		return
	}
	var role *auth.Role
	if req.Role != "" {
		r_ := auth.Role(req.Role)
		role = &r_
	}
	var username *string
	if req.Username != "" {
		username = &req.Username
	}
	view, err := handler.authService.PatchUser(r.Context(), r.PathValue("id"), role, username)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (handler *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	if err := handler.authService.DeleteUser(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) getAdminUserSessions(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	sessions, err := handler.authService.ListActiveSessions(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions, "count": len(sessions)})
}

func (handler *Handler) deleteAdminUserSessions(w http.ResponseWriter, r *http.Request) {
	if !handler.requireAuthService(w) {
		return
	}
	revoked, err := handler.authService.RevokeAllSessionsForUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"revoked": revoked})
}

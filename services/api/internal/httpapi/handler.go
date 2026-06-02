package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"inori-music/services/api/internal/storage"
)

const maxRequestBodyBytes = 1 << 20

// Handler serves versioned administrative HTTP endpoints.
type Handler struct {
	storage                *storage.Service
	adminToken             string
	allowInsecureAdminAuth bool
}

// Option configures HTTP API behavior.
type Option func(*Handler)

// WithAdminToken enables bearer-token authentication for administrative routes.
func WithAdminToken(token string) Option {
	return func(handler *Handler) {
		handler.adminToken = token
	}
}

// WithInsecureAdminAuth disables admin auth and is intended only for local development.
func WithInsecureAdminAuth() Option {
	return func(handler *Handler) {
		handler.allowInsecureAdminAuth = true
	}
}

type storageBackendRequest struct {
	ID          string                `json:"id"`
	Type        storage.BackendType   `json:"type"`
	DisplayName string                `json:"displayName"`
	Enabled     bool                  `json:"enabled"`
	IsDefault   bool                  `json:"isDefault"`
	Priority    int                   `json:"priority"`
	Config      storage.BackendConfig `json:"config"`
}

func (request storageBackendRequest) backend() storage.StorageBackend {
	return storage.StorageBackend{
		ID:          request.ID,
		Type:        request.Type,
		DisplayName: request.DisplayName,
		Enabled:     request.Enabled,
		IsDefault:   request.IsDefault,
		Priority:    request.Priority,
		Config:      request.Config,
	}
}

func NewHandler(storageService *storage.Service, options ...Option) *Handler {
	handler := &Handler{storage: storageService}
	for _, option := range options {
		option(handler)
	}
	return handler
}

func (handler *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.health)
	mux.HandleFunc("GET /api/v1/admin/storage/backends", handler.requireAdmin(handler.listStorageBackends))
	mux.HandleFunc("POST /api/v1/admin/storage/backends", handler.requireAdmin(handler.registerStorageBackend))
	mux.HandleFunc("POST /api/v1/admin/storage/backends/validate", handler.requireAdmin(handler.validateStorageBackend))
	mux.HandleFunc("POST /api/v1/admin/storage/backends/{id}/default", handler.requireAdmin(handler.setDefaultStorageBackend))
	mux.HandleFunc("POST /api/v1/admin/storage/backends/{id}/disable", handler.requireAdmin(handler.disableStorageBackend))
	mux.HandleFunc("/healthz", handler.methodNotAllowed)
	mux.HandleFunc("/api/v1/admin/storage/backends", handler.methodNotAllowed)
	mux.HandleFunc("/api/v1/admin/storage/backends/validate", handler.methodNotAllowed)
	mux.HandleFunc("/api/v1/admin/storage/backends/{id}/default", handler.methodNotAllowed)
	mux.HandleFunc("/api/v1/admin/storage/backends/{id}/disable", handler.methodNotAllowed)
	mux.HandleFunc("/", handler.notFound)
	return mux
}

func (handler *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (handler *Handler) methodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

func (handler *Handler) notFound(w http.ResponseWriter, _ *http.Request) {
	writeAPIError(w, http.StatusNotFound, "not_found", "resource not found")
}

func (handler *Handler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler.allowInsecureAdminAuth {
			next(w, r)
			return
		}
		if handler.adminToken == "" {
			writeAPIError(w, http.StatusServiceUnavailable, "auth_not_configured", "admin authentication is not configured")
			return
		}

		provided, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || provided == "" {
			writeUnauthorized(w)
			return
		}
		providedHash := sha256.Sum256([]byte(provided))
		configuredHash := sha256.Sum256([]byte(handler.adminToken))
		if subtle.ConstantTimeCompare(providedHash[:], configuredHash[:]) != 1 {
			writeUnauthorized(w)
			return
		}

		next(w, r)
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="inori-admin"`)
	writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid admin bearer token is required")
}

func (handler *Handler) listStorageBackends(w http.ResponseWriter, r *http.Request) {
	backends, err := handler.storage.ListBackends(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"backends": backends})
}

func (handler *Handler) registerStorageBackend(w http.ResponseWriter, r *http.Request) {
	var request storageBackendRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, err)
		return
	}
	registered, err := handler.storage.RegisterBackend(r.Context(), request.backend())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, registered)
}

func (handler *Handler) validateStorageBackend(w http.ResponseWriter, r *http.Request) {
	var request storageBackendRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, err)
		return
	}
	validated, err := handler.storage.ValidateBackend(request.backend())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, validated)
}

func (handler *Handler) setDefaultStorageBackend(w http.ResponseWriter, r *http.Request) {
	backend, err := handler.storage.SetDefaultBackend(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, backend)
}

func (handler *Handler) disableStorageBackend(w http.ResponseWriter, r *http.Request) {
	backend, err := handler.storage.DisableBackend(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, backend)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]))
	if contentType != "application/json" {
		return fmt.Errorf("%w: Content-Type must be application/json", storage.ErrInvalidBackend)
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: request body: %v", storage.ErrInvalidBackend, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("%w: request body must contain one JSON value", storage.ErrInvalidBackend)
	}
	return nil
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	switch {
	case errors.Is(err, storage.ErrInvalidBackend):
		status = http.StatusBadRequest
		code = "invalid_backend"
	case errors.Is(err, storage.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, storage.ErrConflict):
		status = http.StatusConflict
		code = "conflict"
	}
	writeAPIError(w, status, code, strings.TrimSpace(err.Error()))
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

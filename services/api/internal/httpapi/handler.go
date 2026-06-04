package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"inori-music/services/api/internal/storage"
)

const maxRequestBodyBytes = 1 << 20

// HandlerOption configures the HTTP API handler.
type HandlerOption func(*Handler)

// WithAdminToken enables Bearer Token authentication for administrator routes.
func WithAdminToken(token string) HandlerOption {
	return func(handler *Handler) {
		handler.adminToken = strings.TrimSpace(token)
	}
}

// WithMediaObjectService enables media object registry routes.
func WithMediaObjectService(mediaObjects *storage.MediaObjectService) HandlerOption {
	return func(handler *Handler) {
		handler.mediaObjects = mediaObjects
	}
}

// Handler serves versioned administrative HTTP endpoints.
type Handler struct {
	storage      *storage.Service
	mediaObjects *storage.MediaObjectService
	adminToken   string
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

type mediaObjectRequest struct {
	ID             string `json:"id"`
	BackendID      string `json:"backendId"`
	ObjectKey      string `json:"objectKey"`
	ContentHash    string `json:"contentHash"`
	SizeBytes      int64  `json:"sizeBytes"`
	MIMEType       string `json:"mimeType"`
	AssetKind      string `json:"assetKind"`
	LifecycleState string `json:"lifecycleState"`
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

func (request mediaObjectRequest) object() storage.MediaObject {
	return storage.MediaObject{
		ID:             request.ID,
		BackendID:      request.BackendID,
		ObjectKey:      request.ObjectKey,
		ContentHash:    request.ContentHash,
		SizeBytes:      request.SizeBytes,
		MIMEType:       request.MIMEType,
		AssetKind:      request.AssetKind,
		LifecycleState: request.LifecycleState,
	}
}

func NewHandler(storageService *storage.Service, options ...HandlerOption) *Handler {
	handler := &Handler{storage: storageService}
	for _, option := range options {
		option(handler)
	}
	return handler
}

func (handler *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.health)
	mux.HandleFunc("GET /api/v1/admin/storage/backends", handler.requireAdminAuth(handler.listStorageBackends))
	mux.HandleFunc("POST /api/v1/admin/storage/backends", handler.requireAdminAuth(handler.registerStorageBackend))
	mux.HandleFunc("POST /api/v1/admin/storage/backends/validate", handler.requireAdminAuth(handler.validateStorageBackend))
	mux.HandleFunc("POST /api/v1/admin/storage/backends/refresh", handler.requireAdminAuth(handler.refreshStorageBackends))
	mux.HandleFunc("POST /api/v1/admin/storage/backends/{id}/default", handler.requireAdminAuth(handler.setDefaultStorageBackend))
	mux.HandleFunc("POST /api/v1/admin/storage/backends/{id}/disable", handler.requireAdminAuth(handler.disableStorageBackend))
	mux.HandleFunc("POST /api/v1/admin/storage/backends/{id}/probe", handler.requireAdminAuth(handler.probeStorageBackend))
	mux.HandleFunc("GET /api/v1/admin/storage/backends/{id}/health", handler.requireAdminAuth(handler.getStorageBackendHealth))
	mux.HandleFunc("GET /api/v1/admin/storage/backends/{id}/capacity", handler.requireAdminAuth(handler.getStorageBackendCapacity))
	mux.HandleFunc("GET /api/v1/admin/media/objects", handler.requireAdminAuth(handler.listMediaObjects))
	mux.HandleFunc("POST /api/v1/admin/media/objects", handler.requireAdminAuth(handler.registerMediaObject))
	mux.HandleFunc("GET /api/v1/admin/media/objects/stats", handler.requireAdminAuth(handler.getMediaObjectStats))
	mux.HandleFunc("POST /api/v1/admin/media/objects/verify", handler.requireAdminAuth(handler.verifyMediaObjects))
	mux.HandleFunc("GET /api/v1/admin/media/objects/{id}", handler.requireAdminAuth(handler.getMediaObject))
	mux.HandleFunc("POST /api/v1/admin/media/objects/{id}/verify", handler.requireAdminAuth(handler.verifyMediaObject))
	mux.HandleFunc("/healthz", handler.methodNotAllowed)
	mux.HandleFunc("/api/v1/admin/storage/backends", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/storage/backends/validate", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/storage/backends/refresh", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/storage/backends/{id}/default", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/storage/backends/{id}/disable", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/storage/backends/{id}/probe", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/storage/backends/{id}/health", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/storage/backends/{id}/capacity", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/media/objects", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/media/objects/{id}", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/media/objects/{id}/verify", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/", handler.requireAdminAuth(handler.notFound))
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

func (handler *Handler) requireAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler.adminToken == "" {
			writeAPIError(w, http.StatusServiceUnavailable, "admin_auth_not_configured", "administrator token is not configured")
			return
		}

		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok || !constantTimeTokenEqual(token, handler.adminToken) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="inori-admin"`)
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid administrator bearer token is required")
			return
		}

		next(w, r)
	}
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func constantTimeTokenEqual(candidate string, expected string) bool {
	candidateHash := sha256.Sum256([]byte(candidate))
	expectedHash := sha256.Sum256([]byte(expected))
	return subtle.ConstantTimeCompare(candidateHash[:], expectedHash[:]) == 1
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

func (handler *Handler) refreshStorageBackends(w http.ResponseWriter, r *http.Request) {
	report, err := handler.storage.RefreshEnabledBackends(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
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

func (handler *Handler) probeStorageBackend(w http.ResponseWriter, r *http.Request) {
	result, err := handler.storage.ProbeBackend(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (handler *Handler) getStorageBackendCapacity(w http.ResponseWriter, r *http.Request) {
	report, err := handler.storage.GetBackendCapacity(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (handler *Handler) getStorageBackendHealth(w http.ResponseWriter, r *http.Request) {
	result, err := handler.storage.GetBackendHealth(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (handler *Handler) registerMediaObject(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	var request mediaObjectRequest
	if err := decodeJSONWithSentinel(w, r, &request, storage.ErrInvalidMediaObject); err != nil {
		writeError(w, err)
		return
	}
	registered, err := handler.mediaObjects.RegisterMediaObject(r.Context(), request.object())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, registered)
}

func (handler *Handler) getMediaObjectStats(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	stats, err := handler.mediaObjects.GetMediaObjectStats(r.Context(), r.URL.Query().Get("backendId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) getMediaObject(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	object, err := handler.mediaObjects.GetMediaObject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, object)
}

func (handler *Handler) verifyMediaObjects(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	backendID := strings.TrimSpace(r.URL.Query().Get("backendId"))
	contentHash := strings.TrimSpace(r.URL.Query().Get("contentHash"))
	if (backendID == "" && contentHash == "") || (backendID != "" && contentHash != "") {
		writeError(w, fmt.Errorf("%w: exactly one of backendId or contentHash is required", storage.ErrInvalidMediaObject))
		return
	}
	var (
		report storage.MediaObjectVerificationReport
		err    error
	)
	if backendID != "" {
		report, err = handler.mediaObjects.VerifyMediaObjectsByBackend(r.Context(), backendID)
	} else {
		report, err = handler.mediaObjects.VerifyMediaObjectsByContentHash(r.Context(), contentHash)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (handler *Handler) verifyMediaObject(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	result, err := handler.mediaObjects.VerifyMediaObject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (handler *Handler) listMediaObjects(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	limit, err := parseMediaObjectListInt(r.URL.Query().Get("limit"), "limit", storage.DefaultMediaObjectListLimit)
	if err != nil {
		writeError(w, err)
		return
	}
	offset, err := parseMediaObjectListInt(r.URL.Query().Get("offset"), "offset", 0)
	if err != nil {
		writeError(w, err)
		return
	}
	page, err := handler.mediaObjects.ListMediaObjects(r.Context(), storage.MediaObjectListFilter{
		BackendID:          r.URL.Query().Get("backendId"),
		ContentHash:        r.URL.Query().Get("contentHash"),
		VerificationStatus: r.URL.Query().Get("verificationStatus"),
		Limit:              limit,
		Offset:             offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func parseMediaObjectListInt(raw string, name string, defaultValue int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%w: %s must be an integer", storage.ErrInvalidMediaObject, name)
	}
	if name == "limit" && value < 1 {
		return 0, fmt.Errorf("%w: limit must be positive", storage.ErrInvalidMediaObject)
	}
	return value, nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	return decodeJSONWithSentinel(w, r, target, storage.ErrInvalidBackend)
}

func decodeJSONWithSentinel(w http.ResponseWriter, r *http.Request, target any, sentinel error) error {
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]))
	if contentType != "application/json" {
		return fmt.Errorf("%w: Content-Type must be application/json", sentinel)
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: request body: %v", sentinel, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("%w: request body must contain one JSON value", sentinel)
	}
	return nil
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	switch {
	case errors.Is(err, storage.ErrInvalidMediaObject):
		status = http.StatusBadRequest
		code = "invalid_media_object"
	case errors.Is(err, storage.ErrInvalidBackend):
		status = http.StatusBadRequest
		code = "invalid_backend"
	case errors.Is(err, storage.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, storage.ErrConflict), errors.Is(err, storage.ErrBackendDisabled):
		status = http.StatusConflict
		code = "conflict"
	case errors.Is(err, storage.ErrMediaObjectVerificationUnsupported):
		status = http.StatusUnprocessableEntity
		code = "media_object_verification_unsupported"
	case errors.Is(err, storage.ErrMediaObjectVerificationFailed):
		status = http.StatusUnprocessableEntity
		code = "media_object_verification_failed"
	case errors.Is(err, storage.ErrProbeUnsupported):
		status = http.StatusUnprocessableEntity
		code = "probe_unsupported"
	case errors.Is(err, storage.ErrProbeFailed):
		status = http.StatusUnprocessableEntity
		code = "probe_failed"
	case errors.Is(err, storage.ErrCapacityUnsupported):
		status = http.StatusUnprocessableEntity
		code = "capacity_unsupported"
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

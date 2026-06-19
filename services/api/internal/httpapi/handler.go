package httpapi

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"inori-music/services/api/internal/auth"
	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/history"
	"inori-music/services/api/internal/storage"
)

const maxRequestBodyBytes = 1 << 20

// contextKeyUser is the typed context key used to pass the authenticated user into handlers.
type contextKeyType int

const contextKeyUser contextKeyType = 0

// userFromContext retrieves the authenticated user injected by requireViewerAuth/requireAdminAuth.
func userFromContext(r *http.Request) (auth.User, bool) {
	u, ok := r.Context().Value(contextKeyUser).(auth.User)
	return u, ok
}

// ServiceInfo describes build metadata exposed by public diagnostic endpoints.
type ServiceInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

// ReadinessCheck describes one public startup readiness check.
type ReadinessCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// ReadinessReport describes whether the API process is ready for admin traffic.
type ReadinessReport struct {
	Ready  bool             `json:"ready"`
	Checks []ReadinessCheck `json:"checks"`
}

type requestMetricKey struct {
	Method string
	Path   string
	Status int
}

type requestMetricValue struct {
	Count           uint64
	DurationSeconds float64
}

type requestMetricSnapshot struct {
	Key             requestMetricKey
	Count           uint64
	DurationSeconds float64
}

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

// WithAuthService enables session-based authentication routes.
func WithAuthService(authSvc *auth.Service) HandlerOption {
	return func(handler *Handler) {
		handler.authService = authSvc
	}
}

// WithCatalogService enables music catalog administration routes.
func WithCatalogService(catalogSvc *catalog.Service) HandlerOption {
	return func(handler *Handler) {
		handler.catalogService = catalogSvc
	}
}

// WithHistoryService enables user playback history routes.
func WithHistoryService(svc *history.Service) HandlerOption {
	return func(handler *Handler) {
		handler.historyService = svc
	}
}

// withCatalogMediaReader wires the media object service into the catalog service
// as a MediaObjectReader after both have been set via options. Called from Routes().
func (handler *Handler) withCatalogMediaReader() {
	if handler.catalogService != nil && handler.mediaObjects != nil {
		handler.catalogService.WithMediaObjectReader(&mediaObjectReaderAdapter{svc: handler.mediaObjects})
	}
}

// mediaObjectReaderAdapter bridges *storage.MediaObjectService to catalog.MediaObjectReader
// without creating a direct import dependency between the catalog and storage packages.
type mediaObjectReaderAdapter struct {
	svc *storage.MediaObjectService
}

func (a *mediaObjectReaderAdapter) GetMediaObjectInfo(ctx context.Context, id string) (catalog.MediaObjectInfo, error) {
	oid, assetKind, lifecycleState, mimeType, err := a.svc.GetMediaObjectInfoForImport(ctx, id)
	if err != nil {
		return catalog.MediaObjectInfo{}, err
	}
	return catalog.MediaObjectInfo{
		ID:             oid,
		AssetKind:      assetKind,
		LifecycleState: lifecycleState,
		MIMEType:       mimeType,
	}, nil
}

// WithServiceInfo configures build metadata returned by public diagnostic endpoints.
func WithServiceInfo(info ServiceInfo) HandlerOption {
	return func(handler *Handler) {
		handler.info = normalizeServiceInfo(info)
	}
}

// Handler serves versioned administrative HTTP endpoints.
type Handler struct {
	storage        *storage.Service
	mediaObjects   *storage.MediaObjectService
	authService    *auth.Service
	catalogService *catalog.Service
	historyService *history.Service
	adminToken     string
	info           ServiceInfo
	metricsMu      sync.Mutex
	requestMetrics map[requestMetricKey]requestMetricValue
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

type mediaObjectLifecycleRequest struct {
	LifecycleState string `json:"lifecycleState"`
}

type mediaObjectSelectionRequest struct {
	BackendID          string `json:"backendId,omitempty"`
	ContentHash        string `json:"contentHash,omitempty"`
	VerificationStatus string `json:"verificationStatus,omitempty"`
	LifecycleState     string `json:"lifecycleState,omitempty"`
	AssetKind          string `json:"assetKind,omitempty"`
}

type mediaObjectBulkLifecycleRequest struct {
	Filter         mediaObjectSelectionRequest `json:"filter"`
	LifecycleState string                      `json:"lifecycleState"`
	DryRun         bool                        `json:"dryRun"`
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

func (request mediaObjectSelectionRequest) filter() storage.MediaObjectSelectionFilter {
	return storage.MediaObjectSelectionFilter{
		BackendID:          request.BackendID,
		ContentHash:        request.ContentHash,
		VerificationStatus: request.VerificationStatus,
		LifecycleState:     request.LifecycleState,
		AssetKind:          request.AssetKind,
	}
}

func defaultServiceInfo() ServiceInfo {
	return ServiceInfo{Name: "inori-api", Version: "dev", Commit: "unknown", BuildTime: "unknown"}
}

func normalizeServiceInfo(info ServiceInfo) ServiceInfo {
	info.Name = strings.TrimSpace(info.Name)
	info.Version = strings.TrimSpace(info.Version)
	info.Commit = strings.TrimSpace(info.Commit)
	info.BuildTime = strings.TrimSpace(info.BuildTime)
	defaults := defaultServiceInfo()
	if info.Name == "" {
		info.Name = defaults.Name
	}
	if info.Version == "" {
		info.Version = defaults.Version
	}
	if info.Commit == "" {
		info.Commit = defaults.Commit
	}
	if info.BuildTime == "" {
		info.BuildTime = defaults.BuildTime
	}
	return info
}

func (handler *Handler) readinessReport() ReadinessReport {
	checks := []ReadinessCheck{
		readinessCheck("storage_service", handler.storage != nil, "storage service is configured", "storage service is not configured"),
		readinessCheck("media_registry", handler.mediaObjects != nil, "media object registry is configured", "media object registry is not configured"),
		readinessCheck("admin_auth", handler.adminToken != "", "admin bearer token is configured", "admin bearer token is not configured"),
	}
	report := ReadinessReport{Ready: true, Checks: checks}
	for _, check := range checks {
		if check.Status != "ok" {
			report.Ready = false
			break
		}
	}
	return report
}

func readinessCheck(name string, ok bool, okMessage string, failureMessage string) ReadinessCheck {
	if ok {
		return ReadinessCheck{Name: name, Status: "ok", Message: okMessage}
	}
	return ReadinessCheck{Name: name, Status: "failed", Message: failureMessage}
}

func NewHandler(storageService *storage.Service, options ...HandlerOption) *Handler {
	handler := &Handler{storage: storageService, info: defaultServiceInfo(), requestMetrics: make(map[requestMetricKey]requestMetricValue)}
	for _, option := range options {
		option(handler)
	}
	return handler
}

func (handler *Handler) Routes() http.Handler {
	handler.withCatalogMediaReader()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.health)
	mux.HandleFunc("GET /metrics", handler.metrics)
	mux.HandleFunc("GET /readyz", handler.readiness)
	mux.HandleFunc("GET /versionz", handler.version)
	mux.HandleFunc("POST /api/v1/auth/login", handler.login)
	mux.HandleFunc("POST /api/v1/auth/logout", handler.logout)
	mux.HandleFunc("GET /api/v1/me", handler.requireViewerAuth(handler.getMe))
	mux.HandleFunc("POST /api/v1/me/change-password", handler.requireViewerAuth(handler.changePassword))
	mux.HandleFunc("GET /api/v1/me/sessions", handler.requireViewerAuth(handler.getMyActiveSessions))
	mux.HandleFunc("POST /api/v1/me/sessions/revoke-all", handler.requireViewerAuth(handler.revokeMyOtherSessions))
	mux.HandleFunc("POST /api/v1/me/sessions/revoke-all-devices", handler.requireViewerAuth(handler.revokeAllMySessions))
	mux.HandleFunc("GET /api/v1/admin/users", handler.requireAdminAuth(handler.listUsers))
	mux.HandleFunc("POST /api/v1/admin/users", handler.requireAdminAuth(handler.createUser))
	mux.HandleFunc("POST /api/v1/admin/users/{id}/disable", handler.requireAdminAuth(handler.disableUser))
	mux.HandleFunc("POST /api/v1/admin/users/{id}/enable", handler.requireAdminAuth(handler.enableUser))
	mux.HandleFunc("PATCH /api/v1/admin/users/{id}", handler.requireAdminAuth(handler.patchAdminUser))
	mux.HandleFunc("GET /api/v1/admin/users/{id}", handler.requireAdminAuth(handler.getAdminUser))
	mux.HandleFunc("DELETE /api/v1/admin/users/{id}", handler.requireAdminAuth(handler.deleteUser))
	mux.HandleFunc("GET /api/v1/admin/users/{id}/sessions", handler.requireAdminAuth(handler.getAdminUserSessions))
	mux.HandleFunc("DELETE /api/v1/admin/users/{id}/sessions", handler.requireAdminAuth(handler.deleteAdminUserSessions))
	mux.HandleFunc("POST /api/v1/admin/users/{id}/change-password", handler.requireAdminAuth(handler.forceChangePassword))
	mux.HandleFunc("GET /api/v1/admin/catalog/artists", handler.requireAdminAuth(handler.listArtists))
	mux.HandleFunc("POST /api/v1/admin/catalog/artists", handler.requireAdminAuth(handler.createArtist))
	mux.HandleFunc("GET /api/v1/admin/catalog/artists/{id}", handler.requireAdminAuth(handler.getArtist))
	mux.HandleFunc("GET /api/v1/admin/catalog/artists/{id}/albums", handler.requireAdminAuth(handler.listAlbumsByArtist))
	mux.HandleFunc("GET /api/v1/admin/catalog/artists/{id}/tracks", handler.requireAdminAuth(handler.listTracksByArtist))
	mux.HandleFunc("PATCH /api/v1/admin/catalog/artists/{id}", handler.requireAdminAuth(handler.patchArtist))
	mux.HandleFunc("DELETE /api/v1/admin/catalog/artists/{id}", handler.requireAdminAuth(handler.deleteArtist))
	mux.HandleFunc("GET /api/v1/admin/catalog/albums", handler.requireAdminAuth(handler.listAlbums))
	mux.HandleFunc("POST /api/v1/admin/catalog/albums", handler.requireAdminAuth(handler.createAlbum))
	mux.HandleFunc("GET /api/v1/admin/catalog/albums/{id}", handler.requireAdminAuth(handler.getAlbum))
	mux.HandleFunc("GET /api/v1/admin/catalog/albums/{id}/tracks", handler.requireAdminAuth(handler.listTracksByAlbum))
	mux.HandleFunc("PATCH /api/v1/admin/catalog/albums/{id}", handler.requireAdminAuth(handler.patchAlbum))
	mux.HandleFunc("DELETE /api/v1/admin/catalog/albums/{id}", handler.requireAdminAuth(handler.deleteAlbum))
	mux.HandleFunc("GET /api/v1/admin/catalog/tracks", handler.requireAdminAuth(handler.listTracks))
	mux.HandleFunc("POST /api/v1/admin/catalog/tracks", handler.requireAdminAuth(handler.createTrack))
	mux.HandleFunc("GET /api/v1/admin/catalog/tracks/{id}", handler.requireAdminAuth(handler.getTrack))
	mux.HandleFunc("PATCH /api/v1/admin/catalog/tracks/{id}", handler.requireAdminAuth(handler.patchTrack))
	mux.HandleFunc("DELETE /api/v1/admin/catalog/tracks/{id}", handler.requireAdminAuth(handler.deleteTrack))
	mux.HandleFunc("POST /api/v1/admin/catalog/tracks/{id}/relink", handler.requireAdminAuth(handler.relinkTrack))
	mux.HandleFunc("POST /api/v1/admin/catalog/import", handler.requireAdminAuth(handler.importTrack))
	mux.HandleFunc("POST /api/v1/admin/catalog/batch-import", handler.requireAdminAuth(handler.batchImportTracks))
	mux.HandleFunc("GET /api/v1/admin/catalog/search", handler.requireAdminAuth(handler.searchCatalog))
	mux.HandleFunc("GET /api/v1/catalog/artists", handler.requireViewerAuth(handler.listArtists))
	mux.HandleFunc("GET /api/v1/catalog/artists/{id}", handler.requireViewerAuth(handler.getArtist))
	mux.HandleFunc("GET /api/v1/catalog/artists/{id}/albums", handler.requireViewerAuth(handler.listAlbumsByArtist))
	mux.HandleFunc("GET /api/v1/catalog/artists/{id}/tracks", handler.requireViewerAuth(handler.listTracksByArtist))
	mux.HandleFunc("GET /api/v1/catalog/albums", handler.requireViewerAuth(handler.listAlbums))
	mux.HandleFunc("GET /api/v1/catalog/albums/{id}", handler.requireViewerAuth(handler.getAlbum))
	mux.HandleFunc("GET /api/v1/catalog/albums/{id}/tracks", handler.requireViewerAuth(handler.listTracksByAlbum))
	mux.HandleFunc("GET /api/v1/catalog/tracks", handler.requireViewerAuth(handler.listTracks))
	mux.HandleFunc("GET /api/v1/catalog/tracks/{id}", handler.requireViewerAuth(handler.getTrack))
	mux.HandleFunc("GET /api/v1/catalog/tracks/{id}/playback", handler.requireViewerAuth(handler.getTrackPlayback))
	mux.HandleFunc("GET /api/v1/catalog/search", handler.requireViewerAuth(handler.searchCatalog))
	mux.HandleFunc("GET /api/v1/admin/catalog/stats", handler.requireAdminAuth(handler.getCatalogStats))
	mux.HandleFunc("GET /api/v1/admin/catalog/stats/artists", handler.requireAdminAuth(handler.getArtistStatsBreakdown))
	mux.HandleFunc("GET /api/v1/admin/catalog/stats/albums", handler.requireAdminAuth(handler.getAlbumStatsBreakdown))
	mux.HandleFunc("GET /api/v1/admin/catalog/stats/playlists", handler.requireAdminAuth(handler.getPlaylistStatsBreakdown))
	mux.HandleFunc("GET /api/v1/admin/catalog/recently-added", handler.requireAdminAuth(handler.getRecentlyAdded))
	mux.HandleFunc("GET /api/v1/admin/catalog/recently-updated", handler.requireAdminAuth(handler.getRecentlyUpdated))
	mux.HandleFunc("GET /api/v1/admin/catalog/playlists", handler.requireAdminAuth(handler.listPlaylists))
	mux.HandleFunc("POST /api/v1/admin/catalog/playlists", handler.requireAdminAuth(handler.createPlaylist))
	mux.HandleFunc("GET /api/v1/admin/catalog/playlists/{id}", handler.requireAdminAuth(handler.getPlaylist))
	mux.HandleFunc("PATCH /api/v1/admin/catalog/playlists/{id}", handler.requireAdminAuth(handler.patchPlaylist))
	mux.HandleFunc("DELETE /api/v1/admin/catalog/playlists/{id}", handler.requireAdminAuth(handler.deletePlaylist))
	mux.HandleFunc("POST /api/v1/admin/catalog/playlists/{id}/tracks", handler.requireAdminAuth(handler.addPlaylistTrack))
	mux.HandleFunc("PUT /api/v1/admin/catalog/playlists/{id}/tracks", handler.requireAdminAuth(handler.setPlaylistTracks))
	mux.HandleFunc("DELETE /api/v1/admin/catalog/playlists/{id}/tracks/{trackId}", handler.requireAdminAuth(handler.removePlaylistTrack))
	mux.HandleFunc("GET /api/v1/admin/catalog/playlists/{id}/tracks", handler.requireAdminAuth(handler.getPlaylistTracks))
	mux.HandleFunc("GET /api/v1/catalog/playlists", handler.requireViewerAuth(handler.listPlaylists))
	mux.HandleFunc("GET /api/v1/catalog/playlists/{id}", handler.requireViewerAuth(handler.getPlaylist))
	mux.HandleFunc("GET /api/v1/catalog/playlists/{id}/tracks", handler.requireViewerAuth(handler.getPlaylistTracks))
	mux.HandleFunc("GET /api/v1/catalog/recently-added", handler.requireViewerAuth(handler.getRecentlyAdded))
	mux.HandleFunc("GET /api/v1/catalog/recently-updated", handler.requireViewerAuth(handler.getRecentlyUpdated))
	mux.HandleFunc("GET /api/v1/catalog/stats", handler.requireViewerAuth(handler.getCatalogStats))
	mux.HandleFunc("GET /api/v1/catalog/stats/artists", handler.requireViewerAuth(handler.getArtistStatsBreakdown))
	mux.HandleFunc("GET /api/v1/catalog/stats/albums", handler.requireViewerAuth(handler.getAlbumStatsBreakdown))
	mux.HandleFunc("GET /api/v1/catalog/stats/playlists", handler.requireViewerAuth(handler.getPlaylistStatsBreakdown))
	mux.HandleFunc("POST /api/v1/me/history", handler.requireViewerAuth(handler.recordPlayEvent))
	mux.HandleFunc("GET /api/v1/me/history", handler.requireViewerAuth(handler.listPlayEvents))
	mux.HandleFunc("DELETE /api/v1/me/history", handler.requireViewerAuth(handler.clearHistory))
	mux.HandleFunc("GET /api/v1/me/history/stats", handler.requireViewerAuth(handler.getMyHistoryStats))
	mux.HandleFunc("GET /api/v1/me/history/top-tracks", handler.requireViewerAuth(handler.getMyTopTracks))
	mux.HandleFunc("GET /api/v1/me/history/timeline", handler.requireViewerAuth(handler.getMyHistoryTimeline))
	mux.HandleFunc("GET /api/v1/me/history/summary", handler.requireViewerAuth(handler.getMyHistorySummary))
	mux.HandleFunc("POST /api/v1/me/history/batch-delete", handler.requireViewerAuth(handler.batchDeleteMyEvents))
	mux.HandleFunc("GET /api/v1/me/history/tracks/{trackId}", handler.requireViewerAuth(handler.getMyTrackHistory))
	mux.HandleFunc("GET /api/v1/me/history/tracks/{trackId}/stats", handler.requireViewerAuth(handler.getMyTrackStats))
	mux.HandleFunc("GET /api/v1/me/history/tracks/{trackId}/timeline", handler.requireViewerAuth(handler.getMyTrackTimeline))
	mux.HandleFunc("GET /api/v1/me/history/tracks/{trackId}/summary", handler.requireViewerAuth(handler.getMyTrackSummary))
	mux.HandleFunc("/api/v1/me/history/tracks/{trackId}/timeline", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/me/history/tracks/{trackId}/stats", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/me/history/tracks/{trackId}", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("GET /api/v1/me/history/{eventId}", handler.requireViewerAuth(handler.getMyEvent))
	mux.HandleFunc("PATCH /api/v1/me/history/{eventId}", handler.requireViewerAuth(handler.patchMyEvent))
	mux.HandleFunc("DELETE /api/v1/me/history/{eventId}", handler.requireViewerAuth(handler.deleteMyEvent))
	mux.HandleFunc("/api/v1/me/history/{eventId}", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/me/history", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("GET /api/v1/admin/history/stats", handler.requireAdminAuth(handler.getAdminHistoryStats))
	mux.HandleFunc("GET /api/v1/admin/history/top-tracks", handler.requireAdminAuth(handler.getAdminTopTracks))
	mux.HandleFunc("GET /api/v1/admin/history/top-users", handler.requireAdminAuth(handler.getAdminTopUsers))
	mux.HandleFunc("GET /api/v1/admin/history/users/{userId}/stats", handler.requireAdminAuth(handler.getAdminUserStats))
	mux.HandleFunc("GET /api/v1/admin/history/users/{userId}/top-tracks", handler.requireAdminAuth(handler.getAdminUserTopTracks))
	mux.HandleFunc("GET /api/v1/admin/history/users/{userId}/timeline", handler.requireAdminAuth(handler.getAdminUserTimeline))
	mux.HandleFunc("GET /api/v1/admin/history/users/{userId}/history-summary", handler.requireAdminAuth(handler.getAdminUserHistorySummary))
	mux.HandleFunc("/api/v1/admin/history/users/{userId}/stats", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/history/users/{userId}/top-tracks", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/history/users/{userId}/timeline", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/history/users/{userId}/history-summary", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("GET /api/v1/admin/history/users/{userId}", handler.requireAdminAuth(handler.getAdminUserHistory))
	mux.HandleFunc("GET /api/v1/admin/history/tracks/{trackId}/stats", handler.requireAdminAuth(handler.getAdminTrackStats))
	mux.HandleFunc("GET /api/v1/admin/history/tracks/{trackId}/top-listeners", handler.requireAdminAuth(handler.getAdminTrackTopListeners))
	mux.HandleFunc("GET /api/v1/admin/history/tracks/{trackId}/timeline", handler.requireAdminAuth(handler.getAdminTrackTimeline))
	mux.HandleFunc("GET /api/v1/admin/history/tracks/{trackId}/history-summary", handler.requireAdminAuth(handler.getAdminTrackHistorySummary))
	mux.HandleFunc("/api/v1/admin/history/tracks/{trackId}/stats", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/history/tracks/{trackId}/top-listeners", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/history/tracks/{trackId}/timeline", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/history/tracks/{trackId}/history-summary", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("GET /api/v1/admin/history/tracks/{trackId}", handler.requireAdminAuth(handler.getAdminTrackHistory))
	mux.HandleFunc("DELETE /api/v1/admin/history/users/{userId}", handler.requireAdminAuth(handler.deleteAdminUserHistory))
	mux.HandleFunc("DELETE /api/v1/admin/history/tracks/{trackId}", handler.requireAdminAuth(handler.deleteAdminTrackHistory))
	mux.HandleFunc("DELETE /api/v1/admin/history", handler.requireAdminAuth(handler.deleteAdminHistoryWindow))
	mux.HandleFunc("/api/v1/admin/history/users/{userId}", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/history/tracks/{trackId}", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("GET /api/v1/admin/history/timeline", handler.requireAdminAuth(handler.getAdminHistoryTimeline))
	mux.HandleFunc("POST /api/v1/admin/history/batch-delete", handler.requireAdminAuth(handler.batchDeleteAdminEvents))
	mux.HandleFunc("GET /api/v1/admin/history/summary", handler.requireAdminAuth(handler.getAdminHistorySummary))
	mux.HandleFunc("GET /api/v1/admin/history/{eventId}", handler.requireAdminAuth(handler.getAdminEvent))
	mux.HandleFunc("PATCH /api/v1/admin/history/{eventId}", handler.requireAdminAuth(handler.patchAdminEvent))
	mux.HandleFunc("DELETE /api/v1/admin/history/{eventId}", handler.requireAdminAuth(handler.deleteAdminEvent))
	mux.HandleFunc("GET /api/v1/admin/history", handler.requireAdminAuth(handler.getAdminAllHistory))
	mux.HandleFunc("/api/v1/admin/history/{eventId}", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/history", handler.requireAdminAuth(handler.methodNotAllowed))
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
	mux.HandleFunc("GET /api/v1/admin/media/objects/duplicates", handler.requireAdminAuth(handler.getMediaObjectDuplicates))
	mux.HandleFunc("POST /api/v1/admin/media/objects/lifecycle", handler.requireAdminAuth(handler.setMediaObjectsLifecycle))
	mux.HandleFunc("POST /api/v1/admin/media/objects/verify", handler.requireAdminAuth(handler.verifyMediaObjects))
	mux.HandleFunc("GET /api/v1/admin/media/objects/{id}", handler.requireAdminAuth(handler.getMediaObject))
	mux.HandleFunc("GET /api/v1/admin/media/objects/{id}/timeline", handler.requireAdminAuth(handler.getMediaObjectTimeline))
	mux.HandleFunc("POST /api/v1/admin/media/objects/{id}/lifecycle", handler.requireAdminAuth(handler.setMediaObjectLifecycle))
	mux.HandleFunc("POST /api/v1/admin/media/objects/{id}/verify", handler.requireAdminAuth(handler.verifyMediaObject))
	mux.HandleFunc("/healthz", handler.methodNotAllowed)
	mux.HandleFunc("/metrics", handler.methodNotAllowed)
	mux.HandleFunc("/readyz", handler.methodNotAllowed)
	mux.HandleFunc("/versionz", handler.methodNotAllowed)
	mux.HandleFunc("/api/v1/auth/login", handler.methodNotAllowed)
	mux.HandleFunc("/api/v1/auth/logout", handler.methodNotAllowed)
	mux.HandleFunc("/api/v1/me", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/me/change-password", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/me/sessions", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/me/sessions/revoke-all", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/me/sessions/revoke-all-devices", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/users", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/users/{id}/disable", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/users/{id}/enable", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/users/{id}/sessions", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/users/{id}/change-password", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/users/{id}", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/artists", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/artists/{id}", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/artists/{id}/albums", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/artists/{id}/tracks", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/albums", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/albums/{id}", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/albums/{id}/tracks", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/tracks", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/tracks/{id}", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/tracks/{id}/relink", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/import", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/batch-import", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/search", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/stats", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/stats/artists", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/stats/albums", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/stats/playlists", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/recently-added", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/recently-updated", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/playlists", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/playlists/{id}", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/playlists/{id}/tracks", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/playlists/{id}/tracks", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/catalog/playlists/{id}/tracks/{trackId}", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/playlists", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/playlists/{id}", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/artists", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/artists/{id}", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/artists/{id}/albums", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/artists/{id}/tracks", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/albums", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/albums/{id}", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/albums/{id}/tracks", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/tracks", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/tracks/{id}", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/tracks/{id}/playback", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/search", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/recently-added", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/recently-updated", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/stats", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/stats/artists", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/stats/albums", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/stats/playlists", handler.requireViewerAuth(handler.methodNotAllowed))
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
	mux.HandleFunc("/api/v1/admin/media/objects/{id}/timeline", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/media/objects/{id}/lifecycle", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/media/objects/{id}/verify", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/catalog/", handler.requireViewerAuth(handler.notFound))
	mux.HandleFunc("/api/v1/admin/", handler.requireAdminAuth(handler.notFound))
	mux.HandleFunc("/", handler.notFound)
	return handler.instrument(mux)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (recorder *statusRecorder) WriteHeader(status int) {
	if recorder.wrote {
		return
	}
	recorder.status = status
	recorder.wrote = true
	recorder.ResponseWriter.WriteHeader(status)
}

func (recorder *statusRecorder) Write(data []byte) (int, error) {
	if !recorder.wrote {
		recorder.WriteHeader(http.StatusOK)
	}
	return recorder.ResponseWriter.Write(data)
}

func (handler *Handler) instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		path := r.Pattern
		if path == "" {
			path = "unmatched"
		}
		handler.recordRequestMetric(r.Method, path, recorder.status, time.Since(started).Seconds())
	})
}

func (handler *Handler) recordRequestMetric(method string, path string, status int, durationSeconds float64) {
	key := requestMetricKey{Method: method, Path: path, Status: status}
	handler.metricsMu.Lock()
	defer handler.metricsMu.Unlock()
	value := handler.requestMetrics[key]
	value.Count++
	value.DurationSeconds += durationSeconds
	handler.requestMetrics[key] = value
}

func (handler *Handler) requestMetricSnapshots() []requestMetricSnapshot {
	handler.metricsMu.Lock()
	defer handler.metricsMu.Unlock()
	snapshots := make([]requestMetricSnapshot, 0, len(handler.requestMetrics))
	for key, value := range handler.requestMetrics {
		snapshots = append(snapshots, requestMetricSnapshot{Key: key, Count: value.Count, DurationSeconds: value.DurationSeconds})
	}
	sort.Slice(snapshots, func(i, j int) bool {
		if snapshots[i].Key.Path != snapshots[j].Key.Path {
			return snapshots[i].Key.Path < snapshots[j].Key.Path
		}
		if snapshots[i].Key.Method != snapshots[j].Key.Method {
			return snapshots[i].Key.Method < snapshots[j].Key.Method
		}
		return snapshots[i].Key.Status < snapshots[j].Key.Status
	})
	return snapshots
}

func (handler *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (handler *Handler) readiness(w http.ResponseWriter, _ *http.Request) {
	report := handler.readinessReport()
	status := http.StatusOK
	if !report.Ready {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, report)
}

func (handler *Handler) metrics(w http.ResponseWriter, _ *http.Request) {
	report := handler.readinessReport()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "# HELP inori_api_ready Whether the API process is ready for admin traffic.")
	fmt.Fprintln(w, "# TYPE inori_api_ready gauge")
	fmt.Fprintf(w, "inori_api_ready %d\n", prometheusBool(report.Ready))
	fmt.Fprintln(w, "# HELP inori_api_readiness_check Readiness status per public startup check.")
	fmt.Fprintln(w, "# TYPE inori_api_readiness_check gauge")
	for _, check := range report.Checks {
		fmt.Fprintf(w, "inori_api_readiness_check{check=\"%s\"} %d\n", prometheusLabelValue(check.Name), prometheusBool(check.Status == "ok"))
	}
	fmt.Fprintln(w, "# HELP inori_api_info Build metadata for the running API process.")
	fmt.Fprintln(w, "# TYPE inori_api_info gauge")
	fmt.Fprintf(w, "inori_api_info{name=\"%s\",version=\"%s\",commit=\"%s\",build_time=\"%s\"} 1\n", prometheusLabelValue(handler.info.Name), prometheusLabelValue(handler.info.Version), prometheusLabelValue(handler.info.Commit), prometheusLabelValue(handler.info.BuildTime))
	fmt.Fprintln(w, "# HELP inori_api_http_requests_total Total HTTP requests handled by method, route pattern, and status.")
	fmt.Fprintln(w, "# TYPE inori_api_http_requests_total counter")
	fmt.Fprintln(w, "# HELP inori_api_http_request_duration_seconds_sum Cumulative HTTP request duration in seconds by method, route pattern, and status.")
	fmt.Fprintln(w, "# TYPE inori_api_http_request_duration_seconds_sum counter")
	for _, metric := range handler.requestMetricSnapshots() {
		method := prometheusLabelValue(metric.Key.Method)
		path := prometheusLabelValue(metric.Key.Path)
		status := prometheusLabelValue(strconv.Itoa(metric.Key.Status))
		fmt.Fprintf(w, "inori_api_http_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n", method, path, status, metric.Count)
		fmt.Fprintf(w, "inori_api_http_request_duration_seconds_sum{method=\"%s\",path=\"%s\",status=\"%s\"} %.6f\n", method, path, status, metric.DurationSeconds)
	}
}

func prometheusBool(value bool) int {
	if value {
		return 1
	}
	return 0
}

func prometheusLabelValue(value string) string {
	return strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\"", "\\\"").Replace(value)
}

func (handler *Handler) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, handler.info)
}

func (handler *Handler) methodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

func (handler *Handler) notFound(w http.ResponseWriter, _ *http.Request) {
	writeAPIError(w, http.StatusNotFound, "not_found", "resource not found")
}

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
	token, session, err := handler.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrBadCredentials), errors.Is(err, auth.ErrUserDisabled):
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		default:
			writeError(w, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{
		Token:     token,
		ExpiresAt: session.ExpiresAt,
		UserID:    session.UserID,
	})
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

// ---- catalog helpers ----

func (handler *Handler) requireCatalogService(w http.ResponseWriter) bool {
	if handler.catalogService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "catalog_not_configured", "catalog service is not configured")
		return false
	}
	return true
}

// ---- import handler ----

type importTrackRequest struct {
	MediaObjectID string `json:"mediaObjectId"`
	Title         string `json:"title"`
	SortTitle     string `json:"sortTitle"`
	ArtistID      string `json:"artistId"`
	AlbumID       string `json:"albumId"`
	TrackNumber   int    `json:"trackNumber"`
	DiscNumber    int    `json:"discNumber"`
	DurationMS    int    `json:"durationMs"`
}

func (handler *Handler) importTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req importTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	track, err := handler.catalogService.ImportTrack(r.Context(), catalog.ImportTrackRequest{
		MediaObjectID: req.MediaObjectID,
		Title:         req.Title,
		SortTitle:     req.SortTitle,
		ArtistID:      req.ArtistID,
		AlbumID:       req.AlbumID,
		TrackNumber:   req.TrackNumber,
		DiscNumber:    req.DiscNumber,
		DurationMS:    req.DurationMS,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, track)
}

// ---- batch import handler ----

type batchImportRequest struct {
	Items []importTrackRequest `json:"items"`
}

func (handler *Handler) batchImportTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req batchImportRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	items := make([]catalog.ImportTrackRequest, len(req.Items))
	for i, it := range req.Items {
		items[i] = catalog.ImportTrackRequest{
			MediaObjectID: it.MediaObjectID,
			Title:         it.Title,
			SortTitle:     it.SortTitle,
			ArtistID:      it.ArtistID,
			AlbumID:       it.AlbumID,
			TrackNumber:   it.TrackNumber,
			DiscNumber:    it.DiscNumber,
			DurationMS:    it.DurationMS,
		}
	}
	result := handler.catalogService.BatchImportTracks(r.Context(), items)
	status := http.StatusOK
	if result.Failed > 0 && result.Imported > 0 {
		status = http.StatusMultiStatus
	} else if result.Failed > 0 && result.Imported == 0 {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, result)
}

// ---- search handler ----

func (handler *Handler) searchCatalog(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_query", "query parameter 'q' is required")
		return
	}
	result, err := handler.catalogService.SearchCatalog(r.Context(), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// ---- catalog pagination & sort helpers ----

const (
	catalogListDefaultLimit = 50
	catalogListMaxLimit     = 500
)

// parseCatalogPage parses limit, offset, sortBy, and sortOrder query parameters.
// limit defaults to catalogListDefaultLimit and is clamped to catalogListMaxLimit.
// sortBy and sortOrder are returned as trimmed lowercase strings; empty strings
// signal "use entity default". Returns false and writes an error when limit or
// offset are invalid.
func parseCatalogPage(w http.ResponseWriter, r *http.Request) (limit, offset int, sortBy, sortOrder string, ok bool) {
	q := r.URL.Query()
	limit = catalogListDefaultLimit
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return 0, 0, "", "", false
		}
		if v > catalogListMaxLimit {
			v = catalogListMaxLimit
		}
		limit = v
	}
	if raw := q.Get("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_offset", "offset must be a non-negative integer")
			return 0, 0, "", "", false
		}
		offset = v
	}
	sortBy = strings.ToLower(strings.TrimSpace(q.Get("sortBy")))
	sortOrder = strings.ToLower(strings.TrimSpace(q.Get("sortOrder")))
	return limit, offset, sortBy, sortOrder, true
}

// normalizeSortOrder returns "asc" or "desc". Empty input → "asc".
// Returns "", false for any other value.
func normalizeSortOrder(raw string) (string, bool) {
	switch raw {
	case "", catalog.CatalogSortOrderAsc:
		return catalog.CatalogSortOrderAsc, true
	case catalog.CatalogSortOrderDesc:
		return catalog.CatalogSortOrderDesc, true
	}
	return "", false
}

// sortCatalogArtists sorts artists in-place by sortBy/sortOrder.
// Invalid sortBy is treated as "name" (default).
func sortCatalogArtists(artists []catalog.Artist, sortBy, sortOrder string) {
	desc := sortOrder == catalog.CatalogSortOrderDesc
	sort.SliceStable(artists, func(i, j int) bool {
		a, b := artists[i], artists[j]
		var less bool
		switch sortBy {
		case catalog.ArtistSortBySortName:
			less = a.SortName < b.SortName
		case catalog.ArtistSortByCreatedAt:
			less = a.CreatedAt.Before(b.CreatedAt)
		case catalog.ArtistSortByUpdatedAt:
			less = a.UpdatedAt.Before(b.UpdatedAt)
		default: // "name"
			less = a.Name < b.Name
		}
		if desc {
			return !less
		}
		return less
	})
}

// sortCatalogAlbums sorts albums in-place by sortBy/sortOrder.
func sortCatalogAlbums(albums []catalog.Album, sortBy, sortOrder string) {
	desc := sortOrder == catalog.CatalogSortOrderDesc
	sort.SliceStable(albums, func(i, j int) bool {
		a, b := albums[i], albums[j]
		var less bool
		switch sortBy {
		case catalog.AlbumSortBySortTitle:
			less = a.SortTitle < b.SortTitle
		case catalog.AlbumSortByReleaseYear:
			less = a.ReleaseYear < b.ReleaseYear
		case catalog.AlbumSortByCreatedAt:
			less = a.CreatedAt.Before(b.CreatedAt)
		case catalog.AlbumSortByUpdatedAt:
			less = a.UpdatedAt.Before(b.UpdatedAt)
		default: // "title"
			less = a.Title < b.Title
		}
		if desc {
			return !less
		}
		return less
	})
}

// sortCatalogTracks sorts tracks in-place by sortBy/sortOrder.
func sortCatalogTracks(tracks []catalog.Track, sortBy, sortOrder string) {
	desc := sortOrder == catalog.CatalogSortOrderDesc
	sort.SliceStable(tracks, func(i, j int) bool {
		a, b := tracks[i], tracks[j]
		var less bool
		switch sortBy {
		case catalog.TrackSortBySortTitle:
			less = a.SortTitle < b.SortTitle
		case catalog.TrackSortByTrackNumber:
			less = a.TrackNumber < b.TrackNumber
		case catalog.TrackSortByDiscNumber:
			less = a.DiscNumber < b.DiscNumber
		case catalog.TrackSortByDurationMS:
			less = a.DurationMS < b.DurationMS
		case catalog.TrackSortByCreatedAt:
			less = a.CreatedAt.Before(b.CreatedAt)
		case catalog.TrackSortByUpdatedAt:
			less = a.UpdatedAt.Before(b.UpdatedAt)
		default: // "title"
			less = a.Title < b.Title
		}
		if desc {
			return !less
		}
		return less
	})
}

// sortCatalogPlaylists sorts playlists in-place by sortBy/sortOrder.
func sortCatalogPlaylists(playlists []catalog.Playlist, sortBy, sortOrder string) {
	desc := sortOrder == catalog.CatalogSortOrderDesc
	sort.SliceStable(playlists, func(i, j int) bool {
		a, b := playlists[i], playlists[j]
		var less bool
		switch sortBy {
		case catalog.PlaylistSortByCreatedAt:
			less = a.CreatedAt.Before(b.CreatedAt)
		case catalog.PlaylistSortByUpdatedAt:
			less = a.UpdatedAt.Before(b.UpdatedAt)
		default: // "name"
			less = a.Name < b.Name
		}
		if desc {
			return !less
		}
		return less
	})
}

// paginateCatalog slices items[offset:offset+limit] and returns the page meta.
func paginateCatalog[T any](items []T, limit, offset int) ([]T, catalog.CatalogPaginationMeta) {
	total := len(items)
	if offset >= total {
		return []T{}, catalog.CatalogPaginationMeta{Limit: limit, Offset: offset, Total: total, HasMore: false}
	}
	end := offset + limit
	if end > total {
		end = total
	}
	page := items[offset:end]
	return page, catalog.CatalogPaginationMeta{
		Limit:   limit,
		Offset:  offset,
		Total:   total,
		HasMore: end < total,
	}
}

// ---- artist handlers ----

type createArtistRequest struct {
	Name     string `json:"name"`
	SortName string `json:"sortName"`
}

func (handler *Handler) listArtists(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	page, err := handler.catalogService.ListArtistsPage(r.Context(), catalog.ListQuery{
		SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: page.Total, HasMore: offset+limit < page.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"artists": page.Items, "pagination": meta})
}

func (handler *Handler) createArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req createArtistRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	artist, err := handler.catalogService.CreateArtist(r.Context(), req.Name, req.SortName)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, artist)
}

func (handler *Handler) getArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	artist, err := handler.catalogService.GetArtist(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, artist)
}

func (handler *Handler) deleteArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	if err := handler.catalogService.DeleteArtist(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listAlbumsByArtist returns paged, sorted albums belonging to the artist identified
// by the {id} path parameter. The parent artist must exist; unknown IDs return 404.
func (handler *Handler) listAlbumsByArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	// Verify the artist exists before listing — produces 404 on unknown IDs.
	if _, err := handler.catalogService.GetArtist(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	albumPage, err := handler.catalogService.ListAlbumsByArtistPage(r.Context(), r.PathValue("id"), catalog.ListQuery{
		SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: albumPage.Total, HasMore: offset+limit < albumPage.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"albums": albumPage.Items, "pagination": meta})
}

// listTracksByArtist returns paged, sorted tracks belonging to the artist identified
// by the {id} path parameter. The parent artist must exist; unknown IDs return 404.
func (handler *Handler) listTracksByArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	if _, err := handler.catalogService.GetArtist(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	trackPage, err := handler.catalogService.ListTracksByArtistPage(r.Context(), r.PathValue("id"), catalog.ListQuery{
		SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: trackPage.Total, HasMore: offset+limit < trackPage.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": trackPage.Items, "pagination": meta})
}

// patchArtistRequest carries the fields that may be changed via PATCH.
// Pointer semantics: nil = leave unchanged, pointer-to-string = set new value.
type patchArtistRequest struct {
	Name     *string `json:"name"`
	SortName *string `json:"sortName"`
}

func (handler *Handler) patchArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req patchArtistRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	artist, err := handler.catalogService.UpdateArtist(r.Context(), r.PathValue("id"), catalog.UpdateArtistRequest{
		Name:     req.Name,
		SortName: req.SortName,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, artist)
}

// ---- album handlers ----

type createAlbumRequest struct {
	Title       string `json:"title"`
	SortTitle   string `json:"sortTitle"`
	ArtistID    string `json:"artistId"`
	ReleaseYear int    `json:"releaseYear"`
}

func (handler *Handler) listAlbums(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	q := catalog.ListQuery{SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset}
	artistID := r.URL.Query().Get("artistId")
	var (
		page catalog.ListPage[catalog.Album]
		err  error
	)
	if artistID != "" {
		page, err = handler.catalogService.ListAlbumsByArtistPage(r.Context(), artistID, q)
	} else {
		page, err = handler.catalogService.ListAlbumsPage(r.Context(), q)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: page.Total, HasMore: offset+limit < page.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"albums": page.Items, "pagination": meta})
}

func (handler *Handler) createAlbum(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req createAlbumRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	album, err := handler.catalogService.CreateAlbum(r.Context(), req.Title, req.SortTitle, req.ArtistID, req.ReleaseYear)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, album)
}

func (handler *Handler) getAlbum(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	album, err := handler.catalogService.GetAlbum(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, album)
}

func (handler *Handler) deleteAlbum(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	if err := handler.catalogService.DeleteAlbum(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listTracksByAlbum returns paged, sorted tracks belonging to the album identified
// by the {id} path parameter. The parent album must exist; unknown IDs return 404.
func (handler *Handler) listTracksByAlbum(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	if _, err := handler.catalogService.GetAlbum(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	trackPage, err := handler.catalogService.ListTracksByAlbumPage(r.Context(), r.PathValue("id"), catalog.ListQuery{
		SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: trackPage.Total, HasMore: offset+limit < trackPage.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": trackPage.Items, "pagination": meta})
}

// patchAlbumRequest carries the fields that may be changed via PATCH.
type patchAlbumRequest struct {
	Title       *string `json:"title"`
	SortTitle   *string `json:"sortTitle"`
	ArtistID    *string `json:"artistId"`
	ReleaseYear *int    `json:"releaseYear"`
}

func (handler *Handler) patchAlbum(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req patchAlbumRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	album, err := handler.catalogService.UpdateAlbum(r.Context(), r.PathValue("id"), catalog.UpdateAlbumRequest{
		Title:       req.Title,
		SortTitle:   req.SortTitle,
		ArtistID:    req.ArtistID,
		ReleaseYear: req.ReleaseYear,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, album)
}

// ---- track handlers ----

type createTrackRequest struct {
	Title         string `json:"title"`
	SortTitle     string `json:"sortTitle"`
	ArtistID      string `json:"artistId"`
	AlbumID       string `json:"albumId"`
	MediaObjectID string `json:"mediaObjectId"`
	TrackNumber   int    `json:"trackNumber"`
	DiscNumber    int    `json:"discNumber"`
	DurationMS    int    `json:"durationMs"`
}

func (handler *Handler) listTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	q := catalog.ListQuery{SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset}
	queryArgs := r.URL.Query()
	artistID := queryArgs.Get("artistId")
	albumID := queryArgs.Get("albumId")
	var (
		page catalog.ListPage[catalog.Track]
		err  error
	)
	switch {
	case albumID != "":
		page, err = handler.catalogService.ListTracksByAlbumPage(r.Context(), albumID, q)
	case artistID != "":
		page, err = handler.catalogService.ListTracksByArtistPage(r.Context(), artistID, q)
	default:
		page, err = handler.catalogService.ListTracksPage(r.Context(), q)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: page.Total, HasMore: offset+limit < page.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": page.Items, "pagination": meta})
}

func (handler *Handler) createTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req createTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	track, err := handler.catalogService.CreateTrack(r.Context(), req.Title, req.SortTitle, req.ArtistID, req.AlbumID, req.MediaObjectID, req.TrackNumber, req.DiscNumber, req.DurationMS)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, track)
}

func (handler *Handler) getTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	track, err := handler.catalogService.GetTrack(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, track)
}

// trackPlaybackDescriptor is the metadata-only response returned by the viewer
// playback endpoint. It carries the fields a client needs to fetch the audio
// file from its storage backend without the server streaming bytes.
// PresignedURL is populated when the backend supports presigned URLs and
// credentials are available; it is omitted otherwise.
type trackPlaybackDescriptor struct {
	TrackID       string `json:"trackId"`
	MediaObjectID string `json:"mediaObjectId"`
	MIMEType      string `json:"mimeType"`
	DurationMS    int    `json:"durationMs"`
	BackendID     string `json:"backendId"`
	BackendType   string `json:"backendType,omitempty"`
	ObjectKey     string `json:"objectKey"`
	PresignedURL  string `json:"presignedUrl,omitempty"`
}

func (handler *Handler) getTrackPlayback(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	track, err := handler.catalogService.GetTrack(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	mo, err := handler.mediaObjects.GetMediaObject(r.Context(), track.MediaObjectID)
	if err != nil {
		writeError(w, err)
		return
	}
	if mo.LifecycleState != string(storage.LifecycleStateActive) ||
		(mo.AssetKind != string(storage.AssetKindOriginalAudio) && mo.AssetKind != string(storage.AssetKindTranscodedAudio)) {
		writeError(w, fmt.Errorf("%w: media object %s is not in a playable state (lifecycleState=%s assetKind=%s)",
			storage.ErrPlaybackUnavailable, mo.ID, mo.LifecycleState, mo.AssetKind))
		return
	}
	backendType := ""
	presignedURL := ""
	if handler.storage != nil {
		backend, backendErr := handler.storage.GetBackend(r.Context(), mo.BackendID)
		if backendErr == nil {
			backendType = string(backend.Type)
			if backend.Capabilities.PresignedURLs {
				if purl, pErr := handler.storage.GeneratePresignedURL(
					r.Context(), mo.BackendID, mo.ObjectKey, storage.DefaultPresignedURLTTL,
				); pErr == nil {
					presignedURL = purl
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, trackPlaybackDescriptor{
		TrackID:       track.ID,
		MediaObjectID: mo.ID,
		MIMEType:      mo.MIMEType,
		DurationMS:    track.DurationMS,
		BackendID:     mo.BackendID,
		BackendType:   backendType,
		ObjectKey:     mo.ObjectKey,
		PresignedURL:  presignedURL,
	})
}

func (handler *Handler) deleteTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	if err := handler.catalogService.DeleteTrack(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// relinkTrackRequest carries the new media object reference for a relink operation.
type relinkTrackRequest struct {
	MediaObjectID string `json:"mediaObjectId"`
}

func (handler *Handler) relinkTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req relinkTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	track, err := handler.catalogService.RelinkTrack(r.Context(), r.PathValue("id"), catalog.RelinkTrackRequest{
		MediaObjectID: req.MediaObjectID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, track)
}

// patchTrackRequest carries the fields that may be changed via PATCH.
type patchTrackRequest struct {
	Title       *string `json:"title"`
	SortTitle   *string `json:"sortTitle"`
	ArtistID    *string `json:"artistId"`
	AlbumID     *string `json:"albumId"`
	TrackNumber *int    `json:"trackNumber"`
	DiscNumber  *int    `json:"discNumber"`
	DurationMS  *int    `json:"durationMs"`
}

func (handler *Handler) patchTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req patchTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	track, err := handler.catalogService.UpdateTrack(r.Context(), r.PathValue("id"), catalog.UpdateTrackRequest{
		Title:       req.Title,
		SortTitle:   req.SortTitle,
		ArtistID:    req.ArtistID,
		AlbumID:     req.AlbumID,
		TrackNumber: req.TrackNumber,
		DiscNumber:  req.DiscNumber,
		DurationMS:  req.DurationMS,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, track)
}

// ---- playlist handlers ----

type createPlaylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type patchPlaylistRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type addPlaylistTrackRequest struct {
	TrackID string `json:"trackId"`
}

type setPlaylistTracksRequest struct {
	TrackIDs []string `json:"trackIds"`
}

func (handler *Handler) listPlaylists(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	pgResult, err := handler.catalogService.ListPlaylistsPage(r.Context(), catalog.ListQuery{
		SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: pgResult.Total, HasMore: offset+limit < pgResult.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"playlists": pgResult.Items, "pagination": meta})
}

func (handler *Handler) createPlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req createPlaylistRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	pl, err := handler.catalogService.CreatePlaylist(r.Context(), req.Name, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, pl)
}

func (handler *Handler) getPlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	pl, err := handler.catalogService.GetPlaylist(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

func (handler *Handler) deletePlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	if err := handler.catalogService.DeletePlaylist(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) patchPlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req patchPlaylistRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	pl, err := handler.catalogService.UpdatePlaylist(r.Context(), r.PathValue("id"), catalog.UpdatePlaylistRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

func (handler *Handler) addPlaylistTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req addPlaylistTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	pl, err := handler.catalogService.AddTrackToPlaylist(r.Context(), r.PathValue("id"), req.TrackID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

func (handler *Handler) removePlaylistTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	pl, err := handler.catalogService.RemoveTrackFromPlaylist(r.Context(), r.PathValue("id"), r.PathValue("trackId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

func (handler *Handler) setPlaylistTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req setPlaylistTracksRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.TrackIDs == nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackIds is required")
		return
	}
	pl, err := handler.catalogService.SetPlaylistTracks(r.Context(), r.PathValue("id"), req.TrackIDs)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

func (handler *Handler) getPlaylistTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	// Playlist tracks have a defined user-curated order; sortBy/sortOrder are not
	// exposed here — only limit/offset for pagination of the ordered list.
	limit, offset, ok := func() (int, int, bool) {
		q := r.URL.Query()
		lim := catalogListDefaultLimit
		off := 0
		if raw := q.Get("limit"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 1 {
				writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
				return 0, 0, false
			}
			if v > catalogListMaxLimit {
				v = catalogListMaxLimit
			}
			lim = v
		}
		if raw := q.Get("offset"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 0 {
				writeAPIError(w, http.StatusBadRequest, "invalid_offset", "offset must be a non-negative integer")
				return 0, 0, false
			}
			off = v
		}
		return lim, off, true
	}()
	if !ok {
		return
	}
	tracks, err := handler.catalogService.GetPlaylistTracks(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	page, meta := paginateCatalog(tracks, limit, offset)
	writeJSON(w, http.StatusOK, map[string]any{"tracks": page, "pagination": meta})
}

func (handler *Handler) getCatalogStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	stats, err := handler.catalogService.GetCatalogStats(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) getArtistStatsBreakdown(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	breakdown, err := handler.catalogService.GetArtistStatsBreakdown(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, breakdown)
}

func (handler *Handler) getAlbumStatsBreakdown(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	breakdown, err := handler.catalogService.GetAlbumStatsBreakdown(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, breakdown)
}

func (handler *Handler) getPlaylistStatsBreakdown(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	breakdown, err := handler.catalogService.GetPlaylistStatsBreakdown(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, breakdown)
}

// ---- playback history handlers ----

func (handler *Handler) requireHistoryService(w http.ResponseWriter) bool {
	if handler.historyService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "history_not_configured", "history service is not configured")
		return false
	}
	return true
}

type recordPlayRequest struct {
	TrackID  string `json:"trackId"`
	PlayedAt string `json:"playedAt,omitempty"` // RFC3339; defaults to server now when empty
}

func (handler *Handler) recordPlayEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	var req recordPlayRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.TrackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	var playedAt time.Time
	if req.PlayedAt != "" {
		t, err := time.Parse(time.RFC3339, req.PlayedAt)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "validation_error", "playedAt must be an RFC3339 timestamp")
			return
		}
		playedAt = t
	}
	event, err := handler.historyService.RecordPlay(r.Context(), user.ID, req.TrackID, playedAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

func (handler *Handler) listPlayEvents(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	q := r.URL.Query()
	limit := 0
	offset := 0
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return
		}
		limit = v
	}
	if raw := q.Get("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_offset", "offset must be a non-negative integer")
			return
		}
		offset = v
	}
	tf, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	asc, ok := parseHistoryOrder(w, r)
	if !ok {
		return
	}
	events, total, err := handler.historyService.ListPlays(r.Context(), history.PlayEventFilter{
		UserID:  user.ID,
		TrackID: q.Get("trackId"),
		Since:   tf.Since,
		Until:   tf.Until,
		Limit:   limit,
		Offset:  offset,
		Asc:     asc,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"hasMore": offset+limit < total && limit > 0,
		},
	})
}

func (handler *Handler) clearHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	if err := handler.historyService.ClearHistory(r.Context(), user.ID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) getMyHistoryStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	stats, err := handler.historyService.GetMyStats(r.Context(), history.UserStatsFilter{
		UserID: user.ID,
		Since:  f.Since,
		Until:  f.Until,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) getMyTopTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	tracks, err := handler.historyService.GetMyTopTracks(r.Context(), history.UserStatsFilter{
		UserID: user.ID,
		Since:  f.Since,
		Until:  f.Until,
	}, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": tracks})
}

func (handler *Handler) getMyHistoryTimeline(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	q := r.URL.Query()

	sinceRaw := q.Get("since")
	untilRaw := q.Get("until")
	if sinceRaw == "" || untilRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_time_bounds", "both since and until are required for timeline queries")
		return
	}

	since, err := time.Parse(time.RFC3339, sinceRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
		return
	}
	until, err := time.Parse(time.RFC3339, untilRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
		return
	}

	gran := history.TimelineGranularity(q.Get("granularity"))
	if gran == "" {
		gran = history.GranularityDay
	}
	switch gran {
	case history.GranularityDay, history.GranularityWeek, history.GranularityMonth:
		// valid
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_granularity", "granularity must be day, week, or month")
		return
	}

	buckets, err := handler.historyService.GetMyTimeline(r.Context(), history.TimelineFilter{
		Since:       since.UTC(),
		Until:       until.UTC(),
		Granularity: gran,
		UserID:      user.ID,
		TrackID:     q.Get("trackId"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"buckets": buckets})
}

// getMyHistorySummary returns combined stats and top-tracks for the authenticated
// viewer in one request.
func (handler *Handler) getMyHistorySummary(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	topN, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	sf := history.UserStatsFilter{UserID: user.ID, Since: f.Since, Until: f.Until}
	stats, err := handler.historyService.GetMyStats(r.Context(), sf)
	if err != nil {
		writeError(w, err)
		return
	}
	if topN <= 0 {
		topN = 10
	}
	tracks, err := handler.historyService.GetMyTopTracks(r.Context(), sf, topN)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"stats": stats, "topTracks": tracks})
}

// getMyTrackHistory returns the calling user's paginated play history for a
// specific track, identified by {trackId} in the path.
func (handler *Handler) getMyTrackHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	q := r.URL.Query()
	limit := 0
	offset := 0
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return
		}
		limit = v
	}
	if raw := q.Get("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_offset", "offset must be a non-negative integer")
			return
		}
		offset = v
	}
	tf, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	asc, ok := parseHistoryOrder(w, r)
	if !ok {
		return
	}
	events, total, err := handler.historyService.ListPlays(r.Context(), history.PlayEventFilter{
		UserID:  user.ID,
		TrackID: trackID,
		Since:   tf.Since,
		Until:   tf.Until,
		Limit:   limit,
		Offset:  offset,
		Asc:     asc,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"hasMore": offset+limit < total && limit > 0,
		},
	})
}

// getMyTrackStats returns aggregate play counts for the authenticated viewer on a specific track.
func (handler *Handler) getMyTrackStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	q := r.URL.Query()
	var sf history.UserStatsFilter
	if sinceRaw := q.Get("since"); sinceRaw != "" {
		t, err := time.Parse(time.RFC3339, sinceRaw)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
			return
		}
		sf.Since = t.UTC()
	}
	if untilRaw := q.Get("until"); untilRaw != "" {
		t, err := time.Parse(time.RFC3339, untilRaw)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
			return
		}
		sf.Until = t.UTC()
	}
	stats, err := handler.historyService.GetMyTrackStats(r.Context(), user.ID, trackID, sf)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// getMyTrackTimeline returns the calling user's play-event counts for a specific
// track grouped by time bucket. {trackId} is required in the path; since and
// until are required query params; granularity defaults to "day".
func (handler *Handler) getMyTrackTimeline(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	q := r.URL.Query()

	sinceRaw := q.Get("since")
	untilRaw := q.Get("until")
	if sinceRaw == "" || untilRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_time_bounds", "both since and until are required for timeline queries")
		return
	}
	since, err := time.Parse(time.RFC3339, sinceRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
		return
	}
	until, err := time.Parse(time.RFC3339, untilRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
		return
	}

	gran := history.TimelineGranularity(q.Get("granularity"))
	if gran == "" {
		gran = history.GranularityDay
	}
	switch gran {
	case history.GranularityDay, history.GranularityWeek, history.GranularityMonth:
		// valid
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_granularity", "granularity must be day, week, or month")
		return
	}

	buckets, err := handler.historyService.GetMyTimeline(r.Context(), history.TimelineFilter{
		Since:       since.UTC(),
		Until:       until.UTC(),
		Granularity: gran,
		UserID:      user.ID,
		TrackID:     trackID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"buckets": buckets})
}

// getMyTrackSummary returns the viewer's per-track play stats combined with their
// overall top tracks for cross-track context.
func (handler *Handler) getMyTrackSummary(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	topN, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	summary, err := handler.historyService.GetMyTrackSummary(r.Context(), user.ID, trackID, history.UserStatsFilter{
		UserID: user.ID,
		Since:  f.Since,
		Until:  f.Until,
	}, topN)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (handler *Handler) getAdminHistoryStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	stats, err := handler.historyService.GetHistoryStats(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) getAdminUserStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	stats, err := handler.historyService.GetAdminUserStats(r.Context(), history.UserStatsFilter{
		UserID: userID,
		Since:  f.Since,
		Until:  f.Until,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) getAdminUserTopTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	tracks, err := handler.historyService.GetAdminUserTopTracks(r.Context(), history.UserStatsFilter{
		UserID: userID,
		Since:  f.Since,
		Until:  f.Until,
	}, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": tracks})
}

// getAdminUserHistorySummary returns combined stats and top-tracks for a specific
// user in one request; intended for admin dashboard use.
func (handler *Handler) getAdminUserHistorySummary(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	topN, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	summary, err := handler.historyService.GetAdminUserSummary(r.Context(), userID, history.UserStatsFilter{
		Since: f.Since,
		Until: f.Until,
	}, topN)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (handler *Handler) getAdminTrackStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	stats, err := handler.historyService.GetTrackStats(r.Context(), history.TrackStatsFilter{
		TrackID: trackID,
		Since:   f.Since,
		Until:   f.Until,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) getAdminTrackTopListeners(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	users, err := handler.historyService.GetTrackTopListeners(r.Context(), history.TrackStatsFilter{
		TrackID: trackID,
		Since:   f.Since,
		Until:   f.Until,
	}, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

// getAdminTrackHistorySummary returns combined stats and top-listeners for a
// specific track in one request; intended for admin dashboard use.
func (handler *Handler) getAdminTrackHistorySummary(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	topN, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	summary, err := handler.historyService.GetTrackSummary(r.Context(), trackID, history.TrackStatsFilter{
		Since: f.Since,
		Until: f.Until,
	}, topN)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// getAdminHistorySummary returns combined system-wide aggregate stats, top tracks,
// and top users in one request; intended for admin dashboard use.
func (handler *Handler) getAdminHistorySummary(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	topN, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	summary, err := handler.historyService.GetGlobalSummary(r.Context(), f, topN)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (handler *Handler) getAdminTopTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	tracks, err := handler.historyService.GetTopTracks(r.Context(), f, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": tracks})
}

func (handler *Handler) getAdminTopUsers(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	users, err := handler.historyService.GetTopUsers(r.Context(), f, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

// parseHistoryAdminFilter parses the optional ?since and ?until query params (RFC3339).
// Returns 400 if either value is unparseable or if since >= until when both are present.
func parseHistoryAdminFilter(w http.ResponseWriter, r *http.Request) (history.StatsFilter, bool) {
	q := r.URL.Query()
	var f history.StatsFilter

	if raw := q.Get("since"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
			return history.StatsFilter{}, false
		}
		f.Since = t.UTC()
	}

	if raw := q.Get("until"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
			return history.StatsFilter{}, false
		}
		f.Until = t.UTC()
	}

	if !f.Since.IsZero() && !f.Until.IsZero() && !f.Since.Before(f.Until) {
		writeAPIError(w, http.StatusBadRequest, "invalid_time_range", "since must be before until")
		return history.StatsFilter{}, false
	}

	return f, true
}

// parseHistoryAdminLimit parses the optional ?limit query param (default 10, max 100).
func parseHistoryAdminLimit(w http.ResponseWriter, r *http.Request) (int, bool) {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return 0, true // service applies default
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
		return 0, false
	}
	return v, true
}

func (handler *Handler) getAdminUserHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	limit, offset, ok := parseHistoryAdminPagination(w, r)
	if !ok {
		return
	}
	tf, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	asc, ok := parseHistoryOrder(w, r)
	if !ok {
		return
	}
	events, total, err := handler.historyService.GetUserHistory(r.Context(), history.PlayEventFilter{
		UserID:  userID,
		TrackID: r.URL.Query().Get("trackId"),
		Since:   tf.Since,
		Until:   tf.Until,
		Limit:   limit,
		Offset:  offset,
		Asc:     asc,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"hasMore": limit > 0 && offset+limit < total,
		},
	})
}

func (handler *Handler) getAdminTrackHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	limit, offset, ok := parseHistoryAdminPagination(w, r)
	if !ok {
		return
	}
	tf, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	asc, ok := parseHistoryOrder(w, r)
	if !ok {
		return
	}
	events, total, err := handler.historyService.GetTrackHistory(r.Context(), history.AdminPlayEventFilter{
		TrackID: trackID,
		UserID:  r.URL.Query().Get("userId"),
		Since:   tf.Since,
		Until:   tf.Until,
		Limit:   limit,
		Offset:  offset,
		Asc:     asc,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"hasMore": limit > 0 && offset+limit < total,
		},
	})
}

func parseHistoryAdminPagination(w http.ResponseWriter, r *http.Request) (limit, offset int, ok bool) {
	q := r.URL.Query()
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return 0, 0, false
		}
		limit = v
	}
	if raw := q.Get("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_offset", "offset must be a non-negative integer")
			return 0, 0, false
		}
		offset = v
	}
	return limit, offset, true
}

// parseHistoryOrder parses the optional ?order=asc|desc query parameter.
// Returns true for ascending, false (default) for descending.
// Writes a 400 and returns ok=false for any value other than "asc" or "desc".
func parseHistoryOrder(w http.ResponseWriter, r *http.Request) (asc bool, ok bool) {
	raw := r.URL.Query().Get("order")
	switch raw {
	case "", "desc":
		return false, true
	case "asc":
		return true, true
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_order", `order must be "asc" or "desc"`)
		return false, false
	}
}

func (handler *Handler) deleteAdminUserHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	if err := handler.historyService.AdminDeleteUserHistory(r.Context(), userID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) deleteAdminTrackHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	if err := handler.historyService.AdminDeleteTrackHistory(r.Context(), trackID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) deleteAdminHistoryWindow(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	if f.Since.IsZero() && f.Until.IsZero() {
		writeAPIError(w, http.StatusBadRequest, "missing_time_filter", "at least one of since or until is required")
		return
	}
	if err := handler.historyService.AdminDeleteHistoryWindow(r.Context(), f); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) getAdminAllHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, offset, ok := parseHistoryAdminPagination(w, r)
	if !ok {
		return
	}
	asc, ok := parseHistoryOrder(w, r)
	if !ok {
		return
	}
	events, total, err := handler.historyService.GetAllHistory(r.Context(), history.GlobalPlayEventFilter{
		UserID:  r.URL.Query().Get("userId"),
		TrackID: r.URL.Query().Get("trackId"),
		Since:   f.Since,
		Until:   f.Until,
		Limit:   limit,
		Offset:  offset,
		Asc:     asc,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"hasMore": limit > 0 && offset+limit < total,
		},
	})
}

func (handler *Handler) getAdminHistoryTimeline(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	q := r.URL.Query()

	sinceRaw := q.Get("since")
	untilRaw := q.Get("until")
	if sinceRaw == "" || untilRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_time_bounds", "both since and until are required for timeline queries")
		return
	}

	since, err := time.Parse(time.RFC3339, sinceRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
		return
	}
	until, err := time.Parse(time.RFC3339, untilRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
		return
	}

	gran := history.TimelineGranularity(q.Get("granularity"))
	if gran == "" {
		gran = history.GranularityDay
	}
	switch gran {
	case history.GranularityDay, history.GranularityWeek, history.GranularityMonth:
		// valid
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_granularity", "granularity must be day, week, or month")
		return
	}

	buckets, err := handler.historyService.GetHistoryTimeline(r.Context(), history.TimelineFilter{
		Since:       since.UTC(),
		Until:       until.UTC(),
		Granularity: gran,
		UserID:      q.Get("userId"),
		TrackID:     q.Get("trackId"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"buckets": buckets})
}

// getAdminUserTimeline returns play-event counts for a specific user grouped by
// time bucket. {userId} is required in the path; since and until are required
// query params; granularity defaults to "day".
func (handler *Handler) getAdminUserTimeline(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	q := r.URL.Query()

	sinceRaw := q.Get("since")
	untilRaw := q.Get("until")
	if sinceRaw == "" || untilRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_time_bounds", "both since and until are required for timeline queries")
		return
	}
	since, err := time.Parse(time.RFC3339, sinceRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
		return
	}
	until, err := time.Parse(time.RFC3339, untilRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
		return
	}

	gran := history.TimelineGranularity(q.Get("granularity"))
	if gran == "" {
		gran = history.GranularityDay
	}
	switch gran {
	case history.GranularityDay, history.GranularityWeek, history.GranularityMonth:
		// valid
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_granularity", "granularity must be day, week, or month")
		return
	}

	buckets, err := handler.historyService.GetHistoryTimeline(r.Context(), history.TimelineFilter{
		Since:       since.UTC(),
		Until:       until.UTC(),
		Granularity: gran,
		UserID:      userID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"buckets": buckets})
}

// getAdminTrackTimeline returns play-event counts for a specific track grouped by
// time bucket. {trackId} is required in the path; since and until are required
// query params; granularity defaults to "day".
func (handler *Handler) getAdminTrackTimeline(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	q := r.URL.Query()

	sinceRaw := q.Get("since")
	untilRaw := q.Get("until")
	if sinceRaw == "" || untilRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_time_bounds", "both since and until are required for timeline queries")
		return
	}
	since, err := time.Parse(time.RFC3339, sinceRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
		return
	}
	until, err := time.Parse(time.RFC3339, untilRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
		return
	}

	gran := history.TimelineGranularity(q.Get("granularity"))
	if gran == "" {
		gran = history.GranularityDay
	}
	switch gran {
	case history.GranularityDay, history.GranularityWeek, history.GranularityMonth:
		// valid
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_granularity", "granularity must be day, week, or month")
		return
	}

	buckets, err := handler.historyService.GetHistoryTimeline(r.Context(), history.TimelineFilter{
		Since:       since.UTC(),
		Until:       until.UTC(),
		Granularity: gran,
		TrackID:     trackID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"buckets": buckets})
}

func (handler *Handler) getAdminEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	e, err := handler.historyService.GetEventByID(r.Context(), eventID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (handler *Handler) deleteAdminEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	if err := handler.historyService.DeleteEventByID(r.Context(), eventID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) getMyEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	e, err := handler.historyService.GetMyEvent(r.Context(), user.ID, eventID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (handler *Handler) deleteMyEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	if err := handler.historyService.DeleteMyEvent(r.Context(), user.ID, eventID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) patchAdminEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	var body struct {
		PlayedAt string `json:"playedAt"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, err)
		return
	}
	if body.PlayedAt == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_played_at", "playedAt is required")
		return
	}
	t, err := time.Parse(time.RFC3339, body.PlayedAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_played_at", "playedAt must be an RFC3339 timestamp")
		return
	}
	e, err := handler.historyService.UpdateEventByID(r.Context(), eventID, t)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (handler *Handler) patchMyEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	var body struct {
		PlayedAt string `json:"playedAt"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, err)
		return
	}
	if body.PlayedAt == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_played_at", "playedAt is required")
		return
	}
	t, err := time.Parse(time.RFC3339, body.PlayedAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_played_at", "playedAt must be an RFC3339 timestamp")
		return
	}
	e, err := handler.historyService.UpdateMyEvent(r.Context(), user.ID, eventID, t)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (handler *Handler) batchDeleteAdminEvents(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, err)
		return
	}
	if len(body.IDs) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_ids", "ids must be a non-empty array")
		return
	}
	if len(body.IDs) > history.MaxBatchDeleteIDs {
		writeAPIError(w, http.StatusBadRequest, "invalid_ids",
			fmt.Sprintf("ids must not exceed %d entries", history.MaxBatchDeleteIDs))
		return
	}
	deleted, err := handler.historyService.BatchDeleteEvents(r.Context(), body.IDs)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
}

func (handler *Handler) batchDeleteMyEvents(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, err)
		return
	}
	if len(body.IDs) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_ids", "ids must be a non-empty array")
		return
	}
	if len(body.IDs) > history.MaxBatchDeleteIDs {
		writeAPIError(w, http.StatusBadRequest, "invalid_ids",
			fmt.Sprintf("ids must not exceed %d entries", history.MaxBatchDeleteIDs))
		return
	}
	deleted, err := handler.historyService.BatchDeleteMyEvents(r.Context(), user.ID, body.IDs)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
}

func (handler *Handler) getRecentlyAdded(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	kind, limit, ok := parseRecentCatalogQuery(w, r)
	if !ok {
		return
	}
	result, err := handler.catalogService.GetRecentlyAdded(r.Context(), kind, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (handler *Handler) getRecentlyUpdated(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	kind, limit, ok := parseRecentCatalogQuery(w, r)
	if !ok {
		return
	}
	result, err := handler.catalogService.GetRecentlyUpdated(r.Context(), kind, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func parseRecentCatalogQuery(w http.ResponseWriter, r *http.Request) (string, int, bool) {
	q := r.URL.Query()
	kind := q.Get("kind")
	limit := 0
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return "", 0, false
		}
		limit = v
	}
	return kind, limit, true
}

// requireViewerAuth allows any session-authenticated user (admin or viewer role).
// The static bootstrap adminToken is intentionally NOT accepted here — catalog
// browse requires a real session. Returns 503 when no auth service is configured,
// 401 when no valid bearer token is supplied.
func (handler *Handler) requireViewerAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler.authService == nil {
			writeAPIError(w, http.StatusServiceUnavailable, "auth_not_configured", "authentication service is not configured")
			return
		}
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			w.Header().Set("WWW-Authenticate", `Bearer realm="inori"`)
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
			return
		}
		user, err := handler.authService.ValidateToken(r.Context(), token)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Bearer realm="inori"`)
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), contextKeyUser, user)))
	}
}

func (handler *Handler) requireAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Reject immediately when no authentication method is available.
		if handler.authService == nil && handler.adminToken == "" {
			writeAPIError(w, http.StatusServiceUnavailable, "admin_auth_not_configured", "administrator token is not configured")
			return
		}

		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			w.Header().Set("WWW-Authenticate", `Bearer realm="inori-admin"`)
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid administrator bearer token is required")
			return
		}

		// Session token path: validate against auth service when available.
		if handler.authService != nil {
			user, err := handler.authService.ValidateToken(r.Context(), token)
			if err == nil {
				if user.Role != auth.RoleAdmin {
					writeAPIError(w, http.StatusForbidden, "unauthorized", "administrator role required")
					return
				}
				next(w, r.WithContext(context.WithValue(r.Context(), contextKeyUser, user)))
				return
			}
			// Fall through to static token check to allow bootstrap token.
		}

		// Static bootstrap token fallback.
		if handler.adminToken == "" {
			w.Header().Set("WWW-Authenticate", `Bearer realm="inori-admin"`)
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid administrator bearer token is required")
			return
		}
		if !constantTimeTokenEqual(token, handler.adminToken) {
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

func (handler *Handler) getMediaObjectDuplicates(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	minCopies, err := parseMediaObjectMinCopies(r.URL.Query().Get("minCopies"))
	if err != nil {
		writeError(w, err)
		return
	}
	report, err := handler.mediaObjects.GetMediaObjectDuplicates(r.Context(), r.URL.Query().Get("backendId"), minCopies)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
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

func (handler *Handler) setMediaObjectsLifecycle(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	var request mediaObjectBulkLifecycleRequest
	if err := decodeJSONWithSentinel(w, r, &request, storage.ErrInvalidMediaObject); err != nil {
		writeError(w, err)
		return
	}
	report, err := handler.mediaObjects.SetMediaObjectLifecycleStateByFilterWithOptions(r.Context(), request.Filter.filter(), request.LifecycleState, storage.MediaObjectLifecycleUpdateOptions{DryRun: request.DryRun})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (handler *Handler) setMediaObjectLifecycle(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	var request mediaObjectLifecycleRequest
	if err := decodeJSONWithSentinel(w, r, &request, storage.ErrInvalidMediaObject); err != nil {
		writeError(w, err)
		return
	}
	object, err := handler.mediaObjects.SetMediaObjectLifecycleState(r.Context(), r.PathValue("id"), request.LifecycleState)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, object)
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

func (handler *Handler) getMediaObjectTimeline(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	timeline, err := handler.mediaObjects.GetMediaObjectTimeline(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, timeline)
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
		LifecycleState:     r.URL.Query().Get("lifecycleState"),
		AssetKind:          r.URL.Query().Get("assetKind"),
		SortBy:             r.URL.Query().Get("sortBy"),
		SortOrder:          r.URL.Query().Get("sortOrder"),
		Limit:              limit,
		Offset:             offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func parseMediaObjectMinCopies(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 2, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%w: minCopies must be an integer", storage.ErrInvalidMediaObject)
	}
	if value < 2 {
		return 0, fmt.Errorf("%w: minCopies must be at least 2", storage.ErrInvalidMediaObject)
	}
	return value, nil
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
	case errors.Is(err, auth.ErrInvalidUser):
		status = http.StatusBadRequest
		code = "invalid_user"
	case errors.Is(err, auth.ErrUserNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, auth.ErrUserConflict):
		status = http.StatusConflict
		code = "conflict"
	case errors.Is(err, auth.ErrUserDisabled):
		status = http.StatusForbidden
		code = "user_disabled"
	case errors.Is(err, auth.ErrBadCredentials):
		status = http.StatusUnauthorized
		code = "unauthorized"
	case errors.Is(err, catalog.ErrInvalidArtist), errors.Is(err, catalog.ErrInvalidAlbum), errors.Is(err, catalog.ErrInvalidTrack),
		errors.Is(err, catalog.ErrInvalidPlaylist):
		status = http.StatusBadRequest
		code = "invalid_catalog_entity"
	case errors.Is(err, catalog.ErrArtistNotFound), errors.Is(err, catalog.ErrAlbumNotFound), errors.Is(err, catalog.ErrTrackNotFound),
		errors.Is(err, catalog.ErrPlaylistNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, catalog.ErrArtistConflict), errors.Is(err, catalog.ErrAlbumConflict), errors.Is(err, catalog.ErrTrackConflict):
		status = http.StatusConflict
		code = "conflict"
	case errors.Is(err, catalog.ErrImportRejected):
		status = http.StatusUnprocessableEntity
		code = "import_rejected"
	case errors.Is(err, catalog.ErrRelinkRejected):
		status = http.StatusUnprocessableEntity
		code = "relink_rejected"
	case errors.Is(err, storage.ErrPlaybackUnavailable):
		status = http.StatusUnprocessableEntity
		code = "playback_unavailable"
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
	case errors.Is(err, history.ErrEventNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, history.ErrEventForbidden):
		status = http.StatusForbidden
		code = "event_forbidden"
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

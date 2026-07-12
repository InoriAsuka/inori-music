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
	"inori-music/services/api/internal/favorites"
	"inori-music/services/api/internal/history"
	"inori-music/services/api/internal/ratelimit"
	"inori-music/services/api/internal/search"
	"inori-music/services/api/internal/storage"
	"inori-music/services/api/internal/streamsign"
	"inori-music/services/api/internal/userplaylist"
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

// WithFavoritesService enables user favorites routes.
func WithFavoritesService(svc *favorites.Service) HandlerOption {
	return func(handler *Handler) {
		handler.favoritesService = svc
	}
}

// WithUserPlaylistService enables user playlist routes.
func WithUserPlaylistService(svc *userplaylist.Service) HandlerOption {
	return func(handler *Handler) {
		handler.userPlaylistService = svc
	}
}

// WithSearchService attaches an optional search backend; if nil the handler falls
// back to catalogService.SearchCatalog for every search request.
func WithSearchService(svc search.Service) HandlerOption {
	return func(handler *Handler) {
		handler.searchSvc = svc
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

// WithLoginLimiter attaches a rate limiter for failed login attempts.
func WithLoginLimiter(l *ratelimit.Limiter) HandlerOption {
	return func(handler *Handler) {
		handler.loginLimiter = l
	}
}

// WithStreamSigner attaches an HMAC signer for server-proxied stream URLs.
func WithStreamSigner(s *streamsign.Signer) HandlerOption {
	return func(handler *Handler) {
		handler.streamSigner = s
	}
}

// WithCORSOrigins sets the list of allowed CORS origins. When the list is empty
// the middleware reflects any request Origin back (permissive development mode).
func WithCORSOrigins(origins []string) HandlerOption {
	return func(handler *Handler) {
		cp := make([]string, len(origins))
		copy(cp, origins)
		handler.corsOrigins = cp
	}
}

// Handler serves versioned administrative HTTP endpoints.
type Handler struct {
	storage             *storage.Service
	mediaObjects        *storage.MediaObjectService
	authService         *auth.Service
	catalogService      *catalog.Service
	historyService      *history.Service
	favoritesService    *favorites.Service
	userPlaylistService *userplaylist.Service
	searchSvc           search.Service
	loginLimiter        *ratelimit.Limiter
	streamSigner        *streamsign.Signer
	adminToken          string
	corsOrigins         []string
	info                ServiceInfo
	metricsMu           sync.Mutex
	requestMetrics      map[requestMetricKey]requestMetricValue
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
		readinessCheck("catalog_service", handler.catalogService != nil, "catalog service is configured", "catalog service is not configured"),
		readinessCheck("history_service", handler.historyService != nil, "history service is configured", "history service is not configured"),
		readinessCheck("favorites_service", handler.favoritesService != nil, "favorites service is configured", "favorites service is not configured"),
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
	mux.HandleFunc("GET /api/v1/catalog/albums/{id}/artwork", handler.requireViewerAuth(handler.getAlbumArtwork))
	mux.HandleFunc("GET /api/v1/catalog/albums/{id}/tracks", handler.requireViewerAuth(handler.listTracksByAlbum))
	mux.HandleFunc("GET /api/v1/catalog/tracks", handler.requireViewerAuth(handler.listTracks))
	mux.HandleFunc("GET /api/v1/catalog/tracks/{id}", handler.requireViewerAuth(handler.getTrack))
	mux.HandleFunc("GET /api/v1/catalog/tracks/{id}/playback", handler.requireViewerAuth(handler.getTrackPlayback))
	mux.HandleFunc("GET /api/v1/catalog/tracks/{id}/stream", handler.streamTrack)
	mux.HandleFunc("POST /api/v1/catalog/tracks/{id}/lyrics", handler.requireAdminAuth(handler.uploadTrackLyrics))
	mux.HandleFunc("GET /api/v1/catalog/tracks/{id}/lyrics", handler.requireViewerAuth(handler.getTrackLyrics))
	mux.HandleFunc("DELETE /api/v1/catalog/tracks/{id}/lyrics", handler.requireAdminAuth(handler.deleteTrackLyrics))
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
	mux.HandleFunc("POST /api/v1/me/favorites/tracks/{trackId}", handler.requireViewerAuth(handler.addFavoriteTrack))
	mux.HandleFunc("DELETE /api/v1/me/favorites/tracks/{trackId}", handler.requireViewerAuth(handler.removeFavoriteTrack))
	mux.HandleFunc("GET /api/v1/me/favorites/tracks", handler.requireViewerAuth(handler.listFavoriteTracks))
	mux.HandleFunc("/api/v1/me/favorites/tracks/{trackId}", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/me/favorites/tracks", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("GET /api/v1/admin/favorites/users/{userId}/tracks", handler.requireAdminAuth(handler.adminListUserFavorites))
	mux.HandleFunc("DELETE /api/v1/admin/favorites/users/{userId}/tracks", handler.requireAdminAuth(handler.adminClearUserFavorites))
	mux.HandleFunc("DELETE /api/v1/admin/favorites/users/{userId}/tracks/{trackId}", handler.requireAdminAuth(handler.adminRemoveUserFavoriteTrack))
	mux.HandleFunc("/api/v1/admin/favorites/users/{userId}/tracks/{trackId}", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/favorites/users/{userId}/tracks", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("POST /api/v1/me/playlists", handler.requireViewerAuth(handler.createUserPlaylist))
	mux.HandleFunc("GET /api/v1/me/playlists", handler.requireViewerAuth(handler.listUserPlaylists))
	mux.HandleFunc("GET /api/v1/me/playlists/{id}", handler.requireViewerAuth(handler.getUserPlaylist))
	mux.HandleFunc("PATCH /api/v1/me/playlists/{id}", handler.requireViewerAuth(handler.patchUserPlaylist))
	mux.HandleFunc("DELETE /api/v1/me/playlists/{id}", handler.requireViewerAuth(handler.deleteUserPlaylist))
	mux.HandleFunc("POST /api/v1/me/playlists/{id}/tracks", handler.requireViewerAuth(handler.addUserPlaylistTrack))
	mux.HandleFunc("DELETE /api/v1/me/playlists/{id}/tracks/{trackId}", handler.requireViewerAuth(handler.removeUserPlaylistTrack))
	mux.HandleFunc("GET /api/v1/me/playlists/{id}/tracks", handler.requireViewerAuth(handler.getUserPlaylistTracks))
	mux.HandleFunc("PUT /api/v1/me/playlists/{id}/tracks", handler.requireViewerAuth(handler.setUserPlaylistTracks))
	mux.HandleFunc("/api/v1/me/playlists/{id}/tracks/{trackId}", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/me/playlists/{id}/tracks", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/me/playlists/{id}", handler.requireViewerAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/me/playlists", handler.requireViewerAuth(handler.methodNotAllowed))
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
	mux.HandleFunc("GET /api/v1/admin/storage/backends/{id}", handler.requireAdminAuth(handler.getStorageBackend))
	mux.HandleFunc("POST /api/v1/admin/storage/backends/{id}/default", handler.requireAdminAuth(handler.setDefaultStorageBackend))
	mux.HandleFunc("POST /api/v1/admin/storage/backends/{id}/disable", handler.requireAdminAuth(handler.disableStorageBackend))
	mux.HandleFunc("POST /api/v1/admin/storage/backends/{id}/enable", handler.requireAdminAuth(handler.enableStorageBackend))
	mux.HandleFunc("PATCH /api/v1/admin/storage/backends/{id}", handler.requireAdminAuth(handler.patchStorageBackend))
	mux.HandleFunc("DELETE /api/v1/admin/storage/backends/{id}", handler.requireAdminAuth(handler.deleteStorageBackend))
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
	mux.HandleFunc("PATCH /api/v1/admin/media/objects/{id}", handler.requireAdminAuth(handler.patchMediaObject))
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
	mux.HandleFunc("GET /api/v1/admin/storage/backends/validate", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("DELETE /api/v1/admin/storage/backends/validate", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("GET /api/v1/admin/storage/backends/refresh", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("DELETE /api/v1/admin/storage/backends/refresh", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/storage/backends/{id}/default", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/storage/backends/{id}/disable", handler.requireAdminAuth(handler.methodNotAllowed))
	mux.HandleFunc("/api/v1/admin/storage/backends/{id}/enable", handler.requireAdminAuth(handler.methodNotAllowed))
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
	return requestIDMiddleware()(corsMiddleware(handler.corsOrigins)(handler.instrument(mux)))
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
	case errors.Is(err, storage.ErrBackendIsDefault):
		status = http.StatusConflict
		code = "storage_backend_is_default"
	case errors.Is(err, storage.ErrBackendInUse):
		status = http.StatusConflict
		code = "storage_backend_in_use"
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
	case errors.Is(err, userplaylist.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, userplaylist.ErrInvalidInput):
		status = http.StatusBadRequest
		code = "invalid_input"
	case errors.Is(err, userplaylist.ErrForbidden):
		status = http.StatusForbidden
		code = "forbidden"
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

package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/favorites"
	"inori-music/services/api/internal/history"
	"inori-music/services/api/internal/storage"
)

func historyNewService() *history.Service {
	return history.NewService(history.NewMemoryRepository())
}

const testAdminToken = "test-admin-token"

func TestHealthIsPublic(t *testing.T) {
	for _, handler := range []http.Handler{newTestHandler(), newUnauthenticatedTestHandler()} {
		response := performRequestWithoutAuth(t, handler, http.MethodGet, "/healthz", "")
		if response.Code != http.StatusOK {
			t.Fatalf("GET /healthz status = %d, want %d", response.Code, http.StatusOK)
		}
		assertJSONField(t, response, "status", "ok")
	}
}

func TestVersionIsPublic(t *testing.T) {
	handler := newTestHandler()
	response := performRequestWithoutAuth(t, handler, http.MethodGet, "/versionz", "")
	if response.Code != http.StatusOK {
		t.Fatalf("GET /versionz status = %d, want %d", response.Code, http.StatusOK)
	}
	assertJSONField(t, response, "name", "inori-api")
	assertJSONField(t, response, "version", "test-version")
	assertJSONField(t, response, "commit", "test-commit")
	assertJSONField(t, response, "buildTime", "2026-06-05T12:30:00Z")

	defaultResponse := performRequestWithoutAuth(t, newUnauthenticatedTestHandler(), http.MethodGet, "/versionz", "")
	if defaultResponse.Code != http.StatusOK {
		t.Fatalf("default GET /versionz status = %d, want %d", defaultResponse.Code, http.StatusOK)
	}
	assertJSONField(t, defaultResponse, "version", "dev")
}

func TestReadinessIsPublic(t *testing.T) {
	ready := performRequestWithoutAuth(t, newTestHandler(), http.MethodGet, "/readyz", "")
	if ready.Code != http.StatusOK {
		t.Fatalf("GET /readyz status = %d, want %d body = %s", ready.Code, http.StatusOK, ready.Body.String())
	}
	var readyReport ReadinessReport
	decodeResponse(t, ready, &readyReport)
	if !readyReport.Ready || len(readyReport.Checks) != 6 {
		t.Fatalf("ready report = %+v, want ready report with six checks", readyReport)
	}

	unready := performRequestWithoutAuth(t, newUnauthenticatedTestHandler(), http.MethodGet, "/readyz", "")
	if unready.Code != http.StatusServiceUnavailable {
		t.Fatalf("unready GET /readyz status = %d, want %d body = %s", unready.Code, http.StatusServiceUnavailable, unready.Body.String())
	}
	var unreadyReport ReadinessReport
	decodeResponse(t, unready, &unreadyReport)
	if unreadyReport.Ready || len(unreadyReport.Checks) != 6 {
		t.Fatalf("unready report = %+v, want not ready report with six checks", unreadyReport)
	}
	failed := make(map[string]bool)
	for _, check := range unreadyReport.Checks {
		if check.Status == "failed" {
			failed[check.Name] = true
		}
	}
	if !failed["media_registry"] || !failed["admin_auth"] || !failed["catalog_service"] || !failed["history_service"] {
		t.Fatalf("failed checks = %+v, want media_registry, admin_auth, catalog_service, and history_service failures", failed)
	}
}

func TestMetricsIsPublic(t *testing.T) {
	handler := newTestHandler()
	performRequestWithoutAuth(t, handler, http.MethodGet, "/healthz", "")
	response := performRequestWithoutAuth(t, handler, http.MethodGet, "/metrics", "")
	if response.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want %d body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	contentType := response.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		t.Fatalf("Content-Type = %q, want text/plain", contentType)
	}
	body := response.Body.String()
	for _, want := range []string{
		"# HELP inori_api_ready",
		"inori_api_ready 1",
		`inori_api_readiness_check{check="storage_service"} 1`,
		`inori_api_info{name="inori-api",version="test-version",commit="test-commit",build_time="2026-06-05T12:30:00Z"} 1`,
		`inori_api_http_requests_total{method="GET",path="GET /healthz",status="200"} 1`,
		`inori_api_http_request_duration_seconds_sum{method="GET",path="GET /healthz",status="200"}`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics body missing %q: %s", want, body)
		}
	}

	unready := performRequestWithoutAuth(t, newUnauthenticatedTestHandler(), http.MethodGet, "/metrics", "")
	if unready.Code != http.StatusOK || !strings.Contains(unready.Body.String(), "inori_api_ready 0") {
		t.Fatalf("unready metrics status/body = %d %s, want ready gauge 0", unready.Code, unready.Body.String())
	}
}

func TestAdminAuth(t *testing.T) {
	handler := newTestHandler()
	tests := []struct {
		name   string
		header string
		status int
		code   string
	}{
		{name: "missing", status: http.StatusUnauthorized, code: "unauthorized"},
		{name: "malformed", header: "Basic abc", status: http.StatusUnauthorized, code: "unauthorized"},
		{name: "invalid", header: "Bearer wrong-token", status: http.StatusUnauthorized, code: "unauthorized"},
		{name: "case insensitive scheme", header: "bearer " + testAdminToken, status: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performRequestWithAuthHeader(t, handler, http.MethodGet, "/api/v1/admin/storage/backends", "", tt.header)
			if tt.code != "" {
				assertAPIError(t, response, tt.status, tt.code)
				if response.Header().Get("WWW-Authenticate") == "" {
					t.Fatal("WWW-Authenticate header should be set for unauthorized responses")
				}
				return
			}
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.status, response.Body.String())
			}
		})
	}
}

func TestAdminAuthFailsClosedWhenTokenIsNotConfigured(t *testing.T) {
	response := performRequestWithoutAuth(t, newUnauthenticatedTestHandler(), http.MethodGet, "/api/v1/admin/storage/backends", "")
	assertAPIError(t, response, http.StatusServiceUnavailable, "admin_auth_not_configured")
}

func TestUnknownAdminRouteRequiresAuth(t *testing.T) {
	response := performRequestWithoutAuth(t, newTestHandler(), http.MethodGet, "/api/v1/admin/missing", "")
	assertAPIError(t, response, http.StatusUnauthorized, "unauthorized")
}

func newTestHandler() http.Handler {
	repository := storage.NewMemoryRepository()
	return NewHandler(
		storage.NewService(repository),
		WithAdminToken(testAdminToken),
		WithMediaObjectService(storage.NewMediaObjectService(repository, storage.NewMemoryMediaObjectRepository())),
		WithCatalogService(catalog.NewService(catalog.NewMemoryRepository())),
		WithHistoryService(history.NewService(history.NewMemoryRepository())),
		WithFavoritesService(favorites.NewService(favorites.NewMemoryRepository())),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test-version", Commit: "test-commit", BuildTime: "2026-06-05T12:30:00Z"}),
	).Routes()
}

func newUnauthenticatedTestHandler() http.Handler {
	return NewHandler(storage.NewService(storage.NewMemoryRepository())).Routes()
}

// newNoCatalogTestHandler returns a fully authenticated handler with all services
// except catalog, to verify catalog_not_configured 503 responses.
func newNoCatalogTestHandler() http.Handler {
	repository := storage.NewMemoryRepository()
	return NewHandler(
		storage.NewService(repository),
		WithAdminToken(testAdminToken),
		WithMediaObjectService(storage.NewMediaObjectService(repository, storage.NewMemoryMediaObjectRepository())),
		WithHistoryService(history.NewService(history.NewMemoryRepository())),
	).Routes()
}

// newNoHistoryTestHandler returns a fully authenticated handler with all services
// except history, to verify history_not_configured 503 responses.
func newNoHistoryTestHandler() http.Handler {
	repository := storage.NewMemoryRepository()
	return NewHandler(
		storage.NewService(repository),
		WithAdminToken(testAdminToken),
		WithMediaObjectService(storage.NewMediaObjectService(repository, storage.NewMemoryMediaObjectRepository())),
		WithCatalogService(catalog.NewService(catalog.NewMemoryRepository())),
	).Routes()
}

func performRequest(t *testing.T, handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	return performRequestWithContentType(t, handler, method, path, body, true)
}

func performRequestWithContentType(t *testing.T, handler http.Handler, method string, path string, body string, includeContentType bool) *httptest.ResponseRecorder {
	t.Helper()
	return performRequestWithAuthHeaderAndContentType(t, handler, method, path, body, "Bearer "+testAdminToken, includeContentType)
}

func performRequestWithoutAuth(t *testing.T, handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	return performRequestWithAuthHeaderAndContentType(t, handler, method, path, body, "", true)
}

func performRequestWithAuthHeader(t *testing.T, handler http.Handler, method string, path string, body string, authHeader string) *httptest.ResponseRecorder {
	t.Helper()
	return performRequestWithAuthHeaderAndContentType(t, handler, method, path, body, authHeader, true)
}

func performRequestWithAuthHeaderAndContentType(t *testing.T, handler http.Handler, method string, path string, body string, authHeader string, includeContentType bool) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if authHeader != "" {
		request.Header.Set("Authorization", authHeader)
	}
	if body != "" && includeContentType {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func assertAPIError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, status, response.Body.String())
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeResponse(t, response, &body)
	if body.Error.Code != code {
		t.Fatalf("error code = %q, want %q", body.Error.Code, code)
	}
}

func assertJSONField(t *testing.T, response *httptest.ResponseRecorder, field string, want any) {
	t.Helper()
	var body map[string]any
	decodeResponse(t, response, &body)
	if body[field] != want {
		t.Fatalf("field %q = %#v, want %#v", field, body[field], want)
	}
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v, body = %s", err, response.Body.String())
	}
}

// ---- Phase 121: readiness check coverage for catalog and history services ----

func TestReadinessAllConfigured(t *testing.T) {
	// newTestHandler now includes catalog and history services.
	resp := performRequestWithoutAuth(t, newTestHandler(), http.MethodGet, "/readyz", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /readyz status = %d, want 200; body = %s", resp.Code, resp.Body.String())
	}
	var report ReadinessReport
	decodeResponse(t, resp, &report)
	if !report.Ready {
		t.Fatalf("expected ready=true, got false; checks = %+v", report.Checks)
	}
	names := make(map[string]string)
	for _, c := range report.Checks {
		names[c.Name] = c.Status
	}
	for _, want := range []string{"storage_service", "media_registry", "admin_auth", "catalog_service", "history_service", "favorites_service"} {
		if names[want] != "ok" {
			t.Errorf("check %q = %q, want \"ok\"", want, names[want])
		}
	}
}

func TestReadinessMissingCatalog(t *testing.T) {
	repository := storage.NewMemoryRepository()
	h := NewHandler(
		storage.NewService(repository),
		WithAdminToken(testAdminToken),
		WithMediaObjectService(storage.NewMediaObjectService(repository, storage.NewMemoryMediaObjectRepository())),
		WithHistoryService(history.NewService(history.NewMemoryRepository())),
		// catalog service intentionally omitted
	).Routes()
	resp := performRequestWithoutAuth(t, h, http.MethodGet, "/readyz", "")
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz status = %d, want 503; body = %s", resp.Code, resp.Body.String())
	}
	var report ReadinessReport
	decodeResponse(t, resp, &report)
	if report.Ready {
		t.Fatalf("expected ready=false, got true")
	}
	found := false
	for _, c := range report.Checks {
		if c.Name == "catalog_service" && c.Status == "failed" {
			found = true
		}
	}
	if !found {
		t.Errorf("catalog_service check not failed; checks = %+v", report.Checks)
	}
}

func TestReadinessMissingHistory(t *testing.T) {
	repository := storage.NewMemoryRepository()
	h := NewHandler(
		storage.NewService(repository),
		WithAdminToken(testAdminToken),
		WithMediaObjectService(storage.NewMediaObjectService(repository, storage.NewMemoryMediaObjectRepository())),
		WithCatalogService(catalog.NewService(catalog.NewMemoryRepository())),
		// history service intentionally omitted
	).Routes()
	resp := performRequestWithoutAuth(t, h, http.MethodGet, "/readyz", "")
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz status = %d, want 503; body = %s", resp.Code, resp.Body.String())
	}
	var report ReadinessReport
	decodeResponse(t, resp, &report)
	if report.Ready {
		t.Fatalf("expected ready=false, got true")
	}
	found := false
	for _, c := range report.Checks {
		if c.Name == "history_service" && c.Status == "failed" {
			found = true
		}
	}
	if !found {
		t.Errorf("history_service check not failed; checks = %+v", report.Checks)
	}
}

func TestReadinessMissingAdminAuth(t *testing.T) {
	repository := storage.NewMemoryRepository()
	h := NewHandler(
		storage.NewService(repository),
		// admin token intentionally omitted
		WithMediaObjectService(storage.NewMediaObjectService(repository, storage.NewMemoryMediaObjectRepository())),
		WithCatalogService(catalog.NewService(catalog.NewMemoryRepository())),
		WithHistoryService(history.NewService(history.NewMemoryRepository())),
	).Routes()
	resp := performRequestWithoutAuth(t, h, http.MethodGet, "/readyz", "")
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz status = %d, want 503; body = %s", resp.Code, resp.Body.String())
	}
	var report ReadinessReport
	decodeResponse(t, resp, &report)
	if report.Ready {
		t.Fatalf("expected ready=false, got true")
	}
	found := false
	for _, c := range report.Checks {
		if c.Name == "admin_auth" && c.Status == "failed" {
			found = true
		}
	}
	if !found {
		t.Errorf("admin_auth check not failed; checks = %+v", report.Checks)
	}
}

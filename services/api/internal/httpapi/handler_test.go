package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"inori-music/services/api/internal/storage"
)

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

func TestStorageBackendWorkflow(t *testing.T) {
	handler := newTestHandler()
	local := `{"id":"local-main","type":"local","displayName":"Local Media","enabled":true,"isDefault":true,"config":{"local":{"rootPath":"/srv/inori/media"}}}`
	s3 := `{"id":"s3-prod","type":"s3","displayName":"S3","enabled":true,"config":{"s3":{"endpoint":"https://s3.example.com","bucket":"inori","accessKeySecretRef":"S3_ACCESS","secretKeySecretRef":"S3_SECRET"}}}`

	validated := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends/validate", local)
	if validated.Code != http.StatusOK {
		t.Fatalf("validate status = %d body = %s", validated.Code, validated.Body.String())
	}
	assertJSONField(t, validated, "healthStatus", "unknown")

	registeredLocal := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", local)
	if registeredLocal.Code != http.StatusCreated {
		t.Fatalf("register local status = %d body = %s", registeredLocal.Code, registeredLocal.Body.String())
	}
	registeredS3 := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", s3)
	if registeredS3.Code != http.StatusCreated {
		t.Fatalf("register s3 status = %d body = %s", registeredS3.Code, registeredS3.Body.String())
	}

	selected := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends/s3-prod/default", "")
	if selected.Code != http.StatusOK {
		t.Fatalf("set default status = %d body = %s", selected.Code, selected.Body.String())
	}
	assertJSONField(t, selected, "isDefault", true)

	disabled := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends/local-main/disable", "")
	if disabled.Code != http.StatusOK {
		t.Fatalf("disable status = %d body = %s", disabled.Code, disabled.Body.String())
	}
	assertJSONField(t, disabled, "healthStatus", "disabled")

	listed := performRequest(t, handler, http.MethodGet, "/api/v1/admin/storage/backends", "")
	if listed.Code != http.StatusOK {
		t.Fatalf("list status = %d body = %s", listed.Code, listed.Body.String())
	}
	var body struct {
		Backends []storage.StorageBackend `json:"backends"`
	}
	decodeResponse(t, listed, &body)
	if len(body.Backends) != 2 {
		t.Fatalf("backend count = %d, want 2", len(body.Backends))
	}
}

func TestStorageBackendProbeWorkflow(t *testing.T) {
	handler := newTestHandler()
	root := t.TempDir()
	local := `{"id":"local-probe","type":"local","displayName":"Local Probe","enabled":true,"config":{"local":{"rootPath":"` + root + `"}}}`

	registered := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", local)
	if registered.Code != http.StatusCreated {
		t.Fatalf("register status = %d body = %s", registered.Code, registered.Body.String())
	}
	probed := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends/local-probe/probe", "")
	if probed.Code != http.StatusOK {
		t.Fatalf("probe status = %d body = %s", probed.Code, probed.Body.String())
	}
	assertJSONField(t, probed, "status", "healthy")
	health := performRequest(t, handler, http.MethodGet, "/api/v1/admin/storage/backends/local-probe/health", "")
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d body = %s", health.Code, health.Body.String())
	}
	assertJSONField(t, health, "status", "healthy")
}

func TestStorageBackendProbeMissingS3CredentialFailure(t *testing.T) {
	handler := newTestHandler()
	s3 := `{"id":"s3-probe","type":"s3","displayName":"S3","enabled":true,"config":{"s3":{"endpoint":"https://s3.example.com","bucket":"inori","accessKeySecretRef":"A","secretKeySecretRef":"S"}}}`
	registered := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", s3)
	if registered.Code != http.StatusCreated {
		t.Fatalf("register status = %d body = %s", registered.Code, registered.Body.String())
	}
	probed := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends/s3-probe/probe", "")
	assertAPIError(t, probed, http.StatusUnprocessableEntity, "probe_failed")
	health := performRequest(t, handler, http.MethodGet, "/api/v1/admin/storage/backends/s3-probe/health", "")
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d body = %s", health.Code, health.Body.String())
	}
	assertJSONField(t, health, "status", "unhealthy")
}

func TestStorageBackendErrors(t *testing.T) {
	handler := newTestHandler()
	valid := `{"id":"local-main","type":"local","displayName":"Local Media","enabled":true,"config":{"local":{"rootPath":"/srv/inori/media"}}}`

	tests := []struct {
		name            string
		method          string
		path            string
		body            string
		omitContentType bool
		status          int
		code            string
	}{
		{name: "missing content type", method: http.MethodPost, path: "/api/v1/admin/storage/backends", body: valid, omitContentType: true, status: http.StatusBadRequest, code: "invalid_backend"},
		{name: "malformed json", method: http.MethodPost, path: "/api/v1/admin/storage/backends", body: `{`, status: http.StatusBadRequest, code: "invalid_backend"},
		{name: "unknown field", method: http.MethodPost, path: "/api/v1/admin/storage/backends", body: `{"id":"local-main","unknown":true}`, status: http.StatusBadRequest, code: "invalid_backend"},
		{name: "server owned field", method: http.MethodPost, path: "/api/v1/admin/storage/backends", body: `{"id":"local-main","healthStatus":"healthy"}`, status: http.StatusBadRequest, code: "invalid_backend"},
		{name: "oversized body", method: http.MethodPost, path: "/api/v1/admin/storage/backends", body: strings.Repeat(" ", maxRequestBodyBytes+1), status: http.StatusBadRequest, code: "invalid_backend"},
		{name: "not found", method: http.MethodPost, path: "/api/v1/admin/storage/backends/missing/default", status: http.StatusNotFound, code: "not_found"},
		{name: "unknown public route", method: http.MethodGet, path: "/missing", status: http.StatusNotFound, code: "not_found"},
		{name: "unknown admin route", method: http.MethodGet, path: "/api/v1/admin/missing", status: http.StatusNotFound, code: "not_found"},
		{name: "unsupported method", method: http.MethodPut, path: "/api/v1/admin/storage/backends", status: http.StatusMethodNotAllowed, code: "method_not_allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performRequestWithContentType(t, handler, tt.method, tt.path, tt.body, !tt.omitContentType)
			assertAPIError(t, response, tt.status, tt.code)
		})
	}

	first := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", valid)
	if first.Code != http.StatusCreated {
		t.Fatalf("initial registration status = %d body = %s", first.Code, first.Body.String())
	}
	duplicate := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", valid)
	assertAPIError(t, duplicate, http.StatusConflict, "conflict")
}

func newTestHandler() http.Handler {
	return NewHandler(storage.NewService(storage.NewMemoryRepository()), WithAdminToken(testAdminToken)).Routes()
}

func newUnauthenticatedTestHandler() http.Handler {
	return NewHandler(storage.NewService(storage.NewMemoryRepository())).Routes()
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

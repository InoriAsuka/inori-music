package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"inori-music/services/api/internal/storage"
)

func TestHealth(t *testing.T) {
	response := performRequest(t, newTestHandler(), http.MethodGet, "/healthz", "")
	if response.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d", response.Code, http.StatusOK)
	}
	assertJSONField(t, response, "status", "ok")
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
		{name: "unknown route", method: http.MethodGet, path: "/missing", status: http.StatusNotFound, code: "not_found"},
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

const testAdminToken = "0123456789abcdef0123456789abcdef"

func TestAdminAuthentication(t *testing.T) {
	secureHandler := newTestHandler()
	noAuth := performRequestWithToken(t, secureHandler, http.MethodGet, "/api/v1/admin/storage/backends", "", "")
	assertAPIError(t, noAuth, http.StatusUnauthorized, "unauthorized")
	if noAuth.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("missing WWW-Authenticate header")
	}

	badAuth := performRequestWithToken(t, secureHandler, http.MethodGet, "/api/v1/admin/storage/backends", "", "wrong-token")
	assertAPIError(t, badAuth, http.StatusUnauthorized, "unauthorized")

	health := performRequestWithToken(t, secureHandler, http.MethodGet, "/healthz", "", "")
	if health.Code != http.StatusOK {
		t.Fatalf("health without auth status = %d, want %d", health.Code, http.StatusOK)
	}

	unconfiguredHandler := NewHandler(storage.NewService(storage.NewMemoryRepository())).Routes()
	unconfigured := performRequestWithToken(t, unconfiguredHandler, http.MethodGet, "/api/v1/admin/storage/backends", "", testAdminToken)
	assertAPIError(t, unconfigured, http.StatusServiceUnavailable, "auth_not_configured")

	insecureHandler := NewHandler(storage.NewService(storage.NewMemoryRepository()), WithInsecureAdminAuth()).Routes()
	insecure := performRequestWithToken(t, insecureHandler, http.MethodGet, "/api/v1/admin/storage/backends", "", "")
	if insecure.Code != http.StatusOK {
		t.Fatalf("insecure dev status = %d, want %d, body = %s", insecure.Code, http.StatusOK, insecure.Body.String())
	}
}

func newTestHandler() http.Handler {
	return NewHandler(storage.NewService(storage.NewMemoryRepository()), WithAdminToken(testAdminToken)).Routes()
}

func performRequest(t *testing.T, handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	return performRequestWithOptions(t, handler, method, path, body, true, testAdminToken)
}

func performRequestWithContentType(t *testing.T, handler http.Handler, method string, path string, body string, includeContentType bool) *httptest.ResponseRecorder {
	t.Helper()
	return performRequestWithOptions(t, handler, method, path, body, includeContentType, testAdminToken)
}

func performRequestWithToken(t *testing.T, handler http.Handler, method string, path string, body string, token string) *httptest.ResponseRecorder {
	t.Helper()
	return performRequestWithOptions(t, handler, method, path, body, true, token)
}

func performRequestWithOptions(t *testing.T, handler http.Handler, method string, path string, body string, includeContentType bool, token string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
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

package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
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

func TestStorageBackendRefreshAndCapacityWorkflow(t *testing.T) {
	handler := newTestHandler()
	root := t.TempDir()
	local := `{"id":"local-capacity","type":"local","displayName":"Local Capacity","enabled":true,"config":{"local":{"rootPath":"` + root + `"}}}`
	registered := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", local)
	if registered.Code != http.StatusCreated {
		t.Fatalf("register status = %d body = %s", registered.Code, registered.Body.String())
	}
	capacity := performRequest(t, handler, http.MethodGet, "/api/v1/admin/storage/backends/local-capacity/capacity", "")
	if capacity.Code != http.StatusOK {
		t.Fatalf("capacity status = %d body = %s", capacity.Code, capacity.Body.String())
	}
	var capacityBody storage.CapacityReport
	decodeResponse(t, capacity, &capacityBody)
	if capacityBody.TotalBytes == 0 {
		t.Fatal("capacity totalBytes should be positive")
	}
	refresh := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends/refresh", "")
	if refresh.Code != http.StatusOK {
		t.Fatalf("refresh status = %d body = %s", refresh.Code, refresh.Body.String())
	}
	var refreshBody storage.RefreshReport
	decodeResponse(t, refresh, &refreshBody)
	if len(refreshBody.Results) != 1 || refreshBody.Results[0].Probe == nil || refreshBody.Results[0].Capacity == nil {
		t.Fatalf("refresh results = %+v, want probe and capacity", refreshBody.Results)
	}
}

func TestStorageBackendCapacityUnsupported(t *testing.T) {
	handler := newTestHandler()
	s3 := `{"id":"s3-capacity","type":"s3","displayName":"S3","enabled":true,"config":{"s3":{"endpoint":"https://s3.example.com","bucket":"inori","accessKeySecretRef":"A","secretKeySecretRef":"S"}}}`
	registered := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", s3)
	if registered.Code != http.StatusCreated {
		t.Fatalf("register status = %d body = %s", registered.Code, registered.Body.String())
	}
	capacity := performRequest(t, handler, http.MethodGet, "/api/v1/admin/storage/backends/s3-capacity/capacity", "")
	assertAPIError(t, capacity, http.StatusUnprocessableEntity, "capacity_unsupported")
}

func TestMediaObjectRoutesRegisterLookupAndFilter(t *testing.T) {
	handler := newTestHandler()
	backend := `{"id":"local-main","type":"local","displayName":"Local Media","enabled":true,"config":{"local":{"rootPath":"/srv/inori/media"}}}`
	backendResponse := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", backend)
	if backendResponse.Code != http.StatusCreated {
		t.Fatalf("backend register status = %d body = %s", backendResponse.Code, backendResponse.Body.String())
	}

	object := `{"id":"media-1","backendId":"local-main","objectKey":"albums/inori/track-01.flac","contentHash":"sha256:abcdef","sizeBytes":1234,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`
	registered := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects", object)
	if registered.Code != http.StatusCreated {
		t.Fatalf("media register status = %d body = %s", registered.Code, registered.Body.String())
	}
	assertJSONField(t, registered, "id", "media-1")
	for _, extra := range []string{
		`{"id":"media-2","backendId":"local-main","objectKey":"albums/inori/track-02.flac","contentHash":"sha256:2222","sizeBytes":1234,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`,
		`{"id":"media-3","backendId":"local-main","objectKey":"albums/inori/track-03.flac","contentHash":"sha256:3333","sizeBytes":1234,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`,
	} {
		if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects", extra); response.Code != http.StatusCreated {
			t.Fatalf("extra media register status = %d body = %s", response.Code, response.Body.String())
		}
	}

	lookup := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects/media-1", "")
	if lookup.Code != http.StatusOK {
		t.Fatalf("media lookup status = %d body = %s", lookup.Code, lookup.Body.String())
	}
	assertJSONField(t, lookup, "objectKey", "albums/inori/track-01.flac")

	byBackend := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main", "")
	assertMediaObjectListLength(t, byBackend, 3)
	paged := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main&limit=1&offset=1", "")
	assertMediaObjectPage(t, paged, 1, 1, 3, true)
	byHash := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?contentHash=sha256:abcdef", "")
	assertMediaObjectListLength(t, byHash, 1)
	byVerificationStatus := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?verificationStatus=unknown", "")
	assertMediaObjectListLength(t, byVerificationStatus, 3)
	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main&verificationStatus=unknown", ""), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?verificationStatus=stale", ""), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main&limit=0", ""), http.StatusBadRequest, "invalid_media_object")
}

func TestMediaObjectRouteRejectsInvalidInput(t *testing.T) {
	handler := newTestHandler()
	backend := `{"id":"local-main","type":"local","displayName":"Local Media","enabled":true,"config":{"local":{"rootPath":"/srv/inori/media"}}}`
	backendResponse := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", backend)
	if backendResponse.Code != http.StatusCreated {
		t.Fatalf("backend register status = %d body = %s", backendResponse.Code, backendResponse.Body.String())
	}

	unsafeKey := `{"id":"media-unsafe","backendId":"local-main","objectKey":"../escape.flac","contentHash":"sha256:abcdef","sizeBytes":1234,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`
	assertAPIError(t, performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects", unsafeKey), http.StatusBadRequest, "invalid_media_object")

	serverOwned := `{"id":"media-owned","backendId":"local-main","objectKey":"safe.flac","contentHash":"sha256:abcdef","sizeBytes":1234,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active","createdAt":"2026-06-03T00:00:00Z"}`
	assertAPIError(t, performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects", serverOwned), http.StatusBadRequest, "invalid_media_object")

	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects", ""), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main&contentHash=sha256:abcdef", ""), http.StatusBadRequest, "invalid_media_object")
}

func TestMediaObjectRouteRejectsDisabledBackendAndRequiresAuth(t *testing.T) {
	handler := newTestHandler()
	backend := `{"id":"local-disabled","type":"local","displayName":"Disabled","enabled":false,"config":{"local":{"rootPath":"/srv/inori/media"}}}`
	backendResponse := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", backend)
	if backendResponse.Code != http.StatusCreated {
		t.Fatalf("backend register status = %d body = %s", backendResponse.Code, backendResponse.Body.String())
	}

	object := `{"id":"media-disabled","backendId":"local-disabled","objectKey":"safe.flac","contentHash":"sha256:abcdef","sizeBytes":1234,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`
	assertAPIError(t, performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects", object), http.StatusConflict, "conflict")
	assertAPIError(t, performRequestWithoutAuth(t, handler, http.MethodPost, "/api/v1/admin/media/objects", object), http.StatusUnauthorized, "unauthorized")
}

func TestMediaObjectVerificationRoute(t *testing.T) {
	handler := newTestHandler()
	root := t.TempDir()
	content := []byte("verify me")
	objectPath := filepath.Join(root, "albums", "track.flac")
	if err := os.MkdirAll(filepath.Dir(objectPath), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(objectPath, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	backend := `{"id":"local-main","type":"local","displayName":"Local Media","enabled":true,"config":{"local":{"rootPath":"` + root + `"}}}`
	backendResponse := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", backend)
	if backendResponse.Code != http.StatusCreated {
		t.Fatalf("backend register status = %d body = %s", backendResponse.Code, backendResponse.Body.String())
	}
	sum := sha256.Sum256(content)
	object := `{"id":"media-verify","backendId":"local-main","objectKey":"albums/track.flac","contentHash":"sha256:` + hex.EncodeToString(sum[:]) + `","sizeBytes":` + strconv.Itoa(len(content)) + `,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`
	registered := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects", object)
	if registered.Code != http.StatusCreated {
		t.Fatalf("media register status = %d body = %s", registered.Code, registered.Body.String())
	}

	verified := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/media-verify/verify", "")
	if verified.Code != http.StatusOK {
		t.Fatalf("verify status = %d body = %s", verified.Code, verified.Body.String())
	}
	assertJSONField(t, verified, "status", "verified")
}

func TestMediaObjectVerificationRouteErrors(t *testing.T) {
	handler := newTestHandler()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bad.flac"), []byte("actual"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	backend := `{"id":"local-main","type":"local","displayName":"Local Media","enabled":true,"config":{"local":{"rootPath":"` + root + `"}}}`
	if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", backend); response.Code != http.StatusCreated {
		t.Fatalf("backend register status = %d body = %s", response.Code, response.Body.String())
	}
	object := `{"id":"media-bad","backendId":"local-main","objectKey":"bad.flac","contentHash":"sha256:0000","sizeBytes":6,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`
	if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects", object); response.Code != http.StatusCreated {
		t.Fatalf("media register status = %d body = %s", response.Code, response.Body.String())
	}
	assertAPIError(t, performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/media-bad/verify", ""), http.StatusUnprocessableEntity, "media_object_verification_failed")
	assertAPIError(t, performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/missing/verify", ""), http.StatusNotFound, "not_found")
}

func TestMediaObjectBatchVerificationRoute(t *testing.T) {
	handler := newTestHandler()
	root := t.TempDir()
	goodContent := []byte("good")
	if err := os.WriteFile(filepath.Join(root, "good.flac"), goodContent, 0o600); err != nil {
		t.Fatalf("WriteFile(good) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "bad.flac"), []byte("actual"), 0o600); err != nil {
		t.Fatalf("WriteFile(bad) error = %v", err)
	}
	backend := `{"id":"local-main","type":"local","displayName":"Local Media","enabled":true,"config":{"local":{"rootPath":"` + root + `"}}}`
	if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", backend); response.Code != http.StatusCreated {
		t.Fatalf("backend register status = %d body = %s", response.Code, response.Body.String())
	}
	goodSum := sha256.Sum256(goodContent)
	objects := []string{
		`{"id":"good","backendId":"local-main","objectKey":"good.flac","contentHash":"sha256:` + hex.EncodeToString(goodSum[:]) + `","sizeBytes":` + strconv.Itoa(len(goodContent)) + `,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`,
		`{"id":"bad","backendId":"local-main","objectKey":"bad.flac","contentHash":"sha256:0000","sizeBytes":6,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`,
	}
	for _, object := range objects {
		if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects", object); response.Code != http.StatusCreated {
			t.Fatalf("media register status = %d body = %s", response.Code, response.Body.String())
		}
	}

	response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/verify?backendId=local-main", "")
	if response.Code != http.StatusOK {
		t.Fatalf("batch verify status = %d body = %s", response.Code, response.Body.String())
	}
	var report storage.MediaObjectVerificationReport
	decodeResponse(t, response, &report)
	if len(report.Results) != 2 || report.CheckedAt.IsZero() {
		t.Fatalf("batch report = %+v, want two results", report)
	}
	statuses := map[string]string{}
	for _, result := range report.Results {
		statuses[result.MediaObjectID] = result.Status
	}
	if statuses["good"] != "verified" || statuses["bad"] != "failed" {
		t.Fatalf("statuses = %#v, want mixed verification outcomes", statuses)
	}
	failedObjects := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?verificationStatus=failed", "")
	assertMediaObjectListLength(t, failedObjects, 1)
}

func TestMediaObjectBatchVerificationRouteRejectsInvalidFiltersAndRequiresAuth(t *testing.T) {
	handler := newTestHandler()
	assertAPIError(t, performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/verify", ""), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/verify?backendId=a&contentHash=b", ""), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequestWithoutAuth(t, handler, http.MethodPost, "/api/v1/admin/media/objects/verify?backendId=a", ""), http.StatusUnauthorized, "unauthorized")
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
	repository := storage.NewMemoryRepository()
	return NewHandler(
		storage.NewService(repository),
		WithAdminToken(testAdminToken),
		WithMediaObjectService(storage.NewMediaObjectService(repository, storage.NewMemoryMediaObjectRepository())),
	).Routes()
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

func assertMediaObjectListLength(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	var body struct {
		Objects []storage.MediaObject `json:"objects"`
	}
	decodeResponse(t, response, &body)
	if len(body.Objects) != want {
		t.Fatalf("objects length = %d, want %d, body = %s", len(body.Objects), want, response.Body.String())
	}
}

func assertMediaObjectPage(t *testing.T, response *httptest.ResponseRecorder, wantObjects int, wantOffset int, wantTotal int, wantHasMore bool) {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	var body struct {
		Objects    []storage.MediaObject `json:"objects"`
		Pagination struct {
			Limit   int  `json:"limit"`
			Offset  int  `json:"offset"`
			Total   int  `json:"total"`
			HasMore bool `json:"hasMore"`
		} `json:"pagination"`
	}
	decodeResponse(t, response, &body)
	if len(body.Objects) != wantObjects || body.Pagination.Offset != wantOffset || body.Pagination.Total != wantTotal || body.Pagination.HasMore != wantHasMore {
		t.Fatalf("page = %+v, want objects=%d offset=%d total=%d hasMore=%v", body, wantObjects, wantOffset, wantTotal, wantHasMore)
	}
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v, body = %s", err, response.Body.String())
	}
}

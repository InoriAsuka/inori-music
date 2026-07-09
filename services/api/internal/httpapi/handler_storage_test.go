package httpapi

import (
	"net/http"
	"strings"
	"testing"

	"inori-music/services/api/internal/storage"
)

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

// ---- storage backend + album filter HTTP tests (Phase 138) ----

func TestGetStorageBackend(t *testing.T) {
	// Register a backend, then fetch it by ID
	backend := `{"id":"local-1","type":"local","displayName":"Local One","enabled":true,"config":{"local":{"rootPath":"/tmp/t1"}}}`
	registerResp := performRequest(t, newTestHandler(), http.MethodPost, "/api/v1/admin/storage/backends", backend)
	if registerResp.Code != http.StatusCreated {
		t.Fatalf("register status = %d; body = %s", registerResp.Code, registerResp.Body.String())
	}

	h := newTestHandler()
	performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backend)

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/storage/backends/local-1", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("GET backend status = %d; body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	if got["id"] != "local-1" || got["displayName"] != "Local One" {
		t.Errorf("unexpected backend = %v", got)
	}

	// Unknown ID → 404
	resp2 := performRequest(t, h, http.MethodGet, "/api/v1/admin/storage/backends/no-such", "")
	assertAPIError(t, resp2, http.StatusNotFound, "not_found")
}

func TestPatchStorageBackend(t *testing.T) {
	h := newTestHandler()
	backend := `{"id":"local-patch","type":"local","displayName":"Before","enabled":true,"config":{"local":{"rootPath":"/tmp/patch"}}}`
	performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backend)

	// Patch displayName
	resp := performRequest(t, h, http.MethodPatch, "/api/v1/admin/storage/backends/local-patch", `{"displayName":"After"}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("PATCH backend status = %d; body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	if got["displayName"] != "After" {
		t.Errorf("displayName = %v, want After", got["displayName"])
	}

	// Empty displayName → 400
	resp2 := performRequest(t, h, http.MethodPatch, "/api/v1/admin/storage/backends/local-patch", `{"displayName":""}`)
	assertAPIError(t, resp2, http.StatusBadRequest, "invalid_backend")

	// Unknown ID → 404
	resp3 := performRequest(t, h, http.MethodPatch, "/api/v1/admin/storage/backends/no-such", `{"priority":5}`)
	assertAPIError(t, resp3, http.StatusNotFound, "not_found")
}

func TestEnableStorageBackend(t *testing.T) {
	h := newTestHandler()
	// Register and then disable a backend, then enable it
	backend := `{"id":"local-enable","type":"local","displayName":"Enable Test","enabled":true,"config":{"local":{"rootPath":"/tmp/en"}}}`
	performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backend)
	performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends/local-enable/disable", "")

	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends/local-enable/enable", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("enable status = %d; body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	if got["enabled"] != true {
		t.Errorf("enabled = %v, want true", got["enabled"])
	}

	// Idempotent enable of already-enabled backend
	resp2 := performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends/local-enable/enable", "")
	if resp2.Code != http.StatusOK {
		t.Fatalf("idempotent enable status = %d", resp2.Code)
	}
}

func TestDeleteStorageBackendGuards(t *testing.T) {
	h := newTestHandler()
	backend := `{"id":"local-del","type":"local","displayName":"Del Test","enabled":true,"isDefault":false,"config":{"local":{"rootPath":"/tmp/del"}}}`
	performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backend)

	// Set as default then try to delete → 409 storage_backend_is_default
	performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends/local-del/default", "")
	resp := performRequest(t, h, http.MethodDelete, "/api/v1/admin/storage/backends/local-del", "")
	assertAPIError(t, resp, http.StatusConflict, "storage_backend_is_default")

	// Unknown ID → 404
	resp2 := performRequest(t, h, http.MethodDelete, "/api/v1/admin/storage/backends/no-such", "")
	assertAPIError(t, resp2, http.StatusNotFound, "not_found")
}

func TestDeleteStorageBackendSuccess(t *testing.T) {
	h := newTestHandler()
	backend := `{"id":"local-del2","type":"local","displayName":"Del2","enabled":true,"config":{"local":{"rootPath":"/tmp/del2"}}}`
	performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backend)

	resp := performRequest(t, h, http.MethodDelete, "/api/v1/admin/storage/backends/local-del2", "")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("DELETE backend status = %d; body = %s", resp.Code, resp.Body.String())
	}

	// Verify gone
	resp2 := performRequest(t, h, http.MethodGet, "/api/v1/admin/storage/backends/local-del2", "")
	assertAPIError(t, resp2, http.StatusNotFound, "not_found")
}

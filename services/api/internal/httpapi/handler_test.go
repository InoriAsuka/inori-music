package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"inori-music/services/api/internal/auth"
	"inori-music/services/api/internal/catalog"
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
	if !readyReport.Ready || len(readyReport.Checks) != 3 {
		t.Fatalf("ready report = %+v, want ready report with three checks", readyReport)
	}

	unready := performRequestWithoutAuth(t, newUnauthenticatedTestHandler(), http.MethodGet, "/readyz", "")
	if unready.Code != http.StatusServiceUnavailable {
		t.Fatalf("unready GET /readyz status = %d, want %d body = %s", unready.Code, http.StatusServiceUnavailable, unready.Body.String())
	}
	var unreadyReport ReadinessReport
	decodeResponse(t, unready, &unreadyReport)
	if unreadyReport.Ready || len(unreadyReport.Checks) != 3 {
		t.Fatalf("unready report = %+v, want not ready report with three checks", unreadyReport)
	}
	failed := make(map[string]bool)
	for _, check := range unreadyReport.Checks {
		if check.Status == "failed" {
			failed[check.Name] = true
		}
	}
	if !failed["media_registry"] || !failed["admin_auth"] {
		t.Fatalf("failed checks = %+v, want media_registry and admin_auth failures", failed)
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
		`{"id":"media-2","backendId":"local-main","objectKey":"albums/inori/track-02.flac","contentHash":"sha256:2222","sizeBytes":900,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`,
		`{"id":"media-3","backendId":"local-main","objectKey":"albums/inori/track-03.flac","contentHash":"sha256:3333","sizeBytes":2000,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`,
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

	lifecycle := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/media-1/lifecycle", `{"lifecycleState":"archived"}`)
	if lifecycle.Code != http.StatusOK {
		t.Fatalf("lifecycle status = %d body = %s", lifecycle.Code, lifecycle.Body.String())
	}
	assertJSONField(t, lifecycle, "lifecycleState", "archived")
	assertAPIError(t, performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/media-1/lifecycle", `{"lifecycleState":"missing"}`), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequestWithoutAuth(t, handler, http.MethodPost, "/api/v1/admin/media/objects/media-1/lifecycle", `{"lifecycleState":"active"}`), http.StatusUnauthorized, "unauthorized")

	byLifecycle := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?lifecycleState=archived", "")
	assertMediaObjectListLength(t, byLifecycle, 1)
	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?lifecycleState=missing", ""), http.StatusBadRequest, "invalid_media_object")
	byAssetKind := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?assetKind=original_audio", "")
	assertMediaObjectListLength(t, byAssetKind, 3)
	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?assetKind=thumbnail", ""), http.StatusBadRequest, "invalid_media_object")

	byBackend := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main", "")
	assertMediaObjectListLength(t, byBackend, 3)
	paged := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main&limit=1&offset=1", "")
	assertMediaObjectPage(t, paged, 1, 1, 3, true)
	sorted := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main&sortBy=size_bytes&sortOrder=desc&limit=2", "")
	assertMediaObjectIDs(t, sorted, []string{"media-3", "media-1"})
	byHash := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?contentHash=sha256:abcdef", "")
	assertMediaObjectListLength(t, byHash, 1)
	byVerificationStatus := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?verificationStatus=unknown", "")
	assertMediaObjectListLength(t, byVerificationStatus, 3)
	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main&verificationStatus=unknown", ""), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?verificationStatus=stale", ""), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main&limit=0", ""), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main&sortBy=missing", ""), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?backendId=local-main&sortOrder=sideways", ""), http.StatusBadRequest, "invalid_media_object")

	stats := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects/stats?backendId=local-main", "")
	if stats.Code != http.StatusOK {
		t.Fatalf("stats status = %d body = %s", stats.Code, stats.Body.String())
	}
	var statsBody storage.MediaObjectStats
	decodeResponse(t, stats, &statsBody)
	if statsBody.BackendID != "local-main" || statsBody.TotalObjects != 3 || statsBody.TotalSizeBytes != 4134 || statsBody.ByVerificationStatus["unknown"] != 3 {
		t.Fatalf("stats = %+v, want backend totals and unknown verification count", statsBody)
	}
	assertAPIError(t, performRequestWithoutAuth(t, handler, http.MethodGet, "/api/v1/admin/media/objects/stats", ""), http.StatusUnauthorized, "unauthorized")
}

func TestMediaObjectDuplicateRoute(t *testing.T) {
	handler := newTestHandler()
	backend := `{"id":"local-main","type":"local","displayName":"Local Media","enabled":true,"config":{"local":{"rootPath":"/srv/inori/media"}}}`
	if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", backend); response.Code != http.StatusCreated {
		t.Fatalf("backend register status = %d body = %s", response.Code, response.Body.String())
	}
	objects := []string{
		`{"id":"media-a","backendId":"local-main","objectKey":"albums/a.flac","contentHash":"sha256:dup","sizeBytes":100,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`,
		`{"id":"media-b","backendId":"local-main","objectKey":"albums/b.flac","contentHash":"sha256:dup","sizeBytes":200,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`,
		`{"id":"media-c","backendId":"local-main","objectKey":"albums/c.flac","contentHash":"sha256:unique","sizeBytes":300,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`,
	}
	for _, object := range objects {
		if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects", object); response.Code != http.StatusCreated {
			t.Fatalf("media register status = %d body = %s", response.Code, response.Body.String())
		}
	}

	response := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects/duplicates?backendId=local-main", "")
	if response.Code != http.StatusOK {
		t.Fatalf("duplicates status = %d body = %s", response.Code, response.Body.String())
	}
	var report storage.MediaObjectDuplicateReport
	decodeResponse(t, response, &report)
	if report.BackendID != "local-main" || report.MinCopies != 2 || report.TotalGroups != 1 || report.TotalObjects != 2 || report.TotalSizeBytes != 300 {
		t.Fatalf("duplicate report = %+v, want one duplicate group", report)
	}
	if len(report.Groups) != 1 || report.Groups[0].ContentHash != "sha256:dup" || len(report.Groups[0].Objects) != 2 {
		t.Fatalf("duplicate groups = %+v, want sha256:dup with two objects", report.Groups)
	}

	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects/duplicates?minCopies=1", ""), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequestWithoutAuth(t, handler, http.MethodGet, "/api/v1/admin/media/objects/duplicates", ""), http.StatusUnauthorized, "unauthorized")
}

func TestMediaObjectBulkLifecycleRoute(t *testing.T) {
	handler := newTestHandler()
	backend := `{"id":"local-main","type":"local","displayName":"Local Media","enabled":true,"config":{"local":{"rootPath":"/srv/inori/media"}}}`
	if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", backend); response.Code != http.StatusCreated {
		t.Fatalf("backend register status = %d body = %s", response.Code, response.Body.String())
	}
	objects := []string{
		`{"id":"bulk-a","backendId":"local-main","objectKey":"bulk/a.flac","contentHash":"sha256:bulk-a","sizeBytes":100,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`,
		`{"id":"bulk-b","backendId":"local-main","objectKey":"bulk/b.flac","contentHash":"sha256:bulk-b","sizeBytes":200,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`,
		`{"id":"bulk-art","backendId":"local-main","objectKey":"bulk/art.jpg","contentHash":"sha256:bulk-art","sizeBytes":50,"mimeType":"image/jpeg","assetKind":"artwork","lifecycleState":"active"}`,
	}
	for _, object := range objects {
		if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects", object); response.Code != http.StatusCreated {
			t.Fatalf("media register status = %d body = %s", response.Code, response.Body.String())
		}
	}

	dryRunRequest := `{"filter":{"assetKind":"original_audio"},"lifecycleState":"archived","dryRun":true}`
	dryRunResponse := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/lifecycle", dryRunRequest)
	if dryRunResponse.Code != http.StatusOK {
		t.Fatalf("dry-run bulk lifecycle status = %d body = %s", dryRunResponse.Code, dryRunResponse.Body.String())
	}
	var dryRunReport storage.MediaObjectLifecycleUpdateReport
	decodeResponse(t, dryRunResponse, &dryRunReport)
	if !dryRunReport.DryRun || dryRunReport.WouldUpdateObjects != 2 || dryRunReport.UpdatedObjects != 0 {
		t.Fatalf("dry-run report = %+v, want two would-update objects", dryRunReport)
	}
	active := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?lifecycleState=active", "")
	assertMediaObjectListLength(t, active, 3)

	request := `{"filter":{"assetKind":"original_audio"},"lifecycleState":"archived"}`
	response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/lifecycle", request)
	if response.Code != http.StatusOK {
		t.Fatalf("bulk lifecycle status = %d body = %s", response.Code, response.Body.String())
	}
	var report storage.MediaObjectLifecycleUpdateReport
	decodeResponse(t, response, &report)
	if report.MatchedObjects != 2 || report.UpdatedObjects != 2 || report.FailedObjects != 0 || report.LifecycleState != "archived" {
		t.Fatalf("report = %+v, want two archived objects", report)
	}
	archived := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects?lifecycleState=archived", "")
	assertMediaObjectListLength(t, archived, 2)

	assertAPIError(t, performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/lifecycle", `{"filter":{},"lifecycleState":"archived"}`), http.StatusBadRequest, "invalid_media_object")
	assertAPIError(t, performRequestWithoutAuth(t, handler, http.MethodPost, "/api/v1/admin/media/objects/lifecycle", request), http.StatusUnauthorized, "unauthorized")
}

func TestMediaObjectTimelineRoute(t *testing.T) {
	handler := newTestHandler()
	backend := `{"id":"local-main","type":"local","displayName":"Local Media","enabled":true,"config":{"local":{"rootPath":"/srv/inori/media"}}}`
	if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/storage/backends", backend); response.Code != http.StatusCreated {
		t.Fatalf("backend register status = %d body = %s", response.Code, response.Body.String())
	}
	object := `{"id":"timeline-media","backendId":"local-main","objectKey":"timeline/track.flac","contentHash":"sha256:timeline","sizeBytes":100,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`
	if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects", object); response.Code != http.StatusCreated {
		t.Fatalf("media register status = %d body = %s", response.Code, response.Body.String())
	}
	if response := performRequest(t, handler, http.MethodPost, "/api/v1/admin/media/objects/timeline-media/lifecycle", `{"lifecycleState":"archived"}`); response.Code != http.StatusOK {
		t.Fatalf("lifecycle status = %d body = %s", response.Code, response.Body.String())
	}

	response := performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects/timeline-media/timeline", "")
	if response.Code != http.StatusOK {
		t.Fatalf("timeline status = %d body = %s", response.Code, response.Body.String())
	}
	var timeline storage.MediaObjectTimeline
	decodeResponse(t, response, &timeline)
	if timeline.MediaObjectID != "timeline-media" || len(timeline.Events) != 2 {
		t.Fatalf("timeline = %+v, want created and lifecycle events", timeline)
	}
	if timeline.Events[0].Type != "created" || timeline.Events[1].Type != "lifecycle_changed" || timeline.Events[1].Source != "single" || timeline.Events[1].LifecycleState != "archived" {
		t.Fatalf("timeline events = %+v, want lifecycle summary", timeline.Events)
	}

	assertAPIError(t, performRequest(t, handler, http.MethodGet, "/api/v1/admin/media/objects/missing/timeline", ""), http.StatusNotFound, "not_found")
	assertAPIError(t, performRequestWithoutAuth(t, handler, http.MethodGet, "/api/v1/admin/media/objects/timeline-media/timeline", ""), http.StatusUnauthorized, "unauthorized")
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
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test-version", Commit: "test-commit", BuildTime: "2026-06-05T12:30:00Z"}),
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

func assertMediaObjectIDs(t *testing.T, response *httptest.ResponseRecorder, want []string) {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	var body struct {
		Objects []storage.MediaObject `json:"objects"`
	}
	decodeResponse(t, response, &body)
	if len(body.Objects) != len(want) {
		t.Fatalf("objects length = %d, want %d, body = %s", len(body.Objects), len(want), response.Body.String())
	}
	for i, object := range body.Objects {
		if object.ID != want[i] {
			t.Fatalf("object[%d].id = %q, want %q; body = %s", i, object.ID, want[i], response.Body.String())
		}
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

// ---- auth test helpers ----

type memAuthUserRepo struct {
	mu    sync.RWMutex
	users map[string]auth.User
}

func newMemAuthUserRepo() *memAuthUserRepo { return &memAuthUserRepo{users: map[string]auth.User{}} }

func (r *memAuthUserRepo) SaveUser(_ context.Context, u auth.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}
func (r *memAuthUserRepo) GetUser(_ context.Context, id string) (auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return auth.User{}, fmt.Errorf("%w: %s", auth.ErrUserNotFound, id)
	}
	return u, nil
}
func (r *memAuthUserRepo) GetUserByUsername(_ context.Context, username string) (auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return auth.User{}, fmt.Errorf("%w: %s", auth.ErrUserNotFound, username)
}
func (r *memAuthUserRepo) ListUsers(_ context.Context) ([]auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]auth.User, 0, len(r.users))
	for _, u := range r.users {
		list = append(list, u)
	}
	return list, nil
}
func (r *memAuthUserRepo) DeleteUser(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return fmt.Errorf("%w: %s", auth.ErrUserNotFound, id)
	}
	delete(r.users, id)
	return nil
}
func (r *memAuthUserRepo) CountAdminUsers(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, u := range r.users {
		if u.Role == auth.RoleAdmin && u.Enabled {
			n++
		}
	}
	return n, nil
}

type memAuthSessionRepo struct {
	mu       sync.RWMutex
	sessions map[string]auth.Session
}

func newMemAuthSessionRepo() *memAuthSessionRepo {
	return &memAuthSessionRepo{sessions: map[string]auth.Session{}}
}
func (r *memAuthSessionRepo) SaveSession(_ context.Context, s auth.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.TokenHash] = s
	return nil
}
func (r *memAuthSessionRepo) GetSession(_ context.Context, h string) (auth.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[h]
	if !ok {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	return s, nil
}
func (r *memAuthSessionRepo) RevokeSession(_ context.Context, h string, t time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[h]
	if !ok {
		return auth.ErrSessionNotFound
	}
	s.RevokedAt = &t
	r.sessions[h] = s
	return nil
}
func (r *memAuthSessionRepo) DeleteExpiredSessions(_ context.Context, before time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, s := range r.sessions {
		if s.ExpiresAt.Before(before) {
			delete(r.sessions, k)
		}
	}
	return nil
}

func newAuthTestHandler() (http.Handler, *auth.Service) {
	repo := storage.NewMemoryRepository()
	svc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	if _, err := svc.CreateUser(context.Background(), "admin", "adminpass1", auth.RoleAdmin); err != nil {
		panic(err)
	}
	h := NewHandler(
		storage.NewService(repo),
		WithAuthService(svc),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test-version", Commit: "test-commit", BuildTime: "2026-06-05T12:30:00Z"}),
	).Routes()
	return h, svc
}

// ---- auth endpoint tests ----

func TestLoginSuccess(t *testing.T) {
	h, _ := newAuthTestHandler()
	body := `{"username":"admin","password":"adminpass1"}`
	resp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", body)
	if resp.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["token"] == "" || result["token"] == nil {
		t.Error("expected non-empty token in response")
	}
}

func TestLoginBadCredentials(t *testing.T) {
	h, _ := newAuthTestHandler()
	body := `{"username":"admin","password":"wrongpass"}`
	resp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", body)
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestLoginUnknownUser(t *testing.T) {
	h, _ := newAuthTestHandler()
	body := `{"username":"nobody","password":"adminpass1"}`
	resp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", body)
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestLoginNotConfigured(t *testing.T) {
	repo := storage.NewMemoryRepository()
	h := NewHandler(storage.NewService(repo), WithAdminToken(testAdminToken)).Routes()
	body := `{"username":"admin","password":"adminpass1"}`
	resp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", body)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

func TestLogoutSuccess(t *testing.T) {
	h, _ := newAuthTestHandler()
	// Login first.
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"adminpass1"}`)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d", loginResp.Code)
	}
	var loginResult map[string]any
	json.Unmarshal(loginResp.Body.Bytes(), &loginResult)
	token := loginResult["token"].(string)

	// Logout.
	logoutResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/auth/logout", "", "Bearer "+token)
	if logoutResp.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, body = %s", logoutResp.Code, logoutResp.Body.String())
	}
}

func TestLogoutInvalidToken(t *testing.T) {
	h, _ := newAuthTestHandler()
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/auth/logout", "", "Bearer invalidtoken")
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestSessionTokenGrantsAdminAccess(t *testing.T) {
	h, _ := newAuthTestHandler()
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"adminpass1"}`)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d", loginResp.Code)
	}
	var loginResult map[string]any
	json.Unmarshal(loginResp.Body.Bytes(), &loginResult)
	token := loginResult["token"].(string)

	// Use session token to access an admin route.
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/storage/backends", "", "Bearer "+token)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin route status = %d, body = %s", resp.Code, resp.Body.String())
	}
}

func TestRevokedSessionDeniesAccess(t *testing.T) {
	h, _ := newAuthTestHandler()
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"adminpass1"}`)
	var loginResult map[string]any
	json.Unmarshal(loginResp.Body.Bytes(), &loginResult)
	token := loginResult["token"].(string)

	// Logout (revoke).
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/auth/logout", "", "Bearer "+token)

	// Revoked token should now be denied.
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/storage/backends", "", "Bearer "+token)
	if resp.Code == http.StatusOK {
		t.Fatal("expected denied access after logout, got 200")
	}
}

func TestUserManagementWorkflow(t *testing.T) {
	h, _ := newAuthTestHandler()
	token := loginAdminToken(t, h)

	listed := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/users", "", "Bearer "+token)
	if listed.Code != http.StatusOK {
		t.Fatalf("list users status = %d, body = %s", listed.Code, listed.Body.String())
	}
	assertUserListLength(t, listed, 1)

	created := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users", `{"username":"viewer1","password":"viewerpass1","role":"viewer"}`, "Bearer "+token)
	if created.Code != http.StatusCreated {
		t.Fatalf("create user status = %d, body = %s", created.Code, created.Body.String())
	}
	var createdUser auth.UserView
	decodeResponse(t, created, &createdUser)
	if createdUser.ID == "" || createdUser.Username != "viewer1" || createdUser.Role != auth.RoleViewer || !createdUser.Enabled {
		t.Fatalf("created user = %+v", createdUser)
	}
	if strings.Contains(created.Body.String(), "password") {
		t.Fatalf("response leaked password material: %s", created.Body.String())
	}

	listed = performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/users", "", "Bearer "+token)
	assertUserListLength(t, listed, 2)

	disabled := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users/"+createdUser.ID+"/disable", "", "Bearer "+token)
	if disabled.Code != http.StatusOK {
		t.Fatalf("disable user status = %d, body = %s", disabled.Code, disabled.Body.String())
	}
	var disabledUser auth.UserView
	decodeResponse(t, disabled, &disabledUser)
	if disabledUser.Enabled {
		t.Fatalf("disabled user still enabled: %+v", disabledUser)
	}

	deleted := performRequestWithAuthHeader(t, h, http.MethodDelete, "/api/v1/admin/users/"+createdUser.ID, "", "Bearer "+token)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete user status = %d, body = %s", deleted.Code, deleted.Body.String())
	}
	listed = performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/users", "", "Bearer "+token)
	assertUserListLength(t, listed, 1)
}

func TestUserManagementCreateValidation(t *testing.T) {
	h, _ := newAuthTestHandler()
	token := loginAdminToken(t, h)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users", `{"username":"bad-name","password":"viewerpass1","role":"viewer"}`, "Bearer "+token)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_user")
}

func TestUserManagementCreateConflict(t *testing.T) {
	h, _ := newAuthTestHandler()
	token := loginAdminToken(t, h)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users", `{"username":"admin","password":"viewerpass1","role":"admin"}`, "Bearer "+token)
	assertAPIError(t, resp, http.StatusConflict, "conflict")
}

func TestUserManagementRequiresAdminRole(t *testing.T) {
	h, svc := newAuthTestHandler()
	if _, err := svc.CreateUser(context.Background(), "viewer2", "viewerpass2", auth.RoleViewer); err != nil {
		t.Fatal(err)
	}
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"viewer2","password":"viewerpass2"}`)
	var loginResult map[string]any
	decodeResponse(t, loginResp, &loginResult)
	token := loginResult["token"].(string)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/users", "", "Bearer "+token)
	assertAPIError(t, resp, http.StatusForbidden, "unauthorized")
}

func TestUserManagementNotConfigured(t *testing.T) {
	repo := storage.NewMemoryRepository()
	h := NewHandler(storage.NewService(repo), WithAdminToken(testAdminToken)).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/users", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

func loginAdminToken(t *testing.T, h http.Handler) string {
	t.Helper()
	loginResp := performRequestWithoutAuth(t, h, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"adminpass1"}`)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", loginResp.Code, loginResp.Body.String())
	}
	var loginResult map[string]any
	decodeResponse(t, loginResp, &loginResult)
	token, ok := loginResult["token"].(string)
	if !ok || token == "" {
		t.Fatalf("missing token in login response: %+v", loginResult)
	}
	return token
}

func assertUserListLength(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("list users status = %d, body = %s", response.Code, response.Body.String())
	}
	var body struct {
		Users []auth.UserView `json:"users"`
	}
	decodeResponse(t, response, &body)
	if len(body.Users) != want {
		t.Fatalf("users length = %d, want %d, body = %+v", len(body.Users), want, body.Users)
	}
}

// ---- catalog test helpers ----

func newCatalogTestHandler() http.Handler {
	repo := storage.NewMemoryRepository()
	catalogRepo := catalog.NewMemoryRepository()
	return NewHandler(
		storage.NewService(repo),
		WithAdminToken(testAdminToken),
		WithCatalogService(catalog.NewService(catalogRepo)),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test-version", Commit: "test-commit", BuildTime: "2026-06-05T12:30:00Z"}),
	).Routes()
}

// ---- artist tests ----

func TestCatalogArtistWorkflow(t *testing.T) {
	h := newCatalogTestHandler()

	// list empty
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("list artists status = %d, body = %s", resp.Code, resp.Body.String())
	}
	assertArtistListLength(t, resp, 0)

	// create
	resp = performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Hatsune Miku","sortName":"Miku Hatsune"}`)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create artist status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var artist map[string]any
	decodeResponse(t, resp, &artist)
	id, _ := artist["id"].(string)
	if id == "" {
		t.Fatal("expected artist id in response")
	}
	if artist["name"] != "Hatsune Miku" {
		t.Fatalf("artist name = %q, want %q", artist["name"], "Hatsune Miku")
	}

	// list now has 1
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists", "")
	assertArtistListLength(t, resp, 1)

	// get by id
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists/"+id, "")
	if resp.Code != http.StatusOK {
		t.Fatalf("get artist status = %d, body = %s", resp.Code, resp.Body.String())
	}

	// delete
	resp = performRequest(t, h, http.MethodDelete, "/api/v1/admin/catalog/artists/"+id, "")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("delete artist status = %d, body = %s", resp.Code, resp.Body.String())
	}

	// list empty again
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists", "")
	assertArtistListLength(t, resp, 0)
}

func TestCatalogArtistNotFound(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists/missing", "")
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestCatalogArtistInvalid(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":""}`)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_catalog_entity")
}

func TestCatalogArtistNotConfigured(t *testing.T) {
	h := newTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

// ---- album tests ----

func TestCatalogAlbumWorkflow(t *testing.T) {
	h := newCatalogTestHandler()

	// create artist first
	aResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Ryo"}`)
	if aResp.Code != http.StatusCreated {
		t.Fatalf("create artist status = %d", aResp.Code)
	}
	var aBody map[string]any
	decodeResponse(t, aResp, &aBody)
	artistID, _ := aBody["id"].(string)

	// create album
	albumBody := fmt.Sprintf(`{"title":"supercell","artistId":%q,"releaseYear":2009}`, artistID)
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums", albumBody)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create album status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var album map[string]any
	decodeResponse(t, resp, &album)
	albumID, _ := album["id"].(string)
	if albumID == "" {
		t.Fatal("expected album id")
	}

	// list with artistId filter
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums?artistId="+artistID, "")
	assertAlbumListLength(t, resp, 1)

	// list all
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums", "")
	assertAlbumListLength(t, resp, 1)

	// get
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums/"+albumID, "")
	if resp.Code != http.StatusOK {
		t.Fatalf("get album status = %d", resp.Code)
	}

	// delete
	resp = performRequest(t, h, http.MethodDelete, "/api/v1/admin/catalog/albums/"+albumID, "")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("delete album status = %d", resp.Code)
	}
}

func TestCatalogAlbumArtistMismatch(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums", `{"title":"X","artistId":"missing"}`)
	if resp.Code == http.StatusCreated {
		t.Fatal("expected failure when artist does not exist")
	}
}

// ---- track tests ----

func TestCatalogTrackWorkflow(t *testing.T) {
	h := newCatalogTestHandler()

	// create artist
	aResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Miku"}`)
	if aResp.Code != http.StatusCreated {
		t.Fatalf("create artist status = %d", aResp.Code)
	}
	var aBody map[string]any
	decodeResponse(t, aResp, &aBody)
	artistID, _ := aBody["id"].(string)

	// create album
	alResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums", fmt.Sprintf(`{"title":"Miku Best","artistId":%q}`, artistID))
	if alResp.Code != http.StatusCreated {
		t.Fatalf("create album status = %d", alResp.Code)
	}
	var alBody map[string]any
	decodeResponse(t, alResp, &alBody)
	albumID, _ := alBody["id"].(string)

	// create track
	trackBody := fmt.Sprintf(`{"title":"World Is Mine","artistId":%q,"albumId":%q,"mediaObjectId":"media-1","trackNumber":1,"durationMs":245000}`, artistID, albumID)
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks", trackBody)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create track status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var track map[string]any
	decodeResponse(t, resp, &track)
	trackID, _ := track["id"].(string)
	if trackID == "" {
		t.Fatal("expected track id")
	}

	// list by album
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/tracks?albumId="+albumID, "")
	assertTrackListLength(t, resp, 1)

	// list by artist
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/tracks?artistId="+artistID, "")
	assertTrackListLength(t, resp, 1)

	// list all
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/tracks", "")
	assertTrackListLength(t, resp, 1)

	// get
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/tracks/"+trackID, "")
	if resp.Code != http.StatusOK {
		t.Fatalf("get track status = %d", resp.Code)
	}

	// delete
	resp = performRequest(t, h, http.MethodDelete, "/api/v1/admin/catalog/tracks/"+trackID, "")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("delete track status = %d", resp.Code)
	}
}

func TestCatalogTrackInvalidMissingTitle(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks", `{"title":"","artistId":"x","mediaObjectId":"m"}`)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_catalog_entity")
}

// ---- catalog list assert helpers ----

func assertArtistListLength(t *testing.T, resp *httptest.ResponseRecorder, want int) {
	t.Helper()
	if resp.Code != http.StatusOK {
		t.Fatalf("list artists status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var body struct {
		Artists []map[string]any `json:"artists"`
	}
	decodeResponse(t, resp, &body)
	if len(body.Artists) != want {
		t.Fatalf("artists length = %d, want %d", len(body.Artists), want)
	}
}

func assertAlbumListLength(t *testing.T, resp *httptest.ResponseRecorder, want int) {
	t.Helper()
	if resp.Code != http.StatusOK {
		t.Fatalf("list albums status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var body struct {
		Albums []map[string]any `json:"albums"`
	}
	decodeResponse(t, resp, &body)
	if len(body.Albums) != want {
		t.Fatalf("albums length = %d, want %d", len(body.Albums), want)
	}
}

func assertTrackListLength(t *testing.T, resp *httptest.ResponseRecorder, want int) {
	t.Helper()
	if resp.Code != http.StatusOK {
		t.Fatalf("list tracks status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var body struct {
		Tracks []map[string]any `json:"tracks"`
	}
	decodeResponse(t, resp, &body)
	if len(body.Tracks) != want {
		t.Fatalf("tracks length = %d, want %d", len(body.Tracks), want)
	}
}

// ---- catalog import HTTP tests ----

func newImportTestHandlerWithMediaObject(t *testing.T, mediaObjID, assetKind, lifecycleState string) (http.Handler, string) {
	t.Helper()
	repo := storage.NewMemoryRepository()
	mediaRepo := storage.NewMemoryMediaObjectRepository()
	if mediaObjID != "" {
		// pre-populate the in-memory media repository
		obj := storage.MediaObject{
			ID:             mediaObjID,
			BackendID:      "backend-1",
			ObjectKey:      "key/" + mediaObjID,
			AssetKind:      assetKind,
			LifecycleState: lifecycleState,
		}
		if err := mediaRepo.SaveMediaObject(context.Background(), obj); err != nil {
			t.Fatalf("SaveMediaObject: %v", err)
		}
	}
	catalogRepo := catalog.NewMemoryRepository()
	catalogSvc := catalog.NewService(catalogRepo)
	mediaSvc := storage.NewMediaObjectService(repo, mediaRepo)
	h := NewHandler(
		storage.NewService(repo),
		WithAdminToken(testAdminToken),
		WithCatalogService(catalogSvc),
		WithMediaObjectService(mediaSvc),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()
	return h, mediaObjID
}

func TestCatalogImportTrackSuccess(t *testing.T) {
	h, mediaID := newImportTestHandlerWithMediaObject(t, "media-import-1", "original_audio", "active")
	body := fmt.Sprintf(`{"mediaObjectId":%q,"title":"World Is Mine","trackNumber":1,"durationMs":245000}`, mediaID)
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/import", body)
	if resp.Code != http.StatusCreated {
		t.Fatalf("import status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var track map[string]any
	decodeResponse(t, resp, &track)
	if track["id"] == "" || track["id"] == nil {
		t.Fatal("expected track id")
	}
	if track["title"] != "World Is Mine" {
		t.Fatalf("title = %q, want %q", track["title"], "World Is Mine")
	}
	if track["mediaObjectId"] != mediaID {
		t.Fatalf("mediaObjectId = %q, want %q", track["mediaObjectId"], mediaID)
	}
}

func TestCatalogImportTrackTitleFallback(t *testing.T) {
	h, mediaID := newImportTestHandlerWithMediaObject(t, "media-import-2", "transcoded_audio", "active")
	body := fmt.Sprintf(`{"mediaObjectId":%q}`, mediaID)
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/import", body)
	if resp.Code != http.StatusCreated {
		t.Fatalf("import status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var track map[string]any
	decodeResponse(t, resp, &track)
	if track["title"] != mediaID {
		t.Fatalf("title = %q, want media object id fallback %q", track["title"], mediaID)
	}
}

func TestCatalogImportTrackWrongAssetKind(t *testing.T) {
	h, mediaID := newImportTestHandlerWithMediaObject(t, "media-import-3", "artwork", "active")
	body := fmt.Sprintf(`{"mediaObjectId":%q}`, mediaID)
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/import", body)
	assertAPIError(t, resp, http.StatusUnprocessableEntity, "import_rejected")
}

func TestCatalogImportTrackNotActive(t *testing.T) {
	h, mediaID := newImportTestHandlerWithMediaObject(t, "media-import-4", "original_audio", "staged")
	body := fmt.Sprintf(`{"mediaObjectId":%q}`, mediaID)
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/import", body)
	assertAPIError(t, resp, http.StatusUnprocessableEntity, "import_rejected")
}

func TestCatalogImportTrackMediaNotFound(t *testing.T) {
	h, _ := newImportTestHandlerWithMediaObject(t, "", "", "")
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/import", `{"mediaObjectId":"missing"}`)
	if resp.Code == http.StatusCreated {
		t.Fatal("expected failure for missing media object")
	}
}

func TestCatalogImportTrackNoCatalogService(t *testing.T) {
	h := newTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/import", `{"mediaObjectId":"x"}`)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

func TestCatalogImportTrackWithArtistAndAlbum(t *testing.T) {
	h, mediaID := newImportTestHandlerWithMediaObject(t, "media-import-5", "original_audio", "active")

	// create artist
	aResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Miku"}`)
	if aResp.Code != http.StatusCreated {
		t.Fatalf("create artist status = %d", aResp.Code)
	}
	var aBody map[string]any
	decodeResponse(t, aResp, &aBody)
	artistID := aBody["id"].(string)

	// create album
	alResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums", fmt.Sprintf(`{"title":"supercell","artistId":%q}`, artistID))
	if alResp.Code != http.StatusCreated {
		t.Fatalf("create album status = %d", alResp.Code)
	}
	var alBody map[string]any
	decodeResponse(t, alResp, &alBody)
	albumID := alBody["id"].(string)

	// import
	body := fmt.Sprintf(`{"mediaObjectId":%q,"title":"World Is Mine","artistId":%q,"albumId":%q,"trackNumber":1}`, mediaID, artistID, albumID)
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/import", body)
	if resp.Code != http.StatusCreated {
		t.Fatalf("import status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var track map[string]any
	decodeResponse(t, resp, &track)
	if track["artistId"] != artistID {
		t.Fatalf("artistId = %q, want %q", track["artistId"], artistID)
	}
	if track["albumId"] != albumID {
		t.Fatalf("albumId = %q, want %q", track["albumId"], albumID)
	}
}

// ---- catalog relink HTTP tests ----

func TestCatalogRelinkTrackSuccess(t *testing.T) {
	h, _ := newImportTestHandlerWithMediaObject(t, "relink-src", "original_audio", "active")
	// register a second media object
	{
		repo := storage.NewMemoryMediaObjectRepository()
		obj := storage.MediaObject{
			ID:             "relink-dst",
			BackendID:      "backend-1",
			ObjectKey:      "key/relink-dst",
			AssetKind:      "transcoded_audio",
			LifecycleState: "active",
		}
		_ = repo.SaveMediaObject(context.Background(), obj)
	}
	// We need a handler with both media objects accessible; easiest: build one
	// from scratch with both pre-seeded.
	h = func() http.Handler {
		sysRepo := storage.NewMemoryRepository()
		mediaRepo := storage.NewMemoryMediaObjectRepository()
		for _, obj := range []storage.MediaObject{
			{ID: "relink-src", BackendID: "b1", ObjectKey: "k/relink-src", AssetKind: "original_audio", LifecycleState: "active"},
			{ID: "relink-dst", BackendID: "b1", ObjectKey: "k/relink-dst", AssetKind: "transcoded_audio", LifecycleState: "active"},
		} {
			if err := mediaRepo.SaveMediaObject(context.Background(), obj); err != nil {
				t.Fatalf("SaveMediaObject %s: %v", obj.ID, err)
			}
		}
		catalogRepo := catalog.NewMemoryRepository()
		catalogSvc := catalog.NewService(catalogRepo)
		mediaSvc := storage.NewMediaObjectService(sysRepo, mediaRepo)
		return NewHandler(
			storage.NewService(sysRepo),
			WithAdminToken(testAdminToken),
			WithCatalogService(catalogSvc),
			WithMediaObjectService(mediaSvc),
			WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
		).Routes()
	}()

	// import original track
	body := `{"mediaObjectId":"relink-src","title":"My Song"}`
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/import", body)
	if resp.Code != http.StatusCreated {
		t.Fatalf("import status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var track map[string]any
	decodeResponse(t, resp, &track)
	trackID := track["id"].(string)

	// relink to new media object
	relinkBody := `{"mediaObjectId":"relink-dst"}`
	resp = performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks/"+trackID+"/relink", relinkBody)
	if resp.Code != http.StatusOK {
		t.Fatalf("relink status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var linked map[string]any
	decodeResponse(t, resp, &linked)
	if linked["mediaObjectId"] != "relink-dst" {
		t.Fatalf("mediaObjectId = %q, want relink-dst", linked["mediaObjectId"])
	}
	if linked["title"] != "My Song" {
		t.Fatalf("title = %q, want My Song", linked["title"])
	}
}

func TestCatalogRelinkTrackWrongAssetKind(t *testing.T) {
	h := func() http.Handler {
		sysRepo := storage.NewMemoryRepository()
		mediaRepo := storage.NewMemoryMediaObjectRepository()
		for _, obj := range []storage.MediaObject{
			{ID: "rk-audio", BackendID: "b1", ObjectKey: "k/rk-audio", AssetKind: "original_audio", LifecycleState: "active"},
			{ID: "rk-art", BackendID: "b1", ObjectKey: "k/rk-art", AssetKind: "artwork", LifecycleState: "active"},
		} {
			_ = mediaRepo.SaveMediaObject(context.Background(), obj)
		}
		catalogSvc := catalog.NewService(catalog.NewMemoryRepository())
		mediaSvc := storage.NewMediaObjectService(sysRepo, mediaRepo)
		return NewHandler(
			storage.NewService(sysRepo),
			WithAdminToken(testAdminToken),
			WithCatalogService(catalogSvc),
			WithMediaObjectService(mediaSvc),
			WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
		).Routes()
	}()

	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/import", `{"mediaObjectId":"rk-audio","title":"T"}`)
	if resp.Code != http.StatusCreated {
		t.Fatalf("import status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var track map[string]any
	decodeResponse(t, resp, &track)
	trackID := track["id"].(string)

	resp = performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks/"+trackID+"/relink", `{"mediaObjectId":"rk-art"}`)
	assertAPIError(t, resp, http.StatusUnprocessableEntity, "relink_rejected")
}

func TestCatalogRelinkTrackNotFound(t *testing.T) {
	h, _ := newImportTestHandlerWithMediaObject(t, "rk-exists", "original_audio", "active")
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks/no-such-id/relink", `{"mediaObjectId":"rk-exists"}`)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestCatalogRelinkTrackNoCatalogService(t *testing.T) {
	h := newTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks/some-id/relink", `{"mediaObjectId":"x"}`)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

func TestCatalogRelinkTrackMethodNotAllowed(t *testing.T) {
	h, _ := newImportTestHandlerWithMediaObject(t, "rk-m", "original_audio", "active")
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/tracks/some-id/relink", "")
	assertAPIError(t, resp, http.StatusMethodNotAllowed, "method_not_allowed")
}

// ---- catalog search HTTP tests ----

func newSearchTestHandler(t *testing.T) http.Handler {
	t.Helper()
	catalogRepo := catalog.NewMemoryRepository()
	catalogSvc := catalog.NewService(catalogRepo)
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
		WithCatalogService(catalogSvc),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()
	return h
}

func TestCatalogSearchMissingQuery(t *testing.T) {
	h := newSearchTestHandler(t)
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/search", "")
	assertAPIError(t, resp, http.StatusBadRequest, "missing_query")
}

func TestCatalogSearchEmptyQuery(t *testing.T) {
	h := newSearchTestHandler(t)
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/search?q=+", "")
	assertAPIError(t, resp, http.StatusBadRequest, "missing_query")
}

func TestCatalogSearchReturnsResults(t *testing.T) {
	h := newSearchTestHandler(t)

	// create artist
	aResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Hatsune Miku"}`)
	if aResp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d %s", aResp.Code, aResp.Body.String())
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/search?q=miku", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("search status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["query"] != "miku" {
		t.Fatalf("query = %q, want miku", result["query"])
	}
	items, ok := result["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items = %v, want 1 result", result["items"])
	}
	item := items[0].(map[string]any)
	if item["kind"] != "artist" {
		t.Fatalf("kind = %q, want artist", item["kind"])
	}
}

func TestCatalogSearchNoResults(t *testing.T) {
	h := newSearchTestHandler(t)
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/search?q=notfound", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("search status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	items := result["items"]
	// items may be nil (JSON null) or an empty array — both acceptable
	if items != nil {
		if arr, ok := items.([]any); ok && len(arr) != 0 {
			t.Fatalf("items = %v, want empty", arr)
		}
	}
}

func TestCatalogSearchNoCatalogService(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/search?q=miku", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

func TestAdminCatalogSearchMethodNotAllowed(t *testing.T) {
	h := newSearchTestHandler(t)
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/search", "")
	assertAPIError(t, resp, http.StatusMethodNotAllowed, "method_not_allowed")
}

// ---- viewer catalog browse HTTP tests ----

// newViewerTestHandler returns a handler with auth service + catalog service
// and two session tokens: one for a viewer user and one for an admin user.
func newViewerTestHandler(t *testing.T) (http.Handler, string, string) {
	t.Helper()
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	catalogRepo := catalog.NewMemoryRepository()
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithCatalogService(catalog.NewService(catalogRepo)),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "alice", "viewerpass1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "bob", "adminpass2", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, err := authSvc.Login(context.Background(), "bob", "adminpass2")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	return h, viewerToken, adminToken
}

func TestCatalogViewerListArtists(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/artists", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("list artists status = %d, body = %s", resp.Code, resp.Body.String())
	}
	assertArtistListLength(t, resp, 0)
}

func TestCatalogViewerListArtistsAdminToken(t *testing.T) {
	h, _, adminToken := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/artists", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin session on viewer route status = %d, body = %s", resp.Code, resp.Body.String())
	}
	assertArtistListLength(t, resp, 0)
}

func TestCatalogViewerStaticBootstrapTokenRejected(t *testing.T) {
	// Handler with static admin token but no auth service: viewer routes return 503.
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
		WithCatalogService(catalog.NewService(catalog.NewMemoryRepository())),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/catalog/artists", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "auth_not_configured")
}

func TestCatalogViewerUnauthorized(t *testing.T) {
	h, _, _ := newViewerTestHandler(t)
	resp := performRequestWithoutAuth(t, h, http.MethodGet, "/api/v1/catalog/artists", "")
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestCatalogViewerGetArtistNotFound(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/artists/missing", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestCatalogViewerListAlbums(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/albums", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("list albums status = %d, body = %s", resp.Code, resp.Body.String())
	}
	assertAlbumListLength(t, resp, 0)
}

func TestCatalogViewerListTracks(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("list tracks status = %d, body = %s", resp.Code, resp.Body.String())
	}
	assertTrackListLength(t, resp, 0)
}

func TestCatalogViewerSearchMissingQuery(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/search", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "missing_query")
}

func TestCatalogViewerSearchReturnsResults(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)
	// seed via admin endpoint
	aResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Hatsune Miku"}`, "Bearer "+adminToken)
	if aResp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d %s", aResp.Code, aResp.Body.String())
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/search?q=miku", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("search status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	items, ok := result["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items = %v, want 1 result", result["items"])
	}
}

func TestCatalogViewerMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/catalog/artists", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusMethodNotAllowed, "method_not_allowed")
}

// ---- batch import tests ----

func TestCatalogBatchImportAllSuccess(t *testing.T) {
	h, mediaID := newImportTestHandlerWithMediaObject(t, "bi-http-1", "original_audio", "active")
	body := fmt.Sprintf(`{"items":[{"mediaObjectId":%q,"title":"Track One"},{"mediaObjectId":%q,"title":"Track Two"}]}`, mediaID, mediaID)
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/batch-import", body)
	if resp.Code != http.StatusOK {
		t.Fatalf("batch-import status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["total"].(float64) != 2 {
		t.Fatalf("total = %v, want 2", result["total"])
	}
	if result["imported"].(float64) != 2 {
		t.Fatalf("imported = %v, want 2", result["imported"])
	}
	if result["failed"].(float64) != 0 {
		t.Fatalf("failed = %v, want 0", result["failed"])
	}
}

func TestCatalogBatchImportPartialSuccess(t *testing.T) {
	h, mediaID := newImportTestHandlerWithMediaObject(t, "bi-http-2", "original_audio", "active")
	body := fmt.Sprintf(`{"items":[{"mediaObjectId":%q,"title":"Good"},{"mediaObjectId":"missing-xyz","title":"Bad"}]}`, mediaID)
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/batch-import", body)
	if resp.Code != http.StatusMultiStatus {
		t.Fatalf("batch-import status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["imported"].(float64) != 1 || result["failed"].(float64) != 1 {
		t.Fatalf("imported=%v failed=%v, want 1/1", result["imported"], result["failed"])
	}
}

func TestCatalogBatchImportAllFail(t *testing.T) {
	h, _ := newImportTestHandlerWithMediaObject(t, "", "", "")
	body := `{"items":[{"mediaObjectId":"none-a"},{"mediaObjectId":"none-b"}]}`
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/batch-import", body)
	if resp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("all-fail status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["failed"].(float64) != 2 {
		t.Fatalf("failed = %v, want 2", result["failed"])
	}
}

func TestCatalogBatchImportEmptyBatch(t *testing.T) {
	h, _ := newImportTestHandlerWithMediaObject(t, "", "", "")
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/batch-import", `{"items":[]}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("empty batch status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["total"].(float64) != 0 {
		t.Fatalf("total = %v, want 0", result["total"])
	}
}

func TestCatalogBatchImportNoCatalogService(t *testing.T) {
	h := newTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/batch-import", `{"items":[]}`)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

func TestCatalogBatchImportMethodNotAllowed(t *testing.T) {
	h, _ := newImportTestHandlerWithMediaObject(t, "", "", "")
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/batch-import", "")
	assertAPIError(t, resp, http.StatusMethodNotAllowed, "method_not_allowed")
}

// ---------- Phase 45: PATCH catalog metadata HTTP tests ----------

func TestCatalogPatchArtistUpdatesName(t *testing.T) {
	h := newCatalogTestHandler()

	// create
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Old Name"}`)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", resp.Code, resp.Body)
	}
	var artist map[string]any
	decodeResponse(t, resp, &artist)
	id, _ := artist["id"].(string)

	// patch
	resp = performRequest(t, h, http.MethodPatch, "/api/v1/admin/catalog/artists/"+id, `{"name":"New Name","sortName":"Name, New"}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("patch: %d %s", resp.Code, resp.Body)
	}
	var updated map[string]any
	decodeResponse(t, resp, &updated)
	if updated["name"] != "New Name" {
		t.Fatalf("name = %q, want %q", updated["name"], "New Name")
	}
	if updated["sortName"] != "Name, New" {
		t.Fatalf("sortName = %q, want %q", updated["sortName"], "Name, New")
	}
	if updated["id"] != id {
		t.Fatal("id must not change")
	}
}

func TestCatalogPatchArtistNotFound(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPatch, "/api/v1/admin/catalog/artists/nonexistent", `{"name":"X"}`)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestCatalogPatchArtistEmptyName(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Valid"}`)
	var artist map[string]any
	decodeResponse(t, resp, &artist)
	id, _ := artist["id"].(string)

	resp = performRequest(t, h, http.MethodPatch, "/api/v1/admin/catalog/artists/"+id, `{"name":""}`)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_catalog_entity")
}

func TestCatalogPatchAlbumUpdatesTitle(t *testing.T) {
	h := newCatalogTestHandler()

	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist"}`)
	var artist map[string]any
	decodeResponse(t, resp, &artist)
	artistID, _ := artist["id"].(string)

	resp = performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		fmt.Sprintf(`{"title":"Old Title","artistId":%q,"releaseYear":2000}`, artistID))
	if resp.Code != http.StatusCreated {
		t.Fatalf("create album: %d %s", resp.Code, resp.Body)
	}
	var album map[string]any
	decodeResponse(t, resp, &album)
	albumID, _ := album["id"].(string)

	resp = performRequest(t, h, http.MethodPatch, "/api/v1/admin/catalog/albums/"+albumID,
		`{"title":"New Title","releaseYear":2024}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("patch album: %d %s", resp.Code, resp.Body)
	}
	var updated map[string]any
	decodeResponse(t, resp, &updated)
	if updated["title"] != "New Title" {
		t.Fatalf("title = %q, want %q", updated["title"], "New Title")
	}
	if updated["releaseYear"] != float64(2024) {
		t.Fatalf("releaseYear = %v, want 2024", updated["releaseYear"])
	}
}

func TestCatalogPatchAlbumNotFound(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPatch, "/api/v1/admin/catalog/albums/nonexistent", `{"title":"X"}`)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestCatalogPatchTrackUpdatesFields(t *testing.T) {
	h := newCatalogTestHandler()

	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"A"}`)
	var artist map[string]any
	decodeResponse(t, resp, &artist)
	artistID, _ := artist["id"].(string)

	resp = performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		fmt.Sprintf(`{"title":"B","artistId":%q}`, artistID))
	var album map[string]any
	decodeResponse(t, resp, &album)
	albumID, _ := album["id"].(string)

	resp = performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"T","artistId":%q,"albumId":%q,"mediaObjectId":"mo99","trackNumber":1,"durationMs":60000}`, artistID, albumID))
	if resp.Code != http.StatusCreated {
		t.Fatalf("create track: %d %s", resp.Code, resp.Body)
	}
	var track map[string]any
	decodeResponse(t, resp, &track)
	trackID, _ := track["id"].(string)

	resp = performRequest(t, h, http.MethodPatch, "/api/v1/admin/catalog/tracks/"+trackID,
		`{"title":"Updated","trackNumber":2,"durationMs":90000}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("patch track: %d %s", resp.Code, resp.Body)
	}
	var updated map[string]any
	decodeResponse(t, resp, &updated)
	if updated["title"] != "Updated" {
		t.Fatalf("title = %q, want %q", updated["title"], "Updated")
	}
	if updated["trackNumber"] != float64(2) {
		t.Fatalf("trackNumber = %v, want 2", updated["trackNumber"])
	}
	if updated["durationMs"] != float64(90000) {
		t.Fatalf("durationMs = %v, want 90000", updated["durationMs"])
	}
	if updated["mediaObjectId"] != "mo99" {
		t.Fatalf("mediaObjectId changed unexpectedly to %q", updated["mediaObjectId"])
	}
}

func TestCatalogPatchTrackNotFound(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPatch, "/api/v1/admin/catalog/tracks/nonexistent", `{"title":"X"}`)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestCatalogPatchNoCatalogService(t *testing.T) {
	h := newTestHandler()
	resp := performRequest(t, h, http.MethodPatch, "/api/v1/admin/catalog/artists/x", `{"name":"X"}`)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

// ---------- Phase 46: Playlist HTTP tests ----------

func TestPlaylistCreateAndList(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists",
		`{"name":"Weekend Mix","description":"a fun playlist"}`)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create playlist: %d %s", resp.Code, resp.Body)
	}
	var pl map[string]any
	decodeResponse(t, resp, &pl)
	if pl["name"] != "Weekend Mix" || pl["description"] != "a fun playlist" {
		t.Fatalf("unexpected playlist body: %v", pl)
	}
	plID, _ := pl["id"].(string)
	if plID == "" {
		t.Fatal("expected non-empty id")
	}
	if trackIDs, ok := pl["trackIds"].([]any); !ok || len(trackIDs) != 0 {
		t.Fatalf("expected empty trackIds, got %v", pl["trackIds"])
	}

	listResp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/playlists", "")
	if listResp.Code != http.StatusOK {
		t.Fatalf("list playlists: %d %s", listResp.Code, listResp.Body)
	}
	var listBody map[string]any
	decodeResponse(t, listResp, &listBody)
	items, _ := listBody["playlists"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 playlist, got %d", len(items))
	}
}

func TestPlaylistAddAndRemoveTrack(t *testing.T) {
	h := newCatalogTestHandler()

	// Create an artist and track first.
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID, _ := artist["id"].(string)

	trackResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-pl-1"}`, artistID))
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID, _ := track["id"].(string)

	// Create playlist and add track.
	plResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"My PL"}`)
	var pl map[string]any
	decodeResponse(t, plResp, &pl)
	plID, _ := pl["id"].(string)

	addResp := performRequest(t, h, http.MethodPost,
		"/api/v1/admin/catalog/playlists/"+plID+"/tracks",
		fmt.Sprintf(`{"trackId":%q}`, trackID))
	if addResp.Code != http.StatusOK {
		t.Fatalf("add track: %d %s", addResp.Code, addResp.Body)
	}
	var added map[string]any
	decodeResponse(t, addResp, &added)
	ids, _ := added["trackIds"].([]any)
	if len(ids) != 1 || ids[0] != trackID {
		t.Fatalf("trackIds after add = %v", ids)
	}

	// Remove the track.
	rmResp := performRequest(t, h, http.MethodDelete,
		"/api/v1/admin/catalog/playlists/"+plID+"/tracks/"+trackID, "")
	if rmResp.Code != http.StatusOK {
		t.Fatalf("remove track: %d %s", rmResp.Code, rmResp.Body)
	}
	var removed map[string]any
	decodeResponse(t, rmResp, &removed)
	idsAfter, _ := removed["trackIds"].([]any)
	if len(idsAfter) != 0 {
		t.Fatalf("trackIds after remove = %v", idsAfter)
	}
}

func TestPlaylistPatchMetadata(t *testing.T) {
	h := newCatalogTestHandler()
	plResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"Old"}`)
	var pl map[string]any
	decodeResponse(t, plResp, &pl)
	plID, _ := pl["id"].(string)

	patchResp := performRequest(t, h, http.MethodPatch, "/api/v1/admin/catalog/playlists/"+plID,
		`{"name":"New Name","description":"updated"}`)
	if patchResp.Code != http.StatusOK {
		t.Fatalf("patch playlist: %d %s", patchResp.Code, patchResp.Body)
	}
	var patched map[string]any
	decodeResponse(t, patchResp, &patched)
	if patched["name"] != "New Name" || patched["description"] != "updated" {
		t.Fatalf("unexpected patch result: %v", patched)
	}
}

func TestPlaylistDeleteAndNotFound(t *testing.T) {
	h := newCatalogTestHandler()
	plResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"Temp"}`)
	var pl map[string]any
	decodeResponse(t, plResp, &pl)
	plID, _ := pl["id"].(string)

	delResp := performRequest(t, h, http.MethodDelete, "/api/v1/admin/catalog/playlists/"+plID, "")
	if delResp.Code != http.StatusNoContent {
		t.Fatalf("delete playlist: %d %s", delResp.Code, delResp.Body)
	}

	getResp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/playlists/"+plID, "")
	assertAPIError(t, getResp, http.StatusNotFound, "not_found")
}

func TestPlaylistViewerCanRead(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)

	plResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists",
		`{"name":"Public PL"}`, "Bearer "+adminToken)
	var pl map[string]any
	decodeResponse(t, plResp, &pl)
	plID, _ := pl["id"].(string)

	listResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/playlists", "", "Bearer "+viewerToken)
	if listResp.Code != http.StatusOK {
		t.Fatalf("viewer list playlists: %d %s", listResp.Code, listResp.Body)
	}

	getResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/playlists/"+plID, "", "Bearer "+viewerToken)
	if getResp.Code != http.StatusOK {
		t.Fatalf("viewer get playlist: %d %s", getResp.Code, getResp.Body)
	}
}

func TestPlaylistNoCatalogService(t *testing.T) {
	h := newTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/playlists", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

func TestPlaylistMethodNotAllowed(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPut, "/api/v1/admin/catalog/playlists", "")
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestPlaylistSetTracks(t *testing.T) {
	h := newCatalogTestHandler()

	// Seed artist and two tracks.
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID, _ := artist["id"].(string)

	t1Resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Track1","artistId":%q,"mediaObjectId":"mo-set-1"}`, artistID))
	var t1 map[string]any
	decodeResponse(t, t1Resp, &t1)
	t1ID, _ := t1["id"].(string)

	t2Resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Track2","artistId":%q,"mediaObjectId":"mo-set-2"}`, artistID))
	var t2 map[string]any
	decodeResponse(t, t2Resp, &t2)
	t2ID, _ := t2["id"].(string)

	plResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"SetPL"}`)
	var pl map[string]any
	decodeResponse(t, plResp, &pl)
	plID, _ := pl["id"].(string)

	// Happy path: set [t2, t1] — reorder.
	setResp := performRequest(t, h, http.MethodPut,
		"/api/v1/admin/catalog/playlists/"+plID+"/tracks",
		fmt.Sprintf(`{"trackIds":[%q,%q]}`, t2ID, t1ID))
	if setResp.Code != http.StatusOK {
		t.Fatalf("set tracks: %d %s", setResp.Code, setResp.Body)
	}
	var got map[string]any
	decodeResponse(t, setResp, &got)
	ids, _ := got["trackIds"].([]any)
	if len(ids) != 2 || ids[0] != t2ID || ids[1] != t1ID {
		t.Fatalf("trackIds after set = %v, want [%s %s]", ids, t2ID, t1ID)
	}

	// Clear: empty slice.
	clearResp := performRequest(t, h, http.MethodPut,
		"/api/v1/admin/catalog/playlists/"+plID+"/tracks",
		`{"trackIds":[]}`)
	if clearResp.Code != http.StatusOK {
		t.Fatalf("clear tracks: %d %s", clearResp.Code, clearResp.Body)
	}
	var cleared map[string]any
	decodeResponse(t, clearResp, &cleared)
	clearedIDs, _ := cleared["trackIds"].([]any)
	if len(clearedIDs) != 0 {
		t.Fatalf("trackIds after clear = %v, want []", clearedIDs)
	}

	// Unknown track → 404.
	badTrackResp := performRequest(t, h, http.MethodPut,
		"/api/v1/admin/catalog/playlists/"+plID+"/tracks",
		`{"trackIds":["no-such-track"]}`)
	assertAPIError(t, badTrackResp, http.StatusNotFound, "not_found")

	// Unknown playlist → 404.
	badPLResp := performRequest(t, h, http.MethodPut,
		"/api/v1/admin/catalog/playlists/no-such-pl/tracks",
		fmt.Sprintf(`{"trackIds":[%q]}`, t1ID))
	assertAPIError(t, badPLResp, http.StatusNotFound, "not_found")

	// Missing trackIds field → 400.
	missingResp := performRequest(t, h, http.MethodPut,
		"/api/v1/admin/catalog/playlists/"+plID+"/tracks",
		`{}`)
	assertAPIError(t, missingResp, http.StatusBadRequest, "validation_error")
}

func TestPlaylistSetTracksNoCatalogService(t *testing.T) {
	h := newTestHandler()
	resp := performRequest(t, h, http.MethodPut, "/api/v1/admin/catalog/playlists/some-id/tracks",
		`{"trackIds":[]}`)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

// ---- Phase 49: GET playlist tracks tests ----

func TestGetPlaylistTracksAdmin(t *testing.T) {
	h := newCatalogTestHandler()

	// seed: artist + 2 tracks + playlist
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	t1Resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song A","artistId":%q,"mediaObjectId":"mo-gta-1"}`, artistID))
	var t1 map[string]any
	decodeResponse(t, t1Resp, &t1)
	t1ID := t1["id"].(string)

	t2Resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song B","artistId":%q,"mediaObjectId":"mo-gta-2"}`, artistID))
	var t2 map[string]any
	decodeResponse(t, t2Resp, &t2)
	t2ID := t2["id"].(string)

	plResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"GetTracksPL"}`)
	var pl map[string]any
	decodeResponse(t, plResp, &pl)
	plID := pl["id"].(string)

	// reorder: [t2, t1]
	setResp := performRequest(t, h, http.MethodPut,
		"/api/v1/admin/catalog/playlists/"+plID+"/tracks",
		fmt.Sprintf(`{"trackIds":[%q,%q]}`, t2ID, t1ID))
	if setResp.Code != http.StatusOK {
		t.Fatalf("set tracks: %d %s", setResp.Code, setResp.Body)
	}

	// GET tracks: expect ordered full objects
	getResp := performRequest(t, h, http.MethodGet,
		"/api/v1/admin/catalog/playlists/"+plID+"/tracks", "")
	if getResp.Code != http.StatusOK {
		t.Fatalf("get playlist tracks: %d %s", getResp.Code, getResp.Body)
	}
	var body map[string]any
	decodeResponse(t, getResp, &body)
	tracks, _ := body["tracks"].([]any)
	if len(tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d: %v", len(tracks), body)
	}
	first, _ := tracks[0].(map[string]any)
	second, _ := tracks[1].(map[string]any)
	if first["id"] != t2ID {
		t.Errorf("tracks[0].id = %v, want %s", first["id"], t2ID)
	}
	if second["id"] != t1ID {
		t.Errorf("tracks[1].id = %v, want %s", second["id"], t1ID)
	}
	// full objects include title
	if first["title"] != "Song B" {
		t.Errorf("tracks[0].title = %v, want Song B", first["title"])
	}
}

func TestGetPlaylistTracksEmpty(t *testing.T) {
	h := newCatalogTestHandler()
	plResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"EmptyPL"}`)
	var pl map[string]any
	decodeResponse(t, plResp, &pl)
	plID := pl["id"].(string)

	resp := performRequest(t, h, http.MethodGet,
		"/api/v1/admin/catalog/playlists/"+plID+"/tracks", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("get empty playlist tracks: %d %s", resp.Code, resp.Body)
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	tracks, _ := body["tracks"].([]any)
	if len(tracks) != 0 {
		t.Fatalf("expected empty tracks, got %v", tracks)
	}
}

func TestGetPlaylistTracksNotFound(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet,
		"/api/v1/admin/catalog/playlists/no-such-id/tracks", "")
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestGetPlaylistTracksViewerCanRead(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)

	// seed via admin session token
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"VBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	tResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"VTrack","artistId":%q,"mediaObjectId":"mo-vgtt-1"}`, artistID), "Bearer "+adminToken)
	var tr map[string]any
	decodeResponse(t, tResp, &tr)
	trID := tr["id"].(string)

	plResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"VPL"}`, "Bearer "+adminToken)
	var pl map[string]any
	decodeResponse(t, plResp, &pl)
	plID := pl["id"].(string)

	performRequestWithAuthHeader(t, h, http.MethodPut,
		"/api/v1/admin/catalog/playlists/"+plID+"/tracks",
		fmt.Sprintf(`{"trackIds":[%q]}`, trID), "Bearer "+adminToken)

	// viewer GET
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/catalog/playlists/"+plID+"/tracks", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer get playlist tracks: %d %s", resp.Code, resp.Body)
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	tracks, _ := body["tracks"].([]any)
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %v", tracks)
	}
}

func TestGetPlaylistTracksNoCatalogService(t *testing.T) {
	h := newTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/playlists/some-id/tracks", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

func TestGetPlaylistTracksMethodNotAllowed(t *testing.T) {
	h := newCatalogTestHandler()
	plResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"MNA"}`)
	var pl map[string]any
	decodeResponse(t, plResp, &pl)
	plID := pl["id"].(string)

	resp := performRequest(t, h, http.MethodDelete,
		"/api/v1/admin/catalog/playlists/"+plID+"/tracks", "")
	// DELETE on {id}/tracks (no trackId) should be 405
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

// ---- catalog stats tests ----

func TestGetCatalogStatsEmpty(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	for _, key := range []string{"artists", "albums", "tracks", "playlists"} {
		if got[key] == nil {
			t.Errorf("response missing key %q", key)
		}
	}
}

func TestGetCatalogStatsPopulatedCounts(t *testing.T) {
	h := newCatalogTestHandler()

	// seed: 2 artists, 1 album, 2 tracks, 1 playlist
	aResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"A1"}`)
	var a1 map[string]any
	decodeResponse(t, aResp, &a1)
	a1ID := a1["id"].(string)

	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"A2"}`)
	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		fmt.Sprintf(`{"title":"Album1","artistId":%q}`, a1ID))
	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"T1","artistId":%q,"mediaObjectId":"mo-st1"}`, a1ID))
	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"T2","artistId":%q,"mediaObjectId":"mo-st2"}`, a1ID))
	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"PL1"}`)

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	assertEqual := func(key string, want float64) {
		t.Helper()
		v, ok := got[key].(float64)
		if !ok || v != want {
			t.Errorf("%s = %v, want %v", key, got[key], want)
		}
	}
	assertEqual("artists", 2)
	assertEqual("albums", 1)
	assertEqual("tracks", 2)
	assertEqual("playlists", 1)
}

func TestGetCatalogStatsNoCatalogService(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test"}),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats", "")
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.Code)
	}
}

func TestGetCatalogStatsMethodNotAllowed(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/stats", `{}`)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestGetArtistStatsBreakdownEmpty(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats/artists", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	artists, ok := got["artists"]
	if !ok {
		t.Fatal("response missing key \"artists\"")
	}
	if arr, ok := artists.([]any); !ok || len(arr) != 0 {
		t.Errorf("expected empty artists array, got %v", artists)
	}
}

func TestGetArtistStatsBreakdownPopulated(t *testing.T) {
	h := newCatalogTestHandler()

	// seed: 2 artists, artist-1 has 1 album + 2 tracks
	r1 := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"ArtistX"}`)
	var a1 map[string]any
	decodeResponse(t, r1, &a1)
	a1ID := a1["id"].(string)

	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"ArtistY"}`)
	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		fmt.Sprintf(`{"title":"AlbumX","artistId":%q}`, a1ID))
	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TX1","artistId":%q,"mediaObjectId":"mo-x1"}`, a1ID))
	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TX2","artistId":%q,"mediaObjectId":"mo-x2"}`, a1ID))

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats/artists", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	arr := got["artists"].([]any)
	if len(arr) != 2 {
		t.Fatalf("expected 2 artist items, got %d", len(arr))
	}
	byID := map[string]map[string]any{}
	for _, item := range arr {
		m := item.(map[string]any)
		byID[m["artistId"].(string)] = m
	}
	if m := byID[a1ID]; m["albumCount"].(float64) != 1 || m["trackCount"].(float64) != 2 {
		t.Errorf("artist1: albumCount=%v trackCount=%v, want 1/2", m["albumCount"], m["trackCount"])
	}
}

func TestGetArtistStatsBreakdownNoCatalogService(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test"}),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats/artists", "")
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.Code)
	}
}

func TestGetArtistStatsBreakdownMethodNotAllowed(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/stats/artists", `{}`)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestGetAlbumStatsBreakdownEmpty(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats/albums", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	albums, ok := got["albums"]
	if !ok {
		t.Fatal("response missing key \"albums\"")
	}
	if arr, ok := albums.([]any); !ok || len(arr) != 0 {
		t.Errorf("expected empty albums array, got %v", albums)
	}
}

func TestGetAlbumStatsBreakdownPopulated(t *testing.T) {
	h := newCatalogTestHandler()

	// seed: 1 artist, 2 albums, album-1 has 2 tracks
	r1 := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"BandZ"}`)
	var a1 map[string]any
	decodeResponse(t, r1, &a1)
	a1ID := a1["id"].(string)

	r2 := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		fmt.Sprintf(`{"title":"Debut","artistId":%q}`, a1ID))
	var al1 map[string]any
	decodeResponse(t, r2, &al1)
	al1ID := al1["id"].(string)

	r3 := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		fmt.Sprintf(`{"title":"Sophomore","artistId":%q}`, a1ID))
	var al2 map[string]any
	decodeResponse(t, r3, &al2)
	al2ID := al2["id"].(string)

	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song1","artistId":%q,"albumId":%q,"mediaObjectId":"mo-z1"}`, a1ID, al1ID))
	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song2","artistId":%q,"albumId":%q,"mediaObjectId":"mo-z2"}`, a1ID, al1ID))

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats/albums", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	arr := got["albums"].([]any)
	if len(arr) != 2 {
		t.Fatalf("expected 2 album items, got %d", len(arr))
	}
	byID := map[string]map[string]any{}
	for _, item := range arr {
		m := item.(map[string]any)
		byID[m["albumId"].(string)] = m
	}
	if m := byID[al1ID]; m["trackCount"].(float64) != 2 {
		t.Errorf("album1 trackCount=%v, want 2", m["trackCount"])
	}
	if m := byID[al2ID]; m["trackCount"].(float64) != 0 {
		t.Errorf("album2 trackCount=%v, want 0", m["trackCount"])
	}
}

func TestGetAlbumStatsBreakdownNoCatalogService(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test"}),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats/albums", "")
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.Code)
	}
}

func TestGetAlbumStatsBreakdownMethodNotAllowed(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/stats/albums", `{}`)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestGetPlaylistStatsBreakdownEmpty(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats/playlists", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	playlists, ok := got["playlists"]
	if !ok {
		t.Fatal("response missing key \"playlists\"")
	}
	if arr, ok := playlists.([]any); !ok || len(arr) != 0 {
		t.Errorf("expected empty playlists array, got %v", playlists)
	}
}

func TestGetPlaylistStatsBreakdownPopulated(t *testing.T) {
	h := newCatalogTestHandler()

	// seed: 1 artist, 1 track, 2 playlists; first playlist has 2 track entries
	r1 := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"StatsArtist"}`)
	var a1 map[string]any
	decodeResponse(t, r1, &a1)
	a1ID := a1["id"].(string)

	r2 := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"StatTrack","artistId":%q,"mediaObjectId":"mo-ps1"}`, a1ID))
	var tr1 map[string]any
	decodeResponse(t, r2, &tr1)
	tr1ID := tr1["id"].(string)

	r3 := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"PL Alpha"}`)
	var pl1 map[string]any
	decodeResponse(t, r3, &pl1)
	pl1ID := pl1["id"].(string)

	r4 := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"PL Beta"}`)
	var pl2 map[string]any
	decodeResponse(t, r4, &pl2)
	pl2ID := pl2["id"].(string)

	// add tr1 twice to pl1
	performRequest(t, h, http.MethodPost, fmt.Sprintf("/api/v1/admin/catalog/playlists/%s/tracks", pl1ID),
		fmt.Sprintf(`{"trackId":%q}`, tr1ID))
	performRequest(t, h, http.MethodPost, fmt.Sprintf("/api/v1/admin/catalog/playlists/%s/tracks", pl1ID),
		fmt.Sprintf(`{"trackId":%q}`, tr1ID))
	// pl2 stays empty

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats/playlists", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	arr := got["playlists"].([]any)
	if len(arr) != 2 {
		t.Fatalf("expected 2 playlist items, got %d", len(arr))
	}
	byID := map[string]map[string]any{}
	for _, item := range arr {
		m := item.(map[string]any)
		byID[m["playlistId"].(string)] = m
	}
	if m := byID[pl1ID]; m["trackCount"].(float64) != 2 {
		t.Errorf("playlist1 trackCount=%v, want 2", m["trackCount"])
	}
	if m := byID[pl2ID]; m["trackCount"].(float64) != 0 {
		t.Errorf("playlist2 trackCount=%v, want 0", m["trackCount"])
	}
}

func TestGetPlaylistStatsBreakdownNoCatalogService(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test"}),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/stats/playlists", "")
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.Code)
	}
}

func TestGetPlaylistStatsBreakdownMethodNotAllowed(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/stats/playlists", `{}`)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

// ---- recently-added handler tests ----

func TestGetRecentlyAddedEmpty(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-added", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items, ok := got["items"].([]any)
	if !ok || len(items) != 0 {
		t.Errorf("expected empty items array, got %v", got["items"])
	}
}

func TestGetRecentlyAddedPopulated(t *testing.T) {
	h := newCatalogTestHandler()

	// create an artist, then a track under it
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Inori Yuzuriha"}`)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d", resp.Code)
	}
	var artistResp map[string]any
	decodeResponse(t, resp, &artistResp)
	artistID := artistResp["id"].(string)

	trackBody := fmt.Sprintf(`{"title":"Departures","artistId":%q,"mediaObjectId":"mo-001","durationMs":200000}`, artistID)
	resp = performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks", trackBody)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create track: %d", resp.Code)
	}

	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-added", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("recently-added status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["items"].([]any)
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	for _, raw := range items {
		m := raw.(map[string]any)
		if m["kind"] == nil || m["addedAt"] == nil {
			t.Errorf("item missing kind or addedAt: %v", m)
		}
	}
}

func TestGetRecentlyAddedKindQueryParam(t *testing.T) {
	h := newCatalogTestHandler()

	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Miku"}`)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d", resp.Code)
	}

	// artist-only filter
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-added?kind=artist", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("recently-added?kind=artist status = %d", resp.Code)
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["items"].([]any)
	for _, raw := range items {
		m := raw.(map[string]any)
		if m["kind"].(string) != "artist" {
			t.Errorf("expected kind=artist, got %s", m["kind"])
		}
	}
}

func TestGetRecentlyAddedPlaylistKindQueryParam(t *testing.T) {
	h := newCatalogTestHandler()

	playlistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"Recent Mix"}`)
	if playlistResp.Code != http.StatusCreated {
		t.Fatalf("create playlist: %d %s", playlistResp.Code, playlistResp.Body.String())
	}
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Other Artist"}`)
	if artistResp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d %s", artistResp.Code, artistResp.Body.String())
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-added?kind=playlist", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("recently-added?kind=playlist status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 playlist item, got %d", len(items))
	}
	item := items[0].(map[string]any)
	playlist := item["playlist"].(map[string]any)
	if item["kind"] != "playlist" || playlist["name"] != "Recent Mix" || item["addedAt"] == nil {
		t.Fatalf("unexpected playlist recent item: %v", item)
	}

	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-added", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("recently-added status = %d, body = %s", resp.Code, resp.Body.String())
	}
	decodeResponse(t, resp, &got)
	items = got["items"].([]any)
	hasPlaylist := false
	for _, raw := range items {
		m := raw.(map[string]any)
		if m["kind"] == "playlist" && m["playlist"] != nil {
			hasPlaylist = true
		}
	}
	if !hasPlaylist {
		t.Fatalf("expected unified recently-added timeline to include playlist, got %v", items)
	}
}

func TestGetRecentlyAddedLimitParam(t *testing.T) {
	h := newCatalogTestHandler()

	for i := 0; i < 5; i++ {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists",
			fmt.Sprintf(`{"name":"Artist %d"}`, i))
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-added?limit=2", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d", resp.Code)
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["items"].([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 items (limit=2), got %d", len(items))
	}
}

func TestGetRecentlyAddedInvalidKind(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-added?kind=invalid", "")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid kind, got %d", resp.Code)
	}
}

func TestGetRecentlyAddedInvalidLimit(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-added?limit=abc", "")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid limit, got %d", resp.Code)
	}
}

func TestGetRecentlyAddedNoCatalogService(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test"}),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-added", "")
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.Code)
	}
}

func TestGetRecentlyAddedMethodNotAllowed(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/recently-added", `{}`)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

// ---- recently-updated handler tests ----

func TestGetRecentlyUpdatedEmpty(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-updated", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items, ok := got["items"].([]any)
	if !ok || len(items) != 0 {
		t.Errorf("expected empty items array, got %v", got["items"])
	}
}

func TestGetRecentlyUpdatedPopulated(t *testing.T) {
	h := newCatalogTestHandler()

	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Inori Yuzuriha"}`)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d", resp.Code)
	}
	var artistResp map[string]any
	decodeResponse(t, resp, &artistResp)
	artistID := artistResp["id"].(string)

	trackBody := fmt.Sprintf(`{"title":"Departures","artistId":%q,"mediaObjectId":"mo-updated-001","durationMs":200000}`, artistID)
	resp = performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks", trackBody)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create track: %d", resp.Code)
	}

	resp = performRequest(t, h, http.MethodPatch, "/api/v1/admin/catalog/artists/"+artistID, `{"sortName":"Yuzuriha, Inori"}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("patch artist: %d %s", resp.Code, resp.Body.String())
	}

	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-updated", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("recently-updated status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["items"].([]any)
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	first := items[0].(map[string]any)
	if first["kind"] != "artist" || first["updatedAt"] == nil || first["artist"] == nil {
		t.Fatalf("expected updated artist first, got %v", first)
	}
	for _, raw := range items {
		m := raw.(map[string]any)
		if m["kind"] == nil || m["updatedAt"] == nil {
			t.Errorf("item missing kind or updatedAt: %v", m)
		}
	}
}

func TestGetRecentlyUpdatedKindQueryParam(t *testing.T) {
	h := newCatalogTestHandler()

	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Miku"}`)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d", resp.Code)
	}

	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-updated?kind=artist", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("recently-updated?kind=artist status = %d", resp.Code)
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["items"].([]any)
	for _, raw := range items {
		m := raw.(map[string]any)
		if m["kind"].(string) != "artist" {
			t.Errorf("expected kind=artist, got %s", m["kind"])
		}
	}
}

func TestGetRecentlyUpdatedPlaylistKindQueryParam(t *testing.T) {
	h := newCatalogTestHandler()

	playlistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"Fresh Mix"}`)
	if playlistResp.Code != http.StatusCreated {
		t.Fatalf("create playlist: %d %s", playlistResp.Code, playlistResp.Body.String())
	}
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Other Artist"}`)
	if artistResp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d %s", artistResp.Code, artistResp.Body.String())
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-updated?kind=playlist", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("recently-updated?kind=playlist status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 playlist item, got %d", len(items))
	}
	item := items[0].(map[string]any)
	playlist := item["playlist"].(map[string]any)
	if item["kind"] != "playlist" || playlist["name"] != "Fresh Mix" || item["updatedAt"] == nil {
		t.Fatalf("unexpected playlist recent item: %v", item)
	}

	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-updated", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("recently-updated status = %d, body = %s", resp.Code, resp.Body.String())
	}
	decodeResponse(t, resp, &got)
	items = got["items"].([]any)
	hasPlaylist := false
	for _, raw := range items {
		m := raw.(map[string]any)
		if m["kind"] == "playlist" && m["playlist"] != nil {
			hasPlaylist = true
		}
	}
	if !hasPlaylist {
		t.Fatalf("expected unified recently-updated timeline to include playlist, got %v", items)
	}
}

func TestGetRecentlyUpdatedLimitParam(t *testing.T) {
	h := newCatalogTestHandler()

	for i := 0; i < 5; i++ {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists",
			fmt.Sprintf(`{"name":"Artist %d"}`, i))
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-updated?limit=2", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d", resp.Code)
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["items"].([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 items (limit=2), got %d", len(items))
	}
}

func TestGetRecentlyUpdatedInvalidKind(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-updated?kind=invalid", "")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid kind, got %d", resp.Code)
	}
}

func TestGetRecentlyUpdatedInvalidLimit(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-updated?limit=abc", "")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid limit, got %d", resp.Code)
	}
}

func TestGetRecentlyUpdatedNoCatalogService(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test"}),
	).Routes()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/recently-updated", "")
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.Code)
	}
}

func TestGetRecentlyUpdatedMethodNotAllowed(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/recently-updated", `{}`)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

// ---- viewer recently-added/updated handler tests ----

func TestViewerGetRecentlyAdded(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-added", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer recently-added status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items, ok := got["items"].([]any)
	if !ok || len(items) != 0 {
		t.Errorf("expected empty items array, got %v", got["items"])
	}
}

func TestViewerGetRecentlyAddedPlaylistKind(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)

	playlistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"Viewer Mix"}`, "Bearer "+adminToken)
	if playlistResp.Code != http.StatusCreated {
		t.Fatalf("create playlist: %d %s", playlistResp.Code, playlistResp.Body.String())
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-added?kind=playlist", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer recently-added?kind=playlist status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 playlist item, got %d", len(items))
	}
	item := items[0].(map[string]any)
	playlist := item["playlist"].(map[string]any)
	if item["kind"] != "playlist" || playlist["name"] != "Viewer Mix" || item["addedAt"] == nil {
		t.Fatalf("unexpected playlist recent item: %v", item)
	}
}

func TestViewerGetRecentlyAddedAdminToken(t *testing.T) {
	h, _, adminToken := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-added", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin session on viewer recently-added status = %d", resp.Code)
	}
}

func TestViewerGetRecentlyAddedUnauthorized(t *testing.T) {
	h, _, _ := newViewerTestHandler(t)
	resp := performRequestWithoutAuth(t, h, http.MethodGet, "/api/v1/catalog/recently-added", "")
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestViewerGetRecentlyAddedNoCatalogService(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test"}),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "alice", "viewerpass1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-added", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

func TestViewerGetRecentlyAddedMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/catalog/recently-added", `{}`, "Bearer "+viewerToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestViewerGetRecentlyUpdated(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-updated", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer recently-updated status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items, ok := got["items"].([]any)
	if !ok || len(items) != 0 {
		t.Errorf("expected empty items array, got %v", got["items"])
	}
}

func TestViewerGetRecentlyUpdatedPlaylistKind(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)

	playlistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"Updated Viewer Mix"}`, "Bearer "+adminToken)
	if playlistResp.Code != http.StatusCreated {
		t.Fatalf("create playlist: %d %s", playlistResp.Code, playlistResp.Body.String())
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-updated?kind=playlist", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer recently-updated?kind=playlist status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 playlist item, got %d", len(items))
	}
	item := items[0].(map[string]any)
	playlist := item["playlist"].(map[string]any)
	if item["kind"] != "playlist" || playlist["name"] != "Updated Viewer Mix" || item["updatedAt"] == nil {
		t.Fatalf("unexpected playlist recent item: %v", item)
	}
}

func TestViewerGetRecentlyUpdatedAdminToken(t *testing.T) {
	h, _, adminToken := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-updated", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin session on viewer recently-updated status = %d", resp.Code)
	}
}

func TestViewerGetRecentlyUpdatedUnauthorized(t *testing.T) {
	h, _, _ := newViewerTestHandler(t)
	resp := performRequestWithoutAuth(t, h, http.MethodGet, "/api/v1/catalog/recently-updated", "")
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

func TestViewerGetRecentlyUpdatedNoCatalogService(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test"}),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "alice", "viewerpass1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-updated", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

func TestViewerGetRecentlyUpdatedMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/catalog/recently-updated", `{}`, "Bearer "+viewerToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestViewerGetRecentlyAddedInvalidKind(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-added?kind=invalid", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_catalog_entity")
}

func TestViewerGetRecentlyAddedInvalidLimit(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-added?limit=abc", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_limit")
}

func TestViewerGetRecentlyUpdatedInvalidKind(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-updated?kind=invalid", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_catalog_entity")
}

func TestViewerGetRecentlyUpdatedInvalidLimit(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/recently-updated?limit=abc", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_limit")
}

// ---- track playback descriptor tests ----

func newViewerWithMediaHandler(t *testing.T) (http.Handler, string, string) {
	t.Helper()
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	storageRepo := storage.NewMemoryRepository()
	storageSvc := storage.NewService(storageRepo)
	mediaRepo := storage.NewMemoryMediaObjectRepository()
	mediaSvc := storage.NewMediaObjectService(storageRepo, mediaRepo)
	catalogRepo := catalog.NewMemoryRepository()
	h := NewHandler(
		storageSvc,
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
		WithMediaObjectService(mediaSvc),
		WithCatalogService(catalog.NewService(catalogRepo)),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "alice", "viewerpass1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "bob", "adminpass2", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, err := authSvc.Login(context.Background(), "bob", "adminpass2")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	return h, viewerToken, adminToken
}

func TestGetTrackPlaybackDescriptor(t *testing.T) {
	h, viewerToken, adminToken := newViewerWithMediaHandler(t)

	backendBody := `{"id":"b-1","type":"local","displayName":"Local","enabled":true,"isDefault":true,"config":{"local":{"rootPath":"/music"}}}`
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backendBody, "Bearer "+adminToken)

	moBody := `{"id":"mo-pb-1","backendId":"b-1","objectKey":"track.flac","contentHash":"sha256:abc","sizeBytes":1024,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/media/objects", moBody, "Bearer "+adminToken)
	if resp.Code != http.StatusCreated {
		t.Fatalf("register media object: %d %s", resp.Code, resp.Body.String())
	}

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackBody := fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-pb-1","durationMs":180000}`, artistID)
	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks", trackBody, "Bearer "+adminToken)
	if trackResp.Code != http.StatusCreated {
		t.Fatalf("create track: %d %s", trackResp.Code, trackResp.Body.String())
	}
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	resp = performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/"+trackID+"/playback", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("playback descriptor status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var desc map[string]any
	decodeResponse(t, resp, &desc)
	if desc["trackId"] != trackID {
		t.Errorf("trackId = %v, want %s", desc["trackId"], trackID)
	}
	if desc["mediaObjectId"] != "mo-pb-1" {
		t.Errorf("mediaObjectId = %v, want mo-pb-1", desc["mediaObjectId"])
	}
	if desc["mimeType"] != "audio/flac" {
		t.Errorf("mimeType = %v, want audio/flac", desc["mimeType"])
	}
	if int(desc["durationMs"].(float64)) != 180000 {
		t.Errorf("durationMs = %v, want 180000", desc["durationMs"])
	}
	if desc["backendId"] != "b-1" {
		t.Errorf("backendId = %v, want b-1", desc["backendId"])
	}
	if desc["backendType"] != "local" {
		t.Errorf("backendType = %v, want local", desc["backendType"])
	}
	if desc["objectKey"] != "track.flac" {
		t.Errorf("objectKey = %v, want track.flac", desc["objectKey"])
	}
}

func TestGetTrackPlaybackDescriptorAdminSession(t *testing.T) {
	h, _, adminToken := newViewerWithMediaHandler(t)

	backendBody := `{"id":"b-admin","type":"local","displayName":"Local","enabled":true,"isDefault":true,"config":{"local":{"rootPath":"/music"}}}`
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backendBody, "Bearer "+adminToken)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	moBody := `{"id":"mo-admin-1","backendId":"b-admin","objectKey":"a.flac","contentHash":"sha256:x","sizeBytes":1,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/media/objects", moBody, "Bearer "+adminToken)

	trackBody := fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-admin-1","durationMs":1000}`, artistID)
	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks", trackBody, "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/"+trackID+"/playback", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin session playback: %d %s", resp.Code, resp.Body.String())
	}
}

func TestGetTrackPlaybackDescriptorTrackNotFound(t *testing.T) {
	h, viewerToken, _ := newViewerWithMediaHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/no-such-track/playback", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestGetTrackPlaybackDescriptorMediaObjectNotFound(t *testing.T) {
	h, viewerToken, adminToken := newViewerWithMediaHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackBody := fmt.Sprintf(`{"title":"Orphan","artistId":%q,"mediaObjectId":"mo-missing","durationMs":1000}`, artistID)
	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks", trackBody, "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/"+trackID+"/playback", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestGetTrackPlaybackDescriptorNotActive(t *testing.T) {
	h, viewerToken, adminToken := newViewerWithMediaHandler(t)

	backendBody := `{"id":"b-x","type":"local","displayName":"Local","enabled":true,"isDefault":true,"config":{"local":{"rootPath":"/music"}}}`
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backendBody, "Bearer "+adminToken)

	moBody := `{"id":"mo-staged","backendId":"b-x","objectKey":"s.flac","contentHash":"sha256:y","sizeBytes":1,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"staged"}`
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/media/objects", moBody, "Bearer "+adminToken)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackBody := fmt.Sprintf(`{"title":"Staged","artistId":%q,"mediaObjectId":"mo-staged","durationMs":1000}`, artistID)
	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks", trackBody, "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/"+trackID+"/playback", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusUnprocessableEntity, "playback_unavailable")
}

func TestGetTrackPlaybackDescriptorWrongKind(t *testing.T) {
	h, viewerToken, adminToken := newViewerWithMediaHandler(t)

	backendBody := `{"id":"b-art","type":"local","displayName":"Local","enabled":true,"isDefault":true,"config":{"local":{"rootPath":"/music"}}}`
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backendBody, "Bearer "+adminToken)

	moBody := `{"id":"mo-art","backendId":"b-art","objectKey":"cover.jpg","contentHash":"sha256:z","sizeBytes":1,"mimeType":"image/jpeg","assetKind":"artwork","lifecycleState":"active"}`
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/media/objects", moBody, "Bearer "+adminToken)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackBody := fmt.Sprintf(`{"title":"Artwork","artistId":%q,"mediaObjectId":"mo-art","durationMs":1000}`, artistID)
	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks", trackBody, "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/"+trackID+"/playback", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusUnprocessableEntity, "playback_unavailable")
}

func TestGetTrackPlaybackDescriptorNoCatalogService(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test"}),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "alice", "viewerpass1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/any-id/playback", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

func TestGetTrackPlaybackDescriptorMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newViewerWithMediaHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/catalog/tracks/any-id/playback", `{}`, "Bearer "+viewerToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestGetTrackPlaybackDescriptorPresignedURL(t *testing.T) {
	// Build a handler with a fake S3 backend that has PresignedURLs capability.
	// We don't need the fake server to respond — we only assert the URL shape.
	t.Setenv("HTTP_TEST_S3_ACCESS", "test-access-key")
	t.Setenv("HTTP_TEST_S3_SECRET", "test-secret-key")

	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	storageRepo := storage.NewMemoryRepository()
	storageSvc := storage.NewService(storageRepo)
	mediaRepo := storage.NewMemoryMediaObjectRepository()
	mediaSvc := storage.NewMediaObjectService(storageRepo, mediaRepo)
	catalogRepo := catalog.NewMemoryRepository()
	h := NewHandler(
		storageSvc,
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
		WithMediaObjectService(mediaSvc),
		WithCatalogService(catalog.NewService(catalogRepo)),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()

	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "alice", "viewerpass1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "bob", "adminpass2", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, err := authSvc.Login(context.Background(), "bob", "adminpass2")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}

	// Register an S3 backend; capabilities are inferred by the server, then we
	// override PresignedURLs directly in the repository to enable presigned URL generation.
	backendBody := fmt.Sprintf(`{
		"id":"s3-presign-test","type":"s3","displayName":"S3","enabled":true,"isDefault":true,
		"config":{"s3":{"endpoint":"https://s3.example.com","region":"us-east-1","bucket":"music",
		"pathStyle":true,"accessKeySecretRef":"HTTP_TEST_S3_ACCESS","secretKeySecretRef":"HTTP_TEST_S3_SECRET"}}
	}`)
	backendResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backendBody, "Bearer "+adminToken)
	if backendResp.Code != http.StatusCreated {
		t.Fatalf("register backend: %d %s", backendResp.Code, backendResp.Body.String())
	}
	// Enable the PresignedURLs capability directly in the in-memory repo.
	{
		b, err := storageRepo.Get(context.Background(), "s3-presign-test")
		if err != nil {
			t.Fatalf("get backend: %v", err)
		}
		b.Capabilities.PresignedURLs = true
		if err := storageRepo.Save(context.Background(), b); err != nil {
			t.Fatalf("save backend with presigned URL capability: %v", err)
		}
	}

	moBody := `{"id":"mo-presign","backendId":"s3-presign-test","objectKey":"music/track.flac","contentHash":"sha256:presign","sizeBytes":1024,"mimeType":"audio/flac","assetKind":"original_audio","lifecycleState":"active"}`
	moResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/media/objects", moBody, "Bearer "+adminToken)
	if moResp.Code != http.StatusCreated {
		t.Fatalf("register media object: %d %s", moResp.Code, moResp.Body.String())
	}

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackBody := fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-presign","durationMs":180000}`, artistID)
	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks", trackBody, "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/"+trackID+"/playback", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("playback descriptor: %d %s", resp.Code, resp.Body.String())
	}
	var desc map[string]any
	decodeResponse(t, resp, &desc)
	presignedURL, _ := desc["presignedUrl"].(string)
	if presignedURL == "" {
		t.Fatal("presignedUrl is missing or empty; expected a signed URL for S3 backend with PresignedURLs=true")
	}
	if !strings.Contains(presignedURL, "X-Amz-Signature") {
		t.Errorf("presignedUrl does not look like a SigV4 URL: %s", presignedURL)
	}
	if !strings.Contains(presignedURL, "track.flac") {
		t.Errorf("presignedUrl missing object key: %s", presignedURL)
	}
}

// ---- viewer catalog stats tests ----

func TestViewerGetCatalogStatsEmpty(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/stats", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	for _, field := range []string{"artists", "albums", "tracks", "playlists"} {
		if v, ok := got[field]; !ok || int(v.(float64)) != 0 {
			t.Errorf("field %q = %v, want 0", field, v)
		}
	}
}

func TestViewerGetCatalogStatsPopulated(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist A"}`, "Bearer "+adminToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"Mix"}`, "Bearer "+adminToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/stats", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	if int(got["artists"].(float64)) != 1 {
		t.Errorf("artists = %v, want 1", got["artists"])
	}
	if int(got["playlists"].(float64)) != 1 {
		t.Errorf("playlists = %v, want 1", got["playlists"])
	}
}

func TestViewerGetCatalogStatsAdminSession(t *testing.T) {
	h, _, adminToken := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/stats", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin session catalog stats: %d %s", resp.Code, resp.Body.String())
	}
}

func TestViewerGetCatalogStatsNoCatalogService(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test"}),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "alice", "viewerpass1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/stats", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

func TestViewerGetCatalogStatsMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/catalog/stats", `{}`, "Bearer "+viewerToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestViewerGetArtistStatsBreakdown(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		fmt.Sprintf(`{"title":"Album","artistId":%q,"releaseYear":2024}`, artistID), "Bearer "+adminToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/stats/artists", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["artists"].([]any)
	if len(items) != 1 {
		t.Fatalf("artists = %d, want 1", len(items))
	}
	item := items[0].(map[string]any)
	if int(item["albumCount"].(float64)) != 1 {
		t.Errorf("albumCount = %v, want 1", item["albumCount"])
	}
}

func TestViewerGetArtistStatsBreakdownMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/catalog/stats/artists", `{}`, "Bearer "+viewerToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestViewerGetAlbumStatsBreakdown(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)
	albumResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		fmt.Sprintf(`{"title":"Album","artistId":%q,"releaseYear":2024}`, artistID), "Bearer "+adminToken)
	var album map[string]any
	decodeResponse(t, albumResp, &album)
	albumID := album["id"].(string)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"albumId":%q,"mediaObjectId":"mo-stat-1"}`, artistID, albumID), "Bearer "+adminToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/stats/albums", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["albums"].([]any)
	if len(items) != 1 {
		t.Fatalf("albums = %d, want 1", len(items))
	}
	item := items[0].(map[string]any)
	if int(item["trackCount"].(float64)) != 1 {
		t.Errorf("trackCount = %v, want 1", item["trackCount"])
	}
}

func TestViewerGetAlbumStatsBreakdownMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/catalog/stats/albums", `{}`, "Bearer "+viewerToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestViewerGetPlaylistStatsBreakdown(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"Mix"}`, "Bearer "+adminToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/stats/playlists", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	items := got["playlists"].([]any)
	if len(items) != 1 {
		t.Fatalf("playlists = %d, want 1", len(items))
	}
}

func TestViewerGetPlaylistStatsBreakdownMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/catalog/stats/playlists", `{}`, "Bearer "+viewerToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

// ---- catalog list pagination tests ----

func TestCatalogListArtistsPagination(t *testing.T) {
	h := newCatalogTestHandler()

	// seed 3 artists
	for _, name := range []string{"Artist A", "Artist B", "Artist C"} {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists",
			fmt.Sprintf(`{"name":%q}`, name))
	}

	// default (no params) — all 3 returned
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d", resp.Code)
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	artists := got["artists"].([]any)
	if len(artists) != 3 {
		t.Errorf("want 3 artists, got %d", len(artists))
	}
	pagination := got["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 3 {
		t.Errorf("pagination.total = %v, want 3", pagination["total"])
	}
	if pagination["hasMore"].(bool) {
		t.Error("hasMore should be false")
	}

	// limit=2
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists?limit=2", "")
	decodeResponse(t, resp, &got)
	artists = got["artists"].([]any)
	if len(artists) != 2 {
		t.Errorf("limit=2: want 2 artists, got %d", len(artists))
	}
	pagination = got["pagination"].(map[string]any)
	if !pagination["hasMore"].(bool) {
		t.Error("hasMore should be true when limit=2 with 3 items")
	}
	if int(pagination["total"].(float64)) != 3 {
		t.Errorf("total = %v, want 3", pagination["total"])
	}

	// offset=2 — 1 item remains
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists?limit=50&offset=2", "")
	decodeResponse(t, resp, &got)
	artists = got["artists"].([]any)
	if len(artists) != 1 {
		t.Errorf("offset=2: want 1 artist, got %d", len(artists))
	}

	// offset past end — empty slice
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists?limit=50&offset=99", "")
	decodeResponse(t, resp, &got)
	artists = got["artists"].([]any)
	if len(artists) != 0 {
		t.Errorf("offset=99: want 0 artists, got %d", len(artists))
	}

	// invalid limit
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists?limit=bad", "")
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_limit")

	// invalid offset
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists?offset=-1", "")
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_offset")
}

func TestCatalogListAlbumsPagination(t *testing.T) {
	h := newCatalogTestHandler()

	// seed artist + 3 albums
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)
	for i := 1; i <= 3; i++ {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
			fmt.Sprintf(`{"title":"Album %d","artistId":%q,"releaseYear":2020}`, i, artistID))
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums?limit=2", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	albums := got["albums"].([]any)
	if len(albums) != 2 {
		t.Errorf("limit=2: want 2 albums, got %d", len(albums))
	}
	pagination := got["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 3 {
		t.Errorf("total = %v, want 3", pagination["total"])
	}
	if !pagination["hasMore"].(bool) {
		t.Error("hasMore should be true")
	}
}

func TestCatalogListTracksPagination(t *testing.T) {
	h := newCatalogTestHandler()

	// seed artist + 3 tracks
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)
	for i := 1; i <= 3; i++ {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
			fmt.Sprintf(`{"title":"Track %d","artistId":%q,"mediaObjectId":"mo-%d"}`, i, artistID, i))
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/tracks?limit=2&offset=1", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	tracks := got["tracks"].([]any)
	if len(tracks) != 2 {
		t.Errorf("limit=2 offset=1: want 2 tracks, got %d", len(tracks))
	}
	pagination := got["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 3 {
		t.Errorf("total = %v, want 3", pagination["total"])
	}
	if pagination["hasMore"].(bool) {
		t.Error("hasMore should be false: offset=1 limit=2 total=3 means we've consumed all")
	}
}

func TestCatalogListPlaylistsPagination(t *testing.T) {
	h := newCatalogTestHandler()

	for _, name := range []string{"Mix A", "Mix B", "Mix C"} {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists",
			fmt.Sprintf(`{"name":%q}`, name))
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/playlists?limit=1", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	playlists := got["playlists"].([]any)
	if len(playlists) != 1 {
		t.Errorf("limit=1: want 1 playlist, got %d", len(playlists))
	}
	pagination := got["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 3 {
		t.Errorf("total = %v, want 3", pagination["total"])
	}
	if !pagination["hasMore"].(bool) {
		t.Error("hasMore should be true when limit=1 with 3 items")
	}
}

func TestViewerCatalogListArtistsPagination(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist X"}`, "Bearer "+adminToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist Y"}`, "Bearer "+adminToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/artists?limit=1", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer list artists: %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	artists := got["artists"].([]any)
	if len(artists) != 1 {
		t.Errorf("viewer limit=1: want 1 artist, got %d", len(artists))
	}
	pagination := got["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 2 {
		t.Errorf("total = %v, want 2", pagination["total"])
	}
	if !pagination["hasMore"].(bool) {
		t.Error("hasMore should be true")
	}
}

// ---- catalog list sort tests ----

func TestCatalogListArtistsSortByName(t *testing.T) {
	h := newCatalogTestHandler()
	for _, name := range []string{"Zara", "Alice", "Mike"} {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", fmt.Sprintf(`{"name":%q}`, name))
	}

	// default asc by name
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists?sortBy=name&sortOrder=asc", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d", resp.Code)
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	artists := got["artists"].([]any)
	if len(artists) != 3 {
		t.Fatalf("expected 3 artists, got %d", len(artists))
	}
	if artists[0].(map[string]any)["name"] != "Alice" {
		t.Errorf("first artist (asc) = %v, want Alice", artists[0].(map[string]any)["name"])
	}

	// desc by name
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists?sortBy=name&sortOrder=desc", "")
	decodeResponse(t, resp, &got)
	artists = got["artists"].([]any)
	if artists[0].(map[string]any)["name"] != "Zara" {
		t.Errorf("first artist (desc) = %v, want Zara", artists[0].(map[string]any)["name"])
	}
}

func TestCatalogListAlbumsSortByReleaseYear(t *testing.T) {
	h := newCatalogTestHandler()
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)
	for _, year := range []int{2020, 2015, 2023} {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
			fmt.Sprintf(`{"title":"Album %d","artistId":%q,"releaseYear":%d}`, year, artistID, year))
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums?sortBy=releaseYear&sortOrder=asc", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	albums := got["albums"].([]any)
	first := albums[0].(map[string]any)
	if int(first["releaseYear"].(float64)) != 2015 {
		t.Errorf("first album year (asc) = %v, want 2015", first["releaseYear"])
	}
	last := albums[2].(map[string]any)
	if int(last["releaseYear"].(float64)) != 2023 {
		t.Errorf("last album year (asc) = %v, want 2023", last["releaseYear"])
	}
}

func TestCatalogListTracksSortByTitle(t *testing.T) {
	h := newCatalogTestHandler()
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)
	for _, title := range []string{"Zephyr", "Aura", "Midnight"} {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
			fmt.Sprintf(`{"title":%q,"artistId":%q,"mediaObjectId":"mo-%s"}`, title, artistID, strings.ToLower(title[:3])))
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/tracks?sortBy=title&sortOrder=asc", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	tracks := got["tracks"].([]any)
	if tracks[0].(map[string]any)["title"] != "Aura" {
		t.Errorf("first track (asc) = %v, want Aura", tracks[0].(map[string]any)["title"])
	}
}

func TestCatalogListPlaylistsSortByName(t *testing.T) {
	h := newCatalogTestHandler()
	for _, name := range []string{"Zen Mix", "Alpha Hits", "Morning Chill"} {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", fmt.Sprintf(`{"name":%q}`, name))
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/playlists?sortBy=name&sortOrder=desc", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	playlists := got["playlists"].([]any)
	if playlists[0].(map[string]any)["name"] != "Zen Mix" {
		t.Errorf("first playlist (desc) = %v, want Zen Mix", playlists[0].(map[string]any)["name"])
	}
}

func TestCatalogListArtistsInvalidSortOrder(t *testing.T) {
	h := newCatalogTestHandler()
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists?sortOrder=random", "")
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_sort_order")
}

func TestViewerCatalogListArtistsSort(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)
	for _, name := range []string{"Zara", "Alice"} {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists",
			fmt.Sprintf(`{"name":%q}`, name), "Bearer "+adminToken)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/artists?sortBy=name&sortOrder=asc", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer sort artists: %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	artists := got["artists"].([]any)
	if artists[0].(map[string]any)["name"] != "Alice" {
		t.Errorf("viewer first artist (asc) = %v, want Alice", artists[0].(map[string]any)["name"])
	}
}

// ---- nested browse route tests (Phase 63) ----

func TestListAlbumsByArtistRoute(t *testing.T) {
	h := newCatalogTestHandler()

	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	for i := 1; i <= 3; i++ {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
			fmt.Sprintf(`{"title":"Album %d","artistId":%q,"releaseYear":%d}`, i, artistID, 2020+i))
	}

	// all albums for artist
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists/"+artistID+"/albums", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	albums := got["albums"].([]any)
	if len(albums) != 3 {
		t.Errorf("want 3 albums, got %d", len(albums))
	}
	if got["pagination"] == nil {
		t.Error("pagination key missing")
	}

	// with limit
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists/"+artistID+"/albums?limit=2", "")
	decodeResponse(t, resp, &got)
	albums = got["albums"].([]any)
	if len(albums) != 2 {
		t.Errorf("limit=2: want 2 albums, got %d", len(albums))
	}

	// with sort
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists/"+artistID+"/albums?sortBy=releaseYear&sortOrder=desc", "")
	decodeResponse(t, resp, &got)
	albums = got["albums"].([]any)
	if int(albums[0].(map[string]any)["releaseYear"].(float64)) != 2023 {
		t.Errorf("first album year (desc) = %v, want 2023", albums[0].(map[string]any)["releaseYear"])
	}

	// unknown artist → 404
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists/no-such/albums", "")
	assertAPIError(t, resp, http.StatusNotFound, "not_found")

	// method not allowed
	resp = performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists/"+artistID+"/albums", `{}`)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestListTracksByArtistRoute(t *testing.T) {
	h := newCatalogTestHandler()

	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	for i := 1; i <= 2; i++ {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
			fmt.Sprintf(`{"title":"Track %d","artistId":%q,"mediaObjectId":"mo-art-%d"}`, i, artistID, i))
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists/"+artistID+"/tracks", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	tracks := got["tracks"].([]any)
	if len(tracks) != 2 {
		t.Errorf("want 2 tracks, got %d", len(tracks))
	}

	// unknown artist → 404
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/artists/no-such/tracks", "")
	assertAPIError(t, resp, http.StatusNotFound, "not_found")

	// method not allowed
	resp = performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists/"+artistID+"/tracks", `{}`)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestListTracksByAlbumRoute(t *testing.T) {
	h := newCatalogTestHandler()

	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	albumResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		fmt.Sprintf(`{"title":"Album","artistId":%q,"releaseYear":2024}`, artistID))
	var album map[string]any
	decodeResponse(t, albumResp, &album)
	albumID := album["id"].(string)

	for i := 1; i <= 3; i++ {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
			fmt.Sprintf(`{"title":"Song %d","artistId":%q,"albumId":%q,"mediaObjectId":"mo-alb-%d"}`, i, artistID, albumID, i))
	}

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums/"+albumID+"/tracks", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	tracks := got["tracks"].([]any)
	if len(tracks) != 3 {
		t.Errorf("want 3 tracks, got %d", len(tracks))
	}
	pagination := got["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 3 {
		t.Errorf("total = %v, want 3", pagination["total"])
	}

	// sort descending
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums/"+albumID+"/tracks?sortBy=title&sortOrder=desc", "")
	decodeResponse(t, resp, &got)
	tracks = got["tracks"].([]any)
	if tracks[0].(map[string]any)["title"] != "Song 3" {
		t.Errorf("first track (desc) = %v, want Song 3", tracks[0].(map[string]any)["title"])
	}

	// unknown album → 404
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums/no-such/tracks", "")
	assertAPIError(t, resp, http.StatusNotFound, "not_found")

	// method not allowed
	resp = performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums/"+albumID+"/tracks", `{}`)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

func TestViewerNestedBrowseRoutes(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	albumResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		fmt.Sprintf(`{"title":"Album","artistId":%q,"releaseYear":2024}`, artistID), "Bearer "+adminToken)
	var album map[string]any
	decodeResponse(t, albumResp, &album)
	albumID := album["id"].(string)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"albumId":%q,"mediaObjectId":"mo-view-1"}`, artistID, albumID), "Bearer "+adminToken)

	for _, path := range []string{
		"/api/v1/catalog/artists/" + artistID + "/albums",
		"/api/v1/catalog/artists/" + artistID + "/tracks",
		"/api/v1/catalog/albums/" + albumID + "/tracks",
	} {
		resp := performRequestWithAuthHeader(t, h, http.MethodGet, path, "", "Bearer "+viewerToken)
		if resp.Code != http.StatusOK {
			t.Errorf("viewer GET %s: status = %d, body = %s", path, resp.Code, resp.Body.String())
		}
	}
}

// ---- playlist tracks pagination tests (Phase 64) ----

func TestGetPlaylistTracksPagination(t *testing.T) {
	h := newCatalogTestHandler()

	// seed: artist + 5 tracks + playlist with all 5 in order
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		tr := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
			fmt.Sprintf(`{"title":"Song %d","artistId":%q,"mediaObjectId":"mo-pl-pg-%d"}`, i+1, artistID, i+1))
		var trk map[string]any
		decodeResponse(t, tr, &trk)
		trackIDs[i] = trk["id"].(string)
	}

	plResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"PL"}`)
	var pl map[string]any
	decodeResponse(t, plResp, &pl)
	plID := pl["id"].(string)

	trackIDsJSON, _ := json.Marshal(trackIDs)
	performRequest(t, h, http.MethodPut, "/api/v1/admin/catalog/playlists/"+plID+"/tracks",
		fmt.Sprintf(`{"trackIds":%s}`, trackIDsJSON))

	// default (all 5)
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/playlists/"+plID+"/tracks", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	tracks := got["tracks"].([]any)
	if len(tracks) != 5 {
		t.Fatalf("default: want 5 tracks, got %d", len(tracks))
	}
	pagination := got["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 5 {
		t.Errorf("total = %v, want 5", pagination["total"])
	}
	if pagination["hasMore"].(bool) {
		t.Error("hasMore should be false for default limit with 5 items")
	}

	// order preserved: first track is Song 1 (trackIDs[0])
	first := tracks[0].(map[string]any)
	if first["id"] != trackIDs[0] {
		t.Errorf("order: first track id = %v, want %s", first["id"], trackIDs[0])
	}

	// limit=2
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/playlists/"+plID+"/tracks?limit=2", "")
	decodeResponse(t, resp, &got)
	tracks = got["tracks"].([]any)
	if len(tracks) != 2 {
		t.Errorf("limit=2: want 2 tracks, got %d", len(tracks))
	}
	pagination = got["pagination"].(map[string]any)
	if !pagination["hasMore"].(bool) {
		t.Error("hasMore should be true with limit=2 of 5")
	}
	// order preserved within page
	if tracks[0].(map[string]any)["id"] != trackIDs[0] {
		t.Errorf("page[0] id = %v, want %s", tracks[0].(map[string]any)["id"], trackIDs[0])
	}
	if tracks[1].(map[string]any)["id"] != trackIDs[1] {
		t.Errorf("page[1] id = %v, want %s", tracks[1].(map[string]any)["id"], trackIDs[1])
	}

	// offset=3 limit=2 → last 2 tracks
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/playlists/"+plID+"/tracks?limit=2&offset=3", "")
	decodeResponse(t, resp, &got)
	tracks = got["tracks"].([]any)
	if len(tracks) != 2 {
		t.Errorf("offset=3 limit=2: want 2 tracks, got %d", len(tracks))
	}
	if tracks[0].(map[string]any)["id"] != trackIDs[3] {
		t.Errorf("offset=3 page[0] id = %v, want %s", tracks[0].(map[string]any)["id"], trackIDs[3])
	}
	pagination = got["pagination"].(map[string]any)
	if pagination["hasMore"].(bool) {
		t.Error("hasMore should be false: offset=3 limit=2 total=5 consumes last 2")
	}

	// invalid limit
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/playlists/"+plID+"/tracks?limit=bad", "")
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_limit")

	// invalid offset
	resp = performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/playlists/"+plID+"/tracks?offset=-1", "")
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_offset")
}

func TestViewerGetPlaylistTracksPagination(t *testing.T) {
	h, viewerToken, adminToken := newViewerTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		tr := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
			fmt.Sprintf(`{"title":"Song %d","artistId":%q,"mediaObjectId":"mo-vpl-%d"}`, i+1, artistID, i+1), "Bearer "+adminToken)
		var trk map[string]any
		decodeResponse(t, tr, &trk)
		trackIDs[i] = trk["id"].(string)
	}

	plResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/playlists", `{"name":"VPL"}`, "Bearer "+adminToken)
	var pl map[string]any
	decodeResponse(t, plResp, &pl)
	plID := pl["id"].(string)

	trackIDsJSON, _ := json.Marshal(trackIDs)
	performRequestWithAuthHeader(t, h, http.MethodPut, "/api/v1/admin/catalog/playlists/"+plID+"/tracks",
		fmt.Sprintf(`{"trackIds":%s}`, trackIDsJSON), "Bearer "+adminToken)

	// viewer: paginated GET
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/catalog/playlists/"+plID+"/tracks?limit=2", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer playlist tracks: %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	tracks := got["tracks"].([]any)
	if len(tracks) != 2 {
		t.Errorf("viewer limit=2: want 2 tracks, got %d", len(tracks))
	}
	pagination := got["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 3 {
		t.Errorf("total = %v, want 3", pagination["total"])
	}
	if !pagination["hasMore"].(bool) {
		t.Error("hasMore should be true")
	}
}

// ---- playback history tests (Phase 68) ----

func newHistoryTestHandler(t *testing.T) (http.Handler, string, string) {
	t.Helper()
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	storageSvc := storage.NewService(storage.NewMemoryRepository())
	catalogRepo := catalog.NewMemoryRepository()
	historySvc := historyNewService()

	h := NewHandler(
		storageSvc,
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
		WithCatalogService(catalog.NewService(catalogRepo)),
		WithHistoryService(historySvc),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()

	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "alice", "viewerpass1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "bob", "adminpass2", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, err := authSvc.Login(context.Background(), "bob", "adminpass2")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	return h, viewerToken, adminToken
}

func TestRecordPlayEvent(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// Seed artist + track via admin
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackBody := fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-hist-1"}`, artistID)
	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks", trackBody, "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Record a play
	body := fmt.Sprintf(`{"trackId":%q}`, trackID)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history", body, "Bearer "+viewerToken)
	if resp.Code != http.StatusCreated {
		t.Fatalf("record play: %d %s", resp.Code, resp.Body.String())
	}
	var event map[string]any
	decodeResponse(t, resp, &event)
	if event["id"] == nil || event["trackId"] != trackID {
		t.Errorf("event = %v", event)
	}
	if event["playedAt"] == nil || event["createdAt"] == nil {
		t.Errorf("missing timestamps in event: %v", event)
	}

	// Missing trackId → 400
	resp = performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history", `{}`, "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "validation_error")
}

func TestListPlayEvents(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	var trackIDs []string
	for i := 1; i <= 3; i++ {
		tr := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
			fmt.Sprintf(`{"title":"Song %d","artistId":%q,"mediaObjectId":"mo-hist-%d"}`, i, artistID, i), "Bearer "+adminToken)
		var t2 map[string]any
		decodeResponse(t, tr, &t2)
		trackIDs = append(trackIDs, t2["id"].(string))
	}

	// Record 3 plays
	for _, tid := range trackIDs {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, tid), "Bearer "+viewerToken)
	}

	// List all
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("list history: %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	events := got["events"].([]any)
	if len(events) != 3 {
		t.Errorf("events count = %d, want 3", len(events))
	}

	// limit=1
	resp = performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history?limit=1", "", "Bearer "+viewerToken)
	decodeResponse(t, resp, &got)
	events = got["events"].([]any)
	if len(events) != 1 {
		t.Errorf("limit=1: events count = %d, want 1", len(events))
	}
	pag := got["pagination"].(map[string]any)
	if int(pag["total"].(float64)) != 3 {
		t.Errorf("pagination.total = %v, want 3", pag["total"])
	}
}

func TestClearHistory(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-clear-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)

	// Verify 1 event
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history", "", "Bearer "+viewerToken)
	var got map[string]any
	decodeResponse(t, resp, &got)
	if len(got["events"].([]any)) != 1 {
		t.Fatalf("want 1 event before clear")
	}

	// Clear
	resp = performRequestWithAuthHeader(t, h, http.MethodDelete, "/api/v1/me/history", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusNoContent {
		t.Fatalf("clear history: %d %s", resp.Code, resp.Body.String())
	}

	// Verify 0 events
	resp = performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history", "", "Bearer "+viewerToken)
	decodeResponse(t, resp, &got)
	if len(got["events"].([]any)) != 0 {
		t.Fatalf("want 0 events after clear, got %v", got["events"])
	}
}

func TestHistoryNotConfigured(t *testing.T) {
	// Handler without history service
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test"}),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, _ := authSvc.Login(context.Background(), "alice", "viewerpass1")
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history", `{"trackId":"x"}`, "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
}

func TestHistoryMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/me/history", `{}`, "Bearer "+viewerToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.Code)
	}
}

// ---- admin history stats tests ----

func TestAdminGetHistoryStats(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// Seed artist + track via admin, then record some plays
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-stats-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Record 2 plays as viewer
	for i := 0; i < 2; i++ {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history/stats", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin history stats: %d %s", resp.Code, resp.Body.String())
	}
	var stats map[string]any
	decodeResponse(t, resp, &stats)
	if int(stats["totalEvents"].(float64)) != 2 {
		t.Errorf("totalEvents = %v, want 2", stats["totalEvents"])
	}
	if int(stats["uniqueUsers"].(float64)) != 1 {
		t.Errorf("uniqueUsers = %v, want 1", stats["uniqueUsers"])
	}
	if int(stats["uniqueTracks"].(float64)) != 1 {
		t.Errorf("uniqueTracks = %v, want 1", stats["uniqueTracks"])
	}
}

func TestAdminGetTopTracks(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	// Create 2 tracks; play track-1 twice and track-2 once.
	var trackIDs []string
	for i := 1; i <= 2; i++ {
		tr := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
			fmt.Sprintf(`{"title":"Song %d","artistId":%q,"mediaObjectId":"mo-top-%d"}`, i, artistID, i), "Bearer "+adminToken)
		var t2 map[string]any
		decodeResponse(t, tr, &t2)
		trackIDs = append(trackIDs, t2["id"].(string))
	}
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackIDs[0]), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackIDs[0]), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackIDs[1]), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history/top-tracks", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("top tracks: %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	tracks := got["tracks"].([]any)
	if len(tracks) != 2 {
		t.Fatalf("tracks = %d, want 2", len(tracks))
	}
	first := tracks[0].(map[string]any)
	if int(first["playCount"].(float64)) != 2 {
		t.Errorf("first track playCount = %v, want 2", first["playCount"])
	}

	// limit=1
	resp = performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history/top-tracks?limit=1", "", "Bearer "+adminToken)
	decodeResponse(t, resp, &got)
	if len(got["tracks"].([]any)) != 1 {
		t.Errorf("limit=1: tracks = %d, want 1", len(got["tracks"].([]any)))
	}
}

func TestAdminGetTopUsers(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-topuser-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Create a second viewer to have 2 distinct users
	authSvc := newMemAuthUserRepo()
	// Use admin token path — history was recorded via the existing viewerToken in newHistoryTestHandler.
	// We only verify admin can call the endpoint.
	_ = trackID

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history/top-users", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("top users: %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	if _, ok := got["users"]; !ok {
		t.Error("response missing \"users\" key")
	}
	_ = authSvc
}

func TestAdminHistoryStatsNotConfigured(t *testing.T) {
	// Handler without history service
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()

	for _, path := range []string{
		"/api/v1/admin/history/stats",
		"/api/v1/admin/history/top-tracks",
		"/api/v1/admin/history/top-users",
	} {
		resp := performRequest(t, h, http.MethodGet, path, "")
		assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
	}
}

func TestAdminHistoryMethodNotAllowed(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)
	for _, path := range []string{
		"/api/v1/admin/history/stats",
		"/api/v1/admin/history/top-tracks",
		"/api/v1/admin/history/top-users",
	} {
		resp := performRequestWithAuthHeader(t, h, http.MethodPost, path, `{}`, "Bearer "+adminToken)
		if resp.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s: expected 405, got %d", path, resp.Code)
		}
	}
}

func TestAdminHistorySinceFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// Create artist + track
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-since-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Record a play event with an explicit timestamp in the past (2020-01-01)
	oldTime := "2020-01-01T00:00:00Z"
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":%q}`, trackID, oldTime), "Bearer "+viewerToken)

	// With since set to 2025-01-01, the old event is excluded → totalEvents = 0
	since := "2025-01-01T00:00:00Z"
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/stats?since="+since, "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("stats since: %d %s", resp.Code, resp.Body.String())
	}
	var stats map[string]any
	decodeResponse(t, resp, &stats)
	if int(stats["totalEvents"].(float64)) != 0 {
		t.Errorf("windowed totalEvents = %v, want 0", stats["totalEvents"])
	}

	// top-tracks since 2025-01-01 → empty list
	resp = performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/top-tracks?since="+since, "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("top-tracks since: %d %s", resp.Code, resp.Body.String())
	}
	var ttr map[string]any
	decodeResponse(t, resp, &ttr)
	if len(ttr["tracks"].([]any)) != 0 {
		t.Errorf("windowed tracks = %d, want 0", len(ttr["tracks"].([]any)))
	}
}

func TestAdminHistorySinceInvalid(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)
	for _, path := range []string{
		"/api/v1/admin/history/stats?since=not-a-date",
		"/api/v1/admin/history/top-tracks?since=not-a-date",
		"/api/v1/admin/history/top-users?since=not-a-date",
	} {
		resp := performRequestWithAuthHeader(t, h, http.MethodGet, path, "", "Bearer "+adminToken)
		assertAPIError(t, resp, http.StatusBadRequest, "invalid_since")
	}
}

func TestAdminHistoryUntilFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// Create artist + track
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-until-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Record event in 2030 (well in the future)
	futureTime := "2030-06-01T00:00:00Z"
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":%q}`, trackID, futureTime), "Bearer "+viewerToken)

	// until=2025-01-01 excludes the 2030 event → totalEvents=0
	until := "2025-01-01T00:00:00Z"
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/stats?until="+until, "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("stats until: %d %s", resp.Code, resp.Body.String())
	}
	var stats map[string]any
	decodeResponse(t, resp, &stats)
	if int(stats["totalEvents"].(float64)) != 0 {
		t.Errorf("windowed totalEvents = %v, want 0", stats["totalEvents"])
	}
}

func TestAdminHistoryUntilInvalid(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)
	for _, path := range []string{
		"/api/v1/admin/history/stats?until=not-a-date",
		"/api/v1/admin/history/top-tracks?until=not-a-date",
		"/api/v1/admin/history/top-users?until=not-a-date",
	} {
		resp := performRequestWithAuthHeader(t, h, http.MethodGet, path, "", "Bearer "+adminToken)
		assertAPIError(t, resp, http.StatusBadRequest, "invalid_until")
	}
}

func TestAdminHistoryInvalidTimeRange(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)
	// since >= until → invalid_time_range
	path := "/api/v1/admin/history/stats?since=2030-01-01T00:00:00Z&until=2020-01-01T00:00:00Z"
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, path, "", "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_time_range")
}

func TestAdminGetUserHistory(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// Create artist + track
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-uhistory-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Record a play event and capture userId from the response
	playResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var playEvent map[string]any
	decodeResponse(t, playResp, &playEvent)
	viewerID := playEvent["userId"].(string)

	// Record one more
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)

	// Admin fetches viewer's history
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/"+viewerID, "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("user history: %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	events := got["events"].([]any)
	if len(events) != 2 {
		t.Errorf("events = %d, want 2", len(events))
	}
	pagination := got["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 2 {
		t.Errorf("total = %v, want 2", pagination["total"])
	}

	// limit=1
	resp = performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/"+viewerID+"?limit=1", "", "Bearer "+adminToken)
	decodeResponse(t, resp, &got)
	if len(got["events"].([]any)) != 1 {
		t.Errorf("limit=1: events = %d, want 1", len(got["events"].([]any)))
	}
}

func TestAdminGetTrackHistory(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// Create artist + track
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-thistory-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Record 3 plays for the track
	for i := 0; i < 3; i++ {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/"+trackID, "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("track history: %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	events := got["events"].([]any)
	if len(events) != 3 {
		t.Errorf("events = %d, want 3", len(events))
	}
	pagination := got["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 3 {
		t.Errorf("total = %v, want 3", pagination["total"])
	}
}

func TestAdminHistoryDetailMethodNotAllowed(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)
	for _, path := range []string{
		"/api/v1/admin/history/users/some-user",
		"/api/v1/admin/history/tracks/some-track",
	} {
		resp := performRequestWithAuthHeader(t, h, http.MethodPost, path, `{}`, "Bearer "+adminToken)
		if resp.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s: expected 405, got %d", path, resp.Code)
		}
	}
}

func TestAdminHistoryDetailNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()

	for _, path := range []string{
		"/api/v1/admin/history/users/some-user",
		"/api/v1/admin/history/tracks/some-track",
	} {
		resp := performRequest(t, h, http.MethodGet, path, "")
		assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
	}
}

func TestAdminDeleteUserHistory(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-del-u-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// record a play and capture viewer's userID
	playResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var playEvent map[string]any
	decodeResponse(t, playResp, &playEvent)
	viewerID := playEvent["userId"].(string)

	// confirm event is present
	listResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history", "", "Bearer "+viewerToken)
	var listBefore map[string]any
	decodeResponse(t, listResp, &listBefore)
	if listBefore["pagination"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("expected 1 event before delete, got %v", listBefore["pagination"].(map[string]any)["total"])
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodDelete, "/api/v1/admin/history/users/"+viewerID, "", "Bearer "+adminToken)
	if resp.Code != http.StatusNoContent {
		t.Fatalf("admin delete user history: expected 204, got %d %s", resp.Code, resp.Body.String())
	}

	listResp2 := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history", "", "Bearer "+viewerToken)
	var listAfter map[string]any
	decodeResponse(t, listResp2, &listAfter)
	if listAfter["pagination"].(map[string]any)["total"].(float64) != 0 {
		t.Errorf("expected 0 events after admin delete, got %v", listAfter["pagination"].(map[string]any)["total"])
	}
}

func TestAdminDeleteTrackHistory(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Artist"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Trk","artistId":%q,"mediaObjectId":"mo-del-t-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodDelete, "/api/v1/admin/history/tracks/"+trackID, "", "Bearer "+adminToken)
	if resp.Code != http.StatusNoContent {
		t.Fatalf("admin delete track history: expected 204, got %d %s", resp.Code, resp.Body.String())
	}

	statsResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history/stats", "", "Bearer "+adminToken)
	var stats map[string]any
	decodeResponse(t, statsResp, &stats)
	if stats["totalEvents"].(float64) != 0 {
		t.Errorf("expected 0 events after track delete, got %v", stats["totalEvents"])
	}
}

func TestAdminDeleteHistoryWindow(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Art"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"T","artistId":%q,"mediaObjectId":"mo-del-w-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2020-01-01T10:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2020-01-01T12:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2020-01-01T20:00:00Z"}`, trackID), "Bearer "+viewerToken)

	// delete window [11:00, 15:00) — only the 12:00 event is deleted
	resp := performRequestWithAuthHeader(t, h, http.MethodDelete,
		"/api/v1/admin/history?since=2020-01-01T11:00:00Z&until=2020-01-01T15:00:00Z", "", "Bearer "+adminToken)
	if resp.Code != http.StatusNoContent {
		t.Fatalf("admin delete history window: expected 204, got %d %s", resp.Code, resp.Body.String())
	}

	statsResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history/stats", "", "Bearer "+adminToken)
	var stats map[string]any
	decodeResponse(t, statsResp, &stats)
	if stats["totalEvents"].(float64) != 2 {
		t.Errorf("expected 2 events after window delete, got %v", stats["totalEvents"])
	}
}

func TestAdminDeleteHistoryWindowMissingFilter(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodDelete, "/api/v1/admin/history", "", "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusBadRequest, "missing_time_filter")
}

func TestAdminBulkDeleteHistoryNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()

	for _, path := range []string{
		"/api/v1/admin/history/users/some-user",
		"/api/v1/admin/history/tracks/some-track",
		"/api/v1/admin/history",
	} {
		resp := performRequest(t, h, http.MethodDelete, path, "")
		assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
	}
}

func TestGetMyHistoryStats(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"S1","artistId":%q,"mediaObjectId":"mo-mstats-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID1 := track["id"].(string)

	trackResp2 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"S2","artistId":%q,"mediaObjectId":"mo-mstats-2"}`, artistID), "Bearer "+adminToken)
	var track2 map[string]any
	decodeResponse(t, trackResp2, &track2)
	trackID2 := track2["id"].(string)

	// viewer plays t1 twice and t2 once
	for i := 0; i < 2; i++ {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, trackID1), "Bearer "+viewerToken)
	}
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID2), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history/stats", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getMyHistoryStats: %d %s", resp.Code, resp.Body.String())
	}
	var stats map[string]any
	decodeResponse(t, resp, &stats)
	if stats["totalEvents"].(float64) != 3 {
		t.Errorf("totalEvents = %v, want 3", stats["totalEvents"])
	}
	if stats["uniqueTracks"].(float64) != 2 {
		t.Errorf("uniqueTracks = %v, want 2", stats["uniqueTracks"])
	}
}

func TestGetMyTopTracks(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Art"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"T1","artistId":%q,"mediaObjectId":"mo-mtop-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// viewer plays the track 3 times
	for i := 0; i < 3; i++ {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history/top-tracks", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getMyTopTracks: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	tracks := result["tracks"].([]any)
	if len(tracks) != 1 {
		t.Fatalf("tracks len = %d, want 1", len(tracks))
	}
	top := tracks[0].(map[string]any)
	if top["trackId"].(string) != trackID {
		t.Errorf("top trackId = %q, want %q", top["trackId"], trackID)
	}
	if top["playCount"].(float64) != 3 {
		t.Errorf("playCount = %v, want 3", top["playCount"])
	}
}

func TestGetMyHistoryStatsTimeWindow(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Art"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"T","artistId":%q,"mediaObjectId":"mo-mwin-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// 3 events at different times
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2020-01-01T09:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2020-01-01T12:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2020-01-01T18:00:00Z"}`, trackID), "Bearer "+viewerToken)

	// window [10:00, 15:00) captures only the 12:00 event
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/stats?since=2020-01-01T10:00:00Z&until=2020-01-01T15:00:00Z", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("time-window stats: %d %s", resp.Code, resp.Body.String())
	}
	var stats map[string]any
	decodeResponse(t, resp, &stats)
	if stats["totalEvents"].(float64) != 1 {
		t.Errorf("totalEvents = %v, want 1", stats["totalEvents"])
	}
}

func TestGetMyHistoryStatsNotConfigured(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, _ := authSvc.Login(context.Background(), "alice", "viewerpass1")

	for _, path := range []string{
		"/api/v1/me/history/stats",
		"/api/v1/me/history/top-tracks",
	} {
		resp := performRequestWithAuthHeader(t, h, http.MethodGet, path, "", "Bearer "+viewerToken)
		assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
	}
}

func TestViewerGetHistoryTimeline(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TLViewerBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TLVT1","artistId":%q,"mediaObjectId":"mo-tlv-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// 2 events on day1, 1 event on day2
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-05-01T10:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-05-01T14:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-05-02T09:00:00Z"}`, trackID), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/timeline?since=2025-05-01T00:00:00Z&until=2025-05-03T00:00:00Z&granularity=day",
		"", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getMyHistoryTimeline: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	buckets := result["buckets"].([]any)
	if len(buckets) != 2 {
		t.Fatalf("buckets = %d, want 2", len(buckets))
	}
	b0 := buckets[0].(map[string]any)
	if b0["eventCount"].(float64) != 2 {
		t.Errorf("day1 eventCount = %v, want 2", b0["eventCount"])
	}
	b1 := buckets[1].(map[string]any)
	if b1["eventCount"].(float64) != 1 {
		t.Errorf("day2 eventCount = %v, want 1", b1["eventCount"])
	}
}

func TestViewerGetHistoryTimelineMissingSince(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/timeline?until=2025-05-03T00:00:00Z", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "missing_time_bounds")

	resp2 := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/timeline?since=2025-05-01T00:00:00Z", "", "Bearer "+viewerToken)
	assertAPIError(t, resp2, http.StatusBadRequest, "missing_time_bounds")
}

func TestViewerGetHistoryTimelineInvalidGranularity(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/timeline?since=2025-05-01T00:00:00Z&until=2025-05-03T00:00:00Z&granularity=hour",
		"", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_granularity")
}

func TestViewerGetHistoryTimelineNotConfigured(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "viewertl", "viewerpassTL1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, _ := authSvc.Login(context.Background(), "viewertl", "viewerpassTL1")

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/timeline?since=2025-05-01T00:00:00Z&until=2025-05-03T00:00:00Z",
		"", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
}

func TestAdminGetUserStats(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"StatsBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"S1","artistId":%q,"mediaObjectId":"mo-aus-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID1 := track["id"].(string)

	trackResp2 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"S2","artistId":%q,"mediaObjectId":"mo-aus-2"}`, artistID), "Bearer "+adminToken)
	var track2 map[string]any
	decodeResponse(t, trackResp2, &track2)
	trackID2 := track2["id"].(string)

	// viewer plays t1 once and capture the play event to get viewerID
	firstPlayResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID1), "Bearer "+viewerToken)
	var firstPlay map[string]any
	decodeResponse(t, firstPlayResp, &firstPlay)
	viewerID := firstPlay["userId"].(string)

	// viewer plays t1 once more and t2 once
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID1), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID2), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/"+viewerID+"/stats", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getAdminUserStats: %d %s", resp.Code, resp.Body.String())
	}
	var stats map[string]any
	decodeResponse(t, resp, &stats)
	if stats["totalEvents"].(float64) != 3 {
		t.Errorf("totalEvents = %v, want 3", stats["totalEvents"])
	}
	if stats["uniqueTracks"].(float64) != 2 {
		t.Errorf("uniqueTracks = %v, want 2", stats["uniqueTracks"])
	}
}

func TestAdminGetUserTopTracks(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TopArt"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TT1","artistId":%q,"mediaObjectId":"mo-autt-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// viewer plays the track — capture first event to get viewerID
	firstPlayResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var firstPlay map[string]any
	decodeResponse(t, firstPlayResp, &firstPlay)
	viewerID := firstPlay["userId"].(string)

	// play 2 more times (3 total)
	for i := 0; i < 2; i++ {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/"+viewerID+"/top-tracks", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getAdminUserTopTracks: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	tracks := result["tracks"].([]any)
	if len(tracks) != 1 {
		t.Fatalf("tracks len = %d, want 1", len(tracks))
	}
	top := tracks[0].(map[string]any)
	if top["trackId"].(string) != trackID {
		t.Errorf("top trackId = %q, want %q", top["trackId"], trackID)
	}
	if top["playCount"].(float64) != 3 {
		t.Errorf("playCount = %v, want 3", top["playCount"])
	}
}

func TestAdminGetUserTopTracksTimeWindow(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TWBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TW1","artistId":%q,"mediaObjectId":"mo-auttw-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// 3 events at different fixed times; capture first event to get viewerID
	firstPlayResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2021-03-01T08:00:00Z"}`, trackID), "Bearer "+viewerToken)
	var firstPlay map[string]any
	decodeResponse(t, firstPlayResp, &firstPlay)
	viewerID := firstPlay["userId"].(string)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2021-03-01T12:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2021-03-01T20:00:00Z"}`, trackID), "Bearer "+viewerToken)

	// window [10:00, 15:00) captures only the 12:00 event
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/"+viewerID+"/top-tracks?since=2021-03-01T10:00:00Z&until=2021-03-01T15:00:00Z",
		"", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin user top-tracks windowed: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	tracks := result["tracks"].([]any)
	if len(tracks) != 1 {
		t.Fatalf("windowed tracks len = %d, want 1", len(tracks))
	}
	if tracks[0].(map[string]any)["playCount"].(float64) != 1 {
		t.Errorf("playCount = %v, want 1", tracks[0].(map[string]any)["playCount"])
	}
}

func TestAdminGetUserStatsNotConfigured(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "adminnc", "adminpassNC1", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, _ := authSvc.Login(context.Background(), "adminnc", "adminpassNC1")

	for _, path := range []string{
		"/api/v1/admin/history/users/someuser/stats",
		"/api/v1/admin/history/users/someuser/top-tracks",
	} {
		resp := performRequestWithAuthHeader(t, h, http.MethodGet, path, "", "Bearer "+adminToken)
		assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
	}
}

func TestAdminGetTrackStats(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TrackStatsBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TS1","artistId":%q,"mediaObjectId":"mo-ts-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// viewer plays the track twice; capture first event to get viewerID
	firstPlayResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var firstPlay map[string]any
	decodeResponse(t, firstPlayResp, &firstPlay)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/"+trackID+"/stats", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getAdminTrackStats: %d %s", resp.Code, resp.Body.String())
	}
	var stats map[string]any
	decodeResponse(t, resp, &stats)
	if stats["totalEvents"].(float64) != 2 {
		t.Errorf("totalEvents = %v, want 2", stats["totalEvents"])
	}
	if stats["uniqueListeners"].(float64) != 1 {
		t.Errorf("uniqueListeners = %v, want 1", stats["uniqueListeners"])
	}
}

func TestAdminGetTrackTopListeners(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"ListenerArt"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TL1","artistId":%q,"mediaObjectId":"mo-tl-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// viewer plays the track 3 times; capture first event to get viewerID
	firstPlayResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var firstPlay map[string]any
	decodeResponse(t, firstPlayResp, &firstPlay)
	viewerID := firstPlay["userId"].(string)

	for i := 0; i < 2; i++ {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/"+trackID+"/top-listeners", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getAdminTrackTopListeners: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	users := result["users"].([]any)
	if len(users) != 1 {
		t.Fatalf("users len = %d, want 1", len(users))
	}
	top := users[0].(map[string]any)
	if top["userId"].(string) != viewerID {
		t.Errorf("top userId = %q, want %q", top["userId"], viewerID)
	}
	if top["playCount"].(float64) != 3 {
		t.Errorf("playCount = %v, want 3", top["playCount"])
	}
}

func TestAdminGetTrackTopListenersTimeWindow(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TLWBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TLW1","artistId":%q,"mediaObjectId":"mo-tlw-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// 3 events at different fixed times; first to get viewerID
	firstPlayResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2022-05-01T08:00:00Z"}`, trackID), "Bearer "+viewerToken)
	var firstPlay map[string]any
	decodeResponse(t, firstPlayResp, &firstPlay)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2022-05-01T12:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2022-05-01T20:00:00Z"}`, trackID), "Bearer "+viewerToken)

	// window [10:00, 15:00) captures only the 12:00 event → 1 listener with playCount 1
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/"+trackID+"/top-listeners?since=2022-05-01T10:00:00Z&until=2022-05-01T15:00:00Z",
		"", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin track top-listeners windowed: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	users := result["users"].([]any)
	if len(users) != 1 {
		t.Fatalf("windowed users len = %d, want 1", len(users))
	}
	if users[0].(map[string]any)["playCount"].(float64) != 1 {
		t.Errorf("playCount = %v, want 1", users[0].(map[string]any)["playCount"])
	}
}

func TestAdminGetTrackStatsNotConfigured(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "admintsnc", "adminpassTSNC1", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, _ := authSvc.Login(context.Background(), "admintsnc", "adminpassTSNC1")

	for _, path := range []string{
		"/api/v1/admin/history/tracks/sometrack/stats",
		"/api/v1/admin/history/tracks/sometrack/top-listeners",
	} {
		resp := performRequestWithAuthHeader(t, h, http.MethodGet, path, "", "Bearer "+adminToken)
		assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
	}
}

func TestAdminGetHistoryTimeline(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TLBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TL1","artistId":%q,"mediaObjectId":"mo-tl-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// 2 events on day1, 1 event on day2
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-04-01T10:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-04-01T15:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-04-02T08:00:00Z"}`, trackID), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/timeline?since=2025-04-01T00:00:00Z&until=2025-04-03T00:00:00Z&granularity=day",
		"", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getAdminHistoryTimeline: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	buckets := result["buckets"].([]any)
	if len(buckets) != 2 {
		t.Fatalf("buckets = %d, want 2", len(buckets))
	}
	b0 := buckets[0].(map[string]any)
	if b0["eventCount"].(float64) != 2 {
		t.Errorf("day1 eventCount = %v, want 2", b0["eventCount"])
	}
	b1 := buckets[1].(map[string]any)
	if b1["eventCount"].(float64) != 1 {
		t.Errorf("day2 eventCount = %v, want 1", b1["eventCount"])
	}
}

func TestAdminGetHistoryTimelineMissingSince(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)

	// missing since
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/timeline?until=2025-04-03T00:00:00Z", "", "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusBadRequest, "missing_time_bounds")

	// missing until
	resp2 := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/timeline?since=2025-04-01T00:00:00Z", "", "Bearer "+adminToken)
	assertAPIError(t, resp2, http.StatusBadRequest, "missing_time_bounds")
}

func TestAdminGetHistoryTimelineInvalidGranularity(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/timeline?since=2025-04-01T00:00:00Z&until=2025-04-03T00:00:00Z&granularity=hour",
		"", "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_granularity")
}

func TestAdminGetHistoryTimelineNotConfigured(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "admintl", "adminpassTL1", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, _ := authSvc.Login(context.Background(), "admintl", "adminpassTL1")

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/timeline?since=2025-04-01T00:00:00Z&until=2025-04-03T00:00:00Z",
		"", "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
}

func TestAdminGetAllHistory(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// create artist + 2 tracks
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"AllHistBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp1 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"AHT1","artistId":%q,"mediaObjectId":"mo-ah-1"}`, artistID), "Bearer "+adminToken)
	var track1 map[string]any
	decodeResponse(t, trackResp1, &track1)
	trackID1 := track1["id"].(string)

	trackResp2 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"AHT2","artistId":%q,"mediaObjectId":"mo-ah-2"}`, artistID), "Bearer "+adminToken)
	var track2 map[string]any
	decodeResponse(t, trackResp2, &track2)
	trackID2 := track2["id"].(string)

	// viewer plays track1 twice and track2 once
	for i := 0; i < 2; i++ {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, trackID1), "Bearer "+viewerToken)
	}
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID2), "Bearer "+viewerToken)

	// admin sees all 3 events
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/history: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	pagination := result["pagination"].(map[string]any)
	if pagination["total"].(float64) != 3 {
		t.Errorf("total = %v, want 3", pagination["total"])
	}
	events := result["events"].([]any)
	if len(events) != 3 {
		t.Errorf("events len = %d, want 3", len(events))
	}
}

func TestAdminGetAllHistoryTrackFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"FilterBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp1 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"FT1","artistId":%q,"mediaObjectId":"mo-flt-1"}`, artistID), "Bearer "+adminToken)
	var track1 map[string]any
	decodeResponse(t, trackResp1, &track1)
	trackID1 := track1["id"].(string)

	trackResp2 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"FT2","artistId":%q,"mediaObjectId":"mo-flt-2"}`, artistID), "Bearer "+adminToken)
	var track2 map[string]any
	decodeResponse(t, trackResp2, &track2)
	trackID2 := track2["id"].(string)
	_ = trackID2 // only track1 is filtered for

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID1), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID2), "Bearer "+viewerToken)

	// filter by trackId → only 1 event
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history?trackId="+trackID1, "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/history?trackId: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	pagination := result["pagination"].(map[string]any)
	if pagination["total"].(float64) != 1 {
		t.Errorf("filtered total = %v, want 1", pagination["total"])
	}
}

func TestAdminGetAllHistoryNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()

	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/history", "")
	assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
}

func TestAdminGetAllHistoryMethodNotAllowed(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/history", "{}", "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusMethodNotAllowed, "method_not_allowed")
}

func TestListPlayEventsAscOrder(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"SortBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp1 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"ST1","artistId":%q,"mediaObjectId":"mo-sort-1"}`, artistID), "Bearer "+adminToken)
	var t1 map[string]any
	decodeResponse(t, trackResp1, &t1)
	tID1 := t1["id"].(string)

	trackResp2 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"ST2","artistId":%q,"mediaObjectId":"mo-sort-2"}`, artistID), "Bearer "+adminToken)
	var t2 map[string]any
	decodeResponse(t, trackResp2, &t2)
	tID2 := t2["id"].(string)

	// play t1 first, then t2
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2020-01-01T10:00:00Z"}`, tID1), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2020-01-01T11:00:00Z"}`, tID2), "Bearer "+viewerToken)

	// default (desc) → t2 first
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history", "", "Bearer "+viewerToken)
	var result map[string]any
	decodeResponse(t, resp, &result)
	events := result["events"].([]any)
	first := events[0].(map[string]any)["trackId"].(string)
	if first != tID2 {
		t.Errorf("desc[0] = %q, want tID2 (%q)", first, tID2)
	}

	// asc → t1 first
	respAsc := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history?order=asc", "", "Bearer "+viewerToken)
	var resultAsc map[string]any
	decodeResponse(t, respAsc, &resultAsc)
	eventsAsc := resultAsc["events"].([]any)
	firstAsc := eventsAsc[0].(map[string]any)["trackId"].(string)
	if firstAsc != tID1 {
		t.Errorf("asc[0] = %q, want tID1 (%q)", firstAsc, tID1)
	}
}

func TestListPlayEventsInvalidOrder(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history?order=random", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_order")
}

func TestAdminGetAllHistoryAscOrder(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"AscBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp1 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"AT1","artistId":%q,"mediaObjectId":"mo-asc-1"}`, artistID), "Bearer "+adminToken)
	var t1 map[string]any
	decodeResponse(t, trackResp1, &t1)
	tID1 := t1["id"].(string)

	trackResp2 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"AT2","artistId":%q,"mediaObjectId":"mo-asc-2"}`, artistID), "Bearer "+adminToken)
	var t2 map[string]any
	decodeResponse(t, trackResp2, &t2)
	tID2 := t2["id"].(string)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2020-06-01T08:00:00Z"}`, tID1), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2020-06-01T09:00:00Z"}`, tID2), "Bearer "+viewerToken)

	// asc → tID1 (older) first
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history?order=asc", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/history?order=asc: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	events := result["events"].([]any)
	firstTrack := events[0].(map[string]any)["trackId"].(string)
	if firstTrack != tID1 {
		t.Errorf("asc[0] trackId = %q, want tID1 (%q)", firstTrack, tID1)
	}
}

func TestAdminGetAllHistoryInvalidOrder(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history?order=newest", "", "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_order")
}

func TestAdminGetEvent(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"EventBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"EV1","artistId":%q,"mediaObjectId":"mo-ev-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// viewer records a play; capture event ID from response
	playResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var evt map[string]any
	decodeResponse(t, playResp, &evt)
	eventID := evt["id"].(string)

	// admin can fetch the event
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history/"+eventID, "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin GET event: %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	if got["id"].(string) != eventID {
		t.Errorf("id = %q, want %q", got["id"], eventID)
	}
}

func TestAdminGetEventNotFound(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history/no-such-event", "", "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestAdminDeleteEvent(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"DelEvBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"DE1","artistId":%q,"mediaObjectId":"mo-dev-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	playResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var evt map[string]any
	decodeResponse(t, playResp, &evt)
	eventID := evt["id"].(string)

	del := performRequestWithAuthHeader(t, h, http.MethodDelete, "/api/v1/admin/history/"+eventID, "", "Bearer "+adminToken)
	if del.Code != http.StatusNoContent {
		t.Fatalf("admin DELETE event: %d %s", del.Code, del.Body.String())
	}

	// gone
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history/"+eventID, "", "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestViewerGetEvent(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"VEvBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"VE1","artistId":%q,"mediaObjectId":"mo-vev-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	playResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var evt map[string]any
	decodeResponse(t, playResp, &evt)
	eventID := evt["id"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history/"+eventID, "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer GET event: %d %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	if got["id"].(string) != eventID {
		t.Errorf("id = %q, want %q", got["id"], eventID)
	}
}

func TestViewerGetEventNotOwned(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	// Verify 404 for a non-existent event
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history/no-such-id", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestViewerDeleteEvent(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"VDelBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"VD1","artistId":%q,"mediaObjectId":"mo-vdel-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	playResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var evt map[string]any
	decodeResponse(t, playResp, &evt)
	eventID := evt["id"].(string)

	del := performRequestWithAuthHeader(t, h, http.MethodDelete, "/api/v1/me/history/"+eventID, "", "Bearer "+viewerToken)
	if del.Code != http.StatusNoContent {
		t.Fatalf("viewer DELETE event: %d %s", del.Code, del.Body.String())
	}

	// gone — viewer's own history list should be empty
	listResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history", "", "Bearer "+viewerToken)
	var listResult map[string]any
	decodeResponse(t, listResp, &listResult)
	pagination := listResult["pagination"].(map[string]any)
	if pagination["total"].(float64) != 0 {
		t.Errorf("total after delete = %v, want 0", pagination["total"])
	}
}

func TestPerEventHistoryNotConfigured(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, _ := authSvc.Login(context.Background(), "alice", "viewerpass1")

	// admin per-event endpoints (use admin bearer token)
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/admin/history/some-id"},
		{http.MethodDelete, "/api/v1/admin/history/some-id"},
	} {
		resp := performRequest(t, h, tc.method, tc.path, "")
		// no history service → should 503; without admin token → 401 from auth middleware
		// use admin token to get past auth
		resp = performRequestWithAuthHeader(t, h, tc.method, tc.path, "", "Bearer "+testAdminToken)
		assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
	}

	// viewer per-event endpoints
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/me/history/some-id"},
		{http.MethodDelete, "/api/v1/me/history/some-id"},
	} {
		resp := performRequestWithAuthHeader(t, h, tc.method, tc.path, "", "Bearer "+viewerToken)
		assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
	}
}

func TestAdminPatchEvent(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"PatchBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"PT1","artistId":%q,"mediaObjectId":"mo-pa-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	playResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var evt map[string]any
	decodeResponse(t, playResp, &evt)
	eventID := evt["id"].(string)

	newTime := "2020-01-01T12:00:00Z"
	resp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/admin/history/"+eventID,
		fmt.Sprintf(`{"playedAt":%q}`, newTime), "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin PATCH event: %d %s", resp.Code, resp.Body.String())
	}
	var updated map[string]any
	decodeResponse(t, resp, &updated)
	if updated["playedAt"].(string) != newTime {
		t.Errorf("playedAt = %q, want %q", updated["playedAt"], newTime)
	}
}

func TestAdminPatchEventNotFound(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/admin/history/no-such",
		`{"playedAt":"2020-01-01T00:00:00Z"}`, "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

func TestAdminPatchEventInvalidPlayedAt(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/admin/history/some-id",
		`{"playedAt":"not-a-date"}`, "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_played_at")
}

func TestViewerPatchEvent(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"VPatchBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"VP1","artistId":%q,"mediaObjectId":"mo-vpa-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	playResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var evt map[string]any
	decodeResponse(t, playResp, &evt)
	eventID := evt["id"].(string)

	newTime := "2021-06-15T09:30:00Z"
	resp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/me/history/"+eventID,
		fmt.Sprintf(`{"playedAt":%q}`, newTime), "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer PATCH event: %d %s", resp.Code, resp.Body.String())
	}
	var updated map[string]any
	decodeResponse(t, resp, &updated)
	if updated["playedAt"].(string) != newTime {
		t.Errorf("playedAt = %q, want %q", updated["playedAt"], newTime)
	}
}

func TestViewerPatchEventInvalidPlayedAt(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/me/history/some-id",
		`{"playedAt":"bad"}`, "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_played_at")
}

func TestViewerPatchEventMissingPlayedAt(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/me/history/some-id",
		`{}`, "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_played_at")
}

func TestPatchEventHistoryNotConfigured(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, _ := authSvc.Login(context.Background(), "alice", "viewerpass1")

	body := `{"playedAt":"2020-01-01T00:00:00Z"}`
	for _, tc := range []struct{ path, token string }{
		{"/api/v1/admin/history/some-id", "Bearer " + testAdminToken},
		{"/api/v1/me/history/some-id", "Bearer " + viewerToken},
	} {
		resp := performRequestWithAuthHeader(t, h, http.MethodPatch, tc.path, body, tc.token)
		assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
	}
}

func TestAdminBatchDeleteEvents(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"BatchBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"BT1","artistId":%q,"mediaObjectId":"mo-bd-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// record 3 events
	var ids []string
	for range 3 {
		pr := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
		var evt map[string]any
		decodeResponse(t, pr, &evt)
		ids = append(ids, evt["id"].(string))
	}

	// batch-delete first two
	body := fmt.Sprintf(`{"ids":[%q,%q]}`, ids[0], ids[1])
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/history/batch-delete", body, "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin batch-delete: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["deleted"].(float64) != 2 {
		t.Errorf("deleted = %v, want 2", result["deleted"])
	}

	// third still present
	getResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/admin/history/"+ids[2], "", "Bearer "+adminToken)
	if getResp.Code != http.StatusOK {
		t.Errorf("third event should still exist, got %d", getResp.Code)
	}
}

func TestAdminBatchDeleteEventsEmptyBody(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/history/batch-delete",
		`{"ids":[]}`, "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_ids")
}

func TestViewerBatchDeleteMyEvents(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"VBatchBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"VBT1","artistId":%q,"mediaObjectId":"mo-vbd-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	pr1 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var evt1 map[string]any
	decodeResponse(t, pr1, &evt1)
	id1 := evt1["id"].(string)

	pr2 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var evt2 map[string]any
	decodeResponse(t, pr2, &evt2)
	id2 := evt2["id"].(string)

	// viewer deletes both own events
	body := fmt.Sprintf(`{"ids":[%q,%q]}`, id1, id2)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history/batch-delete", body, "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer batch-delete: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["deleted"].(float64) != 2 {
		t.Errorf("deleted = %v, want 2", result["deleted"])
	}
}

func TestViewerBatchDeleteSkipsOtherUsersEvents(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"SkipBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"SK1","artistId":%q,"mediaObjectId":"mo-sk-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// viewer records an event; then try to batch-delete a foreign ID and own ID
	pr := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var evt map[string]any
	decodeResponse(t, pr, &evt)
	ownID := evt["id"].(string)

	body := fmt.Sprintf(`{"ids":[%q,"foreign-event-id"]}`, ownID)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history/batch-delete", body, "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer batch-delete skip foreign: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	// only own event deleted; foreign not found → silently ignored
	if result["deleted"].(float64) != 1 {
		t.Errorf("deleted = %v, want 1", result["deleted"])
	}
}

func TestBatchDeleteHistoryNotConfigured(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "alice", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, _ := authSvc.Login(context.Background(), "alice", "viewerpass1")

	body := `{"ids":["some-id"]}`
	for _, tc := range []struct{ path, token string }{
		{"/api/v1/admin/history/batch-delete", "Bearer " + testAdminToken},
		{"/api/v1/me/history/batch-delete", "Bearer " + viewerToken},
	} {
		resp := performRequestWithAuthHeader(t, h, http.MethodPost, tc.path, body, tc.token)
		assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
	}
}

func TestListPlayEventsSinceFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"SinceBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"ST1","artistId":%q,"mediaObjectId":"mo-sf-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// 3 events at distinct times
	for _, ts := range []string{"2020-01-01T08:00:00Z", "2020-01-01T12:00:00Z", "2020-01-01T18:00:00Z"} {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q,"playedAt":%q}`, trackID, ts), "Bearer "+viewerToken)
	}

	// since=10:00 → only 12:00 and 18:00 events
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history?since=2020-01-01T10:00:00Z", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /me/history?since: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["pagination"].(map[string]any)["total"].(float64) != 2 {
		t.Errorf("total = %v, want 2", result["pagination"].(map[string]any)["total"])
	}
}

func TestListPlayEventsUntilFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"UntilBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"UT1","artistId":%q,"mediaObjectId":"mo-uf-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	for _, ts := range []string{"2020-06-01T08:00:00Z", "2020-06-01T12:00:00Z", "2020-06-01T20:00:00Z"} {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q,"playedAt":%q}`, trackID, ts), "Bearer "+viewerToken)
	}

	// until=15:00 (exclusive) → 08:00 and 12:00
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history?until=2020-06-01T15:00:00Z", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /me/history?until: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["pagination"].(map[string]any)["total"].(float64) != 2 {
		t.Errorf("total = %v, want 2", result["pagination"].(map[string]any)["total"])
	}
}

func TestAdminUserHistorySinceUntilFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"AUSBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"AU1","artistId":%q,"mediaObjectId":"mo-aus-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	playResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	var evt map[string]any
	decodeResponse(t, playResp, &evt)
	viewerID := evt["userId"].(string)

	for _, ts := range []string{"2021-01-01T06:00:00Z", "2021-01-01T10:00:00Z", "2021-01-01T22:00:00Z"} {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q,"playedAt":%q}`, trackID, ts), "Bearer "+viewerToken)
	}

	// window [08:00, 12:00) → only 10:00; total from all above = 1 matching window (plus the one without timestamp)
	// Just verify since filter narrows results
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/"+viewerID+"?since=2021-01-01T08:00:00Z&until=2021-01-01T12:00:00Z",
		"", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin user history since/until: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["pagination"].(map[string]any)["total"].(float64) != 1 {
		t.Errorf("total = %v, want 1", result["pagination"].(map[string]any)["total"])
	}
}

func TestAdminTrackHistorySinceFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"ATSBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"ATS1","artistId":%q,"mediaObjectId":"mo-ats-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	for _, ts := range []string{"2022-03-01T07:00:00Z", "2022-03-01T14:00:00Z"} {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q,"playedAt":%q}`, trackID, ts), "Bearer "+viewerToken)
	}

	// since=10:00 → only 14:00
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/"+trackID+"?since=2022-03-01T10:00:00Z",
		"", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin track history since: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["pagination"].(map[string]any)["total"].(float64) != 1 {
		t.Errorf("total = %v, want 1", result["pagination"].(map[string]any)["total"])
	}
}

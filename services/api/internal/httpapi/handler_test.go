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

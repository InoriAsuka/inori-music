package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"inori-music/services/api/internal/storage"
)

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

// ---- media object PATCH tests (Phase 140) ----

func TestPatchMediaObjectMetadata(t *testing.T) {
	h := newTestHandler()

	// Register a media object first
	body := `{"id":"mo-patch-1","backendId":"local-patch","objectKey":"audio/test.mp3","contentHash":"sha256:abc123def456abc123def456abc123def456abc123def456abc123def456abc1","sizeBytes":1024,"mimeType":"audio/mpeg","assetKind":"original_audio","lifecycleState":"active"}`

	// Need a backend first
	backendBody := `{"id":"local-patch","type":"local","displayName":"Patch Test","enabled":true,"config":{"local":{"rootPath":"/tmp/patch-mo"}}}`
	performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backendBody)

	registerResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/media/objects", body)
	if registerResp.Code != http.StatusCreated {
		t.Fatalf("register media object status = %d; body = %s", registerResp.Code, registerResp.Body.String())
	}

	// Patch mimeType
	patchResp := performRequest(t, h, http.MethodPatch, "/api/v1/admin/media/objects/mo-patch-1",
		`{"mimeType":"audio/flac"}`)
	if patchResp.Code != http.StatusOK {
		t.Fatalf("PATCH media object status = %d; body = %s", patchResp.Code, patchResp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, patchResp, &got)
	if got["mimeType"] != "audio/flac" {
		t.Errorf("mimeType = %v, want audio/flac", got["mimeType"])
	}

	// Patch assetKind
	patchResp2 := performRequest(t, h, http.MethodPatch, "/api/v1/admin/media/objects/mo-patch-1",
		`{"assetKind":"transcoded_audio"}`)
	if patchResp2.Code != http.StatusOK {
		t.Fatalf("PATCH assetKind status = %d; body = %s", patchResp2.Code, patchResp2.Body.String())
	}
	var got2 map[string]any
	decodeResponse(t, patchResp2, &got2)
	if got2["assetKind"] != "transcoded_audio" {
		t.Errorf("assetKind = %v, want transcoded_audio", got2["assetKind"])
	}

	// Invalid assetKind → 400
	patchResp3 := performRequest(t, h, http.MethodPatch, "/api/v1/admin/media/objects/mo-patch-1",
		`{"assetKind":"not_valid"}`)
	assertAPIError(t, patchResp3, http.StatusBadRequest, "invalid_media_object")

	// Invalid mimeType → 400
	patchResp4 := performRequest(t, h, http.MethodPatch, "/api/v1/admin/media/objects/mo-patch-1",
		`{"mimeType":"badformat"}`)
	assertAPIError(t, patchResp4, http.StatusBadRequest, "invalid_media_object")

	// Unknown ID → 404
	patchResp5 := performRequest(t, h, http.MethodPatch, "/api/v1/admin/media/objects/no-such",
		`{"mimeType":"audio/ogg"}`)
	assertAPIError(t, patchResp5, http.StatusNotFound, "not_found")
}

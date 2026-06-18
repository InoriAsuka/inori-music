package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStorageAdminOpenAPIContractCoversRoutes(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	expected := map[string][]string{
		"/healthz":                       {"get"},
		"/metrics":                       {"get"},
		"/readyz":                        {"get"},
		"/versionz":                      {"get"},
		"/api/v1/admin/storage/backends": {"get", "post"},
		"/api/v1/admin/storage/backends/validate":               {"post"},
		"/api/v1/admin/storage/backends/refresh":                {"post"},
		"/api/v1/admin/storage/backends/{id}/default":           {"post"},
		"/api/v1/admin/storage/backends/{id}/disable":           {"post"},
		"/api/v1/admin/storage/backends/{id}/probe":             {"post"},
		"/api/v1/admin/storage/backends/{id}/health":            {"get"},
		"/api/v1/admin/storage/backends/{id}/capacity":          {"get"},
		"/api/v1/admin/media/objects":                           {"get", "post"},
		"/api/v1/admin/media/objects/stats":                     {"get"},
		"/api/v1/admin/media/objects/duplicates":                {"get"},
		"/api/v1/admin/media/objects/lifecycle":                 {"post"},
		"/api/v1/admin/media/objects/{id}":                      {"get"},
		"/api/v1/admin/media/objects/{id}/timeline":             {"get"},
		"/api/v1/admin/media/objects/{id}/lifecycle":            {"post"},
		"/api/v1/admin/media/objects/verify":                    {"post"},
		"/api/v1/admin/media/objects/{id}/verify":               {"post"},
		"/api/v1/admin/catalog/artists":                         {"get", "post"},
		"/api/v1/admin/catalog/artists/{id}":                    {"get", "delete", "patch"},
		"/api/v1/admin/catalog/artists/{id}/albums":              {"get"},
		"/api/v1/admin/catalog/artists/{id}/tracks":              {"get"},
		"/api/v1/admin/catalog/albums":                          {"get", "post"},
		"/api/v1/admin/catalog/albums/{id}":                     {"get", "delete", "patch"},
		"/api/v1/admin/catalog/albums/{id}/tracks":               {"get"},
		"/api/v1/admin/catalog/tracks":                          {"get", "post"},
		"/api/v1/admin/catalog/tracks/{id}":                     {"get", "delete", "patch"},
		"/api/v1/admin/catalog/tracks/{id}/relink":              {"post"},
		"/api/v1/admin/catalog/import":                          {"post"},
		"/api/v1/admin/catalog/batch-import":                    {"post"},
		"/api/v1/admin/catalog/search":                          {"get"},
		"/api/v1/admin/catalog/playlists":                       {"get", "post"},
		"/api/v1/admin/catalog/playlists/{id}":                  {"get", "patch", "delete"},
		"/api/v1/admin/catalog/playlists/{id}/tracks":           {"get", "post", "put"},
		"/api/v1/admin/catalog/playlists/{id}/tracks/{trackId}": {"delete"},
		"/api/v1/catalog/playlists":                             {"get"},
		"/api/v1/catalog/playlists/{id}":                        {"get"},
		"/api/v1/catalog/playlists/{id}/tracks":                 {"get"},
		"/api/v1/catalog/artists":                               {"get"},
		"/api/v1/catalog/artists/{id}":                          {"get"},
		"/api/v1/catalog/artists/{id}/albums":                   {"get"},
		"/api/v1/catalog/artists/{id}/tracks":                   {"get"},
		"/api/v1/catalog/albums":                                {"get"},
		"/api/v1/catalog/albums/{id}":                           {"get"},
		"/api/v1/catalog/albums/{id}/tracks":                    {"get"},
		"/api/v1/catalog/tracks":                                {"get"},
		"/api/v1/catalog/tracks/{id}":                           {"get"},
		"/api/v1/catalog/tracks/{id}/playback":                  {"get"},
		"/api/v1/catalog/search":                                {"get"},
		"/api/v1/catalog/recently-added":                        {"get"},
		"/api/v1/catalog/recently-updated":                      {"get"},
		"/api/v1/catalog/stats":                                  {"get"},
		"/api/v1/catalog/stats/artists":                          {"get"},
		"/api/v1/catalog/stats/albums":                           {"get"},
		"/api/v1/catalog/stats/playlists":                        {"get"},
		"/api/v1/me/history":                                     {"get", "post", "delete"},
		"/api/v1/me/history/stats":                               {"get"},
		"/api/v1/me/history/top-tracks":                         {"get"},
		"/api/v1/me/history/{eventId}":                          {"get", "patch", "delete"},
		"/api/v1/me/history/batch-delete":                       {"post"},
		"/api/v1/admin/catalog/stats":                           {"get"},
		"/api/v1/admin/catalog/stats/artists":                   {"get"},
		"/api/v1/admin/catalog/stats/albums":                    {"get"},
		"/api/v1/admin/catalog/stats/playlists":                 {"get"},
		"/api/v1/admin/catalog/recently-added":                  {"get"},
		"/api/v1/admin/catalog/recently-updated":                {"get"},
		"/api/v1/admin/history/stats":                           {"get"},
		"/api/v1/admin/history/top-tracks":                      {"get"},
		"/api/v1/admin/history/top-users":                       {"get"},
		"/api/v1/admin/history/users/{userId}":                  {"get", "delete"},
		"/api/v1/admin/history/users/{userId}/stats":            {"get"},
		"/api/v1/admin/history/users/{userId}/top-tracks":       {"get"},
		"/api/v1/admin/history/tracks/{trackId}":                {"get", "delete"},
		"/api/v1/admin/history/tracks/{trackId}/stats":          {"get"},
		"/api/v1/admin/history/tracks/{trackId}/top-listeners":  {"get"},
		"/api/v1/admin/history/{eventId}":                       {"get", "patch", "delete"},
		"/api/v1/admin/history":                                 {"get", "delete"},
		"/api/v1/admin/history/batch-delete":                    {"post"},
	}

	for path, methods := range expected {
		pathItem, ok := paths[path].(map[string]any)
		if !ok {
			t.Fatalf("OpenAPI path %q is missing", path)
		}
		for _, method := range methods {
			if _, ok := pathItem[method].(map[string]any); !ok {
				t.Fatalf("OpenAPI operation %s %s is missing", method, path)
			}
		}
	}
}

func TestStorageAdminOpenAPIContractPathParameters(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	components := document["components"].(map[string]any)
	parameters := components["parameters"].(map[string]any)
	for _, want := range []string{"BackendId", "MediaObjectId", "UserId", "CatalogId"} {
		if _, ok := parameters[want].(map[string]any); !ok {
			t.Fatalf("OpenAPI %s path parameter is missing", want)
		}
	}

	// All paths containing {id} must have exactly one $ref path-level parameter.
	validRefs := map[string]bool{
		"#/components/parameters/BackendId":     true,
		"#/components/parameters/MediaObjectId": true,
		"#/components/parameters/UserId":        true,
		"#/components/parameters/CatalogId":     true,
	}
	for path, item := range paths {
		if !strings.Contains(path, "{id}") {
			continue
		}
		pathItem := item.(map[string]any)
		refs, ok := pathItem["parameters"].([]any)
		if !ok || len(refs) != 1 {
			t.Fatalf("path %s parameters = %#v, want exactly one parameter ref", path, pathItem["parameters"])
		}
		ref, _ := refs[0].(map[string]any)["$ref"].(string)
		if !validRefs[ref] {
			t.Fatalf("path %s parameter ref = %q, want a known components/parameters ref", path, ref)
		}
	}
}

func TestStorageAdminOpenAPIContractMediaObjectListQueryParameters(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	listOperation := operation(t, paths, "/api/v1/admin/media/objects", "get")
	parameters := listOperation["parameters"].([]any)
	seen := make(map[string]map[string]any)
	for _, parameter := range parameters {
		item := parameter.(map[string]any)
		name := item["name"].(string)
		seen[name] = item
	}
	for _, name := range []string{"backendId", "contentHash", "verificationStatus", "lifecycleState", "assetKind", "sortBy", "sortOrder", "limit", "offset"} {
		if _, ok := seen[name]; !ok {
			t.Fatalf("media object list query parameter %q is missing", name)
		}
	}
	sortBySchema := seen["sortBy"]["schema"].(map[string]any)
	if !containsString(sortBySchema["enum"].([]any), "size_bytes") || sortBySchema["default"] != "backend_object_key" {
		t.Fatalf("sortBy schema = %#v, want size_bytes enum and backend_object_key default", sortBySchema)
	}
	sortOrderSchema := seen["sortOrder"]["schema"].(map[string]any)
	if !containsString(sortOrderSchema["enum"].([]any), "desc") || sortOrderSchema["default"] != "asc" {
		t.Fatalf("sortOrder schema = %#v, want desc enum and asc default", sortOrderSchema)
	}
}

func TestStorageAdminOpenAPIContractMediaObjectDuplicateParameters(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	duplicateOperation := operation(t, paths, "/api/v1/admin/media/objects/duplicates", "get")
	parameters := duplicateOperation["parameters"].([]any)
	seen := make(map[string]map[string]any)
	for _, parameter := range parameters {
		item := parameter.(map[string]any)
		seen[item["name"].(string)] = item
	}
	if _, ok := seen["backendId"]; !ok {
		t.Fatal("duplicates backendId query parameter is missing")
	}
	minCopiesSchema := seen["minCopies"]["schema"].(map[string]any)
	if minCopiesSchema["minimum"] != float64(2) || minCopiesSchema["default"] != float64(2) {
		t.Fatalf("minCopies schema = %#v, want minimum/default 2", minCopiesSchema)
	}
}

func TestStorageAdminOpenAPIContractSecurity(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)

	health := operation(t, paths, "/healthz", "get")
	if security, ok := health["security"].([]any); !ok || len(security) != 0 {
		t.Fatalf("/healthz security = %#v, want public empty security", health["security"])
	}
	metrics := operation(t, paths, "/metrics", "get")
	if security, ok := metrics["security"].([]any); !ok || len(security) != 0 {
		t.Fatalf("/metrics security = %#v, want public empty security", metrics["security"])
	}
	ready := operation(t, paths, "/readyz", "get")
	if security, ok := ready["security"].([]any); !ok || len(security) != 0 {
		t.Fatalf("/readyz security = %#v, want public empty security", ready["security"])
	}
	version := operation(t, paths, "/versionz", "get")
	if security, ok := version["security"].([]any); !ok || len(security) != 0 {
		t.Fatalf("/versionz security = %#v, want public empty security", version["security"])
	}

	for path, item := range paths {
		if path == "/healthz" || path == "/metrics" || path == "/readyz" || path == "/versionz" {
			continue
		}
		// Login and logout are public endpoints (empty security, no bearerAuth required).
		if path == "/api/v1/auth/login" || path == "/api/v1/auth/logout" {
			continue
		}
		pathItem := item.(map[string]any)
		for method := range pathItem {
			if method == "parameters" {
				continue
			}
			op := operation(t, paths, path, method)
			security, ok := op["security"].([]any)
			if !ok || len(security) != 1 {
				t.Fatalf("%s %s security = %#v, want one bearer security requirement", method, path, op["security"])
			}
			requirement := security[0].(map[string]any)
			if _, ok := requirement["bearerAuth"]; !ok {
				t.Fatalf("%s %s security = %#v, want bearerAuth", method, path, requirement)
			}
		}
	}

	components := document["components"].(map[string]any)
	securitySchemes := components["securitySchemes"].(map[string]any)
	if _, ok := securitySchemes["bearerAuth"].(map[string]any); !ok {
		t.Fatal("OpenAPI bearerAuth security scheme is missing")
	}
}

func TestStorageAdminOpenAPIContractSchemasAndErrors(t *testing.T) {
	document := loadOpenAPIContract(t)
	components := document["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)
	for _, name := range []string{"StorageBackend", "StorageBackendRequest", "BackendConfig", "LocalConfig", "NFSConfig", "SMBConfig", "S3Config", "DistributedConfig", "CapabilitySet", "ProbeResult", "CapacityReport", "RefreshReport", "RefreshResult", "ServiceInfo", "ReadinessCheck", "ReadinessReport", "MediaObject", "MediaObjectRequest", "MediaObjectLifecycleRequest", "MediaObjectLifecycleChange", "MediaObjectTimeline", "MediaObjectTimelineEvent", "MediaObjectSelectionFilter", "MediaObjectBulkLifecycleRequest", "MediaObjectLifecycleUpdateReport", "MediaObjectLifecycleUpdateResult", "MediaObjectStats", "MediaObjectDuplicateReport", "MediaObjectDuplicateGroup", "MediaObjectVerificationResult", "MediaObjectVerificationReport", "PaginationMetadata", "ErrorEnvelope", "CatalogArtist", "CatalogAlbum", "CatalogTrack", "CatalogSearchResult", "SearchResultItem", "SearchResultKind", "CatalogArtistStatItem", "CatalogArtistStatsBreakdown", "CatalogAlbumStatItem", "CatalogAlbumStatsBreakdown", "RecentItemKind", "RecentCatalogItem", "RecentCatalogResult", "UpdatedCatalogItem", "UpdatedCatalogResult", "TrackPlaybackDescriptor", "CatalogPaginationMeta", "PlayEvent", "PlayEventList", "HistoryStats", "TrackPlayCount", "UserPlayCount", "TopTracksResult", "TopUsersResult"} {
		if _, ok := schemas[name].(map[string]any); !ok {
			t.Fatalf("schema %q is missing", name)
		}
	}

	errorEnvelope := schemas["ErrorEnvelope"].(map[string]any)
	errorProperty := errorEnvelope["properties"].(map[string]any)["error"].(map[string]any)
	codeProperty := errorProperty["properties"].(map[string]any)["code"].(map[string]any)
	enums := codeProperty["enum"].([]any)
	for _, code := range []string{"invalid_backend", "invalid_media_object", "unauthorized", "not_found", "method_not_allowed", "conflict", "probe_unsupported", "probe_failed", "capacity_unsupported", "internal_error", "admin_auth_not_configured", "media_registry_not_configured", "media_object_verification_unsupported", "media_object_verification_failed", "auth_not_configured", "invalid_user", "user_disabled", "missing_query", "catalog_not_configured", "invalid_catalog_entity", "import_rejected", "relink_rejected", "validation_error", "invalid_limit", "playback_unavailable", "invalid_offset", "invalid_sort_order", "history_not_configured", "invalid_since", "invalid_until", "invalid_time_range"} {
		if !containsString(enums, code) {
			t.Fatalf("error code %q is missing from OpenAPI enum %#v", code, enums)
		}
	}
}

func TestStorageAdminOpenAPIContractRecentTimelineSchemas(t *testing.T) {	document := loadOpenAPIContract(t)
	components := document["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)

	recentKind := schemas["RecentItemKind"].(map[string]any)
	for _, kind := range []string{"artist", "album", "track", "playlist"} {
		if !containsString(recentKind["enum"].([]any), kind) {
			t.Fatalf("RecentItemKind enum is missing %q", kind)
		}
	}

	for _, schemaName := range []string{"RecentCatalogItem", "UpdatedCatalogItem"} {
		schema := schemas[schemaName].(map[string]any)
		properties := schema["properties"].(map[string]any)
		playlist, ok := properties["playlist"].(map[string]any)
		if !ok {
			t.Fatalf("%s playlist payload property is missing", schemaName)
		}
		if playlist["$ref"] != "#/components/schemas/Playlist" {
			t.Fatalf("%s playlist ref = %#v, want Playlist schema ref", schemaName, playlist["$ref"])
		}
	}
}


func TestStorageAdminOpenAPIContractTrackPlaybackDescriptor(t *testing.T) {
	document := loadOpenAPIContract(t)
	components := document["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)

	desc := schemas["TrackPlaybackDescriptor"].(map[string]any)
	properties := desc["properties"].(map[string]any)
	for _, field := range []string{"trackId", "mediaObjectId", "mimeType", "durationMs", "backendId", "objectKey"} {
		if _, ok := properties[field]; !ok {
			t.Fatalf("TrackPlaybackDescriptor missing field %q", field)
		}
	}
	// presignedUrl is optional (omitempty) — must exist in properties but NOT in required.
	if _, ok := properties["presignedUrl"]; !ok {
		t.Fatal("TrackPlaybackDescriptor missing optional field \"presignedUrl\"")
	}
	required := desc["required"].([]any)
	for _, field := range []string{"trackId", "mediaObjectId", "mimeType", "durationMs", "backendId", "objectKey"} {
		if !containsString(required, field) {
			t.Fatalf("TrackPlaybackDescriptor required is missing %q", field)
		}
	}
	if containsString(required, "presignedUrl") {
		t.Fatal("presignedUrl must not be in required (it is optional)")
	}
}

func TestStorageAdminOpenAPIContractCatalogListSortParams(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	for _, path := range []string{
		"/api/v1/catalog/artists", "/api/v1/catalog/albums",
		"/api/v1/catalog/tracks", "/api/v1/catalog/playlists",
		"/api/v1/admin/catalog/artists", "/api/v1/admin/catalog/albums",
		"/api/v1/admin/catalog/tracks", "/api/v1/admin/catalog/playlists",
		"/api/v1/catalog/artists/{id}/albums", "/api/v1/catalog/artists/{id}/tracks",
		"/api/v1/catalog/albums/{id}/tracks",
		"/api/v1/admin/catalog/artists/{id}/albums", "/api/v1/admin/catalog/artists/{id}/tracks",
		"/api/v1/admin/catalog/albums/{id}/tracks",
	} {
		get := operation(t, paths, path, "get")
		params, _ := get["parameters"].([]any)
		seen := make(map[string]bool)
		for _, p := range params {
			if m, ok := p.(map[string]any); ok {
				seen[m["name"].(string)] = true
			}
		}
		for _, want := range []string{"limit", "offset", "sortBy", "sortOrder"} {
			if !seen[want] {
				t.Errorf("path %s GET is missing query param %q", path, want)
			}
		}
	}
}

func TestStorageAdminOpenAPIContractPlaylistTracksPagination(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	for _, path := range []string{
		"/api/v1/catalog/playlists/{id}/tracks",
		"/api/v1/admin/catalog/playlists/{id}/tracks",
	} {
		get := operation(t, paths, path, "get")
		params, _ := get["parameters"].([]any)
		seen := make(map[string]bool)
		for _, p := range params {
			if m, ok := p.(map[string]any); ok {
				seen[m["name"].(string)] = true
			}
		}
		for _, want := range []string{"limit", "offset"} {
			if !seen[want] {
				t.Errorf("path %s GET is missing pagination param %q", path, want)
			}
		}
		// playlist tracks preserve order — sortBy/sortOrder must NOT be present
		for _, noWant := range []string{"sortBy", "sortOrder"} {
			if seen[noWant] {
				t.Errorf("path %s GET should not have %q (playlist order is user-defined)", path, noWant)
			}
		}
	}
}

func loadOpenAPIContract(t *testing.T) map[string]any {
	t.Helper()
	path := filepath.Clean(filepath.Join("..", "..", "..", "..", "packages", "api-contract", "openapi", "storage-admin.v1.json"))
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(content, &document); err != nil {
		t.Fatalf("decode OpenAPI contract: %v", err)
	}
	if document["openapi"] != "3.1.0" {
		t.Fatalf("openapi version = %#v, want 3.1.0", document["openapi"])
	}
	return document
}

func operation(t *testing.T, paths map[string]any, path string, method string) map[string]any {
	t.Helper()
	pathItem, ok := paths[path].(map[string]any)
	if !ok {
		t.Fatalf("path %q is missing", path)
	}
	op, ok := pathItem[method].(map[string]any)
	if !ok {
		t.Fatalf("operation %s %s is missing", method, path)
	}
	return op
}

func containsString(values []any, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func TestStorageAdminOpenAPIContractAdminHistoryPaths(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	components := document["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)

	for _, path := range []string{
		"/api/v1/admin/history/stats",
		"/api/v1/admin/history/top-tracks",
		"/api/v1/admin/history/top-users",
	} {
		pathItem, ok := paths[path].(map[string]any)
		if !ok {
			t.Fatalf("OpenAPI path %q is missing", path)
		}
		if _, ok := pathItem["get"].(map[string]any); !ok {
			t.Fatalf("OpenAPI GET %s is missing", path)
		}
	}

	for _, name := range []string{"HistoryStats", "TrackPlayCount", "UserPlayCount", "TopTracksResult", "TopUsersResult"} {
		if _, ok := schemas[name].(map[string]any); !ok {
			t.Fatalf("schema %q is missing", name)
		}
	}

	// HistoryStats must have the three required fields
	histStats := schemas["HistoryStats"].(map[string]any)
	histProps := histStats["properties"].(map[string]any)
	for _, field := range []string{"totalEvents", "uniqueUsers", "uniqueTracks"} {
		if _, ok := histProps[field]; !ok {
			t.Errorf("HistoryStats missing property %q", field)
		}
	}
}

func TestStorageAdminOpenAPIContractAdminHistorySinceParam(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)

	for _, path := range []string{
		"/api/v1/admin/history/stats",
		"/api/v1/admin/history/top-tracks",
		"/api/v1/admin/history/top-users",
	} {
		get := operation(t, paths, path, "get")
		params, _ := get["parameters"].([]any)
		seen := false
		for _, p := range params {
			if m, ok := p.(map[string]any); ok && m["name"] == "since" {
				seen = true
				schema, _ := m["schema"].(map[string]any)
				if schema["type"] != "string" || schema["format"] != "date-time" {
					t.Errorf("%s since param schema = %#v, want string/date-time", path, schema)
				}
				if m["required"] == true {
					t.Errorf("%s since param must not be required", path)
				}
			}
		}
		if !seen {
			t.Errorf("%s GET is missing 'since' query parameter", path)
		}
	}
}

func TestStorageAdminOpenAPIContractAdminHistoryUntilParam(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)

	for _, path := range []string{
		"/api/v1/admin/history/stats",
		"/api/v1/admin/history/top-tracks",
		"/api/v1/admin/history/top-users",
	} {
		get := operation(t, paths, path, "get")
		params, _ := get["parameters"].([]any)
		seen := false
		for _, p := range params {
			if m, ok := p.(map[string]any); ok && m["name"] == "until" {
				seen = true
				schema, _ := m["schema"].(map[string]any)
				if schema["type"] != "string" || schema["format"] != "date-time" {
					t.Errorf("%s until param schema = %#v, want string/date-time", path, schema)
				}
				if m["required"] == true {
					t.Errorf("%s until param must not be required", path)
				}
			}
		}
		if !seen {
			t.Errorf("%s GET is missing 'until' query parameter", path)
		}
	}
}

func TestStorageAdminOpenAPIContractAdminHistoryDetailPaths(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)

	cases := []struct {
		path        string
		pathParam   string
		queryFilter string
	}{
		{"/api/v1/admin/history/users/{userId}", "userId", "trackId"},
		{"/api/v1/admin/history/tracks/{trackId}", "trackId", "userId"},
	}

	for _, tc := range cases {
		pathItem, ok := paths[tc.path].(map[string]any)
		if !ok {
			t.Fatalf("OpenAPI path %q is missing", tc.path)
		}
		get, ok := pathItem["get"].(map[string]any)
		if !ok {
			t.Fatalf("OpenAPI GET %s is missing", tc.path)
		}

		params, _ := get["parameters"].([]any)
		paramNames := map[string]bool{}
		for _, p := range params {
			if m, ok := p.(map[string]any); ok {
				paramNames[m["name"].(string)] = true
			}
		}

		for _, want := range []string{tc.pathParam, tc.queryFilter, "limit", "offset"} {
			if !paramNames[want] {
				t.Errorf("%s GET is missing parameter %q", tc.path, want)
			}
		}

		// Response must reference PlayEventList
		resp200, _ := get["responses"].(map[string]any)["200"].(map[string]any)
		if resp200 == nil {
			t.Errorf("%s GET missing 200 response", tc.path)
			continue
		}
		content, _ := resp200["content"].(map[string]any)
		appJSON, _ := content["application/json"].(map[string]any)
		schema, _ := appJSON["schema"].(map[string]any)
		if schema["$ref"] != "#/components/schemas/PlayEventList" {
			t.Errorf("%s 200 response schema = %v, want PlayEventList ref", tc.path, schema)
		}
	}
}

func TestStorageAdminOpenAPIContractAdminHistoryBulkDelete(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	components := document["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)

	// /api/v1/admin/history must have a delete operation with since/until params
	windowPath, ok := paths["/api/v1/admin/history"].(map[string]any)
	if !ok {
		t.Fatal("path /api/v1/admin/history is missing")
	}
	del, ok := windowPath["delete"].(map[string]any)
	if !ok {
		t.Fatal("DELETE /api/v1/admin/history is missing")
	}
	params, _ := del["parameters"].([]any)
	seen := map[string]bool{}
	for _, p := range params {
		if m, ok := p.(map[string]any); ok {
			seen[m["name"].(string)] = true
		}
	}
	for _, want := range []string{"since", "until"} {
		if !seen[want] {
			t.Errorf("DELETE /api/v1/admin/history is missing query param %q", want)
		}
	}

	// users/{userId} and tracks/{trackId} must have delete operations
	for _, path := range []string{
		"/api/v1/admin/history/users/{userId}",
		"/api/v1/admin/history/tracks/{trackId}",
	} {
		pathItem, ok := paths[path].(map[string]any)
		if !ok {
			t.Fatalf("path %q is missing", path)
		}
		if _, ok := pathItem["delete"].(map[string]any); !ok {
			t.Errorf("DELETE %s is missing", path)
		}
	}

	// missing_time_filter must be in the error code enum
	env := schemas["ErrorEnvelope"].(map[string]any)
	codes, _ := env["properties"].(map[string]any)["error"].(map[string]any)["properties"].(map[string]any)["code"].(map[string]any)["enum"].([]any)
	if !containsString(codes, "missing_time_filter") {
		t.Error("error code enum is missing 'missing_time_filter'")
	}
}

func TestStorageAdminOpenAPIContractViewerHistoryStatsPaths(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	components := document["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)

	// UserHistoryStats schema must exist with required fields
	us, ok := schemas["UserHistoryStats"].(map[string]any)
	if !ok {
		t.Fatal("schema UserHistoryStats is missing")
	}
	props, _ := us["properties"].(map[string]any)
	for _, want := range []string{"totalEvents", "uniqueTracks"} {
		if _, ok := props[want]; !ok {
			t.Errorf("UserHistoryStats is missing property %q", want)
		}
	}

	// GET /api/v1/me/history/stats must exist with since/until params
	statsGet := operation(t, paths, "/api/v1/me/history/stats", "get")
	statsParams := map[string]bool{}
	for _, p := range statsGet["parameters"].([]any) {
		if m, ok := p.(map[string]any); ok {
			statsParams[m["name"].(string)] = true
		}
	}
	for _, want := range []string{"since", "until"} {
		if !statsParams[want] {
			t.Errorf("GET /api/v1/me/history/stats missing param %q", want)
		}
	}
	// 200 response must reference UserHistoryStats
	resp200 := statsGet["responses"].(map[string]any)["200"].(map[string]any)
	content := resp200["content"].(map[string]any)["application/json"].(map[string]any)
	schema := content["schema"].(map[string]any)
	if schema["$ref"] != "#/components/schemas/UserHistoryStats" {
		t.Errorf("GET /api/v1/me/history/stats 200 schema = %v, want UserHistoryStats ref", schema)
	}

	// GET /api/v1/me/history/top-tracks must exist with limit/since/until params
	topGet := operation(t, paths, "/api/v1/me/history/top-tracks", "get")
	topParams := map[string]bool{}
	for _, p := range topGet["parameters"].([]any) {
		if m, ok := p.(map[string]any); ok {
			topParams[m["name"].(string)] = true
		}
	}
	for _, want := range []string{"limit", "since", "until"} {
		if !topParams[want] {
			t.Errorf("GET /api/v1/me/history/top-tracks missing param %q", want)
		}
	}
}

func TestStorageAdminOpenAPIContractAdminHistoryGlobalList(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)

	// GET /api/v1/admin/history must exist
	get := operation(t, paths, "/api/v1/admin/history", "get")

	// must carry all six query params
	params := map[string]bool{}
	for _, p := range get["parameters"].([]any) {
		if m, ok := p.(map[string]any); ok {
			params[m["name"].(string)] = true
		}
	}
	for _, want := range []string{"userId", "trackId", "since", "until", "limit", "offset"} {
		if !params[want] {
			t.Errorf("GET /api/v1/admin/history missing query param %q", want)
		}
	}

	// 200 response must reference PlayEventList
	resp200 := get["responses"].(map[string]any)["200"].(map[string]any)
	content := resp200["content"].(map[string]any)["application/json"].(map[string]any)
	schema := content["schema"].(map[string]any)
	if schema["$ref"] != "#/components/schemas/PlayEventList" {
		t.Errorf("GET /api/v1/admin/history 200 schema = %v, want PlayEventList ref", schema)
	}
}

func TestStorageAdminOpenAPIContractHistoryOrderParam(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	components := document["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)

	// All four list endpoints must carry the "order" query param
	listPaths := []string{
		"/api/v1/me/history",
		"/api/v1/admin/history/users/{userId}",
		"/api/v1/admin/history/tracks/{trackId}",
		"/api/v1/admin/history",
	}
	for _, p := range listPaths {
		get := operation(t, paths, p, "get")
		seen := map[string]bool{}
		for _, param := range get["parameters"].([]any) {
			if m, ok := param.(map[string]any); ok {
				seen[m["name"].(string)] = true
			}
		}
		if !seen["order"] {
			t.Errorf("GET %s is missing query param \"order\"", p)
		}
	}

	// invalid_order must be in the error code enum
	env := schemas["ErrorEnvelope"].(map[string]any)
	codes, _ := env["properties"].(map[string]any)["error"].(map[string]any)["properties"].(map[string]any)["code"].(map[string]any)["enum"].([]any)
	if !containsString(codes, "invalid_order") {
		t.Error("error code enum is missing 'invalid_order'")
	}
}

func TestStorageAdminOpenAPIContractPerEventPaths(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	components := document["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)

	// Both per-event paths must carry get and delete operations
	for _, p := range []string{
		"/api/v1/admin/history/{eventId}",
		"/api/v1/me/history/{eventId}",
	} {
		pathItem, ok := paths[p].(map[string]any)
		if !ok {
			t.Fatalf("path %q is missing", p)
		}
		for _, method := range []string{"get", "patch", "delete"} {
			if _, ok := pathItem[method].(map[string]any); !ok {
				t.Errorf("%s %s is missing", method, p)
			}
		}
		// GET must return a PlayEvent
		get := pathItem["get"].(map[string]any)
		resp200 := get["responses"].(map[string]any)["200"].(map[string]any)
		content := resp200["content"].(map[string]any)["application/json"].(map[string]any)
		schema := content["schema"].(map[string]any)
		if schema["$ref"] != "#/components/schemas/PlayEvent" {
			t.Errorf("GET %s 200 schema = %v, want PlayEvent ref", p, schema)
		}
		// PATCH must have a requestBody referencing UpdatePlayEventRequest
		patch := pathItem["patch"].(map[string]any)
		rb, _ := patch["requestBody"].(map[string]any)
		rbContent, _ := rb["content"].(map[string]any)
		rbSchema, _ := rbContent["application/json"].(map[string]any)["schema"].(map[string]any)
		if rbSchema["$ref"] != "#/components/schemas/UpdatePlayEventRequest" {
			t.Errorf("PATCH %s requestBody schema = %v, want UpdatePlayEventRequest ref", p, rbSchema)
		}
	}

	// PlayEvent schema must have required fields
	pe, ok := schemas["PlayEvent"].(map[string]any)
	if !ok {
		t.Fatal("schema PlayEvent is missing")
	}
	props, _ := pe["properties"].(map[string]any)
	for _, want := range []string{"id", "userId", "trackId", "playedAt"} {
		if _, ok := props[want]; !ok {
			t.Errorf("PlayEvent missing property %q", want)
		}
	}

	// UpdatePlayEventRequest schema must exist with playedAt
	upr, ok := schemas["UpdatePlayEventRequest"].(map[string]any)
	if !ok {
		t.Fatal("schema UpdatePlayEventRequest is missing")
	}
	uprProps, _ := upr["properties"].(map[string]any)
	if _, ok := uprProps["playedAt"]; !ok {
		t.Error("UpdatePlayEventRequest missing property \"playedAt\"")
	}

	// event_forbidden and invalid_played_at must be in error code enum
	env2 := schemas["ErrorEnvelope"].(map[string]any)
	codes2, _ := env2["properties"].(map[string]any)["error"].(map[string]any)["properties"].(map[string]any)["code"].(map[string]any)["enum"].([]any)
	for _, want := range []string{"event_forbidden", "invalid_played_at"} {
		if !containsString(codes2, want) {
			t.Errorf("error code enum is missing %q", want)
		}
	}
}

func TestStorageAdminOpenAPIContractBatchDelete(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	components := document["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)

	for _, p := range []string{
		"/api/v1/admin/history/batch-delete",
		"/api/v1/me/history/batch-delete",
	} {
		post := operation(t, paths, p, "post")

		rb, _ := post["requestBody"].(map[string]any)
		rbContent, _ := rb["content"].(map[string]any)
		rbSchema, _ := rbContent["application/json"].(map[string]any)["schema"].(map[string]any)
		if rbSchema["$ref"] != "#/components/schemas/BatchDeleteRequest" {
			t.Errorf("POST %s requestBody schema = %v, want BatchDeleteRequest ref", p, rbSchema)
		}

		resp200 := post["responses"].(map[string]any)["200"].(map[string]any)
		content := resp200["content"].(map[string]any)["application/json"].(map[string]any)
		schema := content["schema"].(map[string]any)
		if schema["$ref"] != "#/components/schemas/BatchDeleteResult" {
			t.Errorf("POST %s 200 schema = %v, want BatchDeleteResult ref", p, schema)
		}
	}

	bdr, ok := schemas["BatchDeleteRequest"].(map[string]any)
	if !ok {
		t.Fatal("schema BatchDeleteRequest is missing")
	}
	if _, ok := bdr["properties"].(map[string]any)["ids"]; !ok {
		t.Error("BatchDeleteRequest missing property \"ids\"")
	}

	bdrResult, ok := schemas["BatchDeleteResult"].(map[string]any)
	if !ok {
		t.Fatal("schema BatchDeleteResult is missing")
	}
	if _, ok := bdrResult["properties"].(map[string]any)["deleted"]; !ok {
		t.Error("BatchDeleteResult missing property \"deleted\"")
	}

	env := schemas["ErrorEnvelope"].(map[string]any)
	codes, _ := env["properties"].(map[string]any)["error"].(map[string]any)["properties"].(map[string]any)["code"].(map[string]any)["enum"].([]any)
	if !containsString(codes, "invalid_ids") {
		t.Error("error code enum is missing 'invalid_ids'")
	}
}

func TestStorageAdminOpenAPIContractListSinceUntilParams(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)

	for _, p := range []string{
		"/api/v1/me/history",
		"/api/v1/admin/history/users/{userId}",
		"/api/v1/admin/history/tracks/{trackId}",
	} {
		get := operation(t, paths, p, "get")
		seen := map[string]bool{}
		for _, param := range get["parameters"].([]any) {
			if m, ok := param.(map[string]any); ok {
				seen[m["name"].(string)] = true
			}
		}
		for _, want := range []string{"since", "until"} {
			if !seen[want] {
				t.Errorf("GET %s is missing query param %q", p, want)
			}
		}
	}
}

func TestStorageAdminOpenAPIContractAdminUserStatsPaths(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	components := document["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)

	// UserHistoryStats schema must already exist (added in Phase 74)
	if _, ok := schemas["UserHistoryStats"].(map[string]any); !ok {
		t.Fatal("schema UserHistoryStats is missing")
	}

	// GET /api/v1/admin/history/users/{userId}/stats must exist
	statsGet := operation(t, paths, "/api/v1/admin/history/users/{userId}/stats", "get")
	statsParams := map[string]bool{}
	for _, p := range statsGet["parameters"].([]any) {
		if m, ok := p.(map[string]any); ok {
			statsParams[m["name"].(string)] = true
		}
	}
	for _, want := range []string{"since", "until"} {
		if !statsParams[want] {
			t.Errorf("GET /api/v1/admin/history/users/{userId}/stats missing param %q", want)
		}
	}
	// 200 response must reference UserHistoryStats
	resp200 := statsGet["responses"].(map[string]any)["200"].(map[string]any)
	content := resp200["content"].(map[string]any)["application/json"].(map[string]any)
	schema := content["schema"].(map[string]any)
	if schema["$ref"] != "#/components/schemas/UserHistoryStats" {
		t.Errorf("GET /api/v1/admin/history/users/{userId}/stats 200 schema = %v, want UserHistoryStats ref", schema)
	}

	// GET /api/v1/admin/history/users/{userId}/top-tracks must exist
	topGet := operation(t, paths, "/api/v1/admin/history/users/{userId}/top-tracks", "get")
	topParams := map[string]bool{}
	for _, p := range topGet["parameters"].([]any) {
		if m, ok := p.(map[string]any); ok {
			topParams[m["name"].(string)] = true
		}
	}
	for _, want := range []string{"limit", "since", "until"} {
		if !topParams[want] {
			t.Errorf("GET /api/v1/admin/history/users/{userId}/top-tracks missing param %q", want)
		}
	}
}

func TestStorageAdminOpenAPIContractAdminTrackStatsPaths(t *testing.T) {
	document := loadOpenAPIContract(t)
	paths := document["paths"].(map[string]any)
	components := document["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)

	// TrackHistoryStats schema must exist with required fields
	ts, ok := schemas["TrackHistoryStats"].(map[string]any)
	if !ok {
		t.Fatal("schema TrackHistoryStats is missing")
	}
	props, _ := ts["properties"].(map[string]any)
	for _, want := range []string{"totalEvents", "uniqueListeners"} {
		if _, ok := props[want]; !ok {
			t.Errorf("TrackHistoryStats is missing property %q", want)
		}
	}

	// GET /api/v1/admin/history/tracks/{trackId}/stats must exist with since/until params
	statsGet := operation(t, paths, "/api/v1/admin/history/tracks/{trackId}/stats", "get")
	statsParams := map[string]bool{}
	for _, p := range statsGet["parameters"].([]any) {
		if m, ok := p.(map[string]any); ok {
			statsParams[m["name"].(string)] = true
		}
	}
	for _, want := range []string{"since", "until"} {
		if !statsParams[want] {
			t.Errorf("GET /api/v1/admin/history/tracks/{trackId}/stats missing param %q", want)
		}
	}
	// 200 response must reference TrackHistoryStats
	resp200 := statsGet["responses"].(map[string]any)["200"].(map[string]any)
	content := resp200["content"].(map[string]any)["application/json"].(map[string]any)
	schema := content["schema"].(map[string]any)
	if schema["$ref"] != "#/components/schemas/TrackHistoryStats" {
		t.Errorf("GET /api/v1/admin/history/tracks/{trackId}/stats 200 schema = %v, want TrackHistoryStats ref", schema)
	}

	// GET /api/v1/admin/history/tracks/{trackId}/top-listeners must exist with limit/since/until params
	topGet := operation(t, paths, "/api/v1/admin/history/tracks/{trackId}/top-listeners", "get")
	topParams := map[string]bool{}
	for _, p := range topGet["parameters"].([]any) {
		if m, ok := p.(map[string]any); ok {
			topParams[m["name"].(string)] = true
		}
	}
	for _, want := range []string{"limit", "since", "until"} {
		if !topParams[want] {
			t.Errorf("GET /api/v1/admin/history/tracks/{trackId}/top-listeners missing param %q", want)
		}
	}
	// 200 response must reference TopUsersResult
	topResp200 := topGet["responses"].(map[string]any)["200"].(map[string]any)
	topContent := topResp200["content"].(map[string]any)["application/json"].(map[string]any)
	topSchema := topContent["schema"].(map[string]any)
	if topSchema["$ref"] != "#/components/schemas/TopUsersResult" {
		t.Errorf("GET /api/v1/admin/history/tracks/{trackId}/top-listeners 200 schema = %v, want TopUsersResult ref", topSchema)
	}
}
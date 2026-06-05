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
		"/healthz":                                     {"get"},
		"/api/v1/admin/storage/backends":               {"get", "post"},
		"/api/v1/admin/storage/backends/validate":      {"post"},
		"/api/v1/admin/storage/backends/refresh":       {"post"},
		"/api/v1/admin/storage/backends/{id}/default":  {"post"},
		"/api/v1/admin/storage/backends/{id}/disable":  {"post"},
		"/api/v1/admin/storage/backends/{id}/probe":    {"post"},
		"/api/v1/admin/storage/backends/{id}/health":   {"get"},
		"/api/v1/admin/storage/backends/{id}/capacity": {"get"},
		"/api/v1/admin/media/objects":                  {"get", "post"},
		"/api/v1/admin/media/objects/stats":            {"get"},
		"/api/v1/admin/media/objects/duplicates":       {"get"},
		"/api/v1/admin/media/objects/lifecycle":        {"post"},
		"/api/v1/admin/media/objects/{id}":             {"get"},
		"/api/v1/admin/media/objects/{id}/lifecycle":   {"post"},
		"/api/v1/admin/media/objects/verify":           {"post"},
		"/api/v1/admin/media/objects/{id}/verify":      {"post"},
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
	if _, ok := parameters["BackendId"].(map[string]any); !ok {
		t.Fatal("OpenAPI BackendId path parameter is missing")
	}
	if _, ok := parameters["MediaObjectId"].(map[string]any); !ok {
		t.Fatal("OpenAPI MediaObjectId path parameter is missing")
	}

	for path, item := range paths {
		if !strings.Contains(path, "{id}") {
			continue
		}
		pathItem := item.(map[string]any)
		refs, ok := pathItem["parameters"].([]any)
		if !ok || len(refs) != 1 {
			t.Fatalf("path %s parameters = %#v, want BackendId reference", path, pathItem["parameters"])
		}
		want := "#/components/parameters/BackendId"
		if strings.HasPrefix(path, "/api/v1/admin/media/objects/") {
			want = "#/components/parameters/MediaObjectId"
		}
		ref := refs[0].(map[string]any)["$ref"]
		if ref != want {
			t.Fatalf("path %s parameter ref = %#v, want %s", path, ref, want)
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

	for path, item := range paths {
		if path == "/healthz" {
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
	for _, name := range []string{"StorageBackend", "StorageBackendRequest", "BackendConfig", "LocalConfig", "NFSConfig", "SMBConfig", "S3Config", "DistributedConfig", "CapabilitySet", "ProbeResult", "CapacityReport", "RefreshReport", "RefreshResult", "MediaObject", "MediaObjectRequest", "MediaObjectLifecycleRequest", "MediaObjectLifecycleChange", "MediaObjectSelectionFilter", "MediaObjectBulkLifecycleRequest", "MediaObjectLifecycleUpdateReport", "MediaObjectLifecycleUpdateResult", "MediaObjectStats", "MediaObjectDuplicateReport", "MediaObjectDuplicateGroup", "MediaObjectVerificationResult", "MediaObjectVerificationReport", "PaginationMetadata", "ErrorEnvelope"} {
		if _, ok := schemas[name].(map[string]any); !ok {
			t.Fatalf("schema %q is missing", name)
		}
	}

	errorEnvelope := schemas["ErrorEnvelope"].(map[string]any)
	errorProperty := errorEnvelope["properties"].(map[string]any)["error"].(map[string]any)
	codeProperty := errorProperty["properties"].(map[string]any)["code"].(map[string]any)
	enums := codeProperty["enum"].([]any)
	for _, code := range []string{"invalid_backend", "invalid_media_object", "unauthorized", "not_found", "method_not_allowed", "conflict", "probe_unsupported", "probe_failed", "capacity_unsupported", "internal_error", "admin_auth_not_configured", "media_registry_not_configured", "media_object_verification_unsupported", "media_object_verification_failed"} {
		if !containsString(enums, code) {
			t.Fatalf("error code %q is missing from OpenAPI enum %#v", code, enums)
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

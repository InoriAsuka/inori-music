package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"inori-music/services/api/internal/auth"
	"inori-music/services/api/internal/playerstate"
	"inori-music/services/api/internal/searchhistory"
	"inori-music/services/api/internal/storage"
)

// mustMarshalJSON is a test helper that marshals v to JSON bytes.
func mustMarshalJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func newPlayerStateTestHandler(t *testing.T) (http.Handler, string, string) {
	t.Helper()
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	psSvc := playerstate.NewService(playerstate.NewMemoryRepository())

	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
		WithPlayerstateService(psSvc),
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

func TestGetPlayerStateNotFound(t *testing.T) {
	h, viewerToken, _ := newPlayerStateTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/player-state", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}

func TestPutPlayerState(t *testing.T) {
	h, viewerToken, _ := newPlayerStateTestHandler(t)

	body := `{"queue":["t1","t2"],"currentIndex":0,"positionSeconds":45.5,"repeat":"all","shuffle":false,"volume":0.8,"speed":1.0,"status":"playing"}`
	resp := performRequestWithAuthHeader(t, h, http.MethodPut, "/api/v1/me/player-state", body, "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var state map[string]any
	json.Unmarshal(resp.Body.Bytes(), &state)
	if state["updatedAt"] == nil {
		t.Error("expected updatedAt in response")
	}
	if state["queue"] == nil {
		t.Error("expected queue in response")
	}
}

func TestPutPlayerStateQueueCap(t *testing.T) {
	h, viewerToken, _ := newPlayerStateTestHandler(t)

	queue := make([]string, 501)
	for i := range queue {
		queue[i] = "track"
	}
	body := `{"queue":` + string(mustMarshalJSON(queue)) + `,"currentIndex":0,"positionSeconds":0,"repeat":"off","shuffle":false,"volume":0,"speed":1,"status":"idle"}`
	resp := performRequestWithAuthHeader(t, h, http.MethodPut, "/api/v1/me/player-state", body, "Bearer "+viewerToken)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for queue > 500, got %d", resp.Code)
	}
}

func newSearchHistoryTestHandler(t *testing.T) (http.Handler, string, string) {
	t.Helper()
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	shSvc := searchhistory.NewService(searchhistory.NewMemoryRepository())

	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
		WithSearchHistoryService(shSvc),
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

func TestGetSearchHistoryEmpty(t *testing.T) {
	h, viewerToken, _ := newSearchHistoryTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/search-history", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	var queries []string
	json.Unmarshal(resp.Body.Bytes(), &queries)
	if len(queries) != 0 {
		t.Errorf("expected empty list, got %v", queries)
	}
}

func TestPutSearchHistory(t *testing.T) {
	h, viewerToken, _ := newSearchHistoryTestHandler(t)

	body := `["query1","query2","query3"]`
	resp := performRequestWithAuthHeader(t, h, http.MethodPut, "/api/v1/me/search-history", body, "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestPutSearchHistoryCap20(t *testing.T) {
	h, viewerToken, _ := newSearchHistoryTestHandler(t)

	queries := make([]string, 30)
	for i := range queries {
		queries[i] = "query" + string(rune('a'+i%26))
	}
	body := string(mustMarshalJSON(queries))
	resp := performRequestWithAuthHeader(t, h, http.MethodPut, "/api/v1/me/search-history", body, "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	// Verify only 20 returned (some deduped to ~26 unique, capped at 20)
	getResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/search-history", "", "Bearer "+viewerToken)
	var result []string
	json.Unmarshal(getResp.Body.Bytes(), &result)
	if len(result) != 20 {
		t.Errorf("expected 20 entries, got %d: %v", len(result), result)
	}
}

func TestDeleteSearchHistory(t *testing.T) {
	h, viewerToken, _ := newSearchHistoryTestHandler(t)

	// Add history first
	body := `["q1","q2"]`
	performRequestWithAuthHeader(t, h, http.MethodPut, "/api/v1/me/search-history", body, "Bearer "+viewerToken)

	// Delete
	resp := performRequestWithAuthHeader(t, h, http.MethodDelete, "/api/v1/me/search-history", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	// Verify empty
	getResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/search-history", "", "Bearer "+viewerToken)
	var result []string
	json.Unmarshal(getResp.Body.Bytes(), &result)
	if len(result) != 0 {
		t.Errorf("expected empty after delete, got %v", result)
	}
}

func TestEmptyWriteClearsSearchHistory(t *testing.T) {
	h, viewerToken, _ := newSearchHistoryTestHandler(t)

	body := `["q1","q2"]`
	performRequestWithAuthHeader(t, h, http.MethodPut, "/api/v1/me/search-history", body, "Bearer "+viewerToken)

	// Empty write should clear
	resp := performRequestWithAuthHeader(t, h, http.MethodPut, "/api/v1/me/search-history", `[]`, "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	getResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/search-history", "", "Bearer "+viewerToken)
	var result []string
	json.Unmarshal(getResp.Body.Bytes(), &result)
	if len(result) != 0 {
		t.Errorf("expected empty after clear, got %v", result)
	}
}

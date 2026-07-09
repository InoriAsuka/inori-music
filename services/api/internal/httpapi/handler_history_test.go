package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"inori-music/services/api/internal/auth"
	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/storage"
)

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

func TestViewerGetMyTrackHistory(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// Create artist + track
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song","artistId":%q,"mediaObjectId":"mo-mytrack-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Record 2 plays for this track
	for i := 0; i < 2; i++ {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/tracks/"+trackID, "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
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
}

func TestViewerGetMyTrackHistoryFiltersToOwnUser(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// Create a second viewer
	createResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/users",
		`{"username":"viewer2","password":"password99","role":"viewer"}`, "Bearer "+adminToken)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create viewer2: %d %s", createResp.Code, createResp.Body.String())
	}
	loginResp := performRequest(t, h, http.MethodPost, "/api/v1/auth/login",
		`{"username":"viewer2","password":"password99"}`)
	var loginBody map[string]any
	decodeResponse(t, loginResp, &loginBody)
	viewer2Token := loginBody["token"].(string)

	// Create track
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"Band2"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)
	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"Song2","artistId":%q,"mediaObjectId":"mo-mytrack-2"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// viewer1 records 1 play, viewer2 records 2 plays
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	for i := 0; i < 2; i++ {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewer2Token)
	}

	// viewer1 should only see their 1 play
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/tracks/"+trackID, "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d", resp.Code)
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	events := got["events"].([]any)
	if len(events) != 1 {
		t.Errorf("events = %d, want 1 (own plays only)", len(events))
	}
}

func TestViewerGetMyTrackHistoryMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/history/tracks/some-track", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.Code)
	}
}

func TestViewerGetMyTrackStats(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// Create track
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"BandStats"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)
	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"SongStats","artistId":%q,"mediaObjectId":"mo-tkstats-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Record 3 plays
	for i := 0; i < 3; i++ {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q}`, trackID), "Bearer "+viewerToken)
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/tracks/"+trackID+"/stats", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	if int(got["totalPlays"].(float64)) != 3 {
		t.Errorf("totalPlays = %v, want 3", got["totalPlays"])
	}
	if got["firstPlayedAt"] == nil || got["lastPlayedAt"] == nil {
		t.Error("firstPlayedAt and lastPlayedAt should be populated")
	}
}

func TestViewerGetMyTrackStatsNoPlays(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// Create track (no plays)
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"BandZero"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)
	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"SongZero","artistId":%q,"mediaObjectId":"mo-tkzero-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/tracks/"+trackID+"/stats", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var got map[string]any
	decodeResponse(t, resp, &got)
	if int(got["totalPlays"].(float64)) != 0 {
		t.Errorf("totalPlays = %v, want 0", got["totalPlays"])
	}
}

func TestViewerGetMyTrackStatsMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/history/tracks/some-track/stats", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.Code)
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

func TestViewerGetHistoryTimelineTrackIdFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TLFilterBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp1 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TLF1","artistId":%q,"mediaObjectId":"mo-tlf-1"}`, artistID), "Bearer "+adminToken)
	var track1 map[string]any
	decodeResponse(t, trackResp1, &track1)
	trackID1 := track1["id"].(string)

	trackResp2 := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TLF2","artistId":%q,"mediaObjectId":"mo-tlf-2"}`, artistID), "Bearer "+adminToken)
	var track2 map[string]any
	decodeResponse(t, trackResp2, &track2)
	trackID2 := track2["id"].(string)

	// track1: 2 plays on day1; track2: 1 play on day1
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-05-01T10:00:00Z"}`, trackID1), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-05-01T14:00:00Z"}`, trackID1), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-05-01T11:00:00Z"}`, trackID2), "Bearer "+viewerToken)

	// Without filter: 3 events in day1 bucket
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/timeline?since=2025-05-01T00:00:00Z&until=2025-05-02T00:00:00Z",
		"", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("timeline without filter: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	bucketsAll := result["buckets"].([]any)
	if len(bucketsAll) != 1 || int(bucketsAll[0].(map[string]any)["eventCount"].(float64)) != 3 {
		t.Errorf("without trackId filter: expected 1 bucket with 3 events, got %v", bucketsAll)
	}

	// With trackId filter: only 2 events for track1
	resp2 := performRequestWithAuthHeader(t, h, http.MethodGet,
		fmt.Sprintf("/api/v1/me/history/timeline?since=2025-05-01T00:00:00Z&until=2025-05-02T00:00:00Z&trackId=%s", trackID1),
		"", "Bearer "+viewerToken)
	if resp2.Code != http.StatusOK {
		t.Fatalf("timeline with trackId filter: %d %s", resp2.Code, resp2.Body.String())
	}
	var result2 map[string]any
	decodeResponse(t, resp2, &result2)
	bucketsFiltered := result2["buckets"].([]any)
	if len(bucketsFiltered) != 1 || int(bucketsFiltered[0].(map[string]any)["eventCount"].(float64)) != 2 {
		t.Errorf("with trackId filter: expected 1 bucket with 2 events, got %v", bucketsFiltered)
	}
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

func TestAdminGetTrackTimeline(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TLTrackBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TLTrk1","artistId":%q,"mediaObjectId":"mo-tltk-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// 2 events on day1, 1 on day2
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-06-01T10:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-06-01T15:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-06-02T09:00:00Z"}`, trackID), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/"+trackID+"/timeline?since=2025-06-01T00:00:00Z&until=2025-06-03T00:00:00Z&granularity=day",
		"", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getAdminTrackTimeline: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	buckets := result["buckets"].([]any)
	if len(buckets) != 2 {
		t.Fatalf("buckets = %d, want 2", len(buckets))
	}
	b0 := buckets[0].(map[string]any)
	if int(b0["eventCount"].(float64)) != 2 {
		t.Errorf("day1 eventCount = %v, want 2", b0["eventCount"])
	}
}

func TestAdminGetTrackTimelineMissingBounds(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/some-track/timeline?until=2025-06-03T00:00:00Z",
		"", "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusBadRequest, "missing_time_bounds")
}

func TestAdminGetTrackTimelineMethodNotAllowed(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/admin/history/tracks/some-track/timeline", "", "Bearer "+adminToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.Code)
	}
}

func TestAdminGetUserTimeline(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TLUserBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TLUsr1","artistId":%q,"mediaObjectId":"mo-tlu-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Get viewer user ID from /api/v1/me
	meResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me", "", "Bearer "+viewerToken)
	var me map[string]any
	decodeResponse(t, meResp, &me)
	userID := me["id"].(string)

	// 2 events on day1, 1 on day2
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-07-01T10:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-07-01T14:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-07-02T09:00:00Z"}`, trackID), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/"+userID+"/timeline?since=2025-07-01T00:00:00Z&until=2025-07-03T00:00:00Z&granularity=day",
		"", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getAdminUserTimeline: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	buckets := result["buckets"].([]any)
	if len(buckets) != 2 {
		t.Fatalf("buckets = %d, want 2", len(buckets))
	}
	b0 := buckets[0].(map[string]any)
	if int(b0["eventCount"].(float64)) != 2 {
		t.Errorf("day1 eventCount = %v, want 2", b0["eventCount"])
	}
}

func TestAdminGetUserTimelineMissingBounds(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/some-user/timeline?until=2025-07-03T00:00:00Z",
		"", "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusBadRequest, "missing_time_bounds")
}

func TestAdminGetUserTimelineMethodNotAllowed(t *testing.T) {
	h, _, adminToken := newHistoryTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/admin/history/users/some-user/timeline", "", "Bearer "+adminToken)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.Code)
	}
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

func TestViewerGetMyTrackTimeline(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TLMyTrackBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TLMyTrk1","artistId":%q,"mediaObjectId":"mo-tlmytk-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// 2 events on day1, 1 on day2 — all same viewer, same track
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-07-01T10:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-07-01T15:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-07-02T09:00:00Z"}`, trackID), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/tracks/"+trackID+"/timeline?since=2025-07-01T00:00:00Z&until=2025-07-03T00:00:00Z&granularity=day",
		"", "Bearer "+viewerToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	buckets, ok := body["buckets"].([]any)
	if !ok {
		t.Fatalf("expected buckets array, got %T", body["buckets"])
	}
	if len(buckets) != 2 {
		t.Fatalf("expected 2 day buckets, got %d", len(buckets))
	}
	b0 := buckets[0].(map[string]any)
	if int(b0["eventCount"].(float64)) != 2 {
		t.Errorf("day1 bucket: expected 2 events, got %v", b0["eventCount"])
	}
	b1 := buckets[1].(map[string]any)
	if int(b1["eventCount"].(float64)) != 1 {
		t.Errorf("day2 bucket: expected 1 event, got %v", b1["eventCount"])
	}
}

func TestViewerGetMyTrackTimelineMissingBounds(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/tracks/some-track/timeline?since=2025-07-01T00:00:00Z",
		"", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusBadRequest, "missing_time_bounds")
}

func TestViewerGetMyTrackTimelineMethodNotAllowed(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	resp := performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/history/tracks/some-track/timeline",
		"", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusMethodNotAllowed, "method_not_allowed")
}

func TestAdminGetUserHistorySummary(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// record 3 plays for track-sum-1, 1 play for track-sum-2 under viewerToken's user
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"track-sum-1","playedAt":"2025-08-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"track-sum-1","playedAt":"2025-08-01T11:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"track-sum-1","playedAt":"2025-08-01T12:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"track-sum-2","playedAt":"2025-08-01T13:00:00Z"}`, "Bearer "+viewerToken)

	// get viewer's userID
	meResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me", "", "Bearer "+viewerToken)
	var meBody map[string]any
	decodeResponse(t, meResp, &meBody)
	userID := meBody["id"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/"+userID+"/history-summary",
		"", "Bearer "+adminToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	stats, ok := body["stats"].(map[string]any)
	if !ok {
		t.Fatalf("expected stats object, got %T", body["stats"])
	}
	if int(stats["totalEvents"].(float64)) != 4 {
		t.Errorf("stats.totalEvents = %v, want 4", stats["totalEvents"])
	}
	topTracks, ok := body["topTracks"].([]any)
	if !ok {
		t.Fatalf("expected topTracks array, got %T", body["topTracks"])
	}
	if len(topTracks) == 0 {
		t.Error("expected at least one top track")
	}
}

func TestAdminGetUserHistorySummaryWithTopN(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	for _, tid := range []string{"ts-a", "ts-b", "ts-c"} {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-08-02T10:00:00Z"}`, tid), "Bearer "+viewerToken)
	}

	meResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me", "", "Bearer "+viewerToken)
	var meBody map[string]any
	decodeResponse(t, meResp, &meBody)
	userID := meBody["id"].(string)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/"+userID+"/history-summary?limit=1",
		"", "Bearer "+adminToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	topTracks := body["topTracks"].([]any)
	if len(topTracks) != 1 {
		t.Errorf("expected 1 top track with limit=1, got %d", len(topTracks))
	}
}

func TestAdminGetUserHistorySummaryNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/some-user/history-summary",
		"", "Bearer "+testAdminToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
}

func TestAdminGetTrackHistorySummary(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// 3 plays for user1, 1 play for user2 — all on track-hsum
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"track-hsum","playedAt":"2025-09-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"track-hsum","playedAt":"2025-09-01T11:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"track-hsum","playedAt":"2025-09-01T12:00:00Z"}`, "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/track-hsum/history-summary",
		"", "Bearer "+adminToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	stats, ok := body["stats"].(map[string]any)
	if !ok {
		t.Fatalf("expected stats object, got %T", body["stats"])
	}
	if int(stats["totalEvents"].(float64)) != 3 {
		t.Errorf("stats.totalEvents = %v, want 3", stats["totalEvents"])
	}
	topListeners, ok := body["topListeners"].([]any)
	if !ok {
		t.Fatalf("expected topListeners array, got %T", body["topListeners"])
	}
	if len(topListeners) == 0 {
		t.Error("expected at least one top listener")
	}
}

func TestAdminGetTrackHistorySummaryWithTopN(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"track-hsum2","playedAt":"2025-09-02T10:00:00Z"}`, "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/track-hsum2/history-summary?limit=1",
		"", "Bearer "+adminToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	topListeners := body["topListeners"].([]any)
	if len(topListeners) != 1 {
		t.Errorf("expected 1 top listener with limit=1, got %d", len(topListeners))
	}
}

func TestAdminGetTrackHistorySummaryNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/some-track/history-summary",
		"", "Bearer "+testAdminToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
}

func TestViewerGetMyHistorySummary(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"sum-track-a","playedAt":"2025-10-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"sum-track-a","playedAt":"2025-10-01T11:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"sum-track-b","playedAt":"2025-10-01T12:00:00Z"}`, "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/summary",
		"", "Bearer "+viewerToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	stats, ok := body["stats"].(map[string]any)
	if !ok {
		t.Fatalf("expected stats object, got %T", body["stats"])
	}
	if int(stats["totalEvents"].(float64)) != 3 {
		t.Errorf("stats.totalEvents = %v, want 3", stats["totalEvents"])
	}
	topTracks, ok := body["topTracks"].([]any)
	if !ok {
		t.Fatalf("expected topTracks array, got %T", body["topTracks"])
	}
	if len(topTracks) == 0 {
		t.Error("expected at least one top track")
	}
}

func TestViewerGetMyHistorySummaryWithTopN(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	for _, tid := range []string{"s-a", "s-b", "s-c"} {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-10-02T10:00:00Z"}`, tid), "Bearer "+viewerToken)
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/summary?limit=1",
		"", "Bearer "+viewerToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	topTracks := body["topTracks"].([]any)
	if len(topTracks) != 1 {
		t.Errorf("expected 1 top track with limit=1, got %d", len(topTracks))
	}
}

func TestViewerGetMyHistorySummaryNotConfigured(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "viewersum", "viewerpassSUM1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, _ := authSvc.Login(context.Background(), "viewersum", "viewerpassSUM1")

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/history/summary", "", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
}

func TestViewerGetMyTrackStatsSince(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	// 2 plays before cutoff, 1 play after
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"tss-1","playedAt":"2025-11-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"tss-1","playedAt":"2025-11-01T12:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"tss-1","playedAt":"2025-11-03T10:00:00Z"}`, "Bearer "+viewerToken)

	// since=day3: only 1 play in window
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/tracks/tss-1/stats?since=2025-11-03T00:00:00Z",
		"", "Bearer "+viewerToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if int(body["totalPlays"].(float64)) != 1 {
		t.Errorf("totalPlays with since=day3 = %v, want 1", body["totalPlays"])
	}
}

func TestViewerGetMyTrackStatsUntil(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"tsu-1","playedAt":"2025-11-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"tsu-1","playedAt":"2025-11-02T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"tsu-1","playedAt":"2025-11-03T10:00:00Z"}`, "Bearer "+viewerToken)

	// until=day2 (exclusive): 1 play (day1 only)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/tracks/tsu-1/stats?until=2025-11-02T00:00:00Z",
		"", "Bearer "+viewerToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if int(body["totalPlays"].(float64)) != 1 {
		t.Errorf("totalPlays with until=day2 = %v, want 1", body["totalPlays"])
	}
}

func TestAdminGetHistorySummary(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	// 2 events on different tracks so stats reflect aggregation.
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"gs-t1","playedAt":"2025-10-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"gs-t2","playedAt":"2025-10-01T11:00:00Z"}`, "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/summary",
		"", "Bearer "+adminToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	stats, ok := body["stats"].(map[string]any)
	if !ok {
		t.Fatalf("expected stats object, got %T", body["stats"])
	}
	if int(stats["totalEvents"].(float64)) < 2 {
		t.Errorf("stats.totalEvents = %v, want ≥ 2", stats["totalEvents"])
	}
	if _, ok := body["topTracks"].([]any); !ok {
		t.Fatalf("expected topTracks array, got %T", body["topTracks"])
	}
	if _, ok := body["topUsers"].([]any); !ok {
		t.Fatalf("expected topUsers array, got %T", body["topUsers"])
	}
}

func TestAdminGetHistorySummaryWithTopN(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	for _, tid := range []string{"gs2-t1", "gs2-t2", "gs2-t3"} {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-10-02T10:00:00Z"}`, tid), "Bearer "+viewerToken)
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/summary?limit=1",
		"", "Bearer "+adminToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	topTracks := body["topTracks"].([]any)
	if len(topTracks) != 1 {
		t.Errorf("expected 1 top track with limit=1, got %d", len(topTracks))
	}
}

func TestAdminGetHistorySummaryNotConfigured(t *testing.T) {
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
	).Routes()
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/summary",
		"", "Bearer "+testAdminToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
}

func TestViewerGetMyTrackSummary(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	// Record 2 plays for the specific track and 1 for another.
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"ts-track","playedAt":"2025-11-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"ts-track","playedAt":"2025-11-01T11:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"ts-other","playedAt":"2025-11-01T12:00:00Z"}`, "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/tracks/ts-track/summary",
		"", "Bearer "+viewerToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	stats, ok := body["stats"].(map[string]any)
	if !ok {
		t.Fatalf("expected stats object, got %T", body["stats"])
	}
	if int(stats["totalPlays"].(float64)) != 2 {
		t.Errorf("stats.totalPlays = %v, want 2", stats["totalPlays"])
	}
	if _, ok := body["recentTracks"].([]any); !ok {
		t.Fatalf("expected recentTracks array, got %T", body["recentTracks"])
	}
}

func TestViewerGetMyTrackSummaryWithTopN(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	for _, tid := range []string{"tsn-a", "tsn-b", "tsn-c"} {
		performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
			fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-11-02T10:00:00Z"}`, tid), "Bearer "+viewerToken)
	}

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/tracks/tsn-a/summary?limit=1",
		"", "Bearer "+viewerToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	recentTracks := body["recentTracks"].([]any)
	if len(recentTracks) != 1 {
		t.Errorf("expected 1 recent track with limit=1, got %d", len(recentTracks))
	}
}

func TestViewerGetMyTrackSummaryNotConfigured(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "viewertsumNC", "viewertsumNC1!", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, _ := authSvc.Login(context.Background(), "viewertsumNC", "viewertsumNC1!")
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/tracks/some-track/summary",
		"", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "history_not_configured")
}

func TestAdminGetTrackStatsSince(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"ts-tw","playedAt":"2025-01-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"ts-tw","playedAt":"2025-01-01T12:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"ts-tw","playedAt":"2025-01-03T10:00:00Z"}`, "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/ts-tw/stats?since=2025-01-02T00:00:00Z",
		"", "Bearer "+adminToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if int(body["totalEvents"].(float64)) != 1 {
		t.Errorf("totalEvents with since=day2 = %v, want 1", body["totalEvents"])
	}
}

func TestAdminGetTrackStatsUntil(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"ts-tw2","playedAt":"2025-02-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"ts-tw2","playedAt":"2025-02-01T12:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"ts-tw2","playedAt":"2025-02-03T10:00:00Z"}`, "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/ts-tw2/stats?until=2025-02-02T00:00:00Z",
		"", "Bearer "+adminToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	if int(body["totalEvents"].(float64)) != 2 {
		t.Errorf("totalEvents with until=day2 = %v, want 2", body["totalEvents"])
	}
}

func TestAdminGetHistoryStatsSince(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"hs-tw","playedAt":"2025-03-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"hs-tw","playedAt":"2025-03-01T12:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"hs-tw","playedAt":"2025-03-03T10:00:00Z"}`, "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/stats?since=2025-03-02T00:00:00Z",
		"", "Bearer "+adminToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body2 map[string]any
	decodeResponse(t, resp, &body2)
	if int(body2["totalEvents"].(float64)) != 1 {
		t.Errorf("totalEvents with since=day2 = %v, want 1", body2["totalEvents"])
	}
}

func TestAdminGetHistoryStatsUntil(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"hs-tw2","playedAt":"2025-04-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"hs-tw2","playedAt":"2025-04-01T12:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"hs-tw2","playedAt":"2025-04-03T10:00:00Z"}`, "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/stats?until=2025-04-02T00:00:00Z",
		"", "Bearer "+adminToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body3 map[string]any
	decodeResponse(t, resp, &body3)
	if int(body3["totalEvents"].(float64)) != 2 {
		t.Errorf("totalEvents with until=day2 = %v, want 2", body3["totalEvents"])
	}
}

func TestViewerGetMyTopTracksSince(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"mtt-a","playedAt":"2025-05-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"mtt-a","playedAt":"2025-05-01T12:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"mtt-b","playedAt":"2025-05-03T10:00:00Z"}`, "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/top-tracks?since=2025-05-02T00:00:00Z",
		"", "Bearer "+viewerToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	tracks := body["tracks"].([]any)
	if len(tracks) != 1 {
		t.Errorf("len(tracks) with since=day2 = %d, want 1", len(tracks))
	}
}

func TestViewerGetMyTopTracksUntil(t *testing.T) {
	h, viewerToken, _ := newHistoryTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"mtu-a","playedAt":"2025-06-01T10:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"mtu-a","playedAt":"2025-06-01T12:00:00Z"}`, "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		`{"trackId":"mtu-b","playedAt":"2025-06-03T10:00:00Z"}`, "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history/top-tracks?until=2025-06-02T00:00:00Z",
		"", "Bearer "+viewerToken)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	decodeResponse(t, resp, &body)
	tracks := body["tracks"].([]any)
	if len(tracks) != 1 {
		t.Errorf("len(tracks) with until=day2 = %d, want 1", len(tracks))
	}
}

// TestAdminGetUserTimelineSinceFilter verifies that ?since restricts the
// timeline to events on or after that timestamp.
func TestAdminGetUserTimelineSinceFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TLUSince"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TLUSince1","artistId":%q,"mediaObjectId":"mo-tlus-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	meResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me", "", "Bearer "+viewerToken)
	var me map[string]any
	decodeResponse(t, meResp, &me)
	userID := me["id"].(string)

	// Event on day1 (before since) and day2 (within range)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-08-01T10:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-08-02T10:00:00Z"}`, trackID), "Bearer "+viewerToken)

	// since=day2: only day2 event should appear
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/users/"+userID+"/timeline?since=2025-08-02T00:00:00Z&until=2025-08-03T00:00:00Z&granularity=day",
		"", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getAdminUserTimeline since filter: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	buckets := result["buckets"].([]any)
	if len(buckets) != 1 {
		t.Fatalf("buckets = %d, want 1 (since filter)", len(buckets))
	}
	b0 := buckets[0].(map[string]any)
	if int(b0["eventCount"].(float64)) != 1 {
		t.Errorf("day2 eventCount = %v, want 1", b0["eventCount"])
	}
}

// TestAdminGetTrackTimelineSinceFilter verifies that ?since restricts the
// timeline to events on or after that timestamp.
func TestAdminGetTrackTimelineSinceFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"TLTkSince"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"TLTkSince1","artistId":%q,"mediaObjectId":"mo-tltks-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Event on day1 (before since) and day2 (within range)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-09-01T10:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-09-02T10:00:00Z"}`, trackID), "Bearer "+viewerToken)

	// since=day2: only day2 event should appear
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history/tracks/"+trackID+"/timeline?since=2025-09-02T00:00:00Z&until=2025-09-03T00:00:00Z&granularity=day",
		"", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("getAdminTrackTimeline since filter: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	buckets := result["buckets"].([]any)
	if len(buckets) != 1 {
		t.Fatalf("buckets = %d, want 1 (since filter)", len(buckets))
	}
	b0 := buckets[0].(map[string]any)
	if int(b0["eventCount"].(float64)) != 1 {
		t.Errorf("day2 eventCount = %v, want 1", b0["eventCount"])
	}
}

// TestViewerListPlayEventsTrackIdFilter verifies that GET /api/v1/me/history?trackId=
// returns only events for the specified track.
func TestViewerListPlayEventsTrackIdFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"FilterArtist"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	track1Resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"FilterTrk1","artistId":%q,"mediaObjectId":"mo-ftf-1"}`, artistID), "Bearer "+adminToken)
	var track1 map[string]any
	decodeResponse(t, track1Resp, &track1)
	trackID1 := track1["id"].(string)

	track2Resp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"FilterTrk2","artistId":%q,"mediaObjectId":"mo-ftf-2"}`, artistID), "Bearer "+adminToken)
	var track2 map[string]any
	decodeResponse(t, track2Resp, &track2)
	trackID2 := track2["id"].(string)

	// 2 events for track1, 1 for track2
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID1), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID1), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q}`, trackID2), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/history?trackId="+trackID1,
		"", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("listPlayEvents trackId filter: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	events := result["events"].([]any)
	if len(events) != 2 {
		t.Errorf("events with trackId filter = %d, want 2", len(events))
	}
	for _, ev := range events {
		e := ev.(map[string]any)
		if e["trackId"] != trackID1 {
			t.Errorf("event trackId = %v, want %v", e["trackId"], trackID1)
		}
	}
}

// TestAdminGetAllHistorySinceFilter verifies that GET /api/v1/admin/history?since=
// restricts the global event list to events on or after the given timestamp.
func TestAdminGetAllHistorySinceFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"SinceFilterBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"SFT1","artistId":%q,"mediaObjectId":"mo-agh-s-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Event before and after cutoff
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-10-01T10:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-10-03T10:00:00Z"}`, trackID), "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history?since=2025-10-02T00:00:00Z",
		"", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin/history?since: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	pagination := result["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 1 {
		t.Errorf("total with since filter = %v, want 1", pagination["total"])
	}
}

// TestAdminGetAllHistoryUntilFilter verifies that GET /api/v1/admin/history?until=
// excludes events on or after the given timestamp.
func TestAdminGetAllHistoryUntilFilter(t *testing.T) {
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"UntilFilterBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"UFT1","artistId":%q,"mediaObjectId":"mo-agh-u-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	// Event before and after cutoff
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-11-01T10:00:00Z"}`, trackID), "Bearer "+viewerToken)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/me/history",
		fmt.Sprintf(`{"trackId":%q,"playedAt":"2025-11-03T10:00:00Z"}`, trackID), "Bearer "+viewerToken)

	// until=Nov 2: only the first event should match (Nov 3 is excluded)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/history?until=2025-11-02T00:00:00Z",
		"", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin/history?until: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	pagination := result["pagination"].(map[string]any)
	if int(pagination["total"].(float64)) != 1 {
		t.Errorf("total with until filter = %v, want 1", pagination["total"])
	}
}

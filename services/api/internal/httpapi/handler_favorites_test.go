package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"inori-music/services/api/internal/auth"
	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/favorites"
	"inori-music/services/api/internal/history"
	"inori-music/services/api/internal/storage"
)

// ---- favorites handler tests (Phase 137) ----

// newFavoritesTestHandler builds a handler with auth, catalog, and favorites wired up.
// Returns: handler, viewer bearer token, admin bearer token, trackID of a seeded track.
func newFavoritesTestHandler(t *testing.T) (http.Handler, string, string, string) {
	t.Helper()
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	catalogRepo := catalog.NewMemoryRepository()
	catalogSvc := catalog.NewService(catalogRepo)
	favSvc := favorites.NewService(favorites.NewMemoryRepository())

	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
		WithCatalogService(catalogSvc),
		WithHistoryService(history.NewService(history.NewMemoryRepository())),
		WithFavoritesService(favSvc),
	).Routes()

	// Create viewer and admin users
	if _, err := authSvc.CreateUser(context.Background(), "viewer1", "viewerpass1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "viewer1", "viewerpass1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "admin1", "adminpass1", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, err := authSvc.Login(context.Background(), "admin1", "adminpass1")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}

	// Seed an artist and track
	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists",
		`{"name":"FavBand"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	trackResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"FavSong","artistId":%q,"mediaObjectId":"fav-mo-1"}`, artistID), "Bearer "+adminToken)
	var track map[string]any
	decodeResponse(t, trackResp, &track)
	trackID := track["id"].(string)

	return h, viewerToken, adminToken, trackID
}

func TestAddFavoriteTrack(t *testing.T) {
	h, viewerToken, _, trackID := newFavoritesTestHandler(t)

	// Add favorite — expect 200
	resp := performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/favorites/tracks/"+trackID, "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("POST /me/favorites/tracks/{id} status = %d, want 200; body = %s", resp.Code, resp.Body.String())
	}

	// Idempotent: second add is also 200
	resp2 := performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/favorites/tracks/"+trackID, "", "Bearer "+viewerToken)
	if resp2.Code != http.StatusOK {
		t.Fatalf("duplicate POST favorites status = %d, want 200", resp2.Code)
	}

	// Without auth → 401
	resp3 := performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/favorites/tracks/"+trackID, "", "")
	if resp3.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated POST favorites status = %d, want 401", resp3.Code)
	}
}

func TestRemoveFavoriteTrack(t *testing.T) {
	h, viewerToken, _, trackID := newFavoritesTestHandler(t)

	// Add then remove
	performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/favorites/tracks/"+trackID, "", "Bearer "+viewerToken)

	resp := performRequestWithAuthHeader(t, h, http.MethodDelete,
		"/api/v1/me/favorites/tracks/"+trackID, "", "Bearer "+viewerToken)
	if resp.Code != http.StatusNoContent {
		t.Fatalf("DELETE /me/favorites/tracks/{id} status = %d, want 204; body = %s", resp.Code, resp.Body.String())
	}

	// Idempotent remove of already-removed → 204
	resp2 := performRequestWithAuthHeader(t, h, http.MethodDelete,
		"/api/v1/me/favorites/tracks/"+trackID, "", "Bearer "+viewerToken)
	if resp2.Code != http.StatusNoContent {
		t.Fatalf("idempotent DELETE favorites status = %d, want 204", resp2.Code)
	}
}

func TestListFavoriteTracks(t *testing.T) {
	h, viewerToken, _, trackID := newFavoritesTestHandler(t)

	// Empty list
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/favorites/tracks", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("empty GET favorites status = %d; body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	pagination := result["pagination"].(map[string]any)
	if pagination["total"].(float64) != 0 {
		t.Fatalf("expected total=0, got %v", pagination["total"])
	}

	// Add track and list again
	performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/favorites/tracks/"+trackID, "", "Bearer "+viewerToken)

	resp2 := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/me/favorites/tracks", "", "Bearer "+viewerToken)
	if resp2.Code != http.StatusOK {
		t.Fatalf("GET favorites after add status = %d; body = %s", resp2.Code, resp2.Body.String())
	}
	var result2 map[string]any
	decodeResponse(t, resp2, &result2)
	pagination2 := result2["pagination"].(map[string]any)
	if pagination2["total"].(float64) != 1 {
		t.Fatalf("expected total=1 after adding favorite, got %v", pagination2["total"])
	}
	// When catalog is wired, response should include full tracks with isFavorite=true
	if tracks, ok := result2["tracks"].([]any); ok {
		if len(tracks) != 1 {
			t.Fatalf("expected 1 track, got %d", len(tracks))
		}
		track := tracks[0].(map[string]any)
		if track["isFavorite"] != true {
			t.Errorf("track isFavorite = %v, want true", track["isFavorite"])
		}
		if track["id"] != trackID {
			t.Errorf("track id = %v, want %q", track["id"], trackID)
		}
	}
}

func TestListFavoritesNotConfigured(t *testing.T) {
	// Handler with auth but no favorites service — token must be valid
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	if _, err := authSvc.CreateUser(context.Background(), "viewer_nc", "pass12345", auth.RoleViewer); err != nil {
		t.Fatal(err)
	}
	tok, _, err := authSvc.Login(context.Background(), "viewer_nc", "pass12345")
	if err != nil {
		t.Fatal(err)
	}
	h := NewHandler(
		storage.NewService(storage.NewMemoryRepository()),
		WithAdminToken(testAdminToken),
		WithAuthService(authSvc),
	).Routes()
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me/favorites/tracks", "", "Bearer "+tok)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "favorites_not_configured")
}

func TestAdminListUserFavorites(t *testing.T) {
	h, viewerToken, adminToken, trackID := newFavoritesTestHandler(t)

	// Add a favorite as viewer
	performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/favorites/tracks/"+trackID, "", "Bearer "+viewerToken)

	// Get viewer user ID
	meResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me", "", "Bearer "+viewerToken)
	var me map[string]any
	decodeResponse(t, meResp, &me)
	userID := me["id"].(string)

	// Admin lists the user's favorites
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/favorites/users/"+userID+"/tracks", "", "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("admin GET user favorites status = %d; body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	pagination := result["pagination"].(map[string]any)
	if pagination["total"].(float64) != 1 {
		t.Fatalf("expected total=1, got %v", pagination["total"])
	}
	trackIDs := result["trackIds"].([]any)
	if len(trackIDs) != 1 || trackIDs[0] != trackID {
		t.Fatalf("trackIds = %v, want [%q]", trackIDs, trackID)
	}
}

func TestAdminClearUserFavorites(t *testing.T) {
	h, viewerToken, adminToken, trackID := newFavoritesTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/favorites/tracks/"+trackID, "", "Bearer "+viewerToken)

	meResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me", "", "Bearer "+viewerToken)
	var me map[string]any
	decodeResponse(t, meResp, &me)
	userID := me["id"].(string)

	// Clear all favorites
	resp := performRequestWithAuthHeader(t, h, http.MethodDelete,
		"/api/v1/admin/favorites/users/"+userID+"/tracks", "", "Bearer "+adminToken)
	if resp.Code != http.StatusNoContent {
		t.Fatalf("admin DELETE all favorites status = %d; body = %s", resp.Code, resp.Body.String())
	}

	// Verify empty
	listResp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/favorites/users/"+userID+"/tracks", "", "Bearer "+adminToken)
	var result map[string]any
	decodeResponse(t, listResp, &result)
	if result["pagination"].(map[string]any)["total"].(float64) != 0 {
		t.Error("expected 0 favorites after clear")
	}
}

func TestAdminRemoveUserFavoriteTrack(t *testing.T) {
	h, viewerToken, adminToken, trackID := newFavoritesTestHandler(t)

	performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/favorites/tracks/"+trackID, "", "Bearer "+viewerToken)

	meResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/me", "", "Bearer "+viewerToken)
	var me map[string]any
	decodeResponse(t, meResp, &me)
	userID := me["id"].(string)

	// Remove single track favorite
	resp := performRequestWithAuthHeader(t, h, http.MethodDelete,
		"/api/v1/admin/favorites/users/"+userID+"/tracks/"+trackID, "", "Bearer "+adminToken)
	if resp.Code != http.StatusNoContent {
		t.Fatalf("admin DELETE single favorite status = %d; body = %s", resp.Code, resp.Body.String())
	}

	// Idempotent
	resp2 := performRequestWithAuthHeader(t, h, http.MethodDelete,
		"/api/v1/admin/favorites/users/"+userID+"/tracks/"+trackID, "", "Bearer "+adminToken)
	if resp2.Code != http.StatusNoContent {
		t.Fatalf("idempotent admin DELETE single favorite status = %d", resp2.Code)
	}
}

func TestCatalogTrackIsFavoriteInViewerList(t *testing.T) {
	h, viewerToken, adminToken, trackID := newFavoritesTestHandler(t)

	// Add favorite
	performRequestWithAuthHeader(t, h, http.MethodPost,
		"/api/v1/me/favorites/tracks/"+trackID, "", "Bearer "+viewerToken)

	// GET viewer catalog/tracks — should carry isFavorite=true for the favorited track
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/catalog/tracks", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /catalog/tracks status = %d; body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	tracks := result["tracks"].([]any)
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	track := tracks[0].(map[string]any)
	if track["isFavorite"] != true {
		t.Errorf("track isFavorite = %v, want true (track was favorited)", track["isFavorite"])
	}

	// Admin catalog/tracks — isFavorite must be false (admin path)
	adminResp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/admin/catalog/tracks", "", "Bearer "+adminToken)
	var adminResult map[string]any
	decodeResponse(t, adminResp, &adminResult)
	adminTracks := adminResult["tracks"].([]any)
	adminTrack := adminTracks[0].(map[string]any)
	if adminTrack["isFavorite"] == true {
		t.Error("admin track isFavorite should be false, got true")
	}
}

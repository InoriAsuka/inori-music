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

// newArtworkViewerTestHandler builds a handler wired with session auth, catalog, storage,
// and media objects. Returns the handler, a viewer bearer token, and an admin bearer token.
func newArtworkViewerTestHandler(t *testing.T) (http.Handler, string, string) {
	t.Helper()
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	repo := storage.NewMemoryRepository()
	mediaRepo := storage.NewMemoryMediaObjectRepository()
	catalogSvc := catalog.NewService(catalog.NewMemoryRepository())
	mediaSvc := storage.NewMediaObjectService(repo, mediaRepo)
	h := NewHandler(
		storage.NewService(repo),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
		WithCatalogService(catalogSvc),
		WithMediaObjectService(mediaSvc),
		WithHistoryService(history.NewService(history.NewMemoryRepository())),
		WithFavoritesService(favorites.NewService(favorites.NewMemoryRepository())),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "viewerart", "passartwork1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "viewerart", "passartwork1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "adminart", "adminartwork1", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, err := authSvc.Login(context.Background(), "adminart", "adminartwork1")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	return h, viewerToken, adminToken
}

// newArtworkViewerTestHandlerWithMedia is like newArtworkViewerTestHandler but pre-registers
// a media object with the given ID.
func newArtworkViewerTestHandlerWithMedia(t *testing.T, mediaObjectID string) (http.Handler, string, string) {
	t.Helper()
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	repo := storage.NewMemoryRepository()
	mediaRepo := storage.NewMemoryMediaObjectRepository()
	if mediaObjectID != "" {
		if err := mediaRepo.SaveMediaObject(context.Background(), storage.MediaObject{
			ID:             mediaObjectID,
			BackendID:      "b1",
			ObjectKey:      "images/" + mediaObjectID,
			AssetKind:      "artwork",
			LifecycleState: "active",
		}); err != nil {
			t.Fatalf("SaveMediaObject: %v", err)
		}
	}
	catalogSvc := catalog.NewService(catalog.NewMemoryRepository())
	mediaSvc := storage.NewMediaObjectService(repo, mediaRepo)
	h := NewHandler(
		storage.NewService(repo),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
		WithCatalogService(catalogSvc),
		WithMediaObjectService(mediaSvc),
		WithHistoryService(history.NewService(history.NewMemoryRepository())),
		WithFavoritesService(favorites.NewService(favorites.NewMemoryRepository())),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "viewerart2", "passartwork2", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "viewerart2", "passartwork2")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "adminart2", "adminartwork2", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, err := authSvc.Login(context.Background(), "adminart2", "adminartwork2")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	return h, viewerToken, adminToken
}

// seedArtworkAlbumWithToken creates an artist + album via admin endpoints using the given
// session token, then optionally patches the artwork media object ID. Returns the album ID.
func seedArtworkAlbumWithToken(t *testing.T, h http.Handler, adminToken, artworkMediaObjectID string) string {
	t.Helper()
	aResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists",
		`{"name":"Test Artist"}`, "Bearer "+adminToken)
	if aResp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d %s", aResp.Code, aResp.Body.String())
	}
	var aBody map[string]any
	decodeResponse(t, aResp, &aBody)
	artistID := aBody["id"].(string)

	albumBody := fmt.Sprintf(`{"title":"Test Album","artistId":%q}`, artistID)
	alResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		albumBody, "Bearer "+adminToken)
	if alResp.Code != http.StatusCreated {
		t.Fatalf("create album: %d %s", alResp.Code, alResp.Body.String())
	}
	var album map[string]any
	decodeResponse(t, alResp, &album)
	albumID := album["id"].(string)

	if artworkMediaObjectID != "" {
		patchBody := fmt.Sprintf(`{"artworkMediaObjectId":%q}`, artworkMediaObjectID)
		pResp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/admin/catalog/albums/"+albumID,
			patchBody, "Bearer "+adminToken)
		if pResp.Code != http.StatusOK {
			t.Fatalf("patch album artwork: %d %s", pResp.Code, pResp.Body.String())
		}
	}
	return albumID
}

// TestGetAlbumArtwork_NoArtwork verifies that an album without artwork returns 404 no_artwork.
func TestGetAlbumArtwork_NoArtwork(t *testing.T) {
	h, viewerToken, adminToken := newArtworkViewerTestHandler(t)
	albumID := seedArtworkAlbumWithToken(t, h, adminToken, "")
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/albums/"+albumID+"/artwork",
		"", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusNotFound, "no_artwork")
}

// TestGetAlbumArtwork_AlbumNotFound verifies that a non-existent album returns 404.
func TestGetAlbumArtwork_AlbumNotFound(t *testing.T) {
	h, viewerToken, _ := newArtworkViewerTestHandler(t)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/albums/no-such-id/artwork",
		"", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

// TestGetAlbumArtwork_CatalogNotConfigured verifies that a missing catalog service returns 503.
// The no-catalog handler uses auth service (not static admin token) so viewer routes work.
func TestGetAlbumArtwork_CatalogNotConfigured(t *testing.T) {
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	repo := storage.NewMemoryRepository()
	h := NewHandler(
		storage.NewService(repo),
		WithAuthService(authSvc),
		// No catalog service
		WithHistoryService(history.NewService(history.NewMemoryRepository())),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()
	if _, err := authSvc.CreateUser(context.Background(), "viewernocat", "passnocat1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "viewernocat", "passnocat1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/albums/any-id/artwork",
		"", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusServiceUnavailable, "catalog_not_configured")
}

// TestGetAlbumArtwork_WithArtworkMediaObject verifies that an album with an artwork
// media object ID reaches the presign step. The in-memory backend has no presigned URL
// support, so the handler returns 503 presign_failed — confirming that GetAlbum →
// GetMediaObject → GeneratePresignedURL is wired correctly end-to-end.
func TestGetAlbumArtwork_WithArtworkMediaObject(t *testing.T) {
	const mediaObjectID = "artwork-mo-1"
	h, viewerToken, adminToken := newArtworkViewerTestHandlerWithMedia(t, mediaObjectID)
	albumID := seedArtworkAlbumWithToken(t, h, adminToken, mediaObjectID)
	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/albums/"+albumID+"/artwork",
		"", "Bearer "+viewerToken)
	// Memory backend has no presigned URL support → presign_failed 503.
	// This confirms the code reached the presign step (album found, media object found).
	assertAPIError(t, resp, http.StatusServiceUnavailable, "presign_failed")
}

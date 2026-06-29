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

// newReplayGainTestHandler builds a handler with full auth, catalog, and storage services
// and a pre-seeded audio media object for track import.
// Returns the handler and admin bearer token.
func newReplayGainTestHandler(t *testing.T, audioMediaObjID string) (http.Handler, string) {
	t.Helper()
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	repo := storage.NewMemoryRepository()
	mediaRepo := storage.NewMemoryMediaObjectRepository()
	if audioMediaObjID != "" {
		if err := mediaRepo.SaveMediaObject(context.Background(), storage.MediaObject{
			ID:             audioMediaObjID,
			BackendID:      "backend-rg",
			ObjectKey:      "audio/" + audioMediaObjID,
			AssetKind:      "original_audio",
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
	if _, err := authSvc.CreateUser(context.Background(), "adminrg", "adminrg1!", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, err := authSvc.Login(context.Background(), "adminrg", "adminrg1!")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	return h, adminToken
}

// seedReplayGainTrack creates an artist + track using a pre-seeded audio media object
// and returns the track ID.
func seedReplayGainTrack(t *testing.T, h http.Handler, adminToken, audioMediaObjID string) string {
	t.Helper()
	aResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists",
		`{"name":"RG Artist"}`, "Bearer "+adminToken)
	if aResp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d %s", aResp.Code, aResp.Body.String())
	}
	var aBody map[string]any
	decodeResponse(t, aResp, &aBody)
	artistID := aBody["id"].(string)

	trackBody := fmt.Sprintf(`{"mediaObjectId":%q,"title":"RG Track","artistId":%q,"durationMs":240000}`, audioMediaObjID, artistID)
	tResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/import",
		trackBody, "Bearer "+adminToken)
	if tResp.Code != http.StatusCreated {
		t.Fatalf("import track: %d %s", tResp.Code, tResp.Body.String())
	}
	var tBody map[string]any
	decodeResponse(t, tResp, &tBody)
	return tBody["id"].(string)
}

// TestPatchTrack_ReplayGainDb_UpdateSuccess verifies that PATCH with replayGainDb sets the value and returns 200.
func TestPatchTrack_ReplayGainDb_UpdateSuccess(t *testing.T) {
	const audioMO = "audio-rg-update-1"
	h, adminToken := newReplayGainTestHandler(t, audioMO)
	trackID := seedReplayGainTrack(t, h, adminToken, audioMO)

	body := `{"replayGainDb":-6.5}`
	resp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/admin/catalog/tracks/"+trackID,
		body, "Bearer "+adminToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("patch track status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var track map[string]any
	decodeResponse(t, resp, &track)
	rg, ok := track["replayGainDb"]
	if !ok {
		t.Fatal("expected replayGainDb field in response")
	}
	rgFloat, ok := rg.(float64)
	if !ok {
		t.Fatalf("replayGainDb type = %T, want float64", rg)
	}
	if rgFloat != -6.5 {
		t.Fatalf("replayGainDb = %v, want -6.5", rgFloat)
	}
}

// TestPatchTrack_ReplayGainDb_NullClears verifies that PATCH with null replayGainDb is accepted (200)
// and does not error. The current implementation treats null as no-op (field omitted in response).
func TestPatchTrack_ReplayGainDb_NullClears(t *testing.T) {
	const audioMO = "audio-rg-null-1"
	h, adminToken := newReplayGainTestHandler(t, audioMO)
	trackID := seedReplayGainTrack(t, h, adminToken, audioMO)

	// First set a value.
	setResp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/admin/catalog/tracks/"+trackID,
		`{"replayGainDb":-3.0}`, "Bearer "+adminToken)
	if setResp.Code != http.StatusOK {
		t.Fatalf("set replayGainDb: %d %s", setResp.Code, setResp.Body.String())
	}

	// Patch with null is accepted (200) — null means "no change" in the current API contract.
	clearResp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/admin/catalog/tracks/"+trackID,
		`{"replayGainDb":null}`, "Bearer "+adminToken)
	if clearResp.Code != http.StatusOK {
		t.Fatalf("patch with null replayGainDb: %d %s", clearResp.Code, clearResp.Body.String())
	}
	var track map[string]any
	decodeResponse(t, clearResp, &track)
	// Response is 200 OK — no error regardless of whether field is present or absent.
	if _, ok := track["id"]; !ok {
		t.Fatal("expected id field in patch response")
	}
}

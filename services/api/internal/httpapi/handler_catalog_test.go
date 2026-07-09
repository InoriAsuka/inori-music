package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"inori-music/services/api/internal/auth"
	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/storage"
)

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
	h := newNoCatalogTestHandler()
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
	h := newNoCatalogTestHandler()
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
	h := newNoCatalogTestHandler()
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
	h := newNoCatalogTestHandler()
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
	h := newNoCatalogTestHandler()
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
	h := newNoCatalogTestHandler()
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
	h := newNoCatalogTestHandler()
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
	h := newNoCatalogTestHandler()
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

func TestAlbumReleaseYearFilter(t *testing.T) {
	h := newTestHandler()

	// Seed artist
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"YearBand"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	// Seed albums with different release years
	for _, year := range []int{2010, 2015, 2020} {
		performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
			fmt.Sprintf(`{"title":"Album %d","artistId":%q,"releaseYear":%d}`, year, artistID, year))
	}

	// No filter → all 3
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums?limit=10", "")
	var all map[string]any
	decodeResponse(t, resp, &all)
	if all["pagination"].(map[string]any)["total"].(float64) != 3 {
		t.Fatalf("expected 3 albums, got %v", all["pagination"])
	}

	// releaseYearMin=2015 → 2015 and 2020
	resp2 := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums?releaseYearMin=2015&limit=10", "")
	var filtered map[string]any
	decodeResponse(t, resp2, &filtered)
	if filtered["pagination"].(map[string]any)["total"].(float64) != 2 {
		t.Fatalf("expected 2 albums with year>=2015, got %v", filtered["pagination"])
	}

	// releaseYearMax=2015 → 2010 and 2015
	resp3 := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums?releaseYearMax=2015&limit=10", "")
	var filtered2 map[string]any
	decodeResponse(t, resp3, &filtered2)
	if filtered2["pagination"].(map[string]any)["total"].(float64) != 2 {
		t.Fatalf("expected 2 albums with year<=2015, got %v", filtered2["pagination"])
	}

	// releaseYearMin > releaseYearMax → 400
	resp4 := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums?releaseYearMin=2020&releaseYearMax=2010", "")
	assertAPIError(t, resp4, http.StatusBadRequest, "validation_error")

	// Invalid value → 400
	resp5 := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/albums?releaseYearMin=notanumber", "")
	assertAPIError(t, resp5, http.StatusBadRequest, "validation_error")
}

// ---- catalog search type filter tests (Phase 139) ----

func TestCatalogSearchTypesFilter(t *testing.T) {
	h := newTestHandler()

	// Seed data: one artist, one album, one track
	artistResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/artists", `{"name":"SearchMe"}`)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)

	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		fmt.Sprintf(`{"title":"SearchMe Album","artistId":%q}`, artistID))
	performRequest(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"SearchMe Track","artistId":%q,"mediaObjectId":"mo-search-1"}`, artistID))

	// No types filter → all 3 kinds
	resp := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/search?q=SearchMe", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("search status = %d; body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	items := result["items"].([]any)
	if len(items) != 3 {
		t.Fatalf("expected 3 results (artist+album+track), got %d", len(items))
	}

	// types=track → only tracks
	resp2 := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/search?q=SearchMe&types=track", "")
	var result2 map[string]any
	decodeResponse(t, resp2, &result2)
	items2 := result2["items"].([]any)
	if len(items2) != 1 {
		t.Fatalf("types=track: expected 1 result, got %d", len(items2))
	}
	first := items2[0].(map[string]any)
	if first["kind"] != "track" {
		t.Errorf("types=track: kind = %v, want track", first["kind"])
	}

	// types=artist,album → 2 results
	resp3 := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/search?q=SearchMe&types=artist,album", "")
	var result3 map[string]any
	decodeResponse(t, resp3, &result3)
	items3 := result3["items"].([]any)
	if len(items3) != 2 {
		t.Fatalf("types=artist,album: expected 2 results, got %d", len(items3))
	}

	// Invalid type → 400
	resp4 := performRequest(t, h, http.MethodGet, "/api/v1/admin/catalog/search?q=SearchMe&types=playlist", "")
	assertAPIError(t, resp4, http.StatusBadRequest, "validation_error")
}

func TestViewerCatalogSearchTypesFilter(t *testing.T) {
	// newHistoryTestHandler has auth service wired so viewer tokens work
	h, viewerToken, adminToken := newHistoryTestHandler(t)

	artistResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists",
		`{"name":"ViewerSearch"}`, "Bearer "+adminToken)
	var artist map[string]any
	decodeResponse(t, artistResp, &artist)
	artistID := artist["id"].(string)
	performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/tracks",
		fmt.Sprintf(`{"title":"ViewerSearch Track","artistId":%q,"mediaObjectId":"mo-vs-1"}`, artistID), "Bearer "+adminToken)

	// Viewer path with types=track
	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/catalog/search?q=ViewerSearch&types=track", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("viewer search status = %d; body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	items := result["items"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["kind"] != "track" {
		t.Errorf("viewer search types=track: unexpected items = %v", items)
	}
}

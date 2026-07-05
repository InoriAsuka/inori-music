package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"inori-music/services/api/internal/auth"
	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/favorites"
	"inori-music/services/api/internal/history"
	"inori-music/services/api/internal/storage"
)

// newLyricsTestHandler builds a handler with auth, catalog, storage, and media objects,
// and a pre-seeded audio media object for track import. It also registers a local
// default storage backend rooted at tmpDir so that PutObject works for lyrics upload.
// Returns the handler, viewer bearer token, and admin bearer token.
func newLyricsTestHandler(t *testing.T, audioMediaObjID, tmpDir string) (http.Handler, string, string) {
	t.Helper()
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	repo := storage.NewMemoryRepository()
	mediaRepo := storage.NewMemoryMediaObjectRepository()
	if audioMediaObjID != "" {
		if err := mediaRepo.SaveMediaObject(context.Background(), storage.MediaObject{
			ID:             audioMediaObjID,
			BackendID:      "backend-lyrics",
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

	// Register a local storage backend as default so PutObject works.
	backendBody := fmt.Sprintf(
		`{"id":"backend-lyrics","type":"local","displayName":"Lyrics Test","enabled":true,"isDefault":true,"config":{"local":{"rootPath":%q}}}`,
		tmpDir,
	)
	regResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backendBody)
	if regResp.Code != http.StatusCreated {
		t.Fatalf("register backend: %d %s", regResp.Code, regResp.Body.String())
	}

	if _, err := authSvc.CreateUser(context.Background(), "viewerlyrics", "passlyrics1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "viewerlyrics", "passlyrics1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "adminlyrics", "adminlyrics1", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, err := authSvc.Login(context.Background(), "adminlyrics", "adminlyrics1")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	return h, viewerToken, adminToken
}

// seedLyricsTrack creates an artist + track via admin endpoints using a pre-seeded
// audio media object, and returns the track ID.
func seedLyricsTrack(t *testing.T, h http.Handler, adminToken, audioMediaObjID string) string {
	t.Helper()
	// Create artist.
	aResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists",
		`{"name":"Lyrics Artist"}`, "Bearer "+adminToken)
	if aResp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d %s", aResp.Code, aResp.Body.String())
	}
	var aBody map[string]any
	decodeResponse(t, aResp, &aBody)
	artistID := aBody["id"].(string)

	// Import track with the pre-seeded audio media object.
	trackBody := fmt.Sprintf(`{"mediaObjectId":%q,"title":"Lyrics Track","artistId":%q,"durationMs":180000}`, audioMediaObjID, artistID)
	tResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/import",
		trackBody, "Bearer "+adminToken)
	if tResp.Code != http.StatusCreated {
		t.Fatalf("import track: %d %s", tResp.Code, tResp.Body.String())
	}
	var tBody map[string]any
	decodeResponse(t, tResp, &tBody)
	return tBody["id"].(string)
}

// buildLyricsMultipart builds a multipart/form-data body with a lyrics file.
func buildLyricsMultipart(t *testing.T, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	w.Close()
	return &buf, w.FormDataContentType()
}

// buildLyricsMultipartWithTranslation builds a multipart/form-data body with both a
// lyrics file and an optional translation file field.
func buildLyricsMultipartWithTranslation(t *testing.T, filename, content, translationFilename, translationContent string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	tw, err := w.CreateFormFile("translation", translationFilename)
	if err != nil {
		t.Fatalf("create translation form file: %v", err)
	}
	if _, err := tw.Write([]byte(translationContent)); err != nil {
		t.Fatalf("write translation form file: %v", err)
	}
	w.Close()
	return &buf, w.FormDataContentType()
}

// performMultipartRequest performs a request with a multipart body.
func performMultipartRequest(t *testing.T, h http.Handler, method, path string, body *bytes.Buffer, contentType, authHeader string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", contentType)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// TestUploadTrackLyrics_Success verifies that uploading valid LRC lyrics returns 201.
func TestUploadTrackLyrics_Success(t *testing.T) {
	const audioMO = "audio-lrc-success-1"
	h, _, adminToken := newLyricsTestHandler(t, audioMO, t.TempDir())
	trackID := seedLyricsTrack(t, h, adminToken, audioMO)

	lrc := "[00:01.00]Hello world\n[00:05.00]Second line\n"
	body, ct := buildLyricsMultipart(t, "lyrics.lrc", lrc)
	resp := performMultipartRequest(t, h, http.MethodPost, "/api/v1/catalog/tracks/"+trackID+"/lyrics", body, ct, "Bearer "+adminToken)
	if resp.Code != http.StatusCreated {
		t.Fatalf("upload lyrics status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["mediaObjectId"] == nil {
		t.Fatal("expected mediaObjectId in response")
	}
}

// TestUploadTrackLyrics_BadFormat verifies that uploading an unrecognized format returns 400.
func TestUploadTrackLyrics_BadFormat(t *testing.T) {
	const audioMO = "audio-lrc-badformat-1"
	h, _, adminToken := newLyricsTestHandler(t, audioMO, t.TempDir())
	trackID := seedLyricsTrack(t, h, adminToken, audioMO)

	body, ct := buildLyricsMultipart(t, "bad.txt", "this is not lrc or srt format !!!@@@")
	resp := performMultipartRequest(t, h, http.MethodPost, "/api/v1/catalog/tracks/"+trackID+"/lyrics", body, ct, "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_format")
}

// TestUploadTrackLyrics_TrackNotFound verifies that uploading to a non-existent track returns 404.
func TestUploadTrackLyrics_TrackNotFound(t *testing.T) {
	const audioMO = "audio-lrc-notfound-1"
	h, _, adminToken := newLyricsTestHandler(t, audioMO, t.TempDir())

	lrc := "[00:01.00]Test line\n"
	body, ct := buildLyricsMultipart(t, "lyrics.lrc", lrc)
	resp := performMultipartRequest(t, h, http.MethodPost, "/api/v1/catalog/tracks/no-such-track/lyrics", body, ct, "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusNotFound, "not_found")
}

// TestGetTrackLyrics_Success verifies that GET lyrics after upload returns 200 with content.
func TestGetTrackLyrics_Success(t *testing.T) {
	const audioMO = "audio-lrc-get-1"
	h, viewerToken, adminToken := newLyricsTestHandler(t, audioMO, t.TempDir())
	trackID := seedLyricsTrack(t, h, adminToken, audioMO)

	// Upload first.
	lrc := "[00:01.00]Hello world\n[00:05.00]Second line\n"
	body, ct := buildLyricsMultipart(t, "lyrics.lrc", lrc)
	upResp := performMultipartRequest(t, h, http.MethodPost, "/api/v1/catalog/tracks/"+trackID+"/lyrics", body, ct, "Bearer "+adminToken)
	if upResp.Code != http.StatusCreated {
		t.Fatalf("upload lyrics: %d %s", upResp.Code, upResp.Body.String())
	}

	// Get lyrics.
	getResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/"+trackID+"/lyrics",
		"", "Bearer "+viewerToken)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get lyrics status = %d, body = %s", getResp.Code, getResp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, getResp, &result)
	if result["format"] == nil || result["content"] == nil {
		t.Fatalf("expected format and content fields, got %v", result)
	}
	if result["format"] != "lrc" {
		t.Fatalf("format = %v, want lrc", result["format"])
	}
}

// TestGetTrackLyrics_NoLyrics verifies that GET on a track without lyrics returns 404 no_lyrics.
func TestGetTrackLyrics_NoLyrics(t *testing.T) {
	const audioMO = "audio-lrc-nolyrics-1"
	h, viewerToken, adminToken := newLyricsTestHandler(t, audioMO, t.TempDir())
	trackID := seedLyricsTrack(t, h, adminToken, audioMO)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/"+trackID+"/lyrics",
		"", "Bearer "+viewerToken)
	assertAPIError(t, resp, http.StatusNotFound, "no_lyrics")
}

// TestDeleteTrackLyrics_Success verifies that DELETE lyrics returns 204.
func TestDeleteTrackLyrics_Success(t *testing.T) {
	const audioMO = "audio-lrc-delete-1"
	h, _, adminToken := newLyricsTestHandler(t, audioMO, t.TempDir())
	trackID := seedLyricsTrack(t, h, adminToken, audioMO)

	// Upload first so delete has something to clear.
	lrc := "[00:01.00]Delete me\n"
	body, ct := buildLyricsMultipart(t, "lyrics.lrc", lrc)
	upResp := performMultipartRequest(t, h, http.MethodPost, "/api/v1/catalog/tracks/"+trackID+"/lyrics", body, ct, "Bearer "+adminToken)
	if upResp.Code != http.StatusCreated {
		t.Fatalf("upload lyrics: %d %s", upResp.Code, upResp.Body.String())
	}

	// Delete.
	delResp := performRequestWithAuthHeader(t, h, http.MethodDelete, "/api/v1/catalog/tracks/"+trackID+"/lyrics",
		"", "Bearer "+adminToken)
	if delResp.Code != http.StatusNoContent {
		t.Fatalf("delete lyrics status = %d, body = %s", delResp.Code, delResp.Body.String())
	}
}

// TestUploadTrackLyrics_WithTranslation_Success verifies that uploading lyrics with a
// translation field returns 201 with both mediaObjectId and translationMediaObjectId.
func TestUploadTrackLyrics_WithTranslation_Success(t *testing.T) {
	const audioMO = "audio-lrc-translation-1"
	h, _, adminToken := newLyricsTestHandler(t, audioMO, t.TempDir())
	trackID := seedLyricsTrack(t, h, adminToken, audioMO)

	lrc := "[00:01.00]Hello world\n[00:05.00]Second line\n"
	translation := "你好世界\n第二行\n"
	body, ct := buildLyricsMultipartWithTranslation(t, "lyrics.lrc", lrc, "lyrics.translation.lrc", translation)
	resp := performMultipartRequest(t, h, http.MethodPost, "/api/v1/catalog/tracks/"+trackID+"/lyrics", body, ct, "Bearer "+adminToken)
	if resp.Code != http.StatusCreated {
		t.Fatalf("upload lyrics status = %d, body = %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, resp, &result)
	if result["mediaObjectId"] == nil {
		t.Fatal("expected mediaObjectId in response")
	}
	if result["translationMediaObjectId"] == nil {
		t.Fatal("expected translationMediaObjectId in response")
	}
}

// TestUploadTrackLyrics_TranslationBadFormat verifies that a non-UTF-8 translation field
// returns 400 invalid_format.
func TestUploadTrackLyrics_TranslationBadFormat(t *testing.T) {
	const audioMO = "audio-lrc-translation-badformat-1"
	h, _, adminToken := newLyricsTestHandler(t, audioMO, t.TempDir())
	trackID := seedLyricsTrack(t, h, adminToken, audioMO)

	lrc := "[00:01.00]Hello world\n"
	invalidUTF8 := string([]byte{0xff, 0xfe, 0xfd})
	body, ct := buildLyricsMultipartWithTranslation(t, "lyrics.lrc", lrc, "lyrics.translation.lrc", invalidUTF8)
	resp := performMultipartRequest(t, h, http.MethodPost, "/api/v1/catalog/tracks/"+trackID+"/lyrics", body, ct, "Bearer "+adminToken)
	assertAPIError(t, resp, http.StatusBadRequest, "invalid_format")
}

// TestGetTrackLyrics_WithTranslation verifies that GET returns translation and source
// fields when a translation was uploaded.
func TestGetTrackLyrics_WithTranslation(t *testing.T) {
	const audioMO = "audio-lrc-get-translation-1"
	h, viewerToken, adminToken := newLyricsTestHandler(t, audioMO, t.TempDir())
	trackID := seedLyricsTrack(t, h, adminToken, audioMO)

	lrc := "[00:01.00]Hello world\n[00:05.00]Second line\n"
	translation := "你好世界\n第二行\n"
	body, ct := buildLyricsMultipartWithTranslation(t, "lyrics.lrc", lrc, "lyrics.translation.lrc", translation)
	upResp := performMultipartRequest(t, h, http.MethodPost, "/api/v1/catalog/tracks/"+trackID+"/lyrics", body, ct, "Bearer "+adminToken)
	if upResp.Code != http.StatusCreated {
		t.Fatalf("upload lyrics: %d %s", upResp.Code, upResp.Body.String())
	}

	getResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/"+trackID+"/lyrics",
		"", "Bearer "+viewerToken)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get lyrics status = %d, body = %s", getResp.Code, getResp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, getResp, &result)
	if result["translation"] != translation {
		t.Fatalf("translation = %v, want %q", result["translation"], translation)
	}
	if result["source"] != "manual" {
		t.Fatalf("source = %v, want manual", result["source"])
	}
	if result["translationMediaObjectId"] == nil {
		t.Fatal("expected translationMediaObjectId in response")
	}
}

// TestGetTrackLyrics_NoTranslation verifies that GET omits the translation field when
// no translation was uploaded, while still reporting source.
func TestGetTrackLyrics_NoTranslation(t *testing.T) {
	const audioMO = "audio-lrc-get-notranslation-1"
	h, viewerToken, adminToken := newLyricsTestHandler(t, audioMO, t.TempDir())
	trackID := seedLyricsTrack(t, h, adminToken, audioMO)

	lrc := "[00:01.00]Hello world\n"
	body, ct := buildLyricsMultipart(t, "lyrics.lrc", lrc)
	upResp := performMultipartRequest(t, h, http.MethodPost, "/api/v1/catalog/tracks/"+trackID+"/lyrics", body, ct, "Bearer "+adminToken)
	if upResp.Code != http.StatusCreated {
		t.Fatalf("upload lyrics: %d %s", upResp.Code, upResp.Body.String())
	}

	getResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/"+trackID+"/lyrics",
		"", "Bearer "+viewerToken)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get lyrics status = %d, body = %s", getResp.Code, getResp.Body.String())
	}
	var result map[string]any
	decodeResponse(t, getResp, &result)
	if _, ok := result["translation"]; ok {
		t.Fatalf("expected no translation field, got %v", result["translation"])
	}
	if result["source"] != "manual" {
		t.Fatalf("source = %v, want manual", result["source"])
	}
}

// TestDeleteTrackLyrics_CascadesTranslation verifies that DELETE clears the translation
// reference alongside the primary lyrics reference.
func TestDeleteTrackLyrics_CascadesTranslation(t *testing.T) {
	const audioMO = "audio-lrc-delete-translation-1"
	h, viewerToken, adminToken := newLyricsTestHandler(t, audioMO, t.TempDir())
	trackID := seedLyricsTrack(t, h, adminToken, audioMO)

	lrc := "[00:01.00]Delete me\n"
	translation := "删除我\n"
	body, ct := buildLyricsMultipartWithTranslation(t, "lyrics.lrc", lrc, "lyrics.translation.lrc", translation)
	upResp := performMultipartRequest(t, h, http.MethodPost, "/api/v1/catalog/tracks/"+trackID+"/lyrics", body, ct, "Bearer "+adminToken)
	if upResp.Code != http.StatusCreated {
		t.Fatalf("upload lyrics: %d %s", upResp.Code, upResp.Body.String())
	}

	delResp := performRequestWithAuthHeader(t, h, http.MethodDelete, "/api/v1/catalog/tracks/"+trackID+"/lyrics",
		"", "Bearer "+adminToken)
	if delResp.Code != http.StatusNoContent {
		t.Fatalf("delete lyrics status = %d, body = %s", delResp.Code, delResp.Body.String())
	}

	getResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/tracks/"+trackID+"/lyrics",
		"", "Bearer "+viewerToken)
	assertAPIError(t, getResp, http.StatusNotFound, "no_lyrics")
}

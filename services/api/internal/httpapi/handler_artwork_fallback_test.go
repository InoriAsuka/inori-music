package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"inori-music/services/api/internal/auth"
	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/favorites"
	"inori-music/services/api/internal/history"
	"inori-music/services/api/internal/storage"
	"inori-music/services/api/internal/streamsign"
)

// ---------------------------------------------------------------------------
// Minimal hand-built FLAC fixtures — see internal/artwork/artwork_test.go for
// the full explanation of why dhowden/tag accepts these (no STREAMINFO block
// required, magic + a single terminal metadata block is enough). Duplicated
// here (rather than exported from internal/artwork) because it is test-only
// fixture code, not part of that package's public surface.
// ---------------------------------------------------------------------------

func writeBE32(buf *bytes.Buffer, v uint32) {
	buf.WriteByte(byte(v >> 24))
	buf.WriteByte(byte(v >> 16))
	buf.WriteByte(byte(v >> 8))
	buf.WriteByte(byte(v))
}

func buildFLACWithPicture(mimeType string, data []byte) []byte {
	var body bytes.Buffer
	writeBE32(&body, 3) // picture type 3 = "Cover (front)"
	writeBE32(&body, uint32(len(mimeType)))
	body.WriteString(mimeType)
	writeBE32(&body, 0) // description length
	writeBE32(&body, 0) // width
	writeBE32(&body, 0) // height
	writeBE32(&body, 0) // color depth
	writeBE32(&body, 0) // colors used
	writeBE32(&body, uint32(len(data)))
	body.Write(data)

	var file bytes.Buffer
	file.WriteString("fLaC")
	file.WriteByte(0x80 | 6) // last-block flag set, block type 6 = PICTURE
	blockLen := body.Len()
	file.WriteByte(byte(blockLen >> 16))
	file.WriteByte(byte(blockLen >> 8))
	file.WriteByte(byte(blockLen))
	file.Write(body.Bytes())
	return file.Bytes()
}

func buildFLACWithoutPicture() []byte {
	var file bytes.Buffer
	file.WriteString("fLaC")
	file.WriteByte(0x80 | 1) // last-block flag set, block type 1 = padding
	file.WriteByte(0)
	file.WriteByte(0)
	file.WriteByte(4)
	file.Write([]byte{0, 0, 0, 0})
	return file.Bytes()
}

// signArtworkURL replicates streamsign's HMAC-SHA256 computation to produce
// a validly-signed (or, with an exp in the past, validly-signed-but-expired)
// query string, independent of Signer.Sign's fixed "now + ttl" expiry. This
// is the only way to exercise Verify's expiry check specifically (as opposed
// to its signature-mismatch check).
func signArtworkURL(key, albumID string, exp int64) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(albumArtworkSignaturePayload(albumID)))
	mac.Write([]byte(strconv.FormatInt(exp, 10)))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("exp=%d&sig=%s", exp, sig)
}

// ---------------------------------------------------------------------------
// v5.39.0 guard tests: cover art resolved on demand from the audio file for
// backends that cannot presign (local/NFS/SMB) — see requirement.md v5.39.0
// and .plan/20260814-079-v5.39.0-album-artwork.md.
//
// Before this version, getAlbumArtwork had exactly one way to produce a URL
// (handler.storage.GeneratePresignedURL), and InferCapabilities never sets
// PresignedURLs for BackendTypeLocal — so any album on a local backend
// 503'd, even with ArtworkMediaObjectID correctly linked. That is production's
// exact shape (backendType: "local"). TestGetAlbumArtwork_LocalBackend_
// EmbeddedPicture_Success below is that scenario end to end.
// ---------------------------------------------------------------------------

// newLocalArtworkTestHandler builds a handler wired with session auth,
// catalog, storage, media objects, and a stream signer, plus a real "local"
// storage backend (backendType: "local", matching production) rooted at a
// fresh temp directory. Returns the handler, the underlying MediaObject
// repository (so tests can register media objects pointing at real files
// without going through the admin HTTP surface), the backend's root
// directory, a viewer token, and an admin token.
func newLocalArtworkTestHandler(t *testing.T) (h http.Handler, mediaRepo *storage.MemoryMediaObjectRepository, root, viewerToken, adminToken string) {
	t.Helper()
	root = t.TempDir()
	authSvc := auth.NewService(newMemAuthUserRepo(), newMemAuthSessionRepo(), auth.ServiceConfig{SessionTTL: time.Hour})
	repo := storage.NewMemoryRepository()
	mediaRepo = storage.NewMemoryMediaObjectRepository()
	catalogSvc := catalog.NewService(catalog.NewMemoryRepository())
	mediaSvc := storage.NewMediaObjectService(repo, mediaRepo)
	h = NewHandler(
		storage.NewService(repo),
		WithAuthService(authSvc),
		WithAdminToken(testAdminToken),
		WithCatalogService(catalogSvc),
		WithMediaObjectService(mediaSvc),
		WithHistoryService(history.NewService(history.NewMemoryRepository())),
		WithFavoritesService(favorites.NewService(favorites.NewMemoryRepository())),
		WithStreamSigner(streamsign.NewSigner("test-artwork-signing-key")),
		WithServiceInfo(ServiceInfo{Name: "inori-api", Version: "test", Commit: "c", BuildTime: "t"}),
	).Routes()

	// Register a real "local" backend — the exact deployment shape
	// (backendId "local-music", backendType "local") that 503'd before this fix.
	backendBody := fmt.Sprintf(
		`{"id":"local-artwork-test","type":"local","displayName":"Artwork Test","enabled":true,"isDefault":true,"config":{"local":{"rootPath":%q}}}`,
		root,
	)
	regResp := performRequest(t, h, http.MethodPost, "/api/v1/admin/storage/backends", backendBody)
	if regResp.Code != http.StatusCreated {
		t.Fatalf("register backend: %d %s", regResp.Code, regResp.Body.String())
	}

	if _, err := authSvc.CreateUser(context.Background(), "viewerartfb", "passartfb1", auth.RoleViewer); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	viewerToken, _, err := authSvc.Login(context.Background(), "viewerartfb", "passartfb1")
	if err != nil {
		t.Fatalf("viewer login: %v", err)
	}
	if _, err := authSvc.CreateUser(context.Background(), "adminartfb", "adminartfb1", auth.RoleAdmin); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	adminToken, _, err = authSvc.Login(context.Background(), "adminartfb", "adminartfb1")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	return h, mediaRepo, root, viewerToken, adminToken
}

// seedArtworkTrack writes fileBytes to root/relPath, registers it as a media
// object on the "local-artwork-test" backend, and creates an artist + album +
// track wired to it (discNumber=1, trackNumber=1) via the admin API. When
// linkAlbumArtwork is true, the album's ArtworkMediaObjectID is also set to
// this same media object — reproducing "artwork correctly linked" (Break 2 in
// requirement.md v5.39.0) so the presign path is actually attempted and fails
// on this local backend, rather than short-circuiting on an empty ID before
// ever reaching it. Returns the album ID and track ID.
func seedArtworkTrack(t *testing.T, h http.Handler, mediaRepo *storage.MemoryMediaObjectRepository, adminToken, root, relPath string, fileBytes []byte, linkAlbumArtwork bool) (albumID, trackID string) {
	t.Helper()
	fullPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, fileBytes, 0o644); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}

	mediaObjectID := "mo-" + relPath
	if err := mediaRepo.SaveMediaObject(context.Background(), storage.MediaObject{
		ID:             mediaObjectID,
		BackendID:      "local-artwork-test",
		ObjectKey:      relPath,
		AssetKind:      string(storage.AssetKindOriginalAudio),
		LifecycleState: string(storage.LifecycleStateActive),
		MIMEType:       "audio/flac",
		SizeBytes:      int64(len(fileBytes)),
	}); err != nil {
		t.Fatalf("SaveMediaObject: %v", err)
	}

	aResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/artists",
		`{"name":"Artwork Fallback Artist"}`, "Bearer "+adminToken)
	if aResp.Code != http.StatusCreated {
		t.Fatalf("create artist: %d %s", aResp.Code, aResp.Body.String())
	}
	var aBody map[string]any
	decodeResponse(t, aResp, &aBody)
	artistID := aBody["id"].(string)

	albumBody := fmt.Sprintf(`{"title":"Artwork Fallback Album","artistId":%q}`, artistID)
	alResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/albums",
		albumBody, "Bearer "+adminToken)
	if alResp.Code != http.StatusCreated {
		t.Fatalf("create album: %d %s", alResp.Code, alResp.Body.String())
	}
	var album map[string]any
	decodeResponse(t, alResp, &album)
	albumID = album["id"].(string)

	trackBody := fmt.Sprintf(`{"mediaObjectId":%q,"title":"Artwork Fallback Track","artistId":%q,"albumId":%q,"discNumber":1,"trackNumber":1,"durationMs":180000}`,
		mediaObjectID, artistID, albumID)
	tResp := performRequestWithAuthHeader(t, h, http.MethodPost, "/api/v1/admin/catalog/import",
		trackBody, "Bearer "+adminToken)
	if tResp.Code != http.StatusCreated {
		t.Fatalf("import track: %d %s", tResp.Code, tResp.Body.String())
	}
	var tBody map[string]any
	decodeResponse(t, tResp, &tBody)
	trackID = tBody["id"].(string)

	if linkAlbumArtwork {
		patchBody := fmt.Sprintf(`{"artworkMediaObjectId":%q}`, mediaObjectID)
		pResp := performRequestWithAuthHeader(t, h, http.MethodPatch, "/api/v1/admin/catalog/albums/"+albumID,
			patchBody, "Bearer "+adminToken)
		if pResp.Code != http.StatusOK {
			t.Fatalf("patch album artwork: %d %s", pResp.Code, pResp.Body.String())
		}
	}
	return albumID, trackID
}

// extractSignedURL decodes {"url": "...", "expiresIn": N} from an artwork
// metadata response.
func extractSignedURL(t *testing.T, resp *httptest.ResponseRecorder) string {
	t.Helper()
	var body albumArtworkResponse
	decodeResponse(t, resp, &body)
	if body.URL == "" {
		t.Fatalf("expected non-empty url in response, body = %s", resp.Body.String())
	}
	return body.URL
}

// TestGetAlbumArtwork_LocalBackend_EmbeddedPicture_Success is the single
// most important guard for this version: getAlbumArtwork must succeed for an
// album on a local backend. Before the fix, this exact setup — a "local"
// backend and an album with ArtworkMediaObjectID linked — returned 503
// presign_failed, because GeneratePresignedURL was the only path the handler
// had and InferCapabilities never grants BackendTypeLocal PresignedURLs.
func TestGetAlbumArtwork_LocalBackend_EmbeddedPicture_Success(t *testing.T) {
	h, mediaRepo, root, viewerToken, adminToken := newLocalArtworkTestHandler(t)
	coverBytes := []byte("embedded-cover-bytes-for-local-backend-guard-test")
	flacBytes := buildFLACWithPicture("image/jpeg", coverBytes)
	albumID, _ := seedArtworkTrack(t, h, mediaRepo, adminToken, root, "albums/a1/track01.flac", flacBytes, true)

	metaResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/albums/"+albumID+"/artwork",
		"", "Bearer "+viewerToken)
	if metaResp.Code != http.StatusOK {
		t.Fatalf("GET artwork metadata status = %d, want 200, body = %s", metaResp.Code, metaResp.Body.String())
	}
	url := extractSignedURL(t, metaResp)
	if strings.HasPrefix(url, "http") {
		t.Fatalf("url = %q, want a signed relative path (like streamUrl), not an absolute URL", url)
	}
	if !strings.Contains(url, "exp=") || !strings.Contains(url, "sig=") {
		t.Fatalf("url = %q, want an HMAC-signed query string (exp=&sig=)", url)
	}

	bytesResp := performRequestWithAuthHeader(t, h, http.MethodGet, url, "", "Bearer "+viewerToken)
	if bytesResp.Code != http.StatusOK {
		t.Fatalf("GET artwork file status = %d, want 200, body = %s", bytesResp.Code, bytesResp.Body.String())
	}
	if !bytes.Equal(bytesResp.Body.Bytes(), coverBytes) {
		t.Fatalf("artwork bytes = %q, want %q", bytesResp.Body.Bytes(), coverBytes)
	}
	if ct := bytesResp.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("Content-Type = %q, want image/jpeg", ct)
	}
	if bytesResp.Header().Get("ETag") == "" {
		t.Error("expected a non-empty ETag header")
	}
}

// TestGetAlbumArtwork_LocalBackend_SiblingFileFallback_Success verifies the
// second-priority resolution path: the audio file itself has no embedded
// picture, but the same directory has a sibling "cover.jpg".
func TestGetAlbumArtwork_LocalBackend_SiblingFileFallback_Success(t *testing.T) {
	h, mediaRepo, root, viewerToken, adminToken := newLocalArtworkTestHandler(t)
	flacBytes := buildFLACWithoutPicture()
	albumID, _ := seedArtworkTrack(t, h, mediaRepo, adminToken, root, "albums/a2/track01.flac", flacBytes, false)

	siblingBytes := []byte("sibling-folder-jpg-bytes")
	if err := os.WriteFile(filepath.Join(root, "albums/a2/cover.jpg"), siblingBytes, 0o644); err != nil {
		t.Fatalf("write sibling cover: %v", err)
	}

	metaResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/albums/"+albumID+"/artwork",
		"", "Bearer "+viewerToken)
	if metaResp.Code != http.StatusOK {
		t.Fatalf("GET artwork metadata status = %d, want 200, body = %s", metaResp.Code, metaResp.Body.String())
	}
	url := extractSignedURL(t, metaResp)

	bytesResp := performRequestWithAuthHeader(t, h, http.MethodGet, url, "", "Bearer "+viewerToken)
	if bytesResp.Code != http.StatusOK {
		t.Fatalf("GET artwork file status = %d, want 200, body = %s", bytesResp.Code, bytesResp.Body.String())
	}
	if !bytes.Equal(bytesResp.Body.Bytes(), siblingBytes) {
		t.Fatalf("artwork bytes = %q, want %q (sibling file)", bytesResp.Body.Bytes(), siblingBytes)
	}
	if ct := bytesResp.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("Content-Type = %q, want image/jpeg", ct)
	}
}

// TestGetAlbumArtwork_LocalBackend_NoArtworkAnywhere_404 verifies the honest
// failure mode: no embedded picture, no sibling image file — 404 no_artwork,
// with no server-side placeholder synthesised.
func TestGetAlbumArtwork_LocalBackend_NoArtworkAnywhere_404(t *testing.T) {
	h, mediaRepo, root, viewerToken, adminToken := newLocalArtworkTestHandler(t)
	flacBytes := buildFLACWithoutPicture()
	albumID, _ := seedArtworkTrack(t, h, mediaRepo, adminToken, root, "albums/a3/track01.flac", flacBytes, false)
	// No sibling file written for this album's directory.

	metaResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/albums/"+albumID+"/artwork",
		"", "Bearer "+viewerToken)
	assertAPIError(t, metaResp, http.StatusNotFound, "no_artwork")

	// The bytes endpoint must independently agree — same resolution, same result.
	fileResp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/catalog/albums/"+albumID+"/artwork/file", "", "Bearer "+viewerToken)
	assertAPIError(t, fileResp, http.StatusNotFound, "no_artwork")
}

// TestGetAlbumArtworkFile_SignedURL_TamperedSignatureRejected verifies that
// the artwork-file endpoint validates its HMAC signature, mirroring streamTrack.
func TestGetAlbumArtworkFile_SignedURL_TamperedSignatureRejected(t *testing.T) {
	h, mediaRepo, root, viewerToken, adminToken := newLocalArtworkTestHandler(t)
	flacBytes := buildFLACWithPicture("image/png", []byte("tamper-test-bytes"))
	albumID, _ := seedArtworkTrack(t, h, mediaRepo, adminToken, root, "albums/a4/track01.flac", flacBytes, true)

	metaResp := performRequestWithAuthHeader(t, h, http.MethodGet, "/api/v1/catalog/albums/"+albumID+"/artwork",
		"", "Bearer "+viewerToken)
	url := extractSignedURL(t, metaResp)
	tampered := url + "tampered"

	resp := performRequestWithoutAuth(t, h, http.MethodGet, tampered, "")
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

// TestGetAlbumArtworkFile_SignedURL_ExpiredRejected verifies that a
// validly-signed but expired URL is rejected — distinct from an invalid
// signature, this exercises Verify's time.Now().Unix() > exp check.
func TestGetAlbumArtworkFile_SignedURL_ExpiredRejected(t *testing.T) {
	h, mediaRepo, root, _, adminToken := newLocalArtworkTestHandler(t)
	flacBytes := buildFLACWithPicture("image/png", []byte("expiry-test-bytes"))
	albumID, _ := seedArtworkTrack(t, h, mediaRepo, adminToken, root, "albums/a5/track01.flac", flacBytes, true)

	expiredQuery := signArtworkURL("test-artwork-signing-key", albumID, time.Now().Add(-1*time.Hour).Unix())
	resp := performRequestWithoutAuth(t, h, http.MethodGet,
		"/api/v1/catalog/albums/"+albumID+"/artwork/file?"+expiredQuery, "")
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

// TestGetAlbumArtworkFile_MissingCredentials_Unauthorized verifies that a
// request with neither a Bearer token nor a signed query string is rejected.
func TestGetAlbumArtworkFile_MissingCredentials_Unauthorized(t *testing.T) {
	h, mediaRepo, root, _, adminToken := newLocalArtworkTestHandler(t)
	flacBytes := buildFLACWithPicture("image/png", []byte("no-credentials-bytes"))
	albumID, _ := seedArtworkTrack(t, h, mediaRepo, adminToken, root, "albums/a6/track01.flac", flacBytes, true)

	resp := performRequestWithoutAuth(t, h, http.MethodGet, "/api/v1/catalog/albums/"+albumID+"/artwork/file", "")
	assertAPIError(t, resp, http.StatusUnauthorized, "unauthorized")
}

// TestGetAlbumArtworkFile_BearerToken_Success verifies the dual-auth path:
// a viewer Bearer token works even without a signed query string, exactly
// like streamTrack.
func TestGetAlbumArtworkFile_BearerToken_Success(t *testing.T) {
	h, mediaRepo, root, viewerToken, adminToken := newLocalArtworkTestHandler(t)
	coverBytes := []byte("bearer-auth-cover-bytes")
	flacBytes := buildFLACWithPicture("image/jpeg", coverBytes)
	albumID, _ := seedArtworkTrack(t, h, mediaRepo, adminToken, root, "albums/a7/track01.flac", flacBytes, true)

	resp := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/catalog/albums/"+albumID+"/artwork/file", "", "Bearer "+viewerToken)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %s", resp.Code, resp.Body.String())
	}
	if !bytes.Equal(resp.Body.Bytes(), coverBytes) {
		t.Fatalf("bytes = %q, want %q", resp.Body.Bytes(), coverBytes)
	}
}

// TestGetAlbumArtworkFile_ConditionalGet_304 verifies that the ETag set on
// the bytes endpoint drives a real 304 on a matching If-None-Match, proving
// the Cache-Control/ETag wiring is functional and not just present.
func TestGetAlbumArtworkFile_ConditionalGet_304(t *testing.T) {
	h, mediaRepo, root, viewerToken, adminToken := newLocalArtworkTestHandler(t)
	flacBytes := buildFLACWithPicture("image/jpeg", []byte("conditional-get-bytes"))
	albumID, _ := seedArtworkTrack(t, h, mediaRepo, adminToken, root, "albums/a8/track01.flac", flacBytes, true)

	first := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/catalog/albums/"+albumID+"/artwork/file", "", "Bearer "+viewerToken)
	if first.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", first.Code)
	}
	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag on first response")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/albums/"+albumID+"/artwork/file", nil)
	req.Header.Set("Authorization", "Bearer "+viewerToken)
	req.Header.Set("If-None-Match", etag)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified {
		t.Fatalf("conditional GET status = %d, want 304", rec.Code)
	}
}

// TestGetAlbumArtwork_CachesAcrossRequests demonstrates that the resolved
// image is served from the in-process cache rather than re-read from disk on
// every request: the underlying file is overwritten with different embedded
// artwork between two requests inside the cache TTL, and the second request
// must still return the *first* image's bytes.
func TestGetAlbumArtwork_CachesAcrossRequests(t *testing.T) {
	h, mediaRepo, root, viewerToken, adminToken := newLocalArtworkTestHandler(t)
	originalBytes := []byte("original-cached-cover")
	relPath := "albums/a9/track01.flac"
	albumID, _ := seedArtworkTrack(t, h, mediaRepo, adminToken, root, relPath, buildFLACWithPicture("image/jpeg", originalBytes), true)

	first := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/catalog/albums/"+albumID+"/artwork/file", "", "Bearer "+viewerToken)
	if first.Code != http.StatusOK || !bytes.Equal(first.Body.Bytes(), originalBytes) {
		t.Fatalf("first request = %d %q, want 200 %q", first.Code, first.Body.Bytes(), originalBytes)
	}

	// Overwrite the same file on disk with different embedded artwork.
	mutatedBytes := []byte("mutated-cover-should-not-be-served-yet")
	fullPath := filepath.Join(root, relPath)
	if err := os.WriteFile(fullPath, buildFLACWithPicture("image/png", mutatedBytes), 0o644); err != nil {
		t.Fatalf("overwrite fixture: %v", err)
	}

	second := performRequestWithAuthHeader(t, h, http.MethodGet,
		"/api/v1/catalog/albums/"+albumID+"/artwork/file", "", "Bearer "+viewerToken)
	if second.Code != http.StatusOK {
		t.Fatalf("second request status = %d, want 200", second.Code)
	}
	if !bytes.Equal(second.Body.Bytes(), originalBytes) {
		t.Fatalf("second request bytes = %q, want %q (stale/cached) — the file header was re-parsed instead of served from cache",
			second.Body.Bytes(), originalBytes)
	}
}

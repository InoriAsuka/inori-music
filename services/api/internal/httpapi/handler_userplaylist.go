package httpapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"inori-music/services/api/internal/userplaylist"
)

// ---- user playlist handlers ----

func (handler *Handler) requireUserPlaylistService(w http.ResponseWriter) bool {
	if handler.userPlaylistService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "user_playlists_not_configured", "user playlist service is not configured")
		return false
	}
	return true
}

type createUserPlaylistRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	SourceCatalogID string `json:"sourceCatalogId"`
}

type updateUserPlaylistRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type addUserPlaylistTrackRequest struct {
	TrackID string `json:"trackId"`
}

type setUserPlaylistTracksRequest struct {
	TrackIDs []string `json:"trackIds"`
}

func userPlaylistJSON(p userplaylist.UserPlaylist) map[string]any {
	return map[string]any{
		"id":          p.ID,
		"userId":      p.UserID,
		"name":        p.Name,
		"description": p.Description,
		"trackIds":    p.TrackIDs,
		"createdAt":   p.CreatedAt,
		"updatedAt":   p.UpdatedAt,
	}
}

func (handler *Handler) createUserPlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireUserPlaylistService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	var req createUserPlaylistRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, maxRequestBodyBytes)).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	p, err := handler.userPlaylistService.CreatePlaylist(r.Context(), user.ID, userplaylist.CreateRequest{
		Name:            req.Name,
		Description:     req.Description,
		SourceCatalogID: req.SourceCatalogID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, userPlaylistJSON(p))
}

func (handler *Handler) listUserPlaylists(w http.ResponseWriter, r *http.Request) {
	if !handler.requireUserPlaylistService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	playlists, err := handler.userPlaylistService.ListPlaylists(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	items := make([]map[string]any, len(playlists))
	for i, p := range playlists {
		items[i] = userPlaylistJSON(p)
	}
	writeJSON(w, http.StatusOK, map[string]any{"playlists": items})
}

func (handler *Handler) getUserPlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireUserPlaylistService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	p, err := handler.userPlaylistService.GetPlaylist(r.Context(), user.ID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, userPlaylistJSON(p))
}

func (handler *Handler) patchUserPlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireUserPlaylistService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	var req updateUserPlaylistRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, maxRequestBodyBytes)).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	p, err := handler.userPlaylistService.UpdatePlaylist(r.Context(), user.ID, r.PathValue("id"), userplaylist.UpdateRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, userPlaylistJSON(p))
}

func (handler *Handler) deleteUserPlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireUserPlaylistService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	if err := handler.userPlaylistService.DeletePlaylist(r.Context(), user.ID, r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) addUserPlaylistTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireUserPlaylistService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	var req addUserPlaylistTrackRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, maxRequestBodyBytes)).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if err := handler.userPlaylistService.AddTrack(r.Context(), user.ID, r.PathValue("id"), req.TrackID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (handler *Handler) removeUserPlaylistTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireUserPlaylistService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	if err := handler.userPlaylistService.RemoveTrack(r.Context(), user.ID, r.PathValue("id"), r.PathValue("trackId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) getUserPlaylistTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireUserPlaylistService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	p, err := handler.userPlaylistService.GetPlaylist(r.Context(), user.ID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"trackIds": p.TrackIDs})
}

func (handler *Handler) setUserPlaylistTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireUserPlaylistService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	var req setUserPlaylistTracksRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, maxRequestBodyBytes)).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if err := handler.userPlaylistService.SetTracks(r.Context(), user.ID, r.PathValue("id"), req.TrackIDs); err != nil {
		writeError(w, err)
		return
	}
	p, err := handler.userPlaylistService.GetPlaylist(r.Context(), user.ID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, userPlaylistJSON(p))
}

func (handler *Handler) copyCatalogPlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireUserPlaylistService(w) || !handler.requireCatalogService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	// Copying a catalog playlist to a user playlist doesn't need a body
	// because the name is derived from the source catalog playlist.
	// We pass an empty create request and set SourceCatalogID after creation.
	catalogId := r.PathValue("id")

	// Validate that the catalog playlist exists first
	_, err := handler.catalogService.GetPlaylist(r.Context(), catalogId)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "catalog_playlist_not_found", "catalog playlist not found")
		return
	}

	// Fetch the tracks from the catalog playlist
	tracks, err := handler.catalogService.GetPlaylistTracks(r.Context(), catalogId)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "failed to get playlist tracks")
		return
	}

	var trackIds []string
	for _, t := range tracks {
		trackIds = append(trackIds, t.ID)
	}

	// Create user playlist with the catalog playlist as source
	// Note: we can't directly use CreatePlaylist because it expects a name
	// We need to either: 1) add a new method, or 2) use CreatePlaylist with a name derived from catalog playlist
	// For simplicity, let's use CreatePlaylist and set SourceCatalogID after
	// But we need to get the catalog playlist name first
	pl, err := handler.catalogService.GetPlaylist(r.Context(), catalogId)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "failed to get playlist")
		return
	}

	p, err := handler.userPlaylistService.CreatePlaylist(r.Context(), user.ID, userplaylist.CreateRequest{
		Name:            pl.Name,
		SourceCatalogID: catalogId,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	// Now add all tracks
	for _, tid := range trackIds {
		if err := handler.userPlaylistService.AddTrack(r.Context(), user.ID, p.ID, tid); err != nil {
			// Don't fail the whole operation for a single track failure
			log.Printf("failed to add track %s to playlist %s: %v", tid, p.ID, err)
		}
	}

	// Refresh the playlist data to return the full state
	p, err = handler.userPlaylistService.GetPlaylist(r.Context(), user.ID, p.ID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, userPlaylistJSON(p))
}

package httpapi

import (
	"net/http"

	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/favorites"
)

// ---- favorites handlers ----

func (handler *Handler) requireFavoritesService(w http.ResponseWriter) bool {
	if handler.favoritesService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "favorites_not_configured", "favorites service is not configured")
		return false
	}
	return true
}

func (handler *Handler) addFavoriteTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireFavoritesService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	trackID := r.PathValue("trackId")
	if err := handler.favoritesService.AddFavorite(r.Context(), user.ID, trackID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (handler *Handler) removeFavoriteTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireFavoritesService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	trackID := r.PathValue("trackId")
	if err := handler.favoritesService.RemoveFavorite(r.Context(), user.ID, trackID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) listFavoriteTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireFavoritesService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	limit, err := parseMediaObjectListInt(r.URL.Query().Get("limit"), "limit", favorites.DefaultListLimit)
	if err != nil {
		writeError(w, err)
		return
	}
	offset, err := parseMediaObjectListInt(r.URL.Query().Get("offset"), "offset", 0)
	if err != nil {
		writeError(w, err)
		return
	}
	page, err := handler.favoritesService.ListFavorites(r.Context(), user.ID, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	pagination := map[string]any{
		"limit":   limit,
		"offset":  offset,
		"total":   page.Total,
		"hasMore": offset+limit < page.Total,
	}
	// If catalog service is available, resolve track IDs to full Track objects.
	if handler.catalogService != nil && len(page.TrackIDs) > 0 {
		tracks := make([]catalog.Track, 0, len(page.TrackIDs))
		for _, tid := range page.TrackIDs {
			t, err := handler.catalogService.GetTrack(r.Context(), tid)
			if err == nil {
				tracks = append(tracks, t)
			}
		}
		// All favorites are — by definition — favorited; isFavorite is always true.
		views := make([]trackView, len(tracks))
		for i, t := range tracks {
			views[i] = trackView{Track: t, IsFavorite: true}
		}
		writeJSON(w, http.StatusOK, map[string]any{"tracks": views, "pagination": pagination})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"trackIds":   page.TrackIDs,
		"pagination": pagination,
	})
}

// ---- admin favorites handlers ----

func (handler *Handler) adminListUserFavorites(w http.ResponseWriter, r *http.Request) {
	if !handler.requireFavoritesService(w) {
		return
	}
	userID := r.PathValue("userId")
	limit, err := parseMediaObjectListInt(r.URL.Query().Get("limit"), "limit", favorites.DefaultListLimit)
	if err != nil {
		writeError(w, err)
		return
	}
	offset, err := parseMediaObjectListInt(r.URL.Query().Get("offset"), "offset", 0)
	if err != nil {
		writeError(w, err)
		return
	}
	page, err := handler.favoritesService.ListFavorites(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"trackIds": page.TrackIDs,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   page.Total,
			"hasMore": offset+limit < page.Total,
		},
	})
}

func (handler *Handler) adminClearUserFavorites(w http.ResponseWriter, r *http.Request) {
	if !handler.requireFavoritesService(w) {
		return
	}
	if err := handler.favoritesService.ClearUserFavorites(r.Context(), r.PathValue("userId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) adminRemoveUserFavoriteTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireFavoritesService(w) {
		return
	}
	if err := handler.favoritesService.AdminRemoveFavorite(r.Context(), r.PathValue("userId"), r.PathValue("trackId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

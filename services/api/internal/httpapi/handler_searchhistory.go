package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
)

// ---- search history handlers ----

func (handler *Handler) requireSearchHistoryService(w http.ResponseWriter) bool {
	if handler.searchHistoryService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "search_history_not_configured", "search history service is not configured")
		return false
	}
	return true
}

func (handler *Handler) getMySearchHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireSearchHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	entries, err := handler.searchHistoryService.Get(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (handler *Handler) putMySearchHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireSearchHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	var req []string
	if err := json.NewDecoder(io.LimitReader(r.Body, maxRequestBodyBytes)).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if err := handler.searchHistoryService.Put(r.Context(), user.ID, req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

func (handler *Handler) deleteMySearchHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireSearchHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	if err := handler.searchHistoryService.Delete(r.Context(), user.ID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"inori-music/services/api/internal/playerstate"
)

// ---- player state handlers ----

func (handler *Handler) requirePlayerstateService(w http.ResponseWriter) bool {
	if handler.playerstateService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "playerstate_not_configured", "player state service is not configured")
		return false
	}
	return true
}

// playerStateRequest is the client-supplied player state. UpdatedAt is
// intentionally omitted: the server assigns it (last-write-wins).
type playerStateRequest struct {
	Queue           []string `json:"queue"`
	CurrentIndex    int      `json:"currentIndex"`
	PositionSeconds float64  `json:"positionSeconds"`
	Repeat          string   `json:"repeat"`
	Shuffle         bool     `json:"shuffle"`
	Volume          float64  `json:"volume"`
	Speed           float64  `json:"speed"`
	Status          string   `json:"status"`
}

func (handler *Handler) getMyPlayerState(w http.ResponseWriter, r *http.Request) {
	if !handler.requirePlayerstateService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	state, err := handler.playerstateService.Get(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (handler *Handler) putMyPlayerState(w http.ResponseWriter, r *http.Request) {
	if !handler.requirePlayerstateService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	var req playerStateRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, maxRequestBodyBytes)).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	stored, err := handler.playerstateService.Put(r.Context(), user.ID, playerstate.PlayerState{
		Queue:           req.Queue,
		CurrentIndex:    req.CurrentIndex,
		PositionSeconds: req.PositionSeconds,
		Repeat:          req.Repeat,
		Shuffle:         req.Shuffle,
		Volume:          req.Volume,
		Speed:           req.Speed,
		Status:          req.Status,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stored)
}

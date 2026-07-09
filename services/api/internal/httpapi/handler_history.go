package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"inori-music/services/api/internal/history"
)

// ---- playback history handlers ----

func (handler *Handler) requireHistoryService(w http.ResponseWriter) bool {
	if handler.historyService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "history_not_configured", "history service is not configured")
		return false
	}
	return true
}

type recordPlayRequest struct {
	TrackID  string `json:"trackId"`
	PlayedAt string `json:"playedAt,omitempty"` // RFC3339; defaults to server now when empty
}

func (handler *Handler) recordPlayEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	var req recordPlayRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.TrackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	var playedAt time.Time
	if req.PlayedAt != "" {
		t, err := time.Parse(time.RFC3339, req.PlayedAt)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "validation_error", "playedAt must be an RFC3339 timestamp")
			return
		}
		playedAt = t
	}
	event, err := handler.historyService.RecordPlay(r.Context(), user.ID, req.TrackID, playedAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

func (handler *Handler) listPlayEvents(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	q := r.URL.Query()
	limit := 0
	offset := 0
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return
		}
		limit = v
	}
	if raw := q.Get("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_offset", "offset must be a non-negative integer")
			return
		}
		offset = v
	}
	tf, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	asc, ok := parseHistoryOrder(w, r)
	if !ok {
		return
	}
	events, total, err := handler.historyService.ListPlays(r.Context(), history.PlayEventFilter{
		UserID:  user.ID,
		TrackID: q.Get("trackId"),
		Since:   tf.Since,
		Until:   tf.Until,
		Limit:   limit,
		Offset:  offset,
		Asc:     asc,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"hasMore": offset+limit < total && limit > 0,
		},
	})
}

func (handler *Handler) clearHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	if err := handler.historyService.ClearHistory(r.Context(), user.ID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) getMyHistoryStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	stats, err := handler.historyService.GetMyStats(r.Context(), history.UserStatsFilter{
		UserID: user.ID,
		Since:  f.Since,
		Until:  f.Until,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) getMyTopTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	tracks, err := handler.historyService.GetMyTopTracks(r.Context(), history.UserStatsFilter{
		UserID: user.ID,
		Since:  f.Since,
		Until:  f.Until,
	}, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": tracks})
}

func (handler *Handler) getMyHistoryTimeline(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	q := r.URL.Query()

	sinceRaw := q.Get("since")
	untilRaw := q.Get("until")
	if sinceRaw == "" || untilRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_time_bounds", "both since and until are required for timeline queries")
		return
	}

	since, err := time.Parse(time.RFC3339, sinceRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
		return
	}
	until, err := time.Parse(time.RFC3339, untilRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
		return
	}

	gran := history.TimelineGranularity(q.Get("granularity"))
	if gran == "" {
		gran = history.GranularityDay
	}
	switch gran {
	case history.GranularityDay, history.GranularityWeek, history.GranularityMonth:
		// valid
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_granularity", "granularity must be day, week, or month")
		return
	}

	buckets, err := handler.historyService.GetMyTimeline(r.Context(), history.TimelineFilter{
		Since:       since.UTC(),
		Until:       until.UTC(),
		Granularity: gran,
		UserID:      user.ID,
		TrackID:     q.Get("trackId"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"buckets": buckets})
}

// getMyHistorySummary returns combined stats and top-tracks for the authenticated
// viewer in one request.
func (handler *Handler) getMyHistorySummary(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	topN, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	sf := history.UserStatsFilter{UserID: user.ID, Since: f.Since, Until: f.Until}
	stats, err := handler.historyService.GetMyStats(r.Context(), sf)
	if err != nil {
		writeError(w, err)
		return
	}
	if topN <= 0 {
		topN = 10
	}
	tracks, err := handler.historyService.GetMyTopTracks(r.Context(), sf, topN)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"stats": stats, "topTracks": tracks})
}

// getMyTrackHistory returns the calling user's paginated play history for a
// specific track, identified by {trackId} in the path.
func (handler *Handler) getMyTrackHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	q := r.URL.Query()
	limit := 0
	offset := 0
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return
		}
		limit = v
	}
	if raw := q.Get("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_offset", "offset must be a non-negative integer")
			return
		}
		offset = v
	}
	tf, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	asc, ok := parseHistoryOrder(w, r)
	if !ok {
		return
	}
	events, total, err := handler.historyService.ListPlays(r.Context(), history.PlayEventFilter{
		UserID:  user.ID,
		TrackID: trackID,
		Since:   tf.Since,
		Until:   tf.Until,
		Limit:   limit,
		Offset:  offset,
		Asc:     asc,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"hasMore": offset+limit < total && limit > 0,
		},
	})
}

// getMyTrackStats returns aggregate play counts for the authenticated viewer on a specific track.
func (handler *Handler) getMyTrackStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	q := r.URL.Query()
	var sf history.UserStatsFilter
	if sinceRaw := q.Get("since"); sinceRaw != "" {
		t, err := time.Parse(time.RFC3339, sinceRaw)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
			return
		}
		sf.Since = t.UTC()
	}
	if untilRaw := q.Get("until"); untilRaw != "" {
		t, err := time.Parse(time.RFC3339, untilRaw)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
			return
		}
		sf.Until = t.UTC()
	}
	stats, err := handler.historyService.GetMyTrackStats(r.Context(), user.ID, trackID, sf)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// getMyTrackTimeline returns the calling user's play-event counts for a specific
// track grouped by time bucket. {trackId} is required in the path; since and
// until are required query params; granularity defaults to "day".
func (handler *Handler) getMyTrackTimeline(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	q := r.URL.Query()

	sinceRaw := q.Get("since")
	untilRaw := q.Get("until")
	if sinceRaw == "" || untilRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_time_bounds", "both since and until are required for timeline queries")
		return
	}
	since, err := time.Parse(time.RFC3339, sinceRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
		return
	}
	until, err := time.Parse(time.RFC3339, untilRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
		return
	}

	gran := history.TimelineGranularity(q.Get("granularity"))
	if gran == "" {
		gran = history.GranularityDay
	}
	switch gran {
	case history.GranularityDay, history.GranularityWeek, history.GranularityMonth:
		// valid
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_granularity", "granularity must be day, week, or month")
		return
	}

	buckets, err := handler.historyService.GetMyTimeline(r.Context(), history.TimelineFilter{
		Since:       since.UTC(),
		Until:       until.UTC(),
		Granularity: gran,
		UserID:      user.ID,
		TrackID:     trackID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"buckets": buckets})
}

// getMyTrackSummary returns the viewer's per-track play stats combined with their
// overall top tracks for cross-track context.
func (handler *Handler) getMyTrackSummary(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	topN, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	summary, err := handler.historyService.GetMyTrackSummary(r.Context(), user.ID, trackID, history.UserStatsFilter{
		UserID: user.ID,
		Since:  f.Since,
		Until:  f.Until,
	}, topN)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (handler *Handler) getAdminHistoryStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	stats, err := handler.historyService.GetHistoryStats(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) getAdminUserStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	stats, err := handler.historyService.GetAdminUserStats(r.Context(), history.UserStatsFilter{
		UserID: userID,
		Since:  f.Since,
		Until:  f.Until,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) getAdminUserTopTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	tracks, err := handler.historyService.GetAdminUserTopTracks(r.Context(), history.UserStatsFilter{
		UserID: userID,
		Since:  f.Since,
		Until:  f.Until,
	}, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": tracks})
}

// getAdminUserHistorySummary returns combined stats and top-tracks for a specific
// user in one request; intended for admin dashboard use.
func (handler *Handler) getAdminUserHistorySummary(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	topN, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	summary, err := handler.historyService.GetAdminUserSummary(r.Context(), userID, history.UserStatsFilter{
		Since: f.Since,
		Until: f.Until,
	}, topN)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (handler *Handler) getAdminTrackStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	stats, err := handler.historyService.GetTrackStats(r.Context(), history.TrackStatsFilter{
		TrackID: trackID,
		Since:   f.Since,
		Until:   f.Until,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) getAdminTrackTopListeners(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	users, err := handler.historyService.GetTrackTopListeners(r.Context(), history.TrackStatsFilter{
		TrackID: trackID,
		Since:   f.Since,
		Until:   f.Until,
	}, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

// getAdminTrackHistorySummary returns combined stats and top-listeners for a
// specific track in one request; intended for admin dashboard use.
func (handler *Handler) getAdminTrackHistorySummary(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	topN, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	summary, err := handler.historyService.GetTrackSummary(r.Context(), trackID, history.TrackStatsFilter{
		Since: f.Since,
		Until: f.Until,
	}, topN)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// getAdminHistorySummary returns combined system-wide aggregate stats, top tracks,
// and top users in one request; intended for admin dashboard use.
func (handler *Handler) getAdminHistorySummary(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	topN, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	summary, err := handler.historyService.GetGlobalSummary(r.Context(), f, topN)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (handler *Handler) getAdminTopTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	tracks, err := handler.historyService.GetTopTracks(r.Context(), f, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": tracks})
}

func (handler *Handler) getAdminTopUsers(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, ok := parseHistoryAdminLimit(w, r)
	if !ok {
		return
	}
	users, err := handler.historyService.GetTopUsers(r.Context(), f, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

// parseHistoryAdminFilter parses the optional ?since and ?until query params (RFC3339).
// Returns 400 if either value is unparseable or if since >= until when both are present.
func parseHistoryAdminFilter(w http.ResponseWriter, r *http.Request) (history.StatsFilter, bool) {
	q := r.URL.Query()
	var f history.StatsFilter

	if raw := q.Get("since"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
			return history.StatsFilter{}, false
		}
		f.Since = t.UTC()
	}

	if raw := q.Get("until"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
			return history.StatsFilter{}, false
		}
		f.Until = t.UTC()
	}

	if !f.Since.IsZero() && !f.Until.IsZero() && !f.Since.Before(f.Until) {
		writeAPIError(w, http.StatusBadRequest, "invalid_time_range", "since must be before until")
		return history.StatsFilter{}, false
	}

	return f, true
}

// parseHistoryAdminLimit parses the optional ?limit query param (default 10, max 100).
func parseHistoryAdminLimit(w http.ResponseWriter, r *http.Request) (int, bool) {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return 0, true // service applies default
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
		return 0, false
	}
	return v, true
}

func (handler *Handler) getAdminUserHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	limit, offset, ok := parseHistoryAdminPagination(w, r)
	if !ok {
		return
	}
	tf, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	asc, ok := parseHistoryOrder(w, r)
	if !ok {
		return
	}
	events, total, err := handler.historyService.GetUserHistory(r.Context(), history.PlayEventFilter{
		UserID:  userID,
		TrackID: r.URL.Query().Get("trackId"),
		Since:   tf.Since,
		Until:   tf.Until,
		Limit:   limit,
		Offset:  offset,
		Asc:     asc,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"hasMore": limit > 0 && offset+limit < total,
		},
	})
}

func (handler *Handler) getAdminTrackHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	limit, offset, ok := parseHistoryAdminPagination(w, r)
	if !ok {
		return
	}
	tf, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	asc, ok := parseHistoryOrder(w, r)
	if !ok {
		return
	}
	events, total, err := handler.historyService.GetTrackHistory(r.Context(), history.AdminPlayEventFilter{
		TrackID: trackID,
		UserID:  r.URL.Query().Get("userId"),
		Since:   tf.Since,
		Until:   tf.Until,
		Limit:   limit,
		Offset:  offset,
		Asc:     asc,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"hasMore": limit > 0 && offset+limit < total,
		},
	})
}

func parseHistoryAdminPagination(w http.ResponseWriter, r *http.Request) (limit, offset int, ok bool) {
	q := r.URL.Query()
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return 0, 0, false
		}
		limit = v
	}
	if raw := q.Get("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_offset", "offset must be a non-negative integer")
			return 0, 0, false
		}
		offset = v
	}
	return limit, offset, true
}

// parseHistoryOrder parses the optional ?order=asc|desc query parameter.
// Returns true for ascending, false (default) for descending.
// Writes a 400 and returns ok=false for any value other than "asc" or "desc".
func parseHistoryOrder(w http.ResponseWriter, r *http.Request) (asc bool, ok bool) {
	raw := r.URL.Query().Get("order")
	switch raw {
	case "", "desc":
		return false, true
	case "asc":
		return true, true
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_order", `order must be "asc" or "desc"`)
		return false, false
	}
}

func (handler *Handler) deleteAdminUserHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	if err := handler.historyService.AdminDeleteUserHistory(r.Context(), userID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) deleteAdminTrackHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	if err := handler.historyService.AdminDeleteTrackHistory(r.Context(), trackID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) deleteAdminHistoryWindow(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	if f.Since.IsZero() && f.Until.IsZero() {
		writeAPIError(w, http.StatusBadRequest, "missing_time_filter", "at least one of since or until is required")
		return
	}
	if err := handler.historyService.AdminDeleteHistoryWindow(r.Context(), f); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) getAdminAllHistory(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	f, ok := parseHistoryAdminFilter(w, r)
	if !ok {
		return
	}
	limit, offset, ok := parseHistoryAdminPagination(w, r)
	if !ok {
		return
	}
	asc, ok := parseHistoryOrder(w, r)
	if !ok {
		return
	}
	events, total, err := handler.historyService.GetAllHistory(r.Context(), history.GlobalPlayEventFilter{
		UserID:  r.URL.Query().Get("userId"),
		TrackID: r.URL.Query().Get("trackId"),
		Since:   f.Since,
		Until:   f.Until,
		Limit:   limit,
		Offset:  offset,
		Asc:     asc,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"pagination": map[string]any{
			"limit":   limit,
			"offset":  offset,
			"total":   total,
			"hasMore": limit > 0 && offset+limit < total,
		},
	})
}

func (handler *Handler) getAdminHistoryTimeline(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	q := r.URL.Query()

	sinceRaw := q.Get("since")
	untilRaw := q.Get("until")
	if sinceRaw == "" || untilRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_time_bounds", "both since and until are required for timeline queries")
		return
	}

	since, err := time.Parse(time.RFC3339, sinceRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
		return
	}
	until, err := time.Parse(time.RFC3339, untilRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
		return
	}

	gran := history.TimelineGranularity(q.Get("granularity"))
	if gran == "" {
		gran = history.GranularityDay
	}
	switch gran {
	case history.GranularityDay, history.GranularityWeek, history.GranularityMonth:
		// valid
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_granularity", "granularity must be day, week, or month")
		return
	}

	buckets, err := handler.historyService.GetHistoryTimeline(r.Context(), history.TimelineFilter{
		Since:       since.UTC(),
		Until:       until.UTC(),
		Granularity: gran,
		UserID:      q.Get("userId"),
		TrackID:     q.Get("trackId"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"buckets": buckets})
}

// getAdminUserTimeline returns play-event counts for a specific user grouped by
// time bucket. {userId} is required in the path; since and until are required
// query params; granularity defaults to "day".
func (handler *Handler) getAdminUserTimeline(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	userID := r.PathValue("userId")
	if userID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	q := r.URL.Query()

	sinceRaw := q.Get("since")
	untilRaw := q.Get("until")
	if sinceRaw == "" || untilRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_time_bounds", "both since and until are required for timeline queries")
		return
	}
	since, err := time.Parse(time.RFC3339, sinceRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
		return
	}
	until, err := time.Parse(time.RFC3339, untilRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
		return
	}

	gran := history.TimelineGranularity(q.Get("granularity"))
	if gran == "" {
		gran = history.GranularityDay
	}
	switch gran {
	case history.GranularityDay, history.GranularityWeek, history.GranularityMonth:
		// valid
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_granularity", "granularity must be day, week, or month")
		return
	}

	buckets, err := handler.historyService.GetHistoryTimeline(r.Context(), history.TimelineFilter{
		Since:       since.UTC(),
		Until:       until.UTC(),
		Granularity: gran,
		UserID:      userID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"buckets": buckets})
}

// getAdminTrackTimeline returns play-event counts for a specific track grouped by
// time bucket. {trackId} is required in the path; since and until are required
// query params; granularity defaults to "day".
func (handler *Handler) getAdminTrackTimeline(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	trackID := r.PathValue("trackId")
	if trackID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackId is required")
		return
	}
	q := r.URL.Query()

	sinceRaw := q.Get("since")
	untilRaw := q.Get("until")
	if sinceRaw == "" || untilRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_time_bounds", "both since and until are required for timeline queries")
		return
	}
	since, err := time.Parse(time.RFC3339, sinceRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_since", "since must be an RFC3339 timestamp")
		return
	}
	until, err := time.Parse(time.RFC3339, untilRaw)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_until", "until must be an RFC3339 timestamp")
		return
	}

	gran := history.TimelineGranularity(q.Get("granularity"))
	if gran == "" {
		gran = history.GranularityDay
	}
	switch gran {
	case history.GranularityDay, history.GranularityWeek, history.GranularityMonth:
		// valid
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_granularity", "granularity must be day, week, or month")
		return
	}

	buckets, err := handler.historyService.GetHistoryTimeline(r.Context(), history.TimelineFilter{
		Since:       since.UTC(),
		Until:       until.UTC(),
		Granularity: gran,
		TrackID:     trackID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"buckets": buckets})
}

func (handler *Handler) getAdminEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	e, err := handler.historyService.GetEventByID(r.Context(), eventID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (handler *Handler) deleteAdminEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	if err := handler.historyService.DeleteEventByID(r.Context(), eventID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) getMyEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	e, err := handler.historyService.GetMyEvent(r.Context(), user.ID, eventID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (handler *Handler) deleteMyEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	if err := handler.historyService.DeleteMyEvent(r.Context(), user.ID, eventID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) patchAdminEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	var body struct {
		PlayedAt string `json:"playedAt"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, err)
		return
	}
	if body.PlayedAt == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_played_at", "playedAt is required")
		return
	}
	t, err := time.Parse(time.RFC3339, body.PlayedAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_played_at", "playedAt must be an RFC3339 timestamp")
		return
	}
	e, err := handler.historyService.UpdateEventByID(r.Context(), eventID, t)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (handler *Handler) patchMyEvent(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	eventID := r.PathValue("eventId")
	if eventID == "" {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "eventId is required")
		return
	}
	var body struct {
		PlayedAt string `json:"playedAt"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, err)
		return
	}
	if body.PlayedAt == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_played_at", "playedAt is required")
		return
	}
	t, err := time.Parse(time.RFC3339, body.PlayedAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_played_at", "playedAt must be an RFC3339 timestamp")
		return
	}
	e, err := handler.historyService.UpdateMyEvent(r.Context(), user.ID, eventID, t)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (handler *Handler) batchDeleteAdminEvents(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, err)
		return
	}
	if len(body.IDs) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_ids", "ids must be a non-empty array")
		return
	}
	if len(body.IDs) > history.MaxBatchDeleteIDs {
		writeAPIError(w, http.StatusBadRequest, "invalid_ids",
			fmt.Sprintf("ids must not exceed %d entries", history.MaxBatchDeleteIDs))
		return
	}
	deleted, err := handler.historyService.BatchDeleteEvents(r.Context(), body.IDs)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
}

func (handler *Handler) batchDeleteMyEvents(w http.ResponseWriter, r *http.Request) {
	if !handler.requireHistoryService(w) {
		return
	}
	user, ok := userFromContext(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, err)
		return
	}
	if len(body.IDs) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_ids", "ids must be a non-empty array")
		return
	}
	if len(body.IDs) > history.MaxBatchDeleteIDs {
		writeAPIError(w, http.StatusBadRequest, "invalid_ids",
			fmt.Sprintf("ids must not exceed %d entries", history.MaxBatchDeleteIDs))
		return
	}
	deleted, err := handler.historyService.BatchDeleteMyEvents(r.Context(), user.ID, body.IDs)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
}

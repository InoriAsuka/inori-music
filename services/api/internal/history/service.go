package history

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	DefaultListLimit = 50
	MaxListLimit     = 500
)

// Service coordinates playback history persistence.
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// RecordPlay saves one play event for the given user and track.
// playedAt is the client-reported time; it defaults to now when zero.
func (s *Service) RecordPlay(ctx context.Context, userID, trackID string, playedAt time.Time) (PlayEvent, error) {
	if userID == "" {
		return PlayEvent{}, fmt.Errorf("userID is required")
	}
	if trackID == "" {
		return PlayEvent{}, fmt.Errorf("trackID is required")
	}
	if playedAt.IsZero() {
		playedAt = s.now()
	}
	id, err := newID()
	if err != nil {
		return PlayEvent{}, fmt.Errorf("generate event id: %w", err)
	}
	e := PlayEvent{
		ID:        id,
		UserID:    userID,
		TrackID:   trackID,
		PlayedAt:  playedAt.UTC(),
		CreatedAt: s.now().UTC(),
	}
	if err := s.repo.SavePlayEvent(ctx, e); err != nil {
		return PlayEvent{}, err
	}
	return e, nil
}

// ListPlays returns the play events for a user in reverse-chronological order.
func (s *Service) ListPlays(ctx context.Context, f PlayEventFilter) ([]PlayEvent, int, error) {
	if f.Limit <= 0 {
		f.Limit = DefaultListLimit
	}
	if f.Limit > MaxListLimit {
		f.Limit = MaxListLimit
	}
	return s.repo.ListPlayEvents(ctx, f)
}

// ClearHistory deletes all play events for the given user.
func (s *Service) ClearHistory(ctx context.Context, userID string) error {
	return s.repo.DeletePlayEventsByUser(ctx, userID)
}

const MaxBatchDeleteIDs = 100

// BatchDeleteEvents deletes a set of play events by ID; intended for admin use.
// Returns the count of actually deleted events. Duplicate or unknown IDs are silently ignored.
// Returns an error if ids is empty or exceeds MaxBatchDeleteIDs.
func (s *Service) BatchDeleteEvents(ctx context.Context, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("at least one id is required")
	}
	if len(ids) > MaxBatchDeleteIDs {
		return 0, fmt.Errorf("batch size must not exceed %d", MaxBatchDeleteIDs)
	}
	return s.repo.DeletePlayEventsByIDs(ctx, ids)
}

// BatchDeleteMyEvents deletes own play events by ID; intended for viewer use.
// Events not owned by userID are silently skipped. Returns deleted count.
func (s *Service) BatchDeleteMyEvents(ctx context.Context, userID string, ids []string) (int, error) {
	if userID == "" {
		return 0, fmt.Errorf("userID is required")
	}
	if len(ids) == 0 {
		return 0, fmt.Errorf("at least one id is required")
	}
	if len(ids) > MaxBatchDeleteIDs {
		return 0, fmt.Errorf("batch size must not exceed %d", MaxBatchDeleteIDs)
	}
	return s.repo.DeletePlayEventsByIDsForUser(ctx, userID, ids)
}

// GetEventByID returns any play event by ID; intended for admin use.
func (s *Service) GetEventByID(ctx context.Context, id string) (PlayEvent, error) {
	if id == "" {
		return PlayEvent{}, fmt.Errorf("eventID is required")
	}
	return s.repo.GetPlayEventByID(ctx, id)
}

// DeleteEventByID deletes any play event by ID; intended for admin use.
func (s *Service) DeleteEventByID(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("eventID is required")
	}
	return s.repo.DeletePlayEventByID(ctx, id)
}

// UpdateEventByID updates the playedAt timestamp of any play event; intended for admin use.
func (s *Service) UpdateEventByID(ctx context.Context, id string, playedAt time.Time) (PlayEvent, error) {
	if id == "" {
		return PlayEvent{}, fmt.Errorf("eventID is required")
	}
	if playedAt.IsZero() {
		return PlayEvent{}, fmt.Errorf("playedAt is required")
	}
	return s.repo.UpdatePlayEventByID(ctx, id, playedAt)
}

// GetMyEvent returns a play event by ID only if it belongs to userID; for viewer use.
func (s *Service) GetMyEvent(ctx context.Context, userID, id string) (PlayEvent, error) {
	if id == "" {
		return PlayEvent{}, fmt.Errorf("eventID is required")
	}
	e, err := s.repo.GetPlayEventByID(ctx, id)
	if err != nil {
		return PlayEvent{}, err
	}
	if e.UserID != userID {
		return PlayEvent{}, ErrEventForbidden
	}
	return e, nil
}

// DeleteMyEvent deletes a play event by ID only if it belongs to userID; for viewer use.
func (s *Service) DeleteMyEvent(ctx context.Context, userID, id string) error {
	if id == "" {
		return fmt.Errorf("eventID is required")
	}
	e, err := s.repo.GetPlayEventByID(ctx, id)
	if err != nil {
		return err
	}
	if e.UserID != userID {
		return ErrEventForbidden
	}
	return s.repo.DeletePlayEventByID(ctx, id)
}

// UpdateMyEvent updates the playedAt timestamp of a play event owned by userID; for viewer use.
func (s *Service) UpdateMyEvent(ctx context.Context, userID, id string, playedAt time.Time) (PlayEvent, error) {
	if id == "" {
		return PlayEvent{}, fmt.Errorf("eventID is required")
	}
	if playedAt.IsZero() {
		return PlayEvent{}, fmt.Errorf("playedAt is required")
	}
	e, err := s.repo.GetPlayEventByID(ctx, id)
	if err != nil {
		return PlayEvent{}, err
	}
	if e.UserID != userID {
		return PlayEvent{}, ErrEventForbidden
	}
	return s.repo.UpdatePlayEventByID(ctx, id, playedAt)
}

// GetUserHistory returns paginated play events for any user; intended for admin use.
func (s *Service) GetUserHistory(ctx context.Context, f PlayEventFilter) ([]PlayEvent, int, error) {
	if f.Limit <= 0 {
		f.Limit = DefaultListLimit
	}
	if f.Limit > MaxListLimit {
		f.Limit = MaxListLimit
	}
	return s.repo.ListPlayEvents(ctx, f)
}

// GetTrackHistory returns paginated play events for a specific track across all users; intended for admin use.
func (s *Service) GetTrackHistory(ctx context.Context, f AdminPlayEventFilter) ([]PlayEvent, int, error) {
	if f.Limit <= 0 {
		f.Limit = DefaultListLimit
	}
	if f.Limit > MaxListLimit {
		f.Limit = MaxListLimit
	}
	return s.repo.ListPlayEventsByTrack(ctx, f)
}

// GetAllHistory returns paginated play events across all users and tracks; intended for admin use.
// Any of the GlobalPlayEventFilter fields may be set to further restrict results.
func (s *Service) GetAllHistory(ctx context.Context, f GlobalPlayEventFilter) ([]PlayEvent, int, error) {
	if f.Limit <= 0 {
		f.Limit = DefaultListLimit
	}
	if f.Limit > MaxListLimit {
		f.Limit = MaxListLimit
	}
	return s.repo.ListAllPlayEvents(ctx, f)
}

// AdminDeleteUserHistory deletes all play events for the given user; intended for admin use.
func (s *Service) AdminDeleteUserHistory(ctx context.Context, userID string) error {
	return s.repo.DeletePlayEventsByUserAdmin(ctx, userID)
}

// AdminDeleteTrackHistory deletes all play events for the given track across all users.
func (s *Service) AdminDeleteTrackHistory(ctx context.Context, trackID string) error {
	return s.repo.DeletePlayEventsByTrack(ctx, trackID)
}

// AdminDeleteHistoryWindow deletes play events within the given time bounds.
// At least one bound (Since or Until) must be set.
func (s *Service) AdminDeleteHistoryWindow(ctx context.Context, f StatsFilter) error {
	if f.Since.IsZero() && f.Until.IsZero() {
		return fmt.Errorf("at least one of since or until is required")
	}
	return s.repo.DeletePlayEventsInWindow(ctx, f)
}

// GetHistoryStats returns system-wide aggregate counts for admin use.
// f.Since optionally bounds the query to events on or after that time.
func (s *Service) GetHistoryStats(ctx context.Context, f StatsFilter) (HistoryStats, error) {
	return s.repo.HistoryStats(ctx, f)
}

// GetHistoryTimeline returns play-event counts grouped by time bucket (day/week/month).
// Both Since and Until must be non-zero and Since must be before Until.
// Granularity defaults to GranularityDay when not set.
func (s *Service) GetHistoryTimeline(ctx context.Context, f TimelineFilter) ([]TimelineBucket, error) {
	if f.Since.IsZero() || f.Until.IsZero() {
		return nil, ErrInvalidTimeRange
	}
	if !f.Since.Before(f.Until) {
		return nil, ErrInvalidTimeRange
	}
	switch f.Granularity {
	case GranularityDay, GranularityWeek, GranularityMonth:
		// valid
	case "":
		f.Granularity = GranularityDay
	default:
		return nil, fmt.Errorf("invalid granularity %q: must be day, week, or month", f.Granularity)
	}
	return s.repo.HistoryTimeline(ctx, f)
}

// GetAdminUserStats returns per-user aggregate counts for any user; intended for admin use.
func (s *Service) GetAdminUserStats(ctx context.Context, f UserStatsFilter) (UserHistoryStats, error) {
	if f.UserID == "" {
		return UserHistoryStats{}, fmt.Errorf("userID is required")
	}
	return s.repo.UserHistoryStats(ctx, f)
}

// GetAdminUserTopTracks returns the most-played tracks for any user; intended for admin use.
// limit ≤ 0 defaults to 10 and is clamped to 100.
func (s *Service) GetAdminUserTopTracks(ctx context.Context, f UserStatsFilter, limit int) ([]TrackPlayCount, error) {
	if f.UserID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.UserTopTracks(ctx, f, limit)
}

// GetAdminUserSummary returns a combined stats + top-tracks summary for any user;
// intended for admin use. topN ≤ 0 defaults to 10 and is clamped to 100.
func (s *Service) GetAdminUserSummary(ctx context.Context, userID string, f UserStatsFilter, topN int) (UserHistorySummary, error) {
	if userID == "" {
		return UserHistorySummary{}, fmt.Errorf("userID is required")
	}
	f.UserID = userID
	stats, err := s.GetAdminUserStats(ctx, f)
	if err != nil {
		return UserHistorySummary{}, err
	}
	tracks, err := s.GetAdminUserTopTracks(ctx, f, topN)
	if err != nil {
		return UserHistorySummary{}, err
	}
	return UserHistorySummary{Stats: stats, TopTracks: tracks}, nil
}

// GetTrackStats returns per-track aggregate counts; intended for admin use.
func (s *Service) GetTrackStats(ctx context.Context, f TrackStatsFilter) (TrackHistoryStatsResult, error) {
	if f.TrackID == "" {
		return TrackHistoryStatsResult{}, fmt.Errorf("trackID is required")
	}
	return s.repo.TrackHistoryStats(ctx, f)
}

// GetTrackTopListeners returns the users who have played a track the most; intended for admin use.
// limit ≤ 0 defaults to 10 and is clamped to 100.
func (s *Service) GetTrackTopListeners(ctx context.Context, f TrackStatsFilter, limit int) ([]UserPlayCount, error) {
	if f.TrackID == "" {
		return nil, fmt.Errorf("trackID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.TrackTopListeners(ctx, f, limit)
}

// GetTrackSummary returns a combined stats + top-listeners summary for any track;
// intended for admin use. topN ≤ 0 defaults to 10 and is clamped to 100.
func (s *Service) GetTrackSummary(ctx context.Context, trackID string, f TrackStatsFilter, topN int) (TrackHistorySummary, error) {
	if trackID == "" {
		return TrackHistorySummary{}, fmt.Errorf("trackID is required")
	}
	f.TrackID = trackID
	stats, err := s.GetTrackStats(ctx, f)
	if err != nil {
		return TrackHistorySummary{}, err
	}
	listeners, err := s.GetTrackTopListeners(ctx, f, topN)
	if err != nil {
		return TrackHistorySummary{}, err
	}
	return TrackHistorySummary{Stats: stats, TopListeners: listeners}, nil
}

// GetMyStats returns per-user aggregate counts for the authenticated viewer.
func (s *Service) GetMyStats(ctx context.Context, f UserStatsFilter) (UserHistoryStats, error) {
	if f.UserID == "" {
		return UserHistoryStats{}, fmt.Errorf("userID is required")
	}
	return s.repo.UserHistoryStats(ctx, f)
}

// GetMyTimeline returns the viewer's own play-event counts grouped by time bucket.
// f.UserID must be non-empty (injected from auth context).
// Both f.Since and f.Until must be set and Since must be before Until.
func (s *Service) GetMyTimeline(ctx context.Context, f TimelineFilter) ([]TimelineBucket, error) {
	if f.UserID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	if f.Since.IsZero() || f.Until.IsZero() {
		return nil, ErrInvalidTimeRange
	}
	if !f.Since.Before(f.Until) {
		return nil, ErrInvalidTimeRange
	}
	switch f.Granularity {
	case GranularityDay, GranularityWeek, GranularityMonth:
		// valid
	case "":
		f.Granularity = GranularityDay
	default:
		return nil, fmt.Errorf("invalid granularity %q: must be day, week, or month", f.Granularity)
	}
	return s.repo.HistoryTimeline(ctx, f)
}

// GetMyTopTracks returns the most-played tracks for the authenticated viewer.
// limit ≤ 0 defaults to 10 and is clamped to 100.
func (s *Service) GetMyTopTracks(ctx context.Context, f UserStatsFilter, limit int) ([]TrackPlayCount, error) {
	if f.UserID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.UserTopTracks(ctx, f, limit)
}

// GetMyTrackStats returns aggregate play counts for the authenticated viewer on a specific track.
// f.Since and f.Until optionally restrict the played_at window; f.UserID is ignored.
func (s *Service) GetMyTrackStats(ctx context.Context, userID, trackID string, f UserStatsFilter) (UserTrackStats, error) {
	if userID == "" {
		return UserTrackStats{}, fmt.Errorf("userID is required")
	}
	if trackID == "" {
		return UserTrackStats{}, fmt.Errorf("trackID is required")
	}
	return s.repo.UserTrackPlayStats(ctx, userID, trackID, f)
}

// GetMyTrackSummary returns a combined per-track play stats and overall top tracks
// snapshot for the authenticated viewer. topN ≤ 0 defaults to 10 and is clamped to 100.
func (s *Service) GetMyTrackSummary(ctx context.Context, userID, trackID string, f UserStatsFilter, topN int) (MyTrackSummary, error) {
	if topN <= 0 {
		topN = 10
	}
	if topN > 100 {
		topN = 100
	}
	stats, err := s.GetMyTrackStats(ctx, userID, trackID, f)
	if err != nil {
		return MyTrackSummary{}, err
	}
	recent, err := s.GetMyTopTracks(ctx, UserStatsFilter{UserID: userID, Since: f.Since, Until: f.Until}, topN)
	if err != nil {
		return MyTrackSummary{}, err
	}
	return MyTrackSummary{Stats: stats, RecentTracks: recent}, nil
}

// GetTopTracks returns the most-played tracks across all users.
// limit ≤ 0 defaults to 10 and is clamped to 100.
// f.Since optionally bounds the query to events on or after that time.
func (s *Service) GetTopTracks(ctx context.Context, f StatsFilter, limit int) ([]TrackPlayCount, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.TopTracks(ctx, f, limit)
}

// GetTopUsers returns the users with the most play events.
// limit ≤ 0 defaults to 10 and is clamped to 100.
// f.Since optionally bounds the query to events on or after that time.
func (s *Service) GetTopUsers(ctx context.Context, f StatsFilter, limit int) ([]UserPlayCount, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.TopUsers(ctx, f, limit)
}

// GetGlobalSummary returns a combined system-wide aggregate stats, top tracks,
// and top users snapshot for admin dashboard use.
// topN ≤ 0 defaults to 10 and is clamped to 100.
func (s *Service) GetGlobalSummary(ctx context.Context, f StatsFilter, topN int) (GlobalHistorySummary, error) {
	if topN <= 0 {
		topN = 10
	}
	if topN > 100 {
		topN = 100
	}
	stats, err := s.GetHistoryStats(ctx, f)
	if err != nil {
		return GlobalHistorySummary{}, err
	}
	tracks, err := s.GetTopTracks(ctx, f, topN)
	if err != nil {
		return GlobalHistorySummary{}, err
	}
	users, err := s.GetTopUsers(ctx, f, topN)
	if err != nil {
		return GlobalHistorySummary{}, err
	}
	return GlobalHistorySummary{Stats: stats, TopTracks: tracks, TopUsers: users}, nil
}

func newID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

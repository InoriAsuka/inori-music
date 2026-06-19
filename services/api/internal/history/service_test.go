package history_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"inori-music/services/api/internal/history"
)

func TestRecordPlayRequiresUserAndTrack(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	if _, err := svc.RecordPlay(ctx, "", "track-1", time.Time{}); err == nil {
		t.Error("expected error for empty userID")
	}
	if _, err := svc.RecordPlay(ctx, "user-1", "", time.Time{}); err == nil {
		t.Error("expected error for empty trackID")
	}
}

func TestRecordPlayDefaultsToNow(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	before := time.Now().UTC()
	e, err := svc.RecordPlay(ctx, "u1", "t1", time.Time{})
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}
	if e.PlayedAt.Before(before) || e.PlayedAt.After(after) {
		t.Errorf("PlayedAt %v not between %v and %v", e.PlayedAt, before, after)
	}
}

func TestListPlaysNewestFirst(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	for i := 0; i < 3; i++ {
		_, err := svc.RecordPlay(ctx, "u1", "t1", now.Add(time.Duration(i)*time.Second))
		if err != nil {
			t.Fatalf("RecordPlay %d: %v", i, err)
		}
	}

	events, total, err := svc.ListPlays(ctx, history.PlayEventFilter{UserID: "u1"})
	if err != nil {
		t.Fatalf("ListPlays: %v", err)
	}
	if len(events) != 3 || total != 3 {
		t.Fatalf("events=%d total=%d, want 3/3", len(events), total)
	}
	if !events[0].PlayedAt.After(events[1].PlayedAt) {
		t.Error("events not in newest-first order")
	}
}

func TestListPlaysPagination(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		svc.RecordPlay(ctx, "u1", "t1", now.Add(time.Duration(i)*time.Second))
	}

	events, total, err := svc.ListPlays(ctx, history.PlayEventFilter{UserID: "u1", Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("ListPlays: %v", err)
	}
	if len(events) != 2 || total != 5 {
		t.Fatalf("limit=2: events=%d total=%d, want 2/5", len(events), total)
	}
}

func TestClearHistory(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	svc.RecordPlay(ctx, "u1", "t1", time.Now())
	svc.RecordPlay(ctx, "u2", "t1", time.Now())

	if err := svc.ClearHistory(ctx, "u1"); err != nil {
		t.Fatalf("ClearHistory: %v", err)
	}
	events, total, _ := svc.ListPlays(ctx, history.PlayEventFilter{UserID: "u1"})
	if len(events) != 0 || total != 0 {
		t.Errorf("u1 events after clear = %d, want 0", len(events))
	}
	// u2 unaffected
	events, total, _ = svc.ListPlays(ctx, history.PlayEventFilter{UserID: "u2"})
	if total != 1 {
		t.Errorf("u2 total = %d, want 1", total)
	}
}

func TestGetHistoryStats(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	svc.RecordPlay(ctx, "u1", "t1", now)
	svc.RecordPlay(ctx, "u1", "t2", now)
	svc.RecordPlay(ctx, "u2", "t1", now)

	stats, err := svc.GetHistoryStats(ctx, history.StatsFilter{})
	if err != nil {
		t.Fatalf("GetHistoryStats: %v", err)
	}
	if stats.TotalEvents != 3 {
		t.Errorf("TotalEvents = %d, want 3", stats.TotalEvents)
	}
	if stats.UniqueUsers != 2 {
		t.Errorf("UniqueUsers = %d, want 2", stats.UniqueUsers)
	}
	if stats.UniqueTracks != 2 {
		t.Errorf("UniqueTracks = %d, want 2", stats.UniqueTracks)
	}
}

func TestGetTopTracks(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	// t1 played 3×, t2 played 1×
	for i := 0; i < 3; i++ {
		svc.RecordPlay(ctx, "u1", "t1", now)
	}
	svc.RecordPlay(ctx, "u2", "t2", now)

	tracks, err := svc.GetTopTracks(ctx, history.StatsFilter{}, 0) // 0 → default 10
	if err != nil {
		t.Fatalf("GetTopTracks: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("tracks = %d, want 2", len(tracks))
	}
	if tracks[0].TrackID != "t1" || tracks[0].PlayCount != 3 {
		t.Errorf("first track = %+v, want t1/3", tracks[0])
	}
	if tracks[1].TrackID != "t2" || tracks[1].PlayCount != 1 {
		t.Errorf("second track = %+v, want t2/1", tracks[1])
	}

	// Limit to 1
	top1, _ := svc.GetTopTracks(ctx, history.StatsFilter{}, 1)
	if len(top1) != 1 {
		t.Errorf("limit=1: tracks = %d, want 1", len(top1))
	}
}

func TestGetTopUsers(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	// u1 played 2×, u2 played 1×
	svc.RecordPlay(ctx, "u1", "t1", now)
	svc.RecordPlay(ctx, "u1", "t2", now)
	svc.RecordPlay(ctx, "u2", "t1", now)

	users, err := svc.GetTopUsers(ctx, history.StatsFilter{}, 0) // 0 → default 10
	if err != nil {
		t.Fatalf("GetTopUsers: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("users = %d, want 2", len(users))
	}
	if users[0].UserID != "u1" || users[0].PlayCount != 2 {
		t.Errorf("first user = %+v, want u1/2", users[0])
	}
}

func TestGetHistoryStatsSinceFilter(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()
	cutoff := base.Add(5 * time.Second)

	// 2 events before cutoff, 1 event at/after cutoff
	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u2", "t2", base.Add(2*time.Second))
	svc.RecordPlay(ctx, "u3", "t3", base.Add(10*time.Second))

	// All-time: 3 events, 3 users, 3 tracks
	all, err := svc.GetHistoryStats(ctx, history.StatsFilter{})
	if err != nil {
		t.Fatalf("GetHistoryStats all: %v", err)
	}
	if all.TotalEvents != 3 || all.UniqueUsers != 3 || all.UniqueTracks != 3 {
		t.Errorf("all-time stats = %+v, want 3/3/3", all)
	}

	// Since cutoff: only the event at +10s
	windowed, err := svc.GetHistoryStats(ctx, history.StatsFilter{Since: cutoff})
	if err != nil {
		t.Fatalf("GetHistoryStats since: %v", err)
	}
	if windowed.TotalEvents != 1 {
		t.Errorf("windowed TotalEvents = %d, want 1", windowed.TotalEvents)
	}
	if windowed.UniqueUsers != 1 {
		t.Errorf("windowed UniqueUsers = %d, want 1", windowed.UniqueUsers)
	}
	if windowed.UniqueTracks != 1 {
		t.Errorf("windowed UniqueTracks = %d, want 1", windowed.UniqueTracks)
	}
}

func TestGetTopTracksSinceFilter(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()
	cutoff := base.Add(5 * time.Second)

	// t1 played twice before cutoff, t2 played once after cutoff
	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t1", base.Add(2*time.Second))
	svc.RecordPlay(ctx, "u1", "t2", base.Add(10*time.Second))

	windowed, err := svc.GetTopTracks(ctx, history.StatsFilter{Since: cutoff}, 0)
	if err != nil {
		t.Fatalf("GetTopTracks since: %v", err)
	}
	if len(windowed) != 1 {
		t.Fatalf("windowed tracks = %d, want 1", len(windowed))
	}
	if windowed[0].TrackID != "t2" || windowed[0].PlayCount != 1 {
		t.Errorf("windowed track = %+v, want t2/1", windowed[0])
	}
}

func TestGetTopUsersSinceFilter(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()
	cutoff := base.Add(5 * time.Second)

	// u1 played twice before cutoff, u2 played once after cutoff
	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t1", base.Add(2*time.Second))
	svc.RecordPlay(ctx, "u2", "t1", base.Add(10*time.Second))

	windowed, err := svc.GetTopUsers(ctx, history.StatsFilter{Since: cutoff}, 0)
	if err != nil {
		t.Fatalf("GetTopUsers since: %v", err)
	}
	if len(windowed) != 1 {
		t.Fatalf("windowed users = %d, want 1", len(windowed))
	}
	if windowed[0].UserID != "u2" || windowed[0].PlayCount != 1 {
		t.Errorf("windowed user = %+v, want u2/1", windowed[0])
	}
}

func TestGetHistoryStatsUntilFilter(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	// event at base, event at +5s, event at +10s
	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u2", "t2", base.Add(5*time.Second))
	svc.RecordPlay(ctx, "u3", "t3", base.Add(10*time.Second))

	// until = base+8s excludes the +10s event (exclusive upper bound)
	windowed, err := svc.GetHistoryStats(ctx, history.StatsFilter{Until: base.Add(8 * time.Second)})
	if err != nil {
		t.Fatalf("GetHistoryStats until: %v", err)
	}
	if windowed.TotalEvents != 2 {
		t.Errorf("windowed TotalEvents = %d, want 2", windowed.TotalEvents)
	}
}

func TestGetTopTracksSinceUntilWindow(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	// t1 at base (excluded by since), t2 at +5s (inside window), t3 at +20s (excluded by until)
	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t2", base.Add(5*time.Second))
	svc.RecordPlay(ctx, "u1", "t3", base.Add(20*time.Second))

	f := history.StatsFilter{
		Since: base.Add(3 * time.Second),
		Until: base.Add(10 * time.Second),
	}
	tracks, err := svc.GetTopTracks(ctx, f, 0)
	if err != nil {
		t.Fatalf("GetTopTracks since+until: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("windowed tracks = %d, want 1", len(tracks))
	}
	if tracks[0].TrackID != "t2" {
		t.Errorf("windowed track = %q, want t2", tracks[0].TrackID)
	}
}

func TestGetUserHistory(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	// u1 plays t1 and t2; u2 plays t1
	svc.RecordPlay(ctx, "u1", "t1", now)
	svc.RecordPlay(ctx, "u1", "t2", now.Add(time.Second))
	svc.RecordPlay(ctx, "u2", "t1", now)

	events, total, err := svc.GetUserHistory(ctx, history.PlayEventFilter{UserID: "u1"})
	if err != nil {
		t.Fatalf("GetUserHistory: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	for _, e := range events {
		if e.UserID != "u1" {
			t.Errorf("event userID = %q, want u1", e.UserID)
		}
	}

	// pagination: limit=1 returns 1 event, total still 2
	paged, total2, err := svc.GetUserHistory(ctx, history.PlayEventFilter{UserID: "u1", Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("GetUserHistory paged: %v", err)
	}
	if len(paged) != 1 || total2 != 2 {
		t.Errorf("paged: len=%d total=%d, want 1/2", len(paged), total2)
	}
}

func TestGetTrackHistory(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	// t1 played by u1 (twice) and u2 (once); t2 played by u1
	svc.RecordPlay(ctx, "u1", "t1", now)
	svc.RecordPlay(ctx, "u1", "t1", now.Add(time.Second))
	svc.RecordPlay(ctx, "u2", "t1", now.Add(2*time.Second))
	svc.RecordPlay(ctx, "u1", "t2", now)

	events, total, err := svc.GetTrackHistory(ctx, history.AdminPlayEventFilter{TrackID: "t1"})
	if err != nil {
		t.Fatalf("GetTrackHistory: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	for _, e := range events {
		if e.TrackID != "t1" {
			t.Errorf("event trackID = %q, want t1", e.TrackID)
		}
	}

	// filter by user: only u1 events for t1 (2 events)
	filtered, total2, err := svc.GetTrackHistory(ctx, history.AdminPlayEventFilter{TrackID: "t1", UserID: "u1"})
	if err != nil {
		t.Fatalf("GetTrackHistory filtered: %v", err)
	}
	if total2 != 2 {
		t.Errorf("filtered total = %d, want 2", total2)
	}
	for _, e := range filtered {
		if e.UserID != "u1" {
			t.Errorf("filtered event userID = %q, want u1", e.UserID)
		}
	}
}

func TestAdminDeleteUserHistory(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	svc.RecordPlay(ctx, "u1", "t1", now)
	svc.RecordPlay(ctx, "u1", "t2", now.Add(time.Second))
	svc.RecordPlay(ctx, "u2", "t1", now)

	if err := svc.AdminDeleteUserHistory(ctx, "u1"); err != nil {
		t.Fatalf("AdminDeleteUserHistory: %v", err)
	}

	// u1 events deleted; u2 events intact
	_, total, err := svc.GetUserHistory(ctx, history.PlayEventFilter{UserID: "u1"})
	if err != nil {
		t.Fatalf("GetUserHistory after delete: %v", err)
	}
	if total != 0 {
		t.Errorf("u1 total after delete = %d, want 0", total)
	}

	_, total2, err := svc.GetUserHistory(ctx, history.PlayEventFilter{UserID: "u2"})
	if err != nil {
		t.Fatalf("GetUserHistory u2: %v", err)
	}
	if total2 != 1 {
		t.Errorf("u2 total = %d, want 1", total2)
	}
}

func TestAdminDeleteTrackHistory(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	svc.RecordPlay(ctx, "u1", "t1", now)
	svc.RecordPlay(ctx, "u2", "t1", now.Add(time.Second))
	svc.RecordPlay(ctx, "u1", "t2", now)

	if err := svc.AdminDeleteTrackHistory(ctx, "t1"); err != nil {
		t.Fatalf("AdminDeleteTrackHistory: %v", err)
	}

	_, total, err := svc.GetTrackHistory(ctx, history.AdminPlayEventFilter{TrackID: "t1"})
	if err != nil {
		t.Fatalf("GetTrackHistory after delete: %v", err)
	}
	if total != 0 {
		t.Errorf("t1 total after delete = %d, want 0", total)
	}

	// t2 events intact
	_, total2, err := svc.GetTrackHistory(ctx, history.AdminPlayEventFilter{TrackID: "t2"})
	if err != nil {
		t.Fatalf("GetTrackHistory t2: %v", err)
	}
	if total2 != 1 {
		t.Errorf("t2 total = %d, want 1", total2)
	}
}

func TestAdminDeleteHistoryWindow(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	// 3 events: base, base+5s, base+15s
	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t2", base.Add(5*time.Second))
	svc.RecordPlay(ctx, "u1", "t3", base.Add(15*time.Second))

	// no bounds: expect error
	if err := svc.AdminDeleteHistoryWindow(ctx, history.StatsFilter{}); err == nil {
		t.Fatal("expected error for empty StatsFilter, got nil")
	}

	// delete [base+3s, base+10s) — only t2 (base+5s) is in window
	f := history.StatsFilter{
		Since: base.Add(3 * time.Second),
		Until: base.Add(10 * time.Second),
	}
	if err := svc.AdminDeleteHistoryWindow(ctx, f); err != nil {
		t.Fatalf("AdminDeleteHistoryWindow: %v", err)
	}

	stats, err := svc.GetHistoryStats(ctx, history.StatsFilter{})
	if err != nil {
		t.Fatalf("GetHistoryStats after delete: %v", err)
	}
	if stats.TotalEvents != 2 {
		t.Errorf("TotalEvents after window delete = %d, want 2", stats.TotalEvents)
	}
}

func TestGetMyStats(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	// u1 plays t1 twice and t2 once; u2 plays t3
	svc.RecordPlay(ctx, "u1", "t1", now)
	svc.RecordPlay(ctx, "u1", "t1", now.Add(time.Second))
	svc.RecordPlay(ctx, "u1", "t2", now.Add(2*time.Second))
	svc.RecordPlay(ctx, "u2", "t3", now)

	stats, err := svc.GetMyStats(ctx, history.UserStatsFilter{UserID: "u1"})
	if err != nil {
		t.Fatalf("GetMyStats: %v", err)
	}
	if stats.TotalEvents != 3 {
		t.Errorf("TotalEvents = %d, want 3", stats.TotalEvents)
	}
	if stats.UniqueTracks != 2 {
		t.Errorf("UniqueTracks = %d, want 2", stats.UniqueTracks)
	}

	// empty userID → error
	if _, err := svc.GetMyStats(ctx, history.UserStatsFilter{}); err == nil {
		t.Fatal("expected error for empty UserID, got nil")
	}
}

func TestGetMyTopTracks(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	// u1: t1×3, t2×1; u2: t2×10 (must NOT appear in u1's results)
	for i := 0; i < 3; i++ {
		svc.RecordPlay(ctx, "u1", "t1", now.Add(time.Duration(i)*time.Second))
	}
	svc.RecordPlay(ctx, "u1", "t2", now.Add(10*time.Second))
	for i := 0; i < 10; i++ {
		svc.RecordPlay(ctx, "u2", "t2", now.Add(time.Duration(i)*time.Second))
	}

	tracks, err := svc.GetMyTopTracks(ctx, history.UserStatsFilter{UserID: "u1"}, 0)
	if err != nil {
		t.Fatalf("GetMyTopTracks: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("len = %d, want 2", len(tracks))
	}
	if tracks[0].TrackID != "t1" || tracks[0].PlayCount != 3 {
		t.Errorf("top track = %+v, want t1/3", tracks[0])
	}
	if tracks[1].TrackID != "t2" || tracks[1].PlayCount != 1 {
		t.Errorf("second track = %+v, want t2/1", tracks[1])
	}
}

func TestGetMyTopTracksTimeWindow(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	// u1 plays t1 at base (before window), t2 at +5s (inside), t3 at +20s (after)
	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t2", base.Add(5*time.Second))
	svc.RecordPlay(ctx, "u1", "t3", base.Add(20*time.Second))

	f := history.UserStatsFilter{
		UserID: "u1",
		Since:  base.Add(3 * time.Second),
		Until:  base.Add(10 * time.Second),
	}
	tracks, err := svc.GetMyTopTracks(ctx, f, 0)
	if err != nil {
		t.Fatalf("GetMyTopTracks windowed: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("windowed tracks = %d, want 1", len(tracks))
	}
	if tracks[0].TrackID != "t2" {
		t.Errorf("windowed track = %q, want t2", tracks[0].TrackID)
	}
}

func TestGetAdminUserStats(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	// u1 plays t1 twice and t2 once; u2 plays t3 — u2 must not affect u1's stats
	svc.RecordPlay(ctx, "u1", "t1", now)
	svc.RecordPlay(ctx, "u1", "t1", now.Add(time.Second))
	svc.RecordPlay(ctx, "u1", "t2", now.Add(2*time.Second))
	svc.RecordPlay(ctx, "u2", "t3", now)

	stats, err := svc.GetAdminUserStats(ctx, history.UserStatsFilter{UserID: "u1"})
	if err != nil {
		t.Fatalf("GetAdminUserStats: %v", err)
	}
	if stats.TotalEvents != 3 {
		t.Errorf("TotalEvents = %d, want 3", stats.TotalEvents)
	}
	if stats.UniqueTracks != 2 {
		t.Errorf("UniqueTracks = %d, want 2", stats.UniqueTracks)
	}

	// empty userID → error
	if _, err := svc.GetAdminUserStats(ctx, history.UserStatsFilter{}); err == nil {
		t.Fatal("expected error for empty UserID, got nil")
	}
}

func TestGetAdminUserTopTracks(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	// u1: t1×3, t2×1; u2: t2×10 (must NOT appear in u1's results)
	for i := 0; i < 3; i++ {
		svc.RecordPlay(ctx, "u1", "t1", now.Add(time.Duration(i)*time.Second))
	}
	svc.RecordPlay(ctx, "u1", "t2", now.Add(10*time.Second))
	for i := 0; i < 10; i++ {
		svc.RecordPlay(ctx, "u2", "t2", now.Add(time.Duration(i)*time.Second))
	}

	tracks, err := svc.GetAdminUserTopTracks(ctx, history.UserStatsFilter{UserID: "u1"}, 0)
	if err != nil {
		t.Fatalf("GetAdminUserTopTracks: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("len = %d, want 2", len(tracks))
	}
	if tracks[0].TrackID != "t1" || tracks[0].PlayCount != 3 {
		t.Errorf("top track = %+v, want t1/3", tracks[0])
	}
	if tracks[1].TrackID != "t2" || tracks[1].PlayCount != 1 {
		t.Errorf("second track = %+v, want t2/1", tracks[1])
	}
}

func TestGetAdminUserTopTracksTimeWindow(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	// u1 plays t1 at base (before window), t2 at +5s (inside), t3 at +20s (after)
	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t2", base.Add(5*time.Second))
	svc.RecordPlay(ctx, "u1", "t3", base.Add(20*time.Second))

	f := history.UserStatsFilter{
		UserID: "u1",
		Since:  base.Add(3 * time.Second),
		Until:  base.Add(10 * time.Second),
	}
	tracks, err := svc.GetAdminUserTopTracks(ctx, f, 0)
	if err != nil {
		t.Fatalf("GetAdminUserTopTracks windowed: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("windowed tracks = %d, want 1", len(tracks))
	}
	if tracks[0].TrackID != "t2" {
		t.Errorf("windowed track = %q, want t2", tracks[0].TrackID)
	}
}

func TestGetTrackStats(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	// u1 and u2 play t1; u3 plays t2 — t2 must not affect t1's stats
	svc.RecordPlay(ctx, "u1", "t1", now)
	svc.RecordPlay(ctx, "u1", "t1", now.Add(time.Second))
	svc.RecordPlay(ctx, "u2", "t1", now.Add(2*time.Second))
	svc.RecordPlay(ctx, "u3", "t2", now)

	stats, err := svc.GetTrackStats(ctx, history.TrackStatsFilter{TrackID: "t1"})
	if err != nil {
		t.Fatalf("GetTrackStats: %v", err)
	}
	if stats.TotalEvents != 3 {
		t.Errorf("TotalEvents = %d, want 3", stats.TotalEvents)
	}
	if stats.UniqueListeners != 2 {
		t.Errorf("UniqueListeners = %d, want 2", stats.UniqueListeners)
	}

	// empty trackID → error
	if _, err := svc.GetTrackStats(ctx, history.TrackStatsFilter{}); err == nil {
		t.Fatal("expected error for empty TrackID, got nil")
	}
}

func TestGetTrackTopListeners(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	// u1 plays t1×3, u2 plays t1×1; u2 also plays t2×10 (must NOT appear in t1's results)
	for i := 0; i < 3; i++ {
		svc.RecordPlay(ctx, "u1", "t1", now.Add(time.Duration(i)*time.Second))
	}
	svc.RecordPlay(ctx, "u2", "t1", now.Add(10*time.Second))
	for i := 0; i < 10; i++ {
		svc.RecordPlay(ctx, "u2", "t2", now.Add(time.Duration(i)*time.Second))
	}

	listeners, err := svc.GetTrackTopListeners(ctx, history.TrackStatsFilter{TrackID: "t1"}, 0)
	if err != nil {
		t.Fatalf("GetTrackTopListeners: %v", err)
	}
	if len(listeners) != 2 {
		t.Fatalf("len = %d, want 2", len(listeners))
	}
	if listeners[0].UserID != "u1" || listeners[0].PlayCount != 3 {
		t.Errorf("top listener = %+v, want u1/3", listeners[0])
	}
	if listeners[1].UserID != "u2" || listeners[1].PlayCount != 1 {
		t.Errorf("second listener = %+v, want u2/1", listeners[1])
	}
}

func TestGetTrackTopListenersTimeWindow(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	// u1 plays t1 at base (before window), +5s (inside), +20s (after)
	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u2", "t1", base.Add(5*time.Second))
	svc.RecordPlay(ctx, "u3", "t1", base.Add(20*time.Second))

	f := history.TrackStatsFilter{
		TrackID: "t1",
		Since:   base.Add(3 * time.Second),
		Until:   base.Add(10 * time.Second),
	}
	listeners, err := svc.GetTrackTopListeners(ctx, f, 0)
	if err != nil {
		t.Fatalf("GetTrackTopListeners windowed: %v", err)
	}
	if len(listeners) != 1 {
		t.Fatalf("windowed listeners = %d, want 1", len(listeners))
	}
	if listeners[0].UserID != "u2" {
		t.Errorf("windowed listener = %q, want u2", listeners[0].UserID)
	}
}

func TestGetHistoryTimelineDay(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	// 2 events on day 1, 3 events on day 2
	day1 := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	day2 := time.Date(2024, 1, 11, 9, 0, 0, 0, time.UTC)
	svc.RecordPlay(ctx, "u1", "t1", day1)
	svc.RecordPlay(ctx, "u1", "t2", day1.Add(time.Hour))
	svc.RecordPlay(ctx, "u2", "t1", day2)
	svc.RecordPlay(ctx, "u1", "t1", day2.Add(time.Hour))
	svc.RecordPlay(ctx, "u2", "t2", day2.Add(2*time.Hour))

	f := history.TimelineFilter{
		Since:       time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		Until:       time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
		Granularity: history.GranularityDay,
	}
	buckets, err := svc.GetHistoryTimeline(ctx, f)
	if err != nil {
		t.Fatalf("GetHistoryTimeline: %v", err)
	}
	if len(buckets) != 2 {
		t.Fatalf("buckets = %d, want 2", len(buckets))
	}
	if buckets[0].EventCount != 2 {
		t.Errorf("day1 count = %d, want 2", buckets[0].EventCount)
	}
	if buckets[1].EventCount != 3 {
		t.Errorf("day2 count = %d, want 3", buckets[1].EventCount)
	}
}

func TestGetHistoryTimelineWeek(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	// 2024-01-08 is Monday (week 1); 2024-01-15 is Monday (week 2)
	week1mon := time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)
	week2wed := time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC)
	svc.RecordPlay(ctx, "u1", "t1", week1mon)
	svc.RecordPlay(ctx, "u1", "t1", week1mon.Add(24*time.Hour)) // still week 1
	svc.RecordPlay(ctx, "u2", "t2", week2wed)

	f := history.TimelineFilter{
		Since:       time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		Until:       time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC),
		Granularity: history.GranularityWeek,
	}
	buckets, err := svc.GetHistoryTimeline(ctx, f)
	if err != nil {
		t.Fatalf("GetHistoryTimeline week: %v", err)
	}
	if len(buckets) != 2 {
		t.Fatalf("buckets = %d, want 2", len(buckets))
	}
	if buckets[0].EventCount != 2 {
		t.Errorf("week1 count = %d, want 2", buckets[0].EventCount)
	}
	if buckets[1].EventCount != 1 {
		t.Errorf("week2 count = %d, want 1", buckets[1].EventCount)
	}
}

func TestGetHistoryTimelineUserFilter(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	day := time.Date(2024, 3, 5, 8, 0, 0, 0, time.UTC)
	svc.RecordPlay(ctx, "u1", "t1", day)
	svc.RecordPlay(ctx, "u1", "t1", day.Add(time.Hour))
	svc.RecordPlay(ctx, "u2", "t2", day.Add(2*time.Hour)) // different user — must be excluded

	f := history.TimelineFilter{
		Since:       time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC),
		Until:       time.Date(2024, 3, 6, 0, 0, 0, 0, time.UTC),
		Granularity: history.GranularityDay,
		UserID:      "u1",
	}
	buckets, err := svc.GetHistoryTimeline(ctx, f)
	if err != nil {
		t.Fatalf("GetHistoryTimeline user filter: %v", err)
	}
	if len(buckets) != 1 {
		t.Fatalf("buckets = %d, want 1", len(buckets))
	}
	if buckets[0].EventCount != 2 {
		t.Errorf("u1 count = %d, want 2", buckets[0].EventCount)
	}
}

func TestGetHistoryTimelineInvalidRange(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// both zero
	if _, err := svc.GetHistoryTimeline(ctx, history.TimelineFilter{}); !errors.Is(err, history.ErrInvalidTimeRange) {
		t.Errorf("zero bounds: want ErrInvalidTimeRange, got %v", err)
	}
	// since == until
	if _, err := svc.GetHistoryTimeline(ctx, history.TimelineFilter{Since: base, Until: base}); !errors.Is(err, history.ErrInvalidTimeRange) {
		t.Errorf("equal bounds: want ErrInvalidTimeRange, got %v", err)
	}
	// since > until
	if _, err := svc.GetHistoryTimeline(ctx, history.TimelineFilter{Since: base.Add(time.Hour), Until: base}); !errors.Is(err, history.ErrInvalidTimeRange) {
		t.Errorf("inverted bounds: want ErrInvalidTimeRange, got %v", err)
	}
}

func TestGetMyTimelineDay(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	// u1: 2 events on day1, 1 event on day2; u2: 1 event on day1 (must be excluded)
	day1 := time.Date(2024, 6, 10, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2024, 6, 11, 9, 0, 0, 0, time.UTC)
	svc.RecordPlay(ctx, "u1", "t1", day1)
	svc.RecordPlay(ctx, "u1", "t2", day1.Add(2*time.Hour))
	svc.RecordPlay(ctx, "u1", "t1", day2)
	svc.RecordPlay(ctx, "u2", "t3", day1) // different user — must not appear

	f := history.TimelineFilter{
		Since:       time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC),
		Until:       time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC),
		Granularity: history.GranularityDay,
		UserID:      "u1",
	}
	buckets, err := svc.GetMyTimeline(ctx, f)
	if err != nil {
		t.Fatalf("GetMyTimeline: %v", err)
	}
	if len(buckets) != 2 {
		t.Fatalf("buckets = %d, want 2", len(buckets))
	}
	if buckets[0].EventCount != 2 {
		t.Errorf("day1 count = %d, want 2", buckets[0].EventCount)
	}
	if buckets[1].EventCount != 1 {
		t.Errorf("day2 count = %d, want 1", buckets[1].EventCount)
	}
}

func TestGetMyTimelineTrackFilter(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	day := time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC)
	svc.RecordPlay(ctx, "u1", "t1", day)
	svc.RecordPlay(ctx, "u1", "t1", day.Add(time.Hour))
	svc.RecordPlay(ctx, "u1", "t2", day.Add(2*time.Hour)) // different track — excluded

	f := history.TimelineFilter{
		Since:       time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
		Until:       time.Date(2024, 7, 2, 0, 0, 0, 0, time.UTC),
		Granularity: history.GranularityDay,
		UserID:      "u1",
		TrackID:     "t1",
	}
	buckets, err := svc.GetMyTimeline(ctx, f)
	if err != nil {
		t.Fatalf("GetMyTimeline track filter: %v", err)
	}
	if len(buckets) != 1 {
		t.Fatalf("buckets = %d, want 1", len(buckets))
	}
	if buckets[0].EventCount != 2 {
		t.Errorf("t1 count = %d, want 2", buckets[0].EventCount)
	}
}

func TestGetMyTimelineInvalidRange(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// missing userID
	if _, err := svc.GetMyTimeline(ctx, history.TimelineFilter{Since: base, Until: base.Add(time.Hour)}); err == nil {
		t.Error("expected error for empty UserID, got nil")
	}
	// zero bounds with userID
	if _, err := svc.GetMyTimeline(ctx, history.TimelineFilter{UserID: "u1"}); !errors.Is(err, history.ErrInvalidTimeRange) {
		t.Errorf("zero bounds: want ErrInvalidTimeRange, got %v", err)
	}
	// since >= until with userID
	if _, err := svc.GetMyTimeline(ctx, history.TimelineFilter{UserID: "u1", Since: base, Until: base}); !errors.Is(err, history.ErrInvalidTimeRange) {
		t.Errorf("equal bounds: want ErrInvalidTimeRange, got %v", err)
	}
}

func TestGetAllHistory(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	// two users, two tracks
	svc.RecordPlay(ctx, "u1", "t1", now)
	svc.RecordPlay(ctx, "u1", "t2", now.Add(time.Second))
	svc.RecordPlay(ctx, "u2", "t1", now.Add(2*time.Second))

	// no filter → all 3 events
	events, total, err := svc.GetAllHistory(ctx, history.GlobalPlayEventFilter{})
	if err != nil {
		t.Fatalf("GetAllHistory: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(events) != 3 {
		t.Errorf("events len = %d, want 3", len(events))
	}
}

func TestGetAllHistoryUserFilter(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	svc.RecordPlay(ctx, "u1", "t1", now)
	svc.RecordPlay(ctx, "u2", "t1", now.Add(time.Second))
	svc.RecordPlay(ctx, "u2", "t2", now.Add(2*time.Second))

	// filter by u2 → 2 events
	events, total, err := svc.GetAllHistory(ctx, history.GlobalPlayEventFilter{UserID: "u2"})
	if err != nil {
		t.Fatalf("GetAllHistory user filter: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(events) != 2 {
		t.Errorf("events len = %d, want 2", len(events))
	}
	for _, e := range events {
		if e.UserID != "u2" {
			t.Errorf("unexpected userID %q in filtered result", e.UserID)
		}
	}
}

func TestGetAllHistoryTimeWindow(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t2", base.Add(5*time.Second))
	svc.RecordPlay(ctx, "u2", "t3", base.Add(20*time.Second))

	// window [+3s, +10s) captures only t2 played at +5s
	f := history.GlobalPlayEventFilter{
		Since: base.Add(3 * time.Second),
		Until: base.Add(10 * time.Second),
	}
	events, total, err := svc.GetAllHistory(ctx, f)
	if err != nil {
		t.Fatalf("GetAllHistory time window: %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(events) != 1 || events[0].TrackID != "t2" {
		t.Errorf("events = %+v, want single t2 event", events)
	}
}

func TestListPlaysAscOrder(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	// record 3 events oldest→newest
	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t2", base.Add(time.Second))
	svc.RecordPlay(ctx, "u1", "t3", base.Add(2*time.Second))

	// default (desc) → newest first
	desc, _, _ := svc.ListPlays(ctx, history.PlayEventFilter{UserID: "u1"})
	if desc[0].TrackID != "t3" || desc[2].TrackID != "t1" {
		t.Errorf("desc order wrong: got %v %v %v", desc[0].TrackID, desc[1].TrackID, desc[2].TrackID)
	}

	// asc → oldest first
	asc, _, _ := svc.ListPlays(ctx, history.PlayEventFilter{UserID: "u1", Asc: true})
	if asc[0].TrackID != "t1" || asc[2].TrackID != "t3" {
		t.Errorf("asc order wrong: got %v %v %v", asc[0].TrackID, asc[1].TrackID, asc[2].TrackID)
	}
}

func TestGetAllHistoryAscOrder(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u2", "t2", base.Add(time.Second))
	svc.RecordPlay(ctx, "u1", "t3", base.Add(2*time.Second))

	asc, total, err := svc.GetAllHistory(ctx, history.GlobalPlayEventFilter{Asc: true})
	if err != nil {
		t.Fatalf("GetAllHistory asc: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	// oldest event first
	if asc[0].TrackID != "t1" {
		t.Errorf("asc[0] = %q, want t1", asc[0].TrackID)
	}
	if asc[2].TrackID != "t3" {
		t.Errorf("asc[2] = %q, want t3", asc[2].TrackID)
	}
}

func TestGetEventByID(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	e, err := svc.RecordPlay(ctx, "u1", "t1", time.Now().UTC())
	if err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}

	got, err := svc.GetEventByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("GetEventByID: %v", err)
	}
	if got.ID != e.ID || got.TrackID != "t1" || got.UserID != "u1" {
		t.Errorf("GetEventByID returned %+v, want id=%s user=u1 track=t1", got, e.ID)
	}
}

func TestGetEventByIDNotFound(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	if _, err := svc.GetEventByID(context.Background(), "no-such-id"); !errors.Is(err, history.ErrEventNotFound) {
		t.Errorf("want ErrEventNotFound, got %v", err)
	}
}

func TestDeleteEventByID(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	e, _ := svc.RecordPlay(ctx, "u1", "t1", time.Now().UTC())
	if err := svc.DeleteEventByID(ctx, e.ID); err != nil {
		t.Fatalf("DeleteEventByID: %v", err)
	}
	// should be gone
	if _, err := svc.GetEventByID(ctx, e.ID); !errors.Is(err, history.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound after delete, got %v", err)
	}
}

func TestGetMyEvent(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	e, _ := svc.RecordPlay(ctx, "u1", "t1", time.Now().UTC())

	// owner can fetch
	got, err := svc.GetMyEvent(ctx, "u1", e.ID)
	if err != nil {
		t.Fatalf("GetMyEvent: %v", err)
	}
	if got.ID != e.ID {
		t.Errorf("GetMyEvent id = %q, want %q", got.ID, e.ID)
	}

	// different user gets forbidden
	if _, err := svc.GetMyEvent(ctx, "u2", e.ID); !errors.Is(err, history.ErrEventForbidden) {
		t.Errorf("want ErrEventForbidden, got %v", err)
	}
}

func TestDeleteMyEvent(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	e, _ := svc.RecordPlay(ctx, "u1", "t1", time.Now().UTC())

	// wrong user → forbidden
	if err := svc.DeleteMyEvent(ctx, "u2", e.ID); !errors.Is(err, history.ErrEventForbidden) {
		t.Errorf("want ErrEventForbidden, got %v", err)
	}

	// owner can delete
	if err := svc.DeleteMyEvent(ctx, "u1", e.ID); err != nil {
		t.Fatalf("DeleteMyEvent: %v", err)
	}
	if _, err := svc.GetEventByID(ctx, e.ID); !errors.Is(err, history.ErrEventNotFound) {
		t.Errorf("expected ErrEventNotFound after delete, got %v", err)
	}
}

func TestUpdateEventByID(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	e, _ := svc.RecordPlay(ctx, "u1", "t1", time.Now().UTC())
	newTime := time.Date(2020, 6, 1, 10, 0, 0, 0, time.UTC)

	got, err := svc.UpdateEventByID(ctx, e.ID, newTime)
	if err != nil {
		t.Fatalf("UpdateEventByID: %v", err)
	}
	if !got.PlayedAt.Equal(newTime) {
		t.Errorf("playedAt = %v, want %v", got.PlayedAt, newTime)
	}
	if got.ID != e.ID {
		t.Errorf("id = %q, want %q", got.ID, e.ID)
	}
}

func TestUpdateEventByIDNotFound(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	newTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := svc.UpdateEventByID(context.Background(), "no-such", newTime); !errors.Is(err, history.ErrEventNotFound) {
		t.Errorf("want ErrEventNotFound, got %v", err)
	}
}

func TestUpdateMyEvent(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	e, _ := svc.RecordPlay(ctx, "u1", "t1", time.Now().UTC())
	newTime := time.Date(2021, 3, 15, 8, 0, 0, 0, time.UTC)

	got, err := svc.UpdateMyEvent(ctx, "u1", e.ID, newTime)
	if err != nil {
		t.Fatalf("UpdateMyEvent: %v", err)
	}
	if !got.PlayedAt.Equal(newTime) {
		t.Errorf("playedAt = %v, want %v", got.PlayedAt, newTime)
	}

	// wrong owner → forbidden
	if _, err := svc.UpdateMyEvent(ctx, "u2", e.ID, newTime); !errors.Is(err, history.ErrEventForbidden) {
		t.Errorf("want ErrEventForbidden, got %v", err)
	}
}

func TestBatchDeleteEvents(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	e1, _ := svc.RecordPlay(ctx, "u1", "t1", now)
	e2, _ := svc.RecordPlay(ctx, "u1", "t2", now.Add(time.Second))
	e3, _ := svc.RecordPlay(ctx, "u2", "t1", now.Add(2*time.Second))

	// delete e1 and e3; e2 survives
	deleted, err := svc.BatchDeleteEvents(ctx, []string{e1.ID, e3.ID})
	if err != nil {
		t.Fatalf("BatchDeleteEvents: %v", err)
	}
	if deleted != 2 {
		t.Errorf("deleted = %d, want 2", deleted)
	}
	// e2 still present
	got, err := svc.GetEventByID(ctx, e2.ID)
	if err != nil {
		t.Fatalf("GetEventByID after batch delete: %v", err)
	}
	if got.ID != e2.ID {
		t.Errorf("unexpected id %q", got.ID)
	}
}

func TestBatchDeleteEventsUnknownIDsIgnored(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	e, _ := svc.RecordPlay(ctx, "u1", "t1", time.Now().UTC())

	// include one real ID and one unknown
	deleted, err := svc.BatchDeleteEvents(ctx, []string{e.ID, "no-such-id"})
	if err != nil {
		t.Fatalf("BatchDeleteEvents: %v", err)
	}
	if deleted != 1 {
		t.Errorf("deleted = %d, want 1", deleted)
	}
}

func TestBatchDeleteMyEvents(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	now := time.Now().UTC()

	e1, _ := svc.RecordPlay(ctx, "u1", "t1", now)
	e2, _ := svc.RecordPlay(ctx, "u2", "t2", now.Add(time.Second))

	// u1 tries to batch-delete both; only e1 is hers
	deleted, err := svc.BatchDeleteMyEvents(ctx, "u1", []string{e1.ID, e2.ID})
	if err != nil {
		t.Fatalf("BatchDeleteMyEvents: %v", err)
	}
	if deleted != 1 {
		t.Errorf("deleted = %d, want 1 (only own event)", deleted)
	}
	// e2 (owned by u2) still present
	if _, err := svc.GetEventByID(ctx, e2.ID); err != nil {
		t.Errorf("e2 should still exist, got %v", err)
	}
}

func TestBatchDeleteEventsEmpty(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	if _, err := svc.BatchDeleteEvents(context.Background(), nil); err == nil {
		t.Error("expected error for empty ids")
	}
	if _, err := svc.BatchDeleteMyEvents(context.Background(), "u1", nil); err == nil {
		t.Error("expected error for empty ids")
	}
}

func TestListPlaysSinceFilter(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t2", base.Add(5*time.Second))
	svc.RecordPlay(ctx, "u1", "t3", base.Add(20*time.Second))

	// since +10s: only t3
	events, total, err := svc.ListPlays(ctx, history.PlayEventFilter{
		UserID: "u1",
		Since:  base.Add(10 * time.Second),
	})
	if err != nil {
		t.Fatalf("ListPlays since: %v", err)
	}
	if total != 1 || len(events) != 1 || events[0].TrackID != "t3" {
		t.Errorf("since filter: total=%d events=%v, want 1/t3", total, events)
	}
}

func TestListPlaysUntilFilter(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t2", base.Add(5*time.Second))
	svc.RecordPlay(ctx, "u1", "t3", base.Add(20*time.Second))

	// until +10s (exclusive): t1 and t2
	events, total, err := svc.ListPlays(ctx, history.PlayEventFilter{
		UserID: "u1",
		Until:  base.Add(10 * time.Second),
	})
	if err != nil {
		t.Fatalf("ListPlays until: %v", err)
	}
	if total != 2 {
		t.Errorf("until filter: total=%d, want 2", total)
	}
	for _, e := range events {
		if e.TrackID == "t3" {
			t.Error("t3 should be excluded by until filter")
		}
	}
}

func TestGetUserHistorySinceFilter(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t2", base.Add(5*time.Second))
	svc.RecordPlay(ctx, "u1", "t3", base.Add(20*time.Second))

	events, total, err := svc.GetUserHistory(ctx, history.PlayEventFilter{
		UserID: "u1",
		Since:  base.Add(3 * time.Second),
		Until:  base.Add(15 * time.Second),
	})
	if err != nil {
		t.Fatalf("GetUserHistory since/until: %v", err)
	}
	if total != 1 || len(events) != 1 || events[0].TrackID != "t2" {
		t.Errorf("user history window: total=%d, want 1/t2; got %v", total, events)
	}
}

func TestGetMyTrackStatsNoPlays(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	stats, err := svc.GetMyTrackStats(ctx, "u1", "t1", history.UserStatsFilter{})
	if err != nil {
		t.Fatalf("GetMyTrackStats: %v", err)
	}
	if stats.TotalPlays != 0 {
		t.Errorf("TotalPlays = %d, want 0", stats.TotalPlays)
	}
	if !stats.FirstPlayedAt.IsZero() {
		t.Errorf("FirstPlayedAt should be zero when no plays")
	}
}

func TestGetMyTrackStatsWithPlays(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()
	base := time.Now().UTC()

	svc.RecordPlay(ctx, "u1", "t1", base)
	svc.RecordPlay(ctx, "u1", "t1", base.Add(10*time.Second))
	svc.RecordPlay(ctx, "u1", "t2", base.Add(5*time.Second)) // different track
	svc.RecordPlay(ctx, "u2", "t1", base.Add(3*time.Second)) // different user

	stats, err := svc.GetMyTrackStats(ctx, "u1", "t1", history.UserStatsFilter{})
	if err != nil {
		t.Fatalf("GetMyTrackStats: %v", err)
	}
	if stats.TotalPlays != 2 {
		t.Errorf("TotalPlays = %d, want 2", stats.TotalPlays)
	}
	if stats.FirstPlayedAt.IsZero() {
		t.Error("FirstPlayedAt should not be zero")
	}
	if stats.LastPlayedAt.IsZero() {
		t.Error("LastPlayedAt should not be zero")
	}
	if !stats.FirstPlayedAt.Before(stats.LastPlayedAt) {
		t.Errorf("FirstPlayedAt %v should be before LastPlayedAt %v", stats.FirstPlayedAt, stats.LastPlayedAt)
	}
}

func TestGetMyTrackStatsMissingArgs(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	if _, err := svc.GetMyTrackStats(ctx, "", "t1", history.UserStatsFilter{}); err == nil {
		t.Error("expected error for empty userID")
	}
	if _, err := svc.GetMyTrackStats(ctx, "u1", "", history.UserStatsFilter{}); err == nil {
		t.Error("expected error for empty trackID")
	}
}

func TestGetAdminUserSummary(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		if _, err := svc.RecordPlay(ctx, "u-sum-1", "t-sum-1", now.Add(-time.Duration(i)*time.Hour)); err != nil {
			t.Fatalf("RecordPlay: %v", err)
		}
	}
	if _, err := svc.RecordPlay(ctx, "u-sum-1", "t-sum-2", now.Add(-4*time.Hour)); err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}

	summary, err := svc.GetAdminUserSummary(ctx, "u-sum-1", history.UserStatsFilter{}, 10)
	if err != nil {
		t.Fatalf("GetAdminUserSummary: %v", err)
	}
	if summary.Stats.TotalEvents != 4 {
		t.Errorf("Stats.TotalEvents = %d, want 4", summary.Stats.TotalEvents)
	}
	if len(summary.TopTracks) == 0 {
		t.Error("TopTracks should not be empty")
	}
	if summary.TopTracks[0].TrackID != "t-sum-1" {
		t.Errorf("TopTracks[0].TrackID = %q, want t-sum-1", summary.TopTracks[0].TrackID)
	}
}

func TestGetAdminUserSummaryEmpty(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	summary, err := svc.GetAdminUserSummary(ctx, "u-nobody", history.UserStatsFilter{}, 10)
	if err != nil {
		t.Fatalf("GetAdminUserSummary empty: %v", err)
	}
	if summary.Stats.TotalEvents != 0 {
		t.Errorf("expected 0 total events, got %d", summary.Stats.TotalEvents)
	}
	if len(summary.TopTracks) != 0 {
		t.Errorf("expected no top tracks, got %d", len(summary.TopTracks))
	}
}

func TestGetTrackSummary(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	now := time.Now().UTC()
	// u-ts-1 plays track-ts 3 times, u-ts-2 plays it once
	for i := 0; i < 3; i++ {
		if _, err := svc.RecordPlay(ctx, "u-ts-1", "track-ts", now.Add(-time.Duration(i)*time.Hour)); err != nil {
			t.Fatalf("RecordPlay: %v", err)
		}
	}
	if _, err := svc.RecordPlay(ctx, "u-ts-2", "track-ts", now.Add(-4*time.Hour)); err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}

	summary, err := svc.GetTrackSummary(ctx, "track-ts", history.TrackStatsFilter{}, 10)
	if err != nil {
		t.Fatalf("GetTrackSummary: %v", err)
	}
	if summary.Stats.TotalEvents != 4 {
		t.Errorf("Stats.TotalEvents = %d, want 4", summary.Stats.TotalEvents)
	}
	if summary.Stats.UniqueListeners != 2 {
		t.Errorf("Stats.UniqueListeners = %d, want 2", summary.Stats.UniqueListeners)
	}
	if len(summary.TopListeners) == 0 {
		t.Error("TopListeners should not be empty")
	}
	if summary.TopListeners[0].UserID != "u-ts-1" {
		t.Errorf("TopListeners[0].UserID = %q, want u-ts-1", summary.TopListeners[0].UserID)
	}
}

func TestGetTrackSummaryEmpty(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	summary, err := svc.GetTrackSummary(ctx, "track-nobody", history.TrackStatsFilter{}, 10)
	if err != nil {
		t.Fatalf("GetTrackSummary empty: %v", err)
	}
	if summary.Stats.TotalEvents != 0 {
		t.Errorf("expected 0 total events, got %d", summary.Stats.TotalEvents)
	}
	if len(summary.TopListeners) != 0 {
		t.Errorf("expected no top listeners, got %d", len(summary.TopListeners))
	}
}

func TestGetMyTrackStatsTimeWindow(t *testing.T) {
	svc := history.NewService(history.NewMemoryRepository())
	ctx := context.Background()

	day1 := time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 11, 2, 10, 0, 0, 0, time.UTC)
	day3 := time.Date(2025, 11, 3, 10, 0, 0, 0, time.UTC)

	// 2 plays on day1, 1 play on day3
	if _, err := svc.RecordPlay(ctx, "u-tw", "t-tw", day1); err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}
	if _, err := svc.RecordPlay(ctx, "u-tw", "t-tw", day1.Add(time.Hour)); err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}
	if _, err := svc.RecordPlay(ctx, "u-tw", "t-tw", day3); err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}

	// Window covering only day1-day2 (exclusive): should see 2 plays
	stats, err := svc.GetMyTrackStats(ctx, "u-tw", "t-tw", history.UserStatsFilter{
		Since: day1,
		Until: day2,
	})
	if err != nil {
		t.Fatalf("GetMyTrackStats with window: %v", err)
	}
	if stats.TotalPlays != 2 {
		t.Errorf("TotalPlays in window [day1,day2) = %d, want 2", stats.TotalPlays)
	}

	// No window: should see all 3
	all, err := svc.GetMyTrackStats(ctx, "u-tw", "t-tw", history.UserStatsFilter{})
	if err != nil {
		t.Fatalf("GetMyTrackStats no window: %v", err)
	}
	if all.TotalPlays != 3 {
		t.Errorf("TotalPlays all = %d, want 3", all.TotalPlays)
	}
}

func TestGetGlobalSummary(t *testing.T) {
	ctx := context.Background()
	svc := history.NewService(history.NewMemoryRepository())

	if _, err := svc.RecordPlay(ctx, "u1", "t1", time.Now()); err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}
	if _, err := svc.RecordPlay(ctx, "u1", "t2", time.Now()); err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}
	if _, err := svc.RecordPlay(ctx, "u2", "t1", time.Now()); err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}

	summary, err := svc.GetGlobalSummary(ctx, history.StatsFilter{}, 10)
	if err != nil {
		t.Fatalf("GetGlobalSummary: %v", err)
	}
	if summary.Stats.TotalEvents != 3 {
		t.Errorf("TotalEvents = %d, want 3", summary.Stats.TotalEvents)
	}
	if summary.Stats.UniqueUsers != 2 {
		t.Errorf("UniqueUsers = %d, want 2", summary.Stats.UniqueUsers)
	}
	if summary.Stats.UniqueTracks != 2 {
		t.Errorf("UniqueTracks = %d, want 2", summary.Stats.UniqueTracks)
	}
	if len(summary.TopTracks) == 0 {
		t.Error("TopTracks is empty, want at least 1")
	}
	if len(summary.TopUsers) == 0 {
		t.Error("TopUsers is empty, want at least 1")
	}
}

func TestGetGlobalSummaryWithTopN(t *testing.T) {
	ctx := context.Background()
	svc := history.NewService(history.NewMemoryRepository())

	// Record plays for 5 tracks so topN=2 truncates the result.
	for i := 0; i < 5; i++ {
		trackID := "t-gn-" + string(rune('a'+i))
		if _, err := svc.RecordPlay(ctx, "u-gn", trackID, time.Now()); err != nil {
			t.Fatalf("RecordPlay track %s: %v", trackID, err)
		}
	}

	summary, err := svc.GetGlobalSummary(ctx, history.StatsFilter{}, 2)
	if err != nil {
		t.Fatalf("GetGlobalSummary topN=2: %v", err)
	}
	if len(summary.TopTracks) > 2 {
		t.Errorf("len(TopTracks) = %d, want ≤ 2", len(summary.TopTracks))
	}
	if len(summary.TopUsers) > 2 {
		t.Errorf("len(TopUsers) = %d, want ≤ 2", len(summary.TopUsers))
	}
}

func TestGetMyTrackSummary(t *testing.T) {
	ctx := context.Background()
	svc := history.NewService(history.NewMemoryRepository())

	// 2 plays on t-mts, 1 play on another track for context.
	if _, err := svc.RecordPlay(ctx, "u-mts", "t-mts", time.Now()); err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}
	if _, err := svc.RecordPlay(ctx, "u-mts", "t-mts", time.Now()); err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}
	if _, err := svc.RecordPlay(ctx, "u-mts", "t-other", time.Now()); err != nil {
		t.Fatalf("RecordPlay: %v", err)
	}

	summary, err := svc.GetMyTrackSummary(ctx, "u-mts", "t-mts", history.UserStatsFilter{}, 10)
	if err != nil {
		t.Fatalf("GetMyTrackSummary: %v", err)
	}
	if summary.Stats.TotalPlays != 2 {
		t.Errorf("Stats.TotalPlays = %d, want 2", summary.Stats.TotalPlays)
	}
	if summary.Stats.TrackID != "t-mts" {
		t.Errorf("Stats.TrackID = %q, want t-mts", summary.Stats.TrackID)
	}
	if len(summary.RecentTracks) == 0 {
		t.Error("RecentTracks is empty, want at least 1")
	}
}

func TestGetMyTrackSummaryEmpty(t *testing.T) {
	ctx := context.Background()
	svc := history.NewService(history.NewMemoryRepository())

	// No plays recorded — summary should return zero stats and empty recent tracks.
	summary, err := svc.GetMyTrackSummary(ctx, "u-mts-empty", "t-mts-empty", history.UserStatsFilter{}, 10)
	if err != nil {
		t.Fatalf("GetMyTrackSummary empty: %v", err)
	}
	if summary.Stats.TotalPlays != 0 {
		t.Errorf("Stats.TotalPlays = %d, want 0", summary.Stats.TotalPlays)
	}
	if summary.RecentTracks == nil {
		t.Error("RecentTracks is nil, want empty slice")
	}
}

func TestGetAdminTrackStatsTimeWindow(t *testing.T) {
	ctx := context.Background()
	svc := history.NewService(history.NewMemoryRepository())

	day1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)
	day3 := time.Date(2025, 1, 3, 12, 0, 0, 0, time.UTC)

	// 2 plays on day1, 1 play on day3 — all for track t-atw
	for _, ts := range []time.Time{day1, day1.Add(time.Hour), day3} {
		if _, err := svc.RecordPlay(ctx, "u-atw", "t-atw", ts); err != nil {
			t.Fatalf("RecordPlay: %v", err)
		}
	}

	// Window [day1, day2) — should see 2 plays
	stats, err := svc.GetTrackStats(ctx, history.TrackStatsFilter{
		TrackID: "t-atw",
		Since:   day1,
		Until:   day2,
	})
	if err != nil {
		t.Fatalf("GetTrackStats with window: %v", err)
	}
	if stats.TotalEvents != 2 {
		t.Errorf("TotalEvents in [day1,day2) = %d, want 2", stats.TotalEvents)
	}

	// No window — should see all 3
	all, err := svc.GetTrackStats(ctx, history.TrackStatsFilter{TrackID: "t-atw"})
	if err != nil {
		t.Fatalf("GetTrackStats no window: %v", err)
	}
	if all.TotalEvents != 3 {
		t.Errorf("TotalEvents all = %d, want 3", all.TotalEvents)
	}
}

func TestGetHistoryStatsTimeWindow(t *testing.T) {
	ctx := context.Background()
	svc := history.NewService(history.NewMemoryRepository())

	day1 := time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 3, 2, 12, 0, 0, 0, time.UTC)
	day3 := time.Date(2025, 3, 3, 12, 0, 0, 0, time.UTC)

	if _, err := svc.RecordPlay(ctx, "u-hstw", "t-hstw", day1); err != nil {
		t.Fatalf("RecordPlay day1: %v", err)
	}
	if _, err := svc.RecordPlay(ctx, "u-hstw", "t-hstw", day1.Add(time.Hour)); err != nil {
		t.Fatalf("RecordPlay day1+1h: %v", err)
	}
	if _, err := svc.RecordPlay(ctx, "u-hstw", "t-hstw", day3); err != nil {
		t.Fatalf("RecordPlay day3: %v", err)
	}

	// Window [day1, day2) — should see 2 events
	stats, err := svc.GetHistoryStats(ctx, history.StatsFilter{Since: day1, Until: day2})
	if err != nil {
		t.Fatalf("GetHistoryStats with window: %v", err)
	}
	if stats.TotalEvents != 2 {
		t.Errorf("TotalEvents in [day1,day2) = %d, want 2", stats.TotalEvents)
	}

	// No window — should see all 3
	all, err := svc.GetHistoryStats(ctx, history.StatsFilter{})
	if err != nil {
		t.Fatalf("GetHistoryStats no window: %v", err)
	}
	if all.TotalEvents != 3 {
		t.Errorf("TotalEvents all = %d, want 3", all.TotalEvents)
	}
}

func TestGetMyTopTracksTimeWindowPhase114(t *testing.T) {
	ctx := context.Background()
	svc := history.NewService(history.NewMemoryRepository())

	day1 := time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 5, 2, 12, 0, 0, 0, time.UTC)
	day3 := time.Date(2025, 5, 3, 12, 0, 0, 0, time.UTC)

	// 2 plays for t-a on day1, 1 play for t-b on day3
	if _, err := svc.RecordPlay(ctx, "u-mtt", "t-a", day1); err != nil {
		t.Fatalf("RecordPlay t-a day1: %v", err)
	}
	if _, err := svc.RecordPlay(ctx, "u-mtt", "t-a", day1.Add(time.Hour)); err != nil {
		t.Fatalf("RecordPlay t-a day1+1h: %v", err)
	}
	if _, err := svc.RecordPlay(ctx, "u-mtt", "t-b", day3); err != nil {
		t.Fatalf("RecordPlay t-b day3: %v", err)
	}

	// Window [day1, day2) — only t-a should appear
	tracks, err := svc.GetMyTopTracks(ctx, history.UserStatsFilter{
		UserID: "u-mtt",
		Since:  day1,
		Until:  day2,
	}, 10)
	if err != nil {
		t.Fatalf("GetMyTopTracks with window: %v", err)
	}
	if len(tracks) != 1 || tracks[0].TrackID != "t-a" {
		t.Errorf("GetMyTopTracks [day1,day2) = %v, want [{t-a 2}]", tracks)
	}

	// No window — both tracks should appear
	allTracks, err := svc.GetMyTopTracks(ctx, history.UserStatsFilter{UserID: "u-mtt"}, 10)
	if err != nil {
		t.Fatalf("GetMyTopTracks no window: %v", err)
	}
	if len(allTracks) != 2 {
		t.Errorf("GetMyTopTracks all = %d tracks, want 2", len(allTracks))
	}
}

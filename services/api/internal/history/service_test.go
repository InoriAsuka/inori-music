package history_test

import (
	"context"
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

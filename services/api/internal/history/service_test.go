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

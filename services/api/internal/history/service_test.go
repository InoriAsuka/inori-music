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

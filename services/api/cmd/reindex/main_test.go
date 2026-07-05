package main

import (
	"context"
	"errors"
	"testing"
)

func TestReindexWalkSinglePage(t *testing.T) {
	ctx := context.Background()
	var calls int
	fetchPage := func(_ context.Context, _ string, limit, offset int) (entityPage, error) {
		calls++
		// First call returns 3 of 5; second call (offset=2 by item count) should be empty to terminate.
		if offset >= 3 {
			return entityPage{items: nil, total: 5}, nil
		}
		return entityPage{items: []any{"a", "b", "c"}, total: 5}, nil
	}
	var indexed []any
	ok, failed := reindexWalk(ctx, "test", fetchPage, "id", func(e any) string { return e.(string) }, func(e any) error {
		indexed = append(indexed, e)
		return nil
	})
	if ok != 3 || failed != 0 {
		t.Fatalf("ok=%d failed=%d, want 3/0", ok, failed)
	}
	if len(indexed) != 3 {
		t.Fatalf("indexed=%d, want 3", len(indexed))
	}
	if calls != 2 {
		t.Fatalf("fetchPage called %d times, want 2 (one with data, one empty)", calls)
	}
}

func TestReindexWalkStopsWhenOffsetReachesTotal(t *testing.T) {
	ctx := context.Background()
	totalCalls := 0
	fetchPage := func(_ context.Context, _ string, limit, offset int) (entityPage, error) {
		totalCalls++
		// Always return a full page; termination relies on offset>=total, not emptiness.
		return entityPage{items: []any{"x", "y"}, total: 2}, nil
	}
	ok, failed := reindexWalk(ctx, "test", fetchPage, "id", func(e any) string { return "x" }, func(e any) error { return nil })
	if totalCalls != 1 {
		t.Fatalf("fetchPage called %d times, want 1 (should break after offset>=total)", totalCalls)
	}
	if ok != 2 || failed != 0 {
		t.Fatalf("ok=%d failed=%d, want 2/0", ok, failed)
	}
}

func TestReindexWalkCountsFailuresSeparately(t *testing.T) {
	ctx := context.Background()
	fetchPage := func(_ context.Context, _ string, limit, offset int) (entityPage, error) {
		if offset > 0 {
			return entityPage{items: nil, total: 3}, nil
		}
		return entityPage{items: []any{"ok", "fail", "ok2"}, total: 3}, nil
	}
	ok, failed := reindexWalk(ctx, "test", fetchPage, "id", func(e any) string { return e.(string) }, func(e any) error {
		if e.(string) == "fail" {
			return errors.New("boom")
		}
		return nil
	})
	if ok != 2 || failed != 1 {
		t.Fatalf("ok=%d failed=%d, want 2/1", ok, failed)
	}
}

func TestReindexWalkEmptyFirstPage(t *testing.T) {
	ctx := context.Background()
	var calls int
	fetchPage := func(_ context.Context, _ string, limit, offset int) (entityPage, error) {
		calls++
		return entityPage{items: nil, total: 0}, nil
	}
	ok, failed := reindexWalk(ctx, "test", fetchPage, "id", func(e any) string { return "" }, func(e any) error { return nil })
	if ok != 0 || failed != 0 {
		t.Fatalf("ok=%d failed=%d, want 0/0", ok, failed)
	}
	if calls != 1 {
		t.Fatalf("fetchPage called %d times, want 1", calls)
	}
}

func TestReindexWalkMultiPage(t *testing.T) {
	ctx := context.Background()
	var pageCalls int
	// offset increments by item count: 0 → 2 → 4 → 5, terminates at offset>=total.
	fetchPage := func(_ context.Context, _ string, limit, offset int) (entityPage, error) {
		pageCalls++
		switch offset {
		case 0:
			return entityPage{items: []any{"a", "b"}, total: 5}, nil
		case 2:
			return entityPage{items: []any{"c", "d"}, total: 5}, nil
		case 4:
			return entityPage{items: []any{"e"}, total: 5}, nil
		default:
			return entityPage{items: nil, total: 5}, nil
		}
	}
	var collected []any
	ok, failed := reindexWalk(ctx, "test", fetchPage, "id", func(e any) string { return e.(string) }, func(e any) error {
		collected = append(collected, e)
		return nil
	})
	if ok != 5 || failed != 0 {
		t.Fatalf("ok=%d failed=%d, want 5/0", ok, failed)
	}
	if len(collected) != 5 {
		t.Fatalf("collected=%d, want 5", len(collected))
	}
	// 3 non-empty pages; the 4th call at offset=5 with total=5 hits offset>=total and breaks.
	if pageCalls != 3 {
		t.Fatalf("expected 3 page calls to terminate via offset>=total, got %d", pageCalls)
	}
}

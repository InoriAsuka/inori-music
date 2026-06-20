//go:build integration

package historypg_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"inori-music/services/api/internal/history"
	historypg "inori-music/services/api/internal/history/postgres"
	pgstore "inori-music/services/api/internal/storage/postgres"
)

func setupHistoryTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("inori_test"),
		tcpostgres.WithUsername("inori"),
		tcpostgres.WithPassword("inori"),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	t.Cleanup(pool.Close)

	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}
	defer conn.Release()
	if err := pgstore.Migrate(ctx, conn.Conn()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

// TestRepositoryHistoryStats verifies that HistoryStats correctly counts
// total events, unique users, and unique tracks after recording play events.
func TestRepositoryHistoryStats(t *testing.T) {
	pool := setupHistoryTestDB(t)
	repo := historypg.NewRepository(pool)
	ctx := context.Background()

	now := time.Now().UTC()

	events := []history.PlayEvent{
		{ID: "e1", UserID: "u1", TrackID: "t1", PlayedAt: now, CreatedAt: now},
		{ID: "e2", UserID: "u1", TrackID: "t2", PlayedAt: now, CreatedAt: now},
		{ID: "e3", UserID: "u2", TrackID: "t1", PlayedAt: now, CreatedAt: now},
	}
	for _, e := range events {
		if err := repo.SavePlayEvent(ctx, e); err != nil {
			t.Fatalf("SavePlayEvent %s: %v", e.ID, err)
		}
	}

	stats, err := repo.HistoryStats(ctx, history.StatsFilter{})
	if err != nil {
		t.Fatalf("HistoryStats: %v", err)
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

// TestRepositoryTopTracks verifies that TopTracks returns tracks ordered by
// descending play count.
func TestRepositoryTopTracks(t *testing.T) {
	pool := setupHistoryTestDB(t)
	repo := historypg.NewRepository(pool)
	ctx := context.Background()

	now := time.Now().UTC()

	// t1 played 3 times, t2 played 1 time
	for _, ev := range []history.PlayEvent{
		{ID: "tt1", UserID: "u1", TrackID: "t1", PlayedAt: now, CreatedAt: now},
		{ID: "tt2", UserID: "u1", TrackID: "t1", PlayedAt: now, CreatedAt: now},
		{ID: "tt3", UserID: "u2", TrackID: "t1", PlayedAt: now, CreatedAt: now},
		{ID: "tt4", UserID: "u1", TrackID: "t2", PlayedAt: now, CreatedAt: now},
	} {
		if err := repo.SavePlayEvent(ctx, ev); err != nil {
			t.Fatalf("SavePlayEvent: %v", err)
		}
	}

	top, err := repo.TopTracks(ctx, history.StatsFilter{}, 10)
	if err != nil {
		t.Fatalf("TopTracks: %v", err)
	}
	if len(top) != 2 {
		t.Fatalf("len(top) = %d, want 2", len(top))
	}
	if top[0].TrackID != "t1" || top[0].PlayCount != 3 {
		t.Errorf("top[0] = %+v, want {t1, 3}", top[0])
	}
	if top[1].TrackID != "t2" || top[1].PlayCount != 1 {
		t.Errorf("top[1] = %+v, want {t2, 1}", top[1])
	}
}

// TestRepositoryUserHistoryStats verifies that UserHistoryStats correctly
// counts events and unique tracks scoped to a single user.
func TestRepositoryUserHistoryStats(t *testing.T) {
	pool := setupHistoryTestDB(t)
	repo := historypg.NewRepository(pool)
	ctx := context.Background()

	now := time.Now().UTC()

	for _, ev := range []history.PlayEvent{
		{ID: "us1", UserID: "ua", TrackID: "ta", PlayedAt: now, CreatedAt: now},
		{ID: "us2", UserID: "ua", TrackID: "tb", PlayedAt: now, CreatedAt: now},
		{ID: "us3", UserID: "ua", TrackID: "ta", PlayedAt: now, CreatedAt: now},
		{ID: "us4", UserID: "ub", TrackID: "ta", PlayedAt: now, CreatedAt: now}, // other user — must not appear
	} {
		if err := repo.SavePlayEvent(ctx, ev); err != nil {
			t.Fatalf("SavePlayEvent: %v", err)
		}
	}

	stats, err := repo.UserHistoryStats(ctx, history.UserStatsFilter{UserID: "ua"})
	if err != nil {
		t.Fatalf("UserHistoryStats: %v", err)
	}
	if stats.TotalEvents != 3 {
		t.Errorf("TotalEvents = %d, want 3", stats.TotalEvents)
	}
	if stats.UniqueTracks != 2 {
		t.Errorf("UniqueTracks = %d, want 2", stats.UniqueTracks)
	}
}

// TestRepositoryTrackHistoryStats verifies that TrackHistoryStats correctly
// counts events and unique listeners for a given track.
func TestRepositoryTrackHistoryStats(t *testing.T) {
	pool := setupHistoryTestDB(t)
	repo := historypg.NewRepository(pool)
	ctx := context.Background()

	now := time.Now().UTC()

	for _, ev := range []history.PlayEvent{
		{ID: "ts1", UserID: "u1", TrackID: "tgt", PlayedAt: now, CreatedAt: now},
		{ID: "ts2", UserID: "u2", TrackID: "tgt", PlayedAt: now, CreatedAt: now},
		{ID: "ts3", UserID: "u1", TrackID: "tgt", PlayedAt: now, CreatedAt: now},
		{ID: "ts4", UserID: "u1", TrackID: "other", PlayedAt: now, CreatedAt: now}, // different track
	} {
		if err := repo.SavePlayEvent(ctx, ev); err != nil {
			t.Fatalf("SavePlayEvent: %v", err)
		}
	}

	stats, err := repo.TrackHistoryStats(ctx, history.TrackStatsFilter{TrackID: "tgt"})
	if err != nil {
		t.Fatalf("TrackHistoryStats: %v", err)
	}
	if stats.TotalEvents != 3 {
		t.Errorf("TotalEvents = %d, want 3", stats.TotalEvents)
	}
	if stats.UniqueListeners != 2 {
		t.Errorf("UniqueListeners = %d, want 2", stats.UniqueListeners)
	}
}

// TestRepositoryHistoryTimeline verifies that HistoryTimeline groups events
// into daily buckets and respects Since/Until bounds.
func TestRepositoryHistoryTimeline(t *testing.T) {
	pool := setupHistoryTestDB(t)
	repo := historypg.NewRepository(pool)
	ctx := context.Background()

	day1 := time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 12, 2, 10, 0, 0, 0, time.UTC)
	day3 := time.Date(2025, 12, 3, 10, 0, 0, 0, time.UTC)

	for _, ev := range []history.PlayEvent{
		{ID: "tl1", UserID: "u1", TrackID: "t1", PlayedAt: day1, CreatedAt: day1},
		{ID: "tl2", UserID: "u1", TrackID: "t1", PlayedAt: day1, CreatedAt: day1},
		{ID: "tl3", UserID: "u1", TrackID: "t1", PlayedAt: day2, CreatedAt: day2},
		// day3 event is outside the Since/Until window
		{ID: "tl4", UserID: "u1", TrackID: "t1", PlayedAt: day3, CreatedAt: day3},
	} {
		if err := repo.SavePlayEvent(ctx, ev); err != nil {
			t.Fatalf("SavePlayEvent: %v", err)
		}
	}

	since := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2025, 12, 3, 0, 0, 0, 0, time.UTC)

	buckets, err := repo.HistoryTimeline(ctx, history.TimelineFilter{
		Since:       since,
		Until:       until,
		Granularity: history.GranularityDay,
	})
	if err != nil {
		t.Fatalf("HistoryTimeline: %v", err)
	}
	if len(buckets) != 2 {
		t.Fatalf("len(buckets) = %d, want 2", len(buckets))
	}
	if buckets[0].EventCount != 2 {
		t.Errorf("day1 eventCount = %d, want 2", buckets[0].EventCount)
	}
	if buckets[1].EventCount != 1 {
		t.Errorf("day2 eventCount = %d, want 1", buckets[1].EventCount)
	}
}

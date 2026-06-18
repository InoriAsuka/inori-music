package history

import (
	"context"
	"sort"
	"sync"
)

// MemoryRepository is a thread-safe in-memory history repository for tests and development.
type MemoryRepository struct {
	mu     sync.RWMutex
	events map[string]PlayEvent
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{events: make(map[string]PlayEvent)}
}

func (r *MemoryRepository) SavePlayEvent(_ context.Context, e PlayEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[e.ID] = e
	return nil
}

func (r *MemoryRepository) ListPlayEvents(_ context.Context, f PlayEventFilter) ([]PlayEvent, int, error) {
	r.mu.RLock()
	var all []PlayEvent
	for _, e := range r.events {
		if e.UserID != f.UserID {
			continue
		}
		if f.TrackID != "" && e.TrackID != f.TrackID {
			continue
		}
		all = append(all, e)
	}
	r.mu.RUnlock()

	sort.SliceStable(all, func(i, j int) bool {
		return all[i].PlayedAt.After(all[j].PlayedAt)
	})

	total := len(all)
	start := f.Offset
	if start >= total {
		return []PlayEvent{}, total, nil
	}
	end := start + f.Limit
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (r *MemoryRepository) DeletePlayEventsByUser(_ context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, e := range r.events {
		if e.UserID == userID {
			delete(r.events, id)
		}
	}
	return nil
}

func (r *MemoryRepository) DeletePlayEventsByUserAdmin(_ context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, e := range r.events {
		if e.UserID == userID {
			delete(r.events, id)
		}
	}
	return nil
}

func (r *MemoryRepository) DeletePlayEventsByTrack(_ context.Context, trackID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, e := range r.events {
		if e.TrackID == trackID {
			delete(r.events, id)
		}
	}
	return nil
}

func (r *MemoryRepository) DeletePlayEventsInWindow(_ context.Context, f StatsFilter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, e := range r.events {
		if !f.Since.IsZero() && e.PlayedAt.Before(f.Since) {
			continue
		}
		if !f.Until.IsZero() && !e.PlayedAt.Before(f.Until) {
			continue
		}
		delete(r.events, id)
	}
	return nil
}

func (r *MemoryRepository) ListPlayEventsByTrack(_ context.Context, f AdminPlayEventFilter) ([]PlayEvent, int, error) {
	r.mu.RLock()
	var all []PlayEvent
	for _, e := range r.events {
		if e.TrackID != f.TrackID {
			continue
		}
		if f.UserID != "" && e.UserID != f.UserID {
			continue
		}
		all = append(all, e)
	}
	r.mu.RUnlock()

	sort.SliceStable(all, func(i, j int) bool {
		if all[i].PlayedAt.Equal(all[j].PlayedAt) {
			return all[i].ID < all[j].ID
		}
		return all[i].PlayedAt.After(all[j].PlayedAt)
	})

	total := len(all)
	start := f.Offset
	if start >= total {
		return []PlayEvent{}, total, nil
	}
	end := start + f.Limit
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (r *MemoryRepository) ListAllPlayEvents(_ context.Context, f GlobalPlayEventFilter) ([]PlayEvent, int, error) {
	r.mu.RLock()
	var all []PlayEvent
	for _, e := range r.events {
		if f.UserID != "" && e.UserID != f.UserID {
			continue
		}
		if f.TrackID != "" && e.TrackID != f.TrackID {
			continue
		}
		if !f.Since.IsZero() && e.PlayedAt.Before(f.Since) {
			continue
		}
		if !f.Until.IsZero() && !e.PlayedAt.Before(f.Until) {
			continue
		}
		all = append(all, e)
	}
	r.mu.RUnlock()

	sort.SliceStable(all, func(i, j int) bool {
		if all[i].PlayedAt.Equal(all[j].PlayedAt) {
			return all[i].ID < all[j].ID
		}
		return all[i].PlayedAt.After(all[j].PlayedAt)
	})

	total := len(all)
	start := f.Offset
	if start >= total {
		return []PlayEvent{}, total, nil
	}
	end := start + f.Limit
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (r *MemoryRepository) HistoryStats(_ context.Context, f StatsFilter) (HistoryStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make(map[string]struct{})
	tracks := make(map[string]struct{})
	total := 0
	for _, e := range r.events {
		if !f.Since.IsZero() && e.PlayedAt.Before(f.Since) {
			continue
		}
		if !f.Until.IsZero() && !e.PlayedAt.Before(f.Until) {
			continue
		}
		users[e.UserID] = struct{}{}
		tracks[e.TrackID] = struct{}{}
		total++
	}
	return HistoryStats{
		TotalEvents:  total,
		UniqueUsers:  len(users),
		UniqueTracks: len(tracks),
	}, nil
}

func (r *MemoryRepository) TopTracks(_ context.Context, f StatsFilter, limit int) ([]TrackPlayCount, error) {
	r.mu.RLock()
	counts := make(map[string]int)
	for _, e := range r.events {
		if !f.Since.IsZero() && e.PlayedAt.Before(f.Since) {
			continue
		}
		if !f.Until.IsZero() && !e.PlayedAt.Before(f.Until) {
			continue
		}
		counts[e.TrackID]++
	}
	r.mu.RUnlock()

	result := make([]TrackPlayCount, 0, len(counts))
	for trackID, n := range counts {
		result = append(result, TrackPlayCount{TrackID: trackID, PlayCount: n})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].PlayCount != result[j].PlayCount {
			return result[i].PlayCount > result[j].PlayCount
		}
		return result[i].TrackID < result[j].TrackID
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *MemoryRepository) TopUsers(_ context.Context, f StatsFilter, limit int) ([]UserPlayCount, error) {
	r.mu.RLock()
	counts := make(map[string]int)
	for _, e := range r.events {
		if !f.Since.IsZero() && e.PlayedAt.Before(f.Since) {
			continue
		}
		if !f.Until.IsZero() && !e.PlayedAt.Before(f.Until) {
			continue
		}
		counts[e.UserID]++
	}
	r.mu.RUnlock()

	result := make([]UserPlayCount, 0, len(counts))
	for userID, n := range counts {
		result = append(result, UserPlayCount{UserID: userID, PlayCount: n})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].PlayCount != result[j].PlayCount {
			return result[i].PlayCount > result[j].PlayCount
		}
		return result[i].UserID < result[j].UserID
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *MemoryRepository) UserTopTracks(_ context.Context, f UserStatsFilter, limit int) ([]TrackPlayCount, error) {
	r.mu.RLock()
	counts := make(map[string]int)
	for _, e := range r.events {
		if e.UserID != f.UserID {
			continue
		}
		if !f.Since.IsZero() && e.PlayedAt.Before(f.Since) {
			continue
		}
		if !f.Until.IsZero() && !e.PlayedAt.Before(f.Until) {
			continue
		}
		counts[e.TrackID]++
	}
	r.mu.RUnlock()

	result := make([]TrackPlayCount, 0, len(counts))
	for trackID, n := range counts {
		result = append(result, TrackPlayCount{TrackID: trackID, PlayCount: n})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].PlayCount != result[j].PlayCount {
			return result[i].PlayCount > result[j].PlayCount
		}
		return result[i].TrackID < result[j].TrackID
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *MemoryRepository) UserHistoryStats(_ context.Context, f UserStatsFilter) (UserHistoryStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tracks := make(map[string]struct{})
	total := 0
	for _, e := range r.events {
		if e.UserID != f.UserID {
			continue
		}
		if !f.Since.IsZero() && e.PlayedAt.Before(f.Since) {
			continue
		}
		if !f.Until.IsZero() && !e.PlayedAt.Before(f.Until) {
			continue
		}
		tracks[e.TrackID] = struct{}{}
		total++
	}
	return UserHistoryStats{
		TotalEvents:  total,
		UniqueTracks: len(tracks),
	}, nil
}

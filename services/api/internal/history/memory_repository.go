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

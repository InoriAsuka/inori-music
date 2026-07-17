package searchhistory

import (
	"context"
	"sort"
	"sync"
)

// MemoryRepository is an in-memory search history repository for tests and development.
type MemoryRepository struct {
	mu      sync.RWMutex
	entries map[string][]Entry
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{entries: make(map[string][]Entry)}
}

func (r *MemoryRepository) List(_ context.Context, userID string) ([]Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stored := r.entries[userID]
	out := make([]Entry, len(stored))
	copy(out, stored)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].SearchedAt.After(out[j].SearchedAt)
	})
	if len(out) > MaxEntries {
		out = out[:MaxEntries]
	}
	return out, nil
}

func (r *MemoryRepository) Replace(_ context.Context, userID string, entries []Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := make([]Entry, len(entries))
	copy(stored, entries)
	r.entries[userID] = stored
	return nil
}

func (r *MemoryRepository) Clear(_ context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, userID)
	return nil
}

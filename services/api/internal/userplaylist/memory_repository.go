package userplaylist

import (
	"context"
	"sort"
	"sync"
)

// MemoryRepository is an in-memory user playlist repository for tests and development.
type MemoryRepository struct {
	mu    sync.RWMutex
	store map[string]UserPlaylist // keyed by playlist ID
}

// NewMemoryRepository returns an empty MemoryRepository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{store: make(map[string]UserPlaylist)}
}

func (r *MemoryRepository) Save(_ context.Context, p UserPlaylist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[p.ID] = p
	return nil
}

func (r *MemoryRepository) Get(_ context.Context, id string) (UserPlaylist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.store[id]
	if !ok {
		return UserPlaylist{}, ErrNotFound
	}
	return p, nil
}

func (r *MemoryRepository) ListByUser(_ context.Context, userID string) ([]UserPlaylist, error) {
	r.mu.RLock()
	var result []UserPlaylist
	for _, p := range r.store {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	r.mu.RUnlock()

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	if result == nil {
		result = []UserPlaylist{}
	}
	return result, nil
}

func (r *MemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.store, id)
	return nil
}

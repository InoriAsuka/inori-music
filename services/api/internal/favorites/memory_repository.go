package favorites

import (
	"context"
	"sort"
	"sync"
	"time"
)

type favoriteKey struct {
	userID  string
	trackID string
}

// MemoryRepository is an in-memory favorites repository for tests and development.
type MemoryRepository struct {
	mu      sync.RWMutex
	entries map[favoriteKey]FavoriteEntry
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{entries: make(map[favoriteKey]FavoriteEntry)}
}

func (r *MemoryRepository) AddFavorite(_ context.Context, userID, trackID string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := favoriteKey{userID, trackID}
	if _, ok := r.entries[key]; !ok {
		r.entries[key] = FavoriteEntry{UserID: userID, TrackID: trackID, CreatedAt: now}
	}
	return nil
}

func (r *MemoryRepository) RemoveFavorite(_ context.Context, userID, trackID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, favoriteKey{userID, trackID})
	return nil
}

func (r *MemoryRepository) ListFavorites(_ context.Context, userID string, limit, offset int) (FavoritesPage, error) {
	r.mu.RLock()
	var all []FavoriteEntry
	for key, e := range r.entries {
		if key.userID == userID {
			all = append(all, e)
		}
	}
	r.mu.RUnlock()

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	total := len(all)
	if offset >= total {
		return FavoritesPage{TrackIDs: []string{}, Total: total}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	ids := make([]string, 0, end-offset)
	for _, e := range all[offset:end] {
		ids = append(ids, e.TrackID)
	}
	return FavoritesPage{TrackIDs: ids, Total: total}, nil
}

func (r *MemoryRepository) IsFavorite(_ context.Context, userID, trackID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.entries[favoriteKey{userID, trackID}]
	return ok, nil
}

func (r *MemoryRepository) AreFavorites(_ context.Context, userID string, trackIDs []string) (map[string]bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]bool, len(trackIDs))
	for _, tid := range trackIDs {
		_, ok := r.entries[favoriteKey{userID, tid}]
		result[tid] = ok
	}
	return result, nil
}

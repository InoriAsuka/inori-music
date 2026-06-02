package storage

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// Repository stores server-managed storage backend configuration.
type Repository interface {
	Save(ctx context.Context, backend StorageBackend) error
	Get(ctx context.Context, id string) (StorageBackend, error)
	List(ctx context.Context) ([]StorageBackend, error)
	ClearDefault(ctx context.Context) error
}

// MemoryRepository is a development repository for domain tests and early scaffolding.
type MemoryRepository struct {
	mu       sync.RWMutex
	backends map[string]StorageBackend
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{backends: make(map[string]StorageBackend)}
}

func (repo *MemoryRepository) Save(_ context.Context, backend StorageBackend) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if backend.ID == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidBackend)
	}
	repo.backends[backend.ID] = backend
	return nil
}

func (repo *MemoryRepository) Get(_ context.Context, id string) (StorageBackend, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	backend, ok := repo.backends[id]
	if !ok {
		return StorageBackend{}, fmt.Errorf("%w: %s", ErrNotFound, id)
	}
	return backend, nil
}

func (repo *MemoryRepository) List(_ context.Context) ([]StorageBackend, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	backends := make([]StorageBackend, 0, len(repo.backends))
	for _, backend := range repo.backends {
		backends = append(backends, backend)
	}
	sort.Slice(backends, func(i, j int) bool {
		if backends[i].Priority == backends[j].Priority {
			return backends[i].ID < backends[j].ID
		}
		return backends[i].Priority < backends[j].Priority
	})
	return backends, nil
}

func (repo *MemoryRepository) ClearDefault(_ context.Context) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for id, backend := range repo.backends {
		backend.IsDefault = false
		repo.backends[id] = backend
	}
	return nil
}

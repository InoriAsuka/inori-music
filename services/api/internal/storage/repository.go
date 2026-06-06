package storage

import (
	"context"
	"fmt"
	"sync"
)

// Repository stores server-managed storage backend configuration.
type Repository interface {
	Save(ctx context.Context, backend StorageBackend) error
	SaveWithExclusiveDefault(ctx context.Context, backend StorageBackend) error
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

func (repo *MemoryRepository) SaveWithExclusiveDefault(_ context.Context, backend StorageBackend) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if backend.ID == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidBackend)
	}
	backend.IsDefault = true
	for id, existing := range repo.backends {
		existing.IsDefault = false
		repo.backends[id] = existing
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

	return sortedBackends(repo.backends), nil
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

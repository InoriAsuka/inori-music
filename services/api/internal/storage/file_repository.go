package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const fileRepositorySchemaVersion = 1

// FileRepository stores backend configuration in an atomically rewritten JSON file.
type FileRepository struct {
	mu       sync.RWMutex
	path     string
	backends map[string]StorageBackend
}

type fileRepositoryDocument struct {
	Version  int              `json:"version"`
	Backends []StorageBackend `json:"backends"`
}

// NewFileRepository loads a durable JSON-backed repository from path.
func NewFileRepository(path string) (*FileRepository, error) {
	if path == "" {
		return nil, fmt.Errorf("%w: repository file path is required", ErrInvalidBackend)
	}
	repo := &FileRepository{path: path, backends: make(map[string]StorageBackend)}
	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (repo *FileRepository) Save(ctx context.Context, backend StorageBackend) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if backend.ID == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidBackend)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	repo.backends[backend.ID] = backend
	return repo.persistLocked()
}

func (repo *FileRepository) Get(ctx context.Context, id string) (StorageBackend, error) {
	if err := ctx.Err(); err != nil {
		return StorageBackend{}, err
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	backend, ok := repo.backends[id]
	if !ok {
		return StorageBackend{}, fmt.Errorf("%w: %s", ErrNotFound, id)
	}
	return backend, nil
}

func (repo *FileRepository) List(ctx context.Context) ([]StorageBackend, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	return sortedBackends(repo.backends), nil
}

func (repo *FileRepository) ClearDefault(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	for id, backend := range repo.backends {
		backend.IsDefault = false
		repo.backends[id] = backend
	}
	return repo.persistLocked()
}

func (repo *FileRepository) load() error {
	content, err := os.ReadFile(repo.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("load storage repository %q: %w", repo.path, err)
	}
	if len(content) == 0 {
		return nil
	}

	var document fileRepositoryDocument
	if err := json.Unmarshal(content, &document); err != nil {
		return fmt.Errorf("decode storage repository %q: %w", repo.path, err)
	}
	if document.Version != fileRepositorySchemaVersion {
		return fmt.Errorf("%w: unsupported repository schema version %d", ErrInvalidBackend, document.Version)
	}
	seen := make(map[string]struct{}, len(document.Backends))
	for _, backend := range document.Backends {
		if err := ValidateBackend(&backend); err != nil {
			return fmt.Errorf("validate persisted backend %q: %w", backend.ID, err)
		}
		if _, ok := seen[backend.ID]; ok {
			return fmt.Errorf("%w: duplicate persisted backend id %q", ErrInvalidBackend, backend.ID)
		}
		seen[backend.ID] = struct{}{}
		repo.backends[backend.ID] = backend
	}
	return nil
}

func (repo *FileRepository) persistLocked() error {
	dir := filepath.Dir(repo.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create storage repository directory %q: %w", dir, err)
	}

	temp, err := os.CreateTemp(dir, filepath.Base(repo.path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary storage repository file: %w", err)
	}
	tempPath := temp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tempPath)
		}
	}()

	encoder := json.NewEncoder(temp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(fileRepositoryDocument{Version: fileRepositorySchemaVersion, Backends: sortedBackends(repo.backends)}); err != nil {
		_ = temp.Close()
		return fmt.Errorf("encode storage repository: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync storage repository: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close storage repository: %w", err)
	}
	if err := os.Rename(tempPath, repo.path); err != nil {
		return fmt.Errorf("replace storage repository %q: %w", repo.path, err)
	}
	cleanup = false
	return nil
}

func sortedBackends(backends map[string]StorageBackend) []StorageBackend {
	result := make([]StorageBackend, 0, len(backends))
	for _, backend := range backends {
		result = append(result, backend)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Priority == result[j].Priority {
			return result[i].ID < result[j].ID
		}
		return result[i].Priority < result[j].Priority
	})
	return result
}

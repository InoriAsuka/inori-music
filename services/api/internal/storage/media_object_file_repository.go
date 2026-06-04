package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const mediaObjectFileRepositorySchemaVersion = 1

// FileMediaObjectRepository stores media object metadata in an atomically rewritten JSON file.
type FileMediaObjectRepository struct {
	mu      sync.RWMutex
	path    string
	objects map[string]MediaObject
}

type mediaObjectFileRepositoryDocument struct {
	Version int           `json:"version"`
	Objects []MediaObject `json:"objects"`
}

// NewFileMediaObjectRepository loads a durable JSON-backed media object repository from path.
func NewFileMediaObjectRepository(path string) (*FileMediaObjectRepository, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("%w: media object repository file path is required", ErrInvalidMediaObject)
	}
	repo := &FileMediaObjectRepository{path: path, objects: make(map[string]MediaObject)}
	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (repo *FileMediaObjectRepository) SaveMediaObject(ctx context.Context, object MediaObject) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(object.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidMediaObject)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	repo.objects[object.ID] = object
	return repo.persistLocked()
}

func (repo *FileMediaObjectRepository) GetMediaObject(ctx context.Context, id string) (MediaObject, error) {
	if err := ctx.Err(); err != nil {
		return MediaObject{}, err
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	object, ok := repo.objects[id]
	if !ok {
		return MediaObject{}, fmt.Errorf("%w: media object %s", ErrNotFound, id)
	}
	return object, nil
}

func (repo *FileMediaObjectRepository) ListAllMediaObjects(ctx context.Context) ([]MediaObject, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	objects := make([]MediaObject, 0, len(repo.objects))
	for _, object := range repo.objects {
		objects = append(objects, object)
	}
	return sortedMediaObjects(objects), nil
}

func (repo *FileMediaObjectRepository) ListMediaObjectsByBackend(ctx context.Context, backendID string) ([]MediaObject, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	objects := make([]MediaObject, 0)
	for _, object := range repo.objects {
		if object.BackendID == backendID {
			objects = append(objects, object)
		}
	}
	return sortedMediaObjects(objects), nil
}

func (repo *FileMediaObjectRepository) ListMediaObjectsByContentHash(ctx context.Context, contentHash string) ([]MediaObject, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	objects := make([]MediaObject, 0)
	for _, object := range repo.objects {
		if object.ContentHash == contentHash {
			objects = append(objects, object)
		}
	}
	return sortedMediaObjects(objects), nil
}

func (repo *FileMediaObjectRepository) ListMediaObjectsByVerificationStatus(ctx context.Context, status string) ([]MediaObject, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	objects := make([]MediaObject, 0)
	for _, object := range repo.objects {
		if mediaObjectVerificationStatus(object) == strings.TrimSpace(status) {
			objects = append(objects, object)
		}
	}
	return sortedMediaObjects(objects), nil
}

func (repo *FileMediaObjectRepository) ListMediaObjectsByLifecycleState(ctx context.Context, state string) ([]MediaObject, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	objects := make([]MediaObject, 0)
	for _, object := range repo.objects {
		if object.LifecycleState == strings.TrimSpace(state) {
			objects = append(objects, object)
		}
	}
	return sortedMediaObjects(objects), nil
}

func (repo *FileMediaObjectRepository) load() error {
	content, err := os.ReadFile(repo.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("load media object repository %q: %w", repo.path, err)
	}
	if len(content) == 0 {
		return nil
	}

	var document mediaObjectFileRepositoryDocument
	if err := json.Unmarshal(content, &document); err != nil {
		return fmt.Errorf("decode media object repository %q: %w", repo.path, err)
	}
	if document.Version != mediaObjectFileRepositorySchemaVersion {
		return fmt.Errorf("%w: unsupported media object repository schema version %d", ErrInvalidMediaObject, document.Version)
	}
	for _, object := range document.Objects {
		if strings.TrimSpace(object.ID) == "" {
			return fmt.Errorf("%w: persisted media object id is required", ErrInvalidMediaObject)
		}
		repo.objects[object.ID] = object
	}
	return nil
}

func (repo *FileMediaObjectRepository) persistLocked() error {
	dir := filepath.Dir(repo.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create media object repository directory %q: %w", dir, err)
	}

	temp, err := os.CreateTemp(dir, filepath.Base(repo.path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary media object repository file: %w", err)
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
	if err := encoder.Encode(mediaObjectFileRepositoryDocument{Version: mediaObjectFileRepositorySchemaVersion, Objects: sortedMediaObjectsFromMap(repo.objects)}); err != nil {
		_ = temp.Close()
		return fmt.Errorf("encode media object repository: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync media object repository: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close media object repository: %w", err)
	}
	if err := os.Rename(tempPath, repo.path); err != nil {
		return fmt.Errorf("replace media object repository %q: %w", repo.path, err)
	}
	cleanup = false
	return nil
}

func sortedMediaObjectsFromMap(objects map[string]MediaObject) []MediaObject {
	result := make([]MediaObject, 0, len(objects))
	for _, object := range objects {
		result = append(result, object)
	}
	return sortedMediaObjects(result)
}

package storage

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrInvalidMediaObject = errors.New("invalid media object")

type AssetKind string

const (
	AssetKindOriginalAudio   AssetKind = "original_audio"
	AssetKindTranscodedAudio AssetKind = "transcoded_audio"
	AssetKindArtwork         AssetKind = "artwork"
	AssetKindLyrics          AssetKind = "lyrics"
	AssetKindWaveform        AssetKind = "waveform"
	AssetKindAnalysis        AssetKind = "analysis"
	AssetKindImportPackage   AssetKind = "import_package"
	AssetKindBackup          AssetKind = "backup"
)

type LifecycleState string

const (
	LifecycleStateStaged   LifecycleState = "staged"
	LifecycleStateActive   LifecycleState = "active"
	LifecycleStateArchived LifecycleState = "archived"
	LifecycleStateDeleted  LifecycleState = "deleted"
)

// MediaObjectRepository stores metadata references for binary assets.
type MediaObjectRepository interface {
	SaveMediaObject(ctx context.Context, object MediaObject) error
	GetMediaObject(ctx context.Context, id string) (MediaObject, error)
	ListMediaObjectsByBackend(ctx context.Context, backendID string) ([]MediaObject, error)
	ListMediaObjectsByContentHash(ctx context.Context, contentHash string) ([]MediaObject, error)
}

// MemoryMediaObjectRepository is a development repository for media object metadata.
type MemoryMediaObjectRepository struct {
	mu      sync.RWMutex
	objects map[string]MediaObject
}

func NewMemoryMediaObjectRepository() *MemoryMediaObjectRepository {
	return &MemoryMediaObjectRepository{objects: make(map[string]MediaObject)}
}

func (repo *MemoryMediaObjectRepository) SaveMediaObject(_ context.Context, object MediaObject) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if strings.TrimSpace(object.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidMediaObject)
	}
	repo.objects[object.ID] = object
	return nil
}

func (repo *MemoryMediaObjectRepository) GetMediaObject(_ context.Context, id string) (MediaObject, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	object, ok := repo.objects[id]
	if !ok {
		return MediaObject{}, fmt.Errorf("%w: media object %s", ErrNotFound, id)
	}
	return object, nil
}

func (repo *MemoryMediaObjectRepository) ListMediaObjectsByBackend(_ context.Context, backendID string) ([]MediaObject, error) {
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

func (repo *MemoryMediaObjectRepository) ListMediaObjectsByContentHash(_ context.Context, contentHash string) ([]MediaObject, error) {
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

// MediaObjectService coordinates media object metadata registration rules.
type MediaObjectService struct {
	backendRepository Repository
	mediaRepository   MediaObjectRepository
	verifier          MediaObjectVerifier
	now               func() time.Time
}

func NewMediaObjectService(backendRepository Repository, mediaRepository MediaObjectRepository) *MediaObjectService {
	return &MediaObjectService{backendRepository: backendRepository, mediaRepository: mediaRepository, verifier: NewFilesystemMediaObjectVerifier(), now: time.Now}
}

func (service *MediaObjectService) RegisterMediaObject(ctx context.Context, object MediaObject) (MediaObject, error) {
	if err := ValidateMediaObject(&object); err != nil {
		return MediaObject{}, err
	}
	backend, err := service.backendRepository.Get(ctx, object.BackendID)
	if err != nil {
		return MediaObject{}, err
	}
	if !backend.Enabled {
		return MediaObject{}, fmt.Errorf("%w: backend %s", ErrBackendDisabled, object.BackendID)
	}
	if _, err := service.mediaRepository.GetMediaObject(ctx, object.ID); err == nil {
		return MediaObject{}, fmt.Errorf("%w: media object %s already exists", ErrConflict, object.ID)
	} else if !errors.Is(err, ErrNotFound) {
		return MediaObject{}, err
	}
	now := service.now().UTC()
	object.CreatedAt = now
	object.UpdatedAt = now
	if err := service.mediaRepository.SaveMediaObject(ctx, object); err != nil {
		return MediaObject{}, err
	}
	return object, nil
}

func (service *MediaObjectService) GetMediaObject(ctx context.Context, id string) (MediaObject, error) {
	return service.mediaRepository.GetMediaObject(ctx, id)
}

func (service *MediaObjectService) ListMediaObjectsByBackend(ctx context.Context, backendID string) ([]MediaObject, error) {
	return service.mediaRepository.ListMediaObjectsByBackend(ctx, backendID)
}

func (service *MediaObjectService) ListMediaObjectsByContentHash(ctx context.Context, contentHash string) ([]MediaObject, error) {
	return service.mediaRepository.ListMediaObjectsByContentHash(ctx, strings.TrimSpace(contentHash))
}

func (service *MediaObjectService) VerifyMediaObject(ctx context.Context, id string) (MediaObjectVerificationResult, error) {
	object, err := service.mediaRepository.GetMediaObject(ctx, id)
	if err != nil {
		return MediaObjectVerificationResult{}, err
	}
	return service.verifyMediaObject(ctx, object)
}

func (service *MediaObjectService) VerifyMediaObjectsByBackend(ctx context.Context, backendID string) (MediaObjectVerificationReport, error) {
	objects, err := service.mediaRepository.ListMediaObjectsByBackend(ctx, strings.TrimSpace(backendID))
	if err != nil {
		return MediaObjectVerificationReport{}, err
	}
	return service.verifyMediaObjects(ctx, objects), nil
}

func (service *MediaObjectService) VerifyMediaObjectsByContentHash(ctx context.Context, contentHash string) (MediaObjectVerificationReport, error) {
	objects, err := service.mediaRepository.ListMediaObjectsByContentHash(ctx, strings.TrimSpace(contentHash))
	if err != nil {
		return MediaObjectVerificationReport{}, err
	}
	return service.verifyMediaObjects(ctx, objects), nil
}

func (service *MediaObjectService) verifyMediaObjects(ctx context.Context, objects []MediaObject) MediaObjectVerificationReport {
	checkedAt := service.now().UTC()
	report := MediaObjectVerificationReport{CheckedAt: checkedAt, Results: make([]MediaObjectVerificationResult, 0, len(objects))}
	for _, object := range objects {
		result, err := service.verifyMediaObject(ctx, object)
		if result.VerifiedAt.IsZero() {
			result.VerifiedAt = checkedAt
		}
		if err != nil && result.Message == "" {
			result.Message = err.Error()
		}
		report.Results = append(report.Results, result)
	}
	return report
}

func (service *MediaObjectService) verifyMediaObject(ctx context.Context, object MediaObject) (MediaObjectVerificationResult, error) {
	backend, err := service.backendRepository.Get(ctx, object.BackendID)
	if err != nil {
		return MediaObjectVerificationResult{MediaObjectID: object.ID, BackendID: object.BackendID, ObjectKey: object.ObjectKey, Status: "failed", VerifiedAt: service.now().UTC(), Message: err.Error()}, err
	}
	if !backend.Enabled {
		return MediaObjectVerificationResult{MediaObjectID: object.ID, BackendID: object.BackendID, ObjectKey: object.ObjectKey, Status: "failed", VerifiedAt: service.now().UTC()}, fmt.Errorf("%w: backend %s", ErrBackendDisabled, object.BackendID)
	}
	result, err := service.verifier.Verify(ctx, backend, object)
	result.VerifiedAt = service.now().UTC()
	if result.MediaObjectID == "" {
		result.MediaObjectID = object.ID
		result.BackendID = object.BackendID
		result.ObjectKey = object.ObjectKey
	}
	if err != nil {
		result.Status = "failed"
		result.Message = err.Error()
		return result, err
	}
	return result, nil
}

// ValidateMediaObject checks static media object metadata before persistence.
func ValidateMediaObject(object *MediaObject) error {
	if object == nil {
		return fmt.Errorf("%w: object is required", ErrInvalidMediaObject)
	}
	object.ID = strings.TrimSpace(object.ID)
	object.BackendID = strings.TrimSpace(object.BackendID)
	object.ObjectKey = strings.TrimSpace(object.ObjectKey)
	object.ContentHash = strings.TrimSpace(object.ContentHash)
	object.MIMEType = strings.TrimSpace(object.MIMEType)
	object.AssetKind = strings.TrimSpace(object.AssetKind)
	object.LifecycleState = strings.TrimSpace(object.LifecycleState)

	if object.ID == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidMediaObject)
	}
	if object.BackendID == "" {
		return fmt.Errorf("%w: backend id is required", ErrInvalidMediaObject)
	}
	if err := validateObjectKey(object.ObjectKey); err != nil {
		return err
	}
	if err := validateContentHash(object.ContentHash); err != nil {
		return err
	}
	if object.SizeBytes < 0 {
		return fmt.Errorf("%w: size bytes must be non-negative", ErrInvalidMediaObject)
	}
	if object.MIMEType == "" || !strings.Contains(object.MIMEType, "/") {
		return fmt.Errorf("%w: mime type must be type/subtype", ErrInvalidMediaObject)
	}
	if !validAssetKind(AssetKind(object.AssetKind)) {
		return fmt.Errorf("%w: unsupported asset kind %q", ErrInvalidMediaObject, object.AssetKind)
	}
	if !validLifecycleState(LifecycleState(object.LifecycleState)) {
		return fmt.Errorf("%w: unsupported lifecycle state %q", ErrInvalidMediaObject, object.LifecycleState)
	}
	return nil
}

func validateObjectKey(key string) error {
	if key == "" {
		return fmt.Errorf("%w: object key is required", ErrInvalidMediaObject)
	}
	if strings.HasPrefix(key, "/") || strings.Contains(key, "\\") {
		return fmt.Errorf("%w: object key must be a relative slash-delimited key", ErrInvalidMediaObject)
	}
	clean := path.Clean(key)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || clean != key {
		return fmt.Errorf("%w: object key must be clean and must not traverse parent directories", ErrInvalidMediaObject)
	}
	return nil
}

func validateContentHash(contentHash string) error {
	algorithm, value, ok := strings.Cut(contentHash, ":")
	if !ok || strings.TrimSpace(algorithm) == "" || strings.TrimSpace(value) == "" {
		return fmt.Errorf("%w: content hash must use algorithm:value format", ErrInvalidMediaObject)
	}
	if strings.ContainsAny(algorithm, " /\\") || strings.ContainsAny(value, " /\\") {
		return fmt.Errorf("%w: content hash must not contain spaces or path separators", ErrInvalidMediaObject)
	}
	return nil
}

func validAssetKind(kind AssetKind) bool {
	switch kind {
	case AssetKindOriginalAudio, AssetKindTranscodedAudio, AssetKindArtwork, AssetKindLyrics, AssetKindWaveform, AssetKindAnalysis, AssetKindImportPackage, AssetKindBackup:
		return true
	default:
		return false
	}
}

func validLifecycleState(state LifecycleState) bool {
	switch state {
	case LifecycleStateStaged, LifecycleStateActive, LifecycleStateArchived, LifecycleStateDeleted:
		return true
	default:
		return false
	}
}

func sortedMediaObjects(objects []MediaObject) []MediaObject {
	sort.Slice(objects, func(i, j int) bool {
		if objects[i].BackendID == objects[j].BackendID {
			if objects[i].ObjectKey == objects[j].ObjectKey {
				return objects[i].ID < objects[j].ID
			}
			return objects[i].ObjectKey < objects[j].ObjectKey
		}
		return objects[i].BackendID < objects[j].BackendID
	})
	return objects
}

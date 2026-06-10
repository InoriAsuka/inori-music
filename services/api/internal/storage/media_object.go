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

const (
	DefaultMediaObjectListLimit = 100
	MaxMediaObjectListLimit     = 500
)

// MediaObjectListFilter describes a metadata-only media object list query.
type MediaObjectListFilter struct {
	BackendID          string
	ContentHash        string
	VerificationStatus string
	LifecycleState     string
	AssetKind          string
	Limit              int
	Offset             int
}

// MediaObjectListPage contains a bounded list response and pagination metadata.
type MediaObjectListPage struct {
	Objects    []MediaObject      `json:"objects"`
	Pagination PaginationMetadata `json:"pagination"`
}

// PaginationMetadata describes offset pagination for admin list responses.
type PaginationMetadata struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Total   int  `json:"total"`
	HasMore bool `json:"hasMore"`
}

// MediaObjectStats summarizes metadata-only counts for admin dashboards.
type MediaObjectStats struct {
	BackendID            string         `json:"backendId,omitempty"`
	TotalObjects         int            `json:"totalObjects"`
	TotalSizeBytes       int64          `json:"totalSizeBytes"`
	ByBackendID          map[string]int `json:"byBackendId"`
	ByAssetKind          map[string]int `json:"byAssetKind"`
	ByLifecycleState     map[string]int `json:"byLifecycleState"`
	ByVerificationStatus map[string]int `json:"byVerificationStatus"`
}

// MediaObjectRepository stores metadata references for binary assets.
type MediaObjectRepository interface {
	SaveMediaObject(ctx context.Context, object MediaObject) error
	GetMediaObject(ctx context.Context, id string) (MediaObject, error)
	ListAllMediaObjects(ctx context.Context) ([]MediaObject, error)
	ListMediaObjectsByBackend(ctx context.Context, backendID string) ([]MediaObject, error)
	ListMediaObjectsByContentHash(ctx context.Context, contentHash string) ([]MediaObject, error)
	ListMediaObjectsByVerificationStatus(ctx context.Context, status string) ([]MediaObject, error)
	ListMediaObjectsByLifecycleState(ctx context.Context, state string) ([]MediaObject, error)
	ListMediaObjectsByAssetKind(ctx context.Context, kind string) ([]MediaObject, error)
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

func (repo *MemoryMediaObjectRepository) ListAllMediaObjects(_ context.Context) ([]MediaObject, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	objects := make([]MediaObject, 0, len(repo.objects))
	for _, object := range repo.objects {
		objects = append(objects, object)
	}
	return sortedMediaObjects(objects), nil
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

func (repo *MemoryMediaObjectRepository) ListMediaObjectsByVerificationStatus(_ context.Context, status string) ([]MediaObject, error) {
	status = strings.TrimSpace(status)
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	objects := make([]MediaObject, 0)
	for _, object := range repo.objects {
		if mediaObjectVerificationStatus(object) == status {
			objects = append(objects, object)
		}
	}
	return sortedMediaObjects(objects), nil
}

func (repo *MemoryMediaObjectRepository) ListMediaObjectsByLifecycleState(_ context.Context, state string) ([]MediaObject, error) {
	state = strings.TrimSpace(state)
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	objects := make([]MediaObject, 0)
	for _, object := range repo.objects {
		if object.LifecycleState == state {
			objects = append(objects, object)
		}
	}
	return sortedMediaObjects(objects), nil
}

func (repo *MemoryMediaObjectRepository) ListMediaObjectsByAssetKind(_ context.Context, kind string) ([]MediaObject, error) {
	kind = strings.TrimSpace(kind)
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	objects := make([]MediaObject, 0)
	for _, object := range repo.objects {
		if object.AssetKind == kind {
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

func (service *MediaObjectService) SetMediaObjectLifecycleState(ctx context.Context, id string, state string) (MediaObject, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return MediaObject{}, fmt.Errorf("%w: id is required", ErrInvalidMediaObject)
	}
	state, err := normalizeLifecycleState(state)
	if err != nil {
		return MediaObject{}, err
	}
	object, err := service.mediaRepository.GetMediaObject(ctx, id)
	if err != nil {
		return MediaObject{}, err
	}
	if LifecycleState(object.LifecycleState) == LifecycleStateDeleted && LifecycleState(state) != LifecycleStateDeleted {
		return MediaObject{}, fmt.Errorf("%w: deleted media object %s cannot leave deleted lifecycle state", ErrConflict, id)
	}
	object.LifecycleState = state
	object.UpdatedAt = service.now().UTC()
	if err := service.mediaRepository.SaveMediaObject(ctx, object); err != nil {
		return MediaObject{}, err
	}
	return object, nil
}

func (service *MediaObjectService) ListMediaObjectsByBackend(ctx context.Context, backendID string) ([]MediaObject, error) {
	return service.mediaRepository.ListMediaObjectsByBackend(ctx, backendID)
}

func (service *MediaObjectService) ListMediaObjectsByContentHash(ctx context.Context, contentHash string) ([]MediaObject, error) {
	return service.mediaRepository.ListMediaObjectsByContentHash(ctx, strings.TrimSpace(contentHash))
}

func (service *MediaObjectService) ListMediaObjectsByVerificationStatus(ctx context.Context, status string) ([]MediaObject, error) {
	normalized, err := normalizeMediaObjectVerificationStatus(status)
	if err != nil {
		return nil, err
	}
	return service.mediaRepository.ListMediaObjectsByVerificationStatus(ctx, normalized)
}

func (service *MediaObjectService) ListMediaObjectsByLifecycleState(ctx context.Context, state string) ([]MediaObject, error) {
	normalized, err := normalizeLifecycleState(state)
	if err != nil {
		return nil, err
	}
	return service.mediaRepository.ListMediaObjectsByLifecycleState(ctx, normalized)
}

func (service *MediaObjectService) ListMediaObjectsByAssetKind(ctx context.Context, kind string) ([]MediaObject, error) {
	normalized, err := normalizeAssetKind(kind)
	if err != nil {
		return nil, err
	}
	return service.mediaRepository.ListMediaObjectsByAssetKind(ctx, normalized)
}

func (service *MediaObjectService) ListMediaObjects(ctx context.Context, filter MediaObjectListFilter) (MediaObjectListPage, error) {
	normalized, err := normalizeMediaObjectListFilter(filter)
	if err != nil {
		return MediaObjectListPage{}, err
	}

	var objects []MediaObject
	if normalized.BackendID != "" {
		objects, err = service.mediaRepository.ListMediaObjectsByBackend(ctx, normalized.BackendID)
	} else if normalized.ContentHash != "" {
		objects, err = service.mediaRepository.ListMediaObjectsByContentHash(ctx, normalized.ContentHash)
	} else if normalized.VerificationStatus != "" {
		objects, err = service.mediaRepository.ListMediaObjectsByVerificationStatus(ctx, normalized.VerificationStatus)
	} else if normalized.LifecycleState != "" {
		objects, err = service.mediaRepository.ListMediaObjectsByLifecycleState(ctx, normalized.LifecycleState)
	} else {
		objects, err = service.mediaRepository.ListMediaObjectsByAssetKind(ctx, normalized.AssetKind)
	}
	if err != nil {
		return MediaObjectListPage{}, err
	}
	return paginateMediaObjects(objects, normalized.Limit, normalized.Offset), nil
}

func (service *MediaObjectService) GetMediaObjectStats(ctx context.Context, backendID string) (MediaObjectStats, error) {
	backendID = strings.TrimSpace(backendID)
	var (
		objects []MediaObject
		err     error
	)
	if backendID != "" {
		objects, err = service.mediaRepository.ListMediaObjectsByBackend(ctx, backendID)
	} else {
		objects, err = service.mediaRepository.ListAllMediaObjects(ctx)
	}
	if err != nil {
		return MediaObjectStats{}, err
	}
	return summarizeMediaObjects(backendID, objects), nil
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
		result := MediaObjectVerificationResult{MediaObjectID: object.ID, BackendID: object.BackendID, ObjectKey: object.ObjectKey, Status: "failed", VerifiedAt: service.now().UTC(), Message: err.Error()}
		if saveErr := service.recordMediaObjectVerification(ctx, object, result); saveErr != nil {
			return result, errors.Join(err, saveErr)
		}
		return result, err
	}
	if !backend.Enabled {
		err := fmt.Errorf("%w: backend %s", ErrBackendDisabled, object.BackendID)
		result := MediaObjectVerificationResult{MediaObjectID: object.ID, BackendID: object.BackendID, ObjectKey: object.ObjectKey, Status: "failed", VerifiedAt: service.now().UTC(), Message: err.Error()}
		if saveErr := service.recordMediaObjectVerification(ctx, object, result); saveErr != nil {
			return result, errors.Join(err, saveErr)
		}
		return result, err
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
		if saveErr := service.recordMediaObjectVerification(ctx, object, result); saveErr != nil {
			return result, errors.Join(err, saveErr)
		}
		return result, err
	}
	if saveErr := service.recordMediaObjectVerification(ctx, object, result); saveErr != nil {
		return result, saveErr
	}
	return result, nil
}

func (service *MediaObjectService) recordMediaObjectVerification(ctx context.Context, object MediaObject, result MediaObjectVerificationResult) error {
	object.LastVerification = &result
	object.UpdatedAt = result.VerifiedAt
	return service.mediaRepository.SaveMediaObject(ctx, object)
}

func normalizeMediaObjectListFilter(filter MediaObjectListFilter) (MediaObjectListFilter, error) {
	filter.BackendID = strings.TrimSpace(filter.BackendID)
	filter.ContentHash = strings.TrimSpace(filter.ContentHash)
	filter.VerificationStatus = strings.TrimSpace(filter.VerificationStatus)
	filter.LifecycleState = strings.TrimSpace(filter.LifecycleState)
	filter.AssetKind = strings.TrimSpace(filter.AssetKind)
	filterCount := 0
	for _, value := range []string{filter.BackendID, filter.ContentHash, filter.VerificationStatus, filter.LifecycleState, filter.AssetKind} {
		if value != "" {
			filterCount++
		}
	}
	if filterCount != 1 {
		return MediaObjectListFilter{}, fmt.Errorf("%w: exactly one of backendId, contentHash, verificationStatus, lifecycleState, or assetKind is required", ErrInvalidMediaObject)
	}
	if filter.VerificationStatus != "" {
		normalized, err := normalizeMediaObjectVerificationStatus(filter.VerificationStatus)
		if err != nil {
			return MediaObjectListFilter{}, err
		}
		filter.VerificationStatus = normalized
	}
	if filter.LifecycleState != "" {
		normalized, err := normalizeLifecycleState(filter.LifecycleState)
		if err != nil {
			return MediaObjectListFilter{}, err
		}
		filter.LifecycleState = normalized
	}
	if filter.AssetKind != "" {
		normalized, err := normalizeAssetKind(filter.AssetKind)
		if err != nil {
			return MediaObjectListFilter{}, err
		}
		filter.AssetKind = normalized
	}
	if filter.Limit == 0 {
		filter.Limit = DefaultMediaObjectListLimit
	}
	if filter.Limit < 0 || filter.Limit > MaxMediaObjectListLimit {
		return MediaObjectListFilter{}, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidMediaObject, MaxMediaObjectListLimit)
	}
	if filter.Offset < 0 {
		return MediaObjectListFilter{}, fmt.Errorf("%w: offset must be non-negative", ErrInvalidMediaObject)
	}
	return filter, nil
}

func normalizeAssetKind(kind string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(kind))
	if !validAssetKind(AssetKind(normalized)) {
		return "", fmt.Errorf("%w: assetKind must be one of original_audio, transcoded_audio, artwork, lyrics, waveform, analysis, import_package, or backup", ErrInvalidMediaObject)
	}
	return normalized, nil
}

func normalizeLifecycleState(state string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(state))
	if !validLifecycleState(LifecycleState(normalized)) {
		return "", fmt.Errorf("%w: lifecycleState must be one of staged, active, archived, or deleted", ErrInvalidMediaObject)
	}
	return normalized, nil
}

func normalizeMediaObjectVerificationStatus(status string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "verified", "failed", "unknown":
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: verificationStatus must be one of verified, failed, or unknown", ErrInvalidMediaObject)
	}
}

func summarizeMediaObjects(backendID string, objects []MediaObject) MediaObjectStats {
	stats := MediaObjectStats{
		BackendID:            backendID,
		ByBackendID:          make(map[string]int),
		ByAssetKind:          make(map[string]int),
		ByLifecycleState:     make(map[string]int),
		ByVerificationStatus: map[string]int{"verified": 0, "failed": 0, "unknown": 0},
	}
	for _, object := range objects {
		stats.TotalObjects++
		stats.TotalSizeBytes += object.SizeBytes
		stats.ByBackendID[object.BackendID]++
		stats.ByAssetKind[object.AssetKind]++
		stats.ByLifecycleState[object.LifecycleState]++
		stats.ByVerificationStatus[mediaObjectVerificationStatus(object)]++
	}
	return stats
}

func paginateMediaObjects(objects []MediaObject, limit int, offset int) MediaObjectListPage {
	total := len(objects)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	pageObjects := make([]MediaObject, end-offset)
	copy(pageObjects, objects[offset:end])
	return MediaObjectListPage{
		Objects: pageObjects,
		Pagination: PaginationMetadata{
			Limit:   limit,
			Offset:  offset,
			Total:   total,
			HasMore: end < total,
		},
	}
}

func mediaObjectVerificationStatus(object MediaObject) string {
	if object.LastVerification == nil || strings.TrimSpace(object.LastVerification.Status) == "" {
		return "unknown"
	}
	return strings.ToLower(strings.TrimSpace(object.LastVerification.Status))
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

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
	DefaultMediaObjectSortBy    = "backend_object_key"
	DefaultMediaObjectSortOrder = "asc"
)

const (
	MediaObjectSortByBackendObjectKey = "backend_object_key"
	MediaObjectSortByCreatedAt        = "created_at"
	MediaObjectSortByUpdatedAt        = "updated_at"
	MediaObjectSortBySizeBytes        = "size_bytes"
	MediaObjectSortByObjectKey        = "object_key"
	MediaObjectSortByID               = "id"
)

const (
	MediaObjectSortOrderAscending  = "asc"
	MediaObjectSortOrderDescending = "desc"
)

// MediaObjectSelectionFilter describes a metadata-only media object selection.
type MediaObjectSelectionFilter struct {
	BackendID          string
	ContentHash        string
	VerificationStatus string
	LifecycleState     string
	AssetKind          string
}

// MediaObjectListFilter describes a metadata-only media object list query.
type MediaObjectListFilter struct {
	BackendID          string
	ContentHash        string
	VerificationStatus string
	LifecycleState     string
	AssetKind          string
	SortBy             string
	SortOrder          string
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

// MediaObjectDuplicateReport describes metadata-only duplicate content-hash groups.
type MediaObjectDuplicateReport struct {
	BackendID      string                      `json:"backendId,omitempty"`
	MinCopies      int                         `json:"minCopies"`
	TotalGroups    int                         `json:"totalGroups"`
	TotalObjects   int                         `json:"totalObjects"`
	TotalSizeBytes int64                       `json:"totalSizeBytes"`
	Groups         []MediaObjectDuplicateGroup `json:"groups"`
}

// MediaObjectDuplicateGroup contains media objects sharing the same content hash.
type MediaObjectDuplicateGroup struct {
	ContentHash    string        `json:"contentHash"`
	Count          int           `json:"count"`
	TotalSizeBytes int64         `json:"totalSizeBytes"`
	Objects        []MediaObject `json:"objects"`
}

// MediaObjectLifecycleUpdateOptions controls metadata-only bulk lifecycle behavior.
type MediaObjectLifecycleUpdateOptions struct {
	DryRun bool
}

// MediaObjectLifecycleChange captures the latest committed lifecycle metadata change.
type MediaObjectLifecycleChange struct {
	PreviousLifecycleState string    `json:"previousLifecycleState"`
	LifecycleState         string    `json:"lifecycleState"`
	ChangedAt              time.Time `json:"changedAt"`
	Source                 string    `json:"source"`
}

// MediaObjectLifecycleUpdateReport summarizes a metadata-only bulk lifecycle update.
type MediaObjectLifecycleUpdateReport struct {
	LifecycleState     string                             `json:"lifecycleState"`
	DryRun             bool                               `json:"dryRun"`
	MatchedObjects     int                                `json:"matchedObjects"`
	UpdatedObjects     int                                `json:"updatedObjects"`
	WouldUpdateObjects int                                `json:"wouldUpdateObjects"`
	FailedObjects      int                                `json:"failedObjects"`
	UpdatedAt          time.Time                          `json:"updatedAt"`
	Results            []MediaObjectLifecycleUpdateResult `json:"results"`
}

// MediaObjectLifecycleUpdateResult captures the update outcome for one media object.
type MediaObjectLifecycleUpdateResult struct {
	MediaObjectID          string       `json:"mediaObjectId"`
	PreviousLifecycleState string       `json:"previousLifecycleState"`
	LifecycleState         string       `json:"lifecycleState"`
	Status                 string       `json:"status"`
	Message                string       `json:"message,omitempty"`
	Object                 *MediaObject `json:"object,omitempty"`
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
	state, err := normalizeLifecycleState(state)
	if id == "" {
		return MediaObject{}, fmt.Errorf("%w: id is required", ErrInvalidMediaObject)
	}
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
	previousState := object.LifecycleState
	changedAt := service.now().UTC()
	object.LifecycleState = state
	object.UpdatedAt = changedAt
	object.LastLifecycleChange = &MediaObjectLifecycleChange{PreviousLifecycleState: previousState, LifecycleState: state, ChangedAt: changedAt, Source: "single"}
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

func (service *MediaObjectService) SetMediaObjectLifecycleStateByFilter(ctx context.Context, filter MediaObjectSelectionFilter, state string) (MediaObjectLifecycleUpdateReport, error) {
	return service.SetMediaObjectLifecycleStateByFilterWithOptions(ctx, filter, state, MediaObjectLifecycleUpdateOptions{})
}

func (service *MediaObjectService) SetMediaObjectLifecycleStateByFilterWithOptions(ctx context.Context, filter MediaObjectSelectionFilter, state string, options MediaObjectLifecycleUpdateOptions) (MediaObjectLifecycleUpdateReport, error) {
	normalizedFilter, err := normalizeMediaObjectSelectionFilter(filter)
	if err != nil {
		return MediaObjectLifecycleUpdateReport{}, err
	}
	normalizedState, err := normalizeLifecycleState(state)
	if err != nil {
		return MediaObjectLifecycleUpdateReport{}, err
	}
	objects, err := service.listMediaObjectsForSelection(ctx, normalizedFilter)
	if err != nil {
		return MediaObjectLifecycleUpdateReport{}, err
	}
	updatedAt := service.now().UTC()
	report := MediaObjectLifecycleUpdateReport{LifecycleState: normalizedState, DryRun: options.DryRun, MatchedObjects: len(objects), UpdatedAt: updatedAt, Results: make([]MediaObjectLifecycleUpdateResult, 0, len(objects))}
	for _, object := range objects {
		result := MediaObjectLifecycleUpdateResult{MediaObjectID: object.ID, PreviousLifecycleState: object.LifecycleState, LifecycleState: normalizedState}
		if LifecycleState(object.LifecycleState) == LifecycleStateDeleted && LifecycleState(normalizedState) != LifecycleStateDeleted {
			result.Status = "failed"
			result.Message = fmt.Sprintf("deleted media object %s cannot leave deleted lifecycle state", object.ID)
			report.FailedObjects++
			report.Results = append(report.Results, result)
			continue
		}
		object.LifecycleState = normalizedState
		object.UpdatedAt = updatedAt
		if options.DryRun {
			updatedObject := object
			result.Status = "would_update"
			result.Object = &updatedObject
			report.WouldUpdateObjects++
			report.Results = append(report.Results, result)
			continue
		}
		object.LastLifecycleChange = &MediaObjectLifecycleChange{PreviousLifecycleState: result.PreviousLifecycleState, LifecycleState: normalizedState, ChangedAt: updatedAt, Source: "bulk"}
		if err := service.mediaRepository.SaveMediaObject(ctx, object); err != nil {
			result.Status = "failed"
			result.Message = err.Error()
			report.FailedObjects++
			report.Results = append(report.Results, result)
			continue
		}
		updatedObject := object
		result.Status = "updated"
		result.Object = &updatedObject
		report.UpdatedObjects++
		report.Results = append(report.Results, result)
	}
	return report, nil
}

func (service *MediaObjectService) ListMediaObjects(ctx context.Context, filter MediaObjectListFilter) (MediaObjectListPage, error) {
	normalized, err := normalizeMediaObjectListFilter(filter)
	if err != nil {
		return MediaObjectListPage{}, err
	}

	objects, err := service.listMediaObjectsForSelection(ctx, MediaObjectSelectionFilter{
		BackendID:          normalized.BackendID,
		ContentHash:        normalized.ContentHash,
		VerificationStatus: normalized.VerificationStatus,
		LifecycleState:     normalized.LifecycleState,
		AssetKind:          normalized.AssetKind,
	})
	if err != nil {
		return MediaObjectListPage{}, err
	}
	sortMediaObjectsForList(objects, normalized.SortBy, normalized.SortOrder)
	return paginateMediaObjects(objects, normalized.Limit, normalized.Offset), nil
}

func (service *MediaObjectService) GetMediaObjectStats(ctx context.Context, backendID string) (MediaObjectStats, error) {
	backendID = strings.TrimSpace(backendID)
	objects, err := service.mediaObjectsForOptionalBackend(ctx, backendID)
	if err != nil {
		return MediaObjectStats{}, err
	}
	return summarizeMediaObjects(backendID, objects), nil
}

func (service *MediaObjectService) GetMediaObjectDuplicates(ctx context.Context, backendID string, minCopies int) (MediaObjectDuplicateReport, error) {
	backendID = strings.TrimSpace(backendID)
	if minCopies == 0 {
		minCopies = 2
	}
	if minCopies < 2 {
		return MediaObjectDuplicateReport{}, fmt.Errorf("%w: minCopies must be at least 2", ErrInvalidMediaObject)
	}
	objects, err := service.mediaObjectsForOptionalBackend(ctx, backendID)
	if err != nil {
		return MediaObjectDuplicateReport{}, err
	}
	return summarizeMediaObjectDuplicates(backendID, minCopies, objects), nil
}

func (service *MediaObjectService) mediaObjectsForOptionalBackend(ctx context.Context, backendID string) ([]MediaObject, error) {
	if backendID != "" {
		return service.mediaRepository.ListMediaObjectsByBackend(ctx, backendID)
	}
	return service.mediaRepository.ListAllMediaObjects(ctx)
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

func (service *MediaObjectService) listMediaObjectsForSelection(ctx context.Context, filter MediaObjectSelectionFilter) ([]MediaObject, error) {
	if filter.BackendID != "" {
		return service.mediaRepository.ListMediaObjectsByBackend(ctx, filter.BackendID)
	}
	if filter.ContentHash != "" {
		return service.mediaRepository.ListMediaObjectsByContentHash(ctx, filter.ContentHash)
	}
	if filter.VerificationStatus != "" {
		return service.mediaRepository.ListMediaObjectsByVerificationStatus(ctx, filter.VerificationStatus)
	}
	if filter.LifecycleState != "" {
		return service.mediaRepository.ListMediaObjectsByLifecycleState(ctx, filter.LifecycleState)
	}
	return service.mediaRepository.ListMediaObjectsByAssetKind(ctx, filter.AssetKind)
}

func normalizeMediaObjectSelectionFilter(filter MediaObjectSelectionFilter) (MediaObjectSelectionFilter, error) {
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
		return MediaObjectSelectionFilter{}, fmt.Errorf("%w: exactly one of backendId, contentHash, verificationStatus, lifecycleState, or assetKind is required", ErrInvalidMediaObject)
	}
	if filter.VerificationStatus != "" {
		normalized, err := normalizeMediaObjectVerificationStatus(filter.VerificationStatus)
		if err != nil {
			return MediaObjectSelectionFilter{}, err
		}
		filter.VerificationStatus = normalized
	}
	if filter.LifecycleState != "" {
		normalized, err := normalizeLifecycleState(filter.LifecycleState)
		if err != nil {
			return MediaObjectSelectionFilter{}, err
		}
		filter.LifecycleState = normalized
	}
	if filter.AssetKind != "" {
		normalized, err := normalizeAssetKind(filter.AssetKind)
		if err != nil {
			return MediaObjectSelectionFilter{}, err
		}
		filter.AssetKind = normalized
	}
	return filter, nil
}

func normalizeMediaObjectListFilter(filter MediaObjectListFilter) (MediaObjectListFilter, error) {
	normalizedSelection, err := normalizeMediaObjectSelectionFilter(MediaObjectSelectionFilter{
		BackendID:          filter.BackendID,
		ContentHash:        filter.ContentHash,
		VerificationStatus: filter.VerificationStatus,
		LifecycleState:     filter.LifecycleState,
		AssetKind:          filter.AssetKind,
	})
	if err != nil {
		return MediaObjectListFilter{}, err
	}
	filter.BackendID = normalizedSelection.BackendID
	filter.ContentHash = normalizedSelection.ContentHash
	filter.VerificationStatus = normalizedSelection.VerificationStatus
	filter.LifecycleState = normalizedSelection.LifecycleState
	filter.AssetKind = normalizedSelection.AssetKind
	filter.SortBy = strings.TrimSpace(filter.SortBy)
	filter.SortOrder = strings.TrimSpace(filter.SortOrder)
	if filter.SortBy == "" {
		filter.SortBy = DefaultMediaObjectSortBy
	} else {
		normalized, err := normalizeMediaObjectSortBy(filter.SortBy)
		if err != nil {
			return MediaObjectListFilter{}, err
		}
		filter.SortBy = normalized
	}
	if filter.SortOrder == "" {
		filter.SortOrder = DefaultMediaObjectSortOrder
	} else {
		normalized, err := normalizeMediaObjectSortOrder(filter.SortOrder)
		if err != nil {
			return MediaObjectListFilter{}, err
		}
		filter.SortOrder = normalized
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

func normalizeMediaObjectSortBy(sortBy string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(sortBy))
	switch normalized {
	case MediaObjectSortByBackendObjectKey, MediaObjectSortByCreatedAt, MediaObjectSortByUpdatedAt, MediaObjectSortBySizeBytes, MediaObjectSortByObjectKey, MediaObjectSortByID:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: sortBy must be one of backend_object_key, created_at, updated_at, size_bytes, object_key, or id", ErrInvalidMediaObject)
	}
}

func normalizeMediaObjectSortOrder(order string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(order))
	switch normalized {
	case MediaObjectSortOrderAscending, MediaObjectSortOrderDescending:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: sortOrder must be asc or desc", ErrInvalidMediaObject)
	}
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

func summarizeMediaObjectDuplicates(backendID string, minCopies int, objects []MediaObject) MediaObjectDuplicateReport {
	groupsByHash := make(map[string][]MediaObject)
	for _, object := range objects {
		groupsByHash[object.ContentHash] = append(groupsByHash[object.ContentHash], object)
	}
	report := MediaObjectDuplicateReport{BackendID: backendID, MinCopies: minCopies, Groups: make([]MediaObjectDuplicateGroup, 0)}
	for contentHash, groupObjects := range groupsByHash {
		if len(groupObjects) < minCopies {
			continue
		}
		sortMediaObjectsForList(groupObjects, DefaultMediaObjectSortBy, DefaultMediaObjectSortOrder)
		group := MediaObjectDuplicateGroup{ContentHash: contentHash, Count: len(groupObjects), Objects: groupObjects}
		for _, object := range groupObjects {
			group.TotalSizeBytes += object.SizeBytes
		}
		report.TotalGroups++
		report.TotalObjects += group.Count
		report.TotalSizeBytes += group.TotalSizeBytes
		report.Groups = append(report.Groups, group)
	}
	sort.Slice(report.Groups, func(i, j int) bool {
		if report.Groups[i].Count == report.Groups[j].Count {
			return report.Groups[i].ContentHash < report.Groups[j].ContentHash
		}
		return report.Groups[i].Count > report.Groups[j].Count
	})
	return report
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

func sortMediaObjectsForList(objects []MediaObject, sortBy string, sortOrder string) {
	descending := sortOrder == MediaObjectSortOrderDescending
	sort.SliceStable(objects, func(i, j int) bool {
		if descending {
			return mediaObjectLess(objects[j], objects[i], sortBy)
		}
		return mediaObjectLess(objects[i], objects[j], sortBy)
	})
}

func mediaObjectLess(left MediaObject, right MediaObject, sortBy string) bool {
	switch sortBy {
	case MediaObjectSortByCreatedAt:
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.Before(right.CreatedAt)
		}
	case MediaObjectSortByUpdatedAt:
		if !left.UpdatedAt.Equal(right.UpdatedAt) {
			return left.UpdatedAt.Before(right.UpdatedAt)
		}
	case MediaObjectSortBySizeBytes:
		if left.SizeBytes != right.SizeBytes {
			return left.SizeBytes < right.SizeBytes
		}
	case MediaObjectSortByObjectKey:
		if left.ObjectKey != right.ObjectKey {
			return left.ObjectKey < right.ObjectKey
		}
	case MediaObjectSortByID:
		if left.ID != right.ID {
			return left.ID < right.ID
		}
	default:
		if left.BackendID != right.BackendID {
			return left.BackendID < right.BackendID
		}
		if left.ObjectKey != right.ObjectKey {
			return left.ObjectKey < right.ObjectKey
		}
	}
	return left.ID < right.ID
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

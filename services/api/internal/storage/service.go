package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Service coordinates storage backend administration rules.
type Service struct {
	repository       Repository
	prober           Prober
	capacityProvider CapacityProvider
	now              func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, prober: NewCompositeProber(NewFilesystemProber(), NewS3Prober()), capacityProvider: NewFilesystemCapacityProvider(), now: time.Now}
}

// ValidateBackend checks a backend candidate without persisting it or probing external systems.
func (service *Service) ValidateBackend(backend StorageBackend) (StorageBackend, error) {
	if err := ValidateBackend(&backend); err != nil {
		return StorageBackend{}, err
	}
	return backend, nil
}

func (service *Service) RegisterBackend(ctx context.Context, backend StorageBackend) (StorageBackend, error) {
	if err := ValidateBackend(&backend); err != nil {
		return StorageBackend{}, err
	}
	if _, err := service.repository.Get(ctx, backend.ID); err == nil {
		return StorageBackend{}, fmt.Errorf("%w: backend %s already exists", ErrConflict, backend.ID)
	} else if !errors.Is(err, ErrNotFound) {
		return StorageBackend{}, err
	}

	now := service.now().UTC()
	backend.HealthStatus = HealthStatusUnknown
	backend.LastHealthCheckAt = nil
	backend.CreatedAt = now
	backend.UpdatedAt = now

	if backend.IsDefault {
		if err := ensureDefaultCandidate(backend); err != nil {
			return StorageBackend{}, err
		}
		if err := service.repository.ClearDefault(ctx); err != nil {
			return StorageBackend{}, err
		}
	}

	if err := service.repository.Save(ctx, backend); err != nil {
		return StorageBackend{}, err
	}
	return backend, nil
}

func (service *Service) ListBackends(ctx context.Context) ([]StorageBackend, error) {
	return service.repository.List(ctx)
}

func (service *Service) GetBackendCapacity(ctx context.Context, id string) (CapacityReport, error) {
	backend, err := service.repository.Get(ctx, id)
	if err != nil {
		return CapacityReport{}, err
	}
	if !backend.Enabled {
		return CapacityReport{}, fmt.Errorf("%w: %s", ErrBackendDisabled, id)
	}
	report, err := service.capacityProvider.Capacity(ctx, backend)
	if err != nil {
		return CapacityReport{}, err
	}
	report.BackendID = backend.ID
	report.CheckedAt = service.now().UTC()
	backend.LastCapacity = &report
	backend.UpdatedAt = report.CheckedAt
	if err := service.repository.Save(ctx, backend); err != nil {
		return CapacityReport{}, err
	}
	return report, nil
}

func (service *Service) GetBackendHealth(ctx context.Context, id string) (ProbeResult, error) {
	backend, err := service.repository.Get(ctx, id)
	if err != nil {
		return ProbeResult{}, err
	}
	result := ProbeResult{BackendID: backend.ID, Status: backend.HealthStatus}
	if backend.LastHealthCheckAt != nil {
		result.CheckedAt = *backend.LastHealthCheckAt
	}
	return result, nil
}

func (service *Service) ProbeBackend(ctx context.Context, id string) (ProbeResult, error) {
	backend, err := service.repository.Get(ctx, id)
	if err != nil {
		return ProbeResult{}, err
	}
	if !backend.Enabled {
		return ProbeResult{BackendID: backend.ID, Status: HealthStatusDisabled}, fmt.Errorf("%w: %s", ErrBackendDisabled, id)
	}

	checkedAt := service.now().UTC()
	probeErr := service.prober.Probe(ctx, backend)
	if errors.Is(probeErr, ErrProbeUnsupported) {
		return ProbeResult{BackendID: backend.ID, Status: backend.HealthStatus, Message: probeErr.Error()}, probeErr
	}
	backend.LastHealthCheckAt = &checkedAt
	backend.UpdatedAt = checkedAt
	backend.HealthStatus = HealthStatusHealthy
	result := ProbeResult{BackendID: backend.ID, Status: HealthStatusHealthy, CheckedAt: checkedAt}
	if probeErr != nil {
		backend.HealthStatus = HealthStatusUnhealthy
		result.Status = HealthStatusUnhealthy
		result.Message = probeErr.Error()
	}
	if err := service.repository.Save(ctx, backend); err != nil {
		return ProbeResult{}, err
	}
	if probeErr != nil {
		return result, probeErr
	}
	return result, nil
}

func (service *Service) SetDefaultBackend(ctx context.Context, id string) (StorageBackend, error) {
	backend, err := service.repository.Get(ctx, id)
	if err != nil {
		return StorageBackend{}, err
	}
	if err := ensureDefaultCandidate(backend); err != nil {
		return StorageBackend{}, err
	}
	backend.IsDefault = true
	backend.UpdatedAt = service.now().UTC()

	if err := service.repository.ClearDefault(ctx); err != nil {
		return StorageBackend{}, err
	}
	if err := service.repository.Save(ctx, backend); err != nil {
		return StorageBackend{}, err
	}
	return backend, nil
}

func (service *Service) DisableBackend(ctx context.Context, id string) (StorageBackend, error) {
	backend, err := service.repository.Get(ctx, id)
	if err != nil {
		return StorageBackend{}, err
	}
	if backend.IsDefault {
		return StorageBackend{}, fmt.Errorf("%w: cannot disable default backend %s", ErrConflict, id)
	}
	backend.Enabled = false
	backend.HealthStatus = HealthStatusDisabled
	backend.UpdatedAt = service.now().UTC()
	if err := service.repository.Save(ctx, backend); err != nil {
		return StorageBackend{}, err
	}
	return backend, nil
}

func (service *Service) EnableBackend(ctx context.Context, id string) (StorageBackend, error) {
	backend, err := service.repository.Get(ctx, id)
	if err != nil {
		return StorageBackend{}, err
	}
	if backend.Enabled {
		return backend, nil
	}
	backend.Enabled = true
	backend.HealthStatus = HealthStatusUnknown
	backend.UpdatedAt = service.now().UTC()
	if err := service.repository.Save(ctx, backend); err != nil {
		return StorageBackend{}, err
	}
	return backend, nil
}

func (service *Service) DeleteBackend(ctx context.Context, id string) error {
	backend, err := service.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if backend.IsDefault {
		return fmt.Errorf("%w: cannot delete default backend %s", ErrBackendIsDefault, id)
	}
	return service.repository.Delete(ctx, id)
}

// UpdateBackendRequest carries the fields that may be changed via a PATCH request.
// Nil pointer fields are left unchanged.
type UpdateBackendRequest struct {
	DisplayName *string
	Priority    *int
}

func (service *Service) UpdateBackend(ctx context.Context, id string, req UpdateBackendRequest) (StorageBackend, error) {
	backend, err := service.repository.Get(ctx, id)
	if err != nil {
		return StorageBackend{}, err
	}
	if req.DisplayName != nil {
		name := strings.TrimSpace(*req.DisplayName)
		if name == "" {
			return StorageBackend{}, fmt.Errorf("%w: display name must not be empty", ErrInvalidBackend)
		}
		backend.DisplayName = name
	}
	if req.Priority != nil {
		backend.Priority = *req.Priority
	}
	backend.UpdatedAt = service.now().UTC()
	if err := service.repository.Save(ctx, backend); err != nil {
		return StorageBackend{}, err
	}
	return backend, nil
}

func ensureDefaultCandidate(backend StorageBackend) error {
	if !backend.Enabled {
		return fmt.Errorf("%w: default backend must be enabled", ErrInvalidBackend)
	}
	if backend.HealthStatus == HealthStatusUnhealthy || backend.HealthStatus == HealthStatusDisabled {
		return fmt.Errorf("%w: default backend must not be %s", ErrInvalidBackend, backend.HealthStatus)
	}
	return nil
}

// GetBackend returns a single backend by ID.
func (service *Service) GetBackend(ctx context.Context, id string) (StorageBackend, error) {
	return service.repository.Get(ctx, id)
}

// DefaultPresignedURLTTL is the time-to-live used when generating presigned URLs.
const DefaultPresignedURLTTL = 15 * time.Minute

// GeneratePresignedURL generates an AWS Signature Version 4 presigned GET URL
// for an object on a backend that advertises PresignedURLs capability.
// Returns ErrProbeUnsupported when the backend does not support presigned URLs.
// Returns ErrProbeFailed when credentials cannot be resolved or the config is invalid.
func (service *Service) GeneratePresignedURL(ctx context.Context, backendID string, objectKey string, ttl time.Duration) (string, error) {
	backend, err := service.repository.Get(ctx, backendID)
	if err != nil {
		return "", err
	}
	if !backend.Capabilities.PresignedURLs {
		return "", fmt.Errorf("%w: backend %s does not support presigned URLs", ErrProbeUnsupported, backendID)
	}
	config, ok := s3ProbeConfig(backend)
	if !ok {
		return "", fmt.Errorf("%w: backend %s does not have an S3-compatible configuration", ErrProbeUnsupported, backendID)
	}
	accessKey, secretKey, err := resolveS3ProbeCredentials(config)
	if err != nil {
		return "", err
	}
	return presignS3URL(config, objectKey, accessKey, secretKey, ttl, service.now())
}

// localRootPath returns the local filesystem root for a backend (local/NFS/SMB).
// Returns "", false for backends that don't map to a local path (e.g. S3).
func localRootPath(backend StorageBackend) (string, bool) {
	switch backend.Type {
	case BackendTypeLocal:
		if backend.Config.Local != nil {
			return backend.Config.Local.RootPath, true
		}
	case BackendTypeNFS:
		if backend.Config.NFS != nil {
			return backend.Config.NFS.MountPath, true
		}
	case BackendTypeSMB:
		if backend.Config.SMB != nil {
			return backend.Config.SMB.MountPath, true
		}
	}
	return "", false
}

// PutObject writes r to the object identified by (backendID, objectKey) on a
// filesystem-backed storage backend (local / NFS / SMB). Returns an error for
// backends that do not expose a local path (e.g. S3).
func (service *Service) PutObject(ctx context.Context, backendID, objectKey string, r io.Reader, size int64) error {
	backend, err := service.repository.Get(ctx, backendID)
	if err != nil {
		return err
	}
	root, ok := localRootPath(backend)
	if !ok {
		return fmt.Errorf("%w: backend %s does not support direct object writes", ErrProbeUnsupported, backendID)
	}
	fullPath, err := SafeObjectPath(root, objectKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		return fmt.Errorf("create object directory: %w", err)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create object file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("write object: %w", err)
	}
	return nil
}

// GetObject returns a ReadCloser for the object identified by (backendID, objectKey)
// on a filesystem-backed storage backend (local / NFS / SMB).
func (service *Service) GetObject(ctx context.Context, backendID, objectKey string) (io.ReadCloser, error) {
	backend, err := service.repository.Get(ctx, backendID)
	if err != nil {
		return nil, err
	}
	root, ok := localRootPath(backend)
	if !ok {
		return nil, fmt.Errorf("%w: backend %s does not support direct object reads", ErrProbeUnsupported, backendID)
	}
	fullPath, err := SafeObjectPath(root, objectKey)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: object %s not found", ErrNotFound, objectKey)
		}
		return nil, fmt.Errorf("open object: %w", err)
	}
	return f, nil
}

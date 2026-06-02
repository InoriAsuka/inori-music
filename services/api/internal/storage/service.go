package storage

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Service coordinates storage backend administration rules.
type Service struct {
	repository Repository
	prober     Prober
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, prober: NewCompositeProber(NewFilesystemProber(), NewS3Prober()), now: time.Now}
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

func ensureDefaultCandidate(backend StorageBackend) error {
	if !backend.Enabled {
		return fmt.Errorf("%w: default backend must be enabled", ErrInvalidBackend)
	}
	if backend.HealthStatus == HealthStatusUnhealthy || backend.HealthStatus == HealthStatusDisabled {
		return fmt.Errorf("%w: default backend must not be %s", ErrInvalidBackend, backend.HealthStatus)
	}
	return nil
}

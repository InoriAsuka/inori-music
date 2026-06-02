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
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
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
	if backend.CreatedAt.IsZero() {
		backend.CreatedAt = now
	}
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

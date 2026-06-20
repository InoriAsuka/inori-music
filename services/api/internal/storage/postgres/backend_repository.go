package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/storage"
)

// BackendRepository implements storage.Repository using a PostgreSQL connection pool.
type BackendRepository struct {
	pool *pgxpool.Pool
}

// NewBackendRepository returns a BackendRepository backed by the given pool.
func NewBackendRepository(pool *pgxpool.Pool) *BackendRepository {
	return &BackendRepository{pool: pool}
}

func (r *BackendRepository) Save(ctx context.Context, backend storage.StorageBackend) error {
	capJSON, err := json.Marshal(backend.Capabilities)
	if err != nil {
		return fmt.Errorf("marshal capabilities: %w", err)
	}
	cfgJSON, err := json.Marshal(backend.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	var capRaw, lastCapRaw []byte
	capRaw = capJSON
	if backend.LastCapacity != nil {
		lastCapRaw, err = json.Marshal(backend.LastCapacity)
		if err != nil {
			return fmt.Errorf("marshal last_capacity: %w", err)
		}
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO storage_backends
		    (id, type, display_name, enabled, is_default, priority,
		     health_status, last_health_check_at, last_capacity, capabilities, config,
		     created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (id) DO UPDATE SET
		    type                 = EXCLUDED.type,
		    display_name         = EXCLUDED.display_name,
		    enabled              = EXCLUDED.enabled,
		    is_default           = EXCLUDED.is_default,
		    priority             = EXCLUDED.priority,
		    health_status        = EXCLUDED.health_status,
		    last_health_check_at = EXCLUDED.last_health_check_at,
		    last_capacity        = EXCLUDED.last_capacity,
		    capabilities         = EXCLUDED.capabilities,
		    config               = EXCLUDED.config,
		    updated_at           = EXCLUDED.updated_at`,
		backend.ID, string(backend.Type), backend.DisplayName,
		backend.Enabled, backend.IsDefault, backend.Priority,
		string(backend.HealthStatus), backend.LastHealthCheckAt, lastCapRaw,
		capRaw, cfgJSON,
		backend.CreatedAt.UTC(), backend.UpdatedAt.UTC(),
	)
	return err
}

func (r *BackendRepository) Get(ctx context.Context, id string) (storage.StorageBackend, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, type, display_name, enabled, is_default, priority,
		       health_status, last_health_check_at, last_capacity, capabilities, config,
		       created_at, updated_at
		FROM storage_backends WHERE id = $1`, id)
	backend, err := scanBackend(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return storage.StorageBackend{}, fmt.Errorf("%w: %s", storage.ErrNotFound, id)
		}
		return storage.StorageBackend{}, err
	}
	return backend, nil
}

func (r *BackendRepository) List(ctx context.Context) ([]storage.StorageBackend, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, type, display_name, enabled, is_default, priority,
		       health_status, last_health_check_at, last_capacity, capabilities, config,
		       created_at, updated_at
		FROM storage_backends ORDER BY priority, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var backends []storage.StorageBackend
	for rows.Next() {
		backend, err := scanBackend(rows)
		if err != nil {
			return nil, err
		}
		backends = append(backends, backend)
	}
	return backends, rows.Err()
}

func (r *BackendRepository) ClearDefault(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `UPDATE storage_backends SET is_default = FALSE WHERE is_default = TRUE`)
	return err
}

func (r *BackendRepository) Delete(ctx context.Context, id string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM storage_backends WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", storage.ErrNotFound, id)
	}
	return nil
}

// scanner is satisfied by both pgx.Row and pgx.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanBackend(s scanner) (storage.StorageBackend, error) {
	var (
		b               storage.StorageBackend
		backendType     string
		healthStatus    string
		capJSON         []byte
		cfgJSON         []byte
		lastCapJSON     []byte
		lastHealthCheck *time.Time
	)
	err := s.Scan(
		&b.ID, &backendType, &b.DisplayName,
		&b.Enabled, &b.IsDefault, &b.Priority,
		&healthStatus, &lastHealthCheck, &lastCapJSON,
		&capJSON, &cfgJSON,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return storage.StorageBackend{}, err
	}
	b.Type = storage.BackendType(backendType)
	b.HealthStatus = storage.HealthStatus(healthStatus)
	b.LastHealthCheckAt = lastHealthCheck
	if err := json.Unmarshal(capJSON, &b.Capabilities); err != nil {
		return storage.StorageBackend{}, fmt.Errorf("unmarshal capabilities: %w", err)
	}
	if err := json.Unmarshal(cfgJSON, &b.Config); err != nil {
		return storage.StorageBackend{}, fmt.Errorf("unmarshal config: %w", err)
	}
	if lastCapJSON != nil {
		var cap storage.CapacityReport
		if err := json.Unmarshal(lastCapJSON, &cap); err != nil {
			return storage.StorageBackend{}, fmt.Errorf("unmarshal last_capacity: %w", err)
		}
		b.LastCapacity = &cap
	}
	return b, nil
}

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/storage"
)

// MediaObjectRepository implements storage.MediaObjectRepository using a PostgreSQL connection pool.
type MediaObjectRepository struct {
	pool *pgxpool.Pool
}

// NewMediaObjectRepository returns a MediaObjectRepository backed by the given pool.
func NewMediaObjectRepository(pool *pgxpool.Pool) *MediaObjectRepository {
	return &MediaObjectRepository{pool: pool}
}

func (r *MediaObjectRepository) SaveMediaObject(ctx context.Context, object storage.MediaObject) error {
	var verJSON []byte
	if object.LastVerification != nil {
		var err error
		verJSON, err = json.Marshal(object.LastVerification)
		if err != nil {
			return fmt.Errorf("marshal last_verification: %w", err)
		}
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO media_objects
		    (id, backend_id, object_key, content_hash, size_bytes,
		     mime_type, asset_kind, lifecycle_state, last_verification,
		     created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
		    backend_id        = EXCLUDED.backend_id,
		    object_key        = EXCLUDED.object_key,
		    content_hash      = EXCLUDED.content_hash,
		    size_bytes        = EXCLUDED.size_bytes,
		    mime_type         = EXCLUDED.mime_type,
		    asset_kind        = EXCLUDED.asset_kind,
		    lifecycle_state   = EXCLUDED.lifecycle_state,
		    last_verification = EXCLUDED.last_verification,
		    updated_at        = EXCLUDED.updated_at`,
		object.ID, object.BackendID, object.ObjectKey, object.ContentHash, object.SizeBytes,
		object.MIMEType, object.AssetKind, object.LifecycleState, verJSON,
		object.CreatedAt.UTC(), object.UpdatedAt.UTC(),
	)
	return err
}

func (r *MediaObjectRepository) GetMediaObject(ctx context.Context, id string) (storage.MediaObject, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, backend_id, object_key, content_hash, size_bytes,
		       mime_type, asset_kind, lifecycle_state, last_verification,
		       created_at, updated_at
		FROM media_objects WHERE id = $1`, id)
	obj, err := scanMediaObject(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return storage.MediaObject{}, fmt.Errorf("%w: media object %s", storage.ErrNotFound, id)
		}
		return storage.MediaObject{}, err
	}
	return obj, nil
}

func (r *MediaObjectRepository) ListAllMediaObjects(ctx context.Context) ([]storage.MediaObject, error) {
	return r.queryMediaObjects(ctx, `
		SELECT id, backend_id, object_key, content_hash, size_bytes,
		       mime_type, asset_kind, lifecycle_state, last_verification,
		       created_at, updated_at
		FROM media_objects ORDER BY backend_id, object_key, id`)
}

func (r *MediaObjectRepository) ListMediaObjectsByBackend(ctx context.Context, backendID string) ([]storage.MediaObject, error) {
	return r.queryMediaObjects(ctx, `
		SELECT id, backend_id, object_key, content_hash, size_bytes,
		       mime_type, asset_kind, lifecycle_state, last_verification,
		       created_at, updated_at
		FROM media_objects WHERE backend_id = $1 ORDER BY object_key, id`, backendID)
}

func (r *MediaObjectRepository) ListMediaObjectsByContentHash(ctx context.Context, contentHash string) ([]storage.MediaObject, error) {
	return r.queryMediaObjects(ctx, `
		SELECT id, backend_id, object_key, content_hash, size_bytes,
		       mime_type, asset_kind, lifecycle_state, last_verification,
		       created_at, updated_at
		FROM media_objects WHERE content_hash = $1 ORDER BY backend_id, object_key, id`, contentHash)
}

func (r *MediaObjectRepository) ListMediaObjectsByVerificationStatus(ctx context.Context, status string) ([]storage.MediaObject, error) {
	// last_verification->>'status' is null when no verification has run yet (maps to "unknown").
	if status == "unknown" {
		return r.queryMediaObjects(ctx, `
			SELECT id, backend_id, object_key, content_hash, size_bytes,
			       mime_type, asset_kind, lifecycle_state, last_verification,
			       created_at, updated_at
			FROM media_objects
			WHERE last_verification IS NULL
			   OR last_verification->>'status' IS NULL
			   OR lower(last_verification->>'status') = 'unknown'
			ORDER BY backend_id, object_key, id`)
	}
	return r.queryMediaObjects(ctx, `
		SELECT id, backend_id, object_key, content_hash, size_bytes,
		       mime_type, asset_kind, lifecycle_state, last_verification,
		       created_at, updated_at
		FROM media_objects
		WHERE lower(last_verification->>'status') = lower($1)
		ORDER BY backend_id, object_key, id`, status)
}

func (r *MediaObjectRepository) ListMediaObjectsByLifecycleState(ctx context.Context, state string) ([]storage.MediaObject, error) {
	return r.queryMediaObjects(ctx, `
		SELECT id, backend_id, object_key, content_hash, size_bytes,
		       mime_type, asset_kind, lifecycle_state, last_verification,
		       created_at, updated_at
		FROM media_objects WHERE lifecycle_state = $1 ORDER BY backend_id, object_key, id`, state)
}

func (r *MediaObjectRepository) ListMediaObjectsByAssetKind(ctx context.Context, kind string) ([]storage.MediaObject, error) {
	return r.queryMediaObjects(ctx, `
		SELECT id, backend_id, object_key, content_hash, size_bytes,
		       mime_type, asset_kind, lifecycle_state, last_verification,
		       created_at, updated_at
		FROM media_objects WHERE asset_kind = $1 ORDER BY backend_id, object_key, id`, kind)
}

func (r *MediaObjectRepository) queryMediaObjects(ctx context.Context, sql string, args ...any) ([]storage.MediaObject, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var objects []storage.MediaObject
	for rows.Next() {
		obj, err := scanMediaObject(rows)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}
	return objects, rows.Err()
}

func scanMediaObject(s scanner) (storage.MediaObject, error) {
	var (
		obj     storage.MediaObject
		verJSON []byte
	)
	err := s.Scan(
		&obj.ID, &obj.BackendID, &obj.ObjectKey, &obj.ContentHash, &obj.SizeBytes,
		&obj.MIMEType, &obj.AssetKind, &obj.LifecycleState, &verJSON,
		&obj.CreatedAt, &obj.UpdatedAt,
	)
	if err != nil {
		return storage.MediaObject{}, err
	}
	if verJSON != nil {
		var ver storage.MediaObjectVerificationResult
		if err := json.Unmarshal(verJSON, &ver); err != nil {
			return storage.MediaObject{}, fmt.Errorf("unmarshal last_verification: %w", err)
		}
		obj.LastVerification = &ver
	}
	return obj, nil
}

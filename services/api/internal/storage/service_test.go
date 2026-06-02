package storage

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestServiceRegistersAndSelectsDefaultBackend(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryRepository())

	local, err := service.RegisterBackend(ctx, StorageBackend{
		ID:          "local-main",
		Type:        BackendTypeLocal,
		DisplayName: "Local",
		Enabled:     true,
		IsDefault:   true,
		Config:      BackendConfig{Local: &LocalConfig{RootPath: "/srv/inori/media"}},
	})
	if err != nil {
		t.Fatalf("RegisterBackend(local) error = %v", err)
	}
	if !local.IsDefault {
		t.Fatal("local backend should be default")
	}

	s3, err := service.RegisterBackend(ctx, StorageBackend{
		ID:          "s3-prod",
		Type:        BackendTypeS3,
		DisplayName: "S3",
		Enabled:     true,
		Config:      BackendConfig{S3: &S3Config{Endpoint: "https://s3.example.com", Bucket: "inori", AccessKeySecretRef: "S3_ACCESS", SecretKeySecretRef: "S3_SECRET"}},
	})
	if err != nil {
		t.Fatalf("RegisterBackend(s3) error = %v", err)
	}
	if s3.IsDefault {
		t.Fatal("s3 backend should not be default before selection")
	}

	selected, err := service.SetDefaultBackend(ctx, "s3-prod")
	if err != nil {
		t.Fatalf("SetDefaultBackend() error = %v", err)
	}
	if !selected.IsDefault {
		t.Fatal("selected backend should be default")
	}

	backends, err := service.ListBackends(ctx)
	if err != nil {
		t.Fatalf("ListBackends() error = %v", err)
	}
	defaults := 0
	for _, backend := range backends {
		if backend.IsDefault {
			defaults++
		}
	}
	if defaults != 1 {
		t.Fatalf("default backend count = %d, want 1", defaults)
	}
}

func TestServiceRejectsDuplicateBackendRegistration(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryRepository())
	backend := StorageBackend{
		ID:          "local-main",
		Type:        BackendTypeLocal,
		DisplayName: "Local",
		Enabled:     true,
		Config:      BackendConfig{Local: &LocalConfig{RootPath: "/srv/inori/media"}},
	}

	if _, err := service.RegisterBackend(ctx, backend); err != nil {
		t.Fatalf("RegisterBackend(first) error = %v", err)
	}
	_, err := service.RegisterBackend(ctx, backend)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("RegisterBackend(duplicate) error = %v, want ErrConflict", err)
	}
}

func TestServiceRejectsDisablingDefaultBackend(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryRepository())

	_, err := service.RegisterBackend(ctx, StorageBackend{
		ID:          "local-main",
		Type:        BackendTypeLocal,
		DisplayName: "Local",
		Enabled:     true,
		IsDefault:   true,
		Config:      BackendConfig{Local: &LocalConfig{RootPath: "/srv/inori/media"}},
	})
	if err != nil {
		t.Fatalf("RegisterBackend() error = %v", err)
	}

	_, err = service.DisableBackend(ctx, "local-main")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("DisableBackend() error = %v, want ErrConflict", err)
	}
}

func TestServiceCanDisableNonDefaultBackend(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryRepository())

	_, err := service.RegisterBackend(ctx, StorageBackend{
		ID:          "smb-nas",
		Type:        BackendTypeSMB,
		DisplayName: "SMB NAS",
		Enabled:     true,
		Config:      BackendConfig{SMB: &SMBConfig{MountPath: "/mnt/smb/inori", ExpectedShare: "//nas/music"}},
	})
	if err != nil {
		t.Fatalf("RegisterBackend() error = %v", err)
	}

	disabled, err := service.DisableBackend(ctx, "smb-nas")
	if err != nil {
		t.Fatalf("DisableBackend() error = %v", err)
	}
	if disabled.Enabled || disabled.HealthStatus != HealthStatusDisabled {
		t.Fatalf("disabled backend = %+v, want disabled health state", disabled)
	}
}

func TestServiceRegistrationResetsServerOwnedState(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryRepository())
	checkTime := time.Date(2020, time.January, 2, 3, 4, 5, 0, time.UTC)

	registered, err := service.RegisterBackend(ctx, StorageBackend{
		ID:                "local-main",
		Type:              BackendTypeLocal,
		DisplayName:       "Local",
		Enabled:           true,
		HealthStatus:      HealthStatusHealthy,
		LastHealthCheckAt: &checkTime,
		CreatedAt:         checkTime,
		Config:            BackendConfig{Local: &LocalConfig{RootPath: "/srv/inori/media"}},
	})
	if err != nil {
		t.Fatalf("RegisterBackend() error = %v", err)
	}
	if registered.HealthStatus != HealthStatusUnknown || registered.LastHealthCheckAt != nil {
		t.Fatalf("registered health state = %+v, want reset server-owned health", registered)
	}
	if !registered.CreatedAt.After(checkTime) {
		t.Fatalf("registered createdAt = %v, want server time after %v", registered.CreatedAt, checkTime)
	}
}

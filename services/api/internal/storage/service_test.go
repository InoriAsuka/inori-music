package storage

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
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

func TestServiceProbeBackendUpdatesHealth(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryRepository())
	root := t.TempDir()
	_, err := service.RegisterBackend(ctx, StorageBackend{
		ID:          "local-main",
		Type:        BackendTypeLocal,
		DisplayName: "Local",
		Enabled:     true,
		Config:      BackendConfig{Local: &LocalConfig{RootPath: root}},
	})
	if err != nil {
		t.Fatalf("RegisterBackend() error = %v", err)
	}

	result, err := service.ProbeBackend(ctx, "local-main")
	if err != nil {
		t.Fatalf("ProbeBackend() error = %v", err)
	}
	if result.Status != HealthStatusHealthy || result.CheckedAt.IsZero() {
		t.Fatalf("ProbeBackend() result = %+v, want healthy checked result", result)
	}
	health, err := service.GetBackendHealth(ctx, "local-main")
	if err != nil {
		t.Fatalf("GetBackendHealth() error = %v", err)
	}
	if health.Status != HealthStatusHealthy || !health.CheckedAt.Equal(result.CheckedAt) {
		t.Fatalf("GetBackendHealth() = %+v, want latest probe state %+v", health, result)
	}
}

func TestServiceProbeBackendRecordsMissingS3CredentialFailure(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryRepository())
	_, err := service.RegisterBackend(ctx, StorageBackend{
		ID:          "s3-main",
		Type:        BackendTypeS3,
		DisplayName: "S3",
		Enabled:     true,
		Config:      BackendConfig{S3: &S3Config{Endpoint: "https://s3.example.com", Bucket: "inori", AccessKeySecretRef: "A", SecretKeySecretRef: "S"}},
	})
	if err != nil {
		t.Fatalf("RegisterBackend() error = %v", err)
	}

	result, err := service.ProbeBackend(ctx, "s3-main")
	if !errors.Is(err, ErrProbeFailed) {
		t.Fatalf("ProbeBackend() error = %v, want ErrProbeFailed", err)
	}
	if result.Status != HealthStatusUnhealthy || result.CheckedAt.IsZero() {
		t.Fatalf("ProbeBackend() result = %+v, want unhealthy checked state for failed s3 probe", result)
	}
	health, err := service.GetBackendHealth(ctx, "s3-main")
	if err != nil {
		t.Fatalf("GetBackendHealth() error = %v", err)
	}
	if health.Status != HealthStatusUnhealthy || health.CheckedAt.IsZero() {
		t.Fatalf("GetBackendHealth() = %+v, want unhealthy checked state", health)
	}
}

func TestServiceProbeBackendRejectsDisabledBackend(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryRepository())
	_, err := service.RegisterBackend(ctx, StorageBackend{
		ID:          "local-disabled",
		Type:        BackendTypeLocal,
		DisplayName: "Disabled",
		Enabled:     false,
		Config:      BackendConfig{Local: &LocalConfig{RootPath: t.TempDir()}},
	})
	if err != nil {
		t.Fatalf("RegisterBackend() error = %v", err)
	}

	result, err := service.ProbeBackend(ctx, "local-disabled")
	if !errors.Is(err, ErrBackendDisabled) {
		t.Fatalf("ProbeBackend() error = %v, want ErrBackendDisabled", err)
	}
	if result.Status != HealthStatusDisabled {
		t.Fatalf("ProbeBackend() result = %+v, want disabled", result)
	}
}

func TestServiceProbeBackendRecordsFilesystemFailure(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryRepository())
	missingRoot := filepath.Join(t.TempDir(), "missing")
	_, err := service.RegisterBackend(ctx, StorageBackend{
		ID:          "local-missing",
		Type:        BackendTypeLocal,
		DisplayName: "Missing",
		Enabled:     true,
		Config:      BackendConfig{Local: &LocalConfig{RootPath: missingRoot}},
	})
	if err != nil {
		t.Fatalf("RegisterBackend() error = %v", err)
	}

	result, err := service.ProbeBackend(ctx, "local-missing")
	if !errors.Is(err, ErrProbeFailed) {
		t.Fatalf("ProbeBackend() error = %v, want ErrProbeFailed", err)
	}
	if result.Status != HealthStatusUnhealthy || result.CheckedAt.IsZero() {
		t.Fatalf("ProbeBackend() result = %+v, want unhealthy checked result", result)
	}
	health, err := service.GetBackendHealth(ctx, "local-missing")
	if err != nil {
		t.Fatalf("GetBackendHealth() error = %v", err)
	}
	if health.Status != HealthStatusUnhealthy || health.CheckedAt.IsZero() {
		t.Fatalf("GetBackendHealth() = %+v, want unhealthy checked state", health)
	}
}

func TestGetBackend(t *testing.T) {
	ctx := context.Background()
	svc := NewService(NewMemoryRepository())
	_, err := svc.GetBackend(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetBackend(unknown) = %v, want ErrNotFound", err)
	}

	b := StorageBackend{
		ID:          "local-get",
		Type:        BackendTypeLocal,
		DisplayName: "Local Get",
		Enabled:     true,
		Config:      BackendConfig{Local: &LocalConfig{RootPath: "/srv"}},
	}
	if _, err := svc.RegisterBackend(ctx, b); err != nil {
		t.Fatalf("RegisterBackend: %v", err)
	}
	got, err := svc.GetBackend(ctx, "local-get")
	if err != nil {
		t.Fatalf("GetBackend: %v", err)
	}
	if got.ID != "local-get" {
		t.Errorf("GetBackend().ID = %q, want local-get", got.ID)
	}
}

func TestGeneratePresignedURLUnsupportedBackend(t *testing.T) {
	ctx := context.Background()
	svc := NewService(NewMemoryRepository())
	b := StorageBackend{
		ID:          "local-nopresign",
		Type:        BackendTypeLocal,
		DisplayName: "Local",
		Enabled:     true,
		Config:      BackendConfig{Local: &LocalConfig{RootPath: "/srv"}},
		// Capabilities.PresignedURLs is false by default
	}
	if _, err := svc.RegisterBackend(ctx, b); err != nil {
		t.Fatalf("RegisterBackend: %v", err)
	}
	_, err := svc.GeneratePresignedURL(ctx, "local-nopresign", "key", DefaultPresignedURLTTL)
	if !errors.Is(err, ErrProbeUnsupported) {
		t.Fatalf("GeneratePresignedURL(local) = %v, want ErrProbeUnsupported", err)
	}
}

func TestGeneratePresignedURLMissingCredentials(t *testing.T) {
	ctx := context.Background()
	svc := NewService(NewMemoryRepository())
	b := StorageBackend{
		ID:          "s3-nocreds",
		Type:        BackendTypeS3,
		DisplayName: "S3 No Creds",
		Enabled:     true,
		Capabilities: CapabilitySet{PresignedURLs: true},
		Config: BackendConfig{S3: &S3Config{
			Endpoint:           "https://s3.example.com",
			Region:             "us-east-1",
			Bucket:             "bucket",
			PathStyle:          true,
			AccessKeySecretRef: "MISSING_PS_ACCESS",
			SecretKeySecretRef: "MISSING_PS_SECRET",
		}},
	}
	if _, err := svc.RegisterBackend(ctx, b); err != nil {
		t.Fatalf("RegisterBackend: %v", err)
	}
	_, err := svc.GeneratePresignedURL(ctx, "s3-nocreds", "key.flac", DefaultPresignedURLTTL)
	if !errors.Is(err, ErrProbeFailed) {
		t.Fatalf("GeneratePresignedURL(missing creds) = %v, want ErrProbeFailed", err)
	}
}

func TestGeneratePresignedURLSuccess(t *testing.T) {
	ctx := context.Background()
	t.Setenv("PS2_ACCESS", "access-key")
	t.Setenv("PS2_SECRET", "secret-key")
	svc := NewService(NewMemoryRepository())
	svc.now = func() time.Time { return time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC) }
	b := StorageBackend{
		ID:          "s3-presign",
		Type:        BackendTypeS3,
		DisplayName: "S3 Presign",
		Enabled:     true,
		Capabilities: CapabilitySet{PresignedURLs: true},
		Config: BackendConfig{S3: &S3Config{
			Endpoint:           "https://s3.example.com",
			Region:             "us-east-1",
			Bucket:             "music",
			PathStyle:          true,
			AccessKeySecretRef: "PS2_ACCESS",
			SecretKeySecretRef: "PS2_SECRET",
		}},
	}
	if _, err := svc.RegisterBackend(ctx, b); err != nil {
		t.Fatalf("RegisterBackend: %v", err)
	}
	purl, err := svc.GeneratePresignedURL(ctx, "s3-presign", "tracks/song.flac", DefaultPresignedURLTTL)
	if err != nil {
		t.Fatalf("GeneratePresignedURL: %v", err)
	}
	if !strings.Contains(purl, "X-Amz-Signature") {
		t.Errorf("presigned URL missing X-Amz-Signature: %s", purl)
	}
	if !strings.Contains(purl, "tracks") {
		t.Errorf("presigned URL missing object key: %s", purl)
	}
}

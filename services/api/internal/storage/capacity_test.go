package storage

import (
	"context"
	"errors"
	"testing"
)

func TestFilesystemCapacityProvider(t *testing.T) {
	backend := StorageBackend{ID: "local-main", Type: BackendTypeLocal, Config: BackendConfig{Local: &LocalConfig{RootPath: t.TempDir()}}}
	report, err := NewFilesystemCapacityProvider().Capacity(context.Background(), backend)
	if err != nil {
		t.Fatalf("Capacity() error = %v", err)
	}
	if report.TotalBytes == 0 || report.AvailableBytes > report.TotalBytes || report.UsedBytes != report.TotalBytes-report.AvailableBytes {
		t.Fatalf("Capacity() report = %+v, want coherent filesystem statistics", report)
	}
}

func TestFilesystemCapacityProviderRejectsS3(t *testing.T) {
	backend := StorageBackend{ID: "s3-main", Type: BackendTypeS3, Config: BackendConfig{S3: &S3Config{}}}
	_, err := NewFilesystemCapacityProvider().Capacity(context.Background(), backend)
	if !errors.Is(err, ErrCapacityUnsupported) {
		t.Fatalf("Capacity() error = %v, want ErrCapacityUnsupported", err)
	}
}

func TestServiceGetBackendCapacityRecordsLatestReport(t *testing.T) {
	ctx := context.Background()
	repository := NewMemoryRepository()
	service := NewService(repository)
	_, err := service.RegisterBackend(ctx, StorageBackend{
		ID: "local-main", Type: BackendTypeLocal, DisplayName: "Local", Enabled: true,
		Config: BackendConfig{Local: &LocalConfig{RootPath: t.TempDir()}},
	})
	if err != nil {
		t.Fatalf("RegisterBackend() error = %v", err)
	}
	report, err := service.GetBackendCapacity(ctx, "local-main")
	if err != nil {
		t.Fatalf("GetBackendCapacity() error = %v", err)
	}
	backend, err := repository.Get(ctx, "local-main")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if backend.LastCapacity == nil || !backend.LastCapacity.CheckedAt.Equal(report.CheckedAt) {
		t.Fatalf("LastCapacity = %+v, want report %+v", backend.LastCapacity, report)
	}
}

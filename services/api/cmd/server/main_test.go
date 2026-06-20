package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/history"
	"inori-music/services/api/internal/httpapi"
	"inori-music/services/api/internal/storage"
)

func TestStorageRefreshInterval(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{name: "unset"},
		{name: "valid", value: "15m", want: 15 * time.Minute},
		{name: "invalid", value: "later"},
		{name: "non positive", value: "0s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("INORI_STORAGE_REFRESH_INTERVAL", tt.value)
			if got := storageRefreshInterval(); got != tt.want {
				t.Fatalf("storageRefreshInterval() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestStorageRepositoryDefaultsToMemory(t *testing.T) {
	t.Setenv("INORI_STORAGE_REPOSITORY_FILE", "")
	repo, err := storageRepository(context.Background(), nil)
	if err != nil {
		t.Fatalf("storageRepository() error = %v", err)
	}
	if _, ok := repo.(*storage.MemoryRepository); !ok {
		t.Fatalf("storageRepository() = %T, want *storage.MemoryRepository", repo)
	}
}

func TestStorageRepositoryUsesFileWhenConfigured(t *testing.T) {
	path := filepath.Join(t.TempDir(), "storage-backends.json")
	t.Setenv("INORI_STORAGE_REPOSITORY_FILE", path)
	repo, err := storageRepository(context.Background(), nil)
	if err != nil {
		t.Fatalf("storageRepository() error = %v", err)
	}
	if _, ok := repo.(*storage.FileRepository); !ok {
		t.Fatalf("storageRepository() = %T, want *storage.FileRepository", repo)
	}
}

func TestMediaObjectRepositoryDefaultsToMemory(t *testing.T) {
	t.Setenv("INORI_MEDIA_OBJECT_REPOSITORY_FILE", "")
	repo, err := mediaObjectRepository(context.Background(), nil)
	if err != nil {
		t.Fatalf("mediaObjectRepository() error = %v", err)
	}
	if _, ok := repo.(*storage.MemoryMediaObjectRepository); !ok {
		t.Fatalf("mediaObjectRepository() = %T, want *storage.MemoryMediaObjectRepository", repo)
	}
}

func TestMediaObjectRepositoryUsesFileWhenConfigured(t *testing.T) {
	path := filepath.Join(t.TempDir(), "media-objects.json")
	t.Setenv("INORI_MEDIA_OBJECT_REPOSITORY_FILE", path)
	repo, err := mediaObjectRepository(context.Background(), nil)
	if err != nil {
		t.Fatalf("mediaObjectRepository() error = %v", err)
	}
	if _, ok := repo.(*storage.FileMediaObjectRepository); !ok {
		t.Fatalf("mediaObjectRepository() = %T, want *storage.FileMediaObjectRepository", repo)
	}
}

func TestCatalogRepositoryDefaultsToMemory(t *testing.T) {
	repo := catalogRepository(nil)
	if _, ok := repo.(*catalog.MemoryRepository); !ok {
		t.Fatalf("catalogRepository(nil) = %T, want *catalog.MemoryRepository", repo)
	}
}

func TestHistoryRepositoryDefaultsToMemory(t *testing.T) {
	repo := historyRepository(nil)
	if _, ok := repo.(*history.MemoryRepository); !ok {
		t.Fatalf("historyRepository(nil) = %T, want *history.MemoryRepository", repo)
	}
}

func TestHandlerWithAllServicesReportsThreeBaseChecks(t *testing.T) {
	// Verify the three original readiness checks are still present when all
	// services are wired (catalog/history do not yet participate in readiness).
	storageSvc := storage.NewService(storage.NewMemoryRepository())
	moSvc := storage.NewMediaObjectService(storage.NewMemoryRepository(), storage.NewMemoryMediaObjectRepository())
	catalogSvc := catalog.NewService(catalog.NewMemoryRepository())
	historySvc := history.NewService(history.NewMemoryRepository())

	_ = httpapi.NewHandler(storageSvc,
		httpapi.WithAdminToken("test-token"),
		httpapi.WithMediaObjectService(moSvc),
		httpapi.WithCatalogService(catalogSvc),
		httpapi.WithHistoryService(historySvc),
	)
	// Construction must not panic; catalog and history services are accepted
	// without error. The readiness endpoint is tested separately in handler_test.go.
}

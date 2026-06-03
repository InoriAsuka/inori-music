package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMediaObjectServiceVerifiesFilesystemObject(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	content := []byte("hello inori")
	objectPath := filepath.Join(root, "albums", "track.flac")
	if err := os.MkdirAll(filepath.Dir(objectPath), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(objectPath, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	service := newVerificationService(t, root, content)

	result, err := service.VerifyMediaObject(ctx, "media-original-1")
	if err != nil {
		t.Fatalf("VerifyMediaObject() error = %v", err)
	}
	if result.Status != "verified" || result.SizeBytes != int64(len(content)) || result.VerifiedAt.IsZero() {
		t.Fatalf("VerifyMediaObject() = %+v, want verified result", result)
	}
}

func TestMediaObjectServiceVerificationDetectsHashMismatch(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	objectPath := filepath.Join(root, "albums", "track.flac")
	if err := os.MkdirAll(filepath.Dir(objectPath), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(objectPath, []byte("actual"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	service := newVerificationService(t, root, []byte("expected"))

	result, err := service.VerifyMediaObject(ctx, "media-original-1")
	if !errors.Is(err, ErrMediaObjectVerificationFailed) {
		t.Fatalf("VerifyMediaObject() error = %v, want ErrMediaObjectVerificationFailed", err)
	}
	if result.Status != "failed" || result.Message == "" {
		t.Fatalf("VerifyMediaObject() result = %+v, want failed message", result)
	}
}

func TestMediaObjectServiceVerificationRejectsDisabledBackend(t *testing.T) {
	ctx := context.Background()
	backendRepo := NewMemoryRepository()
	if err := backendRepo.Save(ctx, StorageBackend{ID: "local-main", Enabled: false, Config: BackendConfig{Local: &LocalConfig{RootPath: t.TempDir()}}}); err != nil {
		t.Fatalf("Save(backend) error = %v", err)
	}
	mediaRepo := NewMemoryMediaObjectRepository()
	object := validMediaObject()
	object.ObjectKey = "albums/track.flac"
	if err := mediaRepo.SaveMediaObject(ctx, object); err != nil {
		t.Fatalf("SaveMediaObject() error = %v", err)
	}
	service := NewMediaObjectService(backendRepo, mediaRepo)

	_, err := service.VerifyMediaObject(ctx, object.ID)
	if !errors.Is(err, ErrBackendDisabled) {
		t.Fatalf("VerifyMediaObject() error = %v, want ErrBackendDisabled", err)
	}
}

func TestMediaObjectServiceVerificationRejectsUnsupportedBackend(t *testing.T) {
	ctx := context.Background()
	backendRepo := NewMemoryRepository()
	if err := backendRepo.Save(ctx, StorageBackend{ID: "s3-main", Type: BackendTypeS3, Enabled: true, Config: BackendConfig{S3: &S3Config{Endpoint: "https://s3.example.com", Bucket: "inori", AccessKeySecretRef: "A", SecretKeySecretRef: "S"}}}); err != nil {
		t.Fatalf("Save(backend) error = %v", err)
	}
	mediaRepo := NewMemoryMediaObjectRepository()
	object := validMediaObject()
	object.BackendID = "s3-main"
	if err := mediaRepo.SaveMediaObject(ctx, object); err != nil {
		t.Fatalf("SaveMediaObject() error = %v", err)
	}
	service := NewMediaObjectService(backendRepo, mediaRepo)

	_, err := service.VerifyMediaObject(ctx, object.ID)
	if !errors.Is(err, ErrMediaObjectVerificationUnsupported) {
		t.Fatalf("VerifyMediaObject() error = %v, want ErrMediaObjectVerificationUnsupported", err)
	}
}

func TestMediaObjectServiceVerificationRejectsUnsupportedHashAlgorithm(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	backendRepo := NewMemoryRepository()
	if err := backendRepo.Save(ctx, StorageBackend{ID: "local-main", Type: BackendTypeLocal, Enabled: true, Config: BackendConfig{Local: &LocalConfig{RootPath: root}}}); err != nil {
		t.Fatalf("Save(backend) error = %v", err)
	}
	mediaRepo := NewMemoryMediaObjectRepository()
	object := validMediaObject()
	object.ContentHash = "blake3:abc"
	if err := mediaRepo.SaveMediaObject(ctx, object); err != nil {
		t.Fatalf("SaveMediaObject() error = %v", err)
	}
	service := NewMediaObjectService(backendRepo, mediaRepo)

	_, err := service.VerifyMediaObject(ctx, object.ID)
	if !errors.Is(err, ErrMediaObjectVerificationUnsupported) {
		t.Fatalf("VerifyMediaObject() error = %v, want ErrMediaObjectVerificationUnsupported", err)
	}
}

func newVerificationService(t *testing.T, root string, expectedContent []byte) *MediaObjectService {
	t.Helper()
	ctx := context.Background()
	backendRepo := NewMemoryRepository()
	if err := backendRepo.Save(ctx, StorageBackend{ID: "local-main", Type: BackendTypeLocal, Enabled: true, Config: BackendConfig{Local: &LocalConfig{RootPath: root}}}); err != nil {
		t.Fatalf("Save(backend) error = %v", err)
	}
	mediaRepo := NewMemoryMediaObjectRepository()
	sum := sha256.Sum256(expectedContent)
	object := validMediaObject()
	object.ObjectKey = "albums/track.flac"
	object.ContentHash = "sha256:" + hex.EncodeToString(sum[:])
	object.SizeBytes = int64(len(expectedContent))
	if err := mediaRepo.SaveMediaObject(ctx, object); err != nil {
		t.Fatalf("SaveMediaObject() error = %v", err)
	}
	return NewMediaObjectService(backendRepo, mediaRepo)
}

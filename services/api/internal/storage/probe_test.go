package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFilesystemProberProbeAndCleanup(t *testing.T) {
	root := t.TempDir()
	backend := StorageBackend{
		ID:     "local-main",
		Type:   BackendTypeLocal,
		Config: BackendConfig{Local: &LocalConfig{RootPath: root}},
	}

	if err := NewFilesystemProber().Probe(context.Background(), backend); err != nil {
		t.Fatalf("Probe() error = %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("probe root contains %d entries after cleanup, want 0", len(entries))
	}
}

func TestFilesystemProberSupportsMountedBackendFamilies(t *testing.T) {
	root := t.TempDir()
	tests := []StorageBackend{
		{ID: "nfs", Type: BackendTypeNFS, Config: BackendConfig{NFS: &NFSConfig{MountPath: root}}},
		{ID: "smb", Type: BackendTypeSMB, Config: BackendConfig{SMB: &SMBConfig{MountPath: root}}},
		{ID: "distributed", Type: BackendTypeDistributed, Config: BackendConfig{Distributed: &DistributedConfig{Adapter: "mounted-filesystem", MountPath: root}}},
	}

	for _, backend := range tests {
		t.Run(backend.ID, func(t *testing.T) {
			if err := NewFilesystemProber().Probe(context.Background(), backend); err != nil {
				t.Fatalf("Probe() error = %v", err)
			}
		})
	}
}

func TestFilesystemProberRejectsMissingRoot(t *testing.T) {
	backend := StorageBackend{
		ID:     "missing",
		Type:   BackendTypeLocal,
		Config: BackendConfig{Local: &LocalConfig{RootPath: filepath.Join(t.TempDir(), "missing")}},
	}

	err := NewFilesystemProber().Probe(context.Background(), backend)
	if !errors.Is(err, ErrProbeFailed) {
		t.Fatalf("Probe() error = %v, want ErrProbeFailed", err)
	}
}

func TestFilesystemProberRejectsUnsupportedS3(t *testing.T) {
	backend := StorageBackend{ID: "s3", Type: BackendTypeS3, Config: BackendConfig{S3: &S3Config{}}}
	err := NewFilesystemProber().Probe(context.Background(), backend)
	if !errors.Is(err, ErrProbeUnsupported) {
		t.Fatalf("Probe() error = %v, want ErrProbeUnsupported", err)
	}
}

package storage

import (
	"errors"
	"testing"
)

func TestValidateBackendFamilies(t *testing.T) {
	tests := []struct {
		name     string
		backend  StorageBackend
		assertFn func(t *testing.T, capabilities CapabilitySet)
	}{
		{
			name:    "local",
			backend: StorageBackend{ID: "local-main", Type: BackendTypeLocal, DisplayName: "Local", Enabled: true, Config: BackendConfig{Local: &LocalConfig{RootPath: "/srv/inori/media"}}},
			assertFn: func(t *testing.T, capabilities CapabilitySet) {
				if !capabilities.ServerRangeReads || capabilities.CrossNodeAccess {
					t.Fatalf("unexpected local capabilities: %+v", capabilities)
				}
			},
		},
		{
			name:    "nfs",
			backend: StorageBackend{ID: "nfs-main", Type: BackendTypeNFS, DisplayName: "NFS", Enabled: true, Config: BackendConfig{NFS: &NFSConfig{MountPath: "/mnt/inori", ExpectedRemote: "10.0.0.2:/exports/inori"}}},
			assertFn: func(t *testing.T, capabilities CapabilitySet) {
				if !capabilities.ServerRangeReads || !capabilities.CrossNodeAccess || !capabilities.RequiresMountValidation {
					t.Fatalf("unexpected nfs capabilities: %+v", capabilities)
				}
			},
		},
		{
			name:    "smb",
			backend: StorageBackend{ID: "smb-main", Type: BackendTypeSMB, DisplayName: "SMB", Enabled: true, Config: BackendConfig{SMB: &SMBConfig{MountPath: "/mnt/smb/inori", ExpectedShare: "//nas/music"}}},
			assertFn: func(t *testing.T, capabilities CapabilitySet) {
				if !capabilities.ServerRangeReads || !capabilities.CrossNodeAccess || !capabilities.RequiresMountValidation {
					t.Fatalf("unexpected smb capabilities: %+v", capabilities)
				}
			},
		},
		{
			name:    "s3",
			backend: StorageBackend{ID: "s3-main", Type: BackendTypeS3, DisplayName: "S3", Enabled: true, Config: BackendConfig{S3: &S3Config{Endpoint: "https://s3.example.com", Bucket: "inori", AccessKeySecretRef: "S3_ACCESS", SecretKeySecretRef: "S3_SECRET"}}},
			assertFn: func(t *testing.T, capabilities CapabilitySet) {
				if !capabilities.PresignedURLs || !capabilities.MultipartUpload || !capabilities.RequiresCredentialValidation {
					t.Fatalf("unexpected s3 capabilities: %+v", capabilities)
				}
			},
		},
		{
			name:    "distributed s3 compatible",
			backend: StorageBackend{ID: "ceph-main", Type: BackendTypeDistributed, DisplayName: "Ceph", Enabled: true, Config: BackendConfig{Distributed: &DistributedConfig{Adapter: "s3-compatible", Endpoint: "https://rgw.example.com", Bucket: "inori"}}},
			assertFn: func(t *testing.T, capabilities CapabilitySet) {
				if !capabilities.CrossNodeAccess || !capabilities.PresignedURLs || !capabilities.RequiresCredentialValidation {
					t.Fatalf("unexpected distributed capabilities: %+v", capabilities)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := tt.backend
			if err := ValidateBackend(&backend); err != nil {
				t.Fatalf("ValidateBackend() error = %v", err)
			}
			tt.assertFn(t, backend.Capabilities)
		})
	}
}

func TestValidateBackendRejectsUnsafeOrIncompleteConfig(t *testing.T) {
	tests := []StorageBackend{
		{ID: "missing-config", Type: BackendTypeLocal, DisplayName: "Missing", Enabled: true},
		{ID: "multiple-configs", Type: BackendTypeLocal, DisplayName: "Multiple", Enabled: true, Config: BackendConfig{Local: &LocalConfig{RootPath: "/srv/inori"}, SMB: &SMBConfig{MountPath: "/mnt/smb", ExpectedShare: "//nas/music"}}},
		{ID: "mismatched-config", Type: BackendTypeLocal, DisplayName: "Mismatch", Enabled: true, Config: BackendConfig{NFS: &NFSConfig{MountPath: "/mnt/nfs", ExpectedRemote: "nas:/music"}}},
		{ID: "root-path", Type: BackendTypeLocal, DisplayName: "Root", Enabled: true, Config: BackendConfig{Local: &LocalConfig{RootPath: "/"}}},
		{ID: "bad-smb", Type: BackendTypeSMB, DisplayName: "Bad SMB", Enabled: true, Config: BackendConfig{SMB: &SMBConfig{MountPath: "/mnt/smb", ExpectedShare: "nas/music"}}},
		{ID: "bad-s3", Type: BackendTypeS3, DisplayName: "Bad S3", Enabled: true, Config: BackendConfig{S3: &S3Config{Endpoint: "ftp://s3.example.com", Bucket: "inori", AccessKeySecretRef: "A", SecretKeySecretRef: "S"}}},
		{ID: "bad-distributed", Type: BackendTypeDistributed, DisplayName: "Bad Distributed", Enabled: true, Config: BackendConfig{Distributed: &DistributedConfig{Adapter: "unknown"}}},
	}

	for _, backend := range tests {
		t.Run(backend.ID, func(t *testing.T) {
			candidate := backend
			err := ValidateBackend(&candidate)
			if !errors.Is(err, ErrInvalidBackend) {
				t.Fatalf("ValidateBackend() error = %v, want ErrInvalidBackend", err)
			}
		})
	}
}

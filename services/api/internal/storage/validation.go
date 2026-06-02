package storage

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidBackend = errors.New("invalid storage backend")
	ErrNotFound       = errors.New("storage backend not found")
	ErrConflict       = errors.New("storage backend conflict")
)

// ValidateBackend checks static configuration and fills inferred capabilities.
func ValidateBackend(backend *StorageBackend) error {
	if backend == nil {
		return fmt.Errorf("%w: backend is required", ErrInvalidBackend)
	}

	backend.ID = strings.TrimSpace(backend.ID)
	backend.DisplayName = strings.TrimSpace(backend.DisplayName)
	if backend.ID == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidBackend)
	}
	if backend.DisplayName == "" {
		return fmt.Errorf("%w: display name is required", ErrInvalidBackend)
	}

	if err := validateTypedConfig(backend.Type, backend.Config); err != nil {
		return err
	}

	backend.Capabilities = InferCapabilities(backend.Type, backend.Config)
	if backend.HealthStatus == "" {
		backend.HealthStatus = HealthStatusUnknown
	}
	if !backend.Enabled {
		backend.HealthStatus = HealthStatusDisabled
	}

	return nil
}

// InferCapabilities returns the conservative capability set for a backend family.
func InferCapabilities(backendType BackendType, config BackendConfig) CapabilitySet {
	switch backendType {
	case BackendTypeLocal:
		return CapabilitySet{ServerRangeReads: true}
	case BackendTypeNFS, BackendTypeSMB:
		return CapabilitySet{
			ServerRangeReads:        true,
			CrossNodeAccess:         true,
			RequiresMountValidation: true,
		}
	case BackendTypeS3:
		return CapabilitySet{
			ServerRangeReads:             true,
			PresignedURLs:                true,
			MultipartUpload:              true,
			NativeLifecyclePolicy:        true,
			CrossNodeAccess:              true,
			RequiresCredentialValidation: true,
		}
	case BackendTypeDistributed:
		capabilities := CapabilitySet{CrossNodeAccess: true}
		if config.Distributed == nil {
			return capabilities
		}
		switch config.Distributed.Adapter {
		case "s3-compatible":
			capabilities.ServerRangeReads = true
			capabilities.PresignedURLs = true
			capabilities.MultipartUpload = true
			capabilities.NativeLifecyclePolicy = true
			capabilities.RequiresCredentialValidation = true
		case "mounted-filesystem":
			capabilities.ServerRangeReads = true
			capabilities.RequiresMountValidation = true
		case "dedicated":
			capabilities.ServerRangeReads = true
			capabilities.RequiresCredentialValidation = true
		}
		return capabilities
	default:
		return CapabilitySet{}
	}
}

func validateTypedConfig(backendType BackendType, config BackendConfig) error {
	switch backendType {
	case BackendTypeLocal:
		if config.Local == nil {
			return fmt.Errorf("%w: local config is required", ErrInvalidBackend)
		}
		return validateSafePath("local root path", config.Local.RootPath)
	case BackendTypeNFS:
		if config.NFS == nil {
			return fmt.Errorf("%w: nfs config is required", ErrInvalidBackend)
		}
		if err := validateMountPath("nfs mount path", config.NFS.MountPath); err != nil {
			return err
		}
		if strings.TrimSpace(config.NFS.ExpectedRemote) == "" {
			return fmt.Errorf("%w: nfs expected remote is required", ErrInvalidBackend)
		}
		return nil
	case BackendTypeSMB:
		if config.SMB == nil {
			return fmt.Errorf("%w: smb config is required", ErrInvalidBackend)
		}
		if err := validateMountPath("smb mount path", config.SMB.MountPath); err != nil {
			return err
		}
		if !strings.HasPrefix(strings.TrimSpace(config.SMB.ExpectedShare), "//") {
			return fmt.Errorf("%w: smb expected share must start with //", ErrInvalidBackend)
		}
		return nil
	case BackendTypeS3:
		if config.S3 == nil {
			return fmt.Errorf("%w: s3 config is required", ErrInvalidBackend)
		}
		return validateS3Config(*config.S3)
	case BackendTypeDistributed:
		if config.Distributed == nil {
			return fmt.Errorf("%w: distributed config is required", ErrInvalidBackend)
		}
		return validateDistributedConfig(*config.Distributed)
	default:
		return fmt.Errorf("%w: unsupported backend type %q", ErrInvalidBackend, backendType)
	}
}

func validateSafePath(label string, value string) error {
	path := strings.TrimSpace(value)
	if path == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidBackend, label)
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == string(filepath.Separator) {
		return fmt.Errorf("%w: %s must not be repository current directory or filesystem root", ErrInvalidBackend, label)
	}
	return nil
}

func validateMountPath(label string, value string) error {
	if err := validateSafePath(label, value); err != nil {
		return err
	}
	if !filepath.IsAbs(strings.TrimSpace(value)) {
		return fmt.Errorf("%w: %s must be absolute", ErrInvalidBackend, label)
	}
	return nil
}

func validateS3Config(config S3Config) error {
	if err := validateEndpoint("s3 endpoint", config.Endpoint); err != nil {
		return err
	}
	if strings.TrimSpace(config.Bucket) == "" {
		return fmt.Errorf("%w: s3 bucket is required", ErrInvalidBackend)
	}
	if strings.TrimSpace(config.AccessKeySecretRef) == "" || strings.TrimSpace(config.SecretKeySecretRef) == "" {
		return fmt.Errorf("%w: s3 secret references are required", ErrInvalidBackend)
	}
	return nil
}

func validateDistributedConfig(config DistributedConfig) error {
	switch strings.TrimSpace(config.Adapter) {
	case "s3-compatible":
		if err := validateEndpoint("distributed endpoint", config.Endpoint); err != nil {
			return err
		}
		if strings.TrimSpace(config.Bucket) == "" {
			return fmt.Errorf("%w: distributed bucket is required", ErrInvalidBackend)
		}
	case "mounted-filesystem":
		if err := validateMountPath("distributed mount path", config.MountPath); err != nil {
			return err
		}
	case "dedicated":
		if strings.TrimSpace(config.Endpoint) == "" {
			return fmt.Errorf("%w: distributed endpoint is required", ErrInvalidBackend)
		}
	default:
		return fmt.Errorf("%w: distributed adapter must be s3-compatible, mounted-filesystem, or dedicated", ErrInvalidBackend)
	}
	return nil
}

func validateEndpoint(label string, raw string) error {
	endpoint := strings.TrimSpace(raw)
	if endpoint == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidBackend, label)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("%w: %s must be an absolute URL", ErrInvalidBackend, label)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%w: %s must use http or https", ErrInvalidBackend, label)
	}
	return nil
}

package storage

import "time"

// BackendType identifies the family of a server-managed media storage backend.
type BackendType string

const (
	BackendTypeLocal       BackendType = "local"
	BackendTypeNFS         BackendType = "nfs"
	BackendTypeSMB         BackendType = "smb"
	BackendTypeS3          BackendType = "s3"
	BackendTypeDistributed BackendType = "distributed"
)

// HealthStatus captures the last known operational state of a backend.
type HealthStatus string

const (
	HealthStatusUnknown   HealthStatus = "unknown"
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDisabled  HealthStatus = "disabled"
)

// CapabilitySet describes the operations the service can safely use for a backend.
type CapabilitySet struct {
	ServerRangeReads             bool
	PresignedURLs                bool
	MultipartUpload              bool
	NativeLifecyclePolicy        bool
	CrossNodeAccess              bool
	RequiresMountValidation      bool
	RequiresCredentialValidation bool
}

// StorageBackend is the server-owned configuration record for a media backend.
type StorageBackend struct {
	ID                string
	Type              BackendType
	DisplayName       string
	Enabled           bool
	IsDefault         bool
	Priority          int
	HealthStatus      HealthStatus
	LastHealthCheckAt *time.Time
	Capabilities      CapabilitySet
	Config            BackendConfig
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// BackendConfig stores one backend-family-specific configuration.
type BackendConfig struct {
	Local       *LocalConfig
	NFS         *NFSConfig
	SMB         *SMBConfig
	S3          *S3Config
	Distributed *DistributedConfig
}

type LocalConfig struct {
	RootPath string
}

type NFSConfig struct {
	MountPath      string
	ExpectedRemote string
}

type SMBConfig struct {
	MountPath     string
	ExpectedShare string
}

type S3Config struct {
	Endpoint           string
	Region             string
	Bucket             string
	PathStyle          bool
	AccessKeySecretRef string
	SecretKeySecretRef string
}

type DistributedConfig struct {
	Adapter   string
	Endpoint  string
	Bucket    string
	MountPath string
}

// MediaObject references a stored binary asset without embedding large media in the database.
type MediaObject struct {
	ID             string
	BackendID      string
	ObjectKey      string
	ContentHash    string
	SizeBytes      int64
	MIMEType       string
	AssetKind      string
	LifecycleState string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

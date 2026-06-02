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
	ServerRangeReads             bool `json:"serverRangeReads"`
	PresignedURLs                bool `json:"presignedUrls"`
	MultipartUpload              bool `json:"multipartUpload"`
	NativeLifecyclePolicy        bool `json:"nativeLifecyclePolicy"`
	CrossNodeAccess              bool `json:"crossNodeAccess"`
	RequiresMountValidation      bool `json:"requiresMountValidation"`
	RequiresCredentialValidation bool `json:"requiresCredentialValidation"`
}

// StorageBackend is the server-owned configuration record for a media backend.
type StorageBackend struct {
	ID                string        `json:"id"`
	Type              BackendType   `json:"type"`
	DisplayName       string        `json:"displayName"`
	Enabled           bool          `json:"enabled"`
	IsDefault         bool          `json:"isDefault"`
	Priority          int           `json:"priority"`
	HealthStatus      HealthStatus  `json:"healthStatus"`
	LastHealthCheckAt *time.Time    `json:"lastHealthCheckAt,omitempty"`
	Capabilities      CapabilitySet `json:"capabilities"`
	Config            BackendConfig `json:"config"`
	CreatedAt         time.Time     `json:"createdAt,omitempty"`
	UpdatedAt         time.Time     `json:"updatedAt,omitempty"`
}

// BackendConfig stores exactly one backend-family-specific configuration.
type BackendConfig struct {
	Local       *LocalConfig       `json:"local,omitempty"`
	NFS         *NFSConfig         `json:"nfs,omitempty"`
	SMB         *SMBConfig         `json:"smb,omitempty"`
	S3          *S3Config          `json:"s3,omitempty"`
	Distributed *DistributedConfig `json:"distributed,omitempty"`
}

type LocalConfig struct {
	RootPath string `json:"rootPath"`
}

type NFSConfig struct {
	MountPath      string `json:"mountPath"`
	ExpectedRemote string `json:"expectedRemote"`
}

type SMBConfig struct {
	MountPath     string `json:"mountPath"`
	ExpectedShare string `json:"expectedShare"`
}

type S3Config struct {
	Endpoint           string `json:"endpoint"`
	Region             string `json:"region,omitempty"`
	Bucket             string `json:"bucket"`
	PathStyle          bool   `json:"pathStyle"`
	AccessKeySecretRef string `json:"accessKeySecretRef"`
	SecretKeySecretRef string `json:"secretKeySecretRef"`
}

type DistributedConfig struct {
	Adapter   string `json:"adapter"`
	Endpoint  string `json:"endpoint,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	MountPath string `json:"mountPath,omitempty"`
}

// MediaObject references a stored binary asset without embedding large media in the database.
type MediaObject struct {
	ID             string    `json:"id"`
	BackendID      string    `json:"backendId"`
	ObjectKey      string    `json:"objectKey"`
	ContentHash    string    `json:"contentHash"`
	SizeBytes      int64     `json:"sizeBytes"`
	MIMEType       string    `json:"mimeType"`
	AssetKind      string    `json:"assetKind"`
	LifecycleState string    `json:"lifecycleState"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

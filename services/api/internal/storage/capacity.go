package storage

import (
	"context"
	"errors"
	"time"
)

var ErrCapacityUnsupported = errors.New("storage backend capacity is unsupported")

// CapacityReport records the latest observed storage capacity where the backend exposes it.
type CapacityReport struct {
	BackendID      string    `json:"backendId"`
	TotalBytes     uint64    `json:"totalBytes"`
	AvailableBytes uint64    `json:"availableBytes"`
	UsedBytes      uint64    `json:"usedBytes"`
	CheckedAt      time.Time `json:"checkedAt"`
}

// CapacityProvider reads backend capacity where a portable backend-family implementation exists.
type CapacityProvider interface {
	Capacity(ctx context.Context, backend StorageBackend) (CapacityReport, error)
}

// FilesystemCapacityProvider reads mounted filesystem statistics.
type FilesystemCapacityProvider struct{}

func NewFilesystemCapacityProvider() *FilesystemCapacityProvider {
	return &FilesystemCapacityProvider{}
}

func (provider *FilesystemCapacityProvider) Capacity(ctx context.Context, backend StorageBackend) (CapacityReport, error) {
	if err := ctx.Err(); err != nil {
		return CapacityReport{}, err
	}
	root, err := filesystemProbeRoot(backend)
	if err != nil {
		return CapacityReport{}, fmt.Errorf("%w: %v", ErrCapacityUnsupported, err)
	}
	var stats syscall.Statfs_t
	if err := syscall.Statfs(root, &stats); err != nil {
		return CapacityReport{}, fmt.Errorf("read filesystem capacity: %w", err)
	}
	total := stats.Blocks * uint64(stats.Bsize)
	available := stats.Bavail * uint64(stats.Bsize)
	used := uint64(0)
	if total > available {
		used = total - available
	}
	return CapacityReport{
		BackendID:      backend.ID,
		TotalBytes:     total,
		AvailableBytes: available,
		UsedBytes:      used,
	}, nil
}

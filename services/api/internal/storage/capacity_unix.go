//go:build unix

package storage

import (
	"context"
	"fmt"
	"syscall"
)

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
	return CapacityReport{
		BackendID:      backend.ID,
		TotalBytes:     total,
		AvailableBytes: available,
		UsedBytes:      total - available,
	}, nil
}

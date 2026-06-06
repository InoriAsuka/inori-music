//go:build !unix

package storage

import (
	"context"
	"fmt"
)

// FilesystemCapacityProvider reports unsupported capacity on platforms without Unix statfs support.
type FilesystemCapacityProvider struct{}

func NewFilesystemCapacityProvider() *FilesystemCapacityProvider {
	return &FilesystemCapacityProvider{}
}

func (provider *FilesystemCapacityProvider) Capacity(ctx context.Context, backend StorageBackend) (CapacityReport, error) {
	if err := ctx.Err(); err != nil {
		return CapacityReport{}, err
	}
	if _, err := filesystemProbeRoot(backend); err != nil {
		return CapacityReport{}, fmt.Errorf("%w: %v", ErrCapacityUnsupported, err)
	}
	return CapacityReport{}, fmt.Errorf("%w: filesystem capacity is not implemented on this platform", ErrCapacityUnsupported)
}

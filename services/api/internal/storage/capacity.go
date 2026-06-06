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

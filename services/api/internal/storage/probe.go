package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrProbeUnsupported = errors.New("storage backend probe is unsupported")
	ErrBackendDisabled  = errors.New("storage backend is disabled")
	ErrProbeFailed      = errors.New("storage backend probe failed")
)

var probePayload = []byte("inori-music-storage-probe")

// ProbeResult records the latest externally observed backend health state.
type ProbeResult struct {
	BackendID string       `json:"backendId"`
	Status    HealthStatus `json:"status"`
	CheckedAt time.Time    `json:"checkedAt"`
	Message   string       `json:"message,omitempty"`
}

// Prober verifies that a configured backend can perform the minimum operations needed by media storage.
type Prober interface {
	Probe(ctx context.Context, backend StorageBackend) error
}

// FilesystemProber verifies filesystem-backed storage with an application-owned temporary file.
type FilesystemProber struct{}

func NewFilesystemProber() *FilesystemProber {
	return &FilesystemProber{}
}

func (prober *FilesystemProber) Probe(ctx context.Context, backend StorageBackend) error {
	root, err := filesystemProbeRoot(backend)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrProbeFailed, err)
	}

	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("%w: stat root: %v", ErrProbeFailed, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: root is not a directory", ErrProbeFailed)
	}

	// Read-only local backends (e.g. a music library bind-mounted `:ro` — the
	// common case for a music server) can never pass the write-based probe
	// below by design, not because anything is actually wrong. Stop here with
	// a read-only reachability check instead of always reporting unhealthy.
	// Writable backends (including local backends that don't set ReadOnly)
	// fall through to the original write/read/range-read probe unchanged.
	if backend.Type == BackendTypeLocal && backend.Config.Local != nil && backend.Config.Local.ReadOnly {
		return probeReadableDirectory(root)
	}

	file, err := os.CreateTemp(root, ".inori-music-probe-*")
	if err != nil {
		return fmt.Errorf("%w: create probe file: %v", ErrProbeFailed, err)
	}
	probePath := file.Name()
	defer os.Remove(probePath)
	defer file.Close()

	if _, err := file.Write(probePayload); err != nil {
		return fmt.Errorf("%w: write probe file: %v", ErrProbeFailed, err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("%w: sync probe file: %v", ErrProbeFailed, err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("%w: seek probe file: %v", ErrProbeFailed, err)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("%w: read probe file: %v", ErrProbeFailed, err)
	}
	if !bytes.Equal(content, probePayload) {
		return fmt.Errorf("%w: full read content mismatch", ErrProbeFailed)
	}

	rangeBuffer := make([]byte, 5)
	if _, err := file.ReadAt(rangeBuffer, 6); err != nil {
		return fmt.Errorf("%w: range read probe file: %v", ErrProbeFailed, err)
	}
	if !bytes.Equal(rangeBuffer, probePayload[6:11]) {
		return fmt.Errorf("%w: range read content mismatch", ErrProbeFailed)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("%w: close probe file: %v", ErrProbeFailed, err)
	}
	if err := os.Remove(probePath); err != nil {
		return fmt.Errorf("%w: remove probe file: %v", ErrProbeFailed, err)
	}
	return nil
}

// probeReadableDirectory verifies a mount is reachable using only read
// operations: open the directory and list at least one entry. An empty
// directory (io.EOF from Readdirnames) is a healthy, valid result — a
// freshly mounted read-only library legitimately has nothing in it yet, and
// that is not itself a reason to report the backend unhealthy.
func probeReadableDirectory(root string) error {
	dir, err := os.Open(root)
	if err != nil {
		return fmt.Errorf("%w: open root for read: %v", ErrProbeFailed, err)
	}
	defer dir.Close()

	if _, err := dir.Readdirnames(1); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: list root: %v", ErrProbeFailed, err)
	}
	return nil
}

func filesystemProbeRoot(backend StorageBackend) (string, error) {
	switch backend.Type {
	case BackendTypeLocal:
		if backend.Config.Local != nil {
			return filepath.Clean(backend.Config.Local.RootPath), nil
		}
	case BackendTypeNFS:
		if backend.Config.NFS != nil {
			return filepath.Clean(backend.Config.NFS.MountPath), nil
		}
	case BackendTypeSMB:
		if backend.Config.SMB != nil {
			return filepath.Clean(backend.Config.SMB.MountPath), nil
		}
	case BackendTypeDistributed:
		if backend.Config.Distributed != nil && backend.Config.Distributed.Adapter == "mounted-filesystem" {
			return filepath.Clean(backend.Config.Distributed.MountPath), nil
		}
	}
	return "", fmt.Errorf("%w: backend type %q does not expose a mounted filesystem", ErrProbeUnsupported, backend.Type)
}

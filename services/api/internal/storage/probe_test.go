package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
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

// requireChmodEnforced skips permission-bit-dependent tests where they can't
// be trusted: root bypasses POSIX permission checks entirely (common for
// containerized CI run as root), and Windows' read-only attribute on a
// directory does not block file creation inside it the way POSIX write
// permission does.
func requireChmodEnforced(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("directory permission bits are not POSIX write-protection on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses POSIX permission checks")
	}
}

// TestFilesystemProberReadOnlyLocalSkipsWriteProbe is the regression test for
// the 2026-08-11-adjacent design defect: a local backend's health probe used
// to always write+read+delete a probe file, so a read-only mount (the common
// case for a music server — e.g. `/media/music:ro`) was permanently reported
// unhealthy even though it was perfectly reachable. The root is chmod'd
// genuinely read-only (0555) so this test would fail if Probe() ever
// attempted a write, not just because ReadOnly was set.
func TestFilesystemProberReadOnlyLocalSkipsWriteProbe(t *testing.T) {
	requireChmodEnforced(t)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "existing-track.flac"), []byte("audio"), 0o444); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatalf("chmod root read-only: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) }) // let t.TempDir() clean up afterward

	backend := StorageBackend{
		ID:     "local-readonly",
		Type:   BackendTypeLocal,
		Config: BackendConfig{Local: &LocalConfig{RootPath: root, ReadOnly: true}},
	}

	if err := NewFilesystemProber().Probe(context.Background(), backend); err != nil {
		t.Fatalf("Probe() error = %v, want nil (a read-only mount must not be reported unhealthy)", err)
	}
}

// TestFilesystemProberReadOnlyLocalAcceptsEmptyDirectory ensures a freshly
// mounted, genuinely empty read-only library is healthy — Readdirnames
// returning io.EOF on an empty directory must not be mistaken for a failure.
func TestFilesystemProberReadOnlyLocalAcceptsEmptyDirectory(t *testing.T) {
	root := t.TempDir() // empty, still writable — only emptiness is under test here
	backend := StorageBackend{
		ID:     "local-readonly-empty",
		Type:   BackendTypeLocal,
		Config: BackendConfig{Local: &LocalConfig{RootPath: root, ReadOnly: true}},
	}

	if err := NewFilesystemProber().Probe(context.Background(), backend); err != nil {
		t.Fatalf("Probe() error = %v, want nil (an empty read-only mount is still healthy)", err)
	}
}

// TestFilesystemProberReadOnlyLocalRejectsMissingRoot confirms the read-only
// path still catches a genuinely unreachable mount — this isn't a blanket
// "always healthy" bypass, only a different (read-only) way of checking.
func TestFilesystemProberReadOnlyLocalRejectsMissingRoot(t *testing.T) {
	backend := StorageBackend{
		ID:     "local-readonly-missing",
		Type:   BackendTypeLocal,
		Config: BackendConfig{Local: &LocalConfig{RootPath: filepath.Join(t.TempDir(), "missing"), ReadOnly: true}},
	}

	err := NewFilesystemProber().Probe(context.Background(), backend)
	if !errors.Is(err, ErrProbeFailed) {
		t.Fatalf("Probe() error = %v, want ErrProbeFailed", err)
	}
}

// TestFilesystemProberWritableLocalStillFailsOnReadOnlyMount is the
// differential half of the regression test: on the exact same genuinely
// read-only directory that TestFilesystemProberReadOnlyLocalSkipsWriteProbe
// proves is healthy when ReadOnly is set, a backend that is NOT marked
// ReadOnly (the zero value — every backend registered before this change)
// must still take the write-based probe and correctly report unhealthy. This
// is what "keep the original semantics for writable backends unchanged"
// means in practice: misconfigured/broken writable backends must keep
// failing loudly, not be silently downgraded to healthy.
func TestFilesystemProberWritableLocalStillFailsOnReadOnlyMount(t *testing.T) {
	requireChmodEnforced(t)
	root := t.TempDir()
	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatalf("chmod root read-only: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })

	backend := StorageBackend{
		ID:     "local-should-be-writable",
		Type:   BackendTypeLocal,
		Config: BackendConfig{Local: &LocalConfig{RootPath: root}}, // ReadOnly not set — defaults to false
	}

	err := NewFilesystemProber().Probe(context.Background(), backend)
	if !errors.Is(err, ErrProbeFailed) {
		t.Fatalf("Probe() error = %v, want ErrProbeFailed (a backend not marked ReadOnly must still be probed for write access)", err)
	}
}

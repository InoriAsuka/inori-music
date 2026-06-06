package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrMediaObjectVerificationUnsupported = errors.New("media object verification is unsupported")
	ErrMediaObjectVerificationFailed      = errors.New("media object verification failed")
)

// MediaObjectVerificationReport records batch read-only integrity verification outcomes.
type MediaObjectVerificationReport struct {
	CheckedAt time.Time                       `json:"checkedAt"`
	Results   []MediaObjectVerificationResult `json:"results"`
}

// MediaObjectVerificationResult records a read-only integrity verification outcome.
type MediaObjectVerificationResult struct {
	MediaObjectID string    `json:"mediaObjectId"`
	BackendID     string    `json:"backendId"`
	ObjectKey     string    `json:"objectKey"`
	VerifiedAt    time.Time `json:"verifiedAt"`
	SizeBytes     int64     `json:"sizeBytes"`
	ContentHash   string    `json:"contentHash"`
	Status        string    `json:"status"`
	Message       string    `json:"message,omitempty"`
}

// MediaObjectVerifier verifies object bytes without mutating storage.
type MediaObjectVerifier interface {
	Verify(ctx context.Context, backend StorageBackend, object MediaObject) (MediaObjectVerificationResult, error)
}

// FilesystemMediaObjectVerifier verifies filesystem-backed media objects by reading their bytes.
type FilesystemMediaObjectVerifier struct{}

func NewFilesystemMediaObjectVerifier() *FilesystemMediaObjectVerifier {
	return &FilesystemMediaObjectVerifier{}
}

func (verifier *FilesystemMediaObjectVerifier) Verify(ctx context.Context, backend StorageBackend, object MediaObject) (MediaObjectVerificationResult, error) {
	if err := ctx.Err(); err != nil {
		return MediaObjectVerificationResult{}, fmt.Errorf("%w: %v", ErrMediaObjectVerificationFailed, err)
	}
	root, err := filesystemProbeRoot(backend)
	if err != nil {
		return MediaObjectVerificationResult{}, fmt.Errorf("%w: %v", ErrMediaObjectVerificationUnsupported, err)
	}
	algorithm, expectedHash, ok := strings.Cut(object.ContentHash, ":")
	if !ok || !strings.EqualFold(algorithm, "sha256") {
		return MediaObjectVerificationResult{}, fmt.Errorf("%w: content hash algorithm %q", ErrMediaObjectVerificationUnsupported, algorithm)
	}
	objectPath, err := safeObjectPath(root, object.ObjectKey)
	if err != nil {
		return MediaObjectVerificationResult{}, err
	}
	info, err := os.Stat(objectPath)
	if err != nil {
		return MediaObjectVerificationResult{}, fmt.Errorf("%w: stat object: %v", ErrMediaObjectVerificationFailed, err)
	}
	if !info.Mode().IsRegular() {
		return MediaObjectVerificationResult{}, fmt.Errorf("%w: object is not a regular file", ErrMediaObjectVerificationFailed)
	}
	if info.Size() != object.SizeBytes {
		return MediaObjectVerificationResult{}, fmt.Errorf("%w: size mismatch: got %d want %d", ErrMediaObjectVerificationFailed, info.Size(), object.SizeBytes)
	}
	file, err := os.Open(objectPath)
	if err != nil {
		return MediaObjectVerificationResult{}, fmt.Errorf("%w: open object: %v", ErrMediaObjectVerificationFailed, err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return MediaObjectVerificationResult{}, fmt.Errorf("%w: hash object: %v", ErrMediaObjectVerificationFailed, err)
	}
	actualHash := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actualHash, expectedHash) {
		return MediaObjectVerificationResult{}, fmt.Errorf("%w: sha256 mismatch", ErrMediaObjectVerificationFailed)
	}
	return MediaObjectVerificationResult{
		MediaObjectID: object.ID,
		BackendID:     backend.ID,
		ObjectKey:     object.ObjectKey,
		SizeBytes:     info.Size(),
		ContentHash:   "sha256:" + actualHash,
		Status:        "verified",
	}, nil
}

func safeObjectPath(root string, objectKey string) (string, error) {
	cleanRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", fmt.Errorf("%w: resolve root: %v", ErrMediaObjectVerificationFailed, err)
	}
	candidate := filepath.Join(cleanRoot, filepath.FromSlash(objectKey))
	rel, err := filepath.Rel(cleanRoot, candidate)
	if err != nil {
		return "", fmt.Errorf("%w: resolve object path: %v", ErrMediaObjectVerificationFailed, err)
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", fmt.Errorf("%w: object key escapes backend root", ErrMediaObjectVerificationFailed)
	}
	return candidate, nil
}

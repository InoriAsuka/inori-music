// Package streamsign provides HMAC-SHA256 signed URLs with expiration for
// resources the API proxies bytes for on backends that cannot issue their
// own presigned URLs (local/NFS/SMB storage) — track audio streams
// (GET .../tracks/{id}/stream) and, as of v5.39.0, album artwork bytes
// (GET .../albums/{id}/artwork/file).
//
// The signed payload is an opaque resource ID chosen by the caller. It was
// originally always a track ID (hence the parameter name history); callers
// that sign more than one kind of resource with the same Signer should
// namespace their IDs (e.g. "album:<id>") so a signature minted for one
// resource kind can never be replayed against another, even if the raw IDs
// happened to collide as strings. Track streaming keeps signing bare track
// IDs unchanged — only new callers need to namespace.
package streamsign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

const defaultTTL = 15 * time.Minute

// Signer signs resource URLs with HMAC-SHA256.
type Signer struct {
	key []byte
	ttl time.Duration
}

// NewSigner creates a new Signer with the given secret key.
func NewSigner(key string) *Signer {
	return &Signer{
		key: []byte(key),
		ttl: defaultTTL,
	}
}

// Sign generates a signed URL query string for the given resource ID.
// Returns "exp=<unix>&sig=<base64url>".
func (s *Signer) Sign(resourceID string) string {
	exp := time.Now().Add(s.ttl).Unix()
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(resourceID))
	mac.Write([]byte(strconv.FormatInt(exp, 10)))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("exp=%d&sig=%s", exp, sig)
}

// Verify checks the signature and expiration for the given resource ID.
func (s *Signer) Verify(resourceID string, exp int64, sig string) error {
	if time.Now().Unix() > exp {
		return fmt.Errorf("stream URL expired")
	}
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(resourceID))
	mac.Write([]byte(strconv.FormatInt(exp, 10)))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

// Package streamsign provides HMAC-SHA256 signed stream URLs with expiration.
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

// Signer signs stream URLs with HMAC-SHA256.
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

// Sign generates a signed stream URL query string for the given track ID.
// Returns "exp=<unix>&sig=<base64url>".
func (s *Signer) Sign(trackID string) string {
	exp := time.Now().Add(s.ttl).Unix()
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(trackID))
	mac.Write([]byte(strconv.FormatInt(exp, 10)))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("exp=%d&sig=%s", exp, sig)
}

// Verify checks the signature and expiration for the given track ID.
func (s *Signer) Verify(trackID string, exp int64, sig string) error {
	if time.Now().Unix() > exp {
		return fmt.Errorf("stream URL expired")
	}
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(trackID))
	mac.Write([]byte(strconv.FormatInt(exp, 10)))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

const productionBcryptCost = 12

// bcryptCost returns the bcrypt work factor to use. Test binaries hash at
// bcrypt's minimum cost: dozens of httpapi/auth tests each create a user
// via CreateUser, and at cost 12 a single hash takes ~11s under `go test
// -race` (vs ~0.1s at MinCost) — enough call sites to blow past `go test`'s
// default 600s per-package timeout, which is exactly what happened to the
// race-test CI step. testing.Testing() (Go 1.21+) detects this without any
// env var or per-test plumbing.
func bcryptCost() int {
	if testing.Testing() {
		return bcrypt.MinCost
	}
	return productionBcryptCost
}

// HashPassword returns a bcrypt hash of the plaintext password.
func HashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost())
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword reports whether plaintext matches the stored bcrypt hash.
func CheckPassword(hash, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}

// GenerateToken returns a cryptographically random 32-byte hex token.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hex digest of the plaintext token.
// Only the hash is stored; the plaintext is returned to the caller once.
func HashToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

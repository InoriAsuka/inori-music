package streamsign

import (
	"fmt"
	"testing"
	"time"
)

func TestSigner(t *testing.T) {
	s := NewSigner("test-secret-key")

	t.Run("valid signature", func(t *testing.T) {
		trackID := "123"
		query := s.Sign(trackID)

		var exp int64
		var sig string
		if n, _ := fmt.Sscanf(query, "exp=%d&sig=%s", &exp, &sig); n != 2 {
			t.Fatalf("failed to parse query: %s", query)
		}

		if err := s.Verify(trackID, exp, sig); err != nil {
			t.Errorf("Verify() failed: %v", err)
		}
	})

	t.Run("expired signature", func(t *testing.T) {
		trackID := "123"
		exp := time.Now().Add(-1 * time.Hour).Unix()
		sig := "dummy"

		err := s.Verify(trackID, exp, sig)
		if err == nil {
			t.Error("expected error for expired signature")
		}
	})

	t.Run("tampered signature", func(t *testing.T) {
		trackID := "123"
		query := s.Sign(trackID)

		var exp int64
		var sig string
		fmt.Sscanf(query, "exp=%d&sig=%s", &exp, &sig)

		err := s.Verify("999", exp, sig)
		if err == nil {
			t.Error("expected error for tampered trackID")
		}
	})

	t.Run("mismatched key", func(t *testing.T) {
		s2 := NewSigner("key-a")
		trackID := "123"
		query := s2.Sign(trackID)

		var exp int64
		var sig string
		fmt.Sscanf(query, "exp=%d&sig=%s", &exp, &sig)

		s3 := NewSigner("key-b")
		err := s3.Verify(trackID, exp, sig)
		if err == nil {
			t.Error("expected error when key mismatch")
		}
	})
}

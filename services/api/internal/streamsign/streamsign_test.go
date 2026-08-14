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

	// v5.39.0: album artwork reuses this signer with a namespaced payload
	// ("album:<id>") instead of a bare ID. A signature minted for one
	// resource kind must not verify for the other, even when the raw IDs are
	// identical strings — otherwise a signed track-stream URL could be
	// replayed against the artwork-file endpoint for a same-named album.
	t.Run("namespaced resource ID does not collide with bare ID", func(t *testing.T) {
		const id = "shared-id-123"
		trackQuery := s.Sign(id)
		var exp int64
		var sig string
		fmt.Sscanf(trackQuery, "exp=%d&sig=%s", &exp, &sig)

		if err := s.Verify("album:"+id, exp, sig); err == nil {
			t.Error("expected error verifying a bare-ID signature against a namespaced resource ID")
		}

		albumQuery := s.Sign("album:" + id)
		var albumExp int64
		var albumSig string
		fmt.Sscanf(albumQuery, "exp=%d&sig=%s", &albumExp, &albumSig)

		if err := s.Verify(id, albumExp, albumSig); err == nil {
			t.Error("expected error verifying a namespaced-ID signature against the bare resource ID")
		}
		// The namespaced signature must still verify against its own ID.
		if err := s.Verify("album:"+id, albumExp, albumSig); err != nil {
			t.Errorf("Verify(album:%s) failed: %v", id, err)
		}
	})
}

package ratelimit

import (
	"testing"
)

func TestLimiter(t *testing.T) {
	l := NewLimiter()

	t.Run("initial state unlocked", func(t *testing.T) {
		if l.IsLocked("test-key") {
			t.Error("expected unlocked")
		}
	})

	t.Run("locked after 5 failures", func(t *testing.T) {
		key := "user1"
		for i := 0; i < 5; i++ {
			l.RecordFailure(key)
		}
		if !l.IsLocked(key) {
			t.Error("expected locked after 5 failures")
		}
	})

	t.Run("reset clears lock", func(t *testing.T) {
		key := "user2"
		for i := 0; i < 5; i++ {
			l.RecordFailure(key)
		}
		l.ResetFailures(key)
		if l.IsLocked(key) {
			t.Error("expected unlocked after reset")
		}
	})

	t.Run("exponential backoff", func(t *testing.T) {
		key := "user3"
		for i := 0; i < 6; i++ {
			l.RecordFailure(key)
		}
		if !l.IsLocked(key) {
			t.Error("expected locked")
		}
	})
}

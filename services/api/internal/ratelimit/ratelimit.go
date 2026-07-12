// Package ratelimit provides in-memory login rate limiting.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter tracks failed login attempts per key (IP or username).
type Limiter struct {
	mu      sync.RWMutex
	entries map[string]*entry
}

type entry struct {
	failures    int
	lockedUntil time.Time
}

// NewLimiter creates a new rate limiter.
func NewLimiter() *Limiter {
	return &Limiter{
		entries: make(map[string]*entry),
	}
}

// RecordFailure increments the failure count for the given key.
func (l *Limiter) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[key]
	if !ok {
		e = &entry{}
		l.entries[key] = e
	}
	e.failures++

	if e.failures >= 5 {
		backoff := time.Duration(1<<uint(e.failures-5)) * 60 * time.Second
		if backoff > 1*time.Hour {
			backoff = 1 * time.Hour
		}
		e.lockedUntil = time.Now().Add(backoff)
	}
}

// ResetFailures clears the failure count for the given key.
func (l *Limiter) ResetFailures(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

// IsLocked returns whether the key is currently locked.
func (l *Limiter) IsLocked(key string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	e, ok := l.entries[key]
	if !ok {
		return false
	}
	return time.Now().Before(e.lockedUntil)
}

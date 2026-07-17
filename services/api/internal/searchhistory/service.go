package searchhistory

import (
	"context"
	"strings"
	"time"
)

// Service coordinates per-user search history persistence rules: dedup, newest
// first, capped to MaxEntries. An empty write clears the history.
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// WithClock replaces the time source. Intended for tests.
func (s *Service) WithClock(fn func() time.Time) *Service {
	s.now = fn
	return s
}

// Get returns the user's recent queries, deduplicated and newest first, capped
// to MaxEntries.
func (s *Service) Get(ctx context.Context, userID string) ([]string, error) {
	entries, err := s.repo.List(ctx, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	queries := make([]string, 0, len(entries))
	for _, e := range entries {
		queries = append(queries, e.Query)
	}
	return queries, nil
}

// Put replaces the user's entire search history with queries (interpreted as
// newest first). Blank queries are dropped, duplicates are collapsed to their
// newest occurrence, and the result is capped to MaxEntries. An empty (or
// all-blank) list clears the history.
func (s *Service) Put(ctx context.Context, userID string, queries []string) error {
	userID = strings.TrimSpace(userID)

	now := s.now().UTC()
	seen := make(map[string]struct{}, len(queries))
	entries := make([]Entry, 0, len(queries))
	for i, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		if _, dup := seen[q]; dup {
			continue
		}
		seen[q] = struct{}{}
		// Assign strictly decreasing timestamps so the stored newest-first
		// order is preserved by any time-ordered backend.
		entries = append(entries, Entry{Query: q, SearchedAt: now.Add(-time.Duration(i) * time.Millisecond)})
		if len(entries) >= MaxEntries {
			break
		}
	}

	if len(entries) == 0 {
		return s.repo.Clear(ctx, userID)
	}
	return s.repo.Replace(ctx, userID, entries)
}

// Delete clears the user's search history.
func (s *Service) Delete(ctx context.Context, userID string) error {
	return s.repo.Clear(ctx, strings.TrimSpace(userID))
}

package httpapi

import (
	"sync"
	"time"

	"inori-music/services/api/internal/artwork"
)

// artworkCacheTTL bounds how long a resolved (or confirmed-absent) album
// cover stays in the in-process cache before the next request re-opens and
// re-parses the source file. Resolving cover art means opening an audio
// file, parsing its tags, and possibly listing its directory — cheap once,
// wasteful to repeat on every album-grid render. An on-disk cache is not an
// option here: the media mount this reads from may be read-only
// (storage.LocalConfig.ReadOnly, added v5.35.0 for exactly this reason), so
// caching has to live in process memory and simply be paid again after a
// restart.
const artworkCacheTTL = 10 * time.Minute

// artworkCacheEntry holds one album's resolved cover — or a cached "no
// artwork found" result, so albums confirmed to have none don't pay the
// lookup cost on every request either.
type artworkCacheEntry struct {
	image     artwork.Image
	etag      string
	found     bool
	expiresAt time.Time
}

// artworkCache is a small in-process, TTL-based cache from album ID to its
// resolved cover image. Safe for concurrent use.
type artworkCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]artworkCacheEntry
}

func newArtworkCache(ttl time.Duration) *artworkCache {
	return &artworkCache{ttl: ttl, entries: make(map[string]artworkCacheEntry)}
}

// get returns the cached entry for albumID, or ok=false if there is none or
// it has expired.
func (c *artworkCache) get(albumID string) (artworkCacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[albumID]
	if !ok || time.Now().After(entry.expiresAt) {
		return artworkCacheEntry{}, false
	}
	return entry, true
}

// set stores (or overwrites) the resolution result for albumID.
func (c *artworkCache) set(albumID string, image artwork.Image, etag string, found bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[albumID] = artworkCacheEntry{
		image:     image,
		etag:      etag,
		found:     found,
		expiresAt: time.Now().Add(c.ttl),
	}
}

package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/meilisearch/meilisearch-go"
)

// MeilisearchService implements Service using a Meilisearch backend.
type MeilisearchService struct {
	client meilisearch.ServiceManager
}

// NewMeilisearch connects to Meilisearch and initialises indexes.
// Returns nil (not an error) if Meilisearch is unreachable — caller uses PG fallback.
func NewMeilisearch(host, apiKey string) (*MeilisearchService, error) {
	client := meilisearch.New(host, meilisearch.WithAPIKey(apiKey))
	// Health check
	if !client.IsHealthy() {
		log.Printf("search: meilisearch not reachable at %s — using PG fallback", host)
		return nil, nil
	}
	svc := &MeilisearchService{client: client}
	svc.ensureIndexes()
	return svc, nil
}

func (s *MeilisearchService) ensureIndexes() {
	for _, idx := range []string{"tracks", "albums", "artists"} {
		if _, err := s.client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        idx,
			PrimaryKey: "id",
		}); err != nil {
			// Index may already exist — ignore
		}
	}
}

func (s *MeilisearchService) Search(ctx context.Context, q string, limit int) (SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	var (
		mu       sync.Mutex
		out      SearchResult
		wg       sync.WaitGroup
		firstErr error
	)
	searchIdx := func(idxName string, collector func([]string)) {
		defer wg.Done()
		resp, err := s.client.Index(idxName).Search(q, &meilisearch.SearchRequest{
			Limit: int64(limit),
		})
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = fmt.Errorf("meilisearch %s search: %w", idxName, err)
			}
			mu.Unlock()
			return
		}
		ids := make([]string, 0, len(resp.Hits))
		for _, hit := range resp.Hits {
			if raw, ok := hit["id"]; ok {
				var id string
				if jsonErr := json.Unmarshal(raw, &id); jsonErr == nil && id != "" {
					ids = append(ids, id)
				}
			}
		}
		mu.Lock()
		collector(ids)
		mu.Unlock()
	}

	wg.Add(3)
	go searchIdx("tracks", func(ids []string) { out.Tracks = ids })
	go searchIdx("albums", func(ids []string) { out.Albums = ids })
	go searchIdx("artists", func(ids []string) { out.Artists = ids })
	wg.Wait()

	if firstErr != nil && len(out.Tracks)+len(out.Albums)+len(out.Artists) == 0 {
		return SearchResult{}, firstErr
	}
	return out, nil
}

func (s *MeilisearchService) IndexTrack(_ context.Context, trackID, title, artistName, genre string) error {
	_, err := s.client.Index("tracks").AddDocuments([]map[string]interface{}{
		{"id": trackID, "title": title, "artistName": artistName, "genre": genre},
	}, nil)
	return err
}

func (s *MeilisearchService) IndexAlbum(_ context.Context, albumID, title, artistName string) error {
	_, err := s.client.Index("albums").AddDocuments([]map[string]interface{}{
		{"id": albumID, "title": title, "artistName": artistName},
	}, nil)
	return err
}

func (s *MeilisearchService) IndexArtist(_ context.Context, artistID, name string) error {
	_, err := s.client.Index("artists").AddDocuments([]map[string]interface{}{
		{"id": artistID, "name": name},
	}, nil)
	return err
}

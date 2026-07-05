package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/meilisearch/meilisearch-go"
	"github.com/mozillazg/go-pinyin"
)

// MeilisearchService implements Service using a Meilisearch backend.
type MeilisearchService struct {
	client meilisearch.ServiceManager
}

const (
	highlightPreTag  = "<mark>"
	highlightPostTag = "</mark>"
)

var pinyinArgs = pinyin.NewArgs()

// toPinyin converts Chinese-character text to space-separated lazy pinyin
// (no tone marks), enabling pinyin-based search matches. Non-Chinese input
// (e.g. already-Latin titles) yields an empty string since Pinyin() skips
// characters it can't convert, and searching the empty attribute is harmless.
func toPinyin(s string) string {
	if s == "" {
		return ""
	}
	return strings.Join(pinyin.LazyPinyin(s, pinyinArgs), " ")
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

// searchableAttrsByIndex lists the fields eligible for text matching per index,
// including the pinyin-transliterated shadow fields used for pinyin search.
var searchableAttrsByIndex = map[string][]string{
	"tracks":  {"title", "titlePinyin", "artistName", "genre"},
	"albums":  {"title", "titlePinyin", "artistName"},
	"artists": {"name", "namePinyin"},
}

func (s *MeilisearchService) ensureIndexes() {
	for _, idx := range []string{"tracks", "albums", "artists"} {
		if _, err := s.client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        idx,
			PrimaryKey: "id",
		}); err != nil {
			// Index may already exist — ignore
		}
		attrs := searchableAttrsByIndex[idx]
		if _, err := s.client.Index(idx).UpdateSearchableAttributes(&attrs); err != nil {
			log.Printf("search: update searchable attributes for %s: %v", idx, err)
		}
	}
}

// ClearIndexes deletes all documents from every index without dropping the
// index itself (settings such as searchable attributes are preserved).
// Used by cmd/reindex to start a full rebuild from a clean slate so entities
// deleted from the catalog since the last index build don't linger as ghosts.
func (s *MeilisearchService) ClearIndexes(ctx context.Context) error {
	for _, idx := range []string{"tracks", "albums", "artists"} {
		if _, err := s.client.Index(idx).DeleteAllDocumentsWithContext(ctx, nil); err != nil {
			return fmt.Errorf("clear index %s: %w", idx, err)
		}
	}
	return nil
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
	out.Highlights = make(map[string]string)
	searchIdx := func(idxName, highlightAttr string, collector func([]string)) {
		defer wg.Done()
		resp, err := s.client.Index(idxName).Search(q, &meilisearch.SearchRequest{
			Limit:                 int64(limit),
			AttributesToHighlight: []string{highlightAttr},
			HighlightPreTag:       highlightPreTag,
			HighlightPostTag:      highlightPostTag,
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
		highlights := make(map[string]string, len(resp.Hits))
		for _, hit := range resp.Hits {
			id, snippet := parseSearchHit(hit, highlightAttr)
			if id == "" {
				continue
			}
			ids = append(ids, id)
			if snippet != "" {
				highlights[id] = snippet
			}
		}
		mu.Lock()
		collector(ids)
		for id, snippet := range highlights {
			out.Highlights[id] = snippet
		}
		mu.Unlock()
	}

	wg.Add(3)
	go searchIdx("tracks", "title", func(ids []string) { out.Tracks = ids })
	go searchIdx("albums", "title", func(ids []string) { out.Albums = ids })
	go searchIdx("artists", "name", func(ids []string) { out.Artists = ids })
	wg.Wait()

	if firstErr != nil && len(out.Tracks)+len(out.Albums)+len(out.Artists) == 0 {
		return SearchResult{}, firstErr
	}
	return out, nil
}

// parseSearchHit extracts the id and highlight snippet from one Meilisearch hit.
// It returns id="" when the hit has no usable id, and snippet="" when no highlight
// is present or the snippet doesn't contain the highlight sentinel tag.
func parseSearchHit(hit map[string]json.RawMessage, highlightAttr string) (id string, snippet string) {
	raw, ok := hit["id"]
	if !ok {
		return "", ""
	}
	if jsonErr := json.Unmarshal(raw, &id); jsonErr != nil || id == "" {
		return "", ""
	}
	formattedRaw, ok := hit["_formatted"]
	if !ok {
		return id, ""
	}
	var formatted map[string]json.RawMessage
	if jsonErr := json.Unmarshal(formattedRaw, &formatted); jsonErr != nil {
		return id, ""
	}
	snippetRaw, ok := formatted[highlightAttr]
	if !ok {
		return id, ""
	}
	if jsonErr := json.Unmarshal(snippetRaw, &snippet); jsonErr != nil || !strings.Contains(snippet, highlightPreTag) {
		return id, ""
	}
	return id, snippet
}

func (s *MeilisearchService) IndexTrack(_ context.Context, trackID, title, artistName, genre string) error {
	_, err := s.client.Index("tracks").AddDocuments([]map[string]interface{}{
		{"id": trackID, "title": title, "titlePinyin": toPinyin(title), "artistName": artistName, "genre": genre},
	}, nil)
	return err
}

func (s *MeilisearchService) IndexAlbum(_ context.Context, albumID, title, artistName string) error {
	_, err := s.client.Index("albums").AddDocuments([]map[string]interface{}{
		{"id": albumID, "title": title, "titlePinyin": toPinyin(title), "artistName": artistName},
	}, nil)
	return err
}

func (s *MeilisearchService) IndexArtist(_ context.Context, artistID, name string) error {
	_, err := s.client.Index("artists").AddDocuments([]map[string]interface{}{
		{"id": artistID, "name": name, "namePinyin": toPinyin(name)},
	}, nil)
	return err
}

package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"inori-music/services/api/internal/catalog"
	"inori-music/services/api/internal/storage"
)

// ---- catalog helpers ----

func (handler *Handler) requireCatalogService(w http.ResponseWriter) bool {
	if handler.catalogService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "catalog_not_configured", "catalog service is not configured")
		return false
	}
	return true
}

// ---- import handler ----

type importTrackRequest struct {
	MediaObjectID string `json:"mediaObjectId"`
	Title         string `json:"title"`
	SortTitle     string `json:"sortTitle"`
	ArtistID      string `json:"artistId"`
	AlbumID       string `json:"albumId"`
	TrackNumber   int    `json:"trackNumber"`
	DiscNumber    int    `json:"discNumber"`
	DurationMS    int    `json:"durationMs"`
	Genre         string `json:"genre"`
}

func (handler *Handler) importTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req importTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	track, err := handler.catalogService.ImportTrack(r.Context(), catalog.ImportTrackRequest{
		MediaObjectID: req.MediaObjectID,
		Title:         req.Title,
		SortTitle:     req.SortTitle,
		ArtistID:      req.ArtistID,
		AlbumID:       req.AlbumID,
		TrackNumber:   req.TrackNumber,
		DiscNumber:    req.DiscNumber,
		DurationMS:    req.DurationMS,
		Genre:         req.Genre,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, track)
}

// ---- batch import handler ----

type batchImportRequest struct {
	Items []importTrackRequest `json:"items"`
}

func (handler *Handler) batchImportTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req batchImportRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	items := make([]catalog.ImportTrackRequest, len(req.Items))
	for i, it := range req.Items {
		items[i] = catalog.ImportTrackRequest{
			MediaObjectID: it.MediaObjectID,
			Title:         it.Title,
			SortTitle:     it.SortTitle,
			ArtistID:      it.ArtistID,
			AlbumID:       it.AlbumID,
			TrackNumber:   it.TrackNumber,
			DiscNumber:    it.DiscNumber,
			DurationMS:    it.DurationMS,
			Genre:         it.Genre,
		}
	}
	result := handler.catalogService.BatchImportTracks(r.Context(), items)
	status := http.StatusOK
	if result.Failed > 0 && result.Imported > 0 {
		status = http.StatusMultiStatus
	} else if result.Failed > 0 && result.Imported == 0 {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, result)
}

// ---- search handler ----

func (handler *Handler) searchCatalog(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeAPIError(w, http.StatusBadRequest, "missing_query", "query parameter 'q' is required")
		return
	}
	var result catalog.CatalogSearchResult
	if handler.searchSvc != nil {
		limit := 50
		sr, err := handler.searchSvc.Search(r.Context(), q, limit)
		if err != nil {
			writeError(w, err)
			return
		}
		result = catalog.CatalogSearchResult{Query: q}
		for _, id := range sr.Artists {
			if a, err := handler.catalogService.GetArtist(r.Context(), id); err == nil {
				result.Items = append(result.Items, catalog.SearchResultItem{Kind: catalog.SearchResultArtist, Artist: &a, Highlight: sr.Highlights[id]})
			}
		}
		for _, id := range sr.Albums {
			if a, err := handler.catalogService.GetAlbum(r.Context(), id); err == nil {
				result.Items = append(result.Items, catalog.SearchResultItem{Kind: catalog.SearchResultAlbum, Album: &a, Highlight: sr.Highlights[id]})
			}
		}
		for _, id := range sr.Tracks {
			if t, err := handler.catalogService.GetTrack(r.Context(), id); err == nil {
				result.Items = append(result.Items, catalog.SearchResultItem{Kind: catalog.SearchResultTrack, Track: &t, Highlight: sr.Highlights[id]})
			}
		}
	} else {
		var err error
		result, err = handler.catalogService.SearchCatalog(r.Context(), q)
		if err != nil {
			writeError(w, err)
			return
		}
	}
	// Optional ?types= filter: comma-separated subset of "artist", "album", "track".
	// When absent or empty the full result is returned unchanged.
	if rawTypes := strings.TrimSpace(r.URL.Query().Get("types")); rawTypes != "" {
		allowed := make(map[string]bool)
		for _, t := range strings.Split(rawTypes, ",") {
			switch strings.TrimSpace(t) {
			case "artist", "album", "track":
				allowed[strings.TrimSpace(t)] = true
			default:
				writeAPIError(w, http.StatusBadRequest, "validation_error",
					"types must be a comma-separated list of: artist, album, track")
				return
			}
		}
		filtered := make([]catalog.SearchResultItem, 0, len(result.Items))
		for _, item := range result.Items {
			switch {
			case item.Kind == catalog.SearchResultArtist && allowed["artist"]:
				filtered = append(filtered, item)
			case item.Kind == catalog.SearchResultAlbum && allowed["album"]:
				filtered = append(filtered, item)
			case item.Kind == catalog.SearchResultTrack && allowed["track"]:
				filtered = append(filtered, item)
			}
		}
		result.Items = filtered
	}
	writeJSON(w, http.StatusOK, result)
}

// ---- catalog pagination & sort helpers ----

const (
	catalogListDefaultLimit = 50
	catalogListMaxLimit     = 500
)

// parseCatalogPage parses limit, offset, sortBy, and sortOrder query parameters.
// limit defaults to catalogListDefaultLimit and is clamped to catalogListMaxLimit.
// sortBy and sortOrder are returned as trimmed lowercase strings; empty strings
// signal "use entity default". Returns false and writes an error when limit or
// offset are invalid.
func parseCatalogPage(w http.ResponseWriter, r *http.Request) (limit, offset int, sortBy, sortOrder string, ok bool) {
	q := r.URL.Query()
	limit = catalogListDefaultLimit
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return 0, 0, "", "", false
		}
		if v > catalogListMaxLimit {
			v = catalogListMaxLimit
		}
		limit = v
	}
	if raw := q.Get("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_offset", "offset must be a non-negative integer")
			return 0, 0, "", "", false
		}
		offset = v
	}
	sortBy = strings.ToLower(strings.TrimSpace(q.Get("sortBy")))
	sortOrder = strings.ToLower(strings.TrimSpace(q.Get("sortOrder")))
	return limit, offset, sortBy, sortOrder, true
}

// parseReleaseYearRange parses optional ?releaseYearMin and ?releaseYearMax query params.
// Returns 0,0,true when both are absent. Writes a 400 error and returns _,_,false on invalid input.
func parseReleaseYearRange(w http.ResponseWriter, r *http.Request) (min, max int, ok bool) {
	q := r.URL.Query()
	if raw := q.Get("releaseYearMin"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeAPIError(w, http.StatusBadRequest, "validation_error", "releaseYearMin must be a non-negative integer")
			return 0, 0, false
		}
		min = v
	}
	if raw := q.Get("releaseYearMax"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeAPIError(w, http.StatusBadRequest, "validation_error", "releaseYearMax must be a non-negative integer")
			return 0, 0, false
		}
		max = v
	}
	if min > 0 && max > 0 && min > max {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "releaseYearMin must be <= releaseYearMax")
		return 0, 0, false
	}
	return min, max, true
}

// normalizeSortOrder returns "asc" or "desc". Empty input → "asc".
// Returns "", false for any other value.
func normalizeSortOrder(raw string) (string, bool) {
	switch raw {
	case "", catalog.CatalogSortOrderAsc:
		return catalog.CatalogSortOrderAsc, true
	case catalog.CatalogSortOrderDesc:
		return catalog.CatalogSortOrderDesc, true
	}
	return "", false
}

// sortCatalogArtists sorts artists in-place by sortBy/sortOrder.
// Invalid sortBy is treated as "name" (default).
func sortCatalogArtists(artists []catalog.Artist, sortBy, sortOrder string) {
	desc := sortOrder == catalog.CatalogSortOrderDesc
	sort.SliceStable(artists, func(i, j int) bool {
		a, b := artists[i], artists[j]
		var less bool
		switch sortBy {
		case catalog.ArtistSortBySortName:
			less = a.SortName < b.SortName
		case catalog.ArtistSortByCreatedAt:
			less = a.CreatedAt.Before(b.CreatedAt)
		case catalog.ArtistSortByUpdatedAt:
			less = a.UpdatedAt.Before(b.UpdatedAt)
		default: // "name"
			less = a.Name < b.Name
		}
		if desc {
			return !less
		}
		return less
	})
}

// sortCatalogAlbums sorts albums in-place by sortBy/sortOrder.
func sortCatalogAlbums(albums []catalog.Album, sortBy, sortOrder string) {
	desc := sortOrder == catalog.CatalogSortOrderDesc
	sort.SliceStable(albums, func(i, j int) bool {
		a, b := albums[i], albums[j]
		var less bool
		switch sortBy {
		case catalog.AlbumSortBySortTitle:
			less = a.SortTitle < b.SortTitle
		case catalog.AlbumSortByReleaseYear:
			less = a.ReleaseYear < b.ReleaseYear
		case catalog.AlbumSortByCreatedAt:
			less = a.CreatedAt.Before(b.CreatedAt)
		case catalog.AlbumSortByUpdatedAt:
			less = a.UpdatedAt.Before(b.UpdatedAt)
		default: // "title"
			less = a.Title < b.Title
		}
		if desc {
			return !less
		}
		return less
	})
}

// sortCatalogTracks sorts tracks in-place by sortBy/sortOrder.
func sortCatalogTracks(tracks []catalog.Track, sortBy, sortOrder string) {
	desc := sortOrder == catalog.CatalogSortOrderDesc
	sort.SliceStable(tracks, func(i, j int) bool {
		a, b := tracks[i], tracks[j]
		var less bool
		switch sortBy {
		case catalog.TrackSortBySortTitle:
			less = a.SortTitle < b.SortTitle
		case catalog.TrackSortByTrackNumber:
			less = a.TrackNumber < b.TrackNumber
		case catalog.TrackSortByDiscNumber:
			less = a.DiscNumber < b.DiscNumber
		case catalog.TrackSortByDurationMS:
			less = a.DurationMS < b.DurationMS
		case catalog.TrackSortByCreatedAt:
			less = a.CreatedAt.Before(b.CreatedAt)
		case catalog.TrackSortByUpdatedAt:
			less = a.UpdatedAt.Before(b.UpdatedAt)
		default: // "title"
			less = a.Title < b.Title
		}
		if desc {
			return !less
		}
		return less
	})
}

// sortCatalogPlaylists sorts playlists in-place by sortBy/sortOrder.
func sortCatalogPlaylists(playlists []catalog.Playlist, sortBy, sortOrder string) {
	desc := sortOrder == catalog.CatalogSortOrderDesc
	sort.SliceStable(playlists, func(i, j int) bool {
		a, b := playlists[i], playlists[j]
		var less bool
		switch sortBy {
		case catalog.PlaylistSortByCreatedAt:
			less = a.CreatedAt.Before(b.CreatedAt)
		case catalog.PlaylistSortByUpdatedAt:
			less = a.UpdatedAt.Before(b.UpdatedAt)
		default: // "name"
			less = a.Name < b.Name
		}
		if desc {
			return !less
		}
		return less
	})
}

// paginateCatalog slices items[offset:offset+limit] and returns the page meta.
func paginateCatalog[T any](items []T, limit, offset int) ([]T, catalog.CatalogPaginationMeta) {
	total := len(items)
	if offset >= total {
		return []T{}, catalog.CatalogPaginationMeta{Limit: limit, Offset: offset, Total: total, HasMore: false}
	}
	end := offset + limit
	if end > total {
		end = total
	}
	page := items[offset:end]
	return page, catalog.CatalogPaginationMeta{
		Limit:   limit,
		Offset:  offset,
		Total:   total,
		HasMore: end < total,
	}
}

// ---- artist handlers ----

type createArtistRequest struct {
	Name     string `json:"name"`
	SortName string `json:"sortName"`
}

func (handler *Handler) listArtists(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	page, err := handler.catalogService.ListArtistsPage(r.Context(), catalog.ListQuery{
		SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: page.Total, HasMore: offset+limit < page.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"artists": page.Items, "pagination": meta})
}

func (handler *Handler) createArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req createArtistRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	artist, err := handler.catalogService.CreateArtist(r.Context(), req.Name, req.SortName)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, artist)
}

func (handler *Handler) getArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	artist, err := handler.catalogService.GetArtist(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, artist)
}

func (handler *Handler) deleteArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	if err := handler.catalogService.DeleteArtist(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listAlbumsByArtist returns paged, sorted albums belonging to the artist identified
// by the {id} path parameter. The parent artist must exist; unknown IDs return 404.
func (handler *Handler) listAlbumsByArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	// Verify the artist exists before listing — produces 404 on unknown IDs.
	if _, err := handler.catalogService.GetArtist(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	ryMin, ryMax, ok2 := parseReleaseYearRange(w, r)
	if !ok2 {
		return
	}
	albumPage, err := handler.catalogService.ListAlbumsByArtistPage(r.Context(), r.PathValue("id"), catalog.ListQuery{
		SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset,
		ReleaseYearMin: ryMin, ReleaseYearMax: ryMax,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: albumPage.Total, HasMore: offset+limit < albumPage.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"albums": albumPage.Items, "pagination": meta})
}

// listTracksByArtist returns paged, sorted tracks belonging to the artist identified
// by the {id} path parameter. The parent artist must exist; unknown IDs return 404.
func (handler *Handler) listTracksByArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	if _, err := handler.catalogService.GetArtist(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	trackPage, err := handler.catalogService.ListTracksByArtistPage(r.Context(), r.PathValue("id"), catalog.ListQuery{
		SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: trackPage.Total, HasMore: offset+limit < trackPage.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": trackPage.Items, "pagination": meta})
}

// patchArtistRequest carries the fields that may be changed via PATCH.
// Pointer semantics: nil = leave unchanged, pointer-to-string = set new value.
type patchArtistRequest struct {
	Name     *string `json:"name"`
	SortName *string `json:"sortName"`
}

func (handler *Handler) patchArtist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req patchArtistRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	artist, err := handler.catalogService.UpdateArtist(r.Context(), r.PathValue("id"), catalog.UpdateArtistRequest{
		Name:     req.Name,
		SortName: req.SortName,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, artist)
}

// ---- album handlers ----

type createAlbumRequest struct {
	Title       string `json:"title"`
	SortTitle   string `json:"sortTitle"`
	ArtistID    string `json:"artistId"`
	ReleaseYear int    `json:"releaseYear"`
}

func (handler *Handler) listAlbums(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	q := catalog.ListQuery{SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset}
	artistID := r.URL.Query().Get("artistId")
	if ryMin, ryMax, ok := parseReleaseYearRange(w, r); !ok {
		return
	} else {
		q.ReleaseYearMin = ryMin
		q.ReleaseYearMax = ryMax
	}
	var (
		page catalog.ListPage[catalog.Album]
		err  error
	)
	if artistID != "" {
		page, err = handler.catalogService.ListAlbumsByArtistPage(r.Context(), artistID, q)
	} else {
		page, err = handler.catalogService.ListAlbumsPage(r.Context(), q)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: page.Total, HasMore: offset+limit < page.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"albums": page.Items, "pagination": meta})
}

func (handler *Handler) createAlbum(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req createAlbumRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	album, err := handler.catalogService.CreateAlbum(r.Context(), req.Title, req.SortTitle, req.ArtistID, req.ReleaseYear)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, album)
}

func (handler *Handler) getAlbum(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	album, err := handler.catalogService.GetAlbum(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, album)
}

func (handler *Handler) deleteAlbum(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	if err := handler.catalogService.DeleteAlbum(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listTracksByAlbum returns paged, sorted tracks belonging to the album identified
// by the {id} path parameter. The parent album must exist; unknown IDs return 404.
func (handler *Handler) listTracksByAlbum(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	if _, err := handler.catalogService.GetAlbum(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	trackPage, err := handler.catalogService.ListTracksByAlbumPage(r.Context(), r.PathValue("id"), catalog.ListQuery{
		SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: trackPage.Total, HasMore: offset+limit < trackPage.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": trackPage.Items, "pagination": meta})
}

// patchAlbumRequest carries the fields that may be changed via PATCH.
type patchAlbumRequest struct {
	Title                *string `json:"title"`
	SortTitle            *string `json:"sortTitle"`
	ArtistID             *string `json:"artistId"`
	ReleaseYear          *int    `json:"releaseYear"`
	ArtworkMediaObjectID *string `json:"artworkMediaObjectId"`
}

func (handler *Handler) patchAlbum(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req patchAlbumRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	album, err := handler.catalogService.UpdateAlbum(r.Context(), r.PathValue("id"), catalog.UpdateAlbumRequest{
		Title:                req.Title,
		SortTitle:            req.SortTitle,
		ArtistID:             req.ArtistID,
		ReleaseYear:          req.ReleaseYear,
		ArtworkMediaObjectID: req.ArtworkMediaObjectID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, album)
}

type albumArtworkResponse struct {
	URL       string `json:"url"`
	ExpiresIn int    `json:"expiresIn"`
}

func (handler *Handler) getAlbumArtwork(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	album, err := handler.catalogService.GetAlbum(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if album.ArtworkMediaObjectID == "" {
		writeAPIError(w, http.StatusNotFound, "no_artwork", "no artwork")
		return
	}
	if handler.storage == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "storage_unavailable", "storage service not configured")
		return
	}
	const artworkTTL = 900 * time.Second
	mo, err := handler.mediaObjects.GetMediaObject(r.Context(), album.ArtworkMediaObjectID)
	if err != nil {
		// Propagate the error through the standard error writer so that a
		// missing media object returns 404 via the domain sentinel while a DB
		// or infrastructure failure returns the appropriate 5xx status.
		writeError(w, err)
		return
	}
	url, err := handler.storage.GeneratePresignedURL(r.Context(), mo.BackendID, mo.ObjectKey, artworkTTL)
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "presign_failed", "failed to generate artwork URL")
		return
	}
	// Derive ExpiresIn from the TTL constant so they never drift independently.
	writeJSON(w, http.StatusOK, albumArtworkResponse{URL: url, ExpiresIn: int(artworkTTL / time.Second)})
}

type lyricsResponse struct {
	Format                   string `json:"format"`
	Content                  string `json:"content"`
	Translation              string `json:"translation,omitempty"`
	Source                   string `json:"source,omitempty"`
	TranslationMediaObjectID string `json:"translationMediaObjectId,omitempty"`
}

func (handler *Handler) uploadTrackLyrics(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	if handler.storage == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "storage_unavailable", "storage service not configured")
		return
	}
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_form", "failed to parse multipart form")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "missing_file", "file field required")
		return
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, 512*1024))
	if err != nil {
		writeError(w, err)
		return
	}
	contentStr := string(content)
	format := detectLyricsFormat(contentStr)
	if format == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_format", "file must be LRC or SRT format")
		return
	}
	var translationContent []byte
	hasTranslation := false
	if tfile, _, terr := r.FormFile("translation"); terr == nil {
		hasTranslation = true
		defer tfile.Close()
		translationContent, err = io.ReadAll(io.LimitReader(tfile, 512*1024))
		if err != nil {
			writeError(w, err)
			return
		}
		if !utf8.Valid(translationContent) {
			writeAPIError(w, http.StatusBadRequest, "invalid_format", "translation must be UTF-8 text")
			return
		}
	} else if !errors.Is(terr, http.ErrMissingFile) {
		writeAPIError(w, http.StatusBadRequest, "invalid_form", "failed to read translation field")
		return
	}
	trackID := r.PathValue("id")
	if _, err := handler.catalogService.GetTrack(r.Context(), trackID); err != nil {
		writeError(w, err)
		return
	}
	// Find the default backend to store the lyrics object.
	backends, err := handler.storage.ListBackends(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	var defaultBackend *storage.StorageBackend
	for i := range backends {
		if backends[i].IsDefault && backends[i].Enabled {
			b := backends[i]
			defaultBackend = &b
			break
		}
	}
	if defaultBackend == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "no_default_backend", "no enabled default storage backend configured")
		return
	}
	objectKey := "lyrics/" + trackID + "." + format
	// Generate a content hash for deduplication (algorithm:value format).
	h := sha256.Sum256(content)
	contentHash := "sha256:" + fmt.Sprintf("%x", h)
	mo, err := handler.mediaObjects.RegisterMediaObject(r.Context(), storage.MediaObject{
		ID:             contentHash[:16],
		BackendID:      defaultBackend.ID,
		ObjectKey:      objectKey,
		ContentHash:    contentHash,
		SizeBytes:      int64(len(content)),
		MIMEType:       "text/plain; charset=utf-8",
		AssetKind:      string(storage.AssetKindLyrics),
		LifecycleState: string(storage.LifecycleStateActive),
	})
	if err != nil {
		// If already exists, fetch it to get its current ID.
		existing, getErr := handler.mediaObjects.GetMediaObject(r.Context(), contentHash[:16])
		if getErr != nil {
			writeError(w, err)
			return
		}
		mo = existing
	}
	if err := handler.storage.PutObject(r.Context(), mo.BackendID, mo.ObjectKey, bytes.NewReader(content), int64(len(content))); err != nil {
		writeError(w, err)
		return
	}
	moID := mo.ID
	source := "manual"
	updateReq := catalog.UpdateTrackRequest{
		LyricsMediaObjectID: &moID,
		LyricsSource:        &source,
	}
	var translationMOID string
	if hasTranslation {
		tObjectKey := "lyrics/" + trackID + ".translation." + format
		th := sha256.Sum256(translationContent)
		tContentHash := "sha256:" + fmt.Sprintf("%x", th)
		tmo, err := handler.mediaObjects.RegisterMediaObject(r.Context(), storage.MediaObject{
			ID:             tContentHash[:16],
			BackendID:      defaultBackend.ID,
			ObjectKey:      tObjectKey,
			ContentHash:    tContentHash,
			SizeBytes:      int64(len(translationContent)),
			MIMEType:       "text/plain; charset=utf-8",
			AssetKind:      string(storage.AssetKindLyrics),
			LifecycleState: string(storage.LifecycleStateActive),
		})
		if err != nil {
			existing, getErr := handler.mediaObjects.GetMediaObject(r.Context(), tContentHash[:16])
			if getErr != nil {
				writeError(w, err)
				return
			}
			tmo = existing
		}
		if err := handler.storage.PutObject(r.Context(), tmo.BackendID, tmo.ObjectKey, bytes.NewReader(translationContent), int64(len(translationContent))); err != nil {
			writeError(w, err)
			return
		}
		translationMOID = tmo.ID
		updateReq.LyricsTranslationMediaObjectID = &translationMOID
	}
	if _, err := handler.catalogService.UpdateTrack(r.Context(), trackID, updateReq); err != nil {
		writeError(w, err)
		return
	}
	resp := map[string]string{"mediaObjectId": mo.ID}
	if translationMOID != "" {
		resp["translationMediaObjectId"] = translationMOID
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (handler *Handler) getTrackLyrics(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	track, err := handler.catalogService.GetTrack(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if track.LyricsMediaObjectID == "" {
		writeAPIError(w, http.StatusNotFound, "no_lyrics", "no lyrics for this track")
		return
	}
	if handler.storage == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "storage_unavailable", "storage service not configured")
		return
	}
	mo, err := handler.mediaObjects.GetMediaObject(r.Context(), track.LyricsMediaObjectID)
	if err != nil {
		writeError(w, err)
		return
	}
	rc, err := handler.storage.GetObject(r.Context(), mo.BackendID, mo.ObjectKey)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rc.Close()
	content, err := io.ReadAll(io.LimitReader(rc, 512*1024))
	if err != nil {
		writeError(w, err)
		return
	}
	contentStr := string(content)
	format := detectLyricsFormat(contentStr)
	resp := lyricsResponse{Format: format, Content: contentStr, Source: track.LyricsSource}
	if track.LyricsTranslationMediaObjectID != "" {
		tmo, err := handler.mediaObjects.GetMediaObject(r.Context(), track.LyricsTranslationMediaObjectID)
		if err != nil {
			writeError(w, err)
			return
		}
		trc, err := handler.storage.GetObject(r.Context(), tmo.BackendID, tmo.ObjectKey)
		if err != nil {
			writeError(w, err)
			return
		}
		defer trc.Close()
		translationContent, err := io.ReadAll(io.LimitReader(trc, 512*1024))
		if err != nil {
			writeError(w, err)
			return
		}
		resp.Translation = string(translationContent)
		resp.TranslationMediaObjectID = tmo.ID
	}
	writeJSON(w, http.StatusOK, resp)
}

func (handler *Handler) deleteTrackLyrics(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	track, err := handler.catalogService.GetTrack(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if track.LyricsMediaObjectID == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	empty := ""
	if _, err := handler.catalogService.UpdateTrack(r.Context(), r.PathValue("id"), catalog.UpdateTrackRequest{
		LyricsMediaObjectID:            &empty,
		LyricsTranslationMediaObjectID: &empty,
		LyricsSource:                   &empty,
	}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// detectLyricsFormat returns "lrc", "srt", or "" if unrecognized.
func detectLyricsFormat(content string) string {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "[") {
		return "lrc"
	}
	// SRT starts with a digit (sequence number)
	for _, r := range trimmed {
		if r >= '0' && r <= '9' {
			return "srt"
		}
		break
	}
	return ""
}

// ---- track handlers ----

type createTrackRequest struct {
	Title         string `json:"title"`
	SortTitle     string `json:"sortTitle"`
	ArtistID      string `json:"artistId"`
	AlbumID       string `json:"albumId"`
	MediaObjectID string `json:"mediaObjectId"`
	TrackNumber   int    `json:"trackNumber"`
	DiscNumber    int    `json:"discNumber"`
	DurationMS    int    `json:"durationMs"`
	Genre         string `json:"genre"`
}

func (handler *Handler) listTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	q := catalog.ListQuery{SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset}
	queryArgs := r.URL.Query()
	artistID := queryArgs.Get("artistId")
	albumID := queryArgs.Get("albumId")
	genre := queryArgs.Get("genre")
	q.Genre = genre
	var (
		page catalog.ListPage[catalog.Track]
		err  error
	)
	switch {
	case albumID != "":
		page, err = handler.catalogService.ListTracksByAlbumPage(r.Context(), albumID, q)
	case artistID != "":
		page, err = handler.catalogService.ListTracksByArtistPage(r.Context(), artistID, q)
	default:
		page, err = handler.catalogService.ListTracksPage(r.Context(), q)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: page.Total, HasMore: offset+limit < page.Total,
	}
	if isViewerPath(r) {
		user, _ := userFromContext(r)
		views := handler.annotateTracksWithFavorites(r.Context(), user.ID, page.Items)
		writeJSON(w, http.StatusOK, map[string]any{"tracks": views, "pagination": meta})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": page.Items, "pagination": meta})
}

// trackView wraps a catalog.Track with an isFavorite annotation for viewer responses.
type trackView struct {
	catalog.Track
	IsFavorite bool `json:"isFavorite"`
}

// annotateTracksWithFavorites returns a slice of trackView with IsFavorite populated
// using a single batch lookup. Returns all-false when favoritesService is nil or the
// batch lookup fails (best-effort; does not fail the request).
func (handler *Handler) annotateTracksWithFavorites(ctx context.Context, userID string, tracks []catalog.Track) []trackView {
	views := make([]trackView, len(tracks))
	if len(tracks) == 0 {
		return views
	}
	ids := make([]string, len(tracks))
	for i, t := range tracks {
		ids[i] = t.ID
		views[i] = trackView{Track: t}
	}
	if handler.favoritesService != nil && userID != "" {
		favMap, err := handler.favoritesService.AreFavorites(ctx, userID, ids)
		if err == nil {
			for i, t := range tracks {
				views[i].IsFavorite = favMap[t.ID]
			}
		}
	}
	return views
}

// isViewerPath reports whether the request path is under /api/v1/catalog/ (viewer) vs /api/v1/admin/.
func isViewerPath(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/v1/catalog/") ||
		strings.HasPrefix(r.URL.Path, "/api/v1/me/")
}

func (handler *Handler) createTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req createTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	track, err := handler.catalogService.CreateTrack(r.Context(), req.Title, req.SortTitle, req.ArtistID, req.AlbumID, req.MediaObjectID, req.Genre, req.TrackNumber, req.DiscNumber, req.DurationMS)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, track)
}

func (handler *Handler) getTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	track, err := handler.catalogService.GetTrack(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if isViewerPath(r) {
		user, _ := userFromContext(r)
		views := handler.annotateTracksWithFavorites(r.Context(), user.ID, []catalog.Track{track})
		writeJSON(w, http.StatusOK, views[0])
		return
	}
	writeJSON(w, http.StatusOK, track)
}

// trackPlaybackDescriptor is the metadata-only response returned by the viewer
// playback endpoint. It carries the fields a client needs to fetch the audio
// file from its storage backend without the server streaming bytes.
// PresignedURL is populated when the backend supports presigned URLs and
// credentials are available; it is omitted otherwise.
type trackPlaybackDescriptor struct {
	TrackID       string `json:"trackId"`
	MediaObjectID string `json:"mediaObjectId"`
	MIMEType      string `json:"mimeType"`
	DurationMS    int    `json:"durationMs"`
	BackendID     string `json:"backendId"`
	BackendType   string `json:"backendType,omitempty"`
	ObjectKey     string `json:"objectKey"`
	PresignedURL  string `json:"presignedUrl,omitempty"`
	// StreamURL is the server-proxied streaming URL for backends that do not
	// support presigned URLs (local, NFS, SMB). Clients should prefer
	// PresignedURL when both are present.
	StreamURL string `json:"streamUrl,omitempty"`
}

func (handler *Handler) getTrackPlayback(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	track, err := handler.catalogService.GetTrack(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	mo, err := handler.mediaObjects.GetMediaObject(r.Context(), track.MediaObjectID)
	if err != nil {
		writeError(w, err)
		return
	}
	if mo.LifecycleState != string(storage.LifecycleStateActive) ||
		(mo.AssetKind != string(storage.AssetKindOriginalAudio) && mo.AssetKind != string(storage.AssetKindTranscodedAudio)) {
		writeError(w, fmt.Errorf("%w: media object %s is not in a playable state (lifecycleState=%s assetKind=%s)",
			storage.ErrPlaybackUnavailable, mo.ID, mo.LifecycleState, mo.AssetKind))
		return
	}
	backendType := ""
	presignedURL := ""
	streamURL := ""
	if handler.storage != nil {
		backend, backendErr := handler.storage.GetBackend(r.Context(), mo.BackendID)
		if backendErr == nil {
			backendType = string(backend.Type)
			if backend.Capabilities.PresignedURLs {
				if purl, pErr := handler.storage.GeneratePresignedURL(
					r.Context(), mo.BackendID, mo.ObjectKey, storage.DefaultPresignedURLTTL,
				); pErr == nil {
					presignedURL = purl
				}
			}
			// For filesystem-based backends that cannot presign, expose a
			// server-proxy stream URL so the web client can play via /stream.
			if presignedURL == "" {
				switch backend.Type {
				case storage.BackendTypeLocal, storage.BackendTypeNFS, storage.BackendTypeSMB:
					streamURL = "/api/v1/catalog/tracks/" + track.ID + "/stream"
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, trackPlaybackDescriptor{
		TrackID:       track.ID,
		MediaObjectID: mo.ID,
		MIMEType:      mo.MIMEType,
		DurationMS:    track.DurationMS,
		BackendID:     mo.BackendID,
		BackendType:   backendType,
		ObjectKey:     mo.ObjectKey,
		PresignedURL:  presignedURL,
		StreamURL:     streamURL,
	})
}

// streamTrack proxies audio bytes from a filesystem-based storage backend
// (local, NFS, SMB) directly to the client. It supports HTTP Range requests
// so browsers can seek within the audio file.
//
// Authentication: accepts a Bearer token in the Authorization header OR in the
// ?token= query parameter. The ?token= fallback is required because the HTML
// <audio> element cannot set custom request headers.
func (handler *Handler) streamTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}

	// Authenticate — try header first, fall back to ?token= query param.
	rawToken := r.Header.Get("Authorization")
	if rawToken == "" {
		if qt := r.URL.Query().Get("token"); qt != "" {
			rawToken = "Bearer " + qt
		}
	}
	token, ok := bearerToken(rawToken)
	if !ok {
		w.Header().Set("WWW-Authenticate", `Bearer realm="inori"`)
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}
	if handler.authService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "auth_not_configured", "authentication service is not configured")
		return
	}
	if _, err := handler.authService.ValidateToken(r.Context(), token); err != nil {
		w.Header().Set("WWW-Authenticate", `Bearer realm="inori"`)
		writeAPIError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token is required")
		return
	}

	track, err := handler.catalogService.GetTrack(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	mo, err := handler.mediaObjects.GetMediaObject(r.Context(), track.MediaObjectID)
	if err != nil {
		writeError(w, err)
		return
	}
	if mo.LifecycleState != string(storage.LifecycleStateActive) ||
		(mo.AssetKind != string(storage.AssetKindOriginalAudio) && mo.AssetKind != string(storage.AssetKindTranscodedAudio)) {
		writeError(w, fmt.Errorf("%w: media object %s is not in a playable state", storage.ErrPlaybackUnavailable, mo.ID))
		return
	}
	if handler.storage == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "storage_not_configured", "storage service is not configured")
		return
	}
	backend, err := handler.storage.GetBackend(r.Context(), mo.BackendID)
	if err != nil {
		writeError(w, err)
		return
	}

	// Resolve the local filesystem path for the object.
	var rootPath string
	switch backend.Type {
	case storage.BackendTypeLocal:
		if backend.Config.Local == nil {
			writeAPIError(w, http.StatusServiceUnavailable, "backend_config_missing", "local backend has no config")
			return
		}
		rootPath = backend.Config.Local.RootPath
	case storage.BackendTypeNFS:
		if backend.Config.NFS == nil {
			writeAPIError(w, http.StatusServiceUnavailable, "backend_config_missing", "NFS backend has no config")
			return
		}
		rootPath = backend.Config.NFS.MountPath
	case storage.BackendTypeSMB:
		if backend.Config.SMB == nil {
			writeAPIError(w, http.StatusServiceUnavailable, "backend_config_missing", "SMB backend has no config")
			return
		}
		rootPath = backend.Config.SMB.MountPath
	default:
		writeAPIError(w, http.StatusBadRequest, "stream_unsupported",
			fmt.Sprintf("streaming is not supported for backend type %s; use presigned URLs", backend.Type))
		return
	}

	filePath, err := storage.SafeObjectPath(rootPath, mo.ObjectKey)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_object_key", "object key resolves outside backend root")
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			writeAPIError(w, http.StatusNotFound, "file_not_found", "audio file not found on storage backend")
		} else {
			writeAPIError(w, http.StatusInternalServerError, "file_open_failed", "failed to open audio file")
		}
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "file_stat_failed", "failed to stat audio file")
		return
	}

	mimeType := mo.MIMEType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "no-store")
	// Allows the browser to seek by accepting Range requests.
	http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
}

func (handler *Handler) deleteTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	if err := handler.catalogService.DeleteTrack(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// relinkTrackRequest carries the new media object reference for a relink operation.
type relinkTrackRequest struct {
	MediaObjectID string `json:"mediaObjectId"`
}

func (handler *Handler) relinkTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req relinkTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	track, err := handler.catalogService.RelinkTrack(r.Context(), r.PathValue("id"), catalog.RelinkTrackRequest{
		MediaObjectID: req.MediaObjectID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, track)
}

// patchTrackRequest carries the fields that may be changed via PATCH.
type patchTrackRequest struct {
	Title        *string  `json:"title"`
	SortTitle    *string  `json:"sortTitle"`
	ArtistID     *string  `json:"artistId"`
	AlbumID      *string  `json:"albumId"`
	TrackNumber  *int     `json:"trackNumber"`
	DiscNumber   *int     `json:"discNumber"`
	DurationMS   *int     `json:"durationMs"`
	Genre        *string  `json:"genre"`
	ReplayGainDb *float64 `json:"replayGainDb"`
}

func (handler *Handler) patchTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req patchTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	track, err := handler.catalogService.UpdateTrack(r.Context(), r.PathValue("id"), catalog.UpdateTrackRequest{
		Title:        req.Title,
		SortTitle:    req.SortTitle,
		ArtistID:     req.ArtistID,
		AlbumID:      req.AlbumID,
		TrackNumber:  req.TrackNumber,
		DiscNumber:   req.DiscNumber,
		DurationMS:   req.DurationMS,
		Genre:        req.Genre,
		ReplayGainDb: req.ReplayGainDb,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, track)
}

// ---- playlist handlers ----

type createPlaylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type patchPlaylistRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type addPlaylistTrackRequest struct {
	TrackID string `json:"trackId"`
}

type setPlaylistTracksRequest struct {
	TrackIDs []string `json:"trackIds"`
}

func (handler *Handler) listPlaylists(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	limit, offset, sortBy, sortOrder, ok := parseCatalogPage(w, r)
	if !ok {
		return
	}
	if so, valid := normalizeSortOrder(sortOrder); valid {
		sortOrder = so
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_sort_order", "sortOrder must be asc or desc")
		return
	}
	pgResult, err := handler.catalogService.ListPlaylistsPage(r.Context(), catalog.ListQuery{
		SortBy: sortBy, SortOrder: sortOrder, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	meta := catalog.CatalogPaginationMeta{
		Limit: limit, Offset: offset, Total: pgResult.Total, HasMore: offset+limit < pgResult.Total,
	}
	writeJSON(w, http.StatusOK, map[string]any{"playlists": pgResult.Items, "pagination": meta})
}

func (handler *Handler) createPlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req createPlaylistRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	pl, err := handler.catalogService.CreatePlaylist(r.Context(), req.Name, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, pl)
}

func (handler *Handler) getPlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	pl, err := handler.catalogService.GetPlaylist(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

func (handler *Handler) deletePlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	if err := handler.catalogService.DeletePlaylist(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) patchPlaylist(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req patchPlaylistRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	pl, err := handler.catalogService.UpdatePlaylist(r.Context(), r.PathValue("id"), catalog.UpdatePlaylistRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

func (handler *Handler) addPlaylistTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req addPlaylistTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	pl, err := handler.catalogService.AddTrackToPlaylist(r.Context(), r.PathValue("id"), req.TrackID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

func (handler *Handler) removePlaylistTrack(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	pl, err := handler.catalogService.RemoveTrackFromPlaylist(r.Context(), r.PathValue("id"), r.PathValue("trackId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

func (handler *Handler) setPlaylistTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	var req setPlaylistTracksRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.TrackIDs == nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "trackIds is required")
		return
	}
	pl, err := handler.catalogService.SetPlaylistTracks(r.Context(), r.PathValue("id"), req.TrackIDs)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

func (handler *Handler) getPlaylistTracks(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	// Playlist tracks have a defined user-curated order; sortBy/sortOrder are not
	// exposed here — only limit/offset for pagination of the ordered list.
	limit, offset, ok := func() (int, int, bool) {
		q := r.URL.Query()
		lim := catalogListDefaultLimit
		off := 0
		if raw := q.Get("limit"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 1 {
				writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
				return 0, 0, false
			}
			if v > catalogListMaxLimit {
				v = catalogListMaxLimit
			}
			lim = v
		}
		if raw := q.Get("offset"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 0 {
				writeAPIError(w, http.StatusBadRequest, "invalid_offset", "offset must be a non-negative integer")
				return 0, 0, false
			}
			off = v
		}
		return lim, off, true
	}()
	if !ok {
		return
	}
	tracks, err := handler.catalogService.GetPlaylistTracks(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	page, meta := paginateCatalog(tracks, limit, offset)
	if isViewerPath(r) {
		user, _ := userFromContext(r)
		views := handler.annotateTracksWithFavorites(r.Context(), user.ID, page)
		writeJSON(w, http.StatusOK, map[string]any{"tracks": views, "pagination": meta})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": page, "pagination": meta})
}

func (handler *Handler) getCatalogStats(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	stats, err := handler.catalogService.GetCatalogStats(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) getArtistStatsBreakdown(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	breakdown, err := handler.catalogService.GetArtistStatsBreakdown(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, breakdown)
}

func (handler *Handler) getAlbumStatsBreakdown(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	breakdown, err := handler.catalogService.GetAlbumStatsBreakdown(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, breakdown)
}

func (handler *Handler) getPlaylistStatsBreakdown(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	breakdown, err := handler.catalogService.GetPlaylistStatsBreakdown(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, breakdown)
}

func (handler *Handler) getRecentlyAdded(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	kind, limit, ok := parseRecentCatalogQuery(w, r)
	if !ok {
		return
	}
	result, err := handler.catalogService.GetRecentlyAdded(r.Context(), kind, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (handler *Handler) getRecentlyUpdated(w http.ResponseWriter, r *http.Request) {
	if !handler.requireCatalogService(w) {
		return
	}
	kind, limit, ok := parseRecentCatalogQuery(w, r)
	if !ok {
		return
	}
	result, err := handler.catalogService.GetRecentlyUpdated(r.Context(), kind, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func parseRecentCatalogQuery(w http.ResponseWriter, r *http.Request) (string, int, bool) {
	q := r.URL.Query()
	kind := q.Get("kind")
	limit := 0
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			writeAPIError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return "", 0, false
		}
		limit = v
	}
	return kind, limit, true
}

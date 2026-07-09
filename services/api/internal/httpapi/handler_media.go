package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"inori-music/services/api/internal/storage"
)

type mediaObjectRequest struct {
	ID             string `json:"id"`
	BackendID      string `json:"backendId"`
	ObjectKey      string `json:"objectKey"`
	ContentHash    string `json:"contentHash"`
	SizeBytes      int64  `json:"sizeBytes"`
	MIMEType       string `json:"mimeType"`
	AssetKind      string `json:"assetKind"`
	LifecycleState string `json:"lifecycleState"`
}

type mediaObjectLifecycleRequest struct {
	LifecycleState string `json:"lifecycleState"`
}

type mediaObjectSelectionRequest struct {
	BackendID          string `json:"backendId,omitempty"`
	ContentHash        string `json:"contentHash,omitempty"`
	VerificationStatus string `json:"verificationStatus,omitempty"`
	LifecycleState     string `json:"lifecycleState,omitempty"`
	AssetKind          string `json:"assetKind,omitempty"`
}

type mediaObjectBulkLifecycleRequest struct {
	Filter         mediaObjectSelectionRequest `json:"filter"`
	LifecycleState string                      `json:"lifecycleState"`
	DryRun         bool                        `json:"dryRun"`
}

func (request mediaObjectRequest) object() storage.MediaObject {
	return storage.MediaObject{
		ID:             request.ID,
		BackendID:      request.BackendID,
		ObjectKey:      request.ObjectKey,
		ContentHash:    request.ContentHash,
		SizeBytes:      request.SizeBytes,
		MIMEType:       request.MIMEType,
		AssetKind:      request.AssetKind,
		LifecycleState: request.LifecycleState,
	}
}

func (request mediaObjectSelectionRequest) filter() storage.MediaObjectSelectionFilter {
	return storage.MediaObjectSelectionFilter{
		BackendID:          request.BackendID,
		ContentHash:        request.ContentHash,
		VerificationStatus: request.VerificationStatus,
		LifecycleState:     request.LifecycleState,
		AssetKind:          request.AssetKind,
	}
}
func (handler *Handler) getMediaObjectDuplicates(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	minCopies, err := parseMediaObjectMinCopies(r.URL.Query().Get("minCopies"))
	if err != nil {
		writeError(w, err)
		return
	}
	report, err := handler.mediaObjects.GetMediaObjectDuplicates(r.Context(), r.URL.Query().Get("backendId"), minCopies)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (handler *Handler) registerMediaObject(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	var request mediaObjectRequest
	if err := decodeJSONWithSentinel(w, r, &request, storage.ErrInvalidMediaObject); err != nil {
		writeError(w, err)
		return
	}
	registered, err := handler.mediaObjects.RegisterMediaObject(r.Context(), request.object())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, registered)
}

func (handler *Handler) getMediaObjectStats(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	stats, err := handler.mediaObjects.GetMediaObjectStats(r.Context(), r.URL.Query().Get("backendId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) setMediaObjectsLifecycle(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	var request mediaObjectBulkLifecycleRequest
	if err := decodeJSONWithSentinel(w, r, &request, storage.ErrInvalidMediaObject); err != nil {
		writeError(w, err)
		return
	}
	report, err := handler.mediaObjects.SetMediaObjectLifecycleStateByFilterWithOptions(r.Context(), request.Filter.filter(), request.LifecycleState, storage.MediaObjectLifecycleUpdateOptions{DryRun: request.DryRun})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (handler *Handler) setMediaObjectLifecycle(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	var request mediaObjectLifecycleRequest
	if err := decodeJSONWithSentinel(w, r, &request, storage.ErrInvalidMediaObject); err != nil {
		writeError(w, err)
		return
	}
	object, err := handler.mediaObjects.SetMediaObjectLifecycleState(r.Context(), r.PathValue("id"), request.LifecycleState)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, object)
}

func (handler *Handler) getMediaObject(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	object, err := handler.mediaObjects.GetMediaObject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, object)
}

type patchMediaObjectRequest struct {
	AssetKind *string `json:"assetKind"`
	MIMEType  *string `json:"mimeType"`
}

func (handler *Handler) patchMediaObject(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	var req patchMediaObjectRequest
	if err := decodeJSONWithSentinel(w, r, &req, storage.ErrInvalidMediaObject); err != nil {
		writeError(w, err)
		return
	}
	object, err := handler.mediaObjects.UpdateMediaObjectMetadata(r.Context(), r.PathValue("id"), storage.UpdateMediaObjectMetadataRequest{
		AssetKind: req.AssetKind,
		MIMEType:  req.MIMEType,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, object)
}

func (handler *Handler) getMediaObjectTimeline(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	timeline, err := handler.mediaObjects.GetMediaObjectTimeline(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, timeline)
}

func (handler *Handler) verifyMediaObjects(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	backendID := strings.TrimSpace(r.URL.Query().Get("backendId"))
	contentHash := strings.TrimSpace(r.URL.Query().Get("contentHash"))
	if (backendID == "" && contentHash == "") || (backendID != "" && contentHash != "") {
		writeError(w, fmt.Errorf("%w: exactly one of backendId or contentHash is required", storage.ErrInvalidMediaObject))
		return
	}
	var (
		report storage.MediaObjectVerificationReport
		err    error
	)
	if backendID != "" {
		report, err = handler.mediaObjects.VerifyMediaObjectsByBackend(r.Context(), backendID)
	} else {
		report, err = handler.mediaObjects.VerifyMediaObjectsByContentHash(r.Context(), contentHash)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (handler *Handler) verifyMediaObject(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	result, err := handler.mediaObjects.VerifyMediaObject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (handler *Handler) listMediaObjects(w http.ResponseWriter, r *http.Request) {
	if handler.mediaObjects == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "media_registry_not_configured", "media object registry is not configured")
		return
	}
	limit, err := parseMediaObjectListInt(r.URL.Query().Get("limit"), "limit", storage.DefaultMediaObjectListLimit)
	if err != nil {
		writeError(w, err)
		return
	}
	offset, err := parseMediaObjectListInt(r.URL.Query().Get("offset"), "offset", 0)
	if err != nil {
		writeError(w, err)
		return
	}
	page, err := handler.mediaObjects.ListMediaObjects(r.Context(), storage.MediaObjectListFilter{
		BackendID:          r.URL.Query().Get("backendId"),
		ContentHash:        r.URL.Query().Get("contentHash"),
		VerificationStatus: r.URL.Query().Get("verificationStatus"),
		LifecycleState:     r.URL.Query().Get("lifecycleState"),
		AssetKind:          r.URL.Query().Get("assetKind"),
		SortBy:             r.URL.Query().Get("sortBy"),
		SortOrder:          r.URL.Query().Get("sortOrder"),
		Limit:              limit,
		Offset:             offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func parseMediaObjectMinCopies(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 2, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%w: minCopies must be an integer", storage.ErrInvalidMediaObject)
	}
	if value < 2 {
		return 0, fmt.Errorf("%w: minCopies must be at least 2", storage.ErrInvalidMediaObject)
	}
	return value, nil
}

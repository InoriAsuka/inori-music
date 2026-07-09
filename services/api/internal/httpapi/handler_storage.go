package httpapi

import (
	"fmt"
	"net/http"

	"inori-music/services/api/internal/storage"
)

type storageBackendRequest struct {
	ID          string                `json:"id"`
	Type        storage.BackendType   `json:"type"`
	DisplayName string                `json:"displayName"`
	Enabled     bool                  `json:"enabled"`
	IsDefault   bool                  `json:"isDefault"`
	Priority    int                   `json:"priority"`
	Config      storage.BackendConfig `json:"config"`
}

func (request storageBackendRequest) backend() storage.StorageBackend {
	return storage.StorageBackend{
		ID:          request.ID,
		Type:        request.Type,
		DisplayName: request.DisplayName,
		Enabled:     request.Enabled,
		IsDefault:   request.IsDefault,
		Priority:    request.Priority,
		Config:      request.Config,
	}
}

func (handler *Handler) listStorageBackends(w http.ResponseWriter, r *http.Request) {
	backends, err := handler.storage.ListBackends(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"backends": backends})
}

func (handler *Handler) getStorageBackend(w http.ResponseWriter, r *http.Request) {
	backend, err := handler.storage.GetBackend(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, backend)
}

func (handler *Handler) registerStorageBackend(w http.ResponseWriter, r *http.Request) {
	var request storageBackendRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, err)
		return
	}
	registered, err := handler.storage.RegisterBackend(r.Context(), request.backend())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, registered)
}

func (handler *Handler) refreshStorageBackends(w http.ResponseWriter, r *http.Request) {
	report, err := handler.storage.RefreshEnabledBackends(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (handler *Handler) validateStorageBackend(w http.ResponseWriter, r *http.Request) {
	var request storageBackendRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, err)
		return
	}
	validated, err := handler.storage.ValidateBackend(request.backend())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, validated)
}

func (handler *Handler) setDefaultStorageBackend(w http.ResponseWriter, r *http.Request) {
	backend, err := handler.storage.SetDefaultBackend(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, backend)
}

func (handler *Handler) disableStorageBackend(w http.ResponseWriter, r *http.Request) {
	backend, err := handler.storage.DisableBackend(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, backend)
}

func (handler *Handler) enableStorageBackend(w http.ResponseWriter, r *http.Request) {
	backend, err := handler.storage.EnableBackend(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, backend)
}

func (handler *Handler) deleteStorageBackend(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// Guard: reject if any media object still references this backend.
	if handler.mediaObjects != nil {
		page, err := handler.mediaObjects.ListMediaObjects(r.Context(), storage.MediaObjectListFilter{
			BackendID: id,
			Limit:     1,
		})
		if err != nil {
			writeError(w, err)
			return
		}
		if page.Pagination.Total > 0 {
			writeError(w, fmt.Errorf("%w: backend %s has %d media object(s) — remove or relocate them first",
				storage.ErrBackendInUse, id, page.Pagination.Total))
			return
		}
	}
	if err := handler.storage.DeleteBackend(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type patchStorageBackendRequest struct {
	DisplayName *string `json:"displayName"`
	Priority    *int    `json:"priority"`
}

func (handler *Handler) patchStorageBackend(w http.ResponseWriter, r *http.Request) {
	var req patchStorageBackendRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	backend, err := handler.storage.UpdateBackend(r.Context(), r.PathValue("id"), storage.UpdateBackendRequest{
		DisplayName: req.DisplayName,
		Priority:    req.Priority,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, backend)
}

func (handler *Handler) probeStorageBackend(w http.ResponseWriter, r *http.Request) {
	result, err := handler.storage.ProbeBackend(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (handler *Handler) getStorageBackendCapacity(w http.ResponseWriter, r *http.Request) {
	report, err := handler.storage.GetBackendCapacity(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (handler *Handler) getStorageBackendHealth(w http.ResponseWriter, r *http.Request) {
	result, err := handler.storage.GetBackendHealth(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

package handler

import (
	"net/http"

	"rinofinance-api/internal/adapters/http/dto"
	apppluggy "rinofinance-api/internal/application/pluggy"
)

// PluggyHandler exposes the Open Finance sync endpoint: it pulls a Pluggy
// connection's accounts and transactions into the user's data.
type PluggyHandler struct {
	sync *apppluggy.SyncItemUseCase
}

// NewPluggyHandler wires the dependencies for PluggyHandler.
func NewPluggyHandler(sync *apppluggy.SyncItemUseCase) *PluggyHandler {
	return &PluggyHandler{sync: sync}
}

// Sync handles POST /api/pluggy/sync.
func (h *PluggyHandler) Sync(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req dto.PluggySyncRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	result, err := h.sync.Execute(r.Context(), userID, req.ItemID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewPluggySyncResponse(result))
}

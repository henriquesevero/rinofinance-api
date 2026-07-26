package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"rinofinance-api/internal/adapters/http/dto"
	apppluggy "rinofinance-api/internal/application/pluggy"
)

// PluggyHandler exposes the Open Finance endpoints: an authenticated manual
// sync and a public webhook Pluggy calls to trigger auto-updates.
type PluggyHandler struct {
	sync          *apppluggy.SyncItemUseCase
	webhookSecret string
}

// NewPluggyHandler wires the dependencies for PluggyHandler. webhookSecret is
// the token expected on incoming webhook calls (empty disables the check).
func NewPluggyHandler(sync *apppluggy.SyncItemUseCase, webhookSecret string) *PluggyHandler {
	return &PluggyHandler{sync: sync, webhookSecret: webhookSecret}
}

// Sync handles POST /api/pluggy/sync (authenticated, manual).
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

// Webhook handles POST /api/pluggy/webhook (public). Pluggy calls it when a
// connection refreshes; we re-sync that item in the background so data stays
// up to date without the user clicking anything. It always answers 200 fast
// (Pluggy retries on non-2xx) and does the work asynchronously.
func (h *PluggyHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	if h.webhookSecret != "" && r.URL.Query().Get("token") != h.webhookSecret {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var payload struct {
		Event  string `json:"event"`
		ItemID string `json:"itemId"`
	}
	_ = decodeJSON(r, &payload) // best-effort; we still ack

	// Acknowledge immediately so Pluggy doesn't retry while we work.
	w.WriteHeader(http.StatusOK)

	if payload.ItemID == "" || !isSyncEvent(payload.Event) {
		return
	}
	itemID := payload.ItemID
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		res, err := h.sync.ExecuteByItem(ctx, itemID)
		if err != nil {
			if !errors.Is(err, apppluggy.ErrItemNotLinked) {
				log.Printf("pluggy webhook: erro ao sincronizar item %s: %v", itemID, err)
			}
			return
		}
		log.Printf("pluggy webhook: item %s sincronizado (contas=%d, importadas=%d)", itemID, res.AccountsSynced, res.TransactionsImported)
	}()
}

// isSyncEvent reports whether an event means fresh account/transaction data.
func isSyncEvent(event string) bool {
	switch event {
	case "item/created",
		"item/updated",
		"transactions/created",
		"transactions/updated",
		"transactions/deleted":
		return true
	default:
		return false
	}
}

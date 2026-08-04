package handler

import (
	"net/http"

	"rinofinance-api/internal/adapters/push"
	"rinofinance-api/internal/domain/notification"
)

type NotificationHandler struct {
	subscriptions  notification.Repository
	scheduler      *push.Scheduler
	vapidPublicKey string
}

func NewNotificationHandler(subscriptions notification.Repository, scheduler *push.Scheduler, vapidPublicKey string) *NotificationHandler {
	return &NotificationHandler{subscriptions: subscriptions, scheduler: scheduler, vapidPublicKey: vapidPublicKey}
}

func (h *NotificationHandler) VapidPublicKey(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"publicKey": h.vapidPublicKey})
}

type subscribeRequest struct {
	Endpoint string `json:"endpoint"`
	P256DH   string `json:"p256dh"`
	Auth     string `json:"auth"`
}

func (h *NotificationHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req subscribeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.Endpoint == "" {
		writeError(w, errBadRequest)
		return
	}
	sub := notification.NewPushSubscription(userID, req.Endpoint, req.P256DH, req.Auth)
	if err := h.subscriptions.Save(r.Context(), sub); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req struct {
		Endpoint string `json:"endpoint"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.Endpoint == "" {
		writeError(w, errBadRequest)
		return
	}
	if err := h.subscriptions.DeleteByEndpoint(r.Context(), userID, req.Endpoint); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) SendTest(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	if err := h.scheduler.SendNow(r.Context(), userID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

package notification_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"api-gateway/internal/models"

	"github.com/go-chi/render"
	"github.com/hesoyamTM/nbf-auth/pkg/auth"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"go.uber.org/zap"
)

type NotificationService interface {
	GetNotificationList(ctx context.Context, uid string) ([]models.Notification, error)
	GettingNotification(ctx context.Context, uid string) (<-chan models.Notification, error)
	ReadNotifications(ctx context.Context, ids []string) error
}

type NotificationHandler struct {
	notificationService NotificationService
}

func NewNotificationHandler(notificationService NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	uid, ok := ctx.Value(auth.UID).(string)
	if !ok || uid == "" {
		log.Error("uid not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notifications, err := h.notificationService.GetNotificationList(ctx, uid)
	if err != nil {
		log.Error("Failed to get Notifications", zap.Error(err))
		http.Error(w, "Failed to get Notifications", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, notifications)
	render.Status(r, http.StatusOK)
}

func (h *NotificationHandler) ReadNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var req ReadNotificationsRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("Failed to decode JSON", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.notificationService.ReadNotifications(ctx, req.IDs); err != nil {
		log.Error("Failed to read notification", zap.Error(err))
		http.Error(w, "Failed to read notification", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}

func (h *NotificationHandler) GettingNotifications(w http.ResponseWriter, r *http.Request) {
	// ctx, cancel := context.WithCancel(r.Context())
	// defer cancel()

	ctx := r.Context()

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	uid, ok := ctx.Value(auth.UID).(string)
	if !ok || uid == "" {
		log.Error("uid not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notificationCh, err := h.notificationService.GettingNotification(ctx, uid)
	if err != nil {
		log.Error("Failed to get notification", zap.Error(err))
		http.Error(w, "Failed to get notification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	_, err = fmt.Fprintf(w, "data:\n\n")
	if err != nil {
		log.Error("Failed to write header", zap.Error(err))
		return
	}
	flusher.Flush()

	for {
		select {
		case notification, ok := <-notificationCh:
			if !ok {
				log.Info("Notification channel closed")
				return
			}
			data, err := json.Marshal(notification)
			if err != nil {
				log.Error("Failed to marshal notification", zap.Error(err))
				return
			}

			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				log.Error("Failed to encode notification", zap.Error(err))
				return
			}
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}

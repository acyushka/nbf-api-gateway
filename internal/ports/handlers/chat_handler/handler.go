// Package chat_handler provides a handler for chat service
package chat_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"api-gateway/internal/models"

	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
	"github.com/hesoyamTM/nbf-auth/pkg/auth"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatClient interface {
	SendMessage(
		ctx context.Context,
		chatID string,
		userID string,
		groupID string,
		inputMessageCh <-chan models.InputMessage,
	) (<-chan models.OutputMessage, error)
	GetChatList(
		ctx context.Context,
		userID string,
	) ([]models.Chat, error)
	GetMessageEvents(
		ctx context.Context,
		userID string,
	) (<-chan models.OutputMessage, error)
}

type ChatHandler struct {
	chatClient ChatClient
}

func NewChatHandler(chatClient ChatClient) *ChatHandler {
	return &ChatHandler{
		chatClient: chatClient,
	}
}

func (h *ChatHandler) GetChatList(w http.ResponseWriter, r *http.Request) {
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
	chats, err := h.chatClient.GetChatList(ctx, uid)
	if err != nil {
		log.Error("Failed to get chat list", zap.Error(err))
		http.Error(w, "Failed to get chat list", http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, chats)
}

func (h *ChatHandler) ServeMessages(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())

	defer cancel()

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Failed to upgrade connection", zap.Error(err))
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Error("Failed to close connection", zap.Error(err))
		}
	}()

	inputMessageCh := make(chan models.InputMessage)

	go func() {
		defer close(inputMessageCh)
		defer cancel()

		for {
			var inputMessage models.InputMessage
			err := conn.ReadJSON(&inputMessage)
			if err != nil {
				log.Error("Failed to read message", zap.Error(err))
				return
			}

			inputMessageCh <- inputMessage
		}
	}()

	outputMessageCh, err := h.chatClient.SendMessage(ctx,
		r.URL.Query().Get("chat_id"),
		r.URL.Query().Get("user_id"),
		r.URL.Query().Get("group_id"),
		inputMessageCh,
	)
	if err != nil {
		log.Error("Failed to send message", zap.Error(err))
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case outputMessage, ok := <-outputMessageCh:
			if !ok {
				return
			}
			if err := conn.WriteJSON(outputMessage); err != nil {
				log.Error("Failed to write message", zap.Error(err))
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (h *ChatHandler) GetMessageEvents(w http.ResponseWriter, r *http.Request) {
	// ctx, cancel := context.WithCancel(r.Context())
	// defer cancel()

	ctx := r.Context()

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	log.Info("GetMessageEvents")

	uid, ok := ctx.Value(auth.UID).(string)
	if !ok || uid == "" {
		log.Error("uid not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Info("Got uid", zap.String("uid", uid))

	messageCh, err := h.chatClient.GetMessageEvents(ctx, uid)
	if err != nil {
		log.Error("Failed to get message events", zap.Error(err))
		http.Error(w, "Failed to get message events", http.StatusInternalServerError)
		return
	}

	log.Info("Got message channel")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Access-Control-Expose-Headers", "*")

	log.Info("Set headers")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	log.Info("Flusher")

	_, err = fmt.Fprintf(w, "event: connect\ndata:\n\n")
	if err != nil {
		log.Error("Failed to write header", zap.Error(err))
		return
	}
	flusher.Flush()

	log.Info("Flushed header")

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	log.Info("Heartbeat")

	for {
		select {
		case <-heartbeat.C:
			if _, err := fmt.Fprintf(w, "event: heartbeat\ndata:\n\n"); err != nil {
				log.Error("Failed to write heartbeat", zap.Error(err))
				return
			}
			flusher.Flush()
		case msg, ok := <-messageCh:
			if !ok {
				log.Info("Notification channel closed")
				return
			}
			data, err := json.Marshal(msg)
			if err != nil {
				log.Error("Failed to marshal message", zap.Error(err))
				return
			}

			if _, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", data); err != nil {
				log.Error("Failed to encode message", zap.Error(err))
				return
			}
			flusher.Flush()
		case <-ctx.Done():
			log.Info("Context done")
			return
		}
	}
}

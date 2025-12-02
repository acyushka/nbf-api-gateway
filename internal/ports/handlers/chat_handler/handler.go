// Package chat_handler provides a handler for chat service
package chat_handler

import (
	"context"
	"net/http"

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
		defer func() {
			if err := conn.Close(); err != nil {
				log.Error("Failed to close connection", zap.Error(err))
			}
		}()
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
		case outputMessage := <-outputMessageCh:
			if err := conn.WriteJSON(outputMessage); err != nil {
				log.Error("Failed to write message", zap.Error(err))
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

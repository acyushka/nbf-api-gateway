// Package chat_handler provides a handler for chat service
package chat_handler

import (
	"context"
	"net/http"

	"api-gateway/internal/models"

	"github.com/gorilla/websocket"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ChatClient interface {
	SendMessage(
		ctx context.Context,
		chatID string,
		userID string,
		groupID string,
		inputMessageCh <-chan models.InputMessage,
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

func (h *ChatHandler) ServeMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	defer conn.Close()

	inputMessageCh := make(chan models.InputMessage)

	go func() {
		defer close(inputMessageCh)

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

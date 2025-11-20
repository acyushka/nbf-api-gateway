// Package chat provides a client for chat service
package chat

import (
	"context"
	"fmt"

	"api-gateway/internal/models"

	"github.com/hesoyamTM/nbf-auth/pkg/auth"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	chatv1 "github.com/hesoyamTM/nbf-protos/gen/go/chat"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	api chatv1.ChatServiceClient
}

func New(ctx context.Context, address string) (*Client, error) {
	cc, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		api: chatv1.NewChatServiceClient(cc),
	}, nil
}

func (c *Client) SendMessage(
	ctx context.Context,
	chatID string,
	userID string,
	groupID string,
	inputMessageCh <-chan models.InputMessage,
) (<-chan models.OutputMessage, error) {
	const op = "chat_client.SendMessage"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	outputMessageCh := make(chan models.OutputMessage)

	uid, ok := ctx.Value(auth.UID).(string)
	if !ok {
		return nil, fmt.Errorf("%s: user id not found in context", op)
	}

	md := make(map[string]string)
	md["chat_id"] = chatID
	md["user_id"] = userID
	md["group_id"] = groupID
	md["uid"] = uid

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(md))

	stream, err := c.api.SendMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		defer close(outputMessageCh)

		for {
			resp, err := stream.Recv()
			if err != nil {
				log.Error("failed to receive message from stream", zap.Error(err))
				return
			}

			outputMessageCh <- models.OutputMessage{
				ChatID: resp.GetChatId(),
				Sender: models.ChatUser{
					ID:     resp.GetUser().GetId(),
					Name:   resp.GetUser().GetName(),
					Avatar: resp.GetUser().GetAvatar(),
				},
				Content:     resp.GetText(),
				ContentType: "text", // TODO: add content type
				CreatedAt:   resp.GetCreatedAt().AsTime(),
			}
		}
	}()

	go func() {
		for inputMessage := range inputMessageCh {
			if err := stream.Send(&chatv1.SendMessageRequest{
				UserId: uid,
				Text:   inputMessage.Content,
			}); err != nil {
				log.Error("failed to send message to stream", zap.Error(err))
				return
			}
		}
	}()

	return outputMessageCh, nil
}

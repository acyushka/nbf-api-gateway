package notification

import (
	"context"
	"fmt"

	"api-gateway/internal/models"

	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	notificationv1 "github.com/hesoyamTM/nbf-protos/gen/go/notification"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api notificationv1.NotificationServiceClient
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
		api: notificationv1.NewNotificationServiceClient(cc),
	}, nil
}

func (c *Client) GetNotificationList(
	ctx context.Context,
	userID string,
) ([]models.Notification, error) {
	const op = "notification_client.GetNotificationList"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("GetNotificationList", zap.String("user_id", userID))

	resp, err := c.api.GetNotificationList(
		ctx,
		&notificationv1.GetNotificationListRequest{
			UserId: userID,
		},
	)
	if err != nil {
		log.Error("failed to get notification list", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	notifications := make([]models.Notification, len(resp.GetNotifications()))
	for i, notification := range resp.GetNotifications() {
		notifications[i] = models.Notification{
			ID:        notification.GetId(),
			Title:     notification.GetTitle(),
			Body:      notification.GetBody(),
			IsRead:    notification.GetRead(),
			CreatedAt: notification.GetCreatedAt().AsTime(),
		}
	}

	return notifications, nil
}

func (c *Client) ReadNotifications(
	ctx context.Context,
	ids []string,
) error {
	const op = "notification_client.ReadNotifications"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("ReadNotifications", zap.Strings("ids", ids))

	_, err = c.api.ReadNotifications(
		ctx,
		&notificationv1.ReadNotificationsRequest{
			NotificationIds: ids,
		},
	)
	if err != nil {
		log.Error("failed to read notifications", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) GettingNotification(
	ctx context.Context,
	id string,
) (<-chan models.Notification, error) {
	const op = "notification_client.GetNotification"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("GetNotification", zap.String("id", id))

	stream, err := c.api.GettingNotification(
		ctx,
		&notificationv1.GettingNotificationRequest{
			UserId: id,
		},
	)
	if err != nil {
		log.Error("failed to get notification", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	notificationCh := make(chan models.Notification)

	go func() {
		defer stream.CloseSend()
		defer close(notificationCh)

		for {
			resp, err := stream.Recv()
			if err != nil {
				log.Error("failed to receive notification from stream", zap.Error(err))
				return
			}

			notification := models.Notification{
				ID:        resp.GetNotification().GetId(),
				Title:     resp.GetNotification().GetTitle(),
				Body:      resp.GetNotification().GetBody(),
				IsRead:    resp.GetNotification().GetRead(),
				CreatedAt: resp.GetNotification().GetCreatedAt().AsTime(),
			}

			notificationCh <- notification
		}
	}()

	return notificationCh, nil
}

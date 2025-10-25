package user

import (
	models "api-gateway/internal/ports/handlers/user_handler"
	"context"

	authInt "github.com/hesoyamTM/nbf-auth/pkg/auth"
	userv1 "github.com/hesoyamTM/nbf-protos/gen/go/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserInfo struct {
	ID          string
	Name        string
	Surname     string
	Contacts    []string
	Avatar      string
	Description string
}

type Client struct {
	api userv1.UserClient
}

func New(ctx context.Context, address string) (*Client, error) {
	cc, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInt.SettingMetadataInterceptor()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		api: userv1.NewUserClient(cc),
	}, nil
}

func (c *Client) CreateUser(ctx context.Context, user *models.User) error {
	_, err := c.api.UpdateUser(ctx, &userv1.UpdateUserRequest{
		User: &userv1.UserInfo{
			Id:          user.ID,
			Name:        user.Name,
			Surname:     user.Surname,
			Contacts:    user.Contacts,
			Description: user.Description,
		},
	})

	return err
}

func (c *Client) GetUser(ctx context.Context, uid string) (*models.User, error) {
	resp, err := c.api.GetUser(ctx, &userv1.GetUserRequest{
		Id: uid,
	})
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:          resp.User.GetId(),
		Name:        resp.User.GetName(),
		Surname:     resp.User.GetSurname(),
		Contacts:    resp.User.GetContacts(),
		Description: resp.User.GetDescription(),
	}, nil
}

func (c *Client) GetUsers(ctx context.Context, uids []string) ([]*models.User, error) {
	resp, err := c.api.GetUsers(ctx, &userv1.GetUsersRequest{
		Ids: uids,
	})
	if err != nil {
		return nil, err
	}

	users := make([]*models.User, len(resp.GetUsers()))
	for i, user := range resp.GetUsers() {
		users[i] = &models.User{
			ID:          user.GetId(),
			Name:        user.GetName(),
			Surname:     user.GetSurname(),
			Contacts:    user.GetContacts(),
			Description: user.GetDescription(),
		}
	}

	return users, nil
}

func (c *Client) UpdateUser(ctx context.Context, user *models.User) error {

	_, err := c.api.UpdateUser(ctx, &userv1.UpdateUserRequest{
		User: &userv1.UserInfo{
			Id:          user.ID,
			Name:        user.Name,
			Surname:     user.Surname,
			Contacts:    user.Contacts,
			Description: user.Description,
		},
	})

	return err
}

func (c *Client) DeleteUser(ctx context.Context, uid string) error {
	_, err := c.api.DeleteUser(ctx, &userv1.DeleteUserRequest{
		Id: uid,
	})

	return err
}

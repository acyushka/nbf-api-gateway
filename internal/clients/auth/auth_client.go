package auth

import (
	handler "api-gateway/internal/ports/handlers/auth_handler"
	"context"

	authv1 "github.com/hesoyamTM/nbf-protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api authv1.AuthClient
}

func New(ctx context.Context, address string) (*Client, error) {
	cc, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		api: authv1.NewAuthClient(cc),
	}, nil
}

func (c *Client) Register(ctx context.Context, phone_number, name, surname string) (string, error) {
	resp, err := c.api.Register(ctx, &authv1.RegisterRequest{
		PhoneNumber: phone_number,
		Name:        name,
		Surname:     surname,
	})
	if err != nil {
		return "", err
	}

	return resp.GetToken(), err
}

func (c *Client) Login(ctx context.Context, phone_number string) (string, error) {
	resp, err := c.api.Login(ctx, &authv1.LoginRequest{
		PhoneNumber: phone_number,
	})
	if err != nil {
		return "", err
	}

	return resp.GetToken(), err
}

func (c *Client) Logout(ctx context.Context, refresh_token string) error {
	_, err := c.api.Logout(ctx, &authv1.LogoutRequest{
		RefreshToken: refresh_token,
	})
	if err != nil {
		return err
	}

	return nil
}

/*
func (c *Client) VerifyPhoneNumber(ctx context.Context, token, code string) (handler.Tokens, error) {

}
*/
func (c *Client) RefreshToken(ctx context.Context, token string) (*handler.Tokens, error) {
	resp, err := c.api.RefreshToken(ctx, &authv1.RefreshTokenRequest{
		RefreshToken: token,
	})
	if err != nil {
		return nil, err
	}

	return &handler.Tokens{
		AccessToken:       resp.GetAccessToken(),
		RefreshToken:      resp.GetRefreshToken(),
		Access_expire_at:  resp.GetAccessExpireAt().AsTime(),
		Refresh_expire_at: resp.GetRefreshExpireAt().AsTime(),
	}, nil
}

func (c *Client) GoogleLoginURL(ctx context.Context) (string, error) {
	resp, err := c.api.GoogleLoginURL(ctx, &authv1.GoogleLoginURLRequest{})
	if err != nil {
		return "", err
	}

	return resp.GetUrl(), nil
}

func (c *Client) GoogleAuthorize(ctx context.Context, state, code string) (*handler.Tokens, error) {
	resp, err := c.api.GoogleAuthorize(ctx, &authv1.GoogleAuthorizeRequest{
		State: state,
		Code:  code,
	})
	if err != nil {
		return nil, err
	}

	return &handler.Tokens{
		AccessToken:       resp.GetAccessToken(),
		RefreshToken:      resp.GetRefreshToken(),
		Access_expire_at:  resp.GetAccessExpireAt().AsTime(),
		Refresh_expire_at: resp.GetRefreshExpireAt().AsTime(),
	}, nil
}

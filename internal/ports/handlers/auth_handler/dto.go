package auth_handler

import "time"

//requests

type RegisterRequest struct {
	Name        string
	Surname     string
	PhoneNumber string
}

type LoginRequest struct {
	PhoneNumber string
}

type LogoutRequest struct {
	RefreshToken string
}

//dto for interface

type Tokens struct {
	AccessToken       string
	RefreshToken      string
	Access_expire_at  time.Time
	Refresh_expire_at time.Time
}

/*
type VerifyPhoneResponce struct {
	User_ID           string
	AccessToken       string
	RefreshToken      string
	Access_expire_at  time.Time
	Refresh_expire_at time.Time
}
type RefreshTokenResponce struct {
	User_ID           string
	AccessToken       string
	RefreshToken      string
	Access_expire_at  time.Time
	Refresh_expire_at time.Time
}
type GoogleAuthorizeResponce struct {
	AccessToken       string
	RefreshToken      string
	Access_expire_at  time.Time
	Refresh_expire_at time.Time
}
*/

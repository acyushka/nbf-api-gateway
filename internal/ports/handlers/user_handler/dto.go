package user_handler

import "io"

type User struct {
	ID          string
	Name        string
	Surname     string
	Contacts    []string
	Avatar      string
	Description string
} // @name User

// @Description Get users response
type GetUsersResponse struct {
	users []*User
} // @name GetUsersResponse

// @Description Update user request
type UpdateUserRequest struct {
	Name        string
	Surname     string
	Contacts    []string
	Description string
} // @name UpdateUserRequest

type FilePhoto struct {
	Data        io.Reader `json:"data"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
}

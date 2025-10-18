package user_handler

type User struct {
	ID          string
	Name        string
	Surname     string
	Contacts    []string
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

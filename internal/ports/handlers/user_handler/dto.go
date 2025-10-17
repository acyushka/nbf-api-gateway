package user_handler

type User struct {
	ID          string
	Name        string
	Surname     string
	Contacts    []string
	Description string
}

type GetUsersResponce struct {
	users []*User
}

type UpdateUserRequest struct {
	Name        string
	Surname     string
	Contacts    []string
	Description string
}

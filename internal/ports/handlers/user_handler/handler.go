package user_handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	authorization "github.com/hesoyamTM/nbf-auth/pkg/auth"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"go.uber.org/zap"
)

type UserClient interface {
	CreateUser(ctx context.Context, NewUser *User) error
	GetUser(ctx context.Context, UserID string) (*User, error)
	GetUsers(ctx context.Context, ids []string) ([]*User, error)
	UpdateUser(ctx context.Context, UserModel *User) error
	DeleteUser(ctx context.Context, UserID string) error
}

type UserHandler struct {
	userClient UserClient
}

func NewUserHandler(userClient UserClient) *UserHandler {
	return &UserHandler{
		userClient: userClient,
	}
}

// @Summary Get session
// @Description Get information about user session
// @Tags auth
// @Accept json
// @Produce json
// @Param authorization header string true "Authorization header"
// @Success 200 {object} User
// @Failure 401
// @Failure 404
// @Failure 500
// @Router /user/session [get]
func (c *UserHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	uid, ok := ctx.Value(authorization.UID).(string)
	if !ok || uid == "" {
		log.Error("uid not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := c.userClient.GetUser(ctx, uid)
	if err != nil {
		log.Error("Failed to authorize user", zap.Error(err))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	render.JSON(w, r, user)
}

// @Summary Create user
// @Description Создать нового пользователя в системе
// @Tags user
// @Accept json
// @Param user body User true "User data"
// @Success 201 "User created successfully"
// @Failure 400
// @Failure 500
// @Router /user [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var req User

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("Failed to decode JSON", zap.Error(err))

		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Surname == "" {
		http.Error(w, "Name or surname is empty", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err = h.userClient.CreateUser(ctx, &User{
		ID:          req.ID,
		Name:        req.Name,
		Surname:     req.Surname,
		Contacts:    req.Contacts,
		Description: req.Description,
	})
	if err != nil {
		log.Error("Failed to create user", zap.Error(err))

		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusCreated)
}

// @Summary Get user
// @Description Найти пользователя по ID
// @Tags user
// @Accept json
// @Produce json
// @Param uid path string true "User id"
// @Success 200 {object} User
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /user/{uid} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	uid := chi.URLParam(r, "uid")

	ctx := r.Context()
	user, err := h.userClient.GetUser(ctx, uid)
	if err != nil {
		log.Error("Failed to get user", zap.Error(err))

		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	render.JSON(w, r, user)

	render.Status(r, http.StatusOK)
}

// @Summary Get users
// @Description Найти несколько пользователей по ID
// @Tags user
// @Accept json
// @Produce json
// @Param uids query string true "Comma-separated user IDs" example("id1,id2,id3")
// @Success 200 {object} GetUsersResponse
// @Failure 400
// @Failure 500
// @Router /users [get]
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	uidsParam := r.URL.Query().Get("ids")
	uids := strings.Split(uidsParam, ",")

	if len(uids) == 0 {
		log.Error("No valid user IDs provided")

		http.Error(w, "No valid user IDs provided", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	users, err := h.userClient.GetUsers(ctx, uids)
	if err != nil {
		log.Error("Failed to get users", zap.Error(err))

		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	response := &GetUsersResponse{
		users: users,
	}

	render.JSON(w, r, response)
	render.Status(r, http.StatusOK)
}

// @Summary Update user
// @Description Обновить данные пользователя пользователя в системе
// @Tags user
// @Accept json
// @Param uid path string true "User ID"
// @Param request body UpdateUserRequest true "User data"
// @Success 200
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /user [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var req UpdateUserRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("Failed to decode JSON", zap.Error(err))

		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	uid, ok := ctx.Value(authorization.UID).(string)
	if !ok || uid == "" {
		log.Error("uid not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.userClient.UpdateUser(ctx, &User{
		ID:          uid,
		Name:        req.Name,
		Surname:     req.Surname,
		Contacts:    req.Contacts,
		Description: req.Description,
	})
	if err != nil {
		log.Error("Failed to update user", zap.Error(err))

		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)

}

// @Summary Delete user
// @Description Удалить пользователя из системы
// @Tags user
// @Accept json
// @Param uid path string true "User ID"
// @Success 200
// @Failure 401
// @Failure 500
// @Router /user [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	uid, ok := ctx.Value(authorization.UID).(string)
	if !ok || uid == "" {
		log.Error("uid not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.userClient.DeleteUser(ctx, uid); err != nil {
		log.Error("Failed to delete user", zap.Error(err))

		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}

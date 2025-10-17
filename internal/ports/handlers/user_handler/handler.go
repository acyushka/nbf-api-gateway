package user_handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
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

	response := &GetUsersResponce{
		users: users,
	}

	render.JSON(w, r, response)
	render.Status(r, http.StatusOK)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	uid := chi.URLParam(r, "uid")

	var req UpdateUserRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("Failed to decode JSON", zap.Error(err))

		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
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

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	uid := chi.URLParam(r, "uid")

	ctx := r.Context()
	if err := h.userClient.DeleteUser(ctx, uid); err != nil {
		log.Error("Failed to delete user", zap.Error(err))

		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
}

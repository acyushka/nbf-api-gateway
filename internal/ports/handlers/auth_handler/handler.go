package auth_handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"go.uber.org/zap"
)

type AuthClient interface {
	Register(ctx context.Context, phone_number, name, surname string) (string, error)
	Login(ctx context.Context, phone_number string) (string, error)
	Logout(ctx context.Context, refresh_token string) error
	//VerifyPhoneNumber(ctx context.Context, token, code string) (*Tokens, error)
	RefreshToken(ctx context.Context, token string) (*Tokens, error)
	//YandexLoginURL(ctx context.Context) (string, error)
	//YandexAuthorize(ctx context.Context, state, code string) (*Tokens, error)
	GoogleLoginURL(ctx context.Context) (string, error)
	GoogleAuthorize(ctx context.Context, state, code string) (*Tokens, error)
}

type AuthHandler struct {
	authClient AuthClient
}

func NewAuthHandler(authClient AuthClient) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
	}
}

/*
	func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
		log, err := logger.LoggerFromCtx(r.Context())
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)

			return
		}

		var req RegisterRequest

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			//log

			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		token, err := h.authClient.Register(ctx, req.PhoneNumber, req.Name, req.Surname)
		if err != nil {
			//log

			http.Error(w, "Authorization error", http.StatusUnauthorized)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    token,
			HttpOnly: true,
			Domain:   "localhost",
			Path:     "/",
		})

		render.Status(r, http.StatusOK)
	}

	func (c *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
		log, err := logger.LoggerFromCtx(r.Context())
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)

			return
		}

		var req LoginRequest

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			//log

			http.Error(w, "Invalid JSON", http.StatusBadRequest)

			return
		}

		ctx := r.Context()

		token, err := c.authClient.Login(ctx, req.PhoneNumber)
		if err != nil {
			//log

			http.Error(w, "Authorization error", http.StatusUnauthorized)

			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    token,
			HttpOnly: true,
			Domain:   "localhost",
			Path:     "/",
		})

		render.Status(r, http.StatusOK)

}
*/
func (c *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)

		return
	}

	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		log.Error("Failed to fetch refresh cookie", zap.Error(err))

		http.Error(w, "Failed to fetch cookie", http.StatusUnauthorized)

		return
	}
	if refreshCookie.Value == "" {
		log.Error("refresh token value is empty", zap.Error(err))

		http.Error(w, "Empty value of refresh token", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	if err := c.authClient.Logout(ctx, refreshCookie.Value); err != nil {
		log.Error("Logout failed", zap.Error(err))

		http.Error(w, "Logout failed", http.StatusBadRequest)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Domain:   "localhost",
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Domain:   "localhost",
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})

	render.Status(r, http.StatusOK)
}

/*
func (c *AuthHandler) VerifyPhoneNumber(w http.ResponseWriter, r *http.Request) {

}
*/
func (c *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)

		return
	}

	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		log.Error("Failed to fetch refresh cookie", zap.Error(err))

		http.Error(w, "Failed to fetch cookie", http.StatusUnauthorized)

		return
	}
	if refreshCookie.Value == "" {
		log.Error("Refresh token value is empty", zap.Error(err))

		http.Error(w, "Empty value of refresh token", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	resp, err := c.authClient.RefreshToken(ctx, refreshCookie.Value)
	if err != nil {
		log.Error("Failed to fetch refresh cookie", zap.Error(err))

		http.Error(w, "token refresh failed", http.StatusInternalServerError)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    resp.AccessToken,
		HttpOnly: true,
		Domain:   "localhost",
		Path:     "/",
		Expires:  resp.Access_expire_at,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    resp.RefreshToken,
		HttpOnly: true,
		Domain:   "localhost",
		Path:     "/",
		Expires:  resp.Refresh_expire_at,
	})

	render.Status(r, http.StatusOK)
}

/*
func (c *AuthHandler) YandexLoginURL(w http.ResponseWriter, r *http.Request) {

}

func (c *AuthHandler) YandexAuthorize(w http.ResponseWriter, r *http.Request) {

}
*/
func (c *AuthHandler) GoogleLoginURL(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)

		return
	}

	ctx := r.Context()
	url, err := c.authClient.GoogleLoginURL(ctx)
	if err != nil {
		log.Error("failed to get google url", zap.Error(err))

		http.Error(w, "failed to get google url", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"Url": url,
	})

	render.Status(r, http.StatusOK)
}

func (c *AuthHandler) GoogleAuthorize(w http.ResponseWriter, r *http.Request) {
	log, err := logger.LoggerFromCtx(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)

		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	ctx := r.Context()

	resp, err := c.authClient.GoogleAuthorize(ctx, state, code)
	if err != nil {
		log.Error("Failed to google authorize user", zap.Error(err))

		http.Error(w, "authorization failed", http.StatusInternalServerError)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    resp.AccessToken,
		HttpOnly: true,
		Domain:   "localhost",
		Path:     "/",
		Expires:  resp.Access_expire_at,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    resp.RefreshToken,
		HttpOnly: true,
		Domain:   "localhost",
		Path:     "/",
		Expires:  resp.Refresh_expire_at,
	})

	render.Status(r, http.StatusOK)
}

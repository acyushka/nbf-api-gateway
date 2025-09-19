package httpapp

import (
	"api-gateway/internal/clients/auth"
	"api-gateway/internal/config"
	"api-gateway/internal/ports/handlers/auth_handler"
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"go.uber.org/zap"
)

type App struct {
	router     *chi.Mux
	httpServer *http.Server
}

type Clients struct {
	AuthService_Addr string
}

func New(ctx context.Context, cfg *config.Config, clients Clients) *App {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		panic(err)
	}

	router := chi.NewRouter()

	//New Clients

	AuthClient, err := auth.New(ctx, clients.AuthService_Addr)

	if err != nil {
		log.Error("failed to connect auth client", zap.Error(err))
	}

	//handlers

	AuthHandler := auth_handler.NewAuthHandler(AuthClient)

	loggingMiddleware, err := logger.NewLoggingMiddleware(ctx)
	if err != nil {
		panic(err)
	}

	//middlewares

	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(loggingMiddleware)

	//auth

	router.Get("/api/v1/auth/google/login", AuthHandler.GoogleLoginURL)
	router.Get("/api/v1/auth/google/callback", AuthHandler.GoogleAuthorize)
	router.Delete("/api/v1/auth/logout", AuthHandler.Logout)
	router.Head("/api/v1/auth/refresh", AuthHandler.RefreshToken)

	//server

	httpServer := http.Server{
		Addr:         cfg.HTTP_Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTP_Server.Timeout,
		WriteTimeout: cfg.HTTP_Server.Timeout,
		IdleTimeout:  cfg.HTTP_Server.Idle_Timeout,
	}

	return &App{
		router:     router,
		httpServer: &httpServer,
	}
}

func (a *App) MustStart(ctx context.Context) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		panic(err)
	}

	if err := a.httpServer.ListenAndServe(); err != nil {
		log.Error("failed to start http server", zap.Error(err))

		panic(err)
	}
}

func (a *App) MustStop() {
	panic(a.httpServer.Shutdown(context.Background()))
}

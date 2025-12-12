package httpapp

import (
	"context"
	"net/http"

	"api-gateway/internal/clients/auth"
	"api-gateway/internal/clients/chat"
	"api-gateway/internal/clients/matcher"
	"api-gateway/internal/clients/notification"
	s3 "api-gateway/internal/clients/storage"
	"api-gateway/internal/clients/user"
	"api-gateway/internal/config"
	"api-gateway/internal/ports/handlers/auth_handler"
	"api-gateway/internal/ports/handlers/chat_handler"
	"api-gateway/internal/ports/handlers/matcher_handler"
	"api-gateway/internal/ports/handlers/notification_handler"
	"api-gateway/internal/ports/handlers/user_handler"
	"api-gateway/internal/ports/middlewares"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	authMid "github.com/hesoyamTM/nbf-auth/pkg/auth"
	decodeKeys "github.com/hesoyamTM/nbf-auth/pkg/config"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type App struct {
	router     *chi.Mux
	httpServer *http.Server
}

type Clients struct {
	AuthService_Addr         string
	UserService_Addr         string
	MatcherService_Addr      string
	FileStorageService_Addr  string
	ChatService_Addr         string
	NotificationService_Addr string
}

func New(ctx context.Context, cfg *config.Config, clients Clients) *App {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		panic(err)
	}

	router := chi.NewRouter()

	// New Clients

	AuthClient, err := auth.New(ctx, clients.AuthService_Addr)
	if err != nil {
		log.Error("failed to connect auth client", zap.Error(err))
	}

	UserClient, err := user.New(ctx, clients.UserService_Addr)
	if err != nil {
		log.Error("failed to connect user client", zap.Error(err))
	}

	MatcherClient, err := matcher.New(ctx, clients.MatcherService_Addr)
	if err != nil {
		log.Error("failed to connect matcher client", zap.Error(err))
	}

	FileStorageClient, err := s3.New(ctx, clients.FileStorageService_Addr)
	if err != nil {
		log.Error("failed to connect storage client", zap.Error(err))
	}
	ChatClient, err := chat.New(ctx, clients.ChatService_Addr)
	if err != nil {
		log.Error("failed to connect chat client", zap.Error(err))
	}

	NotificationClient, err := notification.New(ctx, clients.NotificationService_Addr)
	if err != nil {
		log.Error("failed to connect notification client", zap.Error(err))
	}

	// handlers

	AuthHandler := auth_handler.NewAuthHandler(AuthClient, cfg.Domain)
	UserHandler := user_handler.NewUserHandler(UserClient, FileStorageClient)
	MatcherHandler := matcher_handler.NewMatcherHandler(MatcherClient, FileStorageClient)
	ChatHandler := chat_handler.NewChatHandler(ChatClient)
	NotificationHandler := notification_handler.NewNotificationHandler(NotificationClient)

	// middlewares

	loggingMiddleware, err := logger.NewLoggingMiddleware(ctx)
	if err != nil {
		panic(err)
	}

	pubKey, err := decodeKeys.DecodePublicKey(cfg.PublicKey)
	if err != nil {
		panic(err)
	}

	authMiddleware := authMid.NewAuthMiddleware("access_token", AuthClient, pubKey)

	router.Use(middlewares.Cors(cfg))
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(loggingMiddleware)

	// auth
	router.Get("/api/v1/auth/google/login", AuthHandler.GoogleLoginURL)
	router.Get("/api/v1/auth/google/callback", AuthHandler.GoogleAuthorize)
	router.Get("/api/v1/auth/yandex/login", AuthHandler.YandexLoginURL)
	router.Get("/api/v1/auth/yandex/callback", AuthHandler.YandexAuthorize)
	router.Head("/api/v1/auth/refresh", AuthHandler.RefreshToken)
	router.With(authMiddleware).Delete("/api/v1/auth/logout", AuthHandler.Logout)

	// user

	router.Post("/api/v1/user", UserHandler.CreateUser)
	router.Get("/api/v1/user/{uid}", UserHandler.GetUser)
	router.Get("/api/v1/users", UserHandler.GetUsers)
	router.With(authMiddleware).Get("/api/v1/user/session", UserHandler.GetSession)
	router.With(authMiddleware).Put("/api/v1/user", UserHandler.UpdateUser)
	router.With(authMiddleware).Delete("/api/v1/user", UserHandler.DeleteUser)

	// matcher

	router.Get("/api/v1/matcher/form/{uid}", MatcherHandler.GetFormByUser)
	router.With(authMiddleware).Post("/api/v1/matcher/form", MatcherHandler.CreateForm)
	router.With(authMiddleware).Put("/api/v1/matcher/form", MatcherHandler.UpdateForm)
	router.With(authMiddleware).Delete("/api/v1/matcher/form/{uid}", MatcherHandler.DeleteForm)

	router.Get("/api/v1/matcher/group/{gid}", MatcherHandler.GetGroup)
	router.Get("/api/v1/matcher/group/user/{uid}", MatcherHandler.GetGroupByUser)
	router.Get("/api/v1/matcher/group/members/{gid}", MatcherHandler.ListGroupMembers)
	router.With(authMiddleware).Delete("/api/v1/matcher/group/{oid}", MatcherHandler.DeleteGroup)
	router.With(authMiddleware).Delete("/api/v1/matcher/group/user", MatcherHandler.LeaveGroup)
	router.With(authMiddleware).Delete("/api/v1/matcher/group/kick/{uid}", MatcherHandler.KickGroup)

	router.With(authMiddleware).Get("/api/v1/matcher/find/{uid}", MatcherHandler.FindGroups)

	router.Get("/api/v1/matcher/group/{gid}/requests", MatcherHandler.GetRequests)
	router.With(authMiddleware).Get("/api/v1/matcher/group/requests", MatcherHandler.GetRequests)
	router.With(authMiddleware).Post("/api/v1/matcher/group/send", MatcherHandler.SendJoinRequest)
	router.With(authMiddleware).Post("/api/v1/matcher/group/accept", MatcherHandler.AcceptJoinRequest)
	router.With(authMiddleware).Post("/api/v1/matcher/group/reject", MatcherHandler.RejectJoinRequest)

	// chat

	router.With(authMiddleware).Get("/api/v1/chat/messages", ChatHandler.ServeMessages)
	router.With(authMiddleware).Get("/api/v1/chat/list", ChatHandler.GetChatList)
	router.With(authMiddleware).Get("/api/v1/chat/messages/stream", ChatHandler.GetMessageEvents)

	// notification

	router.With(authMiddleware).Get("/api/v1/notifications", NotificationHandler.GetNotifications)
	router.With(authMiddleware).Put("/api/v1/notifications/read", NotificationHandler.ReadNotification)
	router.With(authMiddleware).Get("/api/v1/notifications/stream", NotificationHandler.GettingNotifications)

	// swagger
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// server
	httpServer := http.Server{
		Addr:         cfg.HTTP_Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTP_Server.Timeout,
		WriteTimeout: 0,
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

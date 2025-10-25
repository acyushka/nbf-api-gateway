package main

import (
	_ "api-gateway/docs"
	"api-gateway/internal/app"
	httpapp "api-gateway/internal/app/http"
	"api-gateway/internal/config"
	"context"
	"os"
	"os/signal"
	"syscall"

	cfgtools "github.com/hesoyamTM/nbf-auth/pkg/config"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
)

// @title nbf API
// @version 1.0
// @description API Gateway for nbf-project
// @host localhost:8082
// @BasePath /api/v1
func main() {
	cfg := cfgtools.MustParseConfig[config.Config]()
	ctx, err := logger.SetupLogger(context.Background(), cfg.Env)
	if err != nil {
		panic(err)
	}

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		panic(err)
	}

	log.Debug("Logger is working")

	Clients := httpapp.Clients{
		AuthService_Addr:    cfg.GRPC_Clients.AuthService,
		UserService_Addr:    cfg.GRPC_Clients.UserService,
		MatcherService_Addr: cfg.GRPC_Clients.MatcherService,
	}

	application := app.NewApp(ctx, cfg, Clients)

	go application.HttpApp.MustStart(ctx)
	log.Info("Server started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	application.HttpApp.MustStop()

	log.Info("Server is stopped")
}

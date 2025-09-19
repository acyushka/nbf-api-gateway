package main

import (
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
		AuthService_Addr: cfg.GRPC_Clients.AuthService,
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

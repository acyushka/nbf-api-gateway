package app

import (
	httpapp "api-gateway/internal/app/http"
	"api-gateway/internal/config"
	"context"
)

//слой абстракции
//здесь только конструктор httpApp

type App struct {
	HttpApp *httpapp.App
}

func NewApp(
	ctx context.Context,
	cfg *config.Config,
	clients httpapp.Clients,
) *App {
	httpApp := httpapp.New(ctx, cfg, clients)

	return &App{
		HttpApp: httpApp,
	}
}

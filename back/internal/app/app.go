package app

import (
	"context"
	"net/http"

	"github.com/Kibishi47/mini-games/back/internal/config"
	apihttp "github.com/Kibishi47/mini-games/back/internal/http"
	"github.com/Kibishi47/mini-games/back/internal/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	CFG    config.Config
	DB     *pgxpool.Pool
	Router http.Handler
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	router := apihttp.NewRouter(cfg, pool)

	return &App{
		CFG:    cfg,
		DB:     pool,
		Router: router,
	}, nil
}

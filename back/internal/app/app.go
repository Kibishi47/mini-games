package app

import (
	"context"
	"net/http"

	"github.com/Kibishi47/mini-games/back/internal/config"
	"github.com/Kibishi47/mini-games/back/internal/domain/auth"
	apihttp "github.com/Kibishi47/mini-games/back/internal/http"
	"github.com/Kibishi47/mini-games/back/internal/infra/postgres"
	auth2 "github.com/Kibishi47/mini-games/back/internal/infra/postgres/auth"
	"github.com/Kibishi47/mini-games/back/internal/infra/postgres/user"
	"github.com/Kibishi47/mini-games/back/internal/ws"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	CFG    config.Config
	DB     *pgxpool.Pool
	Router http.Handler
	Hub    *ws.Hub
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// ws
	hub := ws.NewHub()
	go hub.Run()

	// user
	userRepo := user.NewPostgresUserRepository(pool)

	// auth
	tokenRepo := auth2.NewPostgresTokenRepository(pool)
	identityRepo := auth2.NewPostgresIdentityRepository(pool)
	authService := auth.NewService(tokenRepo, identityRepo, userRepo, cfg.AuthTokenTTL)

	// container
	deps := &apihttp.Deps{
		Auth: authService,
		WS:   hub,
	}

	// router
	router := apihttp.NewRouter(cfg, deps)

	return &App{
		CFG:    cfg,
		DB:     pool,
		Router: router,
		Hub:    hub,
	}, nil
}

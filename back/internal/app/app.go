package app

import (
	"context"
	"net/http"

	"github.com/Kibishi47/mini-games/back/internal/config"
	"github.com/Kibishi47/mini-games/back/internal/domain/auth"
	apihttp "github.com/Kibishi47/mini-games/back/internal/http"
	"github.com/Kibishi47/mini-games/back/internal/infra/authrepo"
	"github.com/Kibishi47/mini-games/back/internal/infra/db"
	"github.com/Kibishi47/mini-games/back/internal/infra/userrepo"
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

	// user
	userRepo := userrepo.NewPostgresUserRepository(pool)

	// auth
	tokenRepo := authrepo.NewPostgresTokenRepository(pool)
	identityRepo := authrepo.NewPostgresIdentityRepository(pool)
	authService := auth.NewService(tokenRepo, identityRepo, userRepo, cfg.AuthTokenTTL)

	// container
	deps := &apihttp.Deps{
		DB:   pool,
		Auth: authService,
	}

	// router
	router := apihttp.NewRouter(cfg, deps)

	return &App{
		CFG:    cfg,
		DB:     pool,
		Router: router,
	}, nil
}

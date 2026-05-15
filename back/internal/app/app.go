package app

import (
	"context"
	"net/http"

	"github.com/Kibishi47/mini-games/back/internal/config"
	"github.com/Kibishi47/mini-games/back/internal/domain/auth"
	domaingame "github.com/Kibishi47/mini-games/back/internal/domain/game"
	domainroom "github.com/Kibishi47/mini-games/back/internal/domain/room"
	apihttp "github.com/Kibishi47/mini-games/back/internal/http"
	"github.com/Kibishi47/mini-games/back/internal/infra/postgres"
	auth2 "github.com/Kibishi47/mini-games/back/internal/infra/postgres/auth"
	infraroom "github.com/Kibishi47/mini-games/back/internal/infra/postgres/room"
	"github.com/Kibishi47/mini-games/back/internal/infra/postgres/user"
	"github.com/Kibishi47/mini-games/back/internal/infra/redis"
	"github.com/Kibishi47/mini-games/back/internal/ws"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

type App struct {
	CFG    config.Config
	DB     *pgxpool.Pool
	Redis  *goredis.Client
	Router http.Handler
	Hub    *ws.Hub
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	redisClient, err := redis.NewClient(ctx, cfg.RedisURL, cfg.RedisUsername, cfg.RedisPassword)
	if err != nil {
		return nil, err
	}

	// room repo (needed by Hub)
	roomRepo := infraroom.NewPostgresRoomRepository(pool)

	// user
	userRepo := user.NewPostgresUserRepository(pool)

	// ws
	hub := ws.NewHub(redisClient, roomRepo, userRepo)
	go hub.Run()

	// auth
	tokenRepo := auth2.NewPostgresTokenRepository(pool)
	identityRepo := auth2.NewPostgresIdentityRepository(pool)
	authService := auth.NewService(tokenRepo, identityRepo, userRepo, cfg.AuthTokenTTL)

	// room service
	roomService := domainroom.NewService(roomRepo)

	// game service
	gameService := domaingame.NewService()

	// container
	deps := &apihttp.Deps{
		Auth: authService,
		Room: roomService,
		User: userRepo,
		Game: gameService,
		WS:   hub,
	}

	// router
	router := apihttp.NewRouter(cfg, deps)

	return &App{
		CFG:    cfg,
		DB:     pool,
		Redis:  redisClient,
		Router: router,
		Hub:    hub,
	}, nil
}

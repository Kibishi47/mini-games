package http

import (
	"github.com/Kibishi47/mini-games/back/internal/http/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	DB   *pgxpool.Pool
	Auth handlers.AuthUseCase
}

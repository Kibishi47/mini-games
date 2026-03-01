package auth

import (
	"context"
	"time"

	"github.com/Kibishi47/mini-games/back/internal/domain/user"
	"github.com/Kibishi47/mini-games/back/internal/http/middlewares"
	"github.com/google/uuid"
)

type UseCases interface {
	LoginLocal(ctx context.Context, username string, password string) (string, time.Time, error)
	RegisterLocal(ctx context.Context, username string, password string) (string, time.Time, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	Logout(ctx context.Context, rawToken string) error
}

type Service interface {
	middlewares.Authenticator
	UseCases
}

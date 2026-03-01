package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TokenRepository interface {
	CreateToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	FindUserIDByTokenHash(ctx context.Context, tokenHash string) (uuid.UUID, time.Time, error)
	DeleteTokenHash(ctx context.Context, tokenHash string) error
}

type IdentityRepository interface {
	FindUserIDByProvider(ctx context.Context, provider Provider, providerID string) (uuid.UUID, error)
	CreateIdentity(ctx context.Context, userID uuid.UUID, provider Provider, providerID string) error
}
